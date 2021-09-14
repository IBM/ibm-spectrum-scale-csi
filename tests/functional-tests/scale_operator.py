import copy
import logging
import os.path
import yaml
from kubernetes import client, config
from kubernetes.client.rest import ApiException
import utils.scale_operator_function as scale_function
import utils.scale_operator_object_function as ob
import utils.driver as d
import utils.snapshot as snapshot
import utils.fileset_functions as ff
import utils.cleanup_functions as cleanup
LOGGER = logging.getLogger()


class Scaleoperator:
    def __init__(self, kubeconfig_value, namespace_value):

        self.kubeconfig = kubeconfig_value
        scale_function.set_global_namespace_value(namespace_value)
        ob.set_namespace_value(namespace_value)

        crd_body = self.get_operator_body()
        crd_full_version = crd_body["CustomResourceDefinition"]["apiVersion"].split("/")
        self.crd_version = crd_full_version[1]

    def create(self):

        config.load_kube_config(config_file=self.kubeconfig)

        body = self.get_operator_body()
        if not(scale_function.check_namespace_exists()):
            scale_function.create_namespace()

        if not(scale_function.check_deployment_exists()):
            scale_function.create_deployment(body['Deployment'])

        if not(scale_function.check_cluster_role_exists("ibm-spectrum-scale-csi-operator")):
            scale_function.create_cluster_role(body['ClusterRole'])

        if not(scale_function.check_service_account_exists("ibm-spectrum-scale-csi-operator")):
            scale_function.create_service_account(body['ServiceAccount'])

        if not(scale_function.check_cluster_role_binding_exists("ibm-spectrum-scale-csi-operator")):
            scale_function.create_cluster_role_binding(body['ClusterRoleBinding'])

        if not(scale_function.check_crd_exists(self.crd_version)):
            scale_function.create_crd(body['CustomResourceDefinition'])

    def delete(self, condition=False):

        config.load_kube_config(config_file=self.kubeconfig)
        if ob.check_scaleoperatorobject_is_deployed():   # for edge cases if custom object is not deleted
            ob.delete_custom_object()
            ob.check_scaleoperatorobject_is_deleted()

        if scale_function.check_crd_exists(self.crd_version):
            scale_function.delete_crd(self.crd_version)
        scale_function.check_crd_deleted(self.crd_version)

        if scale_function.check_service_account_exists("ibm-spectrum-scale-csi-operator"):
            scale_function.delete_service_account("ibm-spectrum-scale-csi-operator")
        scale_function.check_service_account_deleted("ibm-spectrum-scale-csi-operator")

        if scale_function.check_cluster_role_binding_exists("ibm-spectrum-scale-csi-operator"):
            scale_function.delete_cluster_role_binding("ibm-spectrum-scale-csi-operator")
        scale_function.check_cluster_role_binding_deleted("ibm-spectrum-scale-csi-operator")

        if scale_function.check_cluster_role_exists("ibm-spectrum-scale-csi-operator"):
            scale_function.delete_cluster_role("ibm-spectrum-scale-csi-operator")
        scale_function.check_cluster_role_deleted("ibm-spectrum-scale-csi-operator")

        if scale_function.check_deployment_exists():
            scale_function.delete_deployment()
        scale_function.check_deployment_deleted()

        if scale_function.check_namespace_exists() and (condition is False):
            scale_function.delete_namespace()
            scale_function.check_namespace_deleted()

    def check(self):

        config.load_kube_config(config_file=self.kubeconfig)
        scale_function.check_namespace_exists()
        scale_function.check_deployment_exists()
        scale_function.check_cluster_role_exists("ibm-spectrum-scale-csi-operator")
        scale_function.check_service_account_exists("ibm-spectrum-scale-csi-operator")
        scale_function.check_cluster_role_binding_exists("ibm-spectrum-scale-csi-operator")
        scale_function.check_crd_exists(self.crd_version)

    def get_operator_body(self):

        path = "../../generated/installer/ibm-spectrum-scale-csi-operator-dev.yaml"
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

        ob.set_namespace_value(self.namespace)

    def create(self):

        config.load_kube_config(config_file=self.kubeconfig)
        LOGGER.info(str(self.temp["custom_object_body"]["spec"]))

        if not(ob.check_secret_exists(self.secret_name)):
            ob.create_secret(self.secret_data, self.secret_name)
        else:
            ob.delete_secret(self.secret_name)
            ob.check_secret_is_deleted(self.secret_name)
            ob.create_secret(self.secret_data, self.secret_name)

        for remote_secret_name in self.temp["remote_secret_names"]:
            remote_secret_data = {"username": self.temp["remote_username"][remote_secret_name],
                                  "password": self.temp["remote_password"][remote_secret_name]}
            if not(ob.check_secret_exists(remote_secret_name)):
                ob.create_secret(remote_secret_data, remote_secret_name)
            else:
                ob.delete_secret(remote_secret_name)
                ob.check_secret_is_deleted(remote_secret_name)
                ob.create_secret(remote_secret_data, remote_secret_name)

        if check_key(self.temp, "local_cacert_name"):
            cacert_name = self.temp["local_cacert_name"]
            if not(check_key(self.temp, "make_cacert_wrong")):
                self.temp["make_cacert_wrong"] = False

            if not(ob.check_configmap_exists(cacert_name)):
                ob.create_configmap(
                    self.temp["cacert_path"], self.temp["make_cacert_wrong"], cacert_name)
            else:
                ob.delete_configmap(cacert_name)
                ob.check_configmap_is_deleted(cacert_name)
                ob.create_configmap(
                    self.temp["cacert_path"], self.temp["make_cacert_wrong"], cacert_name)

        for remote_cacert_name in self.temp["remote_cacert_names"]:

            if not(check_key(self.temp, "make_remote_cacert_wrong")):
                self.temp["make_remote_cacert_wrong"] = False

            if not(ob.check_configmap_exists(remote_cacert_name)):
                ob.create_configmap(
                    self.temp["remote_cacert_path"][remote_cacert_name], self.temp["make_remote_cacert_wrong"], remote_cacert_name)
            else:
                ob.delete_configmap(remote_cacert_name)
                ob.check_configmap_is_deleted(remote_cacert_name)
                ob.create_configmap(
                    self.temp["remote_cacert_path"][remote_cacert_name], self.temp["make_remote_cacert_wrong"], remote_cacert_name)

        if not(ob.check_scaleoperatorobject_is_deployed()):

            ob.create_custom_object(self.temp["custom_object_body"], self.stateful_set_not_created)
        else:
            ob.delete_custom_object()
            ob.check_scaleoperatorobject_is_deleted()
            ob.create_custom_object(self.temp["custom_object_body"], self.stateful_set_not_created)

    def delete(self):

        config.load_kube_config(config_file=self.kubeconfig)

        if ob.check_scaleoperatorobject_is_deployed():
            ob.delete_custom_object()
        ob.check_scaleoperatorobject_is_deleted()

        if ob.check_secret_exists(self.secret_name):
            ob.delete_secret(self.secret_name)
        ob.check_secret_is_deleted(self.secret_name)

        for remote_secret_name in self.temp["remote_secret_names"]:
            if ob.check_secret_exists(remote_secret_name):
                ob.delete_secret(remote_secret_name)
            ob.check_secret_is_deleted(remote_secret_name)
        if check_key(self.temp, "local_cacert_name"):
            if ob.check_configmap_exists(self.temp["local_cacert_name"]):
                ob.delete_configmap(self.temp["local_cacert_name"])
            ob.check_configmap_is_deleted(self.temp["local_cacert_name"])

        for remote_cacert_name in self.temp["remote_cacert_names"]:
            if ob.check_configmap_exists(remote_cacert_name):
                ob.delete_configmap(remote_cacert_name)
            ob.check_configmap_is_deleted(remote_cacert_name)

    def check(self):
        config.load_kube_config(config_file=self.kubeconfig)

        is_deployed = ob.check_scaleoperatorobject_is_deployed()
        if(is_deployed is False):
            return False

        ob.check_scaleoperatorobject_statefulsets_state(
            "ibm-spectrum-scale-csi-attacher")

        ob.check_scaleoperatorobject_statefulsets_state(
            "ibm-spectrum-scale-csi-provisioner")

        val, self.desired_number_scheduled = ob.check_scaleoperatorobject_daemonsets_state()

        # ob.check_pod_running("ibm-spectrum-scale-csi-snapshotter-0")

        return val

    def get_driver_ds_pod_name(self):

        try:
            pod_list_api_instance = client.CoreV1Api()
            pod_list_api_response = pod_list_api_instance.list_namespaced_pod(
                namespace=self.namespace, pretty=True, field_selector="spec.serviceAccountName=ibm-spectrum-scale-csi-node")
            demonset_pod_name = pod_list_api_response.items[0].metadata.name
            LOGGER.debug(str(pod_list_api_response))
            return demonset_pod_name
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->list_namespaced_pod: {e}")
            assert False

    def get_scaleplugin_labelled_nodes(self, label):
        api_instance = client.CoreV1Api()
        label_selector = ""
        for label_val in label:
            label_selector += str(label_val["key"]) + \
                "="+str(label_val["value"])+","
        label_selector = label_selector[0:-1]
        try:
            api_response_2 = api_instance.list_node(
                pretty=True, label_selector=label_selector)
            LOGGER.info(f"{label_selector} labelled nodes are \
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
        d.set_test_namespace_value(self.test_ns)
        d.set_test_nodeselector_value(plugin_nodeselector_labels)
        cleanup.set_keep_objects(self.keep_objects)
        cleanup.set_test_namespace_value(self.test_ns)

    def create_test_ns(self, kubeconfig):
        config.load_kube_config(config_file=kubeconfig)
        scale_function.set_global_namespace_value(self.test_ns)
        if not(scale_function.check_namespace_exists()):
            scale_function.create_namespace()

    def delete_test_ns(self, kubeconfig):
        config.load_kube_config(config_file=kubeconfig)
        scale_function.set_global_namespace_value(self.test_ns)
        if scale_function.check_namespace_exists():
            scale_function.delete_namespace()
        scale_function.check_namespace_deleted()

    def test_dynamic(self, value_sc, value_pvc_passed=None, value_pod_passed=None):
        created_objects = get_cleanup_dict()
        if value_pvc_passed is None:
            value_pvc_passed = copy.deepcopy(self.value_pvc)
        if value_pod_passed is None:
            value_pod_passed = copy.deepcopy(self.value_pod)

        if "permissions" in value_sc.keys() and not(ff.feature_available("permissions")):
            LOGGER.warning("Min required Spectrum Scale version for permissions in storageclass support with CSI is 5.1.1-2")
            LOGGER.warning("Skipping Testcase")
            return

        LOGGER.info(
            f"Testing Dynamic Provisioning with following PVC parameters {str(value_pvc_passed)}")
        sc_name = d.get_random_name("sc")
        config.load_kube_config(config_file=self.kubeconfig)
        d.create_storage_class(value_sc, sc_name, created_objects)
        d.check_storage_class(sc_name)
        for num, _ in enumerate(value_pvc_passed):
            value_pvc_pass = copy.deepcopy(value_pvc_passed[num])
            if (check_key(value_sc, "reason")):
                if not(check_key(value_pvc_pass, "reason")):
                    value_pvc_pass["reason"] = value_sc["reason"]
            LOGGER.info(100*"=")
            pvc_name = d.get_random_name("pvc")
            d.create_pvc(value_pvc_pass, sc_name, pvc_name, created_objects)
            val = d.check_pvc(value_pvc_pass, pvc_name, created_objects)
            if val is True:
                if "permissions" in value_sc.keys():
                    d.check_permissions_for_pvc(pvc_name, value_sc["permissions"], created_objects)

                for num2, _ in enumerate(value_pod_passed):
                    LOGGER.info(100*"-")
                    pod_name = d.get_random_name("pod")
                    if value_sc.keys() >= {"permissions", "gid", "uid"}:
                        value_pod_passed[num2]["gid"] = value_sc["gid"]
                        value_pod_passed[num2]["uid"] = value_sc["uid"]
                    d.create_pod(value_pod_passed[num2], pvc_name, pod_name, created_objects, self.image_name)
                    d.check_pod(value_pod_passed[num2], pod_name, created_objects)
                    cleanup.delete_pod(pod_name, created_objects)
                    cleanup.check_pod_deleted(pod_name, created_objects)
                    if ((value_pvc_pass["access_modes"] == "ReadWriteOnce") and (self.keep_objects is True) and (num2 < (len(value_pod_passed)-1))):
                        pvc_name = d.get_random_name("pvc")
                        d.create_pvc(value_pvc_pass, sc_name, pvc_name, created_objects)
                        val = d.check_pvc(value_pvc_pass, pvc_name, created_objects)
                        if val is not True:
                            break
                LOGGER.info(100*"-")
            vol_name = cleanup.delete_pvc(pvc_name, created_objects)
            cleanup.check_pvc_deleted(pvc_name, vol_name, created_objects)
        LOGGER.info(100*"=")
        cleanup.clean_with_created_objects(created_objects)


    def test_static(self, pv_value, pvc_value, sc_value=False, wrong=None, root_volume=False):

        config.load_kube_config(config_file=self.kubeconfig)
        created_objects = get_cleanup_dict()
        sc_name = ""
        if sc_value is not False:
            sc_name = d.get_random_name("sc")
            d.create_storage_class(sc_value,  sc_name, created_objects)
            d.check_storage_class(sc_name)
        FSUID = ff.get_FSUID()
        cluster_id = self.cluster_id
        if wrong is not None:
            if wrong["id_wrong"] is True:
                cluster_id = int(cluster_id)+1
                cluster_id = str(cluster_id)
            if wrong["FSUID_wrong"] is True:
                FSUID = "AAAA"

        mount_point = ff.get_mount_point()
        if root_volume is False:
            dir_name = d.get_random_name("dir")
            ff.create_dir(dir_name)
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
            pv_name = d.get_random_name("pv")
            d.create_pv(pv_value, pv_name, created_objects, sc_name)
            d.check_pv(pv_name)

            value_pvc_pass = copy.deepcopy(pvc_value[num])
            if (check_key(pv_value, "reason")):
                if not(check_key(value_pvc_pass, "reason")):
                    value_pvc_pass["reason"] = pv_value["reason"]
            LOGGER.info(100*"=")
            pvc_name = d.get_random_name("pvc")
            d.create_pvc(value_pvc_pass, sc_name, pvc_name, created_objects, pv_name)
            val = d.check_pvc(value_pvc_pass, pvc_name, created_objects, pv_name)
            if val is True:
                for num2 in range(0, len(self.value_pod)):
                    LOGGER.info(100*"-")
                    pod_name = d.get_random_name("pod")
                    d.create_pod(self.value_pod[num2], pvc_name, pod_name, created_objects, self.image_name)
                    d.check_pod(self.value_pod[num2], pod_name, created_objects)
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
        sc_name = d.get_random_name("sc")
        config.load_kube_config(config_file=self.kubeconfig)
        d.create_storage_class(value_sc, sc_name, created_objects)
        d.check_storage_class(sc_name)
        pvc_name = d.get_random_name("pvc")
        d.create_pvc(value_pvc_pass, sc_name, pvc_name, created_objects)
        val = d.check_pvc(value_pvc_pass, pvc_name, created_objects)
        if val is True:
            ds_name = d.get_random_name("ds")
            d.create_ds(value_ds_pass, ds_name, pvc_name, created_objects)
            d.check_ds(ds_name, value_ds_pass, created_objects)
        cleanup.clean_with_created_objects(created_objects)

    def sequential_pvc(self, value_sc, num_of_pvc):
        created_objects = get_cleanup_dict()
        sc_name = d.get_random_name("sc")
        config.load_kube_config(config_file=self.kubeconfig)
        d.create_storage_class(value_sc, sc_name, created_objects)
        d.check_storage_class(sc_name)
        pvc_names = []
        number_of_pvc = num_of_pvc
        common_pvc_name = d.get_random_name("pvc")
        for num in range(0, number_of_pvc):
            pvc_names.append(common_pvc_name+"-"+str(num))
        value_pvc_pass = copy.deepcopy(self.value_pvc[0])
        LOGGER.info(100*"-")
        value_pvc_pass["parallel"] = "True"

        for pvc_name in pvc_names:
            LOGGER.info(100*"-")
            d.create_pvc(value_pvc_pass, sc_name, pvc_name, created_objects)

        for pvc_name in pvc_names:
            LOGGER.info(100*"-")
            d.check_pvc(value_pvc_pass, pvc_name, created_objects)

        pod_names = []

        for pvc_name in pvc_names:
            LOGGER.info(100*"-")
            pod_name = d.get_random_name("pod")
            pod_names.append(pod_name)
            d.create_pod(self.value_pod[0], pvc_name, pod_name, created_objects, self.image_name)
            d.check_pod(self.value_pod[0], pod_name, created_objects)

        cleanup.clean_with_created_objects(created_objects)


class Snapshot():
    def __init__(self, kubeconfig, test_namespace, keep_objects, value_pvc, value_vs_class, number_of_snapshots, image_name, cluster_id, plugin_nodeselector_labels):
        config.load_kube_config(config_file=kubeconfig)
        self.value_pvc = value_pvc
        self.value_vs_class = value_vs_class
        self.number_of_snapshots = number_of_snapshots
        self.image_name = image_name
        self.cluster_id = cluster_id
        d.set_test_namespace_value(test_namespace)
        d.set_test_nodeselector_value(plugin_nodeselector_labels)
        snapshot.set_test_namespace_value(test_namespace)
        cleanup.set_keep_objects(keep_objects)
        cleanup.set_test_namespace_value(test_namespace)


    def test_dynamic(self, value_sc, test_restore, value_vs_class=None, number_of_snapshots=None, reason=None, restore_sc=None, restore_pvc=None, value_pod=None):
        if value_vs_class is None:
            value_vs_class = self.value_vs_class
        if number_of_snapshots is None:
            number_of_snapshots = self.number_of_snapshots
        number_of_restore = 1

        if "permissions" in value_sc.keys() and not(ff.feature_available("permissions")):
            LOGGER.warning("Min required Spectrum Scale version for permissions in storageclass support with CSI is 5.1.1-2")
            LOGGER.warning("Skipping Testcase")
            return

        for pvc_value in self.value_pvc:

            created_objects = get_cleanup_dict()
            LOGGER.info("-"*100)
            sc_name = d.get_random_name("sc")
            d.create_storage_class(value_sc, sc_name, created_objects)
            d.check_storage_class(sc_name)

            pvc_name = d.get_random_name("pvc")
            d.create_pvc(pvc_value, sc_name, pvc_name, created_objects)
            val = d.check_pvc(pvc_value, pvc_name, created_objects)

            if val is True and "permissions" in value_sc.keys():
                d.check_permissions_for_pvc(pvc_name, value_sc["permissions"], created_objects)

            pod_name = d.get_random_name("snap-start-pod")
            if value_pod is None:
                value_pod = {"mount_path": "/usr/share/nginx/html/scale", "read_only": "False"} 

            if value_sc.keys() >= {"permissions", "gid", "uid"}:
                value_pod["gid"] = value_sc["gid"]
                value_pod[num2]["uid"] = value_sc["uid"]
            d.create_pod(value_pod, pvc_name, pod_name, created_objects, self.image_name)
            d.check_pod(value_pod, pod_name, created_objects)
            d.create_file_inside_pod(value_pod, pod_name, created_objects)

            vs_class_name = d.get_random_name("vsclass")
            snapshot.create_vs_class(vs_class_name, value_vs_class, created_objects)
            snapshot.check_vs_class(vs_class_name)

            if not(ff.feature_available("snapshot")):
                if reason is None:
                    reason = "Min required Spectrum Scale version for snapshot support with CSI is 5.1.1-0"
                test_restore = False

            vs_name = d.get_random_name("vs")
            for num in range(0, number_of_snapshots):
                snapshot.create_vs(vs_name+"-"+str(num), vs_class_name, pvc_name, created_objects)
                snapshot.check_vs_detail(vs_name+"-"+str(num), pvc_name, value_vs_class, reason, created_objects)

            if test_restore:
                if restore_sc is not None:
                    sc_name = "restore-" + sc_name
                    d.create_storage_class(restore_sc, sc_name, created_objects)
                    d.check_storage_class(sc_name)
                if restore_pvc is not None:
                    pvc_value = restore_pvc

                for num in range(0, number_of_restore):
                    restored_pvc_name = "restored-pvc"+vs_name[2:]+"-"+str(num)
                    snap_pod_name = "snap-end-pod"+vs_name[2:]
                    d.create_pvc_from_snapshot(pvc_value, sc_name, restored_pvc_name, vs_name+"-"+str(num), created_objects)
                    val = d.check_pvc(pvc_value, restored_pvc_name, created_objects)
                    if val is True and "permissions" in value_sc.keys():
                        d.check_permissions_for_pvc(pvc_name, value_sc["permissions"], created_objects)

                    if val is True:
                        if "permissions" in value_sc.keys():
                            run_as_group = value_sc["gid"]
                            run_as_user = value_sc["uid"]
                            d.create_pod(value_pod, restored_pvc_name, snap_pod_name, created_objects, self.image_name, run_as_user, run_as_group)
                        else:
                            d.create_pod(value_pod, restored_pvc_name, snap_pod_name, created_objects, self.image_name)
                        d.check_pod(value_pod, snap_pod_name, created_objects)
                        d.check_file_inside_pod(value_pod, snap_pod_name, created_objects)
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
            sc_name = d.get_random_name("sc")
            d.create_storage_class(value_sc, sc_name, created_objects)
            d.check_storage_class(sc_name)

            pvc_name = d.get_random_name("pvc")
            d.create_pvc(pvc_value, sc_name, pvc_name, created_objects)
            d.check_pvc(pvc_value, pvc_name, created_objects)

            pod_name = d.get_random_name("snap-start-pod")
            value_pod = {"mount_path": "/usr/share/nginx/html/scale", "read_only": "False"}
            d.create_pod(value_pod, pvc_name, pod_name, created_objects, self.image_name)
            d.check_pod(value_pod, pod_name, created_objects)
            d.create_file_inside_pod(value_pod, pod_name, created_objects)

            snapshot_name = d.get_random_name("snapshot")
            volume_name = snapshot.get_pv_name(pvc_name, created_objects)

            FSUID = ff.get_FSUID()
            cluster_id = self.cluster_id
            vs_content_name = d.get_random_name("vscontent")

            vs_name = d.get_random_name("vs")
            for num in range(0, number_of_snapshots):
                ff.create_snapshot(snapshot_name+"-"+str(num), volume_name, created_objects)
                if ff.check_snapshot(snapshot_name+"-"+str(num), volume_name):
                    LOGGER.info(f"snapshot {snapshot_name} exists for {volume_name}")
                else:
                    LOGGER.error(f"snapshot {snapshot_name} does not exists for {volume_name}")
                    cleanup.clean_with_created_objects(created_objects)
                    assert False

                snapshot_handle = cluster_id+';'+FSUID+';'+volume_name+';'+snapshot_name+"-"+str(num)
                body_params = {"deletionPolicy": "Retain", "snapshotHandle": snapshot_handle}
                snapshot.create_vs_content(vs_content_name+"-"+str(num), vs_name+"-"+str(num), body_params, created_objects)
                snapshot.check_vs_content(vs_content_name+"-"+str(num))

                snapshot.create_vs_from_content(vs_name+"-"+str(num), vs_content_name+"-"+str(num), created_objects)
                snapshot.check_vs_detail_for_static(vs_name+"-"+str(num), created_objects)

            if not(ff.feature_available("snapshot")):
                pvc_value["reason"] = "Min required Spectrum Scale version for snapshot support with CSI is 5.1.1-0"

            if test_restore:
                if restore_sc is not None:
                    sc_name = "restore-" + sc_name
                    d.create_storage_class(restore_sc, sc_name, created_objects)
                    d.check_storage_class(sc_name)
                if restore_pvc is not None:
                    pvc_value = restore_pvc

                for num in range(0, number_of_restore):
                    restored_pvc_name = "restored-pvc"+vs_name[2:]+"-"+str(num)
                    snap_pod_name = "snap-end-pod"+vs_name[2:]
                    d.create_pvc_from_snapshot(pvc_value, sc_name, restored_pvc_name, vs_name+"-"+str(num), created_objects)
                    val = d.check_pvc(pvc_value, restored_pvc_name, created_objects)
                    if val is True:
                        d.create_pod(value_pod, restored_pvc_name, snap_pod_name, created_objects, self.image_name)
                        d.check_pod(value_pod, snap_pod_name, created_objects)
                        d.check_file_inside_pod(value_pod, snap_pod_name, created_objects, volume_name)
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


def read_driver_data(clusterconfig, namespace):

    data = get_test_data()

    data["namespace"] = namespace

    try:
        with open(clusterconfig, "r") as f:
            loadcr_yaml = yaml.full_load(f.read())
    except yaml.YAMLError as exc:
        LOGGER.error(f"Error in parsing the cr file {clusterconfig} : {exc}")
        assert False

    for cluster in loadcr_yaml["spec"]["clusters"]:
        if "primary" in cluster.keys():
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


def read_operator_data(clusterconfig, namespace):

    data = get_test_data()

    data["namespace"] = namespace

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
        if "primary" in cluster.keys():
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
    return kubeconfig_value, clusterconfig_value, operator_namespace, test_namespace, runslow_val


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
        "pv": [],
        "dir": [],
        "ds": []
    }
    return created_object
