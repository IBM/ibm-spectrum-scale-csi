import time
import re
import logging
from datetime import datetime, timezone
from kubernetes import client
from kubernetes.client.rest import ApiException
from kubernetes.stream import stream
import utils.fileset_functions as ff
import utils.cleanup_functions as cleanup
from utils.namegenerator import name_generator

LOGGER = logging.getLogger()


def check_key(dict1, key):
    """ checks key is in dictionary or not"""
    if key in dict1.keys():
        return True
    return False


def set_test_namespace_value(namespace_name=None):
    """ sets the test namespace global for use in later functions"""
    global namespace_value
    namespace_value = namespace_name


def get_random_name(type_of):
    """ return random name of type_of"""
    return type_of+"-"+name_generator()


def get_storage_class_parameters(values):
    """
    create parameters for storage class

    Args:
        param1: values - storage class input values

    Returns:
        storage class parameters

    Raises:
        None

    """
    dict_parameters = {}
    list_parameters = ["volBackendFs", "clusterId", "volDirBasePath",
                       "uid", "gid", "filesetType", "parentFileset", "inodeLimit"]
    num = len(list_parameters)
    for val in range(0, num):
        if(check_key(values, list_parameters[val])):
            dict_parameters[list_parameters[val]
                            ] = values[list_parameters[val]]
    return dict_parameters


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
    storage_class_parameters = get_storage_class_parameters(values)
    storage_class_body = client.V1StorageClass(
        api_version="storage.k8s.io/v1",
        kind="StorageClass",
        metadata=storage_class_metadata,
        provisioner="spectrumscale.csi.ibm.com",
        parameters=storage_class_parameters,
        reclaim_policy="Delete"
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
        cleanup.clean_with_created_objects(created_objects)
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
        cleanup.clean_with_created_objects(created_objects)
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
        cleanup.clean_with_created_objects(created_objects)
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
        cleanup.clean_with_created_objects(created_objects)
        assert False


def pvc_bound_fileset_check(api_response, pv_name, pvc_name):
    """
    calculates bound time for pvc and checks fileset created by
    pvc on spectrum scale
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
        if check_key(storage_class_parameters, "volDirBasePath") and check_key(storage_class_parameters, "volBackendFs"):
            return True

    val = ff.created_fileset_exists(volume_name)
    if val is False:
        LOGGER.error(f'PVC Check : Fileset {volume_name} doesn\'t exists')
        return False

    LOGGER.info(f'PVC Check : Fileset {volume_name} has been created successfully')
    return True


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
            cleanup.clean_with_created_objects(created_objects)
            assert False

        if api_response.status.phase == "Bound":
            if(check_key(pvc_values, "reason")):
                LOGGER.error(f'PVC Check : {pvc_name} is BOUND but as the failure reason is provided so\
                asserting the test')
                cleanup.clean_with_created_objects(created_objects)
                assert False
            if(pvc_bound_fileset_check(api_response, pv_name, pvc_name)):
                return True
            cleanup.clean_with_created_objects(created_objects)
            assert False
        else:
            var += 1
            time.sleep(5)
            if(check_key(pvc_values, "reason")):
                time_count = 8
            elif(check_key(pvc_values, "parallel")):
                time_count = 60
            else:
                time_count = 20
            if(var > time_count):
                LOGGER.info("PVC Check : PVC is not BOUND,checking if failure reason is expected")
                field = "involvedObject.name="+pvc_name
                reason = api_instance.list_namespaced_event(
                    namespace=namespace_value, pretty=True, field_selector=field)
                if not(check_key(pvc_values, "reason")):
                    cleanup.clean_with_created_objects(created_objects)
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
                    cleanup.clean_with_created_objects(created_objects)
                    LOGGER.error(f"Failed reason : {str(reason)}")
                    LOGGER.error("PVC Check : PVC is not Bound but FAILED reason does not match")
                    assert False
                else:
                    LOGGER.debug(search_result)
                    LOGGER.info(f"PVC Check : PVC is not Bound and FAILED with expected error {pvc_values['reason']}")
                    con = False


def create_pod(value_pod, pvc_name, pod_name, created_objects, image_name="nginx:1.19.0"):
    """
    creates pod

    Args:
        param1: value_pod - values required for creation of pod
        param2: pvc_name - name of pvc , pod associated with
        param3: pod_name - name of pod to be created
        param4: image_name - name of the pod image (Default:"nginx:1.19.0")

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
    pod_volume_mounts = client.V1VolumeMount(
        name="mypvc", mount_path=value_pod["mount_path"])
    pod_ports = client.V1ContainerPort(container_port=80)
    pod_containers = client.V1Container(
        name="web-server", image=image_name, volume_mounts=[pod_volume_mounts], ports=[pod_ports])
    pod_persistent_volume_claim = client.V1PersistentVolumeClaimVolumeSource(
        claim_name=pvc_name, read_only=value_pod["read_only"])
    pod_volumes = client.V1Volume(
        name="mypvc", persistent_volume_claim=pod_persistent_volume_claim)
    pod_spec = client.V1PodSpec(
        containers=[pod_containers], volumes=[pod_volumes])
    pod_body = client.V1Pod(
        api_version="v1",
        kind="Pod",
        metadata=pod_metadata,
        spec=pod_spec
    )

    try:
        LOGGER.info(f'POD Create : creating pod {pod_name} using {pvc_name} with {image_name} image')
        api_response = api_instance.create_namespaced_pod(
            namespace=namespace_value, body=pod_body, pretty=True)
        LOGGER.debug(str(api_response))
        if pod_name[0:12] == "snap-end-pod":   
            created_objects["restore_pod"].append(pod_name)
        else:
            created_objects["pod"].append(pod_name)
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CoreV1Api->create_namespaced_pod: {e}")
        cleanup.clean_with_created_objects(created_objects)
        assert False


def create_file_inside_pod(value_pod, pod_name, created_objects):
    """
    create snaptestfile inside the pod using touch
    """
    api_instance = client.CoreV1Api()
    LOGGER.info("POD Check : Trying to create snaptestfile on SpectrumScale mount point inside the pod")
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
        LOGGER.info("file snaptestfile created successfully on SpectrumScale mount point inside the pod")
        return
    LOGGER.error("file snaptestfile not created")
    cleanup.clean_with_created_objects(created_objects)
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
        LOGGER.info("POD Check : snaptestfile is succesfully restored from snapshot")
        return
    LOGGER.error("snaptestfile is not restored from snapshot")
    cleanup.clean_with_created_objects(created_objects)
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
    LOGGER.info("POD Check : Trying to create testfile on SpectrumScale mount point inside the pod")
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
        LOGGER.info("POD Check : Deleting testfile from pod's SpectrumScale mount point")
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
        if check_key(value_pod, "reason"):
            cleanup.clean_with_created_objects(created_objects)
            LOGGER.error("Pod should not be able to create file inside the pod as failure REASON provided, so asserting")
            assert False
        return
    if not(check_key(value_pod, "reason")):
        cleanup.clean_with_created_objects(created_objects)
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
        cleanup.clean_with_created_objects(created_objects)
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
                    if not(check_key(value_pod, "reason")):
                        LOGGER.error('FAILED as reason of failure not provided')
                        LOGGER.error(f"POD Check : Reason of failure is : {str(reason)}")
                        cleanup.clean_with_created_objects(created_objects)
                        assert False
                    search_result = re.search(value_pod["reason"], str(reason))
                    if search_result is None:
                        LOGGER.error(f'Failed as reason of failure does not match {value_pod["reason"]}')
                        LOGGER.error(f"POD Check : Reason of failure is : {str(reason)}")
                        cleanup.clean_with_created_objects(created_objects)
                        assert False
                    else:
                        LOGGER.info(f'POD failed with expected reason {value_pod["reason"]}')
                        return
                time.sleep(5)
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod: {e}")
            LOGGER.error("POD Check : POD does not exists on Cluster")
            cleanup.clean_with_created_objects(created_objects)
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
      "ownerReferences":ds_name
    }
  },
  "spec": {
    "selector": {
      "matchLabels": {
        "name": "nginx",
        "ownerReferences":ds_name
      }
    },
    "template": {
      "metadata": {
        "labels": {
          "name": "nginx",
          "ownerReferences":ds_name
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
        ]
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
        cleanup.clean_with_created_objects(created_objects)
        assert False


def check_ds(ds_name,value_ds,created_objects):
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
                if desired_number_scheduled<2:
                    LOGGER.error(f"Not enough nodes for this test, only {desired_number_scheduled} nodes are there")
                    cleanup.clean_with_created_objects(created_objects)
                    assert False

                if "reason" in value_ds:
                    LOGGER.error(f"failure reason provided  {value_ds} , still all pods are running")
                    cleanup.clean_with_created_objects(created_objects)
                    assert False

                LOGGER.info(f"Daemonset Check : daemonset {ds_name} all {current_number_scheduled} pods are Running")

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
       cleanup.clean_with_created_objects(created_objects)
       assert False

    if desired_number_scheduled<2:
        LOGGER.error(f"Not enough nodes for this test, only {desired_number_scheduled} nodes are there")
        cleanup.clean_with_created_objects(created_objects)
        assert False

    if check_ds_pod(ds_name,value_ds,created_objects):
        LOGGER.info(f"Daemonset Check : daemonset {ds_name} pods failed with expected reason {value_ds['reason']}")
        return

    LOGGER.info(f"Daemonset Check : daemonset {ds_name} pods did not fail with expected reason {value_ds['reason']}")
    cleanup.clean_with_created_objects(created_objects)
    assert False

def check_ds_pod(ds_name,value_ds,created_objects):
    api_instance = client.CoreV1Api()
    selector = "ownerReferences="+ds_name
    running_pod_list,pod_list = [],[]
    try:
        api_response = api_instance.list_namespaced_pod(
                namespace=namespace_value,pretty=True, label_selector=selector)
        LOGGER.debug(api_response)
        for pod in api_response.items:
            if pod.status.phase=="Running":
                running_pod_list.append(pod.metadata.name)
            else:
                pod_list.append(pod.metadata.name)
    except ApiException as e:
        LOGGER.error(
                f"Exception when calling CoreV1Api->list_node: {e}")
        cleanup.clean_with_created_objects(created_objects)
        assert False

    if len(running_pod_list)!=1:
        LOGGER.error(f"running pods are {running_pod_list} , only one pod should be running, asserting")
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
