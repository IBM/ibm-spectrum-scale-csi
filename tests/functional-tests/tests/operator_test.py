import time
import re
import random
import logging
import pytest
from kubernetes import client
from kubernetes.client.rest import ApiException
import ibm_spectrum_scale_csi.base_class as baseclass
import ibm_spectrum_scale_csi.common_utils.input_data_functions as inputfunc
LOGGER = logging.getLogger()
pytestmark = pytest.mark.csioperator

@pytest.fixture(scope='session')
def _values(request):

    global kubeconfig_value, clusterconfig_value, namespace_value
    kubeconfig_value, clusterconfig_value, operator_namespace, test_namespace, _, operator_yaml = inputfunc.get_cmd_values(request)
    namespace_value = operator_namespace
    condition = baseclass.kubeobjectfunc.check_ns_exists(kubeconfig_value, namespace_value)
    operator = baseclass.Scaleoperator(kubeconfig_value, namespace_value, operator_yaml)
    read_file = inputfunc.read_operator_data(clusterconfig_value, namespace_value)
    baseclass.filesetfunc.cred_check(read_file)
    fileset_exist = baseclass.filesetfunc.fileset_exists(read_file)
    operator.create()
    operator.check()
    baseclass.kubeobjectfunc.check_nodes_available(
        read_file["pluginNodeSelector"], "pluginNodeSelector")
    baseclass.kubeobjectfunc.check_nodes_available(
        read_file["provisionerNodeSelector"], "provisionerNodeSelector")
    baseclass.kubeobjectfunc.check_nodes_available(
        read_file["attacherNodeSelector"], "attacherNodeSelector")

    yield
    operator.delete(condition)
    if(not(fileset_exist) and baseclass.filesetfunc.fileset_exists(read_file)):
        baseclass.filesetfunc.delete_fileset(read_file)


def test_get_version(_values):
    test = inputfunc.read_operator_data(clusterconfig_value, namespace_value)
    baseclass.filesetfunc.get_scale_version(test)
    baseclass.kubeobjectfunc.get_kubernetes_version(kubeconfig_value)
    baseclass.kubeobjectfunc.get_operator_image()


def test_operator_deploy(_values):

    LOGGER.info("test_operator_deploy")
    LOGGER.info("Every input is correct should run without any error")
    test = inputfunc.read_operator_data(clusterconfig_value, namespace_value)
    operator_object = baseclass.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()
    if operator_object.check() is True:
        LOGGER.info("Operator custom object is deployed successfully")
    else:
        daemonset_pod_name = operator_object.get_driver_ds_pod_name()
        get_logs_api_instance = client.CoreV1Api()
        try:
            get_logs_api_response = get_logs_api_instance.read_namespaced_pod_log(
                name=daemonset_pod_name, namespace=namespace_value, container="ibm-spectrum-scale-csi")
            LOGGER.error(str(get_logs_api_response))
            LOGGER.error(
                "operator custom object should be deployed but it is not deployed hence asserting")
            operator_object.delete()
            assert False
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            assert False
    operator_object.delete()


def test_wrong_cluster_id(_values):
    LOGGER.info("test_wrong_cluster_id : cluster ID is wrong")
    test = inputfunc.read_operator_data(clusterconfig_value, namespace_value)
    wrong_id = str(random.randint(0, 999999999999999999))

    for cluster in test["custom_object_body"]["spec"]["clusters"]:
        if "primary" in cluster.keys():
            cluster["id"] = wrong_id

    operator_object = baseclass.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()
    if operator_object.check() is True:
        LOGGER.error(
            "Operator custom object is deployed successfully not expected")
        assert False
    else:
        daemonset_pod_name = operator_object.get_driver_ds_pod_name()
        get_logs_api_instance = client.CoreV1Api()
        try:
            LOGGER.info(f"Checking for failure reason match in {daemonset_pod_name} pod logs")
            get_logs_api_response = get_logs_api_instance.read_namespaced_pod_log(
                name=daemonset_pod_name, namespace=namespace_value, container="ibm-spectrum-scale-csi")
            LOGGER.debug(str(get_logs_api_response))
            search_result = re.search(
                "Cluster ID doesnt match the cluster", get_logs_api_response)
            LOGGER.debug(search_result)
            assert search_result is not None
            LOGGER.info("'Cluster ID doesnt match the cluster' failure reason matched")
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            assert False
    operator_object.delete()


