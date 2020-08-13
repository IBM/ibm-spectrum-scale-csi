import time
import logging
from kubernetes import client
from kubernetes.client.rest import ApiException
import utils.fileset_functions as ff

if __name__ == "__main__":
    from utils.driver import clean_with_created_objects
LOGGER = logging.getLogger()


def set_test_namespace_value(namespace_name=None):
    """ sets the test namespace global for use in later functions"""
    global namespace_value
    namespace_value = namespace_name


def set_keep_objects(keep_object):
    """ sets the keep_objects global for use in later functions"""
    global keep_objects
    keep_objects = keep_object


def create_vs_class(vs_class_name, body_params):
    """
    create volume snapshot class with vs_class_name
    body_params contains configurable parameters
    """
    class_body = {
        "apiVersion": "snapshot.storage.k8s.io/v1beta1",
        "kind": "VolumeSnapshotClass",
        "metadata": {
            "name": vs_class_name
        },
        "driver": "spectrumscale.csi.ibm.com",
        "deletionPolicy": body_params["deletionPolicy"]
    }

    custom_object_api_instance = client.CustomObjectsApi()
    try:
        custom_object_api_response = custom_object_api_instance.create_cluster_custom_object(
            group="snapshot.storage.k8s.io",
            version="v1beta1",
            plural="volumesnapshotclasses",
            body=class_body,
            pretty=True
        )
        LOGGER.debug(custom_object_api_response)
        LOGGER.info(f"Volume Snapshot Class Create : {vs_class_name} is created with {body_params}")
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CustomObjectsApi->create_namespaced_custom_object: {e}")
        assert False


def delete_vs_class(vs_class_name):
    """
    deletes volume snapshot class vs_class_name
    """
    if keep_objects:
        return
    custom_object_api_instance = client.CustomObjectsApi()
    try:
        custom_object_api_response = custom_object_api_instance.delete_cluster_custom_object(
            group="snapshot.storage.k8s.io",
            version="v1beta1",
            plural="volumesnapshotclasses",
            name=vs_class_name
        )
        LOGGER.debug(custom_object_api_response)
        LOGGER.info(f"Volume Snapshot Class Delete : {vs_class_name} deleted")
    except ApiException as e:
        LOGGER.error(f"Exception when calling CustomObjectsApi->delete_cluster_custom_object_0: {e}")
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
            version="v1beta1",
            plural="volumesnapshotclasses",
            name=vs_class_name
        )
        LOGGER.debug(api_response)
        LOGGER.info(f"Volume Snapshot Class Check : {vs_class_name} exists")
        return True
    except ApiException:
        LOGGER.info(f"volume snapshot class {vs_class_name} does not exists")
        return False


def check_vs_class_deleted(vs_class_name):
    """
    if volume snapshot class vs_class_name  exists ,  assert
    """
    if keep_objects:
        return
    api_instance = client.CustomObjectsApi()
    try:
        api_response = api_instance.get_cluster_custom_object(
            group="snapshot.storage.k8s.io",
            version="v1beta1",
            plural="volumesnapshotclasses",
            name=vs_class_name
        )
        LOGGER.debug(api_response)
        LOGGER.error(f"Volume Snapshot Class Delete : {vs_class_name} is not deleted , asserting")
        assert False
    except ApiException:
        LOGGER.info(f"Volume Snapshot Class Delete : {vs_class_name} deletion confirmed")


def create_vs(vs_name, vs_class_name, pvc_name):
    """
    create volume snapshot vs_name using volume snapshot class vs_class_name
    and pvc pvc_name
    """
    class_body = {
        "apiVersion": "snapshot.storage.k8s.io/v1beta1",
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
            version="v1beta1",
            plural="volumesnapshots",
            body=class_body,
            namespace=namespace_value,
            pretty=True
        )
        LOGGER.debug(custom_object_api_response)
        LOGGER.info(f"Volume Snapshot Create : volume snapshot {vs_name} is created for {pvc_name}")
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CustomObjectsApi->create_namespaced_custom_object: {e}")
        assert False


def create_vs_from_content(vs_name, vs_content_name):
    """
    create volume snapshot vs_name from volume snapshot content vs_content_name
    """
    class_body = {
        "apiVersion": "snapshot.storage.k8s.io/v1beta1",
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
            version="v1beta1",
            plural="volumesnapshots",
            body=class_body,
            namespace=namespace_value,
            pretty=True
        )
        LOGGER.debug(custom_object_api_response)
        LOGGER.info(f"Volume Snapshot Create : volume snapshot {vs_name} is created from {vs_content_name}")
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CustomObjectsApi->create_namespaced_custom_object: {e}")
        assert False


