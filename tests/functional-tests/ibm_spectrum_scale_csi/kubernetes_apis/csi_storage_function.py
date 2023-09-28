import time
import re
import logging
import copy
import urllib3
from datetime import datetime, timezone
from kubernetes import client
from kubernetes.client.rest import ApiException
from kubernetes.stream import stream
import ibm_spectrum_scale_csi.spectrum_scale_apis.fileset_functions as filesetfunc
import ibm_spectrum_scale_csi.common_utils.namegenerator as namegenerator
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
LOGGER = logging.getLogger()


def set_test_namespace_value(namespace_name=None):
    """ sets the test namespace global for use in later functions"""
    global namespace_value
    namespace_value = namespace_name


def set_test_nodeselector_value(plugin_node_selector):
    """ sets the nodeselector global for use in create_pod functions"""
    global nodeselector
    node_selector_labels = {}
    for label_val in plugin_node_selector:
        node_selector_labels[label_val["key"]] = label_val["value"]
    nodeselector = node_selector_labels


def get_random_name(type_of):
    """ return random name of type_of"""
    return f"{type_of}-{namegenerator.name_generator()}"


def create_storage_class(values, sc_name, created_objects):
    """
    creates storage class
    Args:
        param1: values - storage class parameters
        param2: config_value - configuration file
        param3: sc_name - name of storage class to be created
    Returns:
        None
    Raises:
        Raises an exception on kubernetes client api failure and asserts
    """
    global storage_class_parameters
    api_instance = client.StorageV1Api()
    storage_class_metadata = client.V1ObjectMeta(name=sc_name)

    storage_class_parameters = {}
    list_parameters = ["volBackendFs", "clusterId", "volDirBasePath", "uid", "gid",
                       "filesetType", "parentFileset", "inodeLimit", "nodeClass", "permissions",
                       "version", "compression", "tier", "consistencyGroup", "shared"]

    if "version" in values and values["version"] == "2" and "consistencyGroup" not in values:
        values["consistencyGroup"] = get_random_name("cg")

    for sc_parameter in list_parameters:
        if sc_parameter in values:
            storage_class_parameters[sc_parameter] = values[sc_parameter]

    additional_sc_options = {'allow_volume_expansion': None, 'allowed_topologies': None,
                             'mount_options': None, 'volume_binding_mode': None}
    for additional_option in list(additional_sc_options):
        if additional_option in values:
            additional_sc_options[additional_option] = values[additional_option]

    storage_class_body = client.V1StorageClass(
        api_version="storage.k8s.io/v1",
        kind="StorageClass",
        metadata=storage_class_metadata,
        provisioner="spectrumscale.csi.ibm.com",
        parameters=storage_class_parameters,
        reclaim_policy="Delete",
        allow_volume_expansion=additional_sc_options['allow_volume_expansion'],
        allowed_topologies=additional_sc_options['allowed_topologies'],
        mount_options=additional_sc_options['mount_options'],
        volume_binding_mode=additional_sc_options['volume_binding_mode']
    )
    try:
        LOGGER.info(
            f'SC Create : creating storageclass {sc_name} with parameters {str(storage_class_parameters)}')
        api_response = api_instance.create_storage_class(
            body=storage_class_body, pretty=True)
        LOGGER.debug(str(api_response))
        created_objects["sc"].append(sc_name)
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling StorageV1Api->create_storage_class: {e}")
        clean_with_created_objects(created_objects, condition="failed")
        assert False


def check_storage_class(sc_name):
    """
    Checks storage class exists or not

    Args:
       param1: sc_name - name of storage class to be checked

    Returns:
       return True  , if storage class exists
       return False , if storage class does not exists

    Raises:
        None

    """
    if sc_name == "":
        return False
    api_instance = client.StorageV1Api()
    try:
        # TBD: Show StorageClass Parameter in tabular Form
        api_response = api_instance.read_storage_class(
            name=sc_name, pretty=True)
        LOGGER.info(f'SC Check : Storage class {sc_name} does exists on the cluster')
        LOGGER.debug(str(api_response))
        return True
    except ApiException:
        LOGGER.info("strorage class does not exists")
        return False


def create_pv(pv_values, pv_name, created_objects, sc_name=""):
    """
    creates persistent volume

    Args:
        param1: pv_values - values required for creation of pv
        param2: pv_name - name of pv to be created
        param3: sc_name - name of storage class pv associated with

    Returns:
        None

    Raises:
        Raises an exception on kubernetes client api failure and asserts

    """
    api_instance = client.CoreV1Api()
    pv_metadata = client.V1ObjectMeta(name=pv_name)
    pv_csi = client.V1CSIPersistentVolumeSource(
        driver="spectrumscale.csi.ibm.com",
        volume_handle=pv_values["volumeHandle"]
    )
    if pv_values["reclaim_policy"] == "Default":
        pv_spec = client.V1PersistentVolumeSpec(
            access_modes=[pv_values["access_modes"]],
            capacity={"storage": pv_values["storage"]},
            csi=pv_csi,
            storage_class_name=sc_name
        )
    else:
        pv_spec = client.V1PersistentVolumeSpec(
            access_modes=[pv_values["access_modes"]],
            capacity={"storage": pv_values["storage"]},
            csi=pv_csi,
            persistent_volume_reclaim_policy=pv_values["reclaim_policy"],
            storage_class_name=sc_name
        )

    pv_body = client.V1PersistentVolume(
        api_version="v1",
        kind="PersistentVolume",
        metadata=pv_metadata,
        spec=pv_spec
    )
    try:
        LOGGER.info(f'PV Create : Creating PV {pv_name} with {pv_values} parameter')
        api_response = api_instance.create_persistent_volume(
            body=pv_body, pretty=True)
        LOGGER.debug(str(api_response))
        created_objects["pv"].append(pv_name)
    except ApiException as e:
        LOGGER.error(f'PV {pv_name} creation failed hence failing test case ')
        LOGGER.error(
            f"Exception when calling CoreV1Api->create_persistent_volume: {e}")
        clean_with_created_objects(created_objects, condition="failed")
        assert False


def check_pv(pv_name):
    """
    Checks pv exists or not

    Args:
       param1: pv_name - name of persistent volume to be checked

    Returns:
       return True  , if pv exists
       return False , if pv does not exists

    Raises:
        None

    """
    api_instance = client.CoreV1Api()
    try:
        api_response = api_instance.read_persistent_volume(
            name=pv_name, pretty=True)
        LOGGER.debug(str(api_response))
        LOGGER.info(f'PV Check : Created PV {pv_name} exists on cluster')
        return True
    except ApiException:
        LOGGER.info(f'PV {pv_name} does not exists on cluster')
        return False


def create_pvc(pvc_values, sc_name, pvc_name, created_objects, pv_name=None):
    """
    creates persistent volume claim

    Args:
        param1: pvc_values - values required for creation of pvc
        param2: sc_name - name of storage class , pvc associated with
                          if "notusingsc" no storage class
        param3: pvc_name - name of pvc to be created
        param4: pv_name - name of pv , pvc associated with
                          if None , no pv is associated

    Returns:
        None

    Raises:
        Raises an exception on kubernetes client api failure and asserts

    """
    api_instance = client.CoreV1Api()
    pvc_metadata = client.V1ObjectMeta(name=pvc_name)
    pvc_resources = client.V1ResourceRequirements(
        requests={"storage": pvc_values["storage"]})

    pvc_spec = client.V1PersistentVolumeClaimSpec(
        access_modes=[pvc_values["access_modes"]],
        resources=pvc_resources,
        storage_class_name=sc_name,
        volume_name=pv_name
    )

    pvc_body = client.V1PersistentVolumeClaim(
        api_version="v1",
        kind="PersistentVolumeClaim",
        metadata=pvc_metadata,
        spec=pvc_spec
    )

    try:
        LOGGER.info(
            f'PVC Create : Creating pvc {pvc_name} with parameters {str(pvc_values)} and storageclass {str(sc_name)}')
        api_response = api_instance.create_namespaced_persistent_volume_claim(
            namespace=namespace_value, body=pvc_body, pretty=True)
        LOGGER.debug(str(api_response))
        created_objects["pvc"].append(pvc_name)
    except ApiException as e:
        LOGGER.info(f'PVC {pvc_name} creation operation has been failed')
        LOGGER.error(
            f"Exception when calling CoreV1Api->create_namespaced_persistent_volume_claim: {e}")
        clean_with_created_objects(created_objects, condition="failed")
        assert False


def create_pvc_from_snapshot(pvc_values, sc_name, pvc_name, snap_name, created_objects):
    """
    creates persistent volume claim from snapshot

    Args:
        param1: pvc_values - values required for creation of pvc
        param2: sc_name - name of storage class , pvc associated with
        param3: pvc_name - name of pvc to be created
        param4: snap_name - name of snapshot to recover data from

    Returns:
        None

    Raises:
        Raises an exception on kubernetes client api failure and asserts

    """
    api_instance = client.CoreV1Api()
    pvc_metadata = client.V1ObjectMeta(name=pvc_name)
    pvc_resources = client.V1ResourceRequirements(
        requests={"storage": pvc_values["storage"]})
    pvc_data_source = client.V1TypedLocalObjectReference(
        api_group="snapshot.storage.k8s.io",
        kind="VolumeSnapshot",
        name=snap_name
    )

    pvc_spec = client.V1PersistentVolumeClaimSpec(
        access_modes=[pvc_values["access_modes"]],
        resources=pvc_resources,
        storage_class_name=sc_name,
        data_source=pvc_data_source
    )

    pvc_body = client.V1PersistentVolumeClaim(
        api_version="v1",
        kind="PersistentVolumeClaim",
        metadata=pvc_metadata,
        spec=pvc_spec
    )

    try:
        LOGGER.info(
            f'PVC Create from snapshot : Creating pvc {pvc_name} with parameters {str(pvc_values)} and storageclass {str(sc_name)}')
        api_response = api_instance.create_namespaced_persistent_volume_claim(
            namespace=namespace_value, body=pvc_body, pretty=True)
        LOGGER.debug(str(api_response))
        created_objects["restore_pvc"].append(pvc_name)
    except ApiException as e:
        LOGGER.info(f'PVC {pvc_name} creation operation has been failed')
        LOGGER.error(
            f"Exception when calling CoreV1Api->create_namespaced_persistent_volume_claim: {e}")
        clean_with_created_objects(created_objects, condition="failed")
        assert False


