import time
import logging
import yaml
from kubernetes import client
from kubernetes.client.rest import ApiException
LOGGER = logging.getLogger()


def set_global_namespace_value(namespace_name):
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


def create_namespace():
    """
    Create namespace namespace_value(global parameter)

    Args:
       None

    Returns:
       None

    Raises:
        Raises an exception on kubernetes client api failure and asserts

    """
    namespace_api_instance = client.CoreV1Api()
    namespace_metadata = client.V1ObjectMeta(
        name=namespace_value,
        labels={"product": "ibm-spectrum-scale-csi"}
    )
    namespace_body = client.V1Namespace(
        api_version="v1", kind="Namespace", metadata=namespace_metadata)
    try:
        LOGGER.info(f'Creating new Namespace {namespace_value}')
        namespace_api_response = namespace_api_instance.create_namespace(
            body=namespace_body, pretty=True)
        LOGGER.debug(str(namespace_api_response))
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CoreV1Api->create_namespace: {e}")
        assert False


def create_deployment(body):
    """
    Create IBM Spectrum Scale CSI Operator deployment object using operator.yaml file

    Args:
        None

    Returns:
       None

    Raises:
        Raises an exception on kubernetes client api failure and asserts

    """
    deployment_apps_api_instance = client.AppsV1Api()
    try:
        LOGGER.info("Creating Operator Deployment")
        deployment_apps_api_response = deployment_apps_api_instance.create_namespaced_deployment(
            namespace=namespace_value, body=body)
        LOGGER.debug(str(deployment_apps_api_response))
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling RbacAuthorizationV1Api->create_namespaced_deployment: {e}")
        assert False


def create_cluster_role(body):
    """
    Create IBM Spectrum Scale CSI Operator cluster role using role.yaml file

    Args:
       None

    Returns:
       None

    Raises:
        Raises an exception on kubernetes client api failure and asserts

    """
    cluster_role_api_instance = client.RbacAuthorizationV1Api()
    try:
        LOGGER.info("Creating ibm-spectrum-scale-csi-operator ClusterRole ")
        cluster_role_api_response = cluster_role_api_instance.create_cluster_role(
            body=body, pretty=True)
        LOGGER.debug(str(cluster_role_api_response))
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling RbacAuthorizationV1Api->create_cluster_role: {e}")
        assert False


def create_cluster_role_binding(body):
    """
    Create IBM Spectrum Scale CSI Operator ClusterRoleBinding object using role_binding.yaml

    Args:
       None

    Returns:
       None

    Raises:
        Raises an exception on kubernetes client api failure and asserts

    """
    cluster_role_binding_api_instance = client.RbacAuthorizationV1Api()
    body["subjects"][0]["namespace"] = namespace_value
    try:
        LOGGER.info("creating cluster role binding")
        cluster_role_binding_api_response = cluster_role_binding_api_instance.create_cluster_role_binding(
            body=body, pretty=True)
        LOGGER.debug(cluster_role_binding_api_response)
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling RbacAuthorizationV1Api->create_cluster_role_binding: {e}")
        assert False


def create_service_account(body):
    """
    Create IBM Spectrum Scale CSI Operator ServiceAccount using service_account.yaml

    Args:
       None

    Returns:
       None

    Raises:
        Raises an exception on kubernetes client api failure and asserts

    """
    service_account_api_instance = client.CoreV1Api()
    body["metadata"]["namespace"] = namespace_value
    try:
        LOGGER.info("Creating ibm-spectrum-scale-csi-operator ServiceAccount")
        service_account_api_response = service_account_api_instance.create_namespaced_service_account(
            namespace=namespace_value, body=body, pretty=True)
        LOGGER.debug(str(service_account_api_response))
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CoreV1Api->create_namespaced_service_account: {e}")
        assert False


def create_crd(body):
    """
    Create IBM Spectrum Scale CSI Operator CRD (Custom Resource Defination) Object

    Args:
       None

    Returns:
       None

    Raises:
        Raises an ValueError exception but it is expected. hence we pass.

    """
    version = body["apiVersion"].split("/")
    crd_version = version[1]
    LOGGER.info(f"CRD apiVersion is {crd_version}")
    custom_object_api_instance = client.CustomObjectsApi()
    try:
        custom_object_api_response = custom_object_api_instance.create_cluster_custom_object(
            group="apiextensions.k8s.io",
            version=crd_version,
            plural="customresourcedefinitions",
            body=body,
            pretty=True
        )
        LOGGER.debug(custom_object_api_response)
        LOGGER.info(f"Creating IBM SpectrumScale CRD object using csiscaleoperators.csi.ibm.com.crd.yaml file")
    except ValueError as e:
        LOGGER.error(
            f"Exception when calling CustomObjectsApi->create_namespaced_custom_object: {e}")
        LOGGER.info(
            "while there is valuerror expection,but CRD created successfully")
        assert False


