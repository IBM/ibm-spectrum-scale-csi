from datetime import datetime
import pytest
from py.xml import html


def pytest_addoption(parser):
    parser.addoption("--kubeconfig", action="store")
    parser.addoption("--clusterconfig", action="store")
    parser.addoption("--namespace", action="store")
    parser.addoption("--runslow", action="store_true", help="run slow tests")


def pytest_html_results_table_header(cells):
    cells.pop()


@pytest.hookimpl(hookwrapper=True)
def pytest_runtest_makereport(item, call):
    outcome = yield
    report = outcome.get_result()
    report.description = str(item.function.__doc__)


now = datetime.now()
dt_string = now.strftime("%d-%m-%Y-%H-%M-%S")
default_html_path = 'data/report-'+dt_string+'.html'


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