def create_clone_pvc(pvc_values, sc_name, pvc_name, from_pvc_name, created_objects):
    api_instance = client.CoreV1Api()
    pvc_metadata = client.V1ObjectMeta(name=pvc_name)
    pvc_resources = client.V1ResourceRequirements(
        requests={"storage": pvc_values["storage"]})
    pvc_data_source = client.V1TypedLocalObjectReference(
        kind="PersistentVolumeClaim",
        name=from_pvc_name
    )

    pvc_spec = client.V1PersistentVolumeClaimSpec(
        access_modes=[pvc_values["access_modes"]],
        resources=pvc_resources,
        storage_class_name=sc_name,
        data_source=pvc_data_source
    )

    pvc_body = client.V1PersistentVolumeClaim(
        api_version="v1",
        kind="PersistentVolumeClaim",
        metadata=pvc_metadata,
        spec=pvc_spec
    )

    try:
        LOGGER.info(
            f'PVC Create from Clone : Creating pvc {pvc_name} with parameters {str(pvc_values)} and storageclass {str(sc_name)} from PVC {from_pvc_name}')
        api_response = api_instance.create_namespaced_persistent_volume_claim(
            namespace=namespace_value, body=pvc_body, pretty=True)
        LOGGER.debug(str(api_response))
        created_objects["clone_pvc"].append(pvc_name)
    except ApiException as e:
        LOGGER.info(f'PVC {pvc_name} creation operation has been failed')
        LOGGER.error(
            f"Exception when calling CoreV1Api->create_namespaced_persistent_volume_claim: {e}")
        clean_with_created_objects(created_objects, condition="failed")
        assert False


def pvc_bound_fileset_check(api_response, pv_name, pvc_name, pvc_values, created_objects):
    """
    calculates bound time for pvc and checks fileset created by
    pvc on IBM Storage Scale
    """
    now1 = api_response.metadata.creation_timestamp
    now = datetime.now(timezone.utc)
    timediff = now-now1
    minutes = divmod(timediff.seconds, 60)
    msg = 'Time for PVC BOUND : ' + \
        str(minutes[0]) + 'minutes', str(minutes[1]) + 'seconds'
    LOGGER.info(f'PVC Check : {pvc_name} is BOUND succesfully {msg}')
    volume_name = api_response.spec.volume_name
    if volume_name == pv_name:
        LOGGER.info(f"PVC Check : It is case of static pvc , pv {pv_name} is already created")
        return True
    if 'storage_class_parameters' in globals():
        if "volDirBasePath" in storage_class_parameters and "volBackendFs" in storage_class_parameters:
            return True
        if "version" in storage_class_parameters and storage_class_parameters["version"] == "2":
            cg_fileset_name = get_cg_filesetname_from_pv(volume_name, created_objects)
            if not(filesetfunc.created_fileset_exists(cg_fileset_name)):
                LOGGER.error(
                    f'PVC Check : Fileset {cg_fileset_name} doesn\'t exists for version=2 SC')
                return False
            LOGGER.info(
                f'PVC Check : Fileset {cg_fileset_name} has been created successfully for version=2 SC')
            if cg_fileset_name not in created_objects["cg"]:
                created_objects["cg"].append(cg_fileset_name)

    fileset_name = get_filesetname_from_pv(volume_name, created_objects)
    if not(filesetfunc.created_fileset_exists(fileset_name)):
        LOGGER.error(f'PVC Check : Fileset {fileset_name} doesn\'t exists')
        return False

    if not(check_pvc_size(pvc_name, pvc_values["storage"])):
        LOGGER.error(f'PVC Check : PVC {pvc_name} storage does not match storage in PVC status')
        return False

    inode = None
    fileset_append_check = ""
    if 'storage_class_parameters' in globals():
        if "inodeLimit" in storage_class_parameters:
            inode = storage_class_parameters["inodeLimit"]
        elif "filesetType" in storage_class_parameters and storage_class_parameters["filesetType"] == "dependent":
            inode = 0
        if "version" in storage_class_parameters and storage_class_parameters["version"] == "2":
            inode = 0

        if "compression" in storage_class_parameters:
            if storage_class_parameters["compression"] == "true":
                comp = "Z"
            else:
                comp = storage_class_parameters["compression"].upper()
            fileset_append_check = f"{fileset_append_check}-COMPRESS{comp}csi"
            if storage_class_parameters["compression"] == "false":
                fileset_append_check = ""
        if "tier" in storage_class_parameters:
            fileset_append_check = f"{fileset_append_check}-T{storage_class_parameters['tier']}csi"

    if not(filesetfunc.check_fileset_quota(fileset_name, pvc_values["storage"], inode)):
        LOGGER.error(
            f'PVC Check : Fileset {fileset_name} quota does not match requested storage or maxinode is not as expected')
        return False

    LOGGER.info(f'PVC Check : Fileset {fileset_name} has been created successfully')

    if fileset_append_check != "":
        search_result = re.search(fileset_append_check, fileset_name)
        if search_result is None:
            LOGGER.error(
                f"PVC Check : {fileset_append_check} is not matched for fileset name {fileset_name}")
            return False
        LOGGER.info(
            f"PVC Check : For compression and/or tier in {storage_class_parameters}, Fileset name {fileset_name} is appended with correct value {fileset_append_check}")

    return True


def check_pvc_size(pvc_name, expected_size):
    """
    Check PVC size in status matches passed PVC size or not"
    """
    api_instance = client.CoreV1Api()
    count = 12
    while (count > 0):
        try:
            api_response = api_instance.read_namespaced_persistent_volume_claim(
                name=pvc_name, namespace=namespace_value, pretty=True)
            LOGGER.debug(str(api_response))
            LOGGER.info(f'PVC Check: Checking size for pvc {pvc_name}')

            if "storage" in api_response.status.capacity:
                pvc_status_storage = api_response.status.capacity["storage"]

                if pvc_status_storage == expected_size:
                    return True

                power_of_10 = {"M": int(1000**2 / 1024), "G": int(1000 **
                                                                  3 / 1024), "T": int(1000**4 / 1024)}
                power_of_2 = {"Ki": 1, "Mi": int(1024), "Gi": int(1024**2), "Ti": int(1024**3)}

                expected_size_in_Ki = 0
                if expected_size[-1:] in power_of_10:
                    expected_size_in_Ki = int(expected_size[:-1]) * power_of_10[expected_size[-1:]]
                if expected_size[-2:] in power_of_2:
                    expected_size_in_Ki = int(expected_size[:-2]) * power_of_2[expected_size[-2:]]
                if expected_size_in_Ki < int(1024**2):
                    expected_size_in_Ki = int(1024**2)
                if pvc_status_storage[-2:] in power_of_2:
                    pvc_status_storage = int(
                        pvc_status_storage[:-2]) * power_of_2[pvc_status_storage[-2:]]

                if pvc_status_storage >= expected_size_in_Ki:
                    return True
            count -= 1
        except ApiException as e:
            count -= 1

    return False


def check_pvc(pvc_values,  pvc_name, created_objects, pv_name="pvnotavailable"):
    """ checks pvc is BOUND or not
        need to reduce complextity of this function
    """
    api_instance = client.CoreV1Api()
    con = True
    var = 0
    while (con is True):
        try:
            api_response = api_instance.read_namespaced_persistent_volume_claim(
                name=pvc_name, namespace=namespace_value, pretty=True)
            LOGGER.debug(str(api_response))
            LOGGER.info(f'PVC Check: Checking for pvc {pvc_name}')
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_persistent_volume_claim: {e}")
            LOGGER.info(f"PVC Check : PVC {pvc_name} does not exists on the cluster")
            clean_with_created_objects(created_objects, condition="failed")
            assert False

        if api_response.status.phase == "Bound":
            if "reason" in pvc_values:
                LOGGER.error(f'PVC Check : {pvc_name} is BOUND but as the failure reason is provided so\
                asserting the test')
                clean_with_created_objects(created_objects, condition="failed")
                assert False
            if(pvc_bound_fileset_check(api_response, pv_name, pvc_name, pvc_values, created_objects)):
                return True
            clean_with_created_objects(created_objects, condition="failed")
            assert False
        else:
            var += 1
            time.sleep(5)
            if "reason" in pvc_values:
                time_count = 8
            elif "parallel" in pvc_values:
                time_count = 240
            elif "clone" in pvc_values:
                time_count = 120
            else:
                time_count = 20
            if(var > time_count):
                LOGGER.info("PVC Check : PVC is not BOUND,checking if failure reason is expected")
                field = "involvedObject.name="+pvc_name
                reason = api_instance.list_namespaced_event(
                    namespace=namespace_value, pretty=True, field_selector=field)
                if "reason" not in pvc_values:
                    clean_with_created_objects(created_objects, condition="failed")
                    LOGGER.error(str(reason))
                    LOGGER.error(
                        "FAILED as reason for Failure not provides")
                    assert False
                search_result = None
                for item in reason.items:
                    search_result = re.search(
                        pvc_values["reason"], str(item.message))
                    if search_result is not None:
                        break
                if search_result is None:
                    clean_with_created_objects(created_objects, condition="failed")
                    LOGGER.error(f"Failed reason : {str(reason)}")
                    LOGGER.error("PVC Check : PVC is not Bound but FAILED reason does not match")
                    assert False
                else:
                    LOGGER.debug(search_result)
                    LOGGER.info(
                        f"PVC Check : PVC is not Bound and FAILED with expected error {pvc_values['reason']}")
                    con = False