def delete_crd():
    """
    Delete existing IBM Spectrum Scale CSI Operator CRD (Custom Resource Defination) Object

    Args:
       None

    Returns:
       None

    Raises:
        Raises an exception on kubernetes client api failure and asserts

    """
    delete_crd_api_instance = client.ApiextensionsV1beta1Api()
    try:
        delete_crd_api_response = delete_crd_api_instance.delete_custom_resource_definition(
            name="csiscaleoperators.csi.ibm.com", pretty=True)
        LOGGER.debug(str(delete_crd_api_response))
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling ApiextensionsV1beta1Api->delete_custom_resource_definition: {e}")
        assert False


def delete_namespace():
    """
    Delete IBM Spectrum Scale CSI Operator namespace

    Args:
       None

    Returns:
       None

    Raises:
        Raises an exception on kubernetes client api failure and asserts

    """
    delete_namespace_api_instance = client.CoreV1Api()
    try:
        delete_namespace_api_response = delete_namespace_api_instance.delete_namespace(
            name=namespace_value, pretty=True)
        LOGGER.debug(str(delete_namespace_api_response))
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CoreV1Api->delete_namespace: {e}")
        assert False


def delete_deployment():
    """
    Delete IBM Spectrum Scale CSI Operator Deployment object from Operator namespace

    Args:
       None

    Returns:
       None

    Raises:
        Raises an exception on kubernetes client api failure and asserts

    """
    delete_deployment_api_instance = client.AppsV1Api()
    try:
        delete_deployment_api_response = delete_deployment_api_instance.delete_namespaced_deployment(
            name="ibm-spectrum-scale-csi-operator", namespace=namespace_value, pretty=True)
        LOGGER.debug(str(delete_deployment_api_response))
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling ExtensionsV1beta1Api->delete_namespaced_deployment: {e}")
        assert False


def delete_service_account(service_account_name):
    """
    Delete IBM Spectrum Scale CSI Operator ServiceAccount from Operator namespace

    Args:
       param1: service_accout_name - service account name to be deleted

    Returns:
       None

    Raises:
        Raises an exception on kubernetes client api failure and asserts

    """
    delete_service_account_api_instance = client.CoreV1Api()
    try:
        delete_service_account_api_response = delete_service_account_api_instance.delete_namespaced_service_account(
            name=service_account_name, namespace=namespace_value, pretty=True)
        LOGGER.debug(str(delete_service_account_api_response))
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CoreV1Api->delete_namespaced_service_account: {e}")
        assert False


def delete_cluster_role(cluster_role_name):
    """
    Delete IBM Spectrum Scale CSI Operator ClusterRole Object

    Args:
       param1: cluster_role_name - cluster role name to be deleted

    Returns:
       None

    Raises:
        Raises an exception on kubernetes client api failure and asserts

    """
    delete_cluster_role_api_instance = client.RbacAuthorizationV1Api()
    try:
        delete_cluster_role_api_response = delete_cluster_role_api_instance.delete_cluster_role(
            name=cluster_role_name, pretty=True)
        LOGGER.debug(str(delete_cluster_role_api_response))
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling RbacAuthorizationV1Api->delete_cluster_role: {e}")
        assert False


def delete_cluster_role_binding(cluster_role_binding_name):
    """
    Delete IBM Spectrum Scale CSI Operator ClusterRoleBinding Object

    Args:
       param1: cluster_role_name - cluster role name to be deleted

    Returns:
       None

    Raises:
        Raises an exception on kubernetes client api failure and asserts

    """
    delete_cluster_role_binding_api_instance = client.RbacAuthorizationV1Api()
    try:
        delete_cluster_role_binding_api_response = delete_cluster_role_binding_api_instance.delete_cluster_role_binding(
            name=cluster_role_binding_name, pretty=True)
        LOGGER.debug(delete_cluster_role_binding_api_response)
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling RbacAuthorizationV1Api->delete_cluster_role_binding: {e}")
        assert False


def check_crd_deleted():
    """
    Function for checking CRD (Custom Resource Defination) is deleted or not
    If CRD is not deleted in 60 seconds,function asserts

    Args:
       None

    Returns:
       None

    Raises:
        Raises an exception on kubernetes client api failure and asserts

    """
    count = 12
    list_crd_api_instance = client.ApiextensionsV1beta1Api()
    while (count > 0):
        try:
            list_crd_api_response = list_crd_api_instance.read_custom_resource_definition(
                pretty=True, name="ibm-spectrum-scale-csi")
            LOGGER.debug(list_crd_api_response)
            LOGGER.info("still deleting crd")
            count -= 1
            time.sleep(5)

        except ApiException:
            LOGGER.info("crd deleted")
            return

    LOGGER.error("crd is not deleted")
    assert False