def test_wrong_primaryFS(_values):
    LOGGER.info("test_wrong_primaryFS : primaryFS is wrong")
    test = inputfunc.read_operator_data(clusterconfig_value, namespace_value)
    wrong_primaryFs = inputfunc.randomStringDigits()

    for cluster in test["custom_object_body"]["spec"]["clusters"]:
        if "primary" in cluster.keys():
            cluster["primary"]["primaryFs"] = wrong_primaryFs
    test["primaryFs"] = wrong_primaryFs
    operator_object = baseclass.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()
    if operator_object.check() is True:
        LOGGER.error(
            "Operator custom object is deployed successfully not expected")
        assert False
    else:
        daemonset_pod_name = operator_object.get_driver_ds_pod_name()
        get_logs_api_instance = client.CoreV1Api()
        try:
            LOGGER.info(f"Checking for failure reason match in {daemonset_pod_name} pod logs")
            get_logs_api_response = get_logs_api_instance.read_namespaced_pod_log(
                name=daemonset_pod_name, namespace=namespace_value, container="ibm-spectrum-scale-csi")
            LOGGER.debug(str(get_logs_api_response))
            search_result = re.search(
                "Unable to get filesystem", get_logs_api_response)
            LOGGER.debug(search_result)
            assert search_result is not None
            LOGGER.info("'Unable to get filesystem' failure reason matched")
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            assert False
    operator_object.delete()


def test_wrong_guihost(_values):
    LOGGER.info("test_wrong_guihost : gui host is wrong")
    test = inputfunc.read_operator_data(clusterconfig_value, namespace_value)
    wrong_guiHost = inputfunc.randomStringDigits()
    test["guiHost"] = wrong_guiHost
    for cluster in test["custom_object_body"]["spec"]["clusters"]:
        if "primary" in cluster.keys():
            cluster["restApi"][0]["guiHost"] = wrong_guiHost

    operator_object = baseclass.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()
    if operator_object.check() is True:
        LOGGER.error(
            "Operator custom object is deployed successfully not expected")
        assert False
    else:
        daemonset_pod_name = operator_object.get_driver_ds_pod_name()
        get_logs_api_instance = client.CoreV1Api()
        try:
            LOGGER.info(f"Checking for failure reason match in {daemonset_pod_name} pod logs")
            get_logs_api_response = get_logs_api_instance.read_namespaced_pod_log(
                name=daemonset_pod_name, namespace=namespace_value, container="ibm-spectrum-scale-csi")
            LOGGER.debug(str(get_logs_api_response))
            search_result1 = re.search(
                "connection refused", get_logs_api_response)
            LOGGER.debug(search_result1)
            search_result2 = re.search("no such host", get_logs_api_response)
            LOGGER.debug(search_result2)
            assert (search_result1 is not None or search_result2 is not None)
            LOGGER.info("'connection refused' or 'no such host'  failure reason matched")
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            assert False
    operator_object.delete()


def test_wrong_gui_username(_values):
    LOGGER.info("test_wrong_gui_username : gui username is wrong")
    test = inputfunc.read_operator_data(clusterconfig_value, namespace_value)
    test["username"] = inputfunc.randomStringDigits()
    operator_object = baseclass.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()
    if operator_object.check() is True:
        LOGGER.error(
            "Operator custom object is deployed successfully not expected")
        assert False
    else:
        daemonset_pod_name = operator_object.get_driver_ds_pod_name()
        get_logs_api_instance = client.CoreV1Api()
        try:
            LOGGER.info(f"Checking for failure reason match in {daemonset_pod_name} pod logs")
            get_logs_api_response = get_logs_api_instance.read_namespaced_pod_log(
                name=daemonset_pod_name, namespace=namespace_value, container="ibm-spectrum-scale-csi")
            LOGGER.debug(str(get_logs_api_response))
            x = re.search("401 Unauthorized", get_logs_api_response)
            assert x is not None
            LOGGER.info("'401 Unauthorized' failure reason matched")
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            assert False
    operator_object.delete()


def test_wrong_gui_password(_values):
    LOGGER.info("test_wrong_gui_password : gui password is wrong")
    test = inputfunc.read_operator_data(clusterconfig_value, namespace_value)
    test["password"] = inputfunc.randomStringDigits()
    operator_object = baseclass.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()
    operator_object.check()
    LOGGER.info("Checkig if failure reason matches")
    daemonset_pod_name = operator_object.get_driver_ds_pod_name()
    LOGGER.info(f"Checking for failure reason match in {daemonset_pod_name} pod logs")
    get_logs_api_instance = client.CoreV1Api()
    count = 0
    while count < 24:
        try:
            get_logs_api_response = get_logs_api_instance.read_namespaced_pod_log(
                name=daemonset_pod_name, namespace=namespace_value, container="ibm-spectrum-scale-csi")
            LOGGER.debug(str(get_logs_api_response))
            search_result = re.search("401 Unauthorized", get_logs_api_response)
            if search_result is None:
                time.sleep(5)
                count += 1
            else:
                LOGGER.debug(search_result)
                LOGGER.info("'401 Unauthorized' failure reason matched")
                operator_object.delete()
                return
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            assert False

    operator_object.delete()
    LOGGER.error(str(get_logs_api_response))
    LOGGER.error("Asserting as reason of failure does not match")
    assert search_result is not None


def test_wrong_secret_object_name(_values):
    LOGGER.info("test_wrong_secret_object_name : secret object name is wrong")
    test = inputfunc.read_operator_data(clusterconfig_value, namespace_value)
    secret_name_wrong = inputfunc.randomString()

    for cluster in test["custom_object_body"]["spec"]["clusters"]:
        if "primary" in cluster.keys():
            cluster["secrets"] = secret_name_wrong

    test["stateful_set_not_created"] = True
    operator_object = baseclass.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()
    operator_object.delete()


