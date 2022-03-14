import logging
import pytest
import ibm_spectrum_scale_csi.base_class as baseclass
import ibm_spectrum_scale_csi.common_utils.input_data_functions as inputfunc

LOGGER = logging.getLogger()
pytestmark = [pytest.mark.volumesnapshot, pytest.mark.localcluster, pytest.mark.cg]

@pytest.fixture(scope='session', autouse=True)
def values(request, check_csi_operator):
    global data, snapshot_object, kubeconfig_value  # are required in every testcase
    cmd_values = inputfunc.get_pytest_cmd_values(request)
    kubeconfig_value = cmd_values["kubeconfig_value"]
    data = inputfunc.read_driver_data(cmd_values)

    keep_objects = data["keepobjects"]
    if not(data["volBackendFs"] == ""):
        data["primaryFs"] = data["volBackendFs"]

    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_vs_class = {"deletionPolicy": "Delete"}
    number_of_snapshots = 1
    snapshot_object = baseclass.Snapshot(kubeconfig_value, cmd_values["test_namespace"], keep_objects, value_pvc, value_vs_class,
                                       number_of_snapshots, data["image_name"], data["id"], data["pluginNodeSelector"])


@pytest.mark.regression
def test_get_version():
    LOGGER.info("Cluster Details:")
    LOGGER.info("----------------")
    baseclass.filesetfunc.get_scale_version(data)
    baseclass.kubeobjectfunc.get_kubernetes_version(kubeconfig_value)
    baseclass.kubeobjectfunc.get_operator_image()
    baseclass.kubeobjectfunc.get_driver_image()


@pytest.mark.regression
def test_snapshot_cg_pass_1():
    value_sc = {"volBackendFs": data["primaryFs"], "version": "2", "consistencyGroup": None}
    value_vs_class={"deletionPolicy": "Delete", "snapWindow": "15"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, value_vs_class=value_vs_class)


def test_snapshot_cg_pass_2():
    value_sc = {"volBackendFs": data["primaryFs"], "version": "2", "consistencyGroup": "local-test_snapshot_cg_pass_2-cg"}
    value_vs_class={"deletionPolicy": "Delete"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, value_vs_class=value_vs_class, number_of_snapshots=10)


@pytest.mark.regression
def test_snapshot_cg_pass_3():
    value_sc = {"volBackendFs": data["primaryFs"], "version": "2"}
    value_vs_class={"deletionPolicy": "Delete", "snapWindow": "2"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}] * 3
    snapshot_object.test_dynamic(value_sc, test_restore=True, value_vs_class=value_vs_class, value_pvc=value_pvc, number_of_snapshots=3)


def test_snapshot_cg_tier():
    value_sc = {"volBackendFs": data["primaryFs"], "version": "2", "tier": data["tier"]}
    value_vs_class={"deletionPolicy": "Delete", "snapWindow": "15"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, value_vs_class=value_vs_class)


def test_snapshot_cg_compression():
    value_sc = {"volBackendFs": data["primaryFs"], "version": "2", "compression": "true"}
    value_vs_class={"deletionPolicy": "Delete", "snapWindow": "15"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, value_vs_class=value_vs_class)


def test_snapshot_cg_compression_tier():
    value_sc = {"volBackendFs": data["primaryFs"], "version": "2", "tier": data["tier"], "compression": "true"}
    value_vs_class={"deletionPolicy": "Delete", "snapWindow": "15"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, value_vs_class=value_vs_class)