def check_namespace_deleted():
    """
    Function for checking namespace object is deleted or not
    If namespace is not deleted in 120 seconds, Function asserts

    Raises:
        Raises an exception on kubernetes client api failure and asserts

    """
    count = 24
    list_namespace_api_instance = client.CoreV1Api()
    while (count > 0):
        try:
            list_namespace_api_response = list_namespace_api_instance.read_namespace(
                name=namespace_value, pretty=True)
            LOGGER.debug(str(list_namespace_api_response))
            LOGGER.info(f'Still deleting namespace {namespace_value}')
            count = count-1
            time.sleep(5)
        except ApiException:
            LOGGER.info(f'namespace {namespace_value} is deleted')
            return

    LOGGER.error(f'namespace  {namespace_value} is not deleted')
    assert False


def check_deployment_deleted():
    """
    Function for checking deployment is deleted or not
    If deployment is not deleted in 30 seconds, Function asserts

    Raises:
        Raises an exception on kubernetes client api failure and asserts

    """
    count = 6
    api_instance = client.AppsV1Api()
    while (count > 0):
        try:
            api_response = api_instance.read_namespaced_deployment(
                name="ibm-spectrum-scale-csi-operator", namespace=namespace_value, pretty=True)
            LOGGER.debug(str(api_response))
            LOGGER.info('Still Deleting ibm-spectrum-scale-csi-operator deployment')
            count = count-1
            time.sleep(5)
        except ApiException:
            LOGGER.info("Deployment ibm-spectrum-scale-csi-operator is deleted")
            return

    LOGGER.error("deployment is not deleted")
    assert False


def check_service_account_deleted(service_account_name):
    """
    Function to check ServiceAccount is deleted or not
    If ServiceAccount is not deleted in 30 seconds, Function asserts

    Args:
       param1: service_accout_name - service account name to be checked

    Raises:
        Raises an exception on kubernetes client api failure and asserts

    """
    count = 6
    api_instance = client.CoreV1Api()
    while (count > 0):
        try:
            api_response = api_instance.read_namespaced_service_account(
                name=service_account_name, namespace=namespace_value, pretty=True)
            LOGGER.debug(str(api_response))
            LOGGER.info(f'Still deleting ServiceAccount {service_account_name}')
            count = count-1
            time.sleep(5)
        except ApiException:
            LOGGER.info(f'ServiceAccount {service_account_name} is deleted')
            return

    LOGGER.error("service account is not deleted")
    assert False


def check_cluster_role_deleted(cluster_role_name):
    """
    Function to check ClusterRole is deleted or not
    If ClusterRole not deleted in 30 seconds, Function asserts

    Args:
       param1: cluster_role_name - cluster role name to be checked

    Raises:
        Raises an exception on kubernetes client api failure and asserts

    """
    count = 6
    api_instance = client.RbacAuthorizationV1Api()
    while (count > 0):
        try:
            api_response = api_instance.read_cluster_role(
                name=cluster_role_name, pretty=True)
            LOGGER.debug(str(api_response))
            LOGGER.info(f'Still deleting ClusterRole {cluster_role_name} ')
            count = count-1
            time.sleep(5)
        except ApiException:
            LOGGER.info(f'ClusterRole {cluster_role_name} is deleted')
            return

    LOGGER.error(f'ClusterRole {cluster_role_name} is not deleted')
    assert False


def check_cluster_role_binding_deleted(cluster_role_binding_name):
    """
    Function to check ClusterRoleBinding is deleted or not
    If ClusterRoleBinding is not deleted in 30 seconds, Function asserts

    Args:
       param1: cluster_role_binding_name - cluster role binding name to be checked

    Raises:
        Raises an exception on kubernetes client api failure and asserts

    """
    count = 6
    api_instance = client.RbacAuthorizationV1Api()
    while (count > 0):
        try:
            api_response = api_instance.read_cluster_role_binding(
                name=cluster_role_binding_name, pretty=True)
            LOGGER.debug(str(api_response))
            LOGGER.info(f'Still deleting ClusterRoleBinding {cluster_role_binding_name}')
            count = count-1
            time.sleep(5)
        except ApiException:
            LOGGER.info(f'ClusterRoleBinding {cluster_role_binding_name} is deleted')
            return

    LOGGER.error(f'ClusterRoleBinding {cluster_role_binding_name} is not deleted')
    assert False