def test_random_gpfs_primaryFset_name(_values):
    LOGGER.info("test_random_gpfs_primaryFset_name : gpfs primary Fset name is wrong")
    test = inputfunc.read_operator_data(clusterconfig_value, namespace_value)
    random_primaryFset = inputfunc.randomStringDigits()
    test["primaryFset"] = random_primaryFset
    for cluster in test["custom_object_body"]["spec"]["clusters"]:
        if "primary" in cluster.keys():
            cluster["primary"]["primaryFset"] = random_primaryFset

    operator_object = baseclass.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()
    if operator_object.check() is True:
        LOGGER.info("Operator custom object is deployed successfully")
    else:
        get_logs_api_instance = client.CoreV1Api()
        try:
            daemonset_pod_name = operator_object.get_driver_ds_pod_name()
            LOGGER.info(f"Checking for failure reason match in {daemonset_pod_name} pod logs")
            get_logs_api_response = get_logs_api_instance.read_namespaced_pod_log(
                name=daemonset_pod_name, namespace=namespace_value, container="ibm-spectrum-scale-csi")
            LOGGER.debug(str(get_logs_api_response))
            LOGGER.error(
                "operator custom object should be deployed but it is not deployed hence asserting")
            operator_object.delete()
            if(baseclass.filesetfunc.fileset_exists(test)):
                baseclass.filesetfunc.delete_fileset(test)
            assert False
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            operator_object.delete()
            if(baseclass.filesetfunc.fileset_exists(test)):
                baseclass.filesetfunc.delete_fileset(test)
            assert False
    if(baseclass.filesetfunc.fileset_exists(test)):
        baseclass.filesetfunc.delete_fileset(test)
    operator_object.delete()


def test_secureSslMode(_values):
    LOGGER.info("test_secureSslMode")
    LOGGER.info("secureSslMode is True while cacert is not available")
    test = inputfunc.read_operator_data(clusterconfig_value, namespace_value)

    for cluster in test["custom_object_body"]["spec"]["clusters"]:
        if "primary" in cluster.keys():
            cluster["secureSslMode"] = True
            if "cacert" in cluster.keys():
                cluster.pop("cacert")

    operator_object = baseclass.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()
    if operator_object.check() is True:
        LOGGER.error(
            "Operator custom object is deployed successfully not expected")
        assert False
    else:
        daemonset_pod_name = operator_object.get_driver_ds_pod_name()
        get_logs_api_instance = client.CoreV1Api()
        try:
            LOGGER.info(f"Checking for failure reason match in {daemonset_pod_name} pod logs")
            get_logs_api_response = get_logs_api_instance.read_namespaced_pod_log(
                name=daemonset_pod_name, namespace=namespace_value, container="ibm-spectrum-scale-csi")
            LOGGER.debug(str(get_logs_api_response))
            search_result = re.search(
                "CA certificate not specified in secure SSL mode for cluster", str(get_logs_api_response))
            LOGGER.debug(search_result)
            if(search_result is None):
                operator_object.delete()
                LOGGER.error(str(get_logs_api_response))
                LOGGER.error("Reason of failure does not match")
            assert search_result is not None
            LOGGER.info("'CA certificate not specified in secure SSL mode for cluster' failure reason matched")
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            assert False
    if(baseclass.filesetfunc.fileset_exists(test)):
        baseclass.filesetfunc.delete_fileset(test)
    operator_object.delete()


"""
Removing this testcase as scaleHostpath is no longer needed
def test_wrong_gpfs_filesystem_mount_point(_values):
    LOGGER.info("test_wrong_gpfs_filesystem_mount_point")
    LOGGER.info("gpfs filesystem mount point is wrong")
    test = inputfunc.read_operator_data(clusterconfig_value, namespace_value)
    wrong_scaleHostpath = inputfunc.randomStringDigits()
    test["custom_object_body"]["spec"]["scaleHostpath"] = wrong_scaleHostpath
    operator_object = baseclass.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()

    if operator_object.check() is True:
        LOGGER.error(
            "Operator custom object is deployed successfully not expected")
        assert False
    else:
        get_logs_api_instance = client.CoreV1Api()
        daemonset_pod_name = operator_object.get_driver_ds_pod_name()
        try:
            LOGGER.info(f"Checking for failure reason match in {daemonset_pod_name} pod events")
            field = "involvedObject.name="+daemonset_pod_name
            api_response = get_logs_api_instance.list_namespaced_event(
                namespace=namespace_value, pretty="True", field_selector=field)
            LOGGER.debug(str(api_response))
            search_result = re.search(
                'MountVolume.SetUp failed for volume', str(api_response))
            LOGGER.debug(search_result)
            assert search_result is not None
            LOGGER.info("'MountVolume.SetUp failed for volume' failure reason matched")
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            assert False

    operator_object.delete()
"""


