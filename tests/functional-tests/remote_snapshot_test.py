import copy
import logging
import pytest
import utils.fileset_functions as ff
from scale_operator import Snapshot, read_driver_data
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
    keep_objects = data["keepobjects"]
    test_namespace = namespace_value
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"},
                 {"access_modes": "ReadWriteOnce", "storage": "1Gi"}]
    value_vs_class = {"deletionPolicy": "Delete"}
    number_of_snapshots = 1
    remote_data = get_remote_data(data)
    snapshot_object = Snapshot(kubeconfig_value, test_namespace, keep_objects, remote_data, value_pvc, value_vs_class, number_of_snapshots)


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
def test_snapshot_dynamic_expected_fail_1():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    snapshot_object.test_dynamic(value_sc, value_vs_class={"deletionPolicy": "Retain"})


@pytest.mark.skip
def test_snapshot_dynamic_expected_fail_2():
    value_sc = {"volBackendFs": data["remoteFs"],
                "filesetType": "dependent", "clusterId": data["remoteid"]}
    snapshot_object.test_dynamic(value_sc)


@pytest.mark.skip
def test_snapshot_dynamic_expected_fail_3():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc)


def test_snapshot_dynamic_multiple_snapshots():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    snapshot_object.test_dynamic(value_sc, number_of_snapshots=3)


def test_snapshot_dynamic_pass_2():
    value_sc = {"volBackendFs": data["remoteFs"],
                "clusterId": data["remoteid"], "uid": data["r-uid_number"]}
    snapshot_object.test_dynamic(value_sc)


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


def test_snapshot_dynamic_pass_19():
    value_sc = {"volBackendFs": data["remoteFs"],
                "inodeLimit": data["r-inodeLimit"],
                "clusterId": data["remoteid"], "filesetType": "independent"}
    snapshot_object.test_dynamic(value_sc)


def test_snapshot_dynamic_pass_29():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r-gid_number"],
                "uid": data["r-uid_number"], "clusterId": data["remoteid"],
                "filesetType": "independent", "inodeLimit": data["r-inodeLimit"]}
    snapshot_object.test_dynamic(value_sc)


def test_snapshot_dynamic_pass_116():
    value_sc = {"clusterId": data["remoteid"], "gid": data["r-gid_number"],
                "uid": data["r-uid_number"], "volBackendFs": data["remoteFs"]}
    snapshot_object.test_dynamic(value_sc)


def test_snapshot_dynamic_pass_118():
    value_sc = {"clusterId": data["remoteid"], "uid": data["r-uid_number"],
                "inodeLimit": data["r-inodeLimit"],
                "volBackendFs": data["remoteFs"]}
    snapshot_object.test_dynamic(value_sc)


def test_snapshot_dynamic_pass_120():
    value_sc = {"clusterId": data["remoteid"], "gid": data["r-gid_number"],
                "inodeLimit": data["r-inodeLimit"],
                "volBackendFs": data["remoteFs"]}
    snapshot_object.test_dynamic(value_sc)


def test_snapshot_dynamic_pass_188():
    value_sc = {"clusterId": data["remoteid"], "volBackendFs": data["remoteFs"],
                "gid": data["r-gid_number"], "uid": data["r-uid_number"],
                "inodeLimit": data["r-inodeLimit"]}
    snapshot_object.test_dynamic(value_sc)
