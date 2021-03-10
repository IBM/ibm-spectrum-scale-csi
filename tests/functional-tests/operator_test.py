import time
import re
import random
import logging
import pytest
from kubernetes import client
from kubernetes.client.rest import ApiException
import scale_operator as scaleop
from utils.scale_operator_object_function import randomStringDigits, randomString
import utils.fileset_functions as ff
LOGGER = logging.getLogger()


@pytest.fixture(scope='session')
def _values(request):

    global kubeconfig_value, clusterconfig_value, namespace_value
    kubeconfig_value, clusterconfig_value, namespace_value, _ = scaleop.get_cmd_values(request)

    condition = scaleop.check_ns_exists(kubeconfig_value, namespace_value)
    operator = scaleop.Scaleoperator(kubeconfig_value, namespace_value)
    read_file = scaleop.read_operator_data(clusterconfig_value, namespace_value)
    ff.cred_check(read_file)
    fileset_exist = ff.fileset_exists(read_file)
    operator.create()
    operator.check()
    scaleop.check_nodes_available(
        read_file["pluginNodeSelector"], "pluginNodeSelector")
    scaleop.check_nodes_available(
        read_file["provisionerNodeSelector"], "provisionerNodeSelector")
    scaleop.check_nodes_available(
        read_file["attacherNodeSelector"], "attacherNodeSelector")

    yield
    operator.delete(condition)
    if(not(fileset_exist) and ff.fileset_exists(read_file)):
        ff.delete_fileset(read_file)

@pytest.mark.regression
def test_get_version(_values):
    test = scaleop.read_operator_data(clusterconfig_value, namespace_value)
    ff.get_scale_version(test)
    scaleop.get_kubernetes_version(kubeconfig_value)
    scaleop.scale_function.get_operator_image()


