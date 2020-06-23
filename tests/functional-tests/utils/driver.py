import time
import re
import logging
from datetime import datetime, timezone
from kubernetes import client
from kubernetes.client.rest import ApiException
from kubernetes.stream import stream
import utils.fileset_functions as ff
from utils.namegenerator import name_generator

LOGGER = logging.getLogger()
"""
delete_dir
delete_created_fileset
"""


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


def create_storage_class(values, config_value, sc_name):
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
    global test, storage_class_parameters
    test = config_value
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
            f'creating storageclass {sc_name} with parameters {str(storage_class_parameters)}')
        api_response = api_instance.create_storage_class(
            body=storage_class_body, pretty=True)
        LOGGER.debug(str(api_response))
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling StorageV1Api->create_storage_class: {e}")
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
    if sc_name=="":
        return False
    api_instance = client.StorageV1Api()
    try:
        LOGGER.info(f'Storage class {sc_name} does exists on the cluster')
        # TBD: Show StorageClass Parameter in tabular Form
        api_response = api_instance.read_storage_class(
            name=sc_name, pretty=True)
        LOGGER.debug(str(api_response))
        return True
    except ApiException:
        LOGGER.info("strorage class does not exists")
        return False


def create_pv(pv_values, pv_name, sc_name=""):
    """
    creates persistent volume

    Args:
        param1: pv_values - values required for creation of pv
        param2: pv_name - name of pv to be created

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
            storage_class_name = sc_name
        )
    else:
        pv_spec = client.V1PersistentVolumeSpec(
            access_modes=[pv_values["access_modes"]],
            capacity={"storage": pv_values["storage"]},
            csi=pv_csi,
            persistent_volume_reclaim_policy=pv_values["reclaim_policy"],
            storage_class_name = sc_name
        )

    pv_body = client.V1PersistentVolume(
        api_version="v1",
        kind="PersistentVolume",
        metadata=pv_metadata,
        spec=pv_spec
    )
    try:
        LOGGER.info(f'Creating PV {pv_name} with {str(pv_name)} parameter')
        api_response = api_instance.create_persistent_volume(
            body=pv_body, pretty=True)
        LOGGER.debug(str(api_response))
        LOGGER.info(f'PV {pv_name} has been created successfully ')

    except ApiException as e:
        LOGGER.error(f'PV {pv_name} creation failed hence failing test case ')
        LOGGER.error(
            f"Exception when calling CoreV1Api->create_persistent_volume: {e}")
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
        LOGGER.info(f'Created PV {pv_name} exists on cluster')
        return True
    except ApiException:
        LOGGER.info(f'PV {pv_name} does not exists on cluster')
        return False


def create_pvc(pvc_values, sc_name, pvc_name, config_value=None, pv_name=None):
    """
    creates persistent volume claim

    Args:
        param1: pvc_values - values required for creation of pvc
        param2: sc_name - name of storage class , pvc associated with
                          if "notusingsc" no storage class
        param3: pvc_name - name of pvc to be created
        param4: config_value - configuration file
        param5: pv_name - name of pv , pvc associated with
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
    if sc_name == "":
        global test
        test = config_value

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
            f'Creating pvc {pvc_name} with parameters {str(pvc_values)} and storageclass {str(sc_name)}')
        api_response = api_instance.create_namespaced_persistent_volume_claim(
            namespace=namespace_value, body=pvc_body, pretty=True)
        LOGGER.debug(str(api_response))
        LOGGER.info(f'PVC {pvc_name} has been created successfully')
    except ApiException as e:
        LOGGER.info(f'PVC {pvc_name} creation operation has been failed')
        LOGGER.error(
            f"Exception when calling CoreV1Api->create_namespaced_persistent_volume_claim: {e}")
        assert False


def clean_pvc_fail(sc_name, pvc_name, pv_name, dir_name):
    """  cleanup after pvc has failed    """
    LOGGER.info(f'PVC {pvc_name} cleanup operation started')
    delete_pvc(pvc_name)
    check_pvc_deleted(pvc_name)
    if check_pv(pv_name):
        delete_pv(pv_name)
        check_pv_deleted(pv_name)
    if check_storage_class(sc_name):
        delete_storage_class(sc_name)
        check_storage_class_deleted(sc_name)
    if not(dir_name == "nodiravailable"):
        ff.delete_dir(test, dir_name)


def pvc_bound_fileset_check(api_response, pv_name, pvc_name):
    global volume_name
    now1 = api_response.metadata.creation_timestamp
    now = datetime.now(timezone.utc)
    timediff = now-now1
    minutes = divmod(timediff.seconds, 60)
    msg = 'Time for PVC BOUND : ' + \
        str(minutes[0]) + 'minutes', str(minutes[1]) + 'seconds'
    LOGGER.info(f'PVC Check : {pvc_name} is BOUND succesfully {msg}')
    volume_name = api_response.spec.volume_name
    if volume_name == pv_name:
        LOGGER.info("It should be case of static pvc")
        return True
    if 'storage_class_parameters' in globals():
        if check_key(storage_class_parameters, "volDirBasePath") and check_key(storage_class_parameters, "volBackendFs"):
            return True

    val = ff.created_fileset_exists(test, volume_name)
    if val is False:
        return False

    LOGGER.info(f'PVC Check : Fileset {volume_name} has been created successfully')
    return True


