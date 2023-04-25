import time
import logging
import copy
import re
import base64
import urllib3
from kubernetes import client, config
from kubernetes.client.rest import ApiException
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
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


def create_namespace(namespace_name):
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
        name=namespace_name,
        labels={"product": "ibm-spectrum-scale-csi"}
    )
    namespace_body = client.V1Namespace(
        api_version="v1", kind="Namespace", metadata=namespace_metadata)
    try:
        namespace_api_response = namespace_api_instance.create_namespace(
            body=namespace_body, pretty=True)
        LOGGER.debug(str(namespace_api_response))
        LOGGER.info(f'Namespace Create : {namespace_name} is created')
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CoreV1Api->create_namespace: {e}")
        assert False


def create_deployment(body):
    """
    Create IBM Storage Scale CSI Operator deployment object using operator.yaml file

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
            f"Exception when calling AppsV1Api->create_namespaced_deployment: {e}")
        assert False


def create_cluster_role(body):
    """
    Create IBM Storage Scale CSI Operator cluster role using role.yaml file

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
    Create IBM Storage Scale CSI Operator ClusterRoleBinding object using role_binding.yaml

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
    Create IBM Storage Scale CSI Operator ServiceAccount using service_account.yaml

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
    Create IBM Storage Scale CSI Operator CRD (Custom Resource Defination) Object

    Args:
       None

    Returns:
       None

    Raises:
        Raises an ValueError exception but it is expected. hence we pass.

    """
    custom_object_api_instance = client.CustomObjectsApi()
    try:
        custom_object_api_response = custom_object_api_instance.create_cluster_custom_object(
            group="apiextensions.k8s.io",
            version="v1",
            plural="customresourcedefinitions",
            body=body,
            pretty=True
        )
        LOGGER.debug(custom_object_api_response)
        LOGGER.info(
            "Creating IBM StorageScale CRD object using csiscaleoperators.csi.ibm.com.crd.yaml file")
    except ValueError as e:
        LOGGER.error(
            f"Exception when calling CustomObjectsApi->create_namespaced_custom_object: {e}")
        LOGGER.info(
            "while there is valuerror expection,but CRD created successfully")
        assert False


def delete_crd():
    """
    Delete existing IBM Storage Scale CSI Operator CRD (Custom Resource Defination) Object

    Args:
       None

    Returns:
       None

    Raises:
        Raises an exception on kubernetes client api failure and asserts

    """
    crd_name = "csiscaleoperators.csi.ibm.com"
    custom_object_api_instance = client.CustomObjectsApi()
    try:
        custom_object_api_response = custom_object_api_instance.delete_cluster_custom_object(
            group="apiextensions.k8s.io",
            version="v1",
            plural="customresourcedefinitions",
            name=crd_name
        )
        LOGGER.debug(str(custom_object_api_response))
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CustomObjectsApi->delete_cluster_custom_object: {e}")
        assert False


def delete_namespace(namespace_name):
    """
    Delete IBM Storage Scale CSI Operator namespace

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
            name=namespace_name, pretty=True)
        LOGGER.debug(str(delete_namespace_api_response))
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CoreV1Api->delete_namespace: {e}")
        assert False