def check_crd_exists():
    """
    Checks custom resource defination exists or not

    Args:
       None

    Returns:
       return True  , if crd exists
       return False , if crd does not exists

    Raises:
        None

    """
    crd_name = "csiscaleoperators.csi.ibm.com"
    read_crd_api_instance = client.ApiextensionsV1beta1Api()
    try:
        read_crd_api_response = read_crd_api_instance.read_custom_resource_definition(
            pretty=True, name=crd_name)
        LOGGER.debug(str(read_crd_api_response))
        LOGGER.info(f"crd  {crd_name} exists")
        return True
    except ApiException:
        LOGGER.info(f"crd {crd_name} does not exist")
        return False


def check_namespace_exists():
    """
    Checks namespace namespace_value exists or not

    Args:
       None

    Returns:
       return True  , if namespace exists
       return False , if namespace does not exists

    Raises:
        None

    """
    read_namespace_api_instance = client.CoreV1Api()
    try:
        read_namespace_api_response = read_namespace_api_instance.read_namespace(
            name=namespace_value, pretty=True)
        LOGGER.debug(str(read_namespace_api_response))
        LOGGER.info("namespace exists")
        return True
    except ApiException:
        LOGGER.info("namespace does not exists")
        return False


def check_deployment_exists():
    """
    Checks deployment exists or not

    Args:
       None

    Returns:
       return True  , if deployment exists
       return False , if deployment does not exists

    Raises:
        None

    """
    read_deployment_api_instance = client.AppsV1Api()
    try:
        read_deployment_api_response = read_deployment_api_instance.read_namespaced_deployment(
            name="ibm-spectrum-scale-csi-operator", namespace=namespace_value, pretty=True)
        LOGGER.debug(str(read_deployment_api_response))
        LOGGER.info("deployment exists")
        return True
    except ApiException:
        LOGGER.info("deployment does not exists")
        return False


def check_service_account_exists(service_account_name):
    """
    Checks service account exists or not

    Args:
       None

    Returns:
       return True  , if service account exists
       return False , if service account does not exists

    Raises:
        None

    """
    api_instance = client.CoreV1Api()
    try:
        api_response = api_instance.read_namespaced_service_account(
            name=service_account_name, namespace=namespace_value, pretty=True)
        LOGGER.debug(str(api_response))
        LOGGER.info("Service account exists")
        return True
    except ApiException:
        LOGGER.info("Service account does not exists")
        return False


def check_cluster_role_exists(cluster_role_name):
    """
    Checks cluster role exists or not

    Args:
       None

    Returns:
       return True  , if cluster role exists
       return False , if cluster role does not exists

    Raises:
        None

    """
    api_instance = client.RbacAuthorizationV1Api()
    try:
        api_response = api_instance.read_cluster_role(
            name=cluster_role_name, pretty=True)
        LOGGER.debug(str(api_response))
        LOGGER.info("cluster role exists")
        return True
    except ApiException:
        LOGGER.info("cluster role does not exists")
        return False


def check_cluster_role_binding_exists(cluster_role_binding_name):
    """
    Checks cluster role binding exists or not

    Args:
       None

    Returns:
       return True  , if cluster role binding exists
       return False , if cluster role binding does not exists

    Raises:
        None

    """
    api_instance = client.RbacAuthorizationV1Api()
    try:
        api_response = api_instance.read_cluster_role_binding(
            name=cluster_role_binding_name, pretty=True)
        LOGGER.debug(str(api_response))
        LOGGER.info("cluster role binding exists")
        return True
    except ApiException:
        LOGGER.info("cluster role binding does not exists")
        return False

def get_operator_pod_name():
    try:
        pod_list_api_instance = client.CoreV1Api()
        pod_list_api_response = pod_list_api_instance.list_namespaced_pod(
            namespace=namespace_value, pretty=True, field_selector="spec.serviceAccountName=ibm-spectrum-scale-csi-operator")
        operator_pod_name = pod_list_api_response.items[0].metadata.name
        LOGGER.debug(str(pod_list_api_response))
        return operator_pod_name
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CoreV1Api->list_namespaced_pod: {e}")
        assert False


def get_operator_image():
    pod_name = get_operator_pod_name()
    api_instance = client.CoreV1Api()
    try:
        api_response = api_instance.read_namespaced_pod(
                name=pod_name, namespace=namespace_value, pretty=True)
        LOGGER.info("\nCSI operator image    : " + api_response.status.container_statuses[-1].image)
        LOGGER.info("CSI operator image id : " + api_response.status.container_statuses[-1].image_id)
    except ApiException as e:
        LOGGER.info("Unable to get operator image")

