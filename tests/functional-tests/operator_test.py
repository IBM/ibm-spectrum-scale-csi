import time
import re
import random
import logging
import pytest
from kubernetes import client
from kubernetes.client.rest import ApiException
from scale_operator import read_scale_config_file, Scaleoperator, \
    check_nodes_available, Scaleoperatorobject, check_key
from utils.scale_operator_object_function import randomStringDigits, randomString
import utils.fileset_functions as ff
LOGGER = logging.getLogger()


@pytest.fixture(scope='session')
def _values(request):

    global kubeconfig_value, clusterconfig_value, namespace_value
    kubeconfig_value = request.config.option.kubeconfig
    if kubeconfig_value is None:
        kubeconfig_value = "~/.kube/config"
    clusterconfig_value = request.config.option.clusterconfig
    if clusterconfig_value is None:
        clusterconfig_value = "../../operator/deploy/crds/csiscaleoperators.csi.ibm.com_cr.yaml"
    namespace_value = request.config.option.namespace
    if namespace_value is None:
        namespace_value = "ibm-spectrum-scale-csi-driver"
    operator = Scaleoperator(kubeconfig_value)
    read_file = read_scale_config_file(clusterconfig_value, namespace_value)
    operator.create(namespace_value, read_file)
    operator.check()
    check_nodes_available(
        read_file["pluginNodeSelector"], "pluginNodeSelector")
    check_nodes_available(
        read_file["provisionerNodeSelector"], "provisionerNodeSelector")
    check_nodes_available(
        read_file["attacherNodeSelector"], "attacherNodeSelector")

    yield
    operator.delete()
    if(ff.fileset_exists(read_file)):
        ff.delete_fileset(read_file)


def test_operator_deploy(_values):

    LOGGER.info("test_operator_deploy")
    LOGGER.info("Every input is correct should run without any error")
    test = read_scale_config_file(clusterconfig_value, namespace_value)
    if(ff.fileset_exists(test)):
        ff.delete_fileset(test)
    operator_object = Scaleoperatorobject(test)
    operator_object.create(kubeconfig_value)
    if operator_object.check(kubeconfig_value) is True:
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
            operator_object.delete(kubeconfig_value)
            assert False
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            assert False
    operator_object.delete(kubeconfig_value)


def test_wrong_cluster_id(_values):
    LOGGER.info("test_wrong_cluster_id")
    LOGGER.info("cluster ID is wrong")
    test = read_scale_config_file(clusterconfig_value, namespace_value)
    test["id"] = str(random.randint(0, 999999999999999999))
    if(ff.fileset_exists(test)):
        ff.delete_fileset(test)
    operator_object = Scaleoperatorobject(test)
    operator_object.create(kubeconfig_value)
    if operator_object.check(kubeconfig_value) is True:
        LOGGER.error(
            "Operator custom object is deployed successfully not expected")
        assert False
    else:
        demonset_pod_name = operator_object.get_driver_ds_pod_name()
        get_logs_api_instance = client.CoreV1Api()
        try:
            get_logs_api_response = get_logs_api_instance.read_namespaced_pod_log(
                name=demonset_pod_name, namespace=namespace_value, container="ibm-spectrum-scale-csi")
            LOGGER.debug(str(get_logs_api_response))
            search_result = re.search(
                "Cluster ID doesnt match the cluster", get_logs_api_response)
            LOGGER.debug(search_result)
            assert search_result is not None
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            assert False
    operator_object.delete(kubeconfig_value)


def test_wrong_primaryFS(_values):
    LOGGER.info("test_wrong_primaryFS")
    LOGGER.info("primaryFS is wrong")
    test = read_scale_config_file(clusterconfig_value, namespace_value)
    if(ff.fileset_exists(test)):
        ff.delete_fileset(test)
    test["primaryFs"] = randomStringDigits()
    operator_object = Scaleoperatorobject(test)
    operator_object.create(kubeconfig_value)
    if operator_object.check(kubeconfig_value) is True:
        LOGGER.error(
            "Operator custom object is deployed successfully not expected")
        assert False
    else:
        demonset_pod_name = operator_object.get_driver_ds_pod_name()
        get_logs_api_instance = client.CoreV1Api()
        try:
            get_logs_api_response = get_logs_api_instance.read_namespaced_pod_log(
                name=demonset_pod_name, namespace=namespace_value, container="ibm-spectrum-scale-csi")
            LOGGER.debug(str(get_logs_api_response))
            search_result = re.search(
                "Unable to get filesystem", get_logs_api_response)
            LOGGER.debug(search_result)
            assert search_result is not None

        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            assert False
    operator_object.delete(kubeconfig_value)