def create_pod(value_pod, pvc_name, pod_name, created_objects, image_name="nginx:1.19.0"):
    """
    creates pod
    Args:
        param1: value_pod - values required for creation of pod
        param2: pvc_name - name of pvc , pod associated with
        param3: pod_name - name of pod to be created
        param4: image_name - name of the pod image (Default:"nginx:1.19.0")
        param5: run_as_user - value for pod securityContext spec runAsUser
        param6: run_as_group - value for pod securityContext spec runAsGroup
    Returns:
        None
    Raises:
        Raises an exception on kubernetes client api failure and asserts
    """
    if value_pod["read_only"] == "True":
        value_pod["read_only"] = True
    elif value_pod["read_only"] == "False":
        value_pod["read_only"] = False
    api_instance = client.CoreV1Api()
    pod_metadata = client.V1ObjectMeta(name=pod_name, labels={"app": "nginx"})

    pod_persistent_volume_claim = client.V1PersistentVolumeClaimVolumeSource(
        claim_name=pvc_name, read_only=value_pod["read_only"])
    pod_volumes = client.V1Volume(
        name="mypvc", persistent_volume_claim=pod_persistent_volume_claim)

    pod_ports = client.V1ContainerPort(container_port=80)

    if "sub_path" not in value_pod:
        pod_volume_mounts = client.V1VolumeMount(
            name="mypvc", mount_path=value_pod["mount_path"])
        command = ["/bin/sh", "-c", "--"]
        args = ["while true; do sleep 30; done;"]
        pod_containers = client.V1Container(
            name="web-server", image=image_name, volume_mounts=[pod_volume_mounts], ports=[pod_ports], command=command, args=args)
    else:
        list_pod_volume_mount = []
        for iter_num, single_sub_path in enumerate(value_pod["sub_path"]):
            final_mount_path = value_pod["mount_path"] if iter_num == 0 else value_pod["mount_path"]+str(
                iter_num)
            list_pod_volume_mount.append(client.V1VolumeMount(
                name="mypvc", mount_path=final_mount_path, sub_path=single_sub_path, read_only=value_pod["volumemount_readonly"][iter_num]))
        command = ["/bin/sh", "-c", "--"]
        args = ["while true; do sleep 30; done;"]
        pod_containers = client.V1Container(
            name="web-server", image=image_name, volume_mounts=list_pod_volume_mount, ports=[pod_ports],
            command=command, args=args)

    if "fsgroup" in value_pod or "runAsGroup" in value_pod and "runAsUser" in value_pod:
        if "runAsGroup" in value_pod and "runAsUser" in value_pod and "fsgroup" in value_pod and "runasnonroot" in value_pod:
            pod_security_context = client.V1PodSecurityContext(
                run_as_group=int(value_pod["runAsGroup"]), run_as_user=int(value_pod["runAsUser"]),
                fs_group=int(value_pod["fsgroup"]), run_as_non_root=value_pod["runasnonroot"])
        elif "runAsGroup" in value_pod and "runAsUser" in value_pod:
            pod_security_context = client.V1PodSecurityContext(
                run_as_group=int(value_pod["runAsGroup"]), run_as_user=int(value_pod["runAsUser"]))
        elif "fsgroup" in value_pod:
            pod_security_context = client.V1PodSecurityContext(fs_group=int(value_pod["fsgroup"]))
        pod_spec = client.V1PodSpec(
            containers=[pod_containers], volumes=[pod_volumes], node_selector=nodeselector, security_context=pod_security_context)
    else:
        pod_spec = client.V1PodSpec(
            containers=[pod_containers], volumes=[pod_volumes], node_selector=nodeselector)

    pod_body = client.V1Pod(
        api_version="v1",
        kind="Pod",
        metadata=pod_metadata,
        spec=pod_spec
    )

    try:
        LOGGER.info(
            f'POD Create : creating pod {pod_name} using {pvc_name} with {image_name} image with parameters {value_pod}')
        api_response = api_instance.create_namespaced_pod(
            namespace=namespace_value, body=pod_body, pretty=True)
        LOGGER.debug(str(api_response))
        if pod_name[0:12] == "snap-end-pod":
            created_objects["restore_pod"].append(pod_name)
        elif pod_name[0:5] == "clone":
            created_objects["clone_pod"].append(pod_name)
        else:
            created_objects["pod"].append(pod_name)
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CoreV1Api->create_namespaced_pod: {e}")
        clean_with_created_objects(created_objects, condition="failed")
        assert False


def create_file_inside_pod(value_pod, pod_name, created_objects):
    """
    create snaptestfile inside the pod using touch
    """
    api_instance = client.CoreV1Api()
    LOGGER.info("POD Check : Trying to create snaptestfile on StorageScale mount point inside the pod")
    exec_command1 = "touch "+value_pod["mount_path"]+"/snaptestfile"
    exec_command = [
        '/bin/sh',
        '-c',
        exec_command1]
    resp = stream(api_instance.connect_get_namespaced_pod_exec,
                  pod_name,
                  namespace_value,
                  command=exec_command,
                  stderr=True, stdin=False,
                  stdout=True, tty=False)

    if resp == "":
        LOGGER.info("file snaptestfile created successfully on StorageScale mount point inside the pod")
        return

    if "reason" in value_pod:
        LOGGER.warning(f"Cannot write data in pod due to {value_pod}")
        return

    LOGGER.error("file snaptestfile not created")
    LOGGER.error(resp)
    clean_with_created_objects(created_objects, condition="failed")
    assert False


def check_file_inside_pod(value_pod, pod_name, created_objects, volume_name=None):
    """
    check snaptestfile inside the pod using ls
    """
    api_instance = client.CoreV1Api()
    if volume_name is None:
        exec_command1 = "ls "+value_pod["mount_path"]
    else:
        exec_command1 = "ls "+value_pod["mount_path"]+"/"+volume_name+"-data"
    exec_command = [
        '/bin/sh',
        '-c',
        exec_command1]
    resp = stream(api_instance.connect_get_namespaced_pod_exec,
                  pod_name,
                  namespace_value,
                  command=exec_command,
                  stderr=True, stdin=False,
                  stdout=True, tty=False)
    if resp[0:12] == "snaptestfile":
        LOGGER.info("POD Check : snaptestfile is succesfully restored from snapshot or clone")
        return

    if "reason" in value_pod:
        LOGGER.warning(
            f"As snaptestfile cannot be written in pod due to {value_pod}, snaptestfile is not restored")
        return

    LOGGER.error("snaptestfile is not restored from snapshot or clone")
    clean_with_created_objects(created_objects, condition="failed")
    assert False


def check_pod_execution(value_pod, pod_name, created_objects):
    """
    checks can file be created in pod
    if file cannot be created , checks reason , if reason does not mathch , asserts
    Args:
        param1: value_pod - values required for creation of pod
        param2: sc_name - name of storage class , pod associated with
        param3: pvc_name - name of pvc , pod associated with
        param4: pod_name - name of pod to be checked
        param5: dir_name - directory associated with pod
        param6: pv_name - name of pv associated with pod

    Returns:
        None

    Raises:
        None
    """
    api_instance = client.CoreV1Api()
    LOGGER.info("POD Check : Trying to create testfile on StorageScale mount point inside the pod")
    if "fsgroup" in value_pod:
        exec_command1 = "id"
        exec_command = [
            '/bin/sh',
            '-c',
            exec_command1]
        resp = stream(api_instance.connect_get_namespaced_pod_exec,
                      pod_name,
                      namespace_value,
                      command=exec_command,
                      stderr=True, stdin=False,
                      stdout=True, tty=False)
        fsgroup_id = resp.split(",")
        fsgroup_id = fsgroup_id[1].strip()
        if value_pod['fsgroup'] == fsgroup_id:
            LOGGER.info(f"fsGroup ID is {fsgroup_id}")
        else:
            LOGGER.error(f"fsGroup IDs are not matching")
            LOGGER.info(f"Given fsGroup ID : {value_pod['fsgroup']}")
            LOGGER.info(f"Received fsGroup ID : {fsgroup_id}")
            assert False
    exec_command1 = "touch "+value_pod["mount_path"]+"/testfile"
    exec_command = [
        '/bin/sh',
        '-c',
        exec_command1]
    resp = stream(api_instance.connect_get_namespaced_pod_exec,
                  pod_name,
                  namespace_value,
                  command=exec_command,
                  stderr=True, stdin=False,
                  stdout=True, tty=False)
    if resp == "":
        LOGGER.info("POD Check : Create testfile operation completed successfully")
        LOGGER.info("POD Check : Deleting testfile from pod's StorageScale mount point")
        exec_command1 = "rm -rvf "+value_pod["mount_path"]+"/testfile"
        exec_command = [
            '/bin/sh',
            '-c',
            exec_command1]
        resp = stream(api_instance.connect_get_namespaced_pod_exec,
                      pod_name,
                      namespace_value,
                      command=exec_command,
                      stderr=True, stdin=False,
                      stdout=True, tty=False)
        if "reason" in value_pod:
            clean_with_created_objects(created_objects, condition="failed")
            LOGGER.error(
                "Pod should not be able to create file inside the pod as failure REASON provided, so asserting")
            assert False
        return
    if "reason" not in value_pod:
        clean_with_created_objects(created_objects, condition="failed")
        LOGGER.error(str(resp))
        LOGGER.error("FAILED as reason of failure not provided")
        assert False
    search_result1 = re.search(value_pod["reason"], str(resp))
    search_result2 = re.search("Permission denied", str(resp))
    if search_result1 is not None:
        LOGGER.info(str(search_result1))
    if search_result2 is not None:
        LOGGER.info(str(search_result2))
    if not(search_result1 is None and search_result2 is None):
        LOGGER.info("execution of pod failed with expected reason")
    else:
        clean_with_created_objects(created_objects, condition="failed")
        LOGGER.error(str(resp))
        LOGGER.error(
            "execution of pod failed unexpected , reason does not match")
        assert False