def test_unlinked_primaryFset(_values):
    LOGGER.info("test_unlinked_primaryFset")
    LOGGER.info("unlinked primaryFset expected : object created successfully")
    test = inputfunc.read_operator_data(clusterconfig_value, namespace_value)
    if(not(baseclass.filesetfunc.fileset_exists(test))):
        baseclass.filesetfunc.create_fileset(test)
    baseclass.filesetfunc.unlink_fileset(test)
    operator_object = baseclass.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()
    if operator_object.check() is True:
        LOGGER.info("Operator custom object is deployed successfully")
    else:
        get_logs_api_instance = client.CoreV1Api()
        try:
            daemonset_pod_name = operator_object.get_driver_ds_pod_name()
            LOGGER.info(f"Checking for failure reason match in {daemonset_pod_name} pod logs")
            get_logs_api_response = get_logs_api_instance.read_namespaced_pod_log(
                name=daemonset_pod_name, namespace=namespace_value, container="ibm-spectrum-scale-csi")
            LOGGER.error(str(get_logs_api_response))
            LOGGER.error(
                "operator custom object should be deployed but it is not deployed hence asserting")
            operator_object.delete()
            assert False
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            assert False

    operator_object.delete()


def test_existing_primaryFset(_values):
    LOGGER.info("test_existing_primaryFset")
    LOGGER.info(
        "linked existing primaryFset expected : object created successfully")
    test = inputfunc.read_operator_data(clusterconfig_value, namespace_value)
    if(not(baseclass.filesetfunc.fileset_exists(test))):
        baseclass.filesetfunc.create_fileset(test)
    operator_object = baseclass.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()
    if operator_object.check() is True:
        LOGGER.info("Operator custom object is deployed successfully")
    else:
        get_logs_api_instance = client.CoreV1Api()
        try:
            daemonset_pod_name = operator_object.get_driver_ds_pod_name()
            get_logs_api_response = get_logs_api_instance.read_namespaced_pod_log(
                name=daemonset_pod_name, namespace=namespace_value, container="ibm-spectrum-scale-csi")
            LOGGER.error(str(get_logs_api_response))
            LOGGER.error(
                "operator custom object should be deployed but it is not deployed hence asserting")
            operator_object.delete()
            assert False
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            assert False

    operator_object.delete()


def test_unmounted_primaryFS(_values):
    LOGGER.info("test_unmounted_primaryFS")
    LOGGER.info(
        "primaryFS is unmounted and expected : custom object should give error")
    test = inputfunc.read_operator_data(clusterconfig_value, namespace_value)
    baseclass.filesetfunc.unmount_fs(test)
    operator_object = baseclass.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()
    if operator_object.check() is True:
        LOGGER.error(
            "Operator custom object is deployed successfully, it is not expected")
        operator_object.delete()
        baseclass.filesetfunc.mount_fs(test)
        assert False
    else:
        get_logs_api_instance = client.CoreV1Api()
        try:
            daemonset_pod_name = operator_object.get_driver_ds_pod_name()
            LOGGER.info(f"Checking for failure reason match in {daemonset_pod_name} pod logs")
            get_logs_api_response = get_logs_api_instance.read_namespaced_pod_log(
                name=daemonset_pod_name, namespace=namespace_value, container="ibm-spectrum-scale-csi")
            LOGGER.debug(str(get_logs_api_response))
            search_result = re.search(
                'not mounted on GUI node Primary cluster', str(get_logs_api_response))
            if search_result is None:
                LOGGER.error(str(get_logs_api_response))
            LOGGER.debug(search_result)
            operator_object.delete()
            baseclass.filesetfunc.mount_fs(test)
            assert search_result is not None
            LOGGER.info("'not mounted on GUI node Primary cluster' failure reason matched")
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            baseclass.filesetfunc.mount_fs(test)
            assert False
    operator_object.delete()
    baseclass.filesetfunc.mount_fs(test)


def test_non_deafult_attacher(_values):
    LOGGER.info("test_non_deafult_attacher")
    LOGGER.info("attacher image name is changed")
    test = inputfunc.read_operator_data(clusterconfig_value, namespace_value)
    deployment_attacher_image = "quay.io/k8scsi/csi-attacher:v1.2.1"
    test["custom_object_body"]["spec"]["attacher"] = deployment_attacher_image
    operator_object = baseclass.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()
    if operator_object.check() is True:
        LOGGER.info("Operator custom object is deployed successfully")
        baseclass.kubeobjectfunc.check_pod_image(test["csiscaleoperator_name"]+"-attacher-0", deployment_attacher_image)
    else:
        get_logs_api_instance = client.CoreV1Api()
        try:
            daemonset_pod_name = operator_object.get_driver_ds_pod_name()
            LOGGER.info(f"Checking for failure reason match in {daemonset_pod_name} pod events")
            field = "involvedObject.name="+daemonset_pod_name
            api_response = get_logs_api_instance.list_namespaced_event(
                namespace=namespace_value, pretty="True", field_selector=field)
            LOGGER.debug(str(api_response))
            LOGGER.error(
                "operator custom object should be deployed but it is not deployed hence asserting")
            assert False
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            assert False

    operator_object.delete()


