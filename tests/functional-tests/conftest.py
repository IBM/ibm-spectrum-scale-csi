from datetime import datetime
import pytest
from py.xml import html

input_params = {

    "username":"",     #Pass username for primary cluster SpectrumScale GUI (in plain text)
    "password":"",     #Pass password for primary cluster SpectrumScale GUI (in plain text) 
    "port":"443",
    
    "remote-username": "",
    "remote-password": "",
    "remote-port": "443",

    "cacert_path" : "",                #Path of cacert file for primary fileset cluster API cert
    "remote_cacert_path":"",           #Path of cacert file for remote cluster API cert

    "number_of_parallel_pvc":10,

    "remoteFs":"",                # Must provide remote filesystem name on Primary cluster in case of remote_test.py

    "volBackendFs":"",            # OPTIONAL : Should be given in case of driver_test.py and want to use filesytem other than the primartFs
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
