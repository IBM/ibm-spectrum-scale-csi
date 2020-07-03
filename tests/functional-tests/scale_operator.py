import copy
import multiprocessing
import logging
import yaml
from kubernetes import client, config
from kubernetes.client.rest import ApiException
import utils.scale_operator_function as scale_function
import utils.scale_operator_object_function as ob
import utils.driver as d
import utils.snapshot as snapshot
from conftest import input_params
from utils.fileset_functions import get_FSUID, create_dir, delete_dir, get_mount_point
LOGGER = logging.getLogger()


class Scaleoperator:
    def __init__(self, kubeconfig_value, namespace_value):

        self.kubeconfig = kubeconfig_value
        scale_function.set_global_namespace_value(namespace_value)

    def create(self):

        config.load_kube_config(config_file=self.kubeconfig)

        if not(scale_function.check_namespace_exists()):
            scale_function.create_namespace()

        if not(scale_function.check_deployment_exists()):
            scale_function.create_deployment()

        if not(scale_function.check_cluster_role_exists("ibm-spectrum-scale-csi-operator")):
            scale_function.create_cluster_role()

        if not(scale_function.check_service_account_exists("ibm-spectrum-scale-csi-operator")):
            scale_function.create_service_account()

        if not(scale_function.check_cluster_role_binding_exists("ibm-spectrum-scale-csi-operator")):
            scale_function.create_cluster_role_binding()

        if not(scale_function.check_crd_exists()):
            scale_function.create_crd()

    def delete(self):

        config.load_kube_config(config_file=self.kubeconfig)

        if ob.check_scaleoperatorobject_is_deployed():   # for edge cases if custom object is not deleted
            ob.delete_custom_object()
            ob.check_scaleoperatorobject_is_deleted()

        if scale_function.check_crd_exists():
            scale_function.delete_crd()
        scale_function.check_crd_deleted()

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

        if scale_function.check_namespace_exists():
            scale_function.delete_namespace()
        scale_function.check_namespace_deleted()

    def check(self):

        config.load_kube_config(config_file=self.kubeconfig)
        scale_function.check_namespace_exists()
        scale_function.check_deployment_exists()
        scale_function.check_cluster_role_exists("ibm-spectrum-scale-csi-operator")
        scale_function.check_service_account_exists("ibm-spectrum-scale-csi-operator")
        scale_function.check_cluster_role_binding_exists("ibm-spectrum-scale-csi-operator")
        scale_function.check_crd_exists()


