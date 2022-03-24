from datetime import datetime
import pytest
from py.xml import html
import logging
import ibm_spectrum_scale_csi.base_class as baseclass
import ibm_spectrum_scale_csi.common_utils.input_data_functions as inputfunc
import ibm_spectrum_scale_csi.kubernetes_apis.csi_object_function as csiobjectfunc
import ibm_spectrum_scale_csi.kubernetes_apis.csi_storage_function as csistoragefunc
import ibm_spectrum_scale_csi.kubernetes_apis.kubernetes_objects_function as kubeobjectfunc

LOGGER = logging.getLogger()

def pytest_addoption(parser):
    parser.addoption("--kubeconfig", action="store")
    parser.addoption("--clusterconfig", action="store")
    parser.addoption("--testnamespace", default="default",action="store")
    parser.addoption("--operatornamespace", default="ibm-spectrum-scale-csi-driver", action="store")
    parser.addoption("--runslow", action="store_true", help="run slow tests")
    parser.addoption("--operatoryaml", action="store")
    parser.addoption("--testconfig", default="config/test.config", action="store")
    parser.addoption("--createnamespace", action="store_true", help="will create seperate namespace for each testcase")

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
        data_fixture["value_pvc"] = data_fixture["snap_value_pvc"] = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
        data_fixture["value_pod"] = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False"}]
    
    data_fixture["value_vs_class"] = {"deletionPolicy": "Delete"}
    data_fixture["number_of_snapshots"] = 1
    data_fixture["cg_snap_value_pvc"] = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]

    return data_fixture

    
@pytest.fixture(scope='session')
def check_csi_operator(data_fixture):
    baseclass.filesetfunc.cred_check(data_fixture["driver_data"])
    baseclass.filesetfunc.set_data(data_fixture["driver_data"])
    kubeobjectfunc.set_global_namespace_value(data_fixture["cmd_values"]["operator_namespace"])
    csiobjectfunc.set_namespace_value(data_fixture["cmd_values"]["operator_namespace"])
    kubeobjectfunc.get_pod_list_and_check_running("product=ibm-spectrum-scale-csi",5)


@pytest.fixture
def new_namespace(data_fixture):
    if data_fixture["cmd_values"]["createnamespace"] is True:
        data_fixture["cmd_values"]["test_namespace"] = csistoragefunc.get_random_name("ns")
        kubeobjectfunc.create_namespace(data_fixture["cmd_values"]["test_namespace"])
    yield
    if data_fixture["cmd_values"]["createnamespace"] is True and data_fixture["driver_data"]["keepobjects"] is False:
        kubeobjectfunc.delete_namespace(data_fixture["cmd_values"]["test_namespace"])
        kubeobjectfunc.check_namespace_deleted(data_fixture["cmd_values"]["test_namespace"])


@pytest.fixture
def local_cluster_fixture(data_fixture, new_namespace):
    
    if not(data_fixture["driver_data"]["volBackendFs"] == ""):
        data_fixture["driver_data"]["primaryFs"] = data_fixture["driver_data"]["volBackendFs"]

    data_fixture["local_driver_object"] = baseclass.Driver(data_fixture["cmd_values"]["kubeconfig_value"], data_fixture["value_pvc"], 
                               data_fixture["value_pod"], data_fixture["driver_data"]["id"], data_fixture["cmd_values"]["test_namespace"], 
                               data_fixture["driver_data"]["keepobjects"], data_fixture["driver_data"]["image_name"], 
                               data_fixture["driver_data"]["pluginNodeSelector"])

    data_fixture["local_snapshot_object"] = baseclass.Snapshot(data_fixture["cmd_values"]["kubeconfig_value"], 
                               data_fixture["cmd_values"]["test_namespace"], data_fixture["driver_data"]["keepobjects"],
                               data_fixture["snap_value_pvc"], data_fixture["value_vs_class"], data_fixture["number_of_snapshots"], 
                               data_fixture["driver_data"]["image_name"], data_fixture["driver_data"]["id"], 
                               data_fixture["driver_data"]["pluginNodeSelector"])

    data_fixture["local_cg_snapshot_object"] = baseclass.Snapshot(data_fixture["cmd_values"]["kubeconfig_value"], 
                               data_fixture["cmd_values"]["test_namespace"], data_fixture["driver_data"]["keepobjects"],
                               data_fixture["cg_snap_value_pvc"], data_fixture["value_vs_class"], data_fixture["number_of_snapshots"], 
                               data_fixture["driver_data"]["image_name"], data_fixture["driver_data"]["id"], 
                               data_fixture["driver_data"]["pluginNodeSelector"])
                                    
    baseclass.filesetfunc.create_dir(data_fixture["driver_data"]["volDirBasePath"])
    

@pytest.fixture
def remote_cluster_fixture(data_fixture, new_namespace):

    if not("remote" in data_fixture["driver_data"]):
        LOGGER.error("remote data is not provided in CSO")
        assert False

    data_fixture["remote_data"] = inputfunc.get_remote_data(data_fixture["driver_data"])
    baseclass.filesetfunc.cred_check(data_fixture["remote_data"])
    baseclass.filesetfunc.set_data(data_fixture["remote_data"])

    data_fixture["remote_driver_object"] = baseclass.Driver(data_fixture["cmd_values"]["kubeconfig_value"], data_fixture["value_pvc"],
                                data_fixture["value_pod"], data_fixture["remote_data"]["id"], data_fixture["cmd_values"]["test_namespace"], 
                                data_fixture["driver_data"]["keepobjects"],data_fixture["driver_data"]["image_name"], 
                                data_fixture["driver_data"]["pluginNodeSelector"])
    
    data_fixture["remote_snapshot_object"] = baseclass.Snapshot(data_fixture["cmd_values"]["kubeconfig_value"], 
                                data_fixture["cmd_values"]["test_namespace"], data_fixture["driver_data"]["keepobjects"], 
                                data_fixture["snap_value_pvc"], data_fixture["value_vs_class"], data_fixture["number_of_snapshots"], 
                                data_fixture["driver_data"]["image_name"], data_fixture["remote_data"]["id"], 
                                data_fixture["driver_data"]["pluginNodeSelector"])

    data_fixture["remote_cg_snapshot_object"] = baseclass.Snapshot(data_fixture["cmd_values"]["kubeconfig_value"], 
                                data_fixture["cmd_values"]["test_namespace"], data_fixture["driver_data"]["keepobjects"],
                                data_fixture["cg_snap_value_pvc"], data_fixture["value_vs_class"], data_fixture["number_of_snapshots"], 
                                data_fixture["driver_data"]["image_name"], data_fixture["remote_data"]["id"], 
                                data_fixture["driver_data"]["pluginNodeSelector"])

    baseclass.filesetfunc.create_dir(data_fixture["remote_data"]["volDirBasePath"])

