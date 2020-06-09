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


def create_deployment():
    """
    Create IBM Spectrum Scale CSI Operator deployment object in operator namespace using
    deployment_operator_image_for_crd and deployment_driver_image_for_crd parameters from
    config.json file

    Args:
        None

    Returns:
       None

    Raises:
        Raises an exception on kubernetes client api failure and asserts

    """
    deployment_apps_api_instance = client.AppsV1Api()
    filepath = "../../operator/deploy/operator.yaml"
    try:
        with open(filepath, "r") as f:
            loaddep_yaml = yaml.full_load(f.read())
    except yaml.YAMLError as exc:
        print ("Error in configuration file:", exc)
        assert False

    try:
        LOGGER.info("Creating Operator Deployment")
        deployment_apps_api_response = deployment_apps_api_instance.create_namespaced_deployment(
            namespace=namespace_value, body=loaddep_yaml)
        LOGGER.debug(str(deployment_apps_api_response))
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling RbacAuthorizationV1Api->create_namespaced_deployment: {e}")
        assert False


def create_deployment_old(config_file):
    """
    Create IBM Spectrum Scale CSI Operator deployment object in operator namespace using
    deployment_operator_image_for_crd and deployment_driver_image_for_crd parameters from
    config.json file

    Args:
        param1: config_file - configuration json file

    Returns:
       None

    Raises:
        Raises an exception on kubernetes client api failure and asserts

    """

    deployment_apps_api_instance = client.AppsV1Api()
    
    deployment_labels = {
        "app.kubernetes.io/instance": "ibm-spectrum-scale-csi-operator",
        "app.kubernetes.io/managed-by": "ibm-spectrum-scale-csi-operator",
        "app.kubernetes.io/name": "ibm-spectrum-scale-csi-operator",
        "product": "ibm-spectrum-scale-csi",
        "release": "ibm-spectrum-scale-csi-operator"
    }

    deployment_annotations = {
        "productID": "ibm-spectrum-scale-csi-operator",
        "productName": "IBM Spectrum Scale CSI Operator",
        "productVersion": "2.0.0"
    }

    deployment_metadata = client.V1ObjectMeta(
        name="ibm-spectrum-scale-csi-operator", labels=deployment_labels, namespace=namespace_value)

    deployment_selector = client.V1LabelSelector(
        match_labels={"app.kubernetes.io/name": "ibm-spectrum-scale-csi-operator"})

    podtemplate_metadata = client.V1ObjectMeta(
        labels=deployment_labels, annotations=deployment_annotations)

    pod_affinity = client.V1Affinity(
        node_affinity=client.V1NodeAffinity(
            required_during_scheduling_ignored_during_execution=client.V1NodeSelector(
                node_selector_terms=[client.V1NodeSelectorTerm(
                    match_expressions=[client.V1NodeSelectorRequirement(
                        key="beta.kubernetes.io/arch", operator="Exists")]
                )]
            )
        )
    )
    ansible_pod_container = client.V1Container(
        image=config_file["deployment_operator_image_for_crd"],
        command=["/usr/local/bin/ao-logs",
                 "/tmp/ansible-operator/runner", "stdout"],
        liveness_probe=client.V1Probe(_exec=client.V1ExecAction(
            command=["/health_check.sh"]), initial_delay_seconds=10, period_seconds=30),
        readiness_probe=client.V1Probe(_exec=client.V1ExecAction(
            command=["/health_check.sh"]), initial_delay_seconds=3, period_seconds=1),
        name="ansible", image_pull_policy="IfNotPresent",
        security_context=client.V1SecurityContext(
            capabilities=client.V1Capabilities(drop=["ALL"])),
        volume_mounts=[client.V1VolumeMount(
            mount_path="/tmp/ansible-operator/runner", name="runner", read_only=True)],
        env=[client.V1EnvVar(name="CSI_DRIVER_IMAGE", value=config_file["deployment_driver_image_for_crd"])])

    operator_pod_container = client.V1Container(
        image=config_file["deployment_operator_image_for_crd"],
        name="operator", image_pull_policy="IfNotPresent",
        liveness_probe=client.V1Probe(_exec=client.V1ExecAction(
            command=["/health_check.sh"]), initial_delay_seconds=10, period_seconds=30),
        readiness_probe=client.V1Probe(_exec=client.V1ExecAction(
            command=["/health_check.sh"]), initial_delay_seconds=3, period_seconds=1),
        security_context=client.V1SecurityContext(
            capabilities=client.V1Capabilities(drop=["ALL"])),
        env=[client.V1EnvVar(name="WATCH_NAMESPACE",
                             value_from=client.V1EnvVarSource(field_ref=client.V1ObjectFieldSelector(
                                 field_path="metadata.namespace"))),
             client.V1EnvVar(name="POD_NAME", value_from=client.V1EnvVarSource(
                 field_ref=client.V1ObjectFieldSelector(field_path="metadata.name"))),
             client.V1EnvVar(name="OPERATOR_NAME",
                             value="ibm-spectrum-scale-csi-operator"),
             client.V1EnvVar(name="CSI_DRIVER_IMAGE", value=config_file["deployment_driver_image_for_crd"])],
        volume_mounts=[client.V1VolumeMount(
            mount_path="/tmp/ansible-operator/runner", name="runner")]
    )
    pod_spec = client.V1PodSpec(affinity=pod_affinity,
                                containers=[ansible_pod_container,
                                            operator_pod_container],
                                service_account_name="ibm-spectrum-scale-csi-operator",
                                volumes=[client.V1Volume(empty_dir=client.V1EmptyDirVolumeSource(medium="Memory"), name="runner")])

    podtemplate_spec = client.V1PodTemplateSpec(
        metadata=podtemplate_metadata, spec=pod_spec)

    deployment_spec = client.V1DeploymentSpec(
        replicas=1, selector=deployment_selector, template=podtemplate_spec)

    body_dep = client.V1Deployment(
        kind='Deployment', api_version='apps/v1', metadata=deployment_metadata, spec=deployment_spec)
    
    try:
        LOGGER.info("creating deployment for operator")
        deployment_apps_api_response = deployment_apps_api_instance.create_namespaced_deployment(
            namespace=namespace_value, body=body_dep)
        LOGGER.debug(str(deployment_apps_api_response))
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling RbacAuthorizationV1Api->create_namespaced_deployment: {e}")
        assert False