def delete_deployment():
    """
    Delete IBM Storage Scale CSI Operator Deployment object from Operator namespace

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
    Delete IBM Storage Scale CSI Operator ServiceAccount from Operator namespace

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
    Delete IBM Storage Scale CSI Operator ClusterRole Object

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
    Delete IBM Storage Scale CSI Operator ClusterRoleBinding Object

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
    crd_name = "csiscaleoperators.csi.ibm.com"
    custom_object_api_instance = client.CustomObjectsApi()
    while (count > 0):
        try:
            custom_object_api_response = custom_object_api_instance.get_cluster_custom_object(
                group="apiextensions.k8s.io",
                version="v1",
                plural="customresourcedefinitions",
                name=crd_name
            )
            LOGGER.debug(custom_object_api_response)
            LOGGER.info("still deleting crd")
            count -= 1
            time.sleep(5)

        except ApiException:
            LOGGER.info("crd deleted")
            return

    LOGGER.error("crd is not deleted")
    assert False


def check_namespace_deleted(namespace_name):
    """
    Function for checking namespace object is deleted or not
    If namespace is not deleted in 120 seconds, Function asserts

    Raises:
        Raises an exception on kubernetes client api failure and asserts

    """
    count = 18
    list_namespace_api_instance = client.CoreV1Api()
    while (count > 0):
        try:
            list_namespace_api_response = list_namespace_api_instance.read_namespace(
                name=namespace_name, pretty=True)
            LOGGER.debug(str(list_namespace_api_response))
            LOGGER.info(f'Namespace Delete : still deleting {namespace_name}')
            count = count-1
            time.sleep(10)
        except ApiException:
            LOGGER.info(f'Namespace Delete : {namespace_name} is deleted')
            return

    LOGGER.error(f'namespace  {namespace_name} is not deleted')
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
    custom_object_api_instance = client.CustomObjectsApi()
    try:
        custom_object_api_response = custom_object_api_instance.get_cluster_custom_object(
            group="apiextensions.k8s.io",
            version="v1",
            plural="customresourcedefinitions",
            name=crd_name
        )
        LOGGER.debug(str(custom_object_api_response))
        LOGGER.info(f"crd  {crd_name} exists")
        return True
    except ApiException:
        LOGGER.info(f"crd {crd_name} does not exist")
        return False


def check_namespace_exists(namespace_name):
    """
    Checks namespace namespace_name exists or not

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
            name=namespace_name, pretty=True)
        LOGGER.debug(str(read_namespace_api_response))
        LOGGER.info(f"Namespace Check  : {namespace_name} exists")
        return True
    except ApiException:
        LOGGER.info(f"Namespace Check  : {namespace_name} does not exists")
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
        LOGGER.info(f"CSI operator image : {api_response.status.container_statuses[-1].image}")
        LOGGER.info(
            f"CSI operator image id : {api_response.status.container_statuses[-1].image_id}")
    except ApiException:
        LOGGER.info("Unable to get operator image")


def check_ns_exists(passed_kubeconfig_value, namespace_value):
    config.load_kube_config(config_file=passed_kubeconfig_value)
    read_namespace_api_instance = client.CoreV1Api()
    try:
        read_namespace_api_response = read_namespace_api_instance.read_namespace(
            name=namespace_value, pretty=True)
        LOGGER.debug(str(read_namespace_api_response))
        LOGGER.info(f"Namespace Check  : CSI Operator Namespace {namespace_value} exists")
        return True
    except ApiException:
        LOGGER.info(f"Namespace Check  : CSI Operator Namespace {namespace_value} does not exists")
        return False


def get_kubernetes_version(passed_kubeconfig_value):
    config.load_kube_config(config_file=passed_kubeconfig_value)
    api_instance = client.VersionApi()
    try:
        api_response = api_instance.get_code()
        api_response = api_response.__dict__
        LOGGER.info(f"kubernetes version is {api_response['_git_version']}")
        LOGGER.info(f"platform is {api_response['_platform']}")
    except ApiException as e:
        LOGGER.info(f"Kubernetes version cannot be fetched due to {e}")


def check_nodes_available(label, label_name):
    """
    checks number of nodes with label
    if it is 0 , asserts
    """
    api_instance = client.CoreV1Api()
    label_selector = ""
    for label_val in label:
        label_selector += str(label_val["key"])+"="+str(label_val["value"])+","
    label_selector = label_selector[0:-1]
    try:
        api_response_2 = api_instance.list_node(
            pretty=True, label_selector=label_selector)
        if len(api_response_2.items) == 0:
            LOGGER.error(f"0 nodes matches with {label_name}")
            LOGGER.error("please check labels")
            assert False
    except ApiException as e:
        LOGGER.error(f"Exception when calling CoreV1Api->list_node: {e}")
        assert False


def base64encoder(input_str):
    """Takes input string and converts it to base 64 string"""
    message_bytes = input_str.encode('ascii')
    base64_bytes = base64.b64encode(message_bytes)
    base64_message = base64_bytes.decode('ascii')
    return base64_message


def create_secret(secret_data_passed, secret_name):
    """
    Create secret secet_name

    Args:
        param1: secret_data_passed - data for secret body
        param2: secret_name - name of secret to be created

    Returns:
        None

    Raises:
        Raises an exception on kubernetes client api failure and asserts

    """
    secret_api_instance = client.CoreV1Api()
    secret_data = copy.deepcopy(secret_data_passed)
    secret_data["username"] = base64encoder(secret_data["username"])
    secret_data["password"] = base64encoder(secret_data["password"])
    secret_metadata = client.V1ObjectMeta(
        name=secret_name,
        labels={"product": "ibm-spectrum-scale-csi"}
    )
    secret_body = client.V1Secret(
        api_version="v1",
        kind="Secret",
        metadata=secret_metadata,
        data=secret_data
    )
    try:
        LOGGER.info(f'Creating secret {secret_name}')
        secret_api_response = secret_api_instance.create_namespaced_secret(
            namespace=namespace_value,
            body=secret_body,
            pretty=True
        )
        LOGGER.debug(str(secret_api_response))
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CoreV1Api->create_namespaced_secret: {e}")
        assert False


