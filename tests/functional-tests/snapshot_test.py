import logging
import pytest
from scale_operator import read_driver_data, Scaleoperator, check_ns_exists, get_cmd_values,\
    check_nodes_available, Scaleoperatorobject, Snapshot, read_operator_data, get_kubernetes_version
import utils.fileset_functions as ff
LOGGER = logging.getLogger()


@pytest.fixture(scope='session', autouse=True)
def values(request):
    global data, snapshot_object, kubeconfig_value  # are required in every testcase
    kubeconfig_value,clusterconfig_value,namespace_value,runslow_val = get_cmd_values(request)

    data = read_driver_data(clusterconfig_value, namespace_value)
    operator_data = read_operator_data(clusterconfig_value, namespace_value)
    keep_objects = data["keepobjects"]
    test_namespace = namespace_value
    ff.cred_check(data)
    ff.set_data(data)

    operator = Scaleoperator(kubeconfig_value, namespace_value)
    operator_object = Scaleoperatorobject(operator_data, kubeconfig_value)
    condition = check_ns_exists(kubeconfig_value, namespace_value)
    if condition is True:
        if not(operator_object.check()):
            LOGGER.error("Operator custom object is not deployed succesfully")
            assert False
    else:
        operator.create()
        operator.check()
        check_nodes_available(operator_data["pluginNodeSelector"], "pluginNodeSelector")
        check_nodes_available(
            operator_data["provisionerNodeSelector"], "provisionerNodeSelector")
        check_nodes_available(
            operator_data["attacherNodeSelector"], "attacherNodeSelector")
        operator_object.create()
        val = operator_object.check()
        if val is True:
            LOGGER.info("Operator custom object is deployed succesfully")
        else:
            LOGGER.error("Operator custom object is not deployed succesfully")
            assert False
    if runslow_val:
        value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"},
                     {"access_modes": "ReadWriteOnce", "storage": "1Gi"}]
    else:
        value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]

    value_vs_class = {"deletionPolicy": "Delete"}
    number_of_snapshots = 1
    snapshot_object = Snapshot(kubeconfig_value, test_namespace, keep_objects, value_pvc, value_vs_class, number_of_snapshots, data["image_name"])
    if not(data["volBackendFs"] == ""):
        data["primaryFs"] = data["volBackendFs"]
    ff.create_dir(data["volDirBasePath"])

    yield
    if condition is False and not(keep_objects):
        operator_object.delete()
        operator.delete()
        if(ff.fileset_exists(data)):
            ff.delete_fileset(data)


def test_get_version():
    ff.get_scale_version(data)
    get_kubernetes_version(kubeconfig_value)


def test_snapshot_dynamic_pass_1():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


@pytest.mark.skip
def test_snapshot_dynamic_pass_2():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, value_vs_class={"deletionPolicy": "Retain"})


@pytest.mark.skip
def test_snapshot_dynamic_expected_fail_1():
    value_sc = {"volBackendFs": data["primaryFs"],
                "filesetType": "dependent", "clusterId": data["id"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


@pytest.mark.skip
def test_snapshot_dynamic_expected_fail_2():
    value_sc = {"volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_multiple_snapshots():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, number_of_snapshots=3)


@pytest.mark.slow
def test_snapshot_dynamic_multiple_snapshots_256():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, number_of_snapshots=256)


@pytest.mark.slow
def test_snapshot_dynamic_multiple_snapshots_257():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, number_of_snapshots=257)


def test_snapshot_dynamic_pass_3():
    value_sc = {"volBackendFs": data["primaryFs"],
                "clusterId": data["id"], "gid": data["gid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_4():
    value_sc = {"volBackendFs": data["primaryFs"],
                "clusterId": data["id"], "inodeLimit": data["inodeLimit"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_5():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"],
                "inodeLimit": data["inodeLimit"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_6():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_7():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"],
                "inodeLimit": data["inodeLimit"], "gid": data["gid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_8():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"],
                "inodeLimit": data["inodeLimit"], "uid": data["uid_number"],
                "gid": data["gid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_9():
    value_sc = {"volBackendFs": data["primaryFs"],
                "clusterId": data["id"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_10():
    value_sc = {"volBackendFs": data["primaryFs"],
                "inodeLimit": data["inodeLimit"],
                "clusterId": data["id"], "filesetType": "independent"}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_11():
    value_sc = {"volBackendFs": data["primaryFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "clusterId": data["id"],
                "filesetType": "independent", "inodeLimit": data["inodeLimit"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_12():
    value_sc = {"volBackendFs": data["primaryFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_13():
    value_sc = {"uid": data["uid_number"], "volBackendFs": data["primaryFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_14():
    value_sc = {"gid": data["gid_number"], "volBackendFs": data["primaryFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_15():
    value_sc = {"inodeLimit": data["inodeLimit"],
                "volBackendFs": data["primaryFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_16():
    value_sc = {"volBackendFs": data["primaryFs"], "uid": data["uid_number"],
                "gid": data["gid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_17():
    value_sc = {"volBackendFs": data["primaryFs"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_18():
    value_sc = {"volBackendFs": data["primaryFs"], "gid": data["gid_number"],
                "inodeLimit": data["inodeLimit"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_19():
    value_sc = {"clusterId": data["id"], "gid": data["gid_number"],
                "uid": data["uid_number"], "volBackendFs": data["primaryFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_20():
    value_sc = {"clusterId": data["id"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"],
                "volBackendFs": data["primaryFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_21():
    value_sc = {"clusterId": data["id"], "gid": data["gid_number"],
                "inodeLimit": data["inodeLimit"],
                "volBackendFs": data["primaryFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_22():
    value_sc = {"gid": data["gid_number"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"],
                "volBackendFs": data["primaryFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_23():
    value_sc = {"clusterId": data["id"], "volBackendFs": data["primaryFs"],
                "gid": data["gid_number"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_24():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    snapshot_object.test_dynamic(value_sc, test_restore=False)


def test_snapshot_dynamic_pass_25():
    value_sc = {"volBackendFs": data["primaryFs"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"]}
    snapshot_object.test_dynamic(value_sc, test_restore=False)


def test_snapshot_dynamic_pass_26():
    value_sc = {"gid": data["gid_number"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"],
                "volBackendFs": data["primaryFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=False)
