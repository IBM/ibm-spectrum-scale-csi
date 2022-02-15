import copy
import logging
import os.path
import yaml
from kubernetes import client, config
from kubernetes.client.rest import ApiException
import ibm_spectrum_scale_csi.kubernetes_apis.scale_operator_function as csioperatorfunc
import ibm_spectrum_scale_csi.kubernetes_apis.scale_operator_object_function as csiobjectfunc
import ibm_spectrum_scale_csi.kubernetes_apis.volume_functions as volfunc 
import ibm_spectrum_scale_csi.kubernetes_apis.snapshot_functions as snapshotfunc
import ibm_spectrum_scale_csi.spectrum_scale_apis.fileset_functions as filesetfunc
import ibm_spectrum_scale_csi.kubernetes_apis.cleanup_functions as cleanup
LOGGER = logging.getLogger()


class Scaleoperator:
    def __init__(self, kubeconfig_value, namespace_value, operator_yaml):

        self.kubeconfig = kubeconfig_value
        csioperatorfunc.set_global_namespace_value(namespace_value)
        csiobjectfunc.set_namespace_value(namespace_value)
        self.operator_yaml_file_path = operator_yaml
        crd_body = self.get_operator_body()
        crd_full_version = crd_body["CustomResourceDefinition"]["apiVersion"].split("/")
        self.crd_version = crd_full_version[1]

    def create(self):

        config.load_kube_config(config_file=self.kubeconfig)

        body = self.get_operator_body()
        if not(csioperatorfunc.check_namespace_exists()):
            csioperatorfunc.create_namespace()

        if not(csioperatorfunc.check_deployment_exists()):
            csioperatorfunc.create_deployment(body['Deployment'])

        if not(csioperatorfunc.check_cluster_role_exists("ibm-spectrum-scale-csi-operator")):
            csioperatorfunc.create_cluster_role(body['ClusterRole'])

        if not(csioperatorfunc.check_service_account_exists("ibm-spectrum-scale-csi-operator")):
            csioperatorfunc.create_service_account(body['ServiceAccount'])

        if not(csioperatorfunc.check_cluster_role_binding_exists("ibm-spectrum-scale-csi-operator")):
            csioperatorfunc.create_cluster_role_binding(body['ClusterRoleBinding'])

        if not(csioperatorfunc.check_crd_exists(self.crd_version)):
            csioperatorfunc.create_crd(body['CustomResourceDefinition'])

    def delete(self, condition=False):

        config.load_kube_config(config_file=self.kubeconfig)
        if csiobjectfunc.check_scaleoperatorobject_is_deployed():   # for edge cases if custom object is not deleted
            csiobjectfunc.delete_custom_object()
            csiobjectfunc.check_scaleoperatorobject_is_deleted()

        if csioperatorfunc.check_crd_exists(self.crd_version):
            csioperatorfunc.delete_crd(self.crd_version)
        csioperatorfunc.check_crd_deleted(self.crd_version)

        if csioperatorfunc.check_service_account_exists("ibm-spectrum-scale-csi-operator"):
            csioperatorfunc.delete_service_account("ibm-spectrum-scale-csi-operator")
        csioperatorfunc.check_service_account_deleted("ibm-spectrum-scale-csi-operator")

        if csioperatorfunc.check_cluster_role_binding_exists("ibm-spectrum-scale-csi-operator"):
            csioperatorfunc.delete_cluster_role_binding("ibm-spectrum-scale-csi-operator")
        csioperatorfunc.check_cluster_role_binding_deleted("ibm-spectrum-scale-csi-operator")

        if csioperatorfunc.check_cluster_role_exists("ibm-spectrum-scale-csi-operator"):
            csioperatorfunc.delete_cluster_role("ibm-spectrum-scale-csi-operator")
        csioperatorfunc.check_cluster_role_deleted("ibm-spectrum-scale-csi-operator")

        if csioperatorfunc.check_deployment_exists():
            csioperatorfunc.delete_deployment()
        csioperatorfunc.check_deployment_deleted()

        if csioperatorfunc.check_namespace_exists() and (condition is False):
            csioperatorfunc.delete_namespace()
            csioperatorfunc.check_namespace_deleted()

    def check(self):

        config.load_kube_config(config_file=self.kubeconfig)
        csioperatorfunc.check_namespace_exists()
        csioperatorfunc.check_deployment_exists()
        csioperatorfunc.check_cluster_role_exists("ibm-spectrum-scale-csi-operator")
        csioperatorfunc.check_service_account_exists("ibm-spectrum-scale-csi-operator")
        csioperatorfunc.check_cluster_role_binding_exists("ibm-spectrum-scale-csi-operator")
        csioperatorfunc.check_crd_exists(self.crd_version)

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
        self.namespace = test_dict["namespace"]
        self.secret_data = {
            "username": test_dict["username"], "password": test_dict["password"]}

        if check_key(test_dict, "stateful_set_not_created"):
            self.stateful_set_not_created = test_dict["stateful_set_not_created"]
        else:
            self.stateful_set_not_created = False

        csiobjectfunc.set_namespace_value(self.namespace)

    def create(self):

        config.load_kube_config(config_file=self.kubeconfig)
        LOGGER.info(str(self.temp["custom_object_body"]["spec"]))

        if not(csiobjectfunc.check_secret_exists(self.secret_name)):
            csiobjectfunc.create_secret(self.secret_data, self.secret_name)
        else:
            csiobjectfunc.delete_secret(self.secret_name)
            csiobjectfunc.check_secret_is_deleted(self.secret_name)
            csiobjectfunc.create_secret(self.secret_data, self.secret_name)

        for remote_secret_name in self.temp["remote_secret_names"]:
            remote_secret_data = {"username": self.temp["remote_username"][remote_secret_name],
                                  "password": self.temp["remote_password"][remote_secret_name]}
            if not(csiobjectfunc.check_secret_exists(remote_secret_name)):
                csiobjectfunc.create_secret(remote_secret_data, remote_secret_name)
            else:
                csiobjectfunc.delete_secret(remote_secret_name)
                csiobjectfunc.check_secret_is_deleted(remote_secret_name)
                csiobjectfunc.create_secret(remote_secret_data, remote_secret_name)

        if check_key(self.temp, "local_cacert_name"):
            cacert_name = self.temp["local_cacert_name"]
            if not(check_key(self.temp, "make_cacert_wrong")):
                self.temp["make_cacert_wrong"] = False

            if not(csiobjectfunc.check_configmap_exists(cacert_name)):
                csiobjectfunc.create_configmap(
                    self.temp["cacert_path"], self.temp["make_cacert_wrong"], cacert_name)
            else:
                csiobjectfunc.delete_configmap(cacert_name)
                csiobjectfunc.check_configmap_is_deleted(cacert_name)
                csiobjectfunc.create_configmap(
                    self.temp["cacert_path"], self.temp["make_cacert_wrong"], cacert_name)

        for remote_cacert_name in self.temp["remote_cacert_names"]:

            if not(check_key(self.temp, "make_remote_cacert_wrong")):
                self.temp["make_remote_cacert_wrong"] = False

            if not(csiobjectfunc.check_configmap_exists(remote_cacert_name)):
                csiobjectfunc.create_configmap(
                    self.temp["remote_cacert_path"][remote_cacert_name], self.temp["make_remote_cacert_wrong"], remote_cacert_name)
            else:
                csiobjectfunc.delete_configmap(remote_cacert_name)
                csiobjectfunc.check_configmap_is_deleted(remote_cacert_name)
                csiobjectfunc.create_configmap(
                    self.temp["remote_cacert_path"][remote_cacert_name], self.temp["make_remote_cacert_wrong"], remote_cacert_name)

        if not(csiobjectfunc.check_scaleoperatorobject_is_deployed()):

            csiobjectfunc.create_custom_object(self.temp["custom_object_body"], self.stateful_set_not_created)
        else:
            csiobjectfunc.delete_custom_object()
            csiobjectfunc.check_scaleoperatorobject_is_deleted()
            csiobjectfunc.create_custom_object(self.temp["custom_object_body"], self.stateful_set_not_created)

    def delete(self):

        config.load_kube_config(config_file=self.kubeconfig)

        if csiobjectfunc.check_scaleoperatorobject_is_deployed():
            csiobjectfunc.delete_custom_object()
        csiobjectfunc.check_scaleoperatorobject_is_deleted()

        if csiobjectfunc.check_secret_exists(self.secret_name):
            csiobjectfunc.delete_secret(self.secret_name)
        csiobjectfunc.check_secret_is_deleted(self.secret_name)

        for remote_secret_name in self.temp["remote_secret_names"]:
            if csiobjectfunc.check_secret_exists(remote_secret_name):
                csiobjectfunc.delete_secret(remote_secret_name)
            csiobjectfunc.check_secret_is_deleted(remote_secret_name)
        if check_key(self.temp, "local_cacert_name"):
            if csiobjectfunc.check_configmap_exists(self.temp["local_cacert_name"]):
                csiobjectfunc.delete_configmap(self.temp["local_cacert_name"])
            csiobjectfunc.check_configmap_is_deleted(self.temp["local_cacert_name"])

        for remote_cacert_name in self.temp["remote_cacert_names"]:
            if csiobjectfunc.check_configmap_exists(remote_cacert_name):
                csiobjectfunc.delete_configmap(remote_cacert_name)
            csiobjectfunc.check_configmap_is_deleted(remote_cacert_name)

    def check(self, csiscaleoperator_name="ibm-spectrum-scale-csi"):
        config.load_kube_config(config_file=self.kubeconfig)

        is_deployed = csiobjectfunc.check_scaleoperatorobject_is_deployed(csiscaleoperator_name)
        if(is_deployed is False):
            return False

        csiobjectfunc.check_scaleoperatorobject_statefulsets_state(
            csiscaleoperator_name+"-attacher")

        csiobjectfunc.check_scaleoperatorobject_statefulsets_state(
            csiscaleoperator_name+"-provisioner")

        csiobjectfunc.check_scaleoperatorobject_statefulsets_state(
            csiscaleoperator_name+"-snapshotter")

        csiobjectfunc.check_scaleoperatorobject_statefulsets_state(
            csiscaleoperator_name+"-resizer")

        val, self.desired_number_scheduled = csiobjectfunc.check_scaleoperatorobject_daemonsets_state(csiscaleoperator_name)

        # csiobjectfunc.check_pod_running("ibm-spectrum-scale-csi-snapshotter-0")

        return val

    def get_driver_ds_pod_name(self):

        try:
            pod_list_api_instance = client.CoreV1Api()
            pod_list_api_response = pod_list_api_instance.list_namespaced_pod(
                namespace=self.namespace, pretty=True, field_selector="spec.serviceAccountName=ibm-spectrum-scale-csi-node")
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
        volfunc.set_test_namespace_value(self.test_ns)
        volfunc.set_test_nodeselector_value(plugin_nodeselector_labels)
        cleanup.set_keep_objects(self.keep_objects)
        cleanup.set_test_namespace_value(self.test_ns)

    def create_test_ns(self, kubeconfig):
        config.load_kube_config(config_file=kubeconfig)
        csioperatorfunc.set_global_namespace_value(self.test_ns)
        if not(csioperatorfunc.check_namespace_exists()):
            csioperatorfunc.create_namespace()

    def delete_test_ns(self, kubeconfig):
        config.load_kube_config(config_file=kubeconfig)
        csioperatorfunc.set_global_namespace_value(self.test_ns)
        if csioperatorfunc.check_namespace_exists():
            csioperatorfunc.delete_namespace()
        csioperatorfunc.check_namespace_deleted()

    def test_dynamic(self, value_sc, value_pvc_passed=None, value_pod_passed=None, value_clone_passed=None):
        created_objects = get_cleanup_dict()
        if value_pvc_passed is None:
            value_pvc_passed = copy.deepcopy(self.value_pvc)
        if value_pod_passed is None:
            value_pod_passed = copy.deepcopy(self.value_pod)

        if "permissions" in value_sc.keys() and not(filesetfunc.feature_available("permissions")):
            LOGGER.warning("Min required Spectrum Scale version for permissions in storageclass support with CSI is 5.1.1-2")
            LOGGER.warning("Skipping Testcase")
            return

        LOGGER.info(
            f"Testing Dynamic Provisioning with following PVC parameters {str(value_pvc_passed)}")
        sc_name = volfunc.get_random_name("sc")
        config.load_kube_config(config_file=self.kubeconfig)
        volfunc.create_storage_class(value_sc, sc_name, created_objects)
        volfunc.check_storage_class(sc_name)
        for num, _ in enumerate(value_pvc_passed):
            value_pvc_pass = copy.deepcopy(value_pvc_passed[num])
            if (check_key(value_sc, "reason")):
                if not(check_key(value_pvc_pass, "reason")):
                    value_pvc_pass["reason"] = value_sc["reason"]
            LOGGER.info(100*"=")
            pvc_name = volfunc.get_random_name("pvc")
            volfunc.create_pvc(value_pvc_pass, sc_name, pvc_name, created_objects)
            val = volfunc.check_pvc(value_pvc_pass, pvc_name, created_objects)
            if val is True:
                if "permissions" in value_sc.keys():
                    volfunc.check_permissions_for_pvc(pvc_name, value_sc["permissions"], created_objects)

                for num2, _ in enumerate(value_pod_passed):
                    LOGGER.info(100*"-")
                    pod_name = volfunc.get_random_name("pod")
                    if value_sc.keys() >= {"permissions", "gid", "uid"}:
                        value_pod_passed[num2]["gid"] = value_sc["gid"]
                        value_pod_passed[num2]["uid"] = value_sc["uid"]
                    volfunc.create_pod(value_pod_passed[num2], pvc_name, pod_name, created_objects, self.image_name)
                    volfunc.check_pod(value_pod_passed[num2], pod_name, created_objects)
                    if "volume_expansion_storage" in value_pvc_pass:
                        volfunc.expand_and_check_pvc(sc_name, pvc_name, value_pvc_pass, "volume_expansion_storage",
                                               pod_name, value_pod_passed[num2], created_objects)
                    if value_clone_passed is not None:
                        volfunc.clone_and_check_pvc(sc_name, value_sc, pvc_name, pod_name, value_pod_passed[num2], value_clone_passed, created_objects)
                    #cleanup.delete_pod(pod_name, created_objects)
                    #cleanup.check_pod_deleted(pod_name, created_objects)
                    if ((value_pvc_pass["access_modes"] == "ReadWriteOnce") and (num2 < (len(value_pod_passed)-1))):#and (self.keep_objects is True) and (num2 < (len(value_pod_passed)-1))):
                        pvc_name = volfunc.get_random_name("pvc")
                        volfunc.create_pvc(value_pvc_pass, sc_name, pvc_name, created_objects)
                        val = volfunc.check_pvc(value_pvc_pass, pvc_name, created_objects)
                        if val is not True:
                            break
                LOGGER.info(100*"-")
            #vol_name = cleanup.delete_pvc(pvc_name, created_objects)
            #cleanup.check_pvc_deleted(pvc_name, vol_name, created_objects)
        LOGGER.info(100*"=")
        cleanup.clean_with_created_objects(created_objects)

    def test_static(self, pv_value, pvc_value, sc_value=False, wrong=None, root_volume=False):

        config.load_kube_config(config_file=self.kubeconfig)
        created_objects = get_cleanup_dict()
        sc_name = ""
        if sc_value is not False:
            sc_name = volfunc.get_random_name("sc")
            volfunc.create_storage_class(sc_value,  sc_name, created_objects)
            volfunc.check_storage_class(sc_name)
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
            dir_name = volfunc.get_random_name("dir")
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
            pv_name = volfunc.get_random_name("pv")
            volfunc.create_pv(pv_value, pv_name, created_objects, sc_name)
            volfunc.check_pv(pv_name)

            value_pvc_pass = copy.deepcopy(pvc_value[num])
            if (check_key(pv_value, "reason")):
                if not(check_key(value_pvc_pass, "reason")):
                    value_pvc_pass["reason"] = pv_value["reason"]
            LOGGER.info(100*"=")
            pvc_name = volfunc.get_random_name("pvc")
            volfunc.create_pvc(value_pvc_pass, sc_name, pvc_name, created_objects, pv_name)
            val = volfunc.check_pvc(value_pvc_pass, pvc_name, created_objects, pv_name)
            if val is True:
                for num2 in range(0, len(self.value_pod)):
                    LOGGER.info(100*"-")
                    pod_name = volfunc.get_random_name("pod")
                    volfunc.create_pod(self.value_pod[num2], pvc_name, pod_name, created_objects, self.image_name)
                    volfunc.check_pod(self.value_pod[num2], pod_name, created_objects)
                    cleanup.delete_pod(pod_name, created_objects)
                    cleanup.check_pod_deleted(pod_name, created_objects)
                    if value_pvc_pass["access_modes"] == "ReadWriteOnce" and self.keep_objects is True:
                        break
                LOGGER.info(100*"-")
            vol_name = cleanup.delete_pvc(pvc_name, created_objects)
            cleanup.check_pvc_deleted(pvc_name, vol_name, created_objects)
            cleanup.delete_pv(pv_name, created_objects)
            cleanup.check_pv_deleted(pv_name, created_objects)
        LOGGER.info(100*"=")
        cleanup.clean_with_created_objects(created_objects)

    def one_pvc_two_pod(self, value_sc, value_pvc_pass, value_ds_pass):
        created_objects = get_cleanup_dict()
        sc_name = volfunc.get_random_name("sc")
        config.load_kube_config(config_file=self.kubeconfig)
        volfunc.create_storage_class(value_sc, sc_name, created_objects)
        volfunc.check_storage_class(sc_name)
        pvc_name = volfunc.get_random_name("pvc")
        volfunc.create_pvc(value_pvc_pass, sc_name, pvc_name, created_objects)
        val = volfunc.check_pvc(value_pvc_pass, pvc_name, created_objects)
        if val is True:
            ds_name = volfunc.get_random_name("ds")
            volfunc.create_ds(value_ds_pass, ds_name, pvc_name, created_objects)
            volfunc.check_ds(ds_name, value_ds_pass, created_objects)
        cleanup.clean_with_created_objects(created_objects)

    def parallel_pvc(self, value_sc, num_of_pvc, pod_creation=False):
        created_objects = get_cleanup_dict()
        sc_name = volfunc.get_random_name("sc")
        config.load_kube_config(config_file=self.kubeconfig)
        volfunc.create_storage_class(value_sc, sc_name, created_objects)
        volfunc.check_storage_class(sc_name)
        pvc_names = []
        number_of_pvc = num_of_pvc
        common_pvc_name = volfunc.get_random_name("pvc")
        for num in range(0, number_of_pvc):
            pvc_names.append(common_pvc_name+"-"+str(num))
        value_pvc_pass = copy.deepcopy(self.value_pvc[0])
        LOGGER.info(100*"-")
        value_pvc_pass["parallel"] = "True"

        for pvc_name in pvc_names:
            LOGGER.info(100*"-")
            volfunc.create_pvc(value_pvc_pass, sc_name, pvc_name, created_objects)

        for pvc_name in pvc_names:
            LOGGER.info(100*"-")
            volfunc.check_pvc(value_pvc_pass, pvc_name, created_objects)

        if pod_creation is False:
            cleanup.clean_with_created_objects(created_objects)
            return

        pod_names = []

        for pvc_name in pvc_names:
            LOGGER.info(100*"-")
            pod_name = volfunc.get_random_name("pod")
            pod_names.append(pod_name)
            volfunc.create_pod(self.value_pod[0], pvc_name, pod_name, created_objects, self.image_name)
            volfunc.check_pod(self.value_pod[0], pod_name, created_objects)
            cleanup.delete_pod(pod_name, created_objects)
            cleanup.check_pod_deleted(pod_name, created_objects)

        cleanup.clean_with_created_objects(created_objects)