class Scaleoperatorobject:

    def __init__(self, test_dict, kubeconfig_value):

        LOGGER.info("scale operator class object is created")
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
            remote_secret_data = {"username": self.temp["remote-username"][remote_secret_name],
                                  "password": self.temp["remote-password"][remote_secret_name]}
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

        ob.check_scaleoperatorobject_is_deployed()

        ob.check_scaleoperatorobject_statefulsets_state(
            "ibm-spectrum-scale-csi-attacher")

        ob.check_scaleoperatorobject_statefulsets_state(
            "ibm-spectrum-scale-csi-provisioner")

        val, self.desired_number_scheduled = ob.check_scaleoperatorobject_daemonsets_state()

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
        for lable_val in label:
            label_selector += str(lable_val["key"]) + \
                "="+str(lable_val["value"])+","
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

    def __init__(self, kubeconfig_value, value_pvc, value_pod, config_file, test_ns, keep_object):
        self.value_pvc = value_pvc
        self.value_pod = value_pod
        self.config_file = config_file
        self.test_ns = test_ns
        self.keep_objects = keep_object
        self.kubeconfig = kubeconfig_value
        d.set_test_namespace_value(self.test_ns)
        d.set_keep_objects(self.keep_objects)

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

    def test_dynamic(self, value_sc):
        LOGGER.info(
            f"Testing Dynamic Provisioning with following PVC parameters {str(self.value_pvc)}")
        sc_name = d.get_random_name("sc")
        config.load_kube_config(config_file=self.kubeconfig)
        d.create_storage_class(value_sc, self.config_file, sc_name)
        d.check_storage_class(sc_name)
        for num in range(0, len(self.value_pvc)):
            value_pvc_pass = copy.deepcopy(self.value_pvc[num])
            if (check_key(value_sc, "reason")):
                if not(check_key(value_pvc_pass, "reason")):
                    value_pvc_pass["reason"] = value_sc["reason"]
            LOGGER.info(100*"=")
            pvc_name = d.get_random_name("pvc")
            d.create_pvc(value_pvc_pass, sc_name, pvc_name)
            val = d.check_pvc(value_pvc_pass, sc_name, pvc_name)
            if val is True:
                for num2 in range(0, len(self.value_pod)):
                    LOGGER.info(100*"-")
                    pod_name = d.get_random_name("pod")
                    d.create_pod(self.value_pod[num2], pvc_name, pod_name)
                    d.check_pod(self.value_pod[num2], sc_name, pvc_name, pod_name)
                    d.delete_pod(pod_name)
                    d.check_pod_deleted(pod_name)
                    if value_pvc_pass["access_modes"] == "ReadWriteOnce" and self.keep_objects is True:
                        if num2 < (len(self.value_pod)-1):
                            pvc_name = d.get_random_name("pvc")
                            d.create_pvc(value_pvc_pass, sc_name, pvc_name)
                            val = d.check_pvc(value_pvc_pass, sc_name, pvc_name)
                            if val is not True:
                                break
                LOGGER.info(100*"-")
            d.delete_pvc(pvc_name)
            d.check_pvc_deleted(pvc_name)
        LOGGER.info(100*"=")
        if d.check_storage_class(sc_name):
            d.delete_storage_class(sc_name)
        d.check_storage_class_deleted(sc_name)

    def test_static(self, pv_value, pvc_value, sc_value=False, wrong=None, root_volume=False):

        config.load_kube_config(config_file=self.kubeconfig)
        sc_name = ""
        if sc_value is not False:
            sc_name = d.get_random_name("sc")
            d.create_storage_class(sc_value, self.config_file, sc_name)
            d.check_storage_class(sc_name)
        FSUID = get_FSUID(self.config_file)
        cluster_id = self.config_file["id"]
        if wrong is not None:
            if wrong["id_wrong"] is True:
                cluster_id = int(cluster_id)+1
                cluster_id = str(cluster_id)
            if wrong["FSUID_wrong"] is True:
                FSUID = "AAAA"

        mount_point = get_mount_point(self.config_file)
        if root_volume is False:
            dir_name = d.get_random_name("dir")
            create_dir(self.config_file, dir_name)
            pv_value["volumeHandle"] = cluster_id+";"+FSUID + \
                ";path="+mount_point+"/"+dir_name
        elif root_volume is True:
            pv_value["volumeHandle"] = cluster_id+";"+FSUID + \
                ";path="+mount_point
            dir_name = "nodiravailable"

        if pvc_value == "Default":
            pvc_value = copy.deepcopy(self.value_pvc)

        num_final = len(pvc_value)
        for num in range(0, num_final):
            pv_name = d.get_random_name("pv")
            d.create_pv(pv_value, pv_name, sc_name)
            d.check_pv(pv_name)

            value_pvc_pass = copy.deepcopy(pvc_value[num])
            if (check_key(pv_value, "reason")):
                if not(check_key(value_pvc_pass, "reason")):
                    value_pvc_pass["reason"] = pv_value["reason"]
            LOGGER.info(100*"=")
            pvc_name = d.get_random_name("pvc")
            d.create_pvc(value_pvc_pass, sc_name, pvc_name, self.config_file, pv_name)
            val = d.check_pvc(value_pvc_pass, sc_name, pvc_name, dir_name, pv_name)
            if val is True:
                for num2 in range(0, len(self.value_pod)):
                    LOGGER.info(100*"-")
                    pod_name = d.get_random_name("pod")
                    d.create_pod(self.value_pod[num2], pvc_name, pod_name)
                    d.check_pod(self.value_pod[num2], sc_name, pvc_name, pod_name, dir_name, pv_name)
                    d.delete_pod(pod_name)
                    d.check_pod_deleted(pod_name)
                    if value_pvc_pass["access_modes"] == "ReadWriteOnce" and self.keep_objects is True:
                        break
                LOGGER.info(100*"-")
            d.delete_pvc(pvc_name)
            d.check_pvc_deleted(pvc_name)
            d.delete_pv(pv_name)
            d.check_pv_deleted(pv_name)
        LOGGER.info(100*"=")
        delete_dir(self.config_file, dir_name)

        if d.check_storage_class(sc_name):
            d.delete_storage_class(sc_name)
        d.check_storage_class_deleted(sc_name)

    def one_pvc_two_pod(self, value_sc):
        sc_name = d.get_random_name("sc")
        config.load_kube_config(config_file=self.kubeconfig)
        d.create_storage_class(value_sc, self.config_file, sc_name)
        d.check_storage_class(sc_name)
        value_pvc_pass = copy.deepcopy(self.value_pvc[0])
        pvc_name = d.get_random_name("pvc")
        d.create_pvc(value_pvc_pass, sc_name, pvc_name)
        val = d.check_pvc(value_pvc_pass, sc_name, pvc_name)
        if val is True:
            pod_name_1 = d.get_random_name("pod")
            d.create_pod(self.value_pod[0], pvc_name, pod_name_1)
            d.check_pod(self.value_pod[0], sc_name, pvc_name, pod_name_1)
            pod_name_2 = d.get_random_name("pod")
            d.create_pod(self.value_pod[0], pvc_name, pod_name_2)
            d.check_pod(self.value_pod[0], sc_name, pvc_name, pod_name_2)
            d.delete_pod(pod_name_1)
            d.check_pod_deleted(pod_name_1)
            d.delete_pod(pod_name_2)
            d.check_pod_deleted(pod_name_2)
        d.delete_pvc(pvc_name)
        d.check_pvc_deleted(pvc_name)
        if d.check_storage_class(sc_name):
            d.delete_storage_class(sc_name)
        d.check_storage_class_deleted(sc_name)

    def parallel_pvc_process(self, sc_name):
        value_pvc_pass = copy.deepcopy(self.value_pvc[0])
        LOGGER.info(100*"-")
        value_pvc_pass["parallel"] = "True"
        pvc_name = d.get_random_name("pvc")
        d.create_pvc(value_pvc_pass, sc_name, pvc_name)
        val = d.check_pvc(value_pvc_pass, sc_name, pvc_name)
        if val is True:
            pod_name = d.get_random_name("pod")
            d.create_pod(self.value_pod[0], pvc_name, pod_name)
            d.check_pod(self.value_pod[0], sc_name, pvc_name, pod_name)
            d.delete_pod(pod_name)
            d.check_pod_deleted(pod_name)
        d.delete_pvc(pvc_name)
        d.check_pvc_deleted(pvc_name)

    def parallel_pvc(self, value_sc):
        sc_name = d.get_random_name("sc")
        config.load_kube_config(config_file=self.kubeconfig)
        d.create_storage_class(value_sc, self.config_file, sc_name)
        d.check_storage_class(sc_name)
        p = []
        number_of_pvc = self.config_file["number_of_parallel_pvc"]
        for num in range(0, number_of_pvc):
            p.append(multiprocessing.Process(
                target=self.parallel_pvc_process, args=[sc_name]))
            p[num].start()
        for num in range(0, number_of_pvc-1):
            p[num].join()

        LOGGER.info(100*"-")
        if d.check_storage_class(sc_name):
            d.delete_storage_class(sc_name)
        d.check_storage_class_deleted(sc_name)