def check_pvc(pvc_values, sc_name, pvc_name, dir_name="nodiravailable", pv_name="pvnotavailable"):
    """ checks pvc is BOUND or not
        need to reduce complextity of this function
    """
    api_instance = client.CoreV1Api()
    con = True
    var = 0
    global volume_name
    while (con is True):
        try:
            api_response = api_instance.read_namespaced_persistent_volume_claim(
                name=pvc_name, namespace=namespace_value, pretty=True)
            LOGGER.debug(str(api_response))
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_persistent_volume_claim: {e}")
            LOGGER.info(f"PVC {pvc_name} does not exists on the cluster")
            assert False

        if api_response.status.phase == "Bound":   
            if(check_key(pvc_values, "reason")):
                LOGGER.error(f'PVC Check : {pvc_name} is BOUND but as the failure reason is provided so\
                asserting the test')
                volume_name = api_response.spec.volume_name
                clean_pvc_fail(sc_name, pvc_name, pv_name, dir_name)
                assert False
            if(pvc_bound_fileset_check(api_response, pv_name, pvc_name)):
                return True
            clean_pvc_fail(sc_name, pvc_name, pv_name, dir_name)
            LOGGER.error(f'PVC Check : Fileset {volume_name} doesn\'t exists')
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
                LOGGER.info("PVC is not BOUND,checking if failure reason is expected")
                field = "involvedObject.name="+pvc_name
                reason = api_instance.list_namespaced_event(
                    namespace=namespace_value, pretty=True, field_selector=field)
                volume_name = api_response.spec.volume_name
                if volume_name is None:
                    volume_name = "sc-ffwe-sdfsdf-gsv"
                if not(check_key(pvc_values, "reason")):
                    clean_pvc_fail(sc_name, pvc_name, pv_name, dir_name)
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
                    clean_pvc_fail(sc_name, pvc_name, pv_name, dir_name)
                    LOGGER.error(f"Failed reason : {str(reason)}")
                    LOGGER.info("PVC is not Bound but FAILED reason does not match")
                    assert False
                else:
                    LOGGER.debug(search_result)
                    LOGGER.info("PVC is not Bound and FAILED with expected error")
                    con = False