def check_pod(value_pod, pod_name, created_objects):
    """
    checks pod running or not

    Args:
        param1: value_pod - values required for creation of pod
        param2: sc_name - name of storage class , pod associated with
        param3: pvc_name - name of pvc , pod associated with
        param4: pod_name - name of pod to be checked
        param5: dir_name - directory associated with pod
        param6: pv_name - name of pv associated with pod

    Returns:
        None

    Raises:
        None
    """
    api_instance = client.CoreV1Api()
    con = True
    var = 0
    while (con is True):
        try:
            api_response = api_instance.read_namespaced_pod(
                name=pod_name, namespace=namespace_value, pretty=True)
            LOGGER.debug(str(api_response))
            LOGGER.info(f'POD Check: Checking for pod {pod_name}')
            if api_response.status.phase == "Running":
                LOGGER.info(f'POD Check : POD {pod_name} is Running')
                check_pod_execution(value_pod, pod_name, created_objects)
                con = False
            else:
                var += 1
                if(var > 20):
                    LOGGER.error(f'POD Check : POD {pod_name} is not running')
                    field = "involvedObject.name="+pod_name
                    reason = api_instance.list_namespaced_event(
                        namespace=namespace_value, pretty=True, field_selector=field)
                    if "reason" not in value_pod:
                        LOGGER.error('FAILED as reason of failure not provided')
                        LOGGER.error(f"POD Check : Reason of failure is : {str(reason)}")
                        clean_with_created_objects(created_objects, condition="failed")
                        assert False
                    search_result = re.search(value_pod["reason"], str(reason))
                    if search_result is None:
                        LOGGER.error(
                            f'Failed as reason of failure does not match {value_pod["reason"]}')
                        LOGGER.error(f"POD Check : Reason of failure is : {str(reason)}")
                        clean_with_created_objects(created_objects, condition="failed")
                        assert False
                    else:
                        LOGGER.info(f'POD failed with expected reason {value_pod["reason"]}')
                        return
                time.sleep(5)
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod: {e}")
            LOGGER.error("POD Check : POD does not exists on Cluster")
            clean_with_created_objects(created_objects, condition="failed")
            assert False


def create_ds(ds_values, ds_name, pvc_name, created_objects):

    api_instance = client.AppsV1Api()
    if ds_values["read_only"] == "True":
        ds_values["read_only"] = True
    elif ds_values["read_only"] == "False":
        ds_values["read_only"] = False

    ds_body = {
        "apiVersion": "apps/v1",
        "kind": "DaemonSet",
        "metadata": {
            "name": ds_name,
            "labels": {
                "app": "nginx",
                "ownerReferences": ds_name
            }
        },
        "spec": {
            "selector": {
                "matchLabels": {
                    "name": "nginx",
                    "ownerReferences": ds_name
                }
            },
            "template": {
                "metadata": {
                    "labels": {
                        "name": "nginx",
                        "ownerReferences": ds_name
                    }
                },
                "spec": {
                    "containers": [
                        {
                            "name": "web-server",
                            "image": "nginxinc/nginx-unprivileged",
                            "volumeMounts": [
                                {
                                    "name": "mypvc",
                                    "mountPath": ds_values["mount_path"]
                                }
                            ]
                        }
                    ],
                    "volumes": [
                        {
                            "name": "mypvc",
                            "persistentVolumeClaim": {
                                "claimName": pvc_name,
                                "readOnly": ds_values["read_only"]
                            }
                        }
                    ],
                    "nodeSelector": nodeselector
                }
            }
        }
    }

    try:
        LOGGER.info(
            f'Daemonset Create : Creating daemonset {ds_name} with parameters {str(ds_values)} and pvc {str(pvc_name)}')
        api_response = api_instance.create_namespaced_daemon_set(
            namespace=namespace_value, body=ds_body, pretty=True)
        LOGGER.debug(str(api_response))
        created_objects["ds"].append(ds_name)
    except ApiException as e:
        LOGGER.info(f'Daemonset Create : Daemonset {ds_name} creation operation has been failed')
        LOGGER.error(
            f"Exception when calling AppsV1Api->create_namespaced_daemon_set: {e}")
        clean_with_created_objects(created_objects, condition="failed")
        assert False


def check_ds(ds_name, value_ds, created_objects):
    read_daemonsets_api_instance = client.AppsV1Api()
    num = 0
    while (num < 11):
        try:
            read_daemonsets_api_response = read_daemonsets_api_instance.read_namespaced_daemon_set(
                name=ds_name, namespace=namespace_value, pretty=True)
            LOGGER.debug(read_daemonsets_api_response)
            LOGGER.info(f"Daemonset Check : Checking for daemonset {ds_name}")
            current_number_scheduled = read_daemonsets_api_response.status.current_number_scheduled
            desired_number_scheduled = read_daemonsets_api_response.status.desired_number_scheduled
            number_available = read_daemonsets_api_response.status.number_available

            if number_available == current_number_scheduled == desired_number_scheduled:
                if desired_number_scheduled < 2:
                    LOGGER.error(
                        f"Not enough nodes for this test, only {desired_number_scheduled} nodes are there")
                    clean_with_created_objects(created_objects, condition="failed")
                    assert False

                if "reason" in value_ds:
                    LOGGER.error(
                        f"failure reason provided  {value_ds} , still all pods are running")
                    clean_with_created_objects(created_objects, condition="failed")
                    assert False

                LOGGER.info(
                    f"Daemonset Check : daemonset {ds_name} all {current_number_scheduled} pods are Running")

                return

            time.sleep(20)
            num += 1
            LOGGER.info(f"Daemonset Check : waiting for daemonsets {ds_name}")
        except ApiException:
            time.sleep(20)
            num += 1
            LOGGER.info(f"Daemonset Check : waiting for daemonsets {ds_name}")

    if "reason" not in value_ds:
        LOGGER.error(
            f"Daemonset Check : daemonset {ds_name} {number_available}/{desired_number_scheduled} pods are Running, asserting")
        clean_with_created_objects(created_objects, condition="failed")
        assert False

    if desired_number_scheduled < 2:
        LOGGER.error(
            f"Not enough nodes for this test, only {desired_number_scheduled} nodes are there")
        clean_with_created_objects(created_objects, condition="failed")
        assert False

    if check_ds_pod(ds_name, value_ds, created_objects):
        LOGGER.info(
            f"Daemonset Check : daemonset {ds_name} pods failed with expected reason {value_ds['reason']}")
        return

    LOGGER.info(
        f"Daemonset Check : daemonset {ds_name} pods did not fail with expected reason {value_ds['reason']}")
    clean_with_created_objects(created_objects, condition="failed")
    assert False


def check_ds_pod(ds_name, value_ds, created_objects):
    api_instance = client.CoreV1Api()
    selector = "ownerReferences="+ds_name
    running_pod_list, pod_list = [], []
    try:
        api_response = api_instance.list_namespaced_pod(
            namespace=namespace_value, pretty=True, label_selector=selector)
        LOGGER.debug(api_response)
        for pod in api_response.items:
            if pod.status.phase == "Running":
                running_pod_list.append(pod.metadata.name)
            else:
                pod_list.append(pod.metadata.name)
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CoreV1Api->list_node: {e}")
        clean_with_created_objects(created_objects, condition="failed")
        assert False

    if len(running_pod_list) != 1:
        LOGGER.error(
            f"running pods are {running_pod_list} , only one pod should be running, asserting")
        return False

    for pod_name in pod_list:
        check_pod(value_ds, pod_name, created_objects)

    return True


def create_deployment_object():

    deployment_apps_api_instance = client.AppsV1Api()

    deployment_labels = {
        "product": "ibm-spectrum-scale-csi-test",
        "app": "nginx"
    }

    deployment_metadata = client.V1ObjectMeta(
        name="nginx-deployment", labels=deployment_labels, namespace=namespace_value)

    deployment_selector = client.V1LabelSelector(match_labels={"app": "nginx"})

    podtemplate_metadata = client.V1ObjectMeta(labels=deployment_labels)

    nginx_pod_container = client.V1Container(image="nginx:1.14.2", name="nginx")

    pod_spec = client.V1PodSpec(containers=[nginx_pod_container])

    podtemplate_spec = client.V1PodTemplateSpec(
        metadata=podtemplate_metadata, spec=pod_spec)

    deployment_spec = client.V1DeploymentSpec(
        replicas=3, selector=deployment_selector, template=podtemplate_spec)

    body_dep = client.V1Deployment(
        kind='Deployment', api_version='apps/v1', metadata=deployment_metadata, spec=deployment_spec)

    try:
        LOGGER.info("creating deployment for nginx")
        deployment_apps_api_response = deployment_apps_api_instance.create_namespaced_deployment(
            namespace=namespace_value, body=body_dep)
        LOGGER.debug(str(deployment_apps_api_response))
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling RbacAuthorizationV1Api->create_namespaced_deployment: {e}")
        assert False


