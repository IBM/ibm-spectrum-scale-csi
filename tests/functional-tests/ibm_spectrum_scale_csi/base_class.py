import copy
import logging
import pytest
import yaml
from kubernetes import client, config
from kubernetes.client.rest import ApiException
import ibm_spectrum_scale_csi.kubernetes_apis.kubernetes_objects_function as kubeobjectfunc
import ibm_spectrum_scale_csi.kubernetes_apis.csi_object_function as csiobjectfunc
import ibm_spectrum_scale_csi.kubernetes_apis.csi_storage_function as csistoragefunc
import ibm_spectrum_scale_csi.spectrum_scale_apis.fileset_functions as filesetfunc

LOGGER = logging.getLogger()


class Scaleoperator:
    def __init__(self, kubeconfig_value, namespace_value, operator_yaml):

        self.kubeconfig = kubeconfig_value
        self.operator_namespace = namespace_value
        self.operator_yaml_file_path = operator_yaml
        kubeobjectfunc.set_global_namespace_value(self.operator_namespace)
        csiobjectfunc.set_namespace_value(self.operator_namespace)

    def create(self):

        config.load_kube_config(config_file=self.kubeconfig)

        body = self.get_operator_body()
        if not(kubeobjectfunc.check_namespace_exists(self.operator_namespace)):
            kubeobjectfunc.create_namespace(self.operator_namespace)

        if not(kubeobjectfunc.check_deployment_exists()):
            kubeobjectfunc.create_deployment(body['Deployment'])

        if not(kubeobjectfunc.check_cluster_role_exists("ibm-spectrum-scale-csi-operator")):
            kubeobjectfunc.create_cluster_role(body['ClusterRole'])

        if not(kubeobjectfunc.check_service_account_exists("ibm-spectrum-scale-csi-operator")):
            kubeobjectfunc.create_service_account(body['ServiceAccount'])

        if not(kubeobjectfunc.check_cluster_role_binding_exists("ibm-spectrum-scale-csi-operator")):
            kubeobjectfunc.create_cluster_role_binding(body['ClusterRoleBinding'])

        if not(kubeobjectfunc.check_crd_exists()):
            kubeobjectfunc.create_crd(body['CustomResourceDefinition'])

    def delete(self, condition=False):

        config.load_kube_config(config_file=self.kubeconfig)
        if csiobjectfunc.check_scaleoperatorobject_is_deployed():   # for edge cases if custom object is not deleted
            csiobjectfunc.delete_custom_object()
            csiobjectfunc.check_scaleoperatorobject_is_deleted()

        if kubeobjectfunc.check_crd_exists():
            kubeobjectfunc.delete_crd()
        kubeobjectfunc.check_crd_deleted()

        if kubeobjectfunc.check_service_account_exists("ibm-spectrum-scale-csi-operator"):
            kubeobjectfunc.delete_service_account("ibm-spectrum-scale-csi-operator")
        kubeobjectfunc.check_service_account_deleted("ibm-spectrum-scale-csi-operator")

        if kubeobjectfunc.check_cluster_role_binding_exists("ibm-spectrum-scale-csi-operator"):
            kubeobjectfunc.delete_cluster_role_binding("ibm-spectrum-scale-csi-operator")
        kubeobjectfunc.check_cluster_role_binding_deleted("ibm-spectrum-scale-csi-operator")

        if kubeobjectfunc.check_cluster_role_exists("ibm-spectrum-scale-csi-operator"):
            kubeobjectfunc.delete_cluster_role("ibm-spectrum-scale-csi-operator")
        kubeobjectfunc.check_cluster_role_deleted("ibm-spectrum-scale-csi-operator")

        if kubeobjectfunc.check_deployment_exists():
            kubeobjectfunc.delete_deployment()
        kubeobjectfunc.check_deployment_deleted()

        if kubeobjectfunc.check_namespace_exists(self.operator_namespace) and (condition is False):
            kubeobjectfunc.delete_namespace(self.operator_namespace)
            kubeobjectfunc.check_namespace_deleted(self.operator_namespace)

    def check(self):

        config.load_kube_config(config_file=self.kubeconfig)
        kubeobjectfunc.check_namespace_exists(self.operator_namespace)
        kubeobjectfunc.check_deployment_exists()
        kubeobjectfunc.check_cluster_role_exists("ibm-spectrum-scale-csi-operator")
        kubeobjectfunc.check_service_account_exists("ibm-spectrum-scale-csi-operator")
        kubeobjectfunc.check_cluster_role_binding_exists("ibm-spectrum-scale-csi-operator")
        kubeobjectfunc.check_crd_exists()

    def get_operator_body(self):

        path = self.operator_yaml_file_path
        body = {}
        with open(path, 'r') as f:
            manifests = yaml.load_all(f, Loader=yaml.SafeLoader)
            for manifest in manifests:
                kind = manifest['kind']
                body[kind] = manifest
        return body


