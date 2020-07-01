import time
import re
import logging
from kubernetes import client
from kubernetes.client.rest import ApiException
import utils.fileset_functions as ff
LOGGER = logging.getLogger()

def set_test_namespace_value(namespace_name=None):
    """ sets the test namespace global for use in later functions"""
    global namespace_value
    namespace_value = namespace_name

def create_vs_class(vs_class_name,body_params):
    class_body =  {
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
        LOGGER.info(f"volume snapshot class {vs_class_name} is created")
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CustomObjectsApi->create_namespaced_custom_object: {e}")
        assert False

def delete_vs_class(vs_class_name):
    custom_object_api_instance = client.CustomObjectsApi()
    try:
        custom_object_api_response = custom_object_api_instance.delete_cluster_custom_object(
            group="snapshot.storage.k8s.io",
            version="v1beta1",
            plural="volumesnapshotclasses",
            name=vs_class_name
        )
        LOGGER.debug(custom_object_api_response)
        LOGGER.info(f"volume snapshot class {vs_class_name} deleted")
    except ApiException as e:
        LOGGER.error("Exception when calling CustomObjectsApi->delete_cluster_custom_object_0: %s\n" % e)
        assert False   

def check_vs_class(vs_class_name):

    api_instance = client.CustomObjectsApi()
    try:
        api_response = api_instance.get_cluster_custom_object(
            group="snapshot.storage.k8s.io",
            version="v1beta1",
            plural="volumesnapshotclasses",
            name=vs_class_name
        )
        LOGGER.debug(api_response)
        LOGGER.info(f"volume snapshot class {vs_class_name} exists")
        return True
    except ApiException as e:
        LOGGER.info(f"volume snapshot class {vs_class_name} does not exists")
        return False  

def check_vs_class_deleted(vs_class_name):
    api_instance = client.CustomObjectsApi()
    try:
        api_response = api_instance.get_cluster_custom_object(
            group="snapshot.storage.k8s.io",
            version="v1beta1",
            plural="volumesnapshotclasses",
            name=vs_class_name
        )
        LOGGER.debug(api_response)
        LOGGER.error(f"volume snapshot class {vs_class_name} is not deleted , asserting")
        assert False
    except ApiException as e:
        LOGGER.info(f"volume snapshot class {vs_class_name} deletion confirmed")

def create_vs(vs_name,vs_class_name,pvc_name):        

    class_body =  {
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
        LOGGER.info(f"volume snapshot {vs_name} is created")
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CustomObjectsApi->create_namespaced_custom_object: {e}")
        assert False


def delete_vs(vs_name):

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
        LOGGER.info(f"volume snapshot {vs_name} deleted")
    except ApiException as e:
        LOGGER.error("Exception when calling CustomObjectsApi->delete_cluster_custom_object: %s\n" % e)
        assert False


def check_vs(vs_name):

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
        LOGGER.info(f"volume snapshot {vs_name} exists")
        return True
    except ApiException as e:
        LOGGER.info(f"volume snapshot {vs_name} does not exists")
        return False


def check_vs_deleted(vs_name):
    time.sleep(10)
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
        LOGGER.error(f"volume snapshot {vs_name} is not deleted , asserting")
        assert False
    except ApiException as e:
        LOGGER.info(f"volume snapshot {vs_name} deletion confirmed")

def check_vs_detail(vs_name,pvc_name,data):

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
        LOGGER.info(f"volume snapshot {vs_name} exists")
    except ApiException as e:
        LOGGER.info(f"volume snapshot {vs_name} does not exists")
        assert False
    
    uid_name = api_response["metadata"]["uid"]
    snapcontent_name = "snapcontent-" + uid_name
    snapshot_name = "snapshot-" + uid_name
    time.sleep(5)
    try:
        api_response = api_instance.get_cluster_custom_object(
            group="snapshot.storage.k8s.io",
            version="v1beta1",
            plural="volumesnapshotcontents",
            name=snapcontent_name
        )
        LOGGER.debug(api_response)
        LOGGER.info(f"volume snapshot content {snapcontent_name} exists")
    except ApiException as e:
        LOGGER.error(f"volume snapshot content {snapcontent_name} does not exists")
        assert False

    api_instance = client.CoreV1Api()
    try:
        api_response = api_instance.read_namespaced_persistent_volume_claim(
            name=pvc_name, namespace=namespace_value, pretty=True)
        LOGGER.debug(str(api_response))
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CoreV1Api->read_namespaced_persistent_volume_claim: {e}")
        LOGGER.info(f"PVC {pvc_name} does not exists on the cluster")
        assert False
	
    volume_name = api_response.spec.volume_name
    time.sleep(5)
    if ff.check_snapshot(data,snapshot_name,volume_name):
        LOGGER.info(f"snapshot {snapshot_name} exists for {volume_name}")
    else:
        LOGGER.error(f"snapshot {snapshot_name} does not exists for {volume_name}")
        assert False 
