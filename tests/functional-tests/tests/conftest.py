from datetime import datetime
import pytest
from py.xml import html
import logging
import ibm_spectrum_scale_csi.base_class as baseclass
import ibm_spectrum_scale_csi.common_utils.input_data_functions as inputfunc

LOGGER = logging.getLogger()

def pytest_addoption(parser):
    parser.addoption("--kubeconfig", action="store")
    parser.addoption("--clusterconfig", action="store")
    parser.addoption("--testnamespace", action="store")
    parser.addoption("--operatornamespace", action="store")
    parser.addoption("--runslow", action="store_true", help="run slow tests")
    parser.addoption("--operatoryaml", action="store")



def pytest_html_results_table_header(cells):
    cells.pop()


@pytest.hookimpl(hookwrapper=True)
def pytest_runtest_makereport(item, call):
    outcome = yield
    report = outcome.get_result()
    report.description = str(item.function.__doc__)


now = datetime.now()
dt_string = now.strftime("%d-%m-%Y-%H-%M-%S")
default_html_path = 'report-'+dt_string+'.html'


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
def check_csi_operator(request):
    kubeconfig_value, clusterconfig_value, operator_namespace, test_namespace, runslow_val, operator_yaml = inputfunc.get_cmd_values(request)

    data = inputfunc.read_driver_data(clusterconfig_value, test_namespace, operator_namespace, kubeconfig_value)
    operator_data = inputfunc.read_operator_data(clusterconfig_value, operator_namespace, kubeconfig_value)
    keep_objects = data["keepobjects"]
    if not(data["volBackendFs"] == ""):
        data["primaryFs"] = data["volBackendFs"]

    baseclass.filesetfunc.cred_check(data)
    baseclass.filesetfunc.set_data(data)
    operator = baseclass.Scaleoperator(kubeconfig_value, operator_namespace, operator_yaml)
    operator_object = baseclass.Scaleoperatorobject(operator_data, kubeconfig_value)
    condition = baseclass.kubeobjectfunc.check_ns_exists(kubeconfig_value, operator_namespace)
    if condition is True:
        if not(operator_object.check(data["csiscaleoperator_name"])):
            LOGGER.error("Operator custom object is not deployed succesfully")
            assert False
    else:
        operator.create()
        operator.check()
        baseclass.kubeobjectfunc.check_nodes_available(operator_data["pluginNodeSelector"], "pluginNodeSelector")
        baseclass.kubeobjectfunc.check_nodes_available(
            operator_data["provisionerNodeSelector"], "provisionerNodeSelector")
        baseclass.kubeobjectfunc.check_nodes_available(
            operator_data["attacherNodeSelector"], "attacherNodeSelector")
        operator_object.create()
        val = operator_object.check()
        if val is True:
            LOGGER.info("Operator custom object is deployed succesfully")
        else:
            LOGGER.error("Operator custom object is not deployed succesfully")
            assert False

    baseclass.filesetfunc.create_dir(data["volDirBasePath"])

    # driver_object.create_test_ns(kubeconfig_value)
    yield
    # driver_object.delete_test_ns(kubeconfig_value)
    # delete_dir(data, data["volDirBasePath"])
    if condition is False and not(keep_objects):
        operator_object.delete()
        operator.delete()
        if(baseclass.filesetfunc.fileset_exists(data)):
            baseclass.filesetfunc.delete_fileset(data)