class Snapshot():
    def __init__(self, kubeconfig, test_namespace, keep_objects, data, value_pvc, value_vs_class, number_of_snapshots):
        config.load_kube_config(config_file=kubeconfig)
        self.config_file = data
        self.value_pvc = value_pvc
        self.value_vs_class = value_vs_class
        self.number_of_snapshots = number_of_snapshots
        d.set_keep_objects(keep_objects)
        d.set_test_namespace_value(test_namespace)
        snapshot.set_test_namespace_value(test_namespace)
        snapshot.set_keep_objects(keep_objects)

    def test_dynamic(self, value_sc, value_vs_class=None, number_of_snapshots=None):
        if value_vs_class is None:
            value_vs_class = self.value_vs_class
        if number_of_snapshots is None:
            number_of_snapshots = self.number_of_snapshots

        for pvc_value in self.value_pvc:
            LOGGER.info("-"*100)
            sc_name = d.get_random_name("sc")
            d.create_storage_class(value_sc, self.config_file, sc_name)
            d.check_storage_class(sc_name)

            pvc_name = d.get_random_name("pvc")
            d.create_pvc(pvc_value, sc_name, pvc_name)
            d.check_pvc(pvc_value, sc_name, pvc_name)

            vs_class_name = d.get_random_name("vsclass")
            snapshot.create_vs_class(vs_class_name, value_vs_class)
            snapshot.check_vs_class(vs_class_name)

            vs_names = []
            for _ in range(0, number_of_snapshots):
                vs_name = d.get_random_name("vs")
                vs_names.append(vs_name)
                snapshot.create_vs(vs_name, vs_class_name, pvc_name)
                snapshot.check_vs_detail(vs_name, pvc_name, self.config_file, vs_class_name, sc_name)

            for vs_name in vs_names:
                snapshot.delete_vs(vs_name)
                snapshot.check_vs_deleted(vs_name)
            snapshot.delete_vs_class(vs_class_name)
            snapshot.check_vs_class_deleted(vs_class_name)
            d.delete_pvc(pvc_name)
            d.check_pvc_deleted(pvc_name)
            if d.check_storage_class(sc_name):
                d.delete_storage_class(sc_name)
            d.check_storage_class_deleted(sc_name)