class Snapshot():
    def __init__(self, kubeconfig, test_namespace, keep_objects, value_pvc, value_vs_class, number_of_snapshots, image_name, cluster_id, plugin_nodeselector_labels):
        config.load_kube_config(config_file=kubeconfig)
        self.value_pvc = value_pvc
        self.value_vs_class = value_vs_class
        self.number_of_snapshots = number_of_snapshots
        self.image_name = image_name
        self.cluster_id = cluster_id
        volfunc.set_test_namespace_value(test_namespace)
        volfunc.set_test_nodeselector_value(plugin_nodeselector_labels)
        snapshotfunc.set_test_namespace_value(test_namespace)
        cleanup.set_keep_objects(keep_objects)
        cleanup.set_test_namespace_value(test_namespace)

    def test_dynamic(self, value_sc, test_restore, value_vs_class=None, number_of_snapshots=None, reason=None, restore_sc=None, restore_pvc=None, value_pod=None, value_pvc=None, value_clone_passed=None):
        if value_vs_class is None:
            value_vs_class = self.value_vs_class
        if number_of_snapshots is None:
            number_of_snapshots = self.number_of_snapshots
        number_of_restore = 1

        if "permissions" in value_sc.keys() and not(filesetfunc.feature_available("permissions")):
            LOGGER.warning("Min required Spectrum Scale version for permissions in storageclass support with CSI is 5.1.1-2")
            LOGGER.warning("Skipping Testcase")
            return

        if value_pvc is None:
            value_pvc = copy.deepcopy(self.value_pvc)

        for pvc_value in value_pvc:

            created_objects = get_cleanup_dict()
            LOGGER.info("-"*100)
            sc_name = volfunc.get_random_name("sc")
            volfunc.create_storage_class(value_sc, sc_name, created_objects)
            volfunc.check_storage_class(sc_name)

            pvc_name = volfunc.get_random_name("pvc")
            volfunc.create_pvc(pvc_value, sc_name, pvc_name, created_objects)
            val = volfunc.check_pvc(pvc_value, pvc_name, created_objects)

            if val is True and "permissions" in value_sc.keys():
                volfunc.check_permissions_for_pvc(pvc_name, value_sc["permissions"], created_objects)

            pod_name = volfunc.get_random_name("snap-start-pod")
            if value_pod is None:
                value_pod = {"mount_path": "/usr/share/nginx/html/scale", "read_only": "False"}

            if value_sc.keys() >= {"permissions", "gid", "uid"}:
                value_pod["gid"] = value_sc["gid"]
                value_pod["uid"] = value_sc["uid"]
            volfunc.create_pod(value_pod, pvc_name, pod_name, created_objects, self.image_name)
            volfunc.check_pod(value_pod, pod_name, created_objects)
            volfunc.create_file_inside_pod(value_pod, pod_name, created_objects)

            if "presnap_volume_expansion_storage" in pvc_value:
                volfunc.expand_and_check_pvc(sc_name, pvc_name, pvc_value, "presnap_volume_expansion_storage",
                                       pod_name, value_pod, created_objects)

            vs_class_name = volfunc.get_random_name("vsclass")
            snapshotfunc.create_vs_class(vs_class_name, value_vs_class, created_objects)
            snapshotfunc.check_vs_class(vs_class_name)

            if not(filesetfunc.feature_available("snapshot")):
                if reason is None:
                    reason = "Min required Spectrum Scale version for snapshot support with CSI is 5.1.1-0"
                test_restore = False

            vs_name = volfunc.get_random_name("vs")
            for num in range(0, number_of_snapshots):
                snapshotfunc.create_vs(vs_name+"-"+str(num), vs_class_name, pvc_name, created_objects)
                snapshotfunc.check_vs_detail(vs_name+"-"+str(num), pvc_name, value_vs_class, reason, created_objects)

            if test_restore:
                restore_sc_name = sc_name
                if restore_sc is not None:
                    restore_sc_name = "restore-" + restore_sc_name
                    volfunc.create_storage_class(restore_sc, restore_sc_name, created_objects)
                    volfunc.check_storage_class(restore_sc_name)
                else:
                    restore_sc = value_sc
                if restore_pvc is not None:
                    pvc_value = restore_pvc

                for num in range(0, number_of_restore):
                    restored_pvc_name = "restored-pvc"+vs_name[2:]+"-"+str(num)
                    snap_pod_name = "snap-end-pod"+vs_name[2:]
                    volfunc.create_pvc_from_snapshot(pvc_value, restore_sc_name, restored_pvc_name, vs_name+"-"+str(num), created_objects)
                    val = volfunc.check_pvc(pvc_value, restored_pvc_name, created_objects)
                    if val is True and "permissions" in value_sc.keys():
                        volfunc.check_permissions_for_pvc(pvc_name, value_sc["permissions"], created_objects)

                    if val is True:
                        volfunc.create_pod(value_pod, restored_pvc_name, snap_pod_name, created_objects, self.image_name)
                        volfunc.check_pod(value_pod, snap_pod_name, created_objects)
                        volfunc.check_file_inside_pod(value_pod, snap_pod_name, created_objects)

                        if "postsnap_volume_expansion_storage" in pvc_value:
                            volfunc.expand_and_check_pvc(restore_sc_name, restored_pvc_name, pvc_value, "postsnap_volume_expansion_storage",
                                                   snap_pod_name, value_pod, created_objects)

                        if "post_presnap_volume_expansion_storage" in pvc_value:
                            volfunc.expand_and_check_pvc(sc_name, pvc_name, pvc_value, "post_presnap_volume_expansion_storage",
                                                   pod_name, value_pod, created_objects)

                        if value_clone_passed is not None:
                            volfunc.clone_and_check_pvc(restore_sc_name, restore_sc, restored_pvc_name, snap_pod_name, value_pod, value_clone_passed, created_objects)

                        cleanup.delete_pod(snap_pod_name, created_objects)
                        cleanup.check_pod_deleted(snap_pod_name, created_objects)
                    vol_name = cleanup.delete_pvc(restored_pvc_name, created_objects)
                    cleanup.check_pvc_deleted(restored_pvc_name, vol_name, created_objects)

            cleanup.clean_with_created_objects(created_objects)

    def test_static(self, value_sc, test_restore, value_vs_class=None, number_of_snapshots=None, restore_sc=None, restore_pvc=None):
        if value_vs_class is None:
            value_vs_class = self.value_vs_class
        if number_of_snapshots is None:
            number_of_snapshots = self.number_of_snapshots
        number_of_restore = 1

        for pvc_value in self.value_pvc:

            created_objects = get_cleanup_dict()
            LOGGER.info("-"*100)
            sc_name = volfunc.get_random_name("sc")
            volfunc.create_storage_class(value_sc, sc_name, created_objects)
            volfunc.check_storage_class(sc_name)

            pvc_name = volfunc.get_random_name("pvc")
            volfunc.create_pvc(pvc_value, sc_name, pvc_name, created_objects)
            volfunc.check_pvc(pvc_value, pvc_name, created_objects)

            pod_name = volfunc.get_random_name("snap-start-pod")
            value_pod = {"mount_path": "/usr/share/nginx/html/scale", "read_only": "False"}
            volfunc.create_pod(value_pod, pvc_name, pod_name, created_objects, self.image_name)
            volfunc.check_pod(value_pod, pod_name, created_objects)
            volfunc.create_file_inside_pod(value_pod, pod_name, created_objects)

            snapshot_name = volfunc.get_random_name("snapshot")
            volume_name = volfunc.get_pv_for_pvc(pvc_name, created_objects)
            fileset_name = cleanup.get_filesetname_from_pv(volume_name, created_objects)

            FSUID = filesetfunc.get_FSUID()
            cluster_id = self.cluster_id
            vs_content_name = volfunc.get_random_name("vscontent")

            vs_name = volfunc.get_random_name("vs")
            for num in range(0, number_of_snapshots):
                filesetfunc.create_snapshot(snapshot_name+"-"+str(num), fileset_name, created_objects)
                if filesetfunc.check_snapshot_exists(snapshot_name+"-"+str(num), fileset_name):
                    LOGGER.info(f"snapshot {snapshot_name} exists for {fileset_name}")
                else:
                    LOGGER.error(f"snapshot {snapshot_name} does not exists for {fileset_name}")
                    cleanup.clean_with_created_objects(created_objects)
                    assert False

                snapshot_handle = cluster_id+';'+FSUID+';'+fileset_name+';'+snapshot_name+"-"+str(num)
                body_params = {"deletionPolicy": "Retain", "snapshotHandle": snapshot_handle}
                snapshotfunc.create_vs_content(vs_content_name+"-"+str(num), vs_name+"-"+str(num), body_params, created_objects)
                snapshotfunc.check_vs_content(vs_content_name+"-"+str(num))

                snapshotfunc.create_vs_from_content(vs_name+"-"+str(num), vs_content_name+"-"+str(num), created_objects)
                snapshotfunc.check_vs_detail_for_static(vs_name+"-"+str(num), created_objects)

            if not(filesetfunc.feature_available("snapshot")):
                pvc_value["reason"] = "Min required Spectrum Scale version for snapshot support with CSI is 5.1.1-0"

            if test_restore:
                if restore_sc is not None:
                    sc_name = "restore-" + sc_name
                    volfunc.create_storage_class(restore_sc, sc_name, created_objects)
                    volfunc.check_storage_class(sc_name)
                if restore_pvc is not None:
                    pvc_value = restore_pvc

                for num in range(0, number_of_restore):
                    restored_pvc_name = "restored-pvc"+vs_name[2:]+"-"+str(num)
                    snap_pod_name = "snap-end-pod"+vs_name[2:]
                    volfunc.create_pvc_from_snapshot(pvc_value, sc_name, restored_pvc_name, vs_name+"-"+str(num), created_objects)
                    val = volfunc.check_pvc(pvc_value, restored_pvc_name, created_objects)
                    if val is True:
                        volfunc.create_pod(value_pod, restored_pvc_name, snap_pod_name, created_objects, self.image_name)
                        volfunc.check_pod(value_pod, snap_pod_name, created_objects)
                        volfunc.check_file_inside_pod(value_pod, snap_pod_name, created_objects, fileset_name)
                        cleanup.delete_pod(snap_pod_name, created_objects)
                        cleanup.check_pod_deleted(snap_pod_name, created_objects)
                    vol_name = cleanup.delete_pvc(restored_pvc_name, created_objects)
                    cleanup.check_pvc_deleted(restored_pvc_name, vol_name, created_objects)

            cleanup.clean_with_created_objects(created_objects)


