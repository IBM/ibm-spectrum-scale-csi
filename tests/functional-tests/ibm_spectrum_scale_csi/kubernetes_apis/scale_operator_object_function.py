import time
import logging
import string
import random
from kubernetes import client
from kubernetes.client.rest import ApiException
LOGGER = logging.getLogger()


def set_namespace_value(namespace_name):
    """
    Make namespace as global to be used in later functions

    Args:
        param1: namespace_name - namespace name

    Returns:
       None

    Raises:
        None

    """
    global namespace_value
    namespace_value = namespace_name


def create_custom_object(custom_object_body, stateful_set_not_created):
    """
    Create custom object and waits until stateful sets are created.

    Args:
       param1: custom_object_body - custom object body
       param2: stateful_set_not_created - for operator testcases

    Returns:
       None

    Raises:
        Raises an exception on kubernetes client api failure and asserts

    """
    custom_object_api_instance = client.CustomObjectsApi()
    try:
        custom_object_api_response = custom_object_api_instance.create_namespaced_custom_object(
            group="csi.ibm.com",
            version="v1",
            namespace=namespace_value,
            plural="csiscaleoperators",
            body=custom_object_body,
            pretty=True
        )
        LOGGER.debug(custom_object_api_response)
        LOGGER.info("SpectrumScale CSI custom object created")
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CustomObjectsApi->create_namespaced_custom_object: {e}")
        assert False

    num = 0
    while (num < 30):
        read_statefulset_api_instance = client.AppsV1Api()
        try:
            read_statefulset_api_response = read_statefulset_api_instance.read_namespaced_stateful_set(
                name="ibm-spectrum-scale-csi-attacher", namespace=namespace_value, pretty=True)
            LOGGER.debug(str(read_statefulset_api_response))
            LOGGER.info("waiting for statefulsets")
            ready_replicas = read_statefulset_api_response.status.ready_replicas
            replicas = read_statefulset_api_response.status.replicas
            if ready_replicas == replicas:
                if stateful_set_not_created is True:
                    LOGGER.error("Statefulsets should not have been created")
                    assert False
                else:
                    return
            num = num + 1
            time.sleep(20)
        except ApiException:
            LOGGER.info("waiting for statefulsets")
            num = num+1
            time.sleep(20)

    if stateful_set_not_created is True:
        LOGGER.info("Expected Failure ,testcase is passed")
        return
    LOGGER.error("problem while creating custom object, (statefulsets are not created)")
    assert False


def delete_custom_object():
    """
    delete custom object  ( csiscaleoperator )

    Args:
       None

    Returns:
       None

    Raises:
        Raises an exception on kubernetes client api failure and asserts

    """
    delete_co_api_instance = client.CustomObjectsApi()
    delete_co_body = client.V1DeleteOptions()

    try:
        delete_co_api_response = delete_co_api_instance.delete_namespaced_custom_object(
            group="csi.ibm.com",
            version="v1",
            namespace=namespace_value,
            plural="csiscaleoperators",
            body=delete_co_body,
            grace_period_seconds=0,
            name="ibm-spectrum-scale-csi")
        LOGGER.debug(str(delete_co_api_response))
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CustomObjectsApi->delete_namespaced_custom_object: {e}")
        assert False


def check_scaleoperatorobject_is_deleted():
    """
    check  csiscaleoperator deleted or not
    if csiscaleoperator not deleted in 300 seconds , asserts
    """
    count = 30
    list_co_api_instance = client.CustomObjectsApi()
    while (count > 0):
        try:
            list_co_api_response = list_co_api_instance.get_namespaced_custom_object(group="csi.ibm.com",
                                                                                     version="v1",
                                                                                     namespace=namespace_value,
                                                                                     plural="csiscaleoperators",
                                                                                     name="ibm-spectrum-scale-csi"
                                                                                     )
            LOGGER.info("Waiting for custom object deletion")
            LOGGER.debug(str(list_co_api_response))
            count = count-1
            time.sleep(20)
        except ApiException:
            LOGGER.info("SpectrumScale CSI custom object has been deleted")
            return

    LOGGER.error("SpectrumScale CSI custom object is not deleted")
    assert False


