import pytest
from py.xml import html
from datetime import datetime

input_params = {

    "username":"Y3NpYWRtaW4=",
    "password":"UGFzc3cwcmQx",
    "port":"443",
    
    "remote-username": "Y3NpYWRtaW4=",
    "remote-password": "YWRtaW4wMDE=",
    "remote-port": "443",

    "cacert_path":"/root/spectrumscale-trusted-ca.crt",
    "remote_cacert_path":"/root/remote-spectrumscale-trusted-ca.crt",

    "number_of_parallel_pvc":10,

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
    cells.insert(2, html.th('Description'))
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