def get_test_data():
    filepath = "config/test.config"
    try:
        with open(filepath, "r") as f:
            data = yaml.full_load(f.read())
    except yaml.YAMLError as exc:
        print(f"Error in configuration file {filepath} :", exc)
        assert False

    if data['keepobjects'] == "True" or data['keepobjects'] == "true":
        data['keepobjects'] = True
    else:
        data['keepobjects'] = False

    if data.get('remote_username') is None:
        data['remote_username'] = {}
    if data.get('remote_password') is None:
        data['remote_password'] = {}
    if data.get('remote_cacert_path') is None:
        data['remote_cacert_path'] = {}

    return data


def read_driver_data(clusterconfig, namespace, operator_namespace, kubeconfig):

    data = get_test_data()

    data["namespace"] = namespace

    config.load_kube_config(config_file=kubeconfig)
    loadcr_yaml = csiobjectfunc.get_scaleoperatorobject_values(operator_namespace, data["csiscaleoperator_name"])

    if loadcr_yaml is False:
        try:
            with open(clusterconfig, "r") as f:
                loadcr_yaml = yaml.full_load(f.read())
        except yaml.YAMLError as exc:
            LOGGER.error(f"Error in parsing the cr file {clusterconfig} : {exc}")
            assert False

    for cluster in loadcr_yaml["spec"]["clusters"]:
        if "primary" in cluster.keys() and cluster["primary"]["primaryFs"] is not '':
            data["primaryFs"] = cluster["primary"]["primaryFs"]
            data["guiHost"] = cluster["restApi"][0]["guiHost"]
            if check_key(cluster["primary"], "primaryFset"):
                data["primaryFset"] = cluster["primary"]["primaryFset"]
            else:
                data["primaryFset"] = "spectrum-scale-csi-volume-store"
            data["id"] = cluster["id"]

    data["clusters"] = loadcr_yaml["spec"]["clusters"]
    if len(loadcr_yaml["spec"]["clusters"]) > 1:
        data["remote"] = True

    if check_key(loadcr_yaml["spec"], "pluginNodeSelector"):
        data["pluginNodeSelector"] = loadcr_yaml["spec"]["pluginNodeSelector"]
    else:
        data["pluginNodeSelector"] = []

    return data