def test_wrong_guihost(_values):
    LOGGER.info("test_wrong_guihost")
    LOGGER.info("gui host is wrong")
    test = read_scale_config_file(clusterconfig_value, namespace_value)
    if(ff.fileset_exists(test)):
        ff.delete_fileset(test)
    test["guiHost"] = randomStringDigits()
    operator_object = Scaleoperatorobject(test)
    operator_object.create(kubeconfig_value)
    if operator_object.check(kubeconfig_value) is True:
        LOGGER.error(
            "Operator custom object is deployed successfully not expected")
        assert False
    else:
        demonset_pod_name = operator_object.get_driver_ds_pod_name()
        get_logs_api_instance = client.CoreV1Api()
        try:
            get_logs_api_response = get_logs_api_instance.read_namespaced_pod_log(
                name=demonset_pod_name, namespace=namespace_value, container="ibm-spectrum-scale-csi")
            LOGGER.debug(str(get_logs_api_response))
            search_result1 = re.search(
                "connection refused", get_logs_api_response)
            LOGGER.debug(search_result1)
            search_result2 = re.search("no such host", get_logs_api_response)
            LOGGER.debug(search_result2)
            assert (search_result1 is not None or search_result2 is not None)
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            assert False
    operator_object.delete(kubeconfig_value)


def test_wrong_gui_username(_values):
    LOGGER.info("test_wrong_gui_username")
    LOGGER.info("gui username is wrong")
    test = read_scale_config_file(clusterconfig_value, namespace_value)
    if(ff.fileset_exists(test)):
        ff.delete_fileset(test)
    test["username"] = randomStringDigits()
    operator_object = Scaleoperatorobject(test)
    operator_object.create(kubeconfig_value)
    if operator_object.check(kubeconfig_value) is True:
        LOGGER.error(
            "Operator custom object is deployed successfully not expected")
        assert False
    else:
        demonset_pod_name = operator_object.get_driver_ds_pod_name()
        get_logs_api_instance = client.CoreV1Api()
        try:
            get_logs_api_response = get_logs_api_instance.read_namespaced_pod_log(
                name=demonset_pod_name, namespace=namespace_value, container="ibm-spectrum-scale-csi")
            LOGGER.debug(str(get_logs_api_response))
            x = re.search("401 Unauthorized", get_logs_api_response)
            LOGGER.info(x)
            assert x is not None
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            assert False
    operator_object.delete(kubeconfig_value)


def test_wrong_gui_password(_values):
    LOGGER.info("test_wrong_gui_password")
    LOGGER.info("gui password is wrong")
    test = read_scale_config_file(clusterconfig_value, namespace_value)
    if(ff.fileset_exists(test)):
        ff.delete_fileset(test)
    test["password"] = randomStringDigits()
    operator_object = Scaleoperatorobject(test)
    operator_object.create(kubeconfig_value)
    operator_object.check(kubeconfig_value)
    demonset_pod_name = operator_object.get_driver_ds_pod_name()
    get_logs_api_instance = client.CoreV1Api()
    count = 0
    while count < 24:
        try:
            get_logs_api_response = get_logs_api_instance.read_namespaced_pod_log(
                name=demonset_pod_name, namespace=namespace_value, container="ibm-spectrum-scale-csi")
            LOGGER.debug(str(get_logs_api_response))
            search_result = re.search(
                "Error in authentication request", get_logs_api_response)
            if search_result is None:
                time.sleep(5)
            else:
                LOGGER.debug(search_result)
                break
            if count > 23:
                operator_object.delete(kubeconfig_value)
                assert search_result is not None
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            assert False
    operator_object.delete(kubeconfig_value)