def delete_secret(secret_name):
    """
    delete secret secret_name

    Args:
       param1: secret_name - name of secret to be deleted

    Returns:
       None

    Raises:
        Raises an exception on kubernetes client api failure and asserts

    """

    delete_secret_api_instance = client.CoreV1Api()
    try:
        delete_secret_api_response = delete_secret_api_instance.delete_namespaced_secret(
            name=secret_name, namespace=namespace_value, pretty=True)
        LOGGER.info(f'Secret {secret_name} has been deleted')
        LOGGER.debug(str(delete_secret_api_response))
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CoreV1Api->delete_namespaced_secret: {e}")
        assert False


def check_secret_exists(secret_name):
    """
    Checks secret secret_name exists or not

    Args:
       param1: secret_name - name of secret to be checked

    Returns:
       return True  , if secret exists
       return False , if secret does not exist

    Raises:
        None

    """

    api_instance = client.CoreV1Api()
    try:
        api_response = api_instance.read_namespaced_secret(
            name=secret_name, namespace=namespace_value, pretty=True)
        LOGGER.debug(str(api_response))
        LOGGER.info(f'Secret {secret_name} exists')
        return True
    except ApiException:
        LOGGER.info(f'Secret {secret_name} does not exist')
        return False


def check_secret_is_deleted(secret_name):
    """
    checks secret deleted or not
    if secret not deleted in 120 seconds , asserts

    Args:
       param1: secret_name - name of secret to be checked
    """
    var = True
    count = 12
    api_instance = client.CoreV1Api()
    while (var and count > 0):
        try:
            api_response = api_instance.read_namespaced_secret(
                name=secret_name, namespace=namespace_value, pretty=True)
            LOGGER.info("still deleting secret")
            LOGGER.debug(str(api_response))
            count = count-1
            time.sleep(10)
        except ApiException:
            LOGGER.info(f"Secret {secret_name} has been deleted")
            var = False

    if count <= 0:
        LOGGER.error(f"Secret {secret_name} is not deleted")
        assert False


def create_configmap(file_path, make_cacert_wrong, configmap_name):
    """
    Create configmap with file at file_path
    if make_cacert_wrong==True then it makes cacert wrong

    Args:
        param1: file_path - path of cacert file
        param2: make_cacert_wrong - for operator testcase, cacert wrong

    Returns:
       None

    Raises:
        Raises an exception on kubernetes client api failure and asserts

    """
    api_instance = client.CoreV1Api()
    metadata = client.V1ObjectMeta(
        name=configmap_name,
        namespace=namespace_value,
    )
    with open(file_path, 'r') as f:
        file_content = f.read()
    if make_cacert_wrong:
        file_content = file_content[0:50]+file_content[-50:-1]
    data_dict = {}
    data_dict[configmap_name] = file_content
    configmap = client.V1ConfigMap(
        api_version="v1",
        kind="ConfigMap",
        data=data_dict,
        metadata=metadata
    )
    try:
        api_response = api_instance.create_namespaced_config_map(
            namespace=namespace_value,
            body=configmap,
            pretty=True,
        )
        LOGGER.debug(str(api_response))
        LOGGER.info(f"configmap {configmap_name} has been created")

    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CoreV1Api->create_namespaced_config_map: {e}")
        assert False


def delete_configmap(configmap_name):
    """
    deletes configmap

    Args:
       None

    Returns:
       None

    Raises:
        Raises an exception on kubernetes client api failure and asserts

    """
    api_instance = client.CoreV1Api()
    try:
        api_response = api_instance.delete_namespaced_config_map(
            namespace=namespace_value,
            name=configmap_name,
            pretty=True,
        )
        LOGGER.debug(str(api_response))
        LOGGER.info(f"configmap {configmap_name} has been deleted")

    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CoreV1Api->create_namespaced_config_map: {e}")
        assert False


def check_configmap_exists(configmap_name):
    """
    Checks configmap configmap_name exists or not

    Args:
       param1: configmap_name - name of configmap to be checked

    Returns:
       return True  , if configmap exists
       return False , if configmap does not exist

    Raises:
        None

    """

    api_instance = client.CoreV1Api()
    try:
        api_response = api_instance.read_namespaced_config_map(
            namespace=namespace_value,
            name=configmap_name,
            pretty=True,
        )
        LOGGER.debug(str(api_response))
        LOGGER.info(f'configmap {configmap_name} exists')
        return True
    except ApiException:
        LOGGER.info(f'configmap {configmap_name} does not exist')
        return False