def check_ns_exists(passed_kubeconfig_value, namespace_value):
    config.load_kube_config(config_file=passed_kubeconfig_value)
    read_namespace_api_instance = client.CoreV1Api()
    try:
        read_namespace_api_response = read_namespace_api_instance.read_namespace(
            name=namespace_value, pretty=True)
        LOGGER.debug(str(read_namespace_api_response))
        LOGGER.info("namespace exists checking for operator")
        return True
    except ApiException:
        LOGGER.info("namespace does not exists")
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


def check_key(dict1, key):
    if key in dict1.keys():
        return True
    return False


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


def read_operator_data(clusterconfig, namespace, kubeconfig=None):

    data = get_test_data()

    data["namespace"] = namespace

    if kubeconfig is not None:
        config.load_kube_config(config_file=kubeconfig)
        loadcr_yaml = csiobjectfunc.get_scaleoperatorobject_values(namespace, data["csiscaleoperator_name"])
    else:
        loadcr_yaml = False

    if loadcr_yaml is False:
        try:
            with open(clusterconfig, "r") as f:
                loadcr_yaml = yaml.full_load(f.read())
        except yaml.YAMLError as exc:
            LOGGER.error(f"Error in parsing the cr file {clusterconfig} : {exc}")
            assert False

    data["custom_object_body"] = copy.deepcopy(loadcr_yaml)
    data["custom_object_body"]["metadata"]["namespace"] = namespace
    data["remote_secret_names"] = []
    data["remote_cacert_names"] = []
    for cluster in loadcr_yaml["spec"]["clusters"]:
        if "primary" in cluster.keys() and cluster["primary"]["primaryFs"] is not '':
            data["primaryFs"] = cluster["primary"]["primaryFs"]
            data["guiHost"] = cluster["restApi"][0]["guiHost"]
            data["local_secret_name"] = cluster["secrets"]
            if check_key(cluster["primary"], "primaryFset"):
                data["primaryFset"] = cluster["primary"]["primaryFset"]
            else:
                data["primaryFset"] = "spectrum-scale-csi-volume-store"
            if check_key(cluster, "cacert"):
                data["local_cacert_name"] = cluster["cacert"]
        else:
            data["remote_secret_names"].append(cluster["secrets"])
            if check_key(cluster, "cacert"):
                data["remote_cacert_names"].append(cluster["cacert"])

    if check_key(loadcr_yaml["spec"], "attacherNodeSelector"):
        data["attacherNodeSelector"] = loadcr_yaml["spec"]["attacherNodeSelector"]
    else:
        data["attacherNodeSelector"] = []

    if check_key(loadcr_yaml["spec"], "provisionerNodeSelector"):
        data["provisionerNodeSelector"] = loadcr_yaml["spec"]["provisionerNodeSelector"]
    else:
        data["provisionerNodeSelector"] = []

    if check_key(loadcr_yaml["spec"], "pluginNodeSelector"):
        data["pluginNodeSelector"] = loadcr_yaml["spec"]["pluginNodeSelector"]
    else:
        data["pluginNodeSelector"] = []

    if check_key(loadcr_yaml["spec"], "resizerNodeSelector"):
        data["resizerNodeSelector"] = loadcr_yaml["spec"]["resizerNodeSelector"]
    else:
        data["resizerNodeSelector"] = []

    if check_key(loadcr_yaml["spec"], "snapshotterNodeSelector"):
        data["snapshotterNodeSelector"] = loadcr_yaml["spec"]["snapshotterNodeSelector"]
    else:
        data["snapshotterNodeSelector"] = []

    if check_key(data, "local_cacert_name"):
        if data["cacert_path"] == "":
            LOGGER.error("if using cacert , MUST include cacert path in test.config")
            assert False

    for remote_secret_name in data["remote_secret_names"]:
        if not(remote_secret_name in data["remote_username"].keys()):
            LOGGER.error(f"Need username for {remote_secret_name} secret in test.config")
            assert False
        if not(remote_secret_name in data["remote_password"].keys()):
            LOGGER.error(f"Need password for {remote_secret_name} secret in test.config")
            assert False

    for remote_cacert_name in data["remote_cacert_names"]:
        if not(remote_cacert_name in data["remote_cacert_path"].keys()):
            LOGGER.error(f"Need cacert path for {remote_cacert_name} in test.config")
            assert False

    return data