def test_wrong_secret_object_name(_values):
    LOGGER.info("test_wrong_secret_object_name")
    LOGGER.info("secret object name is wrong")
    test = read_scale_config_file(clusterconfig_value, namespace_value)
    test["secrets_name_wrong"] = randomString()
    test["stateful_set_not_created"] = True
    if(ff.fileset_exists(test)):
        ff.delete_fileset(test)
    operator_object = Scaleoperatorobject(test)
    operator_object.create(kubeconfig_value)
    operator_object.delete(kubeconfig_value)


def test_random_gpfs_primaryFset_name(_values):
    LOGGER.info("test_ramdom_gpfs_primaryFset_name")
    LOGGER.info("gpfs primary Fset name is wrong")
    test = read_scale_config_file(clusterconfig_value, namespace_value)
    test["primaryFset"] = randomStringDigits()
    if(ff.fileset_exists(test)):
        ff.delete_fileset(test)
    operator_object = Scaleoperatorobject(test)
    operator_object.create(kubeconfig_value)
    if operator_object.check(kubeconfig_value) is True:
        LOGGER.info("Operator custom object is deployed successfully")
    else:
        get_logs_api_instance = client.CoreV1Api()
        try:
            demonset_pod_name = operator_object.get_driver_ds_pod_name()
            get_logs_api_response = get_logs_api_instance.read_namespaced_pod_log(
                name=demonset_pod_name, namespace=namespace_value, container="ibm-spectrum-scale-csi")
            LOGGER.debug(str(get_logs_api_response))
            LOGGER.error(
                "operator custom object should be deployed but it is not deployed hence asserting")
            operator_object.delete(kubeconfig_value)
            if(ff.fileset_exists(test)):
                ff.delete_fileset(test)
            assert False
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            operator_object.delete(kubeconfig_value)
            if(ff.fileset_exists(test)):
                ff.delete_fileset(test)
            assert False
    if(ff.fileset_exists(test)):
        ff.delete_fileset(test)
    operator_object.delete(kubeconfig_value)


def test_secureSslMode(_values):
    LOGGER.info("test_secureSslMode")
    LOGGER.info("secureSslMode is True while cacert is not available")
    test = read_scale_config_file(clusterconfig_value, namespace_value)
    test["secureSslMode"] = True
    if check_key(test,"cacert_name"):
        test.pop("cacert_name") 

    if(ff.fileset_exists(test)):
        ff.delete_fileset(test)
    operator_object = Scaleoperatorobject(test)
    operator_object.create(kubeconfig_value)
    if operator_object.check(kubeconfig_value) is True:
        LOGGER.error(
            "Operator custom object is deployed successfully not expected")
        assert False
    else:
        demonset_pod_name = operator_object.get_driver_ds_pod_name()
        get_logs_api_instance = client.CoreV1Api()
        try:
            get_logs_api_response = get_logs_api_instance.read_namespaced_pod_log(
                name=demonset_pod_name, namespace=namespace_value, container="ibm-spectrum-scale-csi")
            LOGGER.debug(str(get_logs_api_response))
            search_result = re.search(
                "CA certificate not specified in secure SSL mode for cluster", str(get_logs_api_response))
            LOGGER.debug(search_result)
            if(search_result is None):
                operator_object.delete(kubeconfig_value)
                LOGGER.error(str(get_logs_api_response))
                LOGGER.error("Reason of failure does not match")
            assert search_result is not None
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            assert False
    if(ff.fileset_exists(test)):
        ff.delete_fileset(test)
    operator_object.delete(kubeconfig_value)