class Scaleoperatorobject:

    def __init__(self, test_dict, kubeconfig_value):

        self.kubeconfig = kubeconfig_value
        self.temp = test_dict
        self.secret_name = test_dict["local_secret_name"]
        self.operator_namespace = test_dict["namespace"]
        self.secret_data = {
            "username": test_dict["username"], "password": test_dict["password"]}

        csiobjectfunc.set_namespace_value(self.operator_namespace)

    def create(self):

        config.load_kube_config(config_file=self.kubeconfig)
        LOGGER.info(str(self.temp["custom_object_body"]["spec"]))

        if not(kubeobjectfunc.check_secret_exists(self.secret_name)):
            kubeobjectfunc.create_secret(self.secret_data, self.secret_name)
        else:
            kubeobjectfunc.delete_secret(self.secret_name)
            kubeobjectfunc.check_secret_is_deleted(self.secret_name)
            kubeobjectfunc.create_secret(self.secret_data, self.secret_name)

        for remote_secret_name in self.temp["remote_secret_names"]:
            remote_secret_data = {"username": self.temp["remote_username"][remote_secret_name],
                                  "password": self.temp["remote_password"][remote_secret_name]}
            if not(kubeobjectfunc.check_secret_exists(remote_secret_name)):
                kubeobjectfunc.create_secret(remote_secret_data, remote_secret_name)
            else:
                kubeobjectfunc.delete_secret(remote_secret_name)
                kubeobjectfunc.check_secret_is_deleted(remote_secret_name)
                kubeobjectfunc.create_secret(remote_secret_data, remote_secret_name)

        if "local_cacert_name" in self.temp:
            cacert_name = self.temp["local_cacert_name"]
            if "make_cacert_wrong" not in self.temp:
                self.temp["make_cacert_wrong"] = False

            if not(kubeobjectfunc.check_configmap_exists(cacert_name)):
                kubeobjectfunc.create_configmap(
                    self.temp["cacert_path"], self.temp["make_cacert_wrong"], cacert_name)
            else:
                kubeobjectfunc.delete_configmap(cacert_name)
                kubeobjectfunc.check_configmap_is_deleted(cacert_name)
                kubeobjectfunc.create_configmap(
                    self.temp["cacert_path"], self.temp["make_cacert_wrong"], cacert_name)

        for remote_cacert_name in self.temp["remote_cacert_names"]:

            if "make_remote_cacert_wrong" not in self.temp:
                self.temp["make_remote_cacert_wrong"] = False

            if not(kubeobjectfunc.check_configmap_exists(remote_cacert_name)):
                kubeobjectfunc.create_configmap(
                    self.temp["remote_cacert_path"][remote_cacert_name], self.temp["make_remote_cacert_wrong"], remote_cacert_name)
            else:
                kubeobjectfunc.delete_configmap(remote_cacert_name)
                kubeobjectfunc.check_configmap_is_deleted(remote_cacert_name)
                kubeobjectfunc.create_configmap(
                    self.temp["remote_cacert_path"][remote_cacert_name], self.temp["make_remote_cacert_wrong"], remote_cacert_name)

        if not(csiobjectfunc.check_scaleoperatorobject_is_deployed()):

            csiobjectfunc.create_custom_object(self.temp["custom_object_body"])
        else:
            csiobjectfunc.delete_custom_object()
            csiobjectfunc.check_scaleoperatorobject_is_deleted()
            csiobjectfunc.create_custom_object(self.temp["custom_object_body"])

    def delete(self):

        config.load_kube_config(config_file=self.kubeconfig)

        if csiobjectfunc.check_scaleoperatorobject_is_deployed():
            csiobjectfunc.delete_custom_object()
        csiobjectfunc.check_scaleoperatorobject_is_deleted()

        if kubeobjectfunc.check_secret_exists(self.secret_name):
            kubeobjectfunc.delete_secret(self.secret_name)
        kubeobjectfunc.check_secret_is_deleted(self.secret_name)

        for remote_secret_name in self.temp["remote_secret_names"]:
            if kubeobjectfunc.check_secret_exists(remote_secret_name):
                kubeobjectfunc.delete_secret(remote_secret_name)
            kubeobjectfunc.check_secret_is_deleted(remote_secret_name)
        if "local_cacert_name" in self.temp:
            if kubeobjectfunc.check_configmap_exists(self.temp["local_cacert_name"]):
                kubeobjectfunc.delete_configmap(self.temp["local_cacert_name"])
            kubeobjectfunc.check_configmap_is_deleted(self.temp["local_cacert_name"])

        for remote_cacert_name in self.temp["remote_cacert_names"]:
            if kubeobjectfunc.check_configmap_exists(remote_cacert_name):
                kubeobjectfunc.delete_configmap(remote_cacert_name)
            kubeobjectfunc.check_configmap_is_deleted(remote_cacert_name)

    def check(self, csiscaleoperator_name="ibm-spectrum-scale-csi"):
        config.load_kube_config(config_file=self.kubeconfig)

        is_deployed = csiobjectfunc.check_scaleoperatorobject_is_deployed(csiscaleoperator_name)
        if(is_deployed is False):
            return False

        kubeobjectfunc.get_pod_list_and_check_running("app=ibm-spectrum-scale-csi-attacher", 2)
        kubeobjectfunc.get_pod_list_and_check_running("app=ibm-spectrum-scale-csi-provisioner", 1)
        kubeobjectfunc.get_pod_list_and_check_running("app=ibm-spectrum-scale-csi-resizer", 1)
        kubeobjectfunc.get_pod_list_and_check_running("app=ibm-spectrum-scale-csi-snapshotter", 1)
        LOGGER.info("CSI driver Sidecar pods are Running")

        val, self.desired_number_scheduled = csiobjectfunc.check_scaleoperatorobject_daemonsets_state(
            csiscaleoperator_name)

        # kubeobjectfunc.check_pod_running("ibm-spectrum-scale-csi-snapshotter-0")

        return val

    def get_driver_ds_pod_name(self):

        try:
            pod_list_api_instance = client.CoreV1Api()
            pod_list_api_response = pod_list_api_instance.list_namespaced_pod(
                namespace=self.operator_namespace, pretty=True, field_selector="spec.serviceAccountName=ibm-spectrum-scale-csi-node")
            daemonset_pod_name = pod_list_api_response.items[0].metadata.name
            LOGGER.debug(str(pod_list_api_response))
            return daemonset_pod_name
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->list_namespaced_pod: {e}")
            assert False

    def get_scaleplugin_labeled_nodes(self, label):
        api_instance = client.CoreV1Api()
        label_selector = ""
        for label_val in label:
            label_selector += str(label_val["key"]) + \
                "="+str(label_val["value"])+","
        label_selector = label_selector[0:-1]
        try:
            api_response_2 = api_instance.list_node(
                pretty=True, label_selector=label_selector)
            LOGGER.info(f"{label_selector} labeled nodes are \
                        {str(len(api_response_2.items))}")
            LOGGER.info(f"{str(self.desired_number_scheduled)} daemonset nodes")
            return self.desired_number_scheduled, len(api_response_2.items)
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->list_node: {e}")
            assert False