def test_operator_deploy(_values):

    LOGGER.info("test_operator_deploy")
    LOGGER.info("Every input is correct should run without any error")
    test = scaleop.read_operator_data(clusterconfig_value, namespace_value)
    operator_object = scaleop.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()
    if operator_object.check() is True:
        LOGGER.info("Operator custom object is deployed successfully")
    else:
        demonset_pod_name = operator_object.get_driver_ds_pod_name()
        get_logs_api_instance = client.CoreV1Api()
        try:
            get_logs_api_response = get_logs_api_instance.read_namespaced_pod_log(
                name=demonset_pod_name, namespace=namespace_value, container="ibm-spectrum-scale-csi")
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
    test = scaleop.read_operator_data(clusterconfig_value, namespace_value)
    wrong_id = str(random.randint(0, 999999999999999999))

    for cluster in test["custom_object_body"]["spec"]["clusters"]:
        if "primary" in cluster.keys():
            cluster["id"] = wrong_id

    operator_object = scaleop.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()
    if operator_object.check() is True:
        LOGGER.error(
            "Operator custom object is deployed successfully not expected")
        assert False
    else:
        demonset_pod_name = operator_object.get_driver_ds_pod_name()
        get_logs_api_instance = client.CoreV1Api()
        try:
            LOGGER.info(f"Checking for failure reason match in {demonset_pod_name} pod logs")
            get_logs_api_response = get_logs_api_instance.read_namespaced_pod_log(
                name=demonset_pod_name, namespace=namespace_value, container="ibm-spectrum-scale-csi")
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
    test = scaleop.read_operator_data(clusterconfig_value, namespace_value)
    wrong_primaryFs = randomStringDigits()

    for cluster in test["custom_object_body"]["spec"]["clusters"]:
        if "primary" in cluster.keys():
            cluster["primary"]["primaryFs"] = wrong_primaryFs
    test["primaryFs"] = wrong_primaryFs
    operator_object = scaleop.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()
    if operator_object.check() is True:
        LOGGER.error(
            "Operator custom object is deployed successfully not expected")
        assert False
    else:
        demonset_pod_name = operator_object.get_driver_ds_pod_name()
        get_logs_api_instance = client.CoreV1Api()
        try:
            LOGGER.info(f"Checking for failure reason match in {demonset_pod_name} pod logs")
            get_logs_api_response = get_logs_api_instance.read_namespaced_pod_log(
                name=demonset_pod_name, namespace=namespace_value, container="ibm-spectrum-scale-csi")
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
    test = scaleop.read_operator_data(clusterconfig_value, namespace_value)
    wrong_guiHost = randomStringDigits()
    test["guiHost"] = wrong_guiHost
    for cluster in test["custom_object_body"]["spec"]["clusters"]:
        if "primary" in cluster.keys():
            cluster["restApi"][0]["guiHost"] = wrong_guiHost

    operator_object = scaleop.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()
    if operator_object.check() is True:
        LOGGER.error(
            "Operator custom object is deployed successfully not expected")
        assert False
    else:
        demonset_pod_name = operator_object.get_driver_ds_pod_name()
        get_logs_api_instance = client.CoreV1Api()
        try:
            LOGGER.info(f"Checking for failure reason match in {demonset_pod_name} pod logs")
            get_logs_api_response = get_logs_api_instance.read_namespaced_pod_log(
                name=demonset_pod_name, namespace=namespace_value, container="ibm-spectrum-scale-csi")
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
    test = scaleop.read_operator_data(clusterconfig_value, namespace_value)
    test["username"] = randomStringDigits()
    operator_object = scaleop.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()
    if operator_object.check() is True:
        LOGGER.error(
            "Operator custom object is deployed successfully not expected")
        assert False
    else:
        demonset_pod_name = operator_object.get_driver_ds_pod_name()
        get_logs_api_instance = client.CoreV1Api()
        try:
            LOGGER.info(f"Checking for failure reason match in {demonset_pod_name} pod logs")
            get_logs_api_response = get_logs_api_instance.read_namespaced_pod_log(
                name=demonset_pod_name, namespace=namespace_value, container="ibm-spectrum-scale-csi")
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
    test = scaleop.read_operator_data(clusterconfig_value, namespace_value)
    test["password"] = randomStringDigits()
    operator_object = scaleop.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()
    operator_object.check()
    LOGGER.info("Checkig if failure reason matches")
    demonset_pod_name = operator_object.get_driver_ds_pod_name()
    LOGGER.info(f"Checking for failure reason match in {demonset_pod_name} pod logs")
    get_logs_api_instance = client.CoreV1Api()
    count = 0
    while count < 24:
        try:
            get_logs_api_response = get_logs_api_instance.read_namespaced_pod_log(
                name=demonset_pod_name, namespace=namespace_value, container="ibm-spectrum-scale-csi")
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
    test = scaleop.read_operator_data(clusterconfig_value, namespace_value)
    secret_name_wrong = randomString()

    for cluster in test["custom_object_body"]["spec"]["clusters"]:
        if "primary" in cluster.keys():
            cluster["secrets"] = secret_name_wrong

    test["stateful_set_not_created"] = True
    operator_object = scaleop.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()
    operator_object.delete()


def test_random_gpfs_primaryFset_name(_values):
    LOGGER.info("test_random_gpfs_primaryFset_name : gpfs primary Fset name is wrong")
    test = scaleop.read_operator_data(clusterconfig_value, namespace_value)
    random_primaryFset = randomStringDigits()
    test["primaryFset"] = random_primaryFset
    for cluster in test["custom_object_body"]["spec"]["clusters"]:
        if "primary" in cluster.keys():
            cluster["primary"]["primaryFset"] = random_primaryFset

    operator_object = scaleop.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()
    if operator_object.check() is True:
        LOGGER.info("Operator custom object is deployed successfully")
    else:
        get_logs_api_instance = client.CoreV1Api()
        try:
            demonset_pod_name = operator_object.get_driver_ds_pod_name()
            LOGGER.info(f"Checking for failure reason match in {demonset_pod_name} pod logs")
            get_logs_api_response = get_logs_api_instance.read_namespaced_pod_log(
                name=demonset_pod_name, namespace=namespace_value, container="ibm-spectrum-scale-csi")
            LOGGER.debug(str(get_logs_api_response))
            LOGGER.error(
                "operator custom object should be deployed but it is not deployed hence asserting")
            operator_object.delete()
            if(ff.fileset_exists(test)):
                ff.delete_fileset(test)
            assert False
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            operator_object.delete()
            if(ff.fileset_exists(test)):
                ff.delete_fileset(test)
            assert False
    if(ff.fileset_exists(test)):
        ff.delete_fileset(test)
    operator_object.delete()