def test_wrong_gpfs_filesystem_mount_point(_values):
    LOGGER.info("test_wrong_gpfs_filesystem_mount_point")
    LOGGER.info("gpfs filesystem mount point is wrong")
    test = read_scale_config_file(clusterconfig_value, namespace_value)
    test["scaleHostpath"] = randomStringDigits()
    if(ff.fileset_exists(test)):
        ff.delete_fileset(test)
    operator_object = Scaleoperatorobject(test)
    operator_object.create(kubeconfig_value)

    if operator_object.check(kubeconfig_value) is True:
        LOGGER.error(
            "Operator custom object is deployed successfully not expected")
        assert False
    else:
        get_logs_api_instance = client.CoreV1Api()
        demonset_pod_name = operator_object.get_driver_ds_pod_name()
        try:
            field = "involvedObject.name="+demonset_pod_name
            api_response = get_logs_api_instance.list_namespaced_event(
                namespace=namespace_value, pretty="True", field_selector=field)
            LOGGER.debug(str(api_response))
            search_result = re.search(
                'MountVolume.SetUp failed for volume', str(api_response))
            LOGGER.debug(search_result)
            assert search_result is not None
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            assert False

    operator_object.delete(kubeconfig_value)


def test_unlinked_primaryFset(_values):
    LOGGER.info("test_unlinked_primaryFset")
    LOGGER.info("unlinked primaryFset expected : object created successfully")
    test = read_scale_config_file(clusterconfig_value, namespace_value)
    if(ff.fileset_exists(test)):
        ff.delete_fileset(test)
    ff.create_fileset(test)
    ff.unlink_fileset(test)
    operator_object = Scaleoperatorobject(test)
    operator_object.create(kubeconfig_value)
    if operator_object.check(kubeconfig_value) is True:
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
            operator_object.delete(kubeconfig_value)
            assert False
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            assert False

    operator_object.delete(kubeconfig_value)


def test_existing_primaryFset(_values):
    LOGGER.info("test_existing_primaryFset")
    LOGGER.info(
        "linked existing primaryFset expected : object created successfully")
    test = read_scale_config_file(clusterconfig_value, namespace_value)
    if(ff.fileset_exists(test)):
        ff.delete_fileset(test)
    ff.create_fileset(test)
    operator_object = Scaleoperatorobject(test)
    operator_object.create(kubeconfig_value)
    if operator_object.check(kubeconfig_value) is True:
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
            operator_object.delete(kubeconfig_value)
            assert False
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            assert False

    operator_object.delete(kubeconfig_value)


def test_unmounted_primaryFS(_values):
    LOGGER.info("test_unmounted_primaryFS")
    LOGGER.info(
        "primaryFS is unmounted and expected : custom object successfully deployed")
    test = read_scale_config_file(clusterconfig_value, namespace_value)
    if(ff.fileset_exists(test)):
        ff.delete_fileset(test)
    ff.unmount_fs(test)
    operator_object = Scaleoperatorobject(test)
    operator_object.create(kubeconfig_value)
    if operator_object.check(kubeconfig_value) is True:
        LOGGER.error(
            "Operator custom object is deployed successfully, it is not expected")
        operator_object.delete(kubeconfig_value)
        ff.mount_fs(test)
        assert False
    else:
        get_logs_api_instance = client.CoreV1Api()
        try:
            demonset_pod_name = operator_object.get_driver_ds_pod_name()
            get_logs_api_response = get_logs_api_instance.read_namespaced_pod_log(
                name=demonset_pod_name, namespace=namespace_value, container="ibm-spectrum-scale-csi")
            LOGGER.debug(str(get_logs_api_response))
            search_result = re.search(
                'Unable to link primary fileset', str(get_logs_api_response))
            if search_result is None:
                LOGGER.error(str(get_logs_api_response))
            LOGGER.debug(search_result)
            operator_object.delete(kubeconfig_value)
            ff.mount_fs(test)
            assert search_result is not None
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            ff.mount_fs(test)
            assert False
    operator_object.delete(kubeconfig_value)
    ff.mount_fs(test)


def test_non_deafult_attacher(_values):
    LOGGER.info("test_non_deafult_attacher")
    LOGGER.info("attacher image name is changed")
    test = read_scale_config_file(clusterconfig_value, namespace_value)
    test["deployment_attacher_image"] = "quay.io/k8scsi/csi-attacher:v1.2.1"
    if(ff.fileset_exists(test)):
        ff.delete_fileset(test)
    operator_object = Scaleoperatorobject(test)
    operator_object.create(kubeconfig_value)
    if operator_object.check(kubeconfig_value) is True:
        LOGGER.info("Operator custom object is deployed successfully")
    else:
        get_logs_api_instance = client.CoreV1Api()
        try:
            demonset_pod_name = operator_object.get_driver_ds_pod_name()
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

    operator_object.delete(kubeconfig_value)