def get_pv_for_pvc(pvc_name, created_objects):
    """ 
    return pv name associated with pvc
    """
    api_instance = client.CoreV1Api()
    try:
        api_response = api_instance.read_namespaced_persistent_volume_claim(
            name=pvc_name, namespace=namespace_value, pretty=True)
        LOGGER.debug(str(api_response))
        LOGGER.info(f'PVC Check: Checking for pvc {pvc_name}')
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CoreV1Api->read_namespaced_persistent_volume_claim: {e}")
        LOGGER.info(f"PVC Check : PVC {pvc_name} does not exists on the cluster")
        clean_with_created_objects(created_objects, condition="failed")
        assert False

    return api_response.spec.volume_name


def check_permissions_for_pvc(pvc_name, storage_class_parameters, created_objects):
    """
    get pv and verify permissions for pv
    """
    if "permissions" not in storage_class_parameters.keys() or "shared" not in storage_class_parameters.keys():
        return

    if "permissions" in storage_class_parameters.keys():
        permissions = storage_class_parameters["permissions"]
    elif storage_class_parameters["shared"]=="True":
        permissions = "777"
    else:
        return 

    pv_name = get_pv_for_pvc(pvc_name, created_objects)
    fileset_name = get_filesetname_from_pv(pv_name, created_objects)
    cg_fileset_name = None
    if "version" in storage_class_parameters and storage_class_parameters["version"] == "2":
        cg_fileset_name = get_cg_filesetname_from_pv(pv_name, created_objects)
    if permissions == "":  # assign default permissions 771
        permissions = "771"
    status = filesetfunc.get_and_verify_fileset_permissions(fileset_name, permissions, cg_fileset_name)
    if status is True:
        LOGGER.info(f'PASS: Testing storageclass parameter permissions={permissions} passed.')
    else:
        LOGGER.info(f'FAIL: Testing storageclass parameter permissions={permissions} failed.')
        clean_with_created_objects(created_objects, condition="failed")
        assert False


def expand_pvc(pvc_values, sc_name, pvc_name, created_objects, pv_name=None):
    """
    expand pvc size
    """
    api_instance = client.CoreV1Api()
    pvc_metadata = client.V1ObjectMeta(name=pvc_name)
    pvc_resources = client.V1ResourceRequirements(
        requests={"storage": pvc_values["storage"]})

    pvc_spec = client.V1PersistentVolumeClaimSpec(
        access_modes=[pvc_values["access_modes"]],
        resources=pvc_resources,
        storage_class_name=sc_name,
        volume_name=pv_name
    )

    pvc_body = client.V1PersistentVolumeClaim(
        api_version="v1",
        kind="PersistentVolumeClaim",
        metadata=pvc_metadata,
        spec=pvc_spec
    )

    LOGGER.info(100*"-")
    try:
        LOGGER.info(
            f'PVC Patch : Patching pvc {pvc_name} with parameters {str(pvc_values)} and storageclass {str(sc_name)}')
        api_response = api_instance.patch_namespaced_persistent_volume_claim(
            name=pvc_name, namespace=namespace_value, body=pvc_body, pretty=True)
        LOGGER.debug(str(api_response))
        time.sleep(30)
    except ApiException as e:
        LOGGER.info(f'PVC {pvc_name} patch operation has been failed')
        LOGGER.error(
            f"Exception when calling CoreV1Api->patch_namespaced_persistent_volume_claim: {e}")
        clean_with_created_objects(created_objects, condition="failed")
        assert False


def expand_and_check_pvc(sc_name, pvc_name, value_pvc, expansion_key, pod_name, value_pod, created_objects):

    for expand_storage in value_pvc[expansion_key]:
        value_pvc['storage'] = expand_storage
        expand_pvc(value_pvc, sc_name, pvc_name, created_objects)
        if(check_pvc(value_pvc, pvc_name, created_objects)):
            check_pod(value_pod, pod_name, created_objects)


def clone_and_check_pvc(sc_name, value_sc, pvc_name, pod_name, value_pod, clone_values, created_objects):

    create_file_inside_pod(value_pod, pod_name, created_objects)
    clone_sc_name = sc_name
    number_of_clones = 1 if "number_of_clones" not in clone_values else int(
        clone_values["number_of_clones"])

    if "clone_sc" in clone_values:
        clone_sc_name = "clone-"+get_random_name("sc")
        create_storage_class(clone_values["clone_sc"], clone_sc_name, created_objects)
        check_storage_class(clone_sc_name)
        value_sc = copy.deepcopy(clone_values["clone_sc"])

    for clone_pvc_number, clone_pvc_value in enumerate(clone_values["clone_pvc"]):
        for iter_clone in range(0, number_of_clones):
            clone_pvc_value["clone"] = "True"
            clone_pvc_name = f"clone-{pvc_name}-{clone_pvc_number}-{iter_clone}"
            create_clone_pvc(clone_pvc_value, clone_sc_name,
                             clone_pvc_name, pvc_name, created_objects)
            val = check_pvc(clone_pvc_value, clone_pvc_name, created_objects)
            if val is True:
                check_permissions_for_pvc(
                        clone_pvc_name, value_sc, created_objects)

                if value_sc.keys() >= {"permissions", "gid", "uid"}:
                    value_pod["runAsGroup"] = value_sc["gid"]
                    value_pod["runAsUser"] = value_sc["uid"]
                clone_pod_name = f"clone-pod-{pvc_name}-{clone_pvc_number}-{iter_clone}"
                create_pod(value_pod, clone_pvc_name, clone_pod_name, created_objects)
                check_pod(value_pod, clone_pod_name, created_objects)
                check_file_inside_pod(value_pod, clone_pod_name, created_objects)

            if "clone_chain" in clone_values and clone_values["clone_chain"] > 0:
                clone_values["clone_chain"] -= 1
                clone_and_check_pvc(clone_sc_name, value_sc, clone_pvc_name,
                                    clone_pod_name, value_pod, clone_values, created_objects)

    for pod_name in copy.deepcopy(created_objects["clone_pod"]):
        delete_pod(pod_name, created_objects)
        check_pod_deleted(pod_name, created_objects)
    for pvc_name in copy.deepcopy(created_objects["clone_pvc"]):
        delete_pvc(pvc_name, created_objects)
        check_pvc_deleted(pvc_name, created_objects)


def create_vs_class(vs_class_name, body_params, created_objects):
    """
    create volume snapshot class with vs_class_name
    body_params contains configurable parameters
    """
    class_body = {
        "apiVersion": "snapshot.storage.k8s.io/v1",
        "kind": "VolumeSnapshotClass",
        "metadata": {
            "name": vs_class_name
        },
        "driver": "spectrumscale.csi.ibm.com",
        "deletionPolicy": body_params["deletionPolicy"]
    }

    if "snapWindow" in body_params:
        class_body["parameters"] = {"snapWindow": body_params["snapWindow"]}

    custom_object_api_instance = client.CustomObjectsApi()
    try:
        custom_object_api_response = custom_object_api_instance.create_cluster_custom_object(
            group="snapshot.storage.k8s.io",
            version="v1",
            plural="volumesnapshotclasses",
            body=class_body,
            pretty=True
        )
        LOGGER.debug(custom_object_api_response)
        LOGGER.info(f"Volume Snapshot Class Create : {vs_class_name} is created with {body_params}")
        created_objects["vsclass"].append(vs_class_name)
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CustomObjectsApi->create_namespaced_custom_object: {e}")
        clean_with_created_objects(created_objects, condition="failed")
        assert False


def check_vs_class(vs_class_name):
    """ 
    checks volume snapshot class vs_class_name exists or not
    return True , if vs_class_name exists
    else return False
    """
    api_instance = client.CustomObjectsApi()
    try:
        api_response = api_instance.get_cluster_custom_object(
            group="snapshot.storage.k8s.io",
            version="v1",
            plural="volumesnapshotclasses",
            name=vs_class_name
        )
        LOGGER.debug(api_response)
        LOGGER.info(f"Volume Snapshot Class Check : {vs_class_name} exists")
        return True
    except ApiException:
        LOGGER.info(f"volume snapshot class {vs_class_name} does not exists")
        return False


def create_vs(vs_name, vs_class_name, pvc_name, created_objects):
    """
    create volume snapshot vs_name using volume snapshot class vs_class_name
    and pvc pvc_name
    """
    class_body = {
        "apiVersion": "snapshot.storage.k8s.io/v1",
        "kind": "VolumeSnapshot",
        "metadata": {
                      "name": vs_name
        },
        "spec": {
            "volumeSnapshotClassName": vs_class_name,
            "source": {
                "persistentVolumeClaimName": pvc_name
            }
        }
    }

    custom_object_api_instance = client.CustomObjectsApi()
    try:
        custom_object_api_response = custom_object_api_instance.create_namespaced_custom_object(
            group="snapshot.storage.k8s.io",
            version="v1",
            plural="volumesnapshots",
            body=class_body,
            namespace=namespace_value,
            pretty=True
        )
        LOGGER.debug(custom_object_api_response)
        LOGGER.info(f"Volume Snapshot Create : volume snapshot {vs_name} is created for {pvc_name}")
        created_objects["vs"].append(vs_name)
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CustomObjectsApi->create_namespaced_custom_object: {e}")
        clean_with_created_objects(created_objects, condition="failed")
        assert False