def test_non_deafult_provisioner(_values):
    LOGGER.info("test_non_deafult_provisioner")
    LOGGER.info("provisioner image name is changed")
    test = inputfunc.read_operator_data(clusterconfig_value, namespace_value)
    deployment_provisioner_image = "quay.io/k8scsi/csi-provisioner:v1.6.0"
    test["custom_object_body"]["spec"]["provisioner"] = deployment_provisioner_image
    operator_object = baseclass.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()
    if operator_object.check() is True:
        LOGGER.info("Operator custom object is deployed successfully")
        baseclass.kubeobjectfunc.check_pod_image(test["csiscaleoperator_name"]+"-provisioner-0", deployment_provisioner_image)       
    else:
        get_logs_api_instance = client.CoreV1Api()
        try:
            daemonset_pod_name = operator_object.get_driver_ds_pod_name()
            LOGGER.info(f"Checking for failure reason match in {daemonset_pod_name} pod events")
            field = "involvedObject.name="+daemonset_pod_name
            api_response = get_logs_api_instance.list_namespaced_event(
                namespace=namespace_value, pretty="True", field_selector=field)
            LOGGER.debug(str(api_response))
            LOGGER.error(
                "operator custom object should be deployed but it is not deployed hence asserting")
            assert False
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            assert False

    operator_object.delete()


def test_correct_cacert(_values):
    LOGGER.info("test_secureSslMode with correct cacert file")
    LOGGER.info("correct cacert file is given")
    test = inputfunc.read_operator_data(clusterconfig_value, namespace_value)

    if not("local_cacert_name" in test):
        test["local_cacert_name"] = "test-cacert-configmap"

    for cluster in test["custom_object_body"]["spec"]["clusters"]:
        if "primary" in cluster.keys():
            cluster["secureSslMode"] = True
            if not("cacert" in cluster.keys()):
                cluster["cacert"] = "test-cacert-configmap"

    operator_object = baseclass.Scaleoperatorobject(test, kubeconfig_value)
    if test["cacert_path"] == "":
        LOGGER.info("skipping the test as cacert file path is not given in test.config")
        pytest.skip("path of cacert file is not given")

    operator_object.create()
    if operator_object.check() is True:
        LOGGER.info("Operator custom object is deployed successfully")
    else:
        daemonset_pod_name = operator_object.get_driver_ds_pod_name()
        get_logs_api_instance = client.CoreV1Api()
        try:
            LOGGER.info(f"Checking for failure reason match in {daemonset_pod_name} pod logs")
            get_logs_api_response = get_logs_api_instance.read_namespaced_pod_log(
                name=daemonset_pod_name, namespace=namespace_value, container="ibm-spectrum-scale-csi")
            LOGGER.info(str(get_logs_api_response))
            operator_object.delete()
            LOGGER.error(
                "operator custom object should be deployed but it is not deployed hence asserting")
            assert False
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            assert False

    operator_object.delete()


def test_cacert_with_secureSslMode_false(_values):
    LOGGER.info("test_cacert_with_secureSslMode_false")
    LOGGER.info("secureSslMode is false with correct cacert file")
    test = inputfunc.read_operator_data(clusterconfig_value, namespace_value)

    if not("local_cacert_name" in test):
        test["local_cacert_name"] = "test-cacert-configmap"

    for cluster in test["custom_object_body"]["spec"]["clusters"]:
        if "primary" in cluster.keys():
            cluster["secureSslMode"] = False
            if not("cacert" in cluster.keys()):
                cluster["cacert"] = "test-cacert-configmap"

    operator_object = baseclass.Scaleoperatorobject(test, kubeconfig_value)
    if test["cacert_path"] == "":
        LOGGER.info("skipping the test as cacert file path is not given in test.config")
        pytest.skip("path of cacert file is not given")

    operator_object.create()
    if operator_object.check() is True:
        LOGGER.info("Operator custom object is deployed successfully")
    else:
        daemonset_pod_name = operator_object.get_driver_ds_pod_name()
        get_logs_api_instance = client.CoreV1Api()
        try:
            LOGGER.info(f"Checking for failure reason match in {daemonset_pod_name} pod logs")
            get_logs_api_response = get_logs_api_instance.read_namespaced_pod_log(
                name=daemonset_pod_name, namespace=namespace_value, container="ibm-spectrum-scale-csi")
            LOGGER.info(str(get_logs_api_response))
            operator_object.delete()
            LOGGER.error(
                "operator custom object should be deployed but it is not deployed hence asserting")
            assert False
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            assert False

    operator_object.delete()