def create_cluster_role():
    """
    Create IBM Spectrum Scale CSI Operator cluster role in Operator namespace

    Args:
       None

    Returns:
       None

    Raises:
        Raises an exception on kubernetes client api failure and asserts

    """
    cluster_role_api_instance = client.RbacAuthorizationV1Api()
    pretty = True
    cluster_role_labels = {
        "app.kubernetes.io/instance": "ibm-spectrum-scale-csi-operator",
        "app.kubernetes.io/managed-by": "ibm-spectrum-scale-csi-operator",
        "app.kubernetes.io/name": "ibm-spectrum-scale-csi-operator",
        "product": "ibm-spectrum-scale-csi",
        "release": "ibm-spectrum-scale-csi-operator"
    }

    cluster_role_metadata = client.V1ObjectMeta(
        name="ibm-spectrum-scale-csi-operator", labels=cluster_role_labels, namespace=namespace_value)
    cluster_role_rules = []

    cluster_role_rules.append(client.V1PolicyRule(api_groups=["*"], resources=[
                              'pods', 'persistentvolumeclaims', 'services',
                              'endpoints', 'events', 'configmaps', 'secrets',
                              'secrets/status', 'services/finalizers', 'serviceaccounts', 'securitycontextconstraints'], verbs=["*"]))
    cluster_role_rules.append(client.V1PolicyRule(api_groups=['rbac.authorization.k8s.io'], resources=[
                              'clusterroles', 'clusterrolebindings'], verbs=["*"]))
    cluster_role_rules.append(client.V1PolicyRule(api_groups=['apps'], resources=[
                              'deployments', 'daemonsets', 'replicasets', 'statefulsets'], verbs=["*"]))
    cluster_role_rules.append(client.V1PolicyRule(api_groups=[
                              'monitoring.coreos.com'], resources=['servicemonitors'], verbs=['get', 'create']))
    cluster_role_rules.append(client.V1PolicyRule(
        api_groups=['apps'], resources=['replicasets'], verbs=["get"]))
    cluster_role_rules.append(client.V1PolicyRule(
        api_groups=['csi.ibm.com'], resources=['*'], verbs=["*"]))
    cluster_role_rules.append(client.V1PolicyRule(api_groups=[
                              'security.openshift.io'], resources=['securitycontextconstraints'], verbs=["*"]))
    cluster_role_rules.append(client.V1PolicyRule(api_groups=['storage.k8s.io'], resources=[
                              'volumeattachments', 'storageclasses'], verbs=["*"]))
    cluster_role_rules.append(client.V1PolicyRule(api_groups=['apps'], resource_names=[
                              'ibm-spectrum-scale-csi-operator'], resources=['deployments/finalizers'], verbs=['update']))
    body = client.V1ClusterRole(kind='ClusterRole', api_version='rbac.authorization.k8s.io/v1',
                                metadata=cluster_role_metadata, rules=cluster_role_rules)

    try:
        LOGGER.info("Creating ibm-spectrum-scale-csi-operator ClusterRole ")
        cluster_role_api_response = cluster_role_api_instance.create_cluster_role(
            body, pretty=pretty)
        LOGGER.debug(str(cluster_role_api_response))
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling RbacAuthorizationV1Api->create_cluster_role: {e}")
        assert False


