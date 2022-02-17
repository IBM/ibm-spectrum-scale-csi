import time
import logging
import copy
from kubernetes import client
from kubernetes.client.rest import ApiException
import ibm_spectrum_scale_csi.spectrum_scale_apis.fileset_functions as filesetfunc

LOGGER = logging.getLogger()


def set_test_namespace_value(namespace_name=None):
    """ sets the test namespace global for use in later functions"""
    global namespace_value
    namespace_value = namespace_name


def set_keep_objects(keep_object):
    """ sets the keep_objects global for use in later functions"""
    global keep_objects
    keep_objects = keep_object


def clean_with_created_objects(created_objects):

    for ds_name in copy.deepcopy(created_objects["ds"]):
        delete_ds(ds_name, created_objects)
        check_ds_deleted(ds_name, created_objects)

    for pod_name in copy.deepcopy(created_objects["restore_pod"]):
        delete_pod(pod_name, created_objects)
        check_pod_deleted(pod_name, created_objects)

    for pvc_name in copy.deepcopy(created_objects["restore_pvc"]):
        vol_name = delete_pvc(pvc_name, created_objects)
        check_pvc_deleted(pvc_name, vol_name, created_objects)

    for pod_name in copy.deepcopy(created_objects["clone_pod"]):
        delete_pod(pod_name, created_objects)
        check_pod_deleted(pod_name, created_objects)

    for pvc_name in copy.deepcopy(created_objects["clone_pvc"]):
        vol_name = delete_pvc(pvc_name, created_objects)
        check_pvc_deleted(pvc_name, vol_name, created_objects)

    for vs_name in copy.deepcopy(created_objects["vs"]):
        delete_vs(vs_name, created_objects)
        check_vs_deleted(vs_name, created_objects)

    for vs_content_name in copy.deepcopy(created_objects["vscontent"]):
        delete_vs_content(vs_content_name, created_objects)
        check_vs_content_deleted(vs_content_name, created_objects)

    for scale_snap_data in copy.deepcopy(created_objects["scalesnapshot"]):
        filesetfunc.delete_snapshot(scale_snap_data[0], scale_snap_data[1], created_objects)
        if filesetfunc.check_snapshot_deleted(scale_snap_data[0], scale_snap_data[1]):
            LOGGER.info(f"Scale Snapshot Delete : snapshot {scale_snap_data[0]} of volume {scale_snap_data[1]} deleted successfully")
        else:
            LOGGER.error(f"Scale Snapshot Delete : snapshot {scale_snap_data[0]} of {scale_snap_data[1]} not deleted, asserting")
            clean_with_created_objects(created_objects)
            assert False

    for vs_class_name in copy.deepcopy(created_objects["vsclass"]):
        delete_vs_class(vs_class_name, created_objects)
        check_vs_class_deleted(vs_class_name, created_objects)

    for pod_name in copy.deepcopy(created_objects["pod"]):
        delete_pod(pod_name, created_objects)
        check_pod_deleted(pod_name, created_objects)

    for pvc_name in copy.deepcopy(created_objects["pvc"]):
        vol_name = delete_pvc(pvc_name, created_objects)
        check_pvc_deleted(pvc_name, vol_name, created_objects)

    for pv_name in copy.deepcopy(created_objects["pv"]):
        delete_pv(pv_name, created_objects)
        check_pv_deleted(pv_name, created_objects)

    for dir_name in copy.deepcopy(created_objects["dir"]):
        filesetfunc.delete_dir(dir_name)

    for sc_name in copy.deepcopy(created_objects["sc"]):
        delete_storage_class(sc_name, created_objects)
        check_storage_class_deleted(sc_name, created_objects)


def delete_pod(pod_name, created_objects):
    """ deletes pod pod_name """
    if keep_objects:
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
        clean_with_created_objects(created_objects)
        assert False


def check_pod_deleted(pod_name, created_objects):
    """ checks pod deleted or not , if not deleted , asserts """
    if keep_objects:
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
    clean_with_created_objects(created_objects)
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
        clean_with_created_objects(created_objects)
        assert False

    volume_name = api_response.spec.volume_name
    fileset_name = get_filesetname_from_pv(volume_name, created_objects)

    if keep_objects:
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
        return fileset_name
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CoreV1Api->delete_namespaced_persistent_volume_claim: {e}")
        clean_with_created_objects(created_objects)
        assert False