class Driver:

    def __init__(self, kubeconfig_value, value_pvc, value_pod, cluster_id, test_ns, keep_object, image_name, plugin_nodeselector_labels):
        self.value_pvc = value_pvc
        self.value_pod = value_pod
        self.cluster_id = cluster_id
        self.test_ns = test_ns
        self.keep_objects = keep_object
        self.kubeconfig = kubeconfig_value
        self.image_name = image_name
        csistoragefunc.set_test_namespace_value(self.test_ns)
        csistoragefunc.set_test_nodeselector_value(plugin_nodeselector_labels)
        csistoragefunc.set_keep_objects(self.keep_objects)

    def test_dynamic(self, value_sc, value_pvc_passed=None, value_pod_passed=None, value_clone_passed=None):
        created_objects = get_cleanup_dict()
        if value_pvc_passed is None:
            value_pvc_passed = copy.deepcopy(self.value_pvc)
        if value_pod_passed is None:
            value_pod_passed = copy.deepcopy(self.value_pod)

        if "permissions" in value_sc.keys() and not(filesetfunc.feature_available("permissions")):
            LOGGER.warning(
                "Min required IBM Storage Scale version for permissions in storageclass support with CSI is 5.1.1-2")
            LOGGER.warning("Skipping Testcase")
            return

        LOGGER.info(
            f"Testing Dynamic Provisioning with following PVC parameters {str(value_pvc_passed)}")
        sc_name = csistoragefunc.get_random_name("sc")
        config.load_kube_config(config_file=self.kubeconfig)
        csistoragefunc.create_storage_class(value_sc, sc_name, created_objects)
        csistoragefunc.check_storage_class(sc_name)
        for num, _ in enumerate(value_pvc_passed):
            value_pvc_pass = copy.deepcopy(value_pvc_passed[num])
            if "reason" in value_sc:
                if "reason" not in value_pvc_pass:
                    value_pvc_pass["reason"] = value_sc["reason"]
            LOGGER.info(100*"=")
            pvc_name = csistoragefunc.get_random_name("pvc")
            csistoragefunc.create_pvc(value_pvc_pass, sc_name, pvc_name, created_objects)
            val = csistoragefunc.check_pvc(value_pvc_pass, pvc_name, created_objects)
            if val is True:
                csistoragefunc.check_permissions_for_pvc(
                        pvc_name, value_sc, created_objects)

                for num2, _ in enumerate(value_pod_passed):
                    LOGGER.info(100*"-")
                    pod_name = csistoragefunc.get_random_name("pod")
                    csistoragefunc.create_pod(
                        value_pod_passed[num2], pvc_name, pod_name, created_objects, self.image_name)
                    csistoragefunc.check_pod(value_pod_passed[num2], pod_name, created_objects)
                    if "volume_expansion_storage" in value_pvc_pass:
                        csistoragefunc.expand_and_check_pvc(sc_name, pvc_name, value_pvc_pass, "volume_expansion_storage",
                                                            pod_name, value_pod_passed[num2], created_objects)
                    if value_clone_passed is not None:
                        csistoragefunc.clone_and_check_pvc(
                            sc_name, value_sc, pvc_name, pod_name, value_pod_passed[num2], value_clone_passed, created_objects)
                    # and (self.keep_objects is True) and (num2 < (len(value_pod_passed)-1))):
                    if ((value_pvc_pass["access_modes"] == "ReadWriteOnce") and (num2 < (len(value_pod_passed)-1))):
                        pvc_name = csistoragefunc.get_random_name("pvc")
                        csistoragefunc.create_pvc(value_pvc_pass, sc_name,
                                                  pvc_name, created_objects)
                        val = csistoragefunc.check_pvc(value_pvc_pass, pvc_name, created_objects)
                        if val is not True:
                            break
                LOGGER.info(100*"-")
        LOGGER.info(100*"=")
        csistoragefunc.clean_with_created_objects(created_objects, condition="passed")

    def test_static(self, pv_value, pvc_value, sc_value=False, wrong=None, root_volume=False):

        if filesetfunc.get_scalevalidation() == "False":
            pytest.skip("As scalevalidation is False in config file , GUI communication not allowed. Static cases not possible")
        config.load_kube_config(config_file=self.kubeconfig)
        created_objects = get_cleanup_dict()
        sc_name = ""
        if sc_value is not False:
            sc_name = csistoragefunc.get_random_name("sc")
            csistoragefunc.create_storage_class(sc_value,  sc_name, created_objects)
            csistoragefunc.check_storage_class(sc_name)
        FSUID = filesetfunc.get_FSUID()
        cluster_id = self.cluster_id
        if wrong is not None:
            if wrong["id_wrong"] is True:
                cluster_id = int(cluster_id)+1
                cluster_id = str(cluster_id)
            if wrong["FSUID_wrong"] is True:
                FSUID = "AAAA"

        mount_point = filesetfunc.get_mount_point()
        if root_volume is False:
            dir_name = csistoragefunc.get_random_name("dir")
            filesetfunc.create_dir(dir_name)
            created_objects["dir"].append(dir_name)
            pv_value["volumeHandle"] = cluster_id+";"+FSUID + \
                ";path="+mount_point+"/"+dir_name
        elif root_volume is True:
            pv_value["volumeHandle"] = cluster_id+";"+FSUID + \
                ";path="+mount_point

        if pvc_value == "Default":
            pvc_value = copy.deepcopy(self.value_pvc)

        num_final = len(pvc_value)
        for num in range(0, num_final):
            pv_name = csistoragefunc.get_random_name("pv")
            csistoragefunc.create_pv(pv_value, pv_name, created_objects, sc_name)
            csistoragefunc.check_pv(pv_name)

            value_pvc_pass = copy.deepcopy(pvc_value[num])
            if "reason" in pv_value:
                if "reason" not in value_pvc_pass:
                    value_pvc_pass["reason"] = pv_value["reason"]
            LOGGER.info(100*"=")
            pvc_name = csistoragefunc.get_random_name("pvc")
            csistoragefunc.create_pvc(value_pvc_pass, sc_name, pvc_name, created_objects, pv_name)
            val = csistoragefunc.check_pvc(value_pvc_pass, pvc_name, created_objects, pv_name)
            if val is True:
                for num2 in range(0, len(self.value_pod)):
                    LOGGER.info(100*"-")
                    pod_name = csistoragefunc.get_random_name("pod")
                    csistoragefunc.create_pod(
                        self.value_pod[num2], pvc_name, pod_name, created_objects, self.image_name)
                    csistoragefunc.check_pod(self.value_pod[num2], pod_name, created_objects)
                    csistoragefunc.delete_pod(pod_name, created_objects)
                    csistoragefunc.check_pod_deleted(pod_name, created_objects)
                    if value_pvc_pass["access_modes"] == "ReadWriteOnce" and self.keep_objects == "True":
                        break
                LOGGER.info(100*"-")
            csistoragefunc.delete_pvc(pvc_name, created_objects)
            csistoragefunc.check_pvc_deleted(pvc_name, created_objects)
            csistoragefunc.delete_pv(pv_name, created_objects)
            csistoragefunc.check_pv_deleted(pv_name, created_objects)
        LOGGER.info(100*"=")
        csistoragefunc.clean_with_created_objects(created_objects, condition="passed")

    def one_pvc_two_pod(self, value_sc, value_pvc_pass, value_ds_pass):
        created_objects = get_cleanup_dict()
        sc_name = csistoragefunc.get_random_name("sc")
        config.load_kube_config(config_file=self.kubeconfig)
        csistoragefunc.create_storage_class(value_sc, sc_name, created_objects)
        csistoragefunc.check_storage_class(sc_name)
        pvc_name = csistoragefunc.get_random_name("pvc")
        csistoragefunc.create_pvc(value_pvc_pass, sc_name, pvc_name, created_objects)
        val = csistoragefunc.check_pvc(value_pvc_pass, pvc_name, created_objects)
        if val is True:
            ds_name = csistoragefunc.get_random_name("ds")
            csistoragefunc.create_ds(value_ds_pass, ds_name, pvc_name, created_objects)
            csistoragefunc.check_ds(ds_name, value_ds_pass, created_objects)
        csistoragefunc.clean_with_created_objects(created_objects, condition="passed")

    def parallel_pvc(self, value_sc, num_of_pvc, pod_creation=False):
        created_objects = get_cleanup_dict()
        sc_name = csistoragefunc.get_random_name("sc")
        config.load_kube_config(config_file=self.kubeconfig)
        csistoragefunc.create_storage_class(value_sc, sc_name, created_objects)
        csistoragefunc.check_storage_class(sc_name)
        pvc_names = []
        number_of_pvc = num_of_pvc
        common_pvc_name = csistoragefunc.get_random_name("pvc")
        for num in range(0, number_of_pvc):
            pvc_names.append(common_pvc_name+"-"+str(num))
        value_pvc_pass = copy.deepcopy(self.value_pvc[0])
        LOGGER.info(100*"-")
        value_pvc_pass["parallel"] = "True"

        for pvc_name in pvc_names:
            LOGGER.info(100*"-")
            csistoragefunc.create_pvc(value_pvc_pass, sc_name, pvc_name, created_objects)

        for pvc_name in pvc_names:
            LOGGER.info(100*"-")
            csistoragefunc.check_pvc(value_pvc_pass, pvc_name, created_objects)

        if pod_creation is False:
            csistoragefunc.clean_with_created_objects(created_objects, condition="passed")
            return

        pod_names = []

        for pvc_name in pvc_names:
            LOGGER.info(100*"-")
            pod_name = csistoragefunc.get_random_name("pod")
            pod_names.append(pod_name)
            csistoragefunc.create_pod(
                self.value_pod[0], pvc_name, pod_name, created_objects, self.image_name)
            csistoragefunc.check_pod(self.value_pod[0], pod_name, created_objects)
            csistoragefunc.delete_pod(pod_name, created_objects)
            csistoragefunc.check_pod_deleted(pod_name, created_objects)

        csistoragefunc.clean_with_created_objects(created_objects, condition="passed")


