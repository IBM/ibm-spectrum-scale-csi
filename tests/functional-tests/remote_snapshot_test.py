import copy
import logging
import pytest
import utils.fileset_functions as ff
from scale_operator import read_driver_data, Scaleoperator, check_ns_exists, \
    check_ds_exists, check_nodes_available, Scaleoperatorobject, Snapshot, check_key, \
    read_operator_data, check_pod_running
LOGGER = logging.getLogger()


@pytest.fixture(scope='session', autouse=True)
def values(request):
    global data, snapshot_object  # are required in every testcase
    kubeconfig_value = request.config.option.kubeconfig
    if kubeconfig_value is None:
        kubeconfig_value = "~/.kube/config"
    clusterconfig_value = request.config.option.clusterconfig
    if clusterconfig_value is None:
        clusterconfig_value = "../../operator/deploy/crds/csiscaleoperators.csi.ibm.com_cr.yaml"
    namespace_value = request.config.option.namespace
    if namespace_value is None:
        namespace_value = "ibm-spectrum-scale-csi-driver"
    data = read_driver_data(clusterconfig_value, namespace_value)
    operator_data = read_operator_data(clusterconfig_value, namespace_value)
    keep_objects = data["keepobjects"]
    if not(check_key(data, "remote")):
        LOGGER.error("remote data is not provided in cr file")
        assert False
    test_namespace = namespace_value

    ff.fileset_exists(data)
    operator = Scaleoperator(kubeconfig_value, namespace_value)
    condition = check_ns_exists(kubeconfig_value, namespace_value)
    if condition is True:
        check_ds_exists(kubeconfig_value, namespace_value)
    else:
        operator.create()
        operator.check()
        check_nodes_available(operator_data["pluginNodeSelector"], "pluginNodeSelector")
        check_nodes_available(
            operator_data["provisionerNodeSelector"], "provisionerNodeSelector")
        check_nodes_available(
            operator_data["attacherNodeSelector"], "attacherNodeSelector")
        operator_object = Scaleoperatorobject(operator_data, kubeconfig_value)
        operator_object.create()
        val = operator_object.check()
        if val is True:
            LOGGER.info("Operator custom object is deployed succesfully")
        else:
            LOGGER.error("Operator custom object is not deployed succesfully")
            assert False
    check_pod_running(kubeconfig_value, namespace_value, "ibm-spectrum-scale-csi-snapshotter-0")
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"},
                 {"access_modes": "ReadWriteOnce", "storage": "1Gi"}]
    value_vs_class = {"deletionPolicy": "Delete"}
    number_of_snapshots = 1
    remote_data = get_remote_data(data)
    snapshot_object = Snapshot(kubeconfig_value, test_namespace, keep_objects, remote_data, value_pvc, value_vs_class, number_of_snapshots)
    ff.create_dir(remote_data, remote_data["volDirBasePath"])
    yield
    if condition is False and not(keep_objects):
        operator_object.delete()
        operator.delete()
        if(ff.fileset_exists(data)):
            ff.delete_fileset(data)


def get_remote_data(data_passed):
    remote_data = copy.deepcopy(data_passed)
    remote_data["remoteFs_remote_name"] = ff.get_remoteFs_remotename(data)
    if remote_data["remoteFs_remote_name"] is None:
        LOGGER.error("Unable to get remoteFs , name on remote cluster")
        assert False

    remote_data["primaryFs"] = remote_data["remoteFs_remote_name"]
    remote_data["id"] = remote_data["remoteid"]
    remote_data["port"] = remote_data["remote-port"]
    for cluster in remote_data["clusters"]:
        if cluster["id"] == remote_data["remoteid"]:
            remote_data["guiHost"] = cluster["restApi"][0]["guiHost"]
            remote_sec_name = cluster["secrets"]
            remote_data["username"] = remote_data["remote-username"][remote_sec_name]
            remote_data["password"] = remote_data["remote-password"][remote_sec_name]

    remote_data["volDirBasePath"] = remote_data["r-volDirBasePath"]
    remote_data["parentFileset"] = remote_data["r-parentFileset"]
    remote_data["gid_name"] = remote_data["r-gid_name"]
    remote_data["uid_name"] = remote_data["r-uid_name"]
    remote_data["gid_number"] = remote_data["r-gid_number"]
    remote_data["uid_number"] = remote_data["r-uid_number"]
    remote_data["inodeLimit"] = remote_data["r-inodeLimit"]
    # for get_mount_point function
    remote_data["type_remote"] = {"username": data_passed["username"],
                                  "password": data_passed["password"],
                                  "port": data_passed["port"],
                                  "guiHost": data_passed["guiHost"]}

    return remote_data