def test_wrong_cacert(_values):
    LOGGER.info("secureSslMode true with wrong cacert file")
    LOGGER.info("test_wrong_cacert")
    test = inputfunc.read_operator_data(clusterconfig_value, namespace_value)

    if not("local_cacert_name" in test):
        test["local_cacert_name"] = "test-cacert-configmap"

    for cluster in test["custom_object_body"]["spec"]["clusters"]:
        if "primary" in cluster.keys():
            cluster["secureSslMode"] = True
            if not("cacert" in cluster.keys()):
                cluster["cacert"] = "test-cacert-configmap"

    test["make_cacert_wrong"] = True
    operator_object = baseclass.Scaleoperatorobject(test, kubeconfig_value)
    if test["cacert_path"] == "":
        LOGGER.info("skipping the test as cacert file path is not given in test.config")
        pytest.skip("path of cacert file is not given")

    operator_object.create()
    if operator_object.check() is True:
        LOGGER.error(
            "Operator custom object is deployed successfully not expected")
        operator_object.delete()
        assert False
    else:
        daemonset_pod_name = operator_object.get_driver_ds_pod_name()
        LOGGER.info(f"Checking for failure reason match in {daemonset_pod_name} pod logs")
        get_logs_api_instance = client.CoreV1Api()
        count = 0
        while count < 24:
            try:
                get_logs_api_response = get_logs_api_instance.read_namespaced_pod_log(
                    name=daemonset_pod_name, namespace=namespace_value, container="ibm-spectrum-scale-csi")
                LOGGER.debug(str(get_logs_api_response))
                search_result = re.search(
                    "Error in plugin initialization: Parsing CA cert", get_logs_api_response)
                if search_result is None:
                    time.sleep(5)
                else:
                    LOGGER.debug(search_result)
                    break
                if count > 23:
                    operator_object.delete()
                    assert search_result is not None
            except ApiException as e:
                LOGGER.error(
                    f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
                assert False

    operator_object.delete()


def test_nodeMapping(_values):
    LOGGER.info("test_nodeMapping")
    LOGGER.info("nodeMapping is added to the cr file")
    test = inputfunc.read_operator_data(clusterconfig_value, namespace_value)
    LOGGER.debug(test)
    operator_object = baseclass.Scaleoperatorobject(test, kubeconfig_value)

    if "nodeMapping" not in test['custom_object_body']['spec']:
        LOGGER.info("skipping the test as nodeMapping is not given in cr.yaml file")
        pytest.skip("nodeMapping not in cr file")

    operator_object.create()
    if operator_object.check() is True:
        LOGGER.info("Operator custom object is deployed successfully")
    else:
        get_logs_api_instance = client.CoreV1Api()
        try:
            daemonset_pod_name = operator_object.get_driver_ds_pod_name()
            LOGGER.info(f"Checking for failure reason match in {daemonset_pod_name} pod events")
            field = "involvedObject.name="+daemonset_pod_name
            api_response = get_logs_api_instance.list_namespaced_event(
                namespace=namespace_value, pretty="True", field_selector=field)
            LOGGER.debug(str(api_response))
            LOGGER.error(
                "operator custom object should be deployed but it is not deployed hence asserting")
            assert False
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            assert False

    operator_object.delete()


def test_attacherNodeSelector(_values):
    LOGGER.info("test_attacherNodeSelector")
    LOGGER.info("attacherNodeSelector is added to the cr file")
    test = inputfunc.read_operator_data(clusterconfig_value, namespace_value)
    operator_object = baseclass.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()
    if operator_object.check() is True:
        LOGGER.info("Operator custom object is deployed successfully")
        desired_daemonset_node, labeled_nodes = operator_object.get_scaleplugin_labeled_nodes(
            test["attacherNodeSelector"])
        if desired_daemonset_node == labeled_nodes:
            LOGGER.info("labeled nodes are equal to desired daemonset nodes")
        else:
            LOGGER.error(
                "labeled nodes are not equal to desired daemonset nodes")
            assert False
    else:
        get_logs_api_instance = client.CoreV1Api()
        try:
            daemonset_pod_name = operator_object.get_driver_ds_pod_name()
            LOGGER.info(f"Checking for failure reason match in {daemonset_pod_name} pod events")
            field = "involvedObject.name="+daemonset_pod_name
            api_response = get_logs_api_instance.list_namespaced_event(
                namespace=namespace_value, pretty="True", field_selector=field)
            LOGGER.debug(str(api_response))
            LOGGER.error(
                "operator custom object should be deployed but it is not deployed hence asserting")
            assert False
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            assert False

    operator_object.delete()


def test_provisionerNodeSelector(_values):
    LOGGER.info("test_provisionerNodeSelector")
    LOGGER.info("provisionerNodeSelector is added to the cr file")
    test = inputfunc.read_operator_data(clusterconfig_value, namespace_value)
    operator_object = baseclass.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()
    if operator_object.check() is True:
        LOGGER.info("Operator custom object is deployed successfully")
        desired_daemonset_node, labeled_nodes = operator_object.get_scaleplugin_labeled_nodes(
            test["provisionerNodeSelector"])
        if desired_daemonset_node == labeled_nodes:
            LOGGER.info("labeled nodes are equal to desired daemonset nodes")
        else:
            LOGGER.error(
                "labeled nodes are not equal to desired daemonset nodes")
            assert False
    else:
        daemonset_pod_name = operator_object.get_driver_ds_pod_name()
        get_logs_api_instance = client.CoreV1Api()
        try:
            LOGGER.info(f"Checking for failure reason match in {daemonset_pod_name} pod events")
            field = "involvedObject.name="+daemonset_pod_name
            api_response = get_logs_api_instance.list_namespaced_event(
                namespace=namespace_value, pretty="True", field_selector=field)
            LOGGER.debug(str(api_response))
            LOGGER.error(
                "operator custom object should be deployed but it is not deployed hence asserting")
            assert False
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            assert False

    operator_object.delete()


def test_pluginNodeSelector(_values):
    LOGGER.info("test_pluginNodeSelector")
    LOGGER.info("pluginNodeSelector is added to the cr file")
    test = inputfunc.read_operator_data(clusterconfig_value, namespace_value)
    operator_object = baseclass.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()
    if operator_object.check() is True:
        LOGGER.info("Operator custom object is deployed successfully")
        desired_daemonset_node, labeled_nodes = operator_object.get_scaleplugin_labeled_nodes(
            test["pluginNodeSelector"])
        if desired_daemonset_node == labeled_nodes:
            LOGGER.info("labeled nodes are equal to desired daemonset nodes")
        else:
            LOGGER.error(
                "labeled nodes are not equal to desired daemonset nodes")
            assert False
    else:
        daemonset_pod_name = operator_object.get_driver_ds_pod_name()
        get_logs_api_instance = client.CoreV1Api()
        try:
            LOGGER.info(f"Checking for failure reason match in {daemonset_pod_name} pod events")
            field = "involvedObject.name="+daemonset_pod_name
            api_response = get_logs_api_instance.list_namespaced_event(
                namespace=namespace_value, pretty="True", field_selector=field)
            LOGGER.debug(str(api_response))
            LOGGER.error(
                "operator custom object should be deployed but it is not deployed hence asserting")
            assert False
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            assert False

    operator_object.delete()


def test_non_deafult_snapshotter(_values):
    LOGGER.info("test_non_deafult_snapshotter")
    LOGGER.info("snapshotter image name is changed")
    test = inputfunc.read_operator_data(clusterconfig_value, namespace_value)
    deployment_snapshotter_image = "us.gcr.io/k8s-artifacts-prod/sig-storage/csi-snapshotter:v4.1.1"
    test["custom_object_body"]["spec"]["snapshotter"] = deployment_snapshotter_image
    operator_object = baseclass.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()
    if operator_object.check() is True:
        LOGGER.info("Operator custom object is deployed successfully")
        baseclass.kubeobjectfunc.check_pod_image(test["csiscaleoperator_name"]+"-snapshotter-0", deployment_snapshotter_image)
    else:
        get_logs_api_instance = client.CoreV1Api()
        try:
            daemonset_pod_name = operator_object.get_driver_ds_pod_name()
            LOGGER.info(f"Checking for failure reason match in {daemonset_pod_name} pod events")
            field = "involvedObject.name="+daemonset_pod_name
            api_response = get_logs_api_instance.list_namespaced_event(
                namespace=namespace_value, pretty="True", field_selector=field)
            LOGGER.debug(str(api_response))
            LOGGER.error(
                "operator custom object should be deployed but it is not deployed hence asserting")
            assert False
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            assert False

    operator_object.delete()


def test_non_deafult_livenessprobe(_values):
    LOGGER.info("test_non_deafult_livenessprobe")
    LOGGER.info("livenessprobe image name is changed")
    test = inputfunc.read_operator_data(clusterconfig_value, namespace_value)
    deployment_livenessprobe_image = "us.gcr.io/k8s-artifacts-prod/sig-storage/livenessprobe:v2.3.0"
    test["custom_object_body"]["spec"]["livenessprobe"] = deployment_livenessprobe_image
    operator_object = baseclass.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()
    if operator_object.check() is True:
        LOGGER.info("Operator custom object is deployed successfully")
        daemonset_pod_name = operator_object.get_driver_ds_pod_name()
        baseclass.kubeobjectfunc.check_pod_image(daemonset_pod_name, deployment_livenessprobe_image)
    else:
        get_logs_api_instance = client.CoreV1Api()
        try:
            daemonset_pod_name = operator_object.get_driver_ds_pod_name()
            LOGGER.info(f"Checking for failure reason match in {daemonset_pod_name} pod events")
            field = "involvedObject.name="+daemonset_pod_name
            api_response = get_logs_api_instance.list_namespaced_event(
                namespace=namespace_value, pretty="True", field_selector=field)
            LOGGER.debug(str(api_response))
            LOGGER.error(
                "operator custom object should be deployed but it is not deployed hence asserting")
            assert False
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            assert False

    operator_object.delete()


def test_non_deafult_resizer(_values):
    LOGGER.info("test_non_deafult_resizer")
    LOGGER.info("resizer image name is changed")
    test = inputfunc.read_operator_data(clusterconfig_value, namespace_value)
    deployment_resizer_image = "us.gcr.io/k8s-artifacts-prod/sig-storage/csi-resizer:v1.3.0"
    test["custom_object_body"]["spec"]["resizer"] = deployment_resizer_image
    operator_object = baseclass.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()
    if operator_object.check() is True:
        LOGGER.info("Operator custom object is deployed successfully")
        baseclass.kubeobjectfunc.check_pod_image(test["csiscaleoperator_name"]+"-resizer-0", deployment_resizer_image)
    else:
        get_logs_api_instance = client.CoreV1Api()
        try:
            daemonset_pod_name = operator_object.get_driver_ds_pod_name()
            LOGGER.info(f"Checking for failure reason match in {daemonset_pod_name} pod events")
            field = "involvedObject.name="+daemonset_pod_name
            api_response = get_logs_api_instance.list_namespaced_event(
                namespace=namespace_value, pretty="True", field_selector=field)
            LOGGER.debug(str(api_response))
            LOGGER.error(
                "operator custom object should be deployed but it is not deployed hence asserting")
            assert False
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            assert False

    operator_object.delete()


def test_snapshotterNodeSelector(_values):
    LOGGER.info("test_snapshotterNodeSelector")
    LOGGER.info("snapshotterNodeSelector is added to the cr file")
    test = inputfunc.read_operator_data(clusterconfig_value, namespace_value)
    operator_object = baseclass.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()
    if operator_object.check() is True:
        LOGGER.info("Operator custom object is deployed successfully")
        desired_daemonset_node, labeled_nodes = operator_object.get_scaleplugin_labeled_nodes(
            test["snapshotterNodeSelector"])
        if desired_daemonset_node == labeled_nodes:
            LOGGER.info("labeled nodes are equal to desired daemonset nodes")
        else:
            LOGGER.error(
                "labeled nodes are not equal to desired daemonset nodes")
            assert False
    else:
        get_logs_api_instance = client.CoreV1Api()
        try:
            daemonset_pod_name = operator_object.get_driver_ds_pod_name()
            LOGGER.info(f"Checking for failure reason match in {daemonset_pod_name} pod events")
            field = "involvedObject.name="+daemonset_pod_name
            api_response = get_logs_api_instance.list_namespaced_event(
                namespace=namespace_value, pretty="True", field_selector=field)
            LOGGER.debug(str(api_response))
            LOGGER.error(
                "operator custom object should be deployed but it is not deployed hence asserting")
            assert False
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            assert False

    operator_object.delete()


def test_resizerNodeSelector(_values):
    LOGGER.info("test_resizerNodeSelector")
    LOGGER.info("resizerNodeSelector is added to the cr file")
    test = inputfunc.read_operator_data(clusterconfig_value, namespace_value)
    operator_object = baseclass.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()
    if operator_object.check() is True:
        LOGGER.info("Operator custom object is deployed successfully")
        desired_daemonset_node, labeled_nodes = operator_object.get_scaleplugin_labeled_nodes(
            test["resizerNodeSelector"])
        if desired_daemonset_node == labeled_nodes:
            LOGGER.info("labeled nodes are equal to desired daemonset nodes")
        else:
            LOGGER.error(
                "labeled nodes are not equal to desired daemonset nodes")
            assert False
    else:
        get_logs_api_instance = client.CoreV1Api()
        try:
            daemonset_pod_name = operator_object.get_driver_ds_pod_name()
            LOGGER.info(f"Checking for failure reason match in {daemonset_pod_name} pod events")
            field = "involvedObject.name="+daemonset_pod_name
            api_response = get_logs_api_instance.list_namespaced_event(
                namespace=namespace_value, pretty="True", field_selector=field)
            LOGGER.debug(str(api_response))
            LOGGER.error(
                "operator custom object should be deployed but it is not deployed hence asserting")
            assert False
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            assert False

    operator_object.delete()


def test_wrong_kubeletRootDirPath(_values):
    LOGGER.info("test_wrong_kubeletRootDirPath : kubeletRootDirPath is wrong")
    test = inputfunc.read_operator_data(clusterconfig_value, namespace_value)

    test["custom_object_body"]["spec"]["kubeletRootDirPath"] = f"/{inputfunc.randomString()}/{inputfunc.randomString()}"

    operator_object = baseclass.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()
    if operator_object.check() is True:
        LOGGER.error(
            "Operator custom object is deployed successfully not expected")
        assert False
    else:
        daemonset_pod_name = operator_object.get_driver_ds_pod_name()
        api_instance = client.CoreV1Api()
        try:
            LOGGER.info(f"Checking for failure reason match in {daemonset_pod_name} pod events")
            field = "involvedObject.name="+daemonset_pod_name
            api_response = api_instance.list_namespaced_event(
                namespace=namespace_value, pretty="True", field_selector=field)
            LOGGER.debug(str(api_response))
            search_result = re.search(
                "hostPath type check failed", str(api_response))
            LOGGER.debug(search_result)
            assert search_result is not None
            LOGGER.info("'hostPath type check failed' failure reason matched")
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            assert False
    operator_object.delete()