def create_vs_from_content(vs_name, vs_content_name, created_objects):
    """
    create volume snapshot vs_name from volume snapshot content vs_content_name
    """
    class_body = {
        "apiVersion": "snapshot.storage.k8s.io/v1",
        "kind": "VolumeSnapshot",
        "metadata": {
                      "name": vs_name
        },
        "spec": {
            "source": {
                "volumeSnapshotContentName": vs_content_name
            }
        }
    }

    custom_object_api_instance = client.CustomObjectsApi()
    try:
        custom_object_api_response = custom_object_api_instance.create_namespaced_custom_object(
            group="snapshot.storage.k8s.io",
            version="v1",
            plural="volumesnapshots",
            body=class_body,
            namespace=namespace_value,
            pretty=True
        )
        LOGGER.debug(custom_object_api_response)
        LOGGER.info(
            f"Volume Snapshot Create : volume snapshot {vs_name} is created from {vs_content_name}")
        created_objects["vs"].append(vs_name)
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CustomObjectsApi->create_namespaced_custom_object: {e}")
        clean_with_created_objects(created_objects, condition="failed")
        assert False


def check_vs(vs_name):
    """
    check volume snapshot vs_name exists or not
    return True , if exists
    else return False
    """
    api_instance = client.CustomObjectsApi()
    try:
        api_response = api_instance.get_namespaced_custom_object(
            group="snapshot.storage.k8s.io",
            version="v1",
            plural="volumesnapshots",
            name=vs_name,
            namespace=namespace_value
        )
        LOGGER.debug(api_response)
        LOGGER.info(f"Volume Snapshot Check : volume snapshot {vs_name} has been created")
        return True
    except ApiException:
        LOGGER.info(f"Volume Snapshot Check : volume snapshot {vs_name} does not exists")
        return False


def check_vs_detail_for_static(vs_name, created_objects):
    api_instance = client.CustomObjectsApi()
    try:
        api_response = api_instance.get_namespaced_custom_object(
            group="snapshot.storage.k8s.io",
            version="v1",
            plural="volumesnapshots",
            name=vs_name,
            namespace=namespace_value
        )
        LOGGER.debug(api_response)
        LOGGER.info(f"Volume Snapshot Check : volume snapshot {vs_name} has been created")
    except ApiException:
        LOGGER.info(f"Volume Snapshot Check : volume snapshot {vs_name} does not exists")
        clean_with_created_objects(created_objects, condition="failed")
        assert False

    if check_snapshot_status(vs_name):
        LOGGER.info("volume snapshot status ReadyToUse is true")
    else:
        LOGGER.error("volume snapshot status ReadyToUse is not true")
        clean_with_created_objects(created_objects, condition="failed")
        assert False


def check_vs_detail(vs_name, pvc_name, body_params, reason, created_objects):
    """
    checks volume snapshot vs_name exits , 
    checks volume snapshot content for vs_name is created
    check snapshot is created on IBM Storage Scale
    """
    api_instance = client.CustomObjectsApi()
    try:
        api_response = api_instance.get_namespaced_custom_object(
            group="snapshot.storage.k8s.io",
            version="v1",
            plural="volumesnapshots",
            name=vs_name,
            namespace=namespace_value
        )
        LOGGER.debug(api_response)
        LOGGER.info(f"Volume Snapshot Check : volume snapshot {vs_name} has been created")
    except ApiException:
        LOGGER.info(f"Volume Snapshot Check : volume snapshot {vs_name} does not exists")
        clean_with_created_objects(created_objects, condition="failed")
        assert False

    if check_snapshot_status(vs_name):
        LOGGER.info("volume snapshot status ReadyToUse is true")
    else:
        LOGGER.error("volume snapshot status ReadyToUse is not true")
        api_instance_events = client.CoreV1Api()
        field = "involvedObject.name="+vs_name
        failure_reason = api_instance_events.list_namespaced_event(
            namespace=namespace_value, pretty=True, field_selector=field)
        LOGGER.debug(failure_reason)
        if reason is not None:
            search_result = None
            for item in failure_reason.items:
                search_result = re.search(reason, str(item.message))
                if search_result is not None:
                    LOGGER.info(
                        f"reason {reason} matched in volumesnapshot events, passing the test")
                    return

        LOGGER.error(failure_reason)
        LOGGER.error(f"reason {reason} did not matched in volumesnapshot events")
        clean_with_created_objects(created_objects, condition="failed")
        assert False

    uid_name = api_response["metadata"]["uid"]
    snapcontent_name = "snapcontent-" + uid_name
    time.sleep(2)

    if not(check_vs_content(snapcontent_name)):
        clean_with_created_objects(created_objects, condition="failed")
        assert False

    snapshot_name, fileset_name = get_snapshot_and_related_fileset(
        snapcontent_name, pvc_name, created_objects)

    if filesetfunc.check_snapshot_exists(snapshot_name, fileset_name):
        LOGGER.info(
            f"Snapshot Fileset Check : Snapshot {snapshot_name} exists for Fileset {fileset_name}")
    else:
        LOGGER.error(
            f"Snapshot Fileset Check : Snapshot {snapshot_name} does not exists for Fileset {fileset_name}")
        clean_with_created_objects(created_objects, condition="failed")
        assert False

    if body_params["deletionPolicy"] == "Retain":
        created_objects["vscontent"].append(snapcontent_name)
        created_objects["scalesnapshot"].append([snapshot_name, fileset_name])


def check_snapshot_status(vs_name):
    """
    check status of volume snapshot vs_name
    if status True , return True
    else return False
    """
    api_instance = client.CustomObjectsApi()
    val = 0
    while val < 14:
        try:
            api_response = api_instance.get_namespaced_custom_object_status(
                group="snapshot.storage.k8s.io",
                version="v1",
                plural="volumesnapshots",
                name=vs_name,
                namespace=namespace_value
            )
            LOGGER.debug(api_response)
            LOGGER.info(f"Volume Snapshot Check: Checking for snapshot status of {vs_name}")
            if "status" in api_response.keys() and "readyToUse" in api_response["status"].keys():
                if api_response["status"]["readyToUse"] is True:
                    return True
            time.sleep(15)
            val += 1
        except ApiException:
            time.sleep(15)
            val += 1
    return False


def create_vs_content(vs_content_name, vs_name, body_params, created_objects):
    """
    create volume snapshot content with vs_content_name
    body_params contains configurable parameters
    """
    content_body = {
        "apiVersion": "snapshot.storage.k8s.io/v1",
        "kind": "VolumeSnapshotContent",
        "metadata": {
            "name": vs_content_name
        },
        "spec": {
            "deletionPolicy": body_params["deletionPolicy"],
            "driver": "spectrumscale.csi.ibm.com",
            "source": {
                "snapshotHandle": body_params["snapshotHandle"]
            },
            "volumeSnapshotRef": {
                "name": vs_name,
                "namespace": namespace_value
            }
        }
    }

    custom_object_api_instance = client.CustomObjectsApi()
    try:
        custom_object_api_response = custom_object_api_instance.create_cluster_custom_object(
            group="snapshot.storage.k8s.io",
            version="v1",
            plural="volumesnapshotcontents",
            body=content_body,
            pretty=True
        )
        LOGGER.debug(custom_object_api_response)
        created_objects["vscontent"].append(vs_content_name)
        LOGGER.info(
            f"Volume Snapshot Content Create : {vs_content_name} is created with {body_params}")
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CustomObjectsApi->create_namespaced_custom_object: {e}")
        clean_with_created_objects(created_objects, condition="failed")
        assert False


def check_vs_content(vs_content_name):
    """
    checks volume snapshot content vs_content_name exists or not
    return True , if vs_content_name exists
    else return False
    """
    api_instance = client.CustomObjectsApi()
    try:
        api_response = api_instance.get_cluster_custom_object(
            group="snapshot.storage.k8s.io",
            version="v1",
            plural="volumesnapshotcontents",
            name=vs_content_name
        )
        LOGGER.debug(api_response)
        LOGGER.info(f"Volume Snapshot Content Check : {vs_content_name} exists")
        return True
    except ApiException:
        LOGGER.info(f"Volume Snapshot content {vs_content_name} does not exists")
        return False


def get_snapshot_and_related_fileset(vs_content_name, pvc_name, created_objects):
    """
    checks volume snapshot content vs_content_name 
    return snapshot and its related fileset name on IBM Storage Scale
    by parsing snapshotHandle
    """
    api_instance = client.CustomObjectsApi()
    try:
        api_response = api_instance.get_cluster_custom_object(
            group="snapshot.storage.k8s.io",
            version="v1",
            plural="volumesnapshotcontents",
            name=vs_content_name
        )
        LOGGER.debug(api_response)
        snapshot_handle = api_response["status"]["snapshotHandle"]
        snapshot_handle = snapshot_handle.split(";")
        if len(snapshot_handle) > 5:
            snapshot_name = snapshot_handle[6]
        else:
            snapshot_name = snapshot_handle[3]

        if snapshot_handle[0] == "1":
            return snapshot_name, snapshot_handle[4]

        volume_name = get_pv_for_pvc(pvc_name, created_objects)
        fileset_name = get_filesetname_from_pv(volume_name, created_objects)

        return snapshot_name, fileset_name

    except ApiException:
        LOGGER.info(
            f"Volume Snapshot content {vs_content_name} does not exists, Unable to get snapshotHandle")
        clean_with_created_objects(created_objects, condition="failed")
        assert False


def set_keep_objects(keep_object):
    """ sets the keep_objects global for use in later functions"""
    global keep_objects
    keep_objects = keep_object