def delete_vs(vs_name):
    """
    delete volume snapshot vs_name
    """
    if keep_objects:
        return
    custom_object_api_instance = client.CustomObjectsApi()
    try:
        custom_object_api_response = custom_object_api_instance.delete_namespaced_custom_object(
            group="snapshot.storage.k8s.io",
            version="v1beta1",
            plural="volumesnapshots",
            name=vs_name,
            namespace=namespace_value
        )
        LOGGER.debug(custom_object_api_response)
        LOGGER.info(f"Volume Snapshot Delete : {vs_name} deleted")
    except ApiException as e:
        LOGGER.error(f"Exception when calling CustomObjectsApi->delete_cluster_custom_object: {e}")
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
            version="v1beta1",
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


def check_vs_deleted(vs_name):
    """
    if volume snapshot vs_name exists , it asserts
    """
    if keep_objects:
        return
    api_instance = client.CustomObjectsApi()
    val = 0
    while val < 24:
        try:
            api_response = api_instance.get_namespaced_custom_object(
                group="snapshot.storage.k8s.io",
                version="v1beta1",
                plural="volumesnapshots",
                name=vs_name,
                namespace=namespace_value
            )
            LOGGER.debug(api_response)
            time.sleep(5)
            val += 1
        except ApiException:
            LOGGER.info(f"Volume Snapshot Delete : {vs_name} deletion confirmed")
            return
    LOGGER.error(f"Volume Snapshot Delete : {vs_name} is not deleted , asserting")
    assert False


def check_vs_detail_for_static(vs_name, created_objects):
    api_instance = client.CustomObjectsApi()
    try:
        api_response = api_instance.get_namespaced_custom_object(
            group="snapshot.storage.k8s.io",
            version="v1beta1",
            plural="volumesnapshots",
            name=vs_name,
            namespace=namespace_value
        )
        LOGGER.debug(api_response)
        LOGGER.info(f"Volume Snapshot Check : volume snapshot {vs_name} has been created")
    except ApiException:
        LOGGER.info(f"Volume Snapshot Check : volume snapshot {vs_name} does not exists")
        clean_with_created_objects(created_objects)
        assert False

    if check_snapshot_status(vs_name):
        LOGGER.info("volume snapshot status ReadyToUse is true")
    else:
        LOGGER.error("volume snapshot status ReadyToUse is not true")
        clean_with_created_objects(created_objects)
        assert False


def check_vs_detail(vs_name, pvc_name, body_params, created_objects):
    """
    checks volume snapshot vs_name exits , 
    checks volume snapshot content for vs_name is created
    check snapshot is created on spectrum scale
    """
    api_instance = client.CustomObjectsApi()
    try:
        api_response = api_instance.get_namespaced_custom_object(
            group="snapshot.storage.k8s.io",
            version="v1beta1",
            plural="volumesnapshots",
            name=vs_name,
            namespace=namespace_value
        )
        LOGGER.debug(api_response)
        LOGGER.info(f"Volume Snapshot Check : volume snapshot {vs_name} has been created")
    except ApiException as e:
        LOGGER.info(f"Volume Snapshot Check : volume snapshot {vs_name} does not exists")
        clean_with_created_objects(created_objects)
        assert False

    if check_snapshot_status(vs_name):
        LOGGER.info("volume snapshot status ReadyToUse is true")
    else:
        LOGGER.error("volume snapshot status ReadyToUse is not true")
        clean_with_created_objects(created_objects)
        assert False
    LOGGER.debug(api_response)

    uid_name = api_response["metadata"]["uid"]
    snapcontent_name = "snapcontent-" + uid_name
    snapshot_name = "snapshot-" + uid_name
    time.sleep(2)

    if not(check_vs_content(snapcontent_name)):
        clean_with_created_objects(created_objects)
        assert False

    volume_name = get_pv_name(pvc_name, created_objects)

    if ff.check_snapshot(snapshot_name, volume_name):
        LOGGER.info(f"snapshot {snapshot_name} exists for {volume_name}")
    else:
        LOGGER.error(f"snapshot {snapshot_name} does not exists for {volume_name}")
        clean_with_created_objects(created_objects)
        assert False

    """
    if body_params["deletionPolicy"] == "Retain" and not(keep_objects):
        custom_object_api_instance = client.CustomObjectsApi()
        try:
            custom_object_api_response = custom_object_api_instance.delete_cluster_custom_object(
                group="snapshot.storage.k8s.io",
                version="v1beta1",
                plural="volumesnapshotcontents",
                name=snapcontent_name
            )
            LOGGER.debug(custom_object_api_response)
            LOGGER.info(f"volume snapshot content {snapcontent_name} deleted")
        except ApiException as e:
            LOGGER.error(f"Exception when calling CustomObjectsApi->delete_cluster_custom_object: {e}")
            assert False
    """