def test_non_deafult_provisioner(_values):
    LOGGER.info("test_non_deafult_provisioner")
    LOGGER.info("provisioner image name is changed")
    test = read_scale_config_file(clusterconfig_value, namespace_value)
    test["deployment_provisioner_image"] = "quay.io/k8scsi/csi-provisioner:v1.0.2"

    if(ff.fileset_exists(test)):
        ff.delete_fileset(test)
    operator_object = Scaleoperatorobject(test)
    operator_object.create(kubeconfig_value)
    if operator_object.check(kubeconfig_value) is True:
        LOGGER.info("Operator custom object is deployed successfully")
    else:
        get_logs_api_instance = client.CoreV1Api()
        try:
            demonset_pod_name = operator_object.get_driver_ds_pod_name()
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

    operator_object.delete(kubeconfig_value)


def test_correct_cacert(_values):
    LOGGER.info("test_secureSslMode with correct cacert file")
    LOGGER.info("correct cacert file is given")
    test = read_scale_config_file(clusterconfig_value, namespace_value)
    test["secureSslMode"] = True
    if not(check_key(test,"cacert_name")):
        test["cacert_name"] = "test-cacert-configmap"
    operator_object = Scaleoperatorobject(test)
    if test["cacert_path"] == "":
        LOGGER.info("skipping the test as cacert file path is not given in conftest.py")
        pytest.skip("path of cacert file is not given")

    if(ff.fileset_exists(test)):
        ff.delete_fileset(test)
    operator_object.create(kubeconfig_value)
    if operator_object.check(kubeconfig_value) is True:
        LOGGER.info("Operator custom object is deployed successfully")
    else:
        demonset_pod_name = operator_object.get_driver_ds_pod_name()
        get_logs_api_instance = client.CoreV1Api()
        try:
            get_logs_api_response = get_logs_api_instance.read_namespaced_pod_log(
                name=demonset_pod_name, namespace=namespace_value, container="ibm-spectrum-scale-csi")
            LOGGER.info(str(get_logs_api_response))
            operator_object.delete(kubeconfig_value)
            LOGGER.error(
                "operator custom object should be deployed but it is not deployed hence asserting")
            assert False
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            assert False

    operator_object.delete(kubeconfig_value)


def test_cacert_with_secureSslMode_false(_values):
    LOGGER.info("test_cacert_with_secureSslMode_false")
    LOGGER.info("secureSslMode is false with correct cacert file")
    test = read_scale_config_file(clusterconfig_value, namespace_value)
    test["secureSslMode_explcit"] = False
    if not(check_key(test,"cacert_name")):
        test["cacert_name"] = "test-cacert-configmap"
    operator_object = Scaleoperatorobject(test)
    if test["cacert_path"] == "":
        LOGGER.info("skipping the test as cacert file path is not given in conftest.py")
        pytest.skip("path of cacert file is not given")

    if(ff.fileset_exists(test)):
        ff.delete_fileset(test)
    operator_object.create(kubeconfig_value)
    if operator_object.check(kubeconfig_value) is True:
        LOGGER.info("Operator custom object is deployed successfully")
    else:
        demonset_pod_name = operator_object.get_driver_ds_pod_name()
        get_logs_api_instance = client.CoreV1Api()
        try:
            get_logs_api_response = get_logs_api_instance.read_namespaced_pod_log(
                name=demonset_pod_name, namespace=namespace_value, container="ibm-spectrum-scale-csi")
            LOGGER.info(str(get_logs_api_response))
            operator_object.delete(kubeconfig_value)
            LOGGER.error(
                "operator custom object should be deployed but it is not deployed hence asserting")
            assert False
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            assert False

    operator_object.delete(kubeconfig_value)