def get_remote_data(data_passed):
    remote_data = copy.deepcopy(data_passed)
    remote_data["remoteFs_remote_name"] = filesetfunc.get_remoteFs_remotename(copy.deepcopy(remote_data))
    if remote_data["remoteFs_remote_name"] is None:
        LOGGER.error("Unable to get remoteFs , name on remote cluster")
        assert False

    remote_data["primaryFs"] = remote_data["remoteFs_remote_name"]
    remote_data["id"] = remote_data["remoteid"]
    remote_data["port"] = remote_data["remote_port"]
    for cluster in remote_data["clusters"]:
        if cluster["id"] == remote_data["remoteid"]:
            remote_data["guiHost"] = cluster["restApi"][0]["guiHost"]
            remote_sec_name = cluster["secrets"]
            remote_data["username"] = remote_data["remote_username"][remote_sec_name]
            remote_data["password"] = remote_data["remote_password"][remote_sec_name]

    remote_data["volDirBasePath"] = remote_data["r_volDirBasePath"]
    remote_data["parentFileset"] = remote_data["r_parentFileset"]
    remote_data["gid_name"] = remote_data["r_gid_name"]
    remote_data["uid_name"] = remote_data["r_uid_name"]
    remote_data["gid_number"] = remote_data["r_gid_number"]
    remote_data["uid_number"] = remote_data["r_uid_number"]
    remote_data["inodeLimit"] = remote_data["r_inodeLimit"]
    # for get_mount_point function
    remote_data["type_remote"] = {"username": data_passed["username"],
                                  "password": data_passed["password"],
                                  "port": data_passed["port"],
                                  "guiHost": data_passed["guiHost"]}

    return remote_data