def test_secureSslMode(_values):
    LOGGER.info("test_secureSslMode")
    LOGGER.info("secureSslMode is True while cacert is not available")
    test = scaleop.read_operator_data(clusterconfig_value, namespace_value)

    for cluster in test["custom_object_body"]["spec"]["clusters"]:
        if "primary" in cluster.keys():
            cluster["secureSslMode"] = True
            if "cacert" in cluster.keys():
                cluster.pop("cacert")

    operator_object = scaleop.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()
    if operator_object.check() is True:
        LOGGER.error(
            "Operator custom object is deployed successfully not expected")
        assert False
    else:
        demonset_pod_name = operator_object.get_driver_ds_pod_name()
        get_logs_api_instance = client.CoreV1Api()
        try:
            LOGGER.info(f"Checking for failure reason match in {demonset_pod_name} pod logs")
            get_logs_api_response = get_logs_api_instance.read_namespaced_pod_log(
                name=demonset_pod_name, namespace=namespace_value, container="ibm-spectrum-scale-csi")
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
    if(ff.fileset_exists(test)):
        ff.delete_fileset(test)
    operator_object.delete()

"""
Removing this testcase as scaleHostpath is no longer needed
def test_wrong_gpfs_filesystem_mount_point(_values):
    LOGGER.info("test_wrong_gpfs_filesystem_mount_point")
    LOGGER.info("gpfs filesystem mount point is wrong")
    test = scaleop.read_operator_data(clusterconfig_value, namespace_value)
    wrong_scaleHostpath = randomStringDigits()
    test["custom_object_body"]["spec"]["scaleHostpath"] = wrong_scaleHostpath
    operator_object = scaleop.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()

    if operator_object.check() is True:
        LOGGER.error(
            "Operator custom object is deployed successfully not expected")
        assert False
    else:
        get_logs_api_instance = client.CoreV1Api()
        demonset_pod_name = operator_object.get_driver_ds_pod_name()
        try:
            LOGGER.info(f"Checking for failure reason match in {demonset_pod_name} pod events")
            field = "involvedObject.name="+demonset_pod_name
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
    test = scaleop.read_operator_data(clusterconfig_value, namespace_value)
    if(not(ff.fileset_exists(test))):
        ff.create_fileset(test)
    ff.unlink_fileset(test)
    operator_object = scaleop.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()
    if operator_object.check() is True:
        LOGGER.info("Operator custom object is deployed successfully")
    else:
        get_logs_api_instance = client.CoreV1Api()
        try:
            demonset_pod_name = operator_object.get_driver_ds_pod_name()
            LOGGER.info(f"Checking for failure reason match in {demonset_pod_name} pod logs")
            get_logs_api_response = get_logs_api_instance.read_namespaced_pod_log(
                name=demonset_pod_name, namespace=namespace_value, container="ibm-spectrum-scale-csi")
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
    test = scaleop.read_operator_data(clusterconfig_value, namespace_value)
    if(not(ff.fileset_exists(test))):
        ff.create_fileset(test)
    operator_object = scaleop.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()
    if operator_object.check() is True:
        LOGGER.info("Operator custom object is deployed successfully")
    else:
        get_logs_api_instance = client.CoreV1Api()
        try:
            demonset_pod_name = operator_object.get_driver_ds_pod_name()
            get_logs_api_response = get_logs_api_instance.read_namespaced_pod_log(
                name=demonset_pod_name, namespace=namespace_value, container="ibm-spectrum-scale-csi")
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
    test = scaleop.read_operator_data(clusterconfig_value, namespace_value)
    ff.unmount_fs(test)
    operator_object = scaleop.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()
    if operator_object.check() is True:
        LOGGER.error(
            "Operator custom object is deployed successfully, it is not expected")
        operator_object.delete()
        ff.mount_fs(test)
        assert False
    else:
        get_logs_api_instance = client.CoreV1Api()
        try:
            demonset_pod_name = operator_object.get_driver_ds_pod_name()
            LOGGER.info(f"Checking for failure reason match in {demonset_pod_name} pod logs")
            get_logs_api_response = get_logs_api_instance.read_namespaced_pod_log(
                name=demonset_pod_name, namespace=namespace_value, container="ibm-spectrum-scale-csi")
            LOGGER.debug(str(get_logs_api_response))
            search_result = re.search(
                'not mounted on GUI node Primary cluster', str(get_logs_api_response))
            if search_result is None:
                LOGGER.error(str(get_logs_api_response))
            LOGGER.debug(search_result)
            operator_object.delete()
            ff.mount_fs(test)
            assert search_result is not None
            LOGGER.info("'not mounted on GUI node Primary cluster' failure reason matched")
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            ff.mount_fs(test)
            assert False
    operator_object.delete()
    ff.mount_fs(test)


def test_non_deafult_attacher(_values):
    LOGGER.info("test_non_deafult_attacher")
    LOGGER.info("attacher image name is changed")
    test = scaleop.read_operator_data(clusterconfig_value, namespace_value)
    deployment_attacher_image = "quay.io/k8scsi/csi-attacher:v1.2.1"
    test["custom_object_body"]["spec"]["attacher"] = deployment_attacher_image
    operator_object = scaleop.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()
    if operator_object.check() is True:
        LOGGER.info("Operator custom object is deployed successfully")
    else:
        get_logs_api_instance = client.CoreV1Api()
        try:
            demonset_pod_name = operator_object.get_driver_ds_pod_name()
            LOGGER.info(f"Checking for failure reason match in {demonset_pod_name} pod events")
            field = "involvedObject.name="+demonset_pod_name
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
    test = scaleop.read_operator_data(clusterconfig_value, namespace_value)
    deployment_provisioner_image = "quay.io/k8scsi/csi-provisioner:v1.6.0"
    test["custom_object_body"]["spec"]["provisioner"] = deployment_provisioner_image
    operator_object = scaleop.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()
    if operator_object.check() is True:
        LOGGER.info("Operator custom object is deployed successfully")
    else:
        get_logs_api_instance = client.CoreV1Api()
        try:
            demonset_pod_name = operator_object.get_driver_ds_pod_name()
            LOGGER.info(f"Checking for failure reason match in {demonset_pod_name} pod events")
            field = "involvedObject.name="+demonset_pod_name
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
    test = scaleop.read_operator_data(clusterconfig_value, namespace_value)

    if not("local_cacert_name" in test):
        test["local_cacert_name"] = "test-cacert-configmap"

    for cluster in test["custom_object_body"]["spec"]["clusters"]:
        if "primary" in cluster.keys():
            cluster["secureSslMode"] = True
            if not("cacert" in cluster.keys()):
                cluster["cacert"] = "test-cacert-configmap"

    operator_object = scaleop.Scaleoperatorobject(test, kubeconfig_value)
    if test["cacert_path"] == "":
        LOGGER.info("skipping the test as cacert file path is not given in test.config")
        pytest.skip("path of cacert file is not given")

    operator_object.create()
    if operator_object.check() is True:
        LOGGER.info("Operator custom object is deployed successfully")
    else:
        demonset_pod_name = operator_object.get_driver_ds_pod_name()
        get_logs_api_instance = client.CoreV1Api()
        try:
            LOGGER.info(f"Checking for failure reason match in {demonset_pod_name} pod logs")
            get_logs_api_response = get_logs_api_instance.read_namespaced_pod_log(
                name=demonset_pod_name, namespace=namespace_value, container="ibm-spectrum-scale-csi")
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
    test = scaleop.read_operator_data(clusterconfig_value, namespace_value)

    if not("local_cacert_name" in test):
        test["local_cacert_name"] = "test-cacert-configmap"

    for cluster in test["custom_object_body"]["spec"]["clusters"]:
        if "primary" in cluster.keys():
            cluster["secureSslMode"] = False
            if not("cacert" in cluster.keys()):
                cluster["cacert"] = "test-cacert-configmap"

    operator_object = scaleop.Scaleoperatorobject(test, kubeconfig_value)
    if test["cacert_path"] == "":
        LOGGER.info("skipping the test as cacert file path is not given in test.config")
        pytest.skip("path of cacert file is not given")

    operator_object.create()
    if operator_object.check() is True:
        LOGGER.info("Operator custom object is deployed successfully")
    else:
        demonset_pod_name = operator_object.get_driver_ds_pod_name()
        get_logs_api_instance = client.CoreV1Api()
        try:
            LOGGER.info(f"Checking for failure reason match in {demonset_pod_name} pod logs")
            get_logs_api_response = get_logs_api_instance.read_namespaced_pod_log(
                name=demonset_pod_name, namespace=namespace_value, container="ibm-spectrum-scale-csi")
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
    test = scaleop.read_operator_data(clusterconfig_value, namespace_value)

    if not("local_cacert_name" in test):
        test["local_cacert_name"] = "test-cacert-configmap"

    for cluster in test["custom_object_body"]["spec"]["clusters"]:
        if "primary" in cluster.keys():
            cluster["secureSslMode"] = True
            if not("cacert" in cluster.keys()):
                cluster["cacert"] = "test-cacert-configmap"

    test["make_cacert_wrong"] = True
    operator_object = scaleop.Scaleoperatorobject(test, kubeconfig_value)
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
        demonset_pod_name = operator_object.get_driver_ds_pod_name()
        LOGGER.info(f"Checking for failure reason match in {demonset_pod_name} pod logs")
        get_logs_api_instance = client.CoreV1Api()
        count = 0
        while count < 24:
            try:
                get_logs_api_response = get_logs_api_instance.read_namespaced_pod_log(
                    name=demonset_pod_name, namespace=namespace_value, container="ibm-spectrum-scale-csi")
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
    test = scaleop.read_operator_data(clusterconfig_value, namespace_value)
    operator_object = scaleop.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()
    if operator_object.check() is True:
        LOGGER.info("Operator custom object is deployed successfully")
    else:
        get_logs_api_instance = client.CoreV1Api()
        try:
            demonset_pod_name = operator_object.get_driver_ds_pod_name()
            LOGGER.info(f"Checking for failure reason match in {demonset_pod_name} pod events")
            field = "involvedObject.name="+demonset_pod_name
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
    test = scaleop.read_operator_data(clusterconfig_value, namespace_value)
    operator_object = scaleop.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()
    if operator_object.check() is True:
        LOGGER.info("Operator custom object is deployed successfully")
        desired_daemonset_node, labelled_nodes = operator_object.get_scaleplugin_labelled_nodes(
            test["attacherNodeSelector"])
        if desired_daemonset_node == labelled_nodes:
            LOGGER.info("labelled nodes are equal to desired daemonset nodes")
        else:
            LOGGER.error(
                "labelled nodes are not equal to desired daemonset nodes")
            assert False
    else:
        get_logs_api_instance = client.CoreV1Api()
        try:
            demonset_pod_name = operator_object.get_driver_ds_pod_name()
            LOGGER.info(f"Checking for failure reason match in {demonset_pod_name} pod events")
            field = "involvedObject.name="+demonset_pod_name
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
    test = scaleop.read_operator_data(clusterconfig_value, namespace_value)
    operator_object = scaleop.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()
    if operator_object.check() is True:
        LOGGER.info("Operator custom object is deployed successfully")
        desired_daemonset_node, labelled_nodes = operator_object.get_scaleplugin_labelled_nodes(
            test["provisionerNodeSelector"])
        if desired_daemonset_node == labelled_nodes:
            LOGGER.info("labelled nodes are equal to desired daemonset nodes")
        else:
            LOGGER.error(
                "labelled nodes are not equal to desired daemonset nodes")
            assert False
    else:
        demonset_pod_name = operator_object.get_driver_ds_pod_name()
        get_logs_api_instance = client.CoreV1Api()
        try:
            LOGGER.info(f"Checking for failure reason match in {demonset_pod_name} pod events")
            field = "involvedObject.name="+demonset_pod_name
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
    test = scaleop.read_operator_data(clusterconfig_value, namespace_value)
    operator_object = scaleop.Scaleoperatorobject(test, kubeconfig_value)
    operator_object.create()
    if operator_object.check() is True:
        LOGGER.info("Operator custom object is deployed successfully")
        desired_daemonset_node, labelled_nodes = operator_object.get_scaleplugin_labelled_nodes(
            test["pluginNodeSelector"])
        if desired_daemonset_node == labelled_nodes:
            LOGGER.info("labelled nodes are equal to desired daemonset nodes")
        else:
            LOGGER.error(
                "labelled nodes are not equal to desired daemonset nodes")
            assert False
    else:
        demonset_pod_name = operator_object.get_driver_ds_pod_name()
        get_logs_api_instance = client.CoreV1Api()
        try:
            LOGGER.info(f"Checking for failure reason match in {demonset_pod_name} pod events")
            field = "involvedObject.name="+demonset_pod_name
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