def clean_with_created_objects(created_objects, condition):

    if keep_objects == "onfailure" and condition == "failed":
        return

    for ds_name in copy.deepcopy(created_objects["ds"]):
        delete_ds(ds_name, created_objects)
        check_ds_deleted(ds_name, created_objects)

    for pod_name in copy.deepcopy(created_objects["restore_pod"]):
        delete_pod(pod_name, created_objects)
        check_pod_deleted(pod_name, created_objects)

    for pvc_name in copy.deepcopy(created_objects["restore_pvc"]):
        delete_pvc(pvc_name, created_objects)
        check_pvc_deleted(pvc_name, created_objects)

    for pod_name in copy.deepcopy(created_objects["clone_pod"]):
        delete_pod(pod_name, created_objects)
        check_pod_deleted(pod_name, created_objects)

    for pvc_name in copy.deepcopy(created_objects["clone_pvc"]):
        delete_pvc(pvc_name, created_objects)
        check_pvc_deleted(pvc_name, created_objects)

    for vs_name in copy.deepcopy(created_objects["vs"]):
        delete_vs(vs_name, created_objects)
        check_vs_deleted(vs_name, created_objects)

    for vs_content_name in copy.deepcopy(created_objects["vscontent"]):
        delete_vs_content(vs_content_name, created_objects)
        check_vs_content_deleted(vs_content_name, created_objects)

    for scale_snap_data in copy.deepcopy(created_objects["scalesnapshot"]):
        filesetfunc.delete_snapshot(scale_snap_data[0], scale_snap_data[1], created_objects)
        if filesetfunc.check_snapshot_deleted(scale_snap_data[0], scale_snap_data[1]):
            LOGGER.info(
                f"Scale Snapshot Delete : snapshot {scale_snap_data[0]} of volume {scale_snap_data[1]} deleted successfully")
        else:
            LOGGER.error(
                f"Scale Snapshot Delete : snapshot {scale_snap_data[0]} of {scale_snap_data[1]} not deleted, asserting")
            clean_with_created_objects(created_objects, condition="failed")
            assert False

    for vs_class_name in copy.deepcopy(created_objects["vsclass"]):
        delete_vs_class(vs_class_name, created_objects)
        check_vs_class_deleted(vs_class_name, created_objects)

    for pod_name in copy.deepcopy(created_objects["pod"]):
        delete_pod(pod_name, created_objects)
        check_pod_deleted(pod_name, created_objects)

    for pvc_name in copy.deepcopy(created_objects["pvc"]):
        delete_pvc(pvc_name, created_objects)
        check_pvc_deleted(pvc_name, created_objects)

    for pv_name in copy.deepcopy(created_objects["pv"]):
        delete_pv(pv_name, created_objects)
        check_pv_deleted(pv_name, created_objects)

    for dir_name in copy.deepcopy(created_objects["dir"]):
        filesetfunc.delete_dir(dir_name)

    for sc_name in copy.deepcopy(created_objects["sc"]):
        delete_storage_class(sc_name, created_objects)
        check_storage_class_deleted(sc_name, created_objects)

    for cg_fileset_name in copy.deepcopy(created_objects["cg"]):
        check_cg_fileset_deleted(cg_fileset_name, created_objects)

    for scale_fileset in copy.deepcopy(created_objects["fileset"]):
        filesetfunc.delete_created_fileset(scale_fileset)


def delete_pod_with_graceperiod_0(pod_name, created_objects):
    """ deletes pod pod_name """
    if keep_objects == "True":
        return
    api_instance = client.CoreV1Api()
    try:
        LOGGER.info(f'POD Delete : Deleting pod {pod_name}')
        api_response = api_instance.delete_namespaced_pod(
            name=pod_name, namespace=namespace_value, pretty=True, grace_period_seconds=0)
        LOGGER.debug(str(api_response))
        if pod_name[0:12] == "snap-end-pod":
            created_objects["restore_pod"].remove(pod_name)
        elif pod_name[0:5] == "clone":
            created_objects["clone_pod"].remove(pod_name)
        else:
            created_objects["pod"].remove(pod_name)

    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CoreV1Api->delete_namespaced_pod: {e}")
        clean_with_created_objects(created_objects, condition="failed")
        assert False


def delete_pod(pod_name, created_objects):
    """ deletes pod pod_name """
    if keep_objects == "True":
        return
    api_instance = client.CoreV1Api()
    try:
        LOGGER.info(f'POD Delete : Deleting pod {pod_name}')
        api_response = api_instance.delete_namespaced_pod(
            name=pod_name, namespace=namespace_value, pretty=True)
        LOGGER.debug(str(api_response))
        if pod_name[0:12] == "snap-end-pod":
            created_objects["restore_pod"].remove(pod_name)
        elif pod_name[0:5] == "clone":
            created_objects["clone_pod"].remove(pod_name)
        else:
            created_objects["pod"].remove(pod_name)

    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CoreV1Api->delete_namespaced_pod: {e}")
        clean_with_created_objects(created_objects, condition="failed")
        assert False


def check_pod_deleted(pod_name, created_objects):
    """ checks pod deleted or not , if not deleted , asserts """
    if keep_objects == "True":
        return
    count = 12
    api_instance = client.CoreV1Api()
    while (count > 0):
        try:
            api_response = api_instance.read_namespaced_pod(
                name=pod_name, namespace=namespace_value, pretty=True)
            LOGGER.debug(str(api_response))
            count = count-1
            time.sleep(15)
            LOGGER.info(f'POD Delete : Checking deletion for Pod {pod_name}')
        except ApiException:
            LOGGER.info(f'POD Delete : Pod {pod_name} has been deleted')
            return

    LOGGER.error(f'Pod {pod_name} is still not deleted')
    clean_with_created_objects(created_objects, condition="failed")
    assert False


def delete_pvc(pvc_name, created_objects):
    """ deletes pvc pvc_name and return name of pv associated with it"""

    api_instance = client.CoreV1Api()
    try:
        api_response = api_instance.read_namespaced_persistent_volume_claim(
            name=pvc_name, namespace=namespace_value, pretty=True)
        LOGGER.debug(str(api_response))
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CoreV1Api->read_namespaced_persistent_volume_claim: {e}")
        LOGGER.error(f"PVC {pvc_name} does not exists on the cluster")
        clean_with_created_objects(created_objects, condition="failed")
        assert False

    volume_name = api_response.spec.volume_name
    fileset_name = get_filesetname_from_pv(volume_name, created_objects)

    if keep_objects == "True":
        return fileset_name

    api_instance = client.CoreV1Api()
    try:
        LOGGER.info(f'PVC Delete : Deleting pvc {pvc_name}')
        api_response = api_instance.delete_namespaced_persistent_volume_claim(
            name=pvc_name, namespace=namespace_value, pretty=True, grace_period_seconds=0)
        LOGGER.debug(str(api_response))
        if pvc_name[0:12] == "restored-pvc":
            created_objects["restore_pvc"].remove(pvc_name)
        elif pvc_name[0:5] == "clone":
            created_objects["clone_pvc"].remove(pvc_name)
        else:
            created_objects["pvc"].remove(pvc_name)
        created_objects["fileset"].append(fileset_name)
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CoreV1Api->delete_namespaced_persistent_volume_claim: {e}")
        clean_with_created_objects(created_objects, condition="failed")
        assert False


def check_pvc_deleted(pvc_name, created_objects):
    """ check pvc deleted or not , if not deleted , asserts """
    if keep_objects == "True":
        return
    count = 30
    api_instance = client.CoreV1Api()
    while (count > 0):
        try:
            api_response = api_instance.read_namespaced_persistent_volume_claim(
                name=pvc_name, namespace=namespace_value, pretty=True)
            LOGGER.debug(str(api_response))
            count = count-1
            time.sleep(15)
            LOGGER.info(f'PVC Delete : Checking deletion for pvc {pvc_name}')
        except ApiException:
            LOGGER.info(f'PVC Delete : pvc {pvc_name} deleted')
            return

    LOGGER.error(f'pvc {pvc_name} is not deleted')
    clean_with_created_objects(created_objects, condition="failed")
    assert False


def delete_pv(pv_name, created_objects):
    """ delete pv pv_name """
    if keep_objects == "True":
        return
    api_instance = client.CoreV1Api()
    try:
        LOGGER.info(f'PV Delete : Deleting pv {pv_name}')
        api_response = api_instance.delete_persistent_volume(
            name=pv_name, pretty=True, grace_period_seconds=0)
        LOGGER.debug(str(api_response))
        created_objects["pv"].remove(pv_name)
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CoreV1Api->delete_persistent_volume: {e}")
        clean_with_created_objects(created_objects, condition="failed")
        assert False


def check_pv_deleted(pv_name, created_objects):
    """ checks pv is deleted or not , if not deleted ,asserts"""
    if keep_objects == "True":
        return
    count = 12
    api_instance = client.CoreV1Api()
    while (count > 0):
        try:
            api_response = api_instance.read_persistent_volume(
                name=pv_name, pretty=True)
            LOGGER.debug(str(api_response))
            count = count-1
            time.sleep(15)
            LOGGER.info(f'PV Delete : Checking deletion for PV {pv_name}')
        except ApiException:
            LOGGER.info(f'PV Delete : PV {pv_name} has been deleted')
            return

    LOGGER.error(f'PV {pv_name} is still not deleted')
    clean_with_created_objects(created_objects, condition="failed")
    assert False


def delete_storage_class(sc_name, created_objects):
    """deletes storage class sc_name"""
    if sc_name == "" or keep_objects == "True":
        return
    api_instance = client.StorageV1Api()
    try:
        LOGGER.info(f'SC Delete : deleting storage class {sc_name}')
        api_response = api_instance.delete_storage_class(
            name=sc_name, pretty=True, grace_period_seconds=0)
        LOGGER.debug(str(api_response))
        created_objects["sc"].remove(sc_name)
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling StorageV1Api->delete_storage_class: {e}")
        clean_with_created_objects(created_objects, condition="failed")
        assert False