def create_cluster_role_binding():
    """
    Create IBM Spectrum Scale CSI Operator ClusterRoleBinding object in Operator namepsace

    Args:
       None

    Returns:
       None

    Raises:
        Raises an exception on kubernetes client api failure and asserts

    """
    cluster_role_binding_api_instance = client.RbacAuthorizationV1Api()
    pretty = True
    cluster_role_binding_labels = {
        "app.kubernetes.io/instance": "ibm-spectrum-scale-csi-operator",
        "app.kubernetes.io/managed-by": "ibm-spectrum-scale-csi-operator",
        "app.kubernetes.io/name": "ibm-spectrum-scale-csi-operator",
                                  "product": "ibm-spectrum-scale-csi",
                                  "release": "ibm-spectrum-scale-csi-operator"
    }

    cluster_role_binding_metadata = client.V1ObjectMeta(
        name="ibm-spectrum-scale-csi-operator", labels=cluster_role_binding_labels, namespace=namespace_value)

    cluster_role_binding_role_ref = client.V1RoleRef(
        api_group="rbac.authorization.k8s.io", kind="ClusterRole", name="ibm-spectrum-scale-csi-operator")

    cluster_role_binding_subjects = client.V1Subject(
        kind="ServiceAccount", name="ibm-spectrum-scale-csi-operator", namespace=namespace_value)

    cluster_role_binding_body = client.V1ClusterRoleBinding(kind='ClusterRoleBinding',
                                                            api_version='rbac.authorization.k8s.io/v1',
                                                            metadata=cluster_role_binding_metadata,
                                                            role_ref=cluster_role_binding_role_ref,
                                                            subjects=[cluster_role_binding_subjects])

    try:
        LOGGER.info("creating cluster role binding")
        cluster_role_binding_api_response = cluster_role_binding_api_instance.create_cluster_role_binding(
            cluster_role_binding_body, pretty=pretty)
        LOGGER.debug(cluster_role_binding_api_response)
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling RbacAuthorizationV1Api->create_cluster_role_binding: {e}")
        assert False


def create_service_account():
    """
    Create IBM Spectrum Scale CSI Operator ServiceAccount in Operator namespace

    Args:
       None

    Returns:
       None

    Raises:
        Raises an exception on kubernetes client api failure and asserts

    """
    pretty = True
    service_account_api_instance = client.CoreV1Api()
    service_account_labels = {
        "app.kubernetes.io/instance": "ibm-spectrum-scale-csi-operator",
        "app.kubernetes.io/managed-by": "ibm-spectrum-scale-csi-operator",
        "app.kubernetes.io/name": "ibm-spectrum-scale-csi-operator",
        "product": "ibm-spectrum-scale-csi",
        "release": "ibm-spectrum-scale-csi-operator"
    }

    service_account_metadata = client.V1ObjectMeta(
        name="ibm-spectrum-scale-csi-operator", namespace=namespace_value, labels=service_account_labels)
    service_account_body = client.V1ServiceAccount(
        api_version="v1", kind="ServiceAccount", metadata=service_account_metadata)

    try:
        LOGGER.info("Creating ibm-spectrum-scale-csi-operator ServiceAccount")
        service_account_api_response = service_account_api_instance.create_namespaced_service_account(
            namespace=namespace_value, body=service_account_body, pretty=pretty)
        LOGGER.debug(str(service_account_api_response))
    except ApiException as e:
        LOGGER.error(
            f"Exception when calling CoreV1Api->create_namespaced_service_account: {e}")
        assert False