class Snapshot():
    def __init__(self, kubeconfig, test_namespace, keep_objects, value_pvc, value_vs_class, number_of_snapshots, image_name, cluster_id, plugin_nodeselector_labels):
        config.load_kube_config(config_file=kubeconfig)
        self.value_pvc = value_pvc
        self.value_vs_class = value_vs_class
        self.number_of_snapshots = number_of_snapshots
        self.image_name = image_name
        self.cluster_id = cluster_id
        csistoragefunc.set_test_namespace_value(test_namespace)
        csistoragefunc.set_test_nodeselector_value(plugin_nodeselector_labels)
        csistoragefunc.set_keep_objects(keep_objects)

    def test_dynamic(self, value_sc, test_restore, value_vs_class=None, number_of_snapshots=None, reason=None, restore_sc=None, restore_pvc=None, value_pod=None, value_pvc=None, value_clone_passed=None):
        if value_vs_class is None:
            value_vs_class = self.value_vs_class
        if number_of_snapshots is None:
            number_of_snapshots = self.number_of_snapshots
        number_of_restore = 1

        if "permissions" in value_sc.keys() and not(filesetfunc.feature_available("permissions")):
            LOGGER.warning(
                "Min required IBM Storage Scale version for permissions in storageclass support with CSI is 5.1.1-2")
            LOGGER.warning("Skipping Testcase")
            return

        if value_pvc is None:
            value_pvc = copy.deepcopy(self.value_pvc)

        created_objects = get_cleanup_dict()
        for pvc_value in value_pvc:

            LOGGER.info("-"*100)
            sc_name = csistoragefunc.get_random_name("sc")
            csistoragefunc.create_storage_class(value_sc, sc_name, created_objects)
            csistoragefunc.check_storage_class(sc_name)

            pvc_name = csistoragefunc.get_random_name("pvc")
            csistoragefunc.create_pvc(pvc_value, sc_name, pvc_name, created_objects)
            val = csistoragefunc.check_pvc(pvc_value, pvc_name, created_objects)

            if val is True:
                csistoragefunc.check_permissions_for_pvc(
                    pvc_name, value_sc, created_objects)

            pod_name = csistoragefunc.get_random_name("snap-start-pod")
            if value_pod is None:
                value_pod = {"mount_path": "/usr/share/nginx/html/scale", "read_only": "False"}

            csistoragefunc.create_pod(value_pod, pvc_name, pod_name,
                                      created_objects, self.image_name)
            csistoragefunc.check_pod(value_pod, pod_name, created_objects)
            csistoragefunc.create_file_inside_pod(value_pod, pod_name, created_objects)

            if "presnap_volume_expansion_storage" in pvc_value:
                csistoragefunc.expand_and_check_pvc(sc_name, pvc_name, pvc_value, "presnap_volume_expansion_storage",
                                                    pod_name, value_pod, created_objects)

            vs_class_name = csistoragefunc.get_random_name("vsclass")
            csistoragefunc.create_vs_class(vs_class_name, value_vs_class, created_objects)
            csistoragefunc.check_vs_class(vs_class_name)

            if not(filesetfunc.feature_available("snapshot")):
                if reason is None:
                    reason = "Min required IBM Storage Scale version for snapshot support with CSI is 5.1.1-0"
                test_restore = False

            vs_name = csistoragefunc.get_random_name("vs")
            for num in range(0, number_of_snapshots):
                csistoragefunc.create_vs(vs_name+"-"+str(num), vs_class_name,
                                         pvc_name, created_objects)
                csistoragefunc.check_vs_detail(
                    vs_name+"-"+str(num), pvc_name, value_vs_class, reason, created_objects)

            if test_restore:
                restore_sc_name = sc_name
                if restore_sc is not None:
                    restore_sc_name = "restore-" + restore_sc_name
                    csistoragefunc.create_storage_class(
                        restore_sc, restore_sc_name, created_objects)
                    csistoragefunc.check_storage_class(restore_sc_name)
                else:
                    restore_sc = value_sc
                if restore_pvc is not None:
                    pvc_value = restore_pvc

                for num in range(0, number_of_restore):
                    restored_pvc_name = "restored-pvc"+vs_name[2:]+"-"+str(num)
                    snap_pod_name = "snap-end-pod"+vs_name[2:]
                    csistoragefunc.create_pvc_from_snapshot(
                        pvc_value, restore_sc_name, restored_pvc_name, vs_name+"-"+str(num), created_objects)
                    if reason is not None:
                        pvc_value["reason"]=reason
                    val = csistoragefunc.check_pvc(pvc_value, restored_pvc_name, created_objects)

                    if val is True:
                        csistoragefunc.check_permissions_for_pvc(
                            pvc_name, value_sc, created_objects)
                        csistoragefunc.create_pod(
                            value_pod, restored_pvc_name, snap_pod_name, created_objects, self.image_name)
                        csistoragefunc.check_pod(value_pod, snap_pod_name, created_objects)
                        csistoragefunc.check_file_inside_pod(
                            value_pod, snap_pod_name, created_objects)

                        if "postsnap_volume_expansion_storage" in pvc_value:
                            csistoragefunc.expand_and_check_pvc(restore_sc_name, restored_pvc_name, pvc_value, "postsnap_volume_expansion_storage",
                                                                snap_pod_name, value_pod, created_objects)

                        if "post_presnap_volume_expansion_storage" in pvc_value:
                            csistoragefunc.expand_and_check_pvc(sc_name, pvc_name, pvc_value, "post_presnap_volume_expansion_storage",
                                                                pod_name, value_pod, created_objects)

                        if value_clone_passed is not None:
                            csistoragefunc.clone_and_check_pvc(
                                restore_sc_name, restore_sc, restored_pvc_name, snap_pod_name, value_pod, value_clone_passed, created_objects)

        csistoragefunc.clean_with_created_objects(created_objects, condition="passed")

    def test_static(self, value_sc, test_restore, value_vs_class=None, number_of_snapshots=None, restore_sc=None, restore_pvc=None, reason=None):

        if filesetfunc.get_scalevalidation() == "False":
            pytest.skip("As scalevalidation is False in config file , GUI communication not allowed. Static cases not possible")
        if value_vs_class is None:
            value_vs_class = self.value_vs_class
        if number_of_snapshots is None:
            number_of_snapshots = self.number_of_snapshots
        number_of_restore = 1

        for pvc_value in self.value_pvc:

            created_objects = get_cleanup_dict()
            LOGGER.info("-"*100)
            sc_name = csistoragefunc.get_random_name("sc")
            csistoragefunc.create_storage_class(value_sc, sc_name, created_objects)
            csistoragefunc.check_storage_class(sc_name)

            pvc_name = csistoragefunc.get_random_name("pvc")
            csistoragefunc.create_pvc(pvc_value, sc_name, pvc_name, created_objects)
            csistoragefunc.check_pvc(pvc_value, pvc_name, created_objects)

            pod_name = csistoragefunc.get_random_name("snap-start-pod")
            value_pod = {"mount_path": "/usr/share/nginx/html/scale", "read_only": "False"}
            csistoragefunc.create_pod(value_pod, pvc_name, pod_name,
                                      created_objects, self.image_name)
            csistoragefunc.check_pod(value_pod, pod_name, created_objects)
            csistoragefunc.create_file_inside_pod(value_pod, pod_name, created_objects)

            snapshot_name = csistoragefunc.get_random_name("snapshot")
            volume_name = csistoragefunc.get_pv_for_pvc(pvc_name, created_objects)
            fileset_name = csistoragefunc.get_filesetname_from_pv(volume_name, created_objects)

            FSUID = filesetfunc.get_FSUID()
            cluster_id = self.cluster_id
            vs_content_name = csistoragefunc.get_random_name("vscontent")

            vs_name = csistoragefunc.get_random_name("vs")
            for num in range(0, number_of_snapshots):
                filesetfunc.create_snapshot(snapshot_name+"-"+str(num),
                                            fileset_name, created_objects)
                if filesetfunc.check_snapshot_exists(snapshot_name+"-"+str(num), fileset_name):
                    LOGGER.info(f"snapshot {snapshot_name} exists for {fileset_name}")
                else:
                    LOGGER.error(f"snapshot {snapshot_name} does not exists for {fileset_name}")
                    csistoragefunc.clean_with_created_objects(created_objects, condition="failed")
                    assert False

                snapshot_handle = cluster_id+';'+FSUID+';' + \
                    fileset_name+';'+snapshot_name+"-"+str(num)
                body_params = {"deletionPolicy": "Retain", "snapshotHandle": snapshot_handle}
                csistoragefunc.create_vs_content(
                    vs_content_name+"-"+str(num), vs_name+"-"+str(num), body_params, created_objects)
                csistoragefunc.check_vs_content(vs_content_name+"-"+str(num))

                csistoragefunc.create_vs_from_content(
                    vs_name+"-"+str(num), vs_content_name+"-"+str(num), created_objects)
                csistoragefunc.check_vs_detail_for_static(vs_name+"-"+str(num), created_objects)

            if not(filesetfunc.feature_available("snapshot")):
                pvc_value["reason"] = "Min required IBM Storage Scale version for snapshot support with CSI is 5.1.1-0"

            if test_restore:
                if restore_sc is not None:
                    sc_name = "restore-" + sc_name
                    csistoragefunc.create_storage_class(restore_sc, sc_name, created_objects)
                    csistoragefunc.check_storage_class(sc_name)
                if restore_pvc is not None:
                    pvc_value = restore_pvc

                for num in range(0, number_of_restore):
                    restored_pvc_name = "restored-pvc"+vs_name[2:]+"-"+str(num)
                    snap_pod_name = "snap-end-pod"+vs_name[2:]
                    csistoragefunc.create_pvc_from_snapshot(
                        pvc_value, sc_name, restored_pvc_name, vs_name+"-"+str(num), created_objects)
                    if reason is not None:
                        pvc_value["reason"] = reason
                    val = csistoragefunc.check_pvc(pvc_value, restored_pvc_name, created_objects)
                    if val is True:
                        csistoragefunc.create_pod(
                            value_pod, restored_pvc_name, snap_pod_name, created_objects, self.image_name)
                        csistoragefunc.check_pod(value_pod, snap_pod_name, created_objects)
                        csistoragefunc.check_file_inside_pod(
                            value_pod, snap_pod_name, created_objects, fileset_name)
                        csistoragefunc.delete_pod(snap_pod_name, created_objects)
                        csistoragefunc.check_pod_deleted(snap_pod_name, created_objects)
                    csistoragefunc.delete_pvc(restored_pvc_name, created_objects)
                    csistoragefunc.check_pvc_deleted(restored_pvc_name, created_objects)

            csistoragefunc.clean_with_created_objects(created_objects, condition="passed")


def get_cleanup_dict():
    created_object = {
        "sc": [],
        "pvc": [],
        "pod": [],
        "vs": [],
        "vsclass": [],
        "vscontent": [],
        "scalesnapshot": [],
        "restore_pod": [],
        "restore_pvc": [],
        "clone_pod": [],
        "clone_pvc": [],
        "pv": [],
        "dir": [],
        "ds": [],
        "cg": [],
        "fileset": []
    }
    return created_object
