from datetime import datetime
import pytest
from py.xml import html

input_params = {

    "username":"",      #Pass username for primary cluster SpectrumScale GUI (in plain text)
    "password":"",     #Pass password for primary cluster SpectrumScale GUI (in plain text) 
    "port":"443",
    
    "remote-username": {"guisecretremote":""},  # eg. { "secret_name" : "RestAPI_username" }
    "remote-password": {"guisecretremote":""},  # eg. { "secret_name" : "RestAPI_password" }
    "remote-port": "443",

    "cacert_path" : " ",                #Path of cacert file for primary fileset cluster API cert
    "remote_cacert_path":{"remoteconf1":""},           #Path of cacert file for remote cluster API cert eg. { "cacert_name" : "cacert_path" }

    "number_of_parallel_pvc":10,

    "remoteFs":"",            # Must provide remote filesystem name on Primary cluster in case of remote_test.py
    "remoteid":"",            # Must provide remote cluster id in case of remote_test.py

    "volBackendFs":"",            # OPTIONAL : Should be given in case of driver_test.py and want to use filesytem other than the primartFs

    "keepobjects":False,

    "volDirBasePath":"LW", 
    "parentFileset":"root",
    "gid_name":"nobody",
    "uid_name":"nobody",
    "gid_number":"99",
    "uid_number":"99",
    "inodeLimit":"1024",
 
    "r-volDirBasePath":"LW",
    "r-parentFileset":"root",
    "r-gid_name":"nobody",
    "r-uid_name":"nobody",
    "r-gid_number":"99",
    "r-uid_number":"99",
    "r-inodeLimit":"1024"
}


def pytest_addoption(parser):
    parser.addoption("--kubeconfig", action="store")
    parser.addoption("--clusterconfig", action="store")
    parser.addoption("--namespace", action="store")
    parser.addoption("--runslow", action="store_true",help="run slow tests")


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

