import pytest
from scale_operator import Snapshot, read_driver_data


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
    snapshot_object = Snapshot(kubeconfig_value, test_namespace, keep_objects, data, value_pvc, value_vs_class, number_of_snapshots)
    if not(data["volBackendFs"] == ""):
        data["primaryFs"] = data["volBackendFs"]


def test_snapshot_dynamic_pass_1():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    snapshot_object.test_dynamic(value_sc)


def test_snapshot_dynamic_expected_fail_1():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    snapshot_object.test_dynamic(value_sc, value_vs_class={"deletionPolicy": "Retain"})


def test_snapshot_dynamic_expected_fail_2():
    value_sc = {"volBackendFs": data["primaryFs"],
                "filesetType": "dependent", "clusterId": data["id"]}
    snapshot_object.test_dynamic(value_sc)


def test_snapshot_dynamic_expected_fail_3():
    value_sc = {"volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc)


def test_snapshot_dynamic_multiple_snapshots():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    snapshot_object.test_dynamic(value_sc, number_of_snapshots=3)


def test_snapshot_dynamic_pass_2():
    value_sc = {"volBackendFs": data["primaryFs"],
                "clusterId": data["id"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc)


def test_snapshot_dynamic_pass_3():
    value_sc = {"volBackendFs": data["primaryFs"],
                "clusterId": data["id"], "gid": data["gid_number"]}
    snapshot_object.test_dynamic(value_sc)


def test_snapshot_dynamic_pass_4():
    value_sc = {"volBackendFs": data["primaryFs"],
                "clusterId": data["id"], "inodeLimit": data["inodeLimit"]}
    snapshot_object.test_dynamic(value_sc)


def test_snapshot_dynamic_pass_5():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"],
                "inodeLimit": data["inodeLimit"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc)


def test_snapshot_dynamic_pass_6():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc)


def test_snapshot_dynamic_pass_7():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"],
                "inodeLimit": data["inodeLimit"], "gid": data["gid_number"]}
    snapshot_object.test_dynamic(value_sc)


def test_snapshot_dynamic_pass_8():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"],
                "inodeLimit": data["inodeLimit"], "uid": data["uid_number"],
                "gid": data["gid_number"]}
    snapshot_object.test_dynamic(value_sc)


def test_snapshot_dynamic_pass_19():
    value_sc = {"volBackendFs": data["primaryFs"],
                "inodeLimit": data["inodeLimit"],
                "clusterId": data["id"], "filesetType": "independent"}
    snapshot_object.test_dynamic(value_sc)


def test_snapshot_dynamic_pass_29():
    value_sc = {"volBackendFs": data["primaryFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "clusterId": data["id"],
                "filesetType": "independent", "inodeLimit": data["inodeLimit"]}
    snapshot_object.test_dynamic(value_sc)


def test_snapshot_dynamic_pass_30():
    value_sc = {"volBackendFs": data["primaryFs"]}
    snapshot_object.test_dynamic(value_sc)


def test_snapshot_dynamic_pass_38():
    value_sc = {"uid": data["uid_number"], "volBackendFs": data["primaryFs"]}
    snapshot_object.test_dynamic(value_sc)


def test_snapshot_dynamic_pass_39():
    value_sc = {"gid": data["gid_number"], "volBackendFs": data["primaryFs"]}
    snapshot_object.test_dynamic(value_sc)


def test_snapshot_dynamic_pass_42():
    value_sc = {"inodeLimit": data["inodeLimit"],
                "volBackendFs": data["primaryFs"]}
    snapshot_object.test_dynamic(value_sc)


def test_snapshot_dynamic_pass_68():
    value_sc = {"volBackendFs": data["primaryFs"], "uid": data["uid_number"],
                "gid": data["gid_number"]}
    snapshot_object.test_dynamic(value_sc)


def test_snapshot_dynamic_pass_71():
    value_sc = {"volBackendFs": data["primaryFs"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"]}
    snapshot_object.test_dynamic(value_sc)


def test_snapshot_dynamic_pass_74():
    value_sc = {"volBackendFs": data["primaryFs"], "gid": data["gid_number"],
                "inodeLimit": data["inodeLimit"]}
    snapshot_object.test_dynamic(value_sc)


def test_snapshot_dynamic_pass_116():
    value_sc = {"clusterId": data["id"], "gid": data["gid_number"],
                "uid": data["uid_number"], "volBackendFs": data["primaryFs"]}
    snapshot_object.test_dynamic(value_sc)


def test_snapshot_dynamic_pass_118():
    value_sc = {"clusterId": data["id"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"],
                "volBackendFs": data["primaryFs"]}
    snapshot_object.test_dynamic(value_sc)


def test_snapshot_dynamic_pass_120():
    value_sc = {"clusterId": data["id"], "gid": data["gid_number"],
                "inodeLimit": data["inodeLimit"],
                "volBackendFs": data["primaryFs"]}
    snapshot_object.test_dynamic(value_sc)


def test_snapshot_dynamic_pass_135():
    value_sc = {"gid": data["gid_number"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"],
                "volBackendFs": data["primaryFs"]}
    snapshot_object.test_dynamic(value_sc)


def test_snapshot_dynamic_pass_188():
    value_sc = {"clusterId": data["id"], "volBackendFs": data["primaryFs"],
                "gid": data["gid_number"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"]}
    snapshot_object.test_dynamic(value_sc)