def create_pod(value_pod, pvc_name, pod_name):
    """
    creates pod

    Args:
        param1: value_pod - values required for creation of pod
        param2: pvc_name - name of pvc , pod associated with
        param3: pod_name - name of pod to be created

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
        name="web-server", image="nginx:1.19.0", volume_mounts=[pod_volume_mounts], ports=[pod_ports])
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
        LOGGER.info(f'creating pod {pod_name} with {str(value_pod)}')
        api_response = api_instance.create_namespaced_pod(
            namespace=namespace_value, body=pod_body, pretty=True)
        LOGGER.debug(str(api_response))
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CoreV1Api->create_namespaced_pod: {e}")
        assert False


def clean_pod_fail(sc_name, pvc_name, pv_name, dir_name, pod_name):
    """ cleanup after pod fails """
    delete_pod(pod_name)
    check_pod_deleted(pod_name)
    delete_pvc(pvc_name)
    check_pvc_deleted(pvc_name)
    if check_pv(pv_name):
        delete_pv(pv_name)
        check_pv_deleted(pv_name)
    if check_storage_class(sc_name):
        delete_storage_class(sc_name)
        check_storage_class_deleted(sc_name)
    if not(dir_name == "nodiravailable"):
        ff.delete_dir(test, dir_name)


def check_pod_execution(value_pod, sc_name, pvc_name, pod_name, dir_name, pv_name):
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
            clean_pod_fail(sc_name, pvc_name, pv_name, dir_name, pod_name)
            LOGGER.error("Pod should not be able to create file inside the pod as failure REASON provided, so asserting")
            assert False
        return
    if not(check_key(value_pod, "reason")):
        clean_pod_fail(sc_name, pvc_name, pv_name, dir_name, pod_name)
        LOGGER.error(str(resp))
        LOGGER.info("FAILED as reason of failure not provided")
        assert False
    search_result1 = re.search(value_pod["reason"], str(resp))
    search_result2 = re.search("Permission denied", str(resp))
    LOGGER.info(str(search_result1))
    LOGGER.info(str(search_result2))
    if not(search_result1 is None and search_result2 is None):
        LOGGER.info("execution of pod failed with expected reason")
    else:
        clean_pod_fail(sc_name, pvc_name, pv_name, dir_name, pod_name)
        LOGGER.error(str(resp))
        LOGGER.error(
            "execution of pod failed unexpected , reason does not match")
        assert False


def check_pod(value_pod, sc_name, pvc_name, pod_name, dir_name="nodiravailable", pv_name="pvnotavailable"):
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
            if api_response.status.phase == "Running":
                LOGGER.info(f'POD Check : POD {pod_name} is Running')
                check_pod_execution(value_pod, sc_name,
                                    pvc_name, pod_name, dir_name, pv_name)
                con = False
            else:
                var += 1
                if(var > 20):
                    LOGGER.info(f'POD {pod_name} is not running')
                    field = "involvedObject.name="+pod_name
                    reason = api_instance.list_namespaced_event(
                        namespace=namespace_value, pretty=True, field_selector=field)
                    LOGGER.info(f"POD Check : Reason of failure is : {str(reason)}")
                    con = False
                    clean_pod_fail(sc_name, pvc_name, pv_name,
                                   dir_name, pod_name)
                    assert False
                time.sleep(5)
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod: {e}")
            LOGGER.info("POD Check : POD does not exists on Cluster")
            con = False
            assert False


def delete_pod(pod_name):
    """ deletes pod pod_name """
    api_instance = client.CoreV1Api()
    try:
        LOGGER.info(f'Deleting pod {pod_name}')
        api_response = api_instance.delete_namespaced_pod(
            name=pod_name, namespace=namespace_value, pretty=True, grace_period_seconds=0)
        LOGGER.debug(str(api_response))
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CoreV1Api->delete_namespaced_pod: {e}")
        assert False


def check_pod_deleted(pod_name):
    """ checks pod deleted or not , if not deleted , asserts """
    count = 12
    api_instance = client.CoreV1Api()
    while (count > 0):
        try:
            api_response = api_instance.read_namespaced_pod(
                name=pod_name, namespace=namespace_value, pretty=True)
            LOGGER.debug(str(api_response))
            count = count-1
            time.sleep(5)
        except ApiException:
            LOGGER.info(f'Pod {pod_name} has been deleted')
            return

    LOGGER.info(f'Pod {pod_name} is still not deleted')
    assert False


def delete_pvc(pvc_name):
    """ deletes pvc pvc_name """
    api_instance = client.CoreV1Api()
    try:
        LOGGER.info(f'Deleting pvc {pvc_name}')
        api_response = api_instance.delete_namespaced_persistent_volume_claim(
            name=pvc_name, namespace=namespace_value, pretty=True, grace_period_seconds=0)
        LOGGER.debug(str(api_response))
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CoreV1Api->delete_namespaced_persistent_volume_claim: {e}")
        assert False


def check_pvc_deleted(pvc_name):
    """ check pvc deleted or not , if not deleted , asserts """
    count = 12
    api_instance = client.CoreV1Api()
    while (count > 0):
        try:
            api_response = api_instance.read_namespaced_persistent_volume_claim(
                name=pvc_name, namespace=namespace_value, pretty=True)
            LOGGER.debug(str(api_response))
            count = count-1
            time.sleep(5)
        except ApiException:
            LOGGER.info(f'pvc {pvc_name} deleted')
            ff.delete_created_fileset(test, volume_name)
            return

    LOGGER.info(f'pvc {pvc_name} is not deleted')
    assert False


def delete_pv(pv_name):
    """ delete pv pv_name """
    api_instance = client.CoreV1Api()
    try:
        LOGGER.info(f'Deleting pv {pv_name}')
        api_response = api_instance.delete_persistent_volume(
            name=pv_name, pretty=True, grace_period_seconds=0)
        LOGGER.debug(str(api_response))
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CoreV1Api->delete_persistent_volume: {e}")
        assert False


def check_pv_deleted(pv_name):
    """ checks pv is deleted or not , if not deleted ,asserts"""
    count = 12
    api_instance = client.CoreV1Api()
    while (count > 0):
        try:
            api_response = api_instance.read_persistent_volume(
                name=pv_name, pretty=True)
            LOGGER.debug(str(api_response))
            count = count-1
            time.sleep(5)
        except ApiException:
            LOGGER.info(f'PV {pv_name} has been deleted')
            return
        
    LOGGER.info(f'PV {pv_name} is still not deleted')
    assert False


def delete_storage_class(sc_name):
    """deletes storage class sc_name"""
    if sc_name == "":
        return
    api_instance = client.StorageV1Api()
    try:
        LOGGER.info(f'deleting storage class {sc_name}')
        api_response = api_instance.delete_storage_class(
            name=sc_name, pretty=True, grace_period_seconds=0)
        LOGGER.debug(str(api_response))
        # can use cleanup function if running more than 20 parallel testcases to make
        # sure that every fileset is deleted
        # ff.cleanup(test)
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling StorageV1Api->delete_storage_class: {e}")
        assert False


def check_storage_class_deleted(sc_name):
    """
    checks storage class sc_name deleted
    if sc not deleted , asserts
    """
    if sc_name == "":
        return
    count = 12
    api_instance = client.StorageV1Api()
    while (count > 0):
        try:
            api_response = api_instance.read_storage_class(
                name=sc_name, pretty=True)
            LOGGER.debug(str(api_response))
            count = count-1
            time.sleep(5)
        except ApiException:
            LOGGER.info(f'StorageClass {sc_name} has been deleted')
            return

    LOGGER.info(f'StorageClass {sc_name} is not deleted')
    assert False