def check_storage_class_deleted(sc_name, created_objects):
    """
    checks storage class sc_name deleted
    if sc not deleted , asserts
    """
    if sc_name == "" or keep_objects == "True":
        return
    count = 12
    api_instance = client.StorageV1Api()
    while (count > 0):
        try:
            api_response = api_instance.read_storage_class(
                name=sc_name, pretty=True)
            LOGGER.debug(str(api_response))
            count = count-1
            time.sleep(15)
            LOGGER.info(f'SC Delete : Checking deletion for StorageClass {sc_name}')
        except ApiException:
            LOGGER.info(f'SC Delete : StorageClass {sc_name} has been deleted')
            return

    LOGGER.error(f'StorageClass {sc_name} is not deleted')
    clean_with_created_objects(created_objects, condition="failed")
    assert False


def delete_vs_content(vs_content_name, created_objects):
    """
    deletes volume snapshot content vs_content_name
    """
    if keep_objects == "True":
        return
    custom_object_api_instance = client.CustomObjectsApi()
    try:
        custom_object_api_response = custom_object_api_instance.delete_cluster_custom_object(
            group="snapshot.storage.k8s.io",
            version="v1",
            plural="volumesnapshotcontents",
            name=vs_content_name
        )
        LOGGER.debug(custom_object_api_response)
        created_objects["vscontent"].remove(vs_content_name)
        LOGGER.info(f"Volume Snapshot Content Delete : {vs_content_name} deleted")
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CustomObjectsApi->delete_cluster_custom_object_0: {e}")
        clean_with_created_objects(created_objects, condition="failed")
        assert False


def check_vs_content_deleted(vs_content_name, created_objects):
    """
    if volume snapshot content vs_content_name  exists ,  assert
    """
    if keep_objects == "True":
        return
    api_instance = client.CustomObjectsApi()
    val = 0
    while val < 12:
        try:
            api_response = api_instance.get_cluster_custom_object(
                group="snapshot.storage.k8s.io",
                version="v1beta1",
                plural="volumesnapshotcontents",
                name=vs_content_name
            )
            LOGGER.debug(api_response)
            time.sleep(15)
            LOGGER.info(f"Volume Snapshot Content Delete : Checking deletion {vs_content_name}")
            val += 1
        except ApiException:
            LOGGER.info(f"Volume Snapshot Content Delete : {vs_content_name} deletion confirmed")
            return
    LOGGER.error(f"Volume Snapshot Content Delete : {vs_content_name} is not deleted , asserting")
    clean_with_created_objects(created_objects, condition="failed")
    assert False


def delete_vs(vs_name, created_objects):
    """
    delete volume snapshot vs_name
    """
    if keep_objects == "True":
        return
    custom_object_api_instance = client.CustomObjectsApi()
    try:
        custom_object_api_response = custom_object_api_instance.delete_namespaced_custom_object(
            group="snapshot.storage.k8s.io",
            version="v1",
            plural="volumesnapshots",
            name=vs_name,
            namespace=namespace_value
        )
        LOGGER.debug(custom_object_api_response)
        LOGGER.info(f"Volume Snapshot Delete : {vs_name} deleted")
        created_objects["vs"].remove(vs_name)
    except ApiException as e:
        LOGGER.error(f"Exception when calling CustomObjectsApi->delete_cluster_custom_object: {e}")
        clean_with_created_objects(created_objects, condition="failed")
        assert False


def check_vs_deleted(vs_name, created_objects):
    """
    if volume snapshot vs_name exists , it asserts
    """
    if keep_objects == "True":
        return
    api_instance = client.CustomObjectsApi()
    val = 0
    while val < 12:
        try:
            api_response = api_instance.get_namespaced_custom_object(
                group="snapshot.storage.k8s.io",
                version="v1",
                plural="volumesnapshots",
                name=vs_name,
                namespace=namespace_value
            )
            LOGGER.debug(api_response)
            time.sleep(15)
            LOGGER.info(f"Volume Snapshot Delete : Checking deletion for {vs_name}")
            val += 1
        except ApiException:
            LOGGER.info(f"Volume Snapshot Delete : {vs_name} deletion confirmed")
            return
    LOGGER.error(f"Volume Snapshot Delete : {vs_name} is not deleted , asserting")
    clean_with_created_objects(created_objects, condition="failed")
    assert False


def delete_vs_class(vs_class_name, created_objects):
    """
    deletes volume snapshot class vs_class_name
    """
    if keep_objects == "True":
        return
    custom_object_api_instance = client.CustomObjectsApi()
    try:
        custom_object_api_response = custom_object_api_instance.delete_cluster_custom_object(
            group="snapshot.storage.k8s.io",
            version="v1",
            plural="volumesnapshotclasses",
            name=vs_class_name
        )
        LOGGER.debug(custom_object_api_response)
        LOGGER.info(f"Volume Snapshot Class Delete : {vs_class_name} deleted")
        created_objects["vsclass"].remove(vs_class_name)
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CustomObjectsApi->delete_cluster_custom_object_0: {e}")
        clean_with_created_objects(created_objects, condition="failed")
        assert False


def check_vs_class_deleted(vs_class_name, created_objects):
    """
    if volume snapshot class vs_class_name  exists ,  assert
    """
    if keep_objects == "True":
        return
    api_instance = client.CustomObjectsApi()
    try:
        api_response = api_instance.get_cluster_custom_object(
            group="snapshot.storage.k8s.io",
            version="v1",
            plural="volumesnapshotclasses",
            name=vs_class_name
        )
        LOGGER.debug(api_response)
        LOGGER.error(f"Volume Snapshot Class Delete : {vs_class_name} is not deleted , asserting")
        clean_with_created_objects(created_objects, condition="failed")
        assert False
    except ApiException:
        LOGGER.info(f"Volume Snapshot Class Delete : {vs_class_name} deletion confirmed")


def delete_ds(ds_name, created_objects):
    if keep_objects == "True":
        return
    api_instance = client.AppsV1Api()

    try:
        api_response = api_instance.delete_namespaced_daemon_set(
            name=ds_name, namespace=namespace_value)
        LOGGER.debug(api_response)
        LOGGER.info(f"Daemon Set Delete : {ds_name} deleted")
        created_objects["ds"].remove(ds_name)
    except ApiException as e:
        LOGGER.error(f"Exception when calling AppsV1Api->delete_namespaced_daemon_set: {e}")
        clean_with_created_objects(created_objects, condition="failed")
        assert False


def check_ds_deleted(ds_name, created_objects):
    if keep_objects == "True":
        return

    api_instance = client.AppsV1Api()
    try:
        api_response = api_instance.read_namespaced_daemon_set(
            name=ds_name, namespace=namespace_value)
        LOGGER.debug(api_response)
        LOGGER.error(f"Daemon Set Delete : {ds_name} is not deleted , asserting")
        clean_with_created_objects(created_objects, condition="failed")
        assert False
    except ApiException:
        LOGGER.info(f"Daemon Set Delete : {ds_name} deletion confirmed")


def get_filesetname_from_pv(volume_name, created_objects):
    """
    return filesetname from VolumeHandle of PV
    """
    api_instance = client.CoreV1Api()
    fileset_name = None

    if volume_name is not None:
        try:
            api_response = api_instance.read_persistent_volume(
                name=volume_name, pretty=True)
            LOGGER.debug(str(api_response))
            volume_handle = api_response.spec.csi.volume_handle
            volume_handle = volume_handle.split(";")
            if len(volume_handle) == 3:
                fileset_name = ""
            elif len(volume_handle) <= 4:
                fileset_name = volume_handle[2][12:]
            else:
                fileset_name = volume_handle[5]
            if fileset_name == "":
                fileset_name = "LW"
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_persistent_volume: {e}")
            LOGGER.info(f'PV {volume_name} does not exists on cluster')

    if volume_name is not None and fileset_name is None:
        LOGGER.error(f"Not able to find fileset name for PV {volume_name}")
        clean_with_created_objects(created_objects, condition="failed")
        assert False

    return fileset_name


def get_cg_filesetname_from_pv(volume_name, created_objects):
    """
    return consistency group filesetname from VolumeHandle of PV
    """
    api_instance = client.CoreV1Api()
    cg_fileset_name = None

    if volume_name is not None:
        try:
            api_response = api_instance.read_persistent_volume(
                name=volume_name, pretty=True)
            LOGGER.debug(str(api_response))
            volume_handle = api_response.spec.csi.volume_handle
            volume_handle = volume_handle.split(";")
            cg_fileset_name = volume_handle[4]
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_persistent_volume: {e}")
            LOGGER.info(f'PV {volume_name} does not exists on cluster')

    if volume_name is not None and cg_fileset_name is None:
        LOGGER.error(f"Not able to find cg fileset name for PV {volume_name}")
        clean_with_created_objects(created_objects, condition="failed")
        assert False

    return cg_fileset_name


def check_cg_fileset_deleted(cg_fileset_name, created_objects):
    if keep_objects == "True":
        return

    for _ in range(0, 24):
        LOGGER.info(f"Checking for deletion of consistency group fileset {cg_fileset_name}")
        if filesetfunc.check_fileset_deleted(cg_fileset_name):
            created_objects["cg"].remove(cg_fileset_name)
            LOGGER.info(f"Consistency group fileset {cg_fileset_name} is deleted")
            break
    else:
        created_objects["cg"].remove(cg_fileset_name)
        LOGGER.error(f"Consistency group fileset {cg_fileset_name} is not deleted")
        clean_with_created_objects(created_objects, condition="failed")
        assert False