def create_crd():
    """
    Create IBM Spectrum Scale CSI Operator CRD (Custom Resource Defination) Object

    Args:
       None

    Returns:
       None

    Raises:
        Raises an ValueError exception but it is expected. hence we pass.

    """
    filepath = "../../operator/deploy/crds/csiscaleoperators.csi.ibm.com.crd.yaml"
    try:
        with open(filepath, "r") as f:
            loadcrd_yaml = yaml.full_load(f.read())
    except yaml.YAMLError as exc:
        print ("Error in configuration file:", exc)
        assert False

    crd_api_instance = client.ApiextensionsV1beta1Api()
    try:
        LOGGER.info(
            "Creating IBM SpectrumScale CRD object using csiscaleoperators.csi.ibm.com.crd.yaml file")
        crd_api_response = crd_api_instance.create_custom_resource_definition(
            loadcrd_yaml, pretty=True)
        LOGGER.debug(str(crd_api_response))
    except ValueError:
        LOGGER.info(
            "while there is valuerror expection,but CRD created successfully")


def create_crd_old():
    """
    Create IBM Spectrum Scale CSI Operator CRD (Custom Resource Defination) Object

    Args:
       None

    Returns:
       None

    Raises:
        Raises an ValueError exception but it is expected. hence we pass.

    """

    
    # input to crd_metadata
    crd_labels = {
        "app.kubernetes.io/instance": "ibm-spectrum-scale-csi-operator",
        "app.kubernetes.io/managed-by": "ibm-spectrum-scale-csi-operator",
        "app.kubernetes.io/name": "ibm-spectrum-scale-csi-operator",
        "release": "ibm-spectrum-scale-csi-operator"
    }

    # input to crd_body
    crd_metadata = client.V1ObjectMeta(
        name="csiscaleoperators.csi.ibm.com", labels=crd_labels)

    crd_names = client.V1beta1CustomResourceDefinitionNames(
        kind="CSIScaleOperator",
        list_kind="CSIScaleOperatorList",
        plural="csiscaleoperators",
        singular="csiscaleoperator"
    )

    crd_subresources = client.V1beta1CustomResourceSubresources(status={})
    
    # input to crd_validation     json input
    filepath = "../../operator/deploy/crds/csiscaleoperators.csi.ibm.com.crd.yaml"
    try:
        with open(filepath, "r") as f:
            loadcrd_yaml = yaml.full_load(f.read())
    except yaml.YAMLError as exc:
        print ("Error in configuration file:", exc)
        assert False 
    properties   = loadcrd_yaml['spec']['validation']['openAPIV3Schema']['properties']
    
    crd_open_apiv3_schema = client.V1beta1JSONSchemaProps(
        properties=properties, type="object")
    crd_validation = client.V1beta1CustomResourceValidation(
        open_apiv3_schema=crd_open_apiv3_schema)
    crd_versions = [client.V1beta1CustomResourceDefinitionVersion(
        name="v1", served=True, storage=True)]

    crd_spec = client.V1beta1CustomResourceDefinitionSpec(
        group="csi.ibm.com",
        names=crd_names,
        scope="Namespaced",
        subresources=crd_subresources,
        validation=crd_validation,
        version="v1",
        versions=crd_versions
    )
    
    
    crd_body = client.V1beta1CustomResourceDefinition(
        api_version="apiextensions.k8s.io/v1beta1",
        kind="CustomResourceDefinition",
        metadata=crd_metadata,
        spec=crd_spec)
    
    crd_api_instance = client.ApiextensionsV1beta1Api()
    try:
        LOGGER.info("creating crd")
        crd_api_response = crd_api_instance.create_custom_resource_definition(
            crd_body, pretty=True)
        LOGGER.debug(str(crd_api_response))
    except ValueError:
        LOGGER.info(
            "while there is valuerror expection,but CRD created successfully")


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
    var = True
    count = 12
    list_crd_api_instance = client.ApiextensionsV1beta1Api()
    while (var and count > 0):
        try:
            list_crd_api_response = list_crd_api_instance.read_custom_resource_definition(
                pretty=True, name="ibm-spectrum-scale-csi")
            LOGGER.debug(list_crd_api_response)
            LOGGER.info("still deleting crd")
            count -= 1
            time.sleep(5)

        except ApiException:
            LOGGER.info("crd deleted")
            var = False

    if count <= 0:
        LOGGER.error("crd is not deleted")
        assert False