def test_snapshot_dynamic_pass_1():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    snapshot_object.test_dynamic(value_sc)


@pytest.mark.skip
def test_snapshot_dynamic_pass_2():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    snapshot_object.test_dynamic(value_sc, value_vs_class={"deletionPolicy": "Retain"})


@pytest.mark.skip
def test_snapshot_dynamic_expected_fail_1():
    value_sc = {"volBackendFs": data["remoteFs"],
                "filesetType": "dependent", "clusterId": data["remoteid"]}
    snapshot_object.test_dynamic(value_sc)


@pytest.mark.skip
def test_snapshot_dynamic_expected_fail_2():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc)


def test_snapshot_dynamic_multiple_snapshots():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    snapshot_object.test_dynamic(value_sc, number_of_snapshots=3)


@pytest.mark.slow
def test_snapshot_dynamic_multiple_snapshots_256():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    snapshot_object.test_dynamic(value_sc, number_of_snapshots=256)


@pytest.mark.slow
def test_snapshot_dynamic_multiple_snapshots_257():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    snapshot_object.test_dynamic(value_sc, number_of_snapshots=257)


def test_snapshot_dynamic_pass_3():
    value_sc = {"volBackendFs": data["remoteFs"],
                "clusterId": data["remoteid"], "gid": data["r-gid_number"]}
    snapshot_object.test_dynamic(value_sc)


def test_snapshot_dynamic_pass_4():
    value_sc = {"volBackendFs": data["remoteFs"],
                "clusterId": data["remoteid"], "inodeLimit": data["r-inodeLimit"]}
    snapshot_object.test_dynamic(value_sc)


def test_snapshot_dynamic_pass_5():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"],
                "inodeLimit": data["r-inodeLimit"], "uid": data["r-uid_number"]}
    snapshot_object.test_dynamic(value_sc)


def test_snapshot_dynamic_pass_6():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"],
                "gid": data["r-gid_number"], "uid": data["r-uid_number"]}
    snapshot_object.test_dynamic(value_sc)


def test_snapshot_dynamic_pass_7():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"],
                "inodeLimit": data["r-inodeLimit"], "gid": data["r-gid_number"]}
    snapshot_object.test_dynamic(value_sc)


def test_snapshot_dynamic_pass_8():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"],
                "inodeLimit": data["r-inodeLimit"], "uid": data["r-uid_number"],
                "gid": data["r-gid_number"]}
    snapshot_object.test_dynamic(value_sc)


def test_snapshot_dynamic_pass_9():
    value_sc = {"volBackendFs": data["remoteFs"],
                "clusterId": data["remoteid"], "uid": data["r-uid_number"]}
    snapshot_object.test_dynamic(value_sc)


def test_snapshot_dynamic_pass_10():
    value_sc = {"volBackendFs": data["remoteFs"],
                "inodeLimit": data["r-inodeLimit"],
                "clusterId": data["remoteid"], "filesetType": "independent"}
    snapshot_object.test_dynamic(value_sc)


def test_snapshot_dynamic_pass_11():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r-gid_number"],
                "uid": data["r-uid_number"], "clusterId": data["remoteid"],
                "filesetType": "independent", "inodeLimit": data["r-inodeLimit"]}
    snapshot_object.test_dynamic(value_sc)


def test_snapshot_dynamic_pass_12():
    value_sc = {"clusterId": data["remoteid"], "gid": data["r-gid_number"],
                "uid": data["r-uid_number"], "volBackendFs": data["remoteFs"]}
    snapshot_object.test_dynamic(value_sc)


def test_snapshot_dynamic_pass_13():
    value_sc = {"clusterId": data["remoteid"], "uid": data["r-uid_number"],
                "inodeLimit": data["r-inodeLimit"],
                "volBackendFs": data["remoteFs"]}
    snapshot_object.test_dynamic(value_sc)


def test_snapshot_dynamic_pass_14():
    value_sc = {"clusterId": data["remoteid"], "gid": data["r-gid_number"],
                "inodeLimit": data["r-inodeLimit"],
                "volBackendFs": data["remoteFs"]}
    snapshot_object.test_dynamic(value_sc)


def test_snapshot_dynamic_pass_15():
    value_sc = {"clusterId": data["remoteid"], "volBackendFs": data["remoteFs"],
                "gid": data["r-gid_number"], "uid": data["r-uid_number"],
                "inodeLimit": data["r-inodeLimit"]}
    snapshot_object.test_dynamic(value_sc)