def get_cmd_values(request):
    kubeconfig_value = request.config.option.kubeconfig
    if kubeconfig_value is None:
        if os.path.isfile('config/kubeconfig'):
            kubeconfig_value = 'config/kubeconfig'
        else:
            kubeconfig_value = '~/.kube/config'

    clusterconfig_value = request.config.option.clusterconfig
    if clusterconfig_value is None:
        if os.path.isfile('config/csiscaleoperators.csi.ibm.com_cr.yaml'):
            clusterconfig_value = 'config/csiscaleoperators.csi.ibm.com_cr.yaml'
        else:
            clusterconfig_value = '../../operator/config/samples/csiscaleoperators.csi.ibm.com_cr.yaml'

    test_namespace = request.config.option.testnamespace
    if test_namespace is None:
        test_namespace = 'ibm-spectrum-scale-csi-driver'

    operator_namespace = request.config.option.operatornamespace
    if operator_namespace is None:
        operator_namespace = 'ibm-spectrum-scale-csi-driver'

    runslow_val = request.config.option.runslow

    operator_file = request.config.option.operatoryaml
    if operator_file is None:
        operator_file = '../../generated/installer/ibm-spectrum-scale-csi-operator-dev.yaml'

    return kubeconfig_value, clusterconfig_value, operator_namespace, test_namespace, runslow_val, operator_file


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
        "ds": []
    }
    return created_object
    LOGGER.info("Demo Message")