def get_pv_name(pvc_name, created_objects):
    api_instance = client.CoreV1Api()
    try:
        api_response = api_instance.read_namespaced_persistent_volume_claim(
            name=pvc_name, namespace=namespace_value, pretty=True)
        LOGGER.debug(str(api_response))
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CoreV1Api->read_namespaced_persistent_volume_claim: {e}")
        LOGGER.info(f"PVC {pvc_name} does not exists on the cluster")
        clean_with_created_objects(created_objects)
        assert False

    return api_response.spec.volume_name


def check_snapshot_status(vs_name):
    """
    check status of volume snapshot vs_name
    if status True , return True
    else return False
    """
    api_instance = client.CustomObjectsApi()
    val = 0
    while val < 36:
        try:
            api_response = api_instance.get_namespaced_custom_object_status(
                group="snapshot.storage.k8s.io",
                version="v1beta1",
                plural="volumesnapshots",
                name=vs_name,
                namespace=namespace_value
            )
            LOGGER.debug(api_response)
            if "status" in api_response.keys() and "readyToUse" in api_response["status"].keys():
                if api_response["status"]["readyToUse"] is True:
                    return True
            time.sleep(5)
            val += 1
        except ApiException:
            time.sleep(5)
            val += 1
    return False


def create_vs_content(vs_content_name, vs_name, body_params):
    """
    create volume snapshot content with vs_content_name
    body_params contains configurable parameters
    """
    content_body = {
        "apiVersion": "snapshot.storage.k8s.io/v1beta1",
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
            version="v1beta1",
            plural="volumesnapshotcontents",
            body=content_body,
            pretty=True
        )
        LOGGER.debug(custom_object_api_response)
        LOGGER.info(f"Volume Snapshot Content Create : {vs_content_name} is created with {body_params}")
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CustomObjectsApi->create_namespaced_custom_object: {e}")
        assert False


def delete_vs_content(vs_content_name):
    """
    deletes volume snapshot content vs_content_name
    """
    if keep_objects:
        return
    custom_object_api_instance = client.CustomObjectsApi()
    try:
        custom_object_api_response = custom_object_api_instance.delete_cluster_custom_object(
            group="snapshot.storage.k8s.io",
            version="v1beta1",
            plural="volumesnapshotcontents",
            name=vs_content_name
        )
        LOGGER.debug(custom_object_api_response)
        LOGGER.info(f"Volume Snapshot Content Delete : {vs_content_name} deleted")
    except ApiException as e:
        LOGGER.error(f"Exception when calling CustomObjectsApi->delete_cluster_custom_object_0: {e}")
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
            version="v1beta1",
            plural="volumesnapshotcontents",
            name=vs_content_name
        )
        LOGGER.debug(api_response)
        LOGGER.info(f"Volume Snapshot Content Check : {vs_content_name} exists")
        return True
    except ApiException:
        LOGGER.info(f"Volume Snapshot content {vs_content_name} does not exists")
        return False


def check_vs_content_deleted(vs_content_name):
    """
    if volume snapshot content vs_content_name  exists ,  assert
    """
    if keep_objects:
        return
    api_instance = client.CustomObjectsApi()
    try:
        api_response = api_instance.get_cluster_custom_object(
            group="snapshot.storage.k8s.io",
            version="v1beta1",
            plural="volumesnapshotcontents",
            name=vs_content_name
        )
        LOGGER.debug(api_response)
        LOGGER.error(f"Volume Snapshot Content Delete : {vs_content_name} is not deleted , asserting")
        assert False
    except ApiException:
        LOGGER.info(f"Volume Snapshot Content Delete : {vs_content_name} deletion confirmed")