def read_driver_data(clusterconfig, namespace):

    data = copy.deepcopy(input_params)
    data["namespace"] = namespace

    try:
        with open(clusterconfig, "r") as f:
            loadcr_yaml = yaml.full_load(f.read())
    except yaml.YAMLError as exc:
        LOGGER.error(f"Error in parsing the cr file {clusterconfig} : {exc}")
        assert False

    data["scaleHostpath"] = loadcr_yaml["spec"]["scaleHostpath"]
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

    return data


def check_ns_exists(passed_kubeconfig_value, namespace_value):
    config.load_kube_config(config_file=passed_kubeconfig_value)
    read_namespace_api_instance = client.CoreV1Api()
    try:
        read_namespace_api_response = read_namespace_api_instance.read_namespace(
            name=namespace_value, pretty=True)
        LOGGER.debug(str(read_namespace_api_response))
        LOGGER.info("namespace exists checking daemon sets")
        return True
    except ApiException:
        LOGGER.info("namespace does not exists")
        return False


def check_ds_exists(passed_kubeconfig_value, namespace_value):
    config.load_kube_config(config_file=passed_kubeconfig_value)
    read_daemonsets_api_instance = client.AppsV1Api()
    try:
        read_daemonsets_api_response = read_daemonsets_api_instance.read_namespaced_daemon_set(
            name="ibm-spectrum-scale-csi", namespace=namespace_value, pretty=True)
        current_number_scheduled = read_daemonsets_api_response.status.current_number_scheduled
        desired_number_scheduled = read_daemonsets_api_response.status.desired_number_scheduled
        number_available = read_daemonsets_api_response.status.number_available
        if number_available == current_number_scheduled == desired_number_scheduled:
            LOGGER.info("Daemon sets are up")
        else:
            LOGGER.info("Daemon sets are not available")
            LOGGER.error("Requirements are not satisfied")
            assert False
    except ApiException:
        LOGGER.error("daemonset does not exists")
        LOGGER.error("Requirements are not satisfied")
        assert False


def check_key(dict1, key):
    if key in dict1.keys():
        return True
    return False


def check_nodes_available(label, lable_name):
    api_instance = client.CoreV1Api()
    label_selector = ""
    for lable_val in label:
        label_selector += str(lable_val["key"])+"="+str(lable_val["value"])+","
    label_selector = label_selector[0:-1]
    try:
        api_response_2 = api_instance.list_node(
            pretty=True, label_selector=label_selector)
        if len(api_response_2.items) == 0:
            LOGGER.error(f"0 nodes have provided labels with {lable_name}")
            LOGGER.error("please check labels")
            assert False
    except ApiException as e:
        LOGGER.error(f"Exception when calling CoreV1Api->list_node: {e}")
        assert False


def read_operator_data(clusterconfig, namespace):

    data = copy.deepcopy(input_params)

    data["namespace"] = namespace

    try:
        with open(clusterconfig, "r") as f:
            loadcr_yaml = yaml.full_load(f.read())
    except yaml.YAMLError as exc:
        LOGGER.error(f"Error in parsing the cr file {clusterconfig} : {exc}")
        assert False

    data["scaleHostpath"] = loadcr_yaml["spec"]["scaleHostpath"]
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
            LOGGER.error("if using cacert , MUST include cacert path in conftest.py")
            assert False

    for remote_secret_name in data["remote_secret_names"]:
        if not(remote_secret_name in data["remote-username"].keys()):
            LOGGER.error(f"Need username for {remote_secret_name} secret in conftest")
            assert False
        if not(remote_secret_name in data["remote-password"].keys()):
            LOGGER.error(f"Need password for {remote_secret_name} secret in conftest")
            assert False

    for remote_cacert_name in data["remote_cacert_names"]:
        if not(remote_cacert_name in data["remote_cacert_path"].keys()):
            LOGGER.error(f"Need cacert path for {remote_cacert_name} in conftest")
            assert False

    return data