def check_configmap_is_deleted(configmap_name):
    """
    checks configmap deleted or not
    if configmap not deleted in 120 seconds , asserts

    Args:
       param1: configmap_name - name of configmap to be checked
    """
    count = 12
    api_instance = client.CoreV1Api()
    while (count > 0):
        try:
            api_response = api_instance.read_namespaced_config_map(
                namespace=namespace_value,
                name=configmap_name,
                pretty=True,
            )
            LOGGER.info("still deleting configmap")
            LOGGER.debug(str(api_response))
            count = count-1
            time.sleep(10)
        except ApiException:
            LOGGER.info(f"configmap {configmap_name} deletion confirmed")
            return

    LOGGER.error(f"configmap {configmap_name} is not deleted")
    assert False


def check_pod_running(pod_name):
    """
    checking phase of pod pod_name to be running
    if not running then asserts
    """

    api_instance = client.CoreV1Api()
    val = 0
    while val < 12:
        try:
            api_response = api_instance.read_namespaced_pod(
                name=pod_name, namespace=namespace_value, pretty=True)
            LOGGER.debug(str(api_response))
            if api_response.status.phase == "Running":
                LOGGER.info(f'POD Check : POD {pod_name} is Running')
                return
            time.sleep(5)
            val += 1
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod: {e}")
            LOGGER.info(f"POD Check : POD {pod_name} does not exists on Cluster")
            assert False
    LOGGER.error(f'POD Check : POD {pod_name} is not Running')
    assert False


def get_driver_ds_pod_name():
    try:
        pod_list_api_instance = client.CoreV1Api()
        pod_list_api_response = pod_list_api_instance.list_namespaced_pod(
            namespace=namespace_value, pretty=True, field_selector="spec.serviceAccountName=ibm-spectrum-scale-csi-node")
        daemonset_pod_name = pod_list_api_response.items[0].metadata.name
        LOGGER.debug(str(pod_list_api_response))
        return daemonset_pod_name
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CoreV1Api->list_namespaced_pod: {e}")
        assert False


def get_driver_image():
    pod_name = get_driver_ds_pod_name()
    api_instance = client.CoreV1Api()
    try:
        api_response = api_instance.read_namespaced_pod(
            name=pod_name, namespace=namespace_value, pretty=True)
        for container in api_response.status.container_statuses:
            if(container.name == "ibm-spectrum-scale-csi"):
                LOGGER.info(f"CSI driver image :  {container.image}")
                LOGGER.info(f"CSI driver image id : {container.image_id}")
    except ApiException:
        LOGGER.info("Unable to get driver image")


def check_pod_image(pod_name, image_name):
    """
    checking phase of pod pod_name to be running
    if not running then asserts
    """

    api_instance = client.CoreV1Api()
    try:
        api_response = api_instance.read_namespaced_pod(
            name=pod_name, namespace=namespace_value, pretty=True)
        LOGGER.debug(str(api_response))
        search_result = re.search(image_name, str(api_response))
        LOGGER.info(search_result)
        if search_result is not None:
            LOGGER.info(f"Image {image_name} matched for pod {pod_name}")
            return
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CoreV1Api->read_namespaced_pod: {e}")
        assert False

    LOGGER.error(f"Image {image_name} not matched for pod {pod_name}")
    LOGGER.error(str(api_response))
    assert False


def get_pod_list_and_check_running(label, required_pods):
    api_instance = client.CoreV1Api()
    for _ in range(0, 24):
        try:
            api_response = api_instance.list_pod_for_all_namespaces(
                pretty=True, label_selector=label)
            pod_status = True
            for pod_info in api_response.items:
                if not(pod_info.status.phase == "Running"):
                    pod_status = False
                    break
            if pod_status is True and (len(api_response.items) == required_pods):
                return
            time.sleep(20)
            LOGGER.info(f"Checking for pod with label {label}")
        except ApiException as e:
            LOGGER.error(f"Exception when calling CoreV1Api->list_pod_for_all_namespaces: {e}")
            assert False
    else:
        LOGGER.error(f"Pods with label {label} are not in expected state {api_response}")
        assert False


def get_pod_list_with_label(label):
    api_instance = client.CoreV1Api()
    try:
        api_response = api_instance.list_pod_for_all_namespaces(pretty=True, label_selector=label)
        pod_list = []
        for pod_info in api_response.items:
            pod_list.append(pod_info.metadata.name)
        return pod_list
    except ApiException as e:
        LOGGER.error(f"Exception when calling CoreV1Api->list_pod_for_all_namespaces: {e}")
        assert False
