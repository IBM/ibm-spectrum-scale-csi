from datetime import datetime
import pytest
from py.xml import html
import logging
import copy
import ibm_spectrum_scale_csi.base_class as baseclass
import ibm_spectrum_scale_csi.common_utils.input_data_functions as inputfunc
import ibm_spectrum_scale_csi.kubernetes_apis.csi_object_function as csiobjectfunc
import ibm_spectrum_scale_csi.kubernetes_apis.csi_storage_function as csistoragefunc
import ibm_spectrum_scale_csi.kubernetes_apis.kubernetes_objects_function as kubeobjectfunc

LOGGER = logging.getLogger()


def pytest_addoption(parser):
    parser.addoption("--kubeconfig", action="store")
    parser.addoption("--clusterconfig", action="store")
    parser.addoption("--testnamespace", default="ibm-spectrum-scale-csi-test", action="store")
    parser.addoption("--operatornamespace", default="ibm-spectrum-scale-csi-driver", action="store")
    parser.addoption("--runslow", action="store_true", help="run slow tests")
    parser.addoption("--operatoryaml", action="store")
    parser.addoption("--testconfig", default="config/test.config", action="store")
    parser.addoption("--createnamespace", action="store_true",
                     help="will create seperate namespace for each testcase")


def pytest_html_results_table_header(cells):
    cells.pop()


@pytest.hookimpl(hookwrapper=True)
def pytest_runtest_makereport(item, call):
    outcome = yield
    report = outcome.get_result()
    report.description = str(item.function.__doc__)


now = datetime.now()
dt_string = now.strftime("%d-%m-%Y-%H-%M-%S")
default_html_path = 'html-reports/csi-reports/csi-test-report-'+dt_string+'.html'


@pytest.hookimpl(tryfirst=True)
def pytest_configure(config):
    if not config.option.htmlpath:
        config.option.htmlpath = default_html_path
    config.addinivalue_line("markers", "slow: mark test as slow to run")


def pytest_collection_modifyitems(config, items):
    if config.getoption("--runslow"):
        # --runslow given in cli: do not skip slow tests
        return
    skip_slow = pytest.mark.skip(reason="need --runslow option to run")
    for item in items:
        if "slow" in item.keywords:
            item.add_marker(skip_slow)


@pytest.fixture(scope='session')
def data_fixture(request):
    data_fixture = {}
    data_fixture["cmd_values"] = inputfunc.get_pytest_cmd_values(request)
    data_fixture["driver_data"] = inputfunc.read_driver_data(data_fixture["cmd_values"])
    data_fixture["operator_data"] = inputfunc.read_operator_data(data_fixture["cmd_values"]["clusterconfig_value"],
                                                                 data_fixture["cmd_values"]["operator_namespace"], data_fixture["cmd_values"]["test_config"],
                                                                 data_fixture["cmd_values"]["kubeconfig_value"])
    if data_fixture["cmd_values"]["runslow_val"]:
        data_fixture["value_pvc"] = [{"access_modes": "ReadWriteMany", "storage": "1Gi"},
                                     {"access_modes": "ReadWriteOnce", "storage": "1Gi"},
                                     {"access_modes": "ReadOnlyMany", "storage": "1Gi",
                                      "reason": "ReadOnlyMany is not supported"}
                                     ]
        data_fixture["value_pod"] = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False"},
                                     {"mount_path": "/usr/share/nginx/html/scale",
                                      "read_only": "True", "reason": "Read-only file system"}
                                     ]
        data_fixture["snap_value_pvc"] = [{"access_modes": "ReadWriteMany", "storage": "1Gi"},
                                          {"access_modes": "ReadWriteOnce", "storage": "1Gi"}]
    else:
        data_fixture["value_pvc"] = data_fixture["snap_value_pvc"] = [
            {"access_modes": "ReadWriteMany", "storage": "1Gi"}]
        data_fixture["value_pod"] = [
            {"mount_path": "/usr/share/nginx/html/scale", "read_only": "False"}]

    data_fixture["value_vs_class"] = {"deletionPolicy": "Delete"}
    data_fixture["number_of_snapshots"] = 1
    data_fixture["cg_snap_value_pvc"] = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]

    data_fixture["driver_object"] = baseclass.Driver(data_fixture["cmd_values"]["kubeconfig_value"], data_fixture["value_pvc"],
                                                     data_fixture["value_pod"], data_fixture["driver_data"]["id"], data_fixture["cmd_values"]["test_namespace"],
                                                     data_fixture["driver_data"]["keepobjects"], data_fixture["driver_data"]["image_name"],
                                                     data_fixture["driver_data"]["pluginNodeSelector"])

    data_fixture["snapshot_object"] = baseclass.Snapshot(data_fixture["cmd_values"]["kubeconfig_value"],
                                                         data_fixture["cmd_values"]["test_namespace"], data_fixture["driver_data"]["keepobjects"],
                                                         data_fixture["snap_value_pvc"], data_fixture["value_vs_class"], data_fixture["number_of_snapshots"],
                                                         data_fixture["driver_data"]["image_name"], data_fixture["driver_data"]["id"],
                                                         data_fixture["driver_data"]["pluginNodeSelector"])

    return data_fixture