def check_pvc_deleted(pvc_name, volume_name, created_objects):
    """ check pvc deleted or not , if not deleted , asserts """
    if keep_objects:
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
            filesetfunc.delete_created_fileset(volume_name)
            return

    LOGGER.error(f'pvc {pvc_name} is not deleted')
    clean_with_created_objects(created_objects)
    assert False


def delete_pv(pv_name, created_objects):
    """ delete pv pv_name """
    if keep_objects:
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
        clean_with_created_objects(created_objects)
        assert False


def check_pv_deleted(pv_name, created_objects):
    """ checks pv is deleted or not , if not deleted ,asserts"""
    if keep_objects:
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
    clean_with_created_objects(created_objects)
    assert False


def delete_storage_class(sc_name, created_objects):
    """deletes storage class sc_name"""
    if sc_name == "" or keep_objects:
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
        clean_with_created_objects(created_objects)
        assert False


def check_storage_class_deleted(sc_name, created_objects):
    """
    checks storage class sc_name deleted
    if sc not deleted , asserts
    """
    if sc_name == "" or keep_objects:
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
    clean_with_created_objects(created_objects)
    assert False


def delete_vs_content(vs_content_name, created_objects):
    """
    deletes volume snapshot content vs_content_name
    """
    if keep_objects:
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
        LOGGER.error(f"Exception when calling CustomObjectsApi->delete_cluster_custom_object_0: {e}")
        clean_with_created_objects(created_objects)
        assert False


def check_vs_content_deleted(vs_content_name, created_objects):
    """
    if volume snapshot content vs_content_name  exists ,  assert
    """
    if keep_objects:
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
    clean_with_created_objects(created_objects)
    assert False


def delete_vs(vs_name, created_objects):
    """
    delete volume snapshot vs_name
    """
    if keep_objects:
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
        clean_with_created_objects(created_objects)
        assert False


def check_vs_deleted(vs_name, created_objects):
    """
    if volume snapshot vs_name exists , it asserts
    """
    if keep_objects:
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
    clean_with_created_objects(created_objects)
    assert False


def delete_vs_class(vs_class_name, created_objects):
    """
    deletes volume snapshot class vs_class_name
    """
    if keep_objects:
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
        LOGGER.error(f"Exception when calling CustomObjectsApi->delete_cluster_custom_object_0: {e}")
        clean_with_created_objects(created_objects)
        assert False


def check_vs_class_deleted(vs_class_name, created_objects):
    """
    if volume snapshot class vs_class_name  exists ,  assert
    """
    if keep_objects:
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
        clean_with_created_objects(created_objects)
        assert False
    except ApiException:
        LOGGER.info(f"Volume Snapshot Class Delete : {vs_class_name} deletion confirmed")


def delete_ds(ds_name, created_objects):
    if keep_objects:
        return
    api_instance = client.AppsV1Api()

    try:
        api_response = api_instance.delete_namespaced_daemon_set(name=ds_name, namespace=namespace_value)
        LOGGER.debug(api_response)
        LOGGER.info(f"Daemon Set Delete : {ds_name} deleted")
        created_objects["ds"].remove(ds_name)
    except ApiException as e:
        LOGGER.error(f"Exception when calling AppsV1Api->delete_namespaced_daemon_set: {e}")
        clean_with_created_objects(created_objects)
        assert False


def check_ds_deleted(ds_name, created_objects):
    if keep_objects:
        return

    api_instance = client.AppsV1Api()
    try:
        api_response = api_instance.read_namespaced_daemon_set(name=ds_name, namespace=namespace_value)
        LOGGER.debug(api_response)
        LOGGER.error(f"Daemon Set Delete : {ds_name} is not deleted , asserting")
        clean_with_created_objects(created_objects)
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
            if len(volume_handle)==3:
                fileset_name = ""
            elif len(volume_handle)<=4:
                fileset_name= volume_handle[2][12:]
            else:
                fileset_name= volume_handle[5]
            if fileset_name == "":
                fileset_name = "LW"
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_persistent_volume: {e}")
            LOGGER.info(f'PV {pv_name} does not exists on cluster')

    if volume_name is not None and fileset_name is None:
        LOGGER.error(f"Not able to find fileset name for PV {volume_name}")
        clean_with_created_objects(created_objects)
        assert False

    return fileset_name