def test_wrong_cacert(_values):
    LOGGER.info("secureSslMode true with wrong cacert file")
    LOGGER.info("test_wrong_cacert")
    test = read_scale_config_file(clusterconfig_value, namespace_value)
    test["secureSslMode"] = True
    if not(check_key(test,"cacert_name")):
        test["cacert_name"] = "test-cacert-configmap"
    test["make_cacert_wrong"] = True
    operator_object = Scaleoperatorobject(test)
    if test["cacert_path"] == "":
        LOGGER.info("skipping the test as cacert file path is not given in conftest.py")
        pytest.skip("path of cacert file is not given")

    if(ff.fileset_exists(test)):
        ff.delete_fileset(test)
    operator_object.create(kubeconfig_value)
    if operator_object.check(kubeconfig_value) is True:
        LOGGER.error(
            "Operator custom object is deployed successfully not expected")
        operator_object.delete(kubeconfig_value)
        assert False
    else:
        demonset_pod_name = operator_object.get_driver_ds_pod_name()
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
                    operator_object.delete(kubeconfig_value)
                    assert search_result is not None
            except ApiException as e:
                LOGGER.error(
                    f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
                assert False

    operator_object.delete(kubeconfig_value)


def test_nodeMapping(_values):
    LOGGER.info("test_nodeMapping")
    LOGGER.info("nodeMapping is added to the cr file")
    test = read_scale_config_file(clusterconfig_value, namespace_value)
    operator_object = Scaleoperatorobject(test)
    if(ff.fileset_exists(test)):
        ff.delete_fileset(test)
    operator_object.create(kubeconfig_value)
    if operator_object.check(kubeconfig_value) is True:
        LOGGER.info("Operator custom object is deployed successfully")
    else:
        get_logs_api_instance = client.CoreV1Api()
        try:
            demonset_pod_name = operator_object.get_driver_ds_pod_name()
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

    operator_object.delete(kubeconfig_value)


def test_attacherNodeSelector(_values):
    LOGGER.info("test_attacherNodeSelector")
    LOGGER.info("attacherNodeSelector is added to the cr file")
    test = read_scale_config_file(clusterconfig_value, namespace_value)
    operator_object = Scaleoperatorobject(test)
    if(ff.fileset_exists(test)):
        ff.delete_fileset(test)
    operator_object.create(kubeconfig_value)
    if operator_object.check(kubeconfig_value) is True:
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

    operator_object.delete(kubeconfig_value)


def test_provisionerNodeSelector(_values):
    LOGGER.info("test_provisionerNodeSelector")
    LOGGER.info("provisionerNodeSelector is added to the cr file")
    test = read_scale_config_file(clusterconfig_value, namespace_value)
    operator_object = Scaleoperatorobject(test)
    if(ff.fileset_exists(test)):
        ff.delete_fileset(test)
    operator_object.create(kubeconfig_value)
    if operator_object.check(kubeconfig_value) is True:
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

    operator_object.delete(kubeconfig_value)


def test_pluginNodeSelector(_values):
    LOGGER.info("test_pluginNodeSelector")
    LOGGER.info("pluginNodeSelector is added to the cr file")
    test = read_scale_config_file(clusterconfig_value, namespace_value)
    operator_object = Scaleoperatorobject(test)
    if(ff.fileset_exists(test)):
        ff.delete_fileset(test)
    operator_object.create(kubeconfig_value)
    if operator_object.check(kubeconfig_value) is True:
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

    operator_object.delete(kubeconfig_value)

'''
def test_remote_operator_deploy(_values):

    LOGGER.info("test_remote_operator_deploy")
    LOGGER.info("should run without any error using remote")
    test = read_scale_config_file(clusterconfig_value, namespace_value)
    #if(ff.fileset_exists(test)):
    #    ff.delete_fileset(test)
    #test["remote"] = True
    operator_object = Scaleoperatorobject(test)
    operator_object.create(kubeconfig_value)
    if operator_object.check(kubeconfig_value) is True:
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
            operator_object.delete(kubeconfig_value)
            assert False
        except ApiException as e:
            LOGGER.error(
                f"Exception when calling CoreV1Api->read_namespaced_pod_log: {e}")
            assert False
    operator_object.delete(kubeconfig_value)

'''