@pytest.fixture(scope='session')
def check_csi_operator(data_fixture):
    baseclass.filesetfunc.cred_check(data_fixture["driver_data"])
    baseclass.filesetfunc.set_data(data_fixture["driver_data"])
    operator = baseclass.Scaleoperator(data_fixture["cmd_values"]["kubeconfig_value"],
                                       data_fixture["cmd_values"]["operator_namespace"], data_fixture["cmd_values"]["operator_file"])
    operator_object = baseclass.Scaleoperatorobject(
        data_fixture["operator_data"], data_fixture["cmd_values"]["kubeconfig_value"])
    condition = baseclass.kubeobjectfunc.check_ns_exists(data_fixture["cmd_values"]["kubeconfig_value"],
                                                         data_fixture["cmd_values"]["operator_namespace"])
    if condition is True:
        if not(operator_object.check(data_fixture["driver_data"]["csiscaleoperator_name"])):
            LOGGER.error("Operator custom object is not deployed succesfully")
            assert False
    else:
        operator.create()
        operator.check()
        operator_object.create()
        if operator_object.check():
            LOGGER.info("Operator custom object is deployed succesfully")
        else:
            LOGGER.error("Operator custom object is not deployed succesfully")
            assert False


@pytest.fixture
def new_namespace(data_fixture):
    if data_fixture["cmd_values"]["createnamespace"] is True:
        data_fixture["cmd_values"]["test_namespace"] = csistoragefunc.get_random_name("csi-test")

    if not(kubeobjectfunc.check_namespace_exists(data_fixture["cmd_values"]["test_namespace"])):
        kubeobjectfunc.create_namespace(data_fixture["cmd_values"]["test_namespace"])

    data_fixture["driver_object"].test_ns = data_fixture["cmd_values"]["test_namespace"]
    data_fixture["snapshot_object"].test_namespace = data_fixture["cmd_values"]["test_namespace"]
    csistoragefunc.set_test_namespace_value(data_fixture["cmd_values"]["test_namespace"])
    yield
    if data_fixture["cmd_values"]["createnamespace"] is True and data_fixture["driver_data"]["keepobjects"] == "False":
        kubeobjectfunc.delete_namespace(data_fixture["cmd_values"]["test_namespace"])
        kubeobjectfunc.check_namespace_deleted(data_fixture["cmd_values"]["test_namespace"])


@pytest.fixture
def local_cluster_fixture(data_fixture, new_namespace):

    if data_fixture["driver_data"]["localFs"] in [None, ""]:
        LOGGER.error("Local Filesystem not provided in test.config file")
        assert False

    data_fixture["local_driver_object"] = copy.deepcopy(data_fixture["driver_object"])

    data_fixture["local_snapshot_object"] = copy.deepcopy(data_fixture["snapshot_object"])

    data_fixture["local_cg_snapshot_object"] = copy.deepcopy(data_fixture["snapshot_object"])
    data_fixture["local_cg_snapshot_object"].value_pvc = copy.deepcopy(
        data_fixture["cg_snap_value_pvc"])

    baseclass.filesetfunc.create_dir(data_fixture["driver_data"]["volDirBasePath"])


@pytest.fixture
def remote_cluster_fixture(data_fixture, new_namespace):

    if not("remote" in data_fixture["driver_data"]):
        LOGGER.error("remote data is not provided in CSO")
        assert False

    if data_fixture["driver_data"]["remoteFs"] is "":
        LOGGER.error("Remote Filesystem not provided in test.config file")
        assert False

    data_fixture["remote_data"] = inputfunc.get_remote_data(data_fixture["driver_data"])
    data_fixture["driver_data"]["remoteid"] = data_fixture["remote_data"]["remoteid"]
    baseclass.filesetfunc.cred_check(data_fixture["remote_data"])
    baseclass.filesetfunc.set_data(data_fixture["remote_data"])

    data_fixture["remote_driver_object"] = copy.deepcopy(data_fixture["driver_object"])
    data_fixture["remote_driver_object"].cluster_id = copy.deepcopy(
        data_fixture["remote_data"]["id"])

    data_fixture["remote_snapshot_object"] = copy.deepcopy(data_fixture["snapshot_object"])
    data_fixture["remote_snapshot_object"].cluster_id = copy.deepcopy(
        data_fixture["remote_data"]["id"])

    data_fixture["remote_cg_snapshot_object"] = copy.deepcopy(data_fixture["snapshot_object"])
    data_fixture["remote_snapshot_object"].cluster_id = copy.deepcopy(
        data_fixture["remote_data"]["id"])
    data_fixture["remote_snapshot_object"].value_pvc = copy.deepcopy(
        data_fixture["cg_snap_value_pvc"])

    baseclass.filesetfunc.create_dir(data_fixture["remote_data"]["volDirBasePath"])