def check_namespace_deleted():
    """
    Function for checking namespace object is deleted or not
    If namespace is not deleted in 120 seconds, Function asserts

    Raises:
        Raises an exception on kubernetes client api failure and asserts

    """
    var = True
    count = 24
    list_namespace_api_instance = client.CoreV1Api()
    while (var and count > 0):
        try:
            list_namespace_api_response = list_namespace_api_instance.read_namespace(
                name=namespace_value, pretty=True)
            LOGGER.debug(str(list_namespace_api_response))
            LOGGER.info(f'Still deleting namespace {namespace_value}')
            count = count-1
            time.sleep(5)
        except ApiException:
            LOGGER.info(f'namespace {namespace_value} is deleted')
            var = False

    if count <= 0:
        LOGGER.error(f'namespace  {namespace_value} is not deleted')
        assert False


def check_deployment_deleted():
    """
    Function for checking deployment is deleted or not
    If deployment is not deleted in 30 seconds, Function asserts

    Raises:
        Raises an exception on kubernetes client api failure and asserts

    """
    var = True
    count = 6
    api_instance = client.AppsV1Api()
    while (var and count > 0):
        try:
            api_response = api_instance.read_namespaced_deployment(
                name="ibm-spectrum-scale-csi-operator", namespace=namespace_value, pretty=True)
            LOGGER.debug(str(api_response))
            LOGGER.info(f'Still Deleting ibm-spectrum-scale-csi-operator deployment')
            count = count-1
            time.sleep(5)
        except ApiException:
            LOGGER.info("Deployment ibm-spectrum-scale-csi-operator is deleted")
            var = False

    if count <= 0:
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
    var = True
    count = 6
    api_instance = client.CoreV1Api()
    while (var and count > 0):
        try:
            api_response = api_instance.read_namespaced_service_account(
                name=service_account_name, namespace=namespace_value, pretty=True)
            LOGGER.debug(str(api_response))
            LOGGER.info(f'Still deleting ServiceAccount {service_account_name}')
            count = count-1
            time.sleep(5)
        except ApiException:
            LOGGER.info(f'ServiceAccount {service_account_name} is deleted')
            var = False

    if count <= 0:
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
    var = True
    count = 6
    api_instance = client.RbacAuthorizationV1Api()
    while (var and count > 0):
        try:
            api_response = api_instance.read_cluster_role(
                name=cluster_role_name, pretty=True)
            LOGGER.debug(str(api_response))
            LOGGER.info(f'Still deleting ClusterRole {cluster_role_name} ')
            count = count-1
            time.sleep(5)
        except ApiException:
            LOGGER.info(f'ClusterRole {cluster_role_name} is deleted')
            var = False

    if count <= 0:
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
    var = True
    count = 6
    api_instance = client.RbacAuthorizationV1Api()
    while (var and count > 0):
        try:
            api_response = api_instance.read_cluster_role_binding(
                name=cluster_role_binding_name, pretty=True)
            LOGGER.debug(str(api_response))
            LOGGER.info(f'Still deleting ClusterRoleBinding {cluster_role_binding_name}')
            count = count-1
            time.sleep(5)
        except ApiException:
            LOGGER.info(f'ClusterRoleBinding {cluster_role_binding_name} is deleted')
            var = False

    if count <= 0:
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
    read_crd_api_instance = client.ApiextensionsV1beta1Api()
    try:
        read_crd_api_response = read_crd_api_instance.read_custom_resource_definition(
            pretty=True, name="csiscaleoperators.csi.ibm.com")
        LOGGER.debug(str(read_crd_api_response))
        LOGGER.info("crd exists")
        return True
    except ApiException:
        LOGGER.info("crd does not exist")
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