def check_scaleoperatorobject_is_deployed(csiscaleoperator_name="ibm-spectrum-scale-csi"):
    """
    Checks csiscaleoperator exists or not

    Args:
       None

    Returns:
       return True  , if csiscaleoperator exists
       return False , if csiscaleoperator does not exist

    Raises:
        None

    """
    read_co_api_instance = client.CustomObjectsApi()
    try:
        read_co_api_response = read_co_api_instance.get_namespaced_custom_object(group="csi.ibm.com",
                                                                                 version="v1",
                                                                                 namespace=namespace_value,
                                                                                 plural="csiscaleoperators",
                                                                                 name=csiscaleoperator_name
                                                                                 )
        LOGGER.debug(str(read_co_api_response))
        LOGGER.info("SpectrumScale CSI custom object exists")
        return True
    except ApiException:
        LOGGER.info("SpectrumScale CSI custom object does not exist")
        return False


def check_scaleoperatorobject_statefulsets_state(stateful_name):
    """
    Checks statefulset exists or not
    if not exists , It asserts
    if exists :
        Checks statfulset is up or not
        if statefulsets not up in 120 seconds , it asserts

    Args:
       param1: stateful_name - statefulset name to check

    Returns:
       None

    Raises:
        Raises an exception on kubernetes client api failure.

    """
    read_statefulsets_api_instance = client.AppsV1Api()
    num = 0
    while (num < 30):
        try:
            read_statefulsets_api_response = read_statefulsets_api_instance.read_namespaced_stateful_set(
                name=stateful_name, namespace=namespace_value, pretty=True)
            LOGGER.debug(read_statefulsets_api_response)
            ready_replicas = read_statefulsets_api_response.status.ready_replicas
            replicas = read_statefulsets_api_response.status.replicas
            if ready_replicas == replicas:
                LOGGER.info(f"CSI driver statefulset {stateful_name} is up")
                return
            num += 1
            time.sleep(10)
        except ApiException:
            num += 1
            time.sleep(10)
    LOGGER.info(f"CSI driver statefulset {stateful_name} does not exist")
    assert False


def check_scaleoperatorobject_daemonsets_state(csiscaleoperator_name="ibm-spectrum-scale-csi"):
    """
    Checks daemonset exists or not
    If not exists , It asserts
    if exists :
        Checks daemonset is running or not
        if not running , It asserts

    Args:
       None

    Returns:
       returns True, desired_number_scheduled , if daemonset pods are running
       returns False, desired_number_scheduled , if daemonset pods are not running

    Raises:
        Raises an exception on kubernetes client api failure.

    """
    read_daemonsets_api_instance = client.AppsV1Api()
    num = 0
    while (num < 15):
        try:
            read_daemonsets_api_response = read_daemonsets_api_instance.read_namespaced_daemon_set(
                name=csiscaleoperator_name, namespace=namespace_value, pretty=True)
            LOGGER.debug(read_daemonsets_api_response)
            current_number_scheduled = read_daemonsets_api_response.status.current_number_scheduled
            desired_number_scheduled = read_daemonsets_api_response.status.desired_number_scheduled
            number_available = read_daemonsets_api_response.status.number_available
            if number_available == current_number_scheduled == desired_number_scheduled:
                LOGGER.info("CSI driver daemonset ibm-spectrum-scale-csi's pods are Running")
                return True, desired_number_scheduled

            time.sleep(20)
            num += 1
            LOGGER.info("waiting for daemonsets")
        except ApiException:
            time.sleep(20)
            num += 1
            LOGGER.info("waiting for daemonsets")

    LOGGER.error(
        "Expected CSI driver daemonset ibm-spectrum-scale-csi's pods are not Running")
    return False, desired_number_scheduled


def randomStringDigits(stringLength=6):
    """Generate a random string of letters and digits """
    lettersAndDigits = string.ascii_letters + string.digits
    return ''.join(random.choice(lettersAndDigits) for i in range(stringLength))


def randomString(stringLength=10):
    """Generate a random string of fixed length """
    letters = string.ascii_lowercase
    return ''.join(random.choice(letters) for i in range(stringLength))


def get_scaleoperatorobject_values(namespace_value, csiscaleoperator_name="ibm-spectrum-scale-csi"):
    read_cr_api_instance = client.CustomObjectsApi()
    try:
        read_cr_api_response = read_cr_api_instance.get_namespaced_custom_object(group="csi.ibm.com",
                                  version="v1", namespace=namespace_value, plural="csiscaleoperators", name=csiscaleoperator_name)
        LOGGER.debug(str(read_cr_api_response))
        return read_cr_api_response
    except ApiException:
        return False
