import logging
import pytest
import ibm_spectrum_scale_csi.base_class as baseclass
import ibm_spectrum_scale_csi.common_utils.input_data_functions as inputfunc
LOGGER = logging.getLogger()
pytestmark = [pytest.mark.volumecloning, pytest.mark.localcluster]


@pytest.fixture(autouse=True)
def values(data_fixture, check_csi_operator, local_cluster_fixture):
    global data, driver_object, snapshot_object, kubeconfig_value  # are required in every testcase
    data = data_fixture["driver_data"]
    kubeconfig_value = data_fixture["cmd_values"]["kubeconfig_value"]
    driver_object = data_fixture["local_driver_object"]
    snapshot_object = data_fixture["local_snapshot_object"]


#: Testcase that are expected to pass:
@pytest.mark.regression
def test_get_version():
    LOGGER.info("Cluster Details:")
    LOGGER.info("----------------")
    baseclass.filesetfunc.get_scale_version(data)
    baseclass.kubeobjectfunc.get_kubernetes_version(kubeconfig_value)
    baseclass.kubeobjectfunc.get_operator_image()
    baseclass.kubeobjectfunc.get_driver_image()

def test_driver_volume_cloning_pass_1():
    value_sc = {"volBackendFs": data["localFs"], "clusterId": data["id"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}, {
        "access_modes": "ReadWriteOnce", "storage": "1Gi"}]}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_pass_2():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "clusterId": data["id"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}, {
        "access_modes": "ReadWriteOnce", "storage": "1Gi"}]}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_pass_3():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}, {
        "access_modes": "ReadWriteOnce", "storage": "1Gi"}]}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_pass_4():
    value_sc = {"volBackendFs": data["localFs"], "clusterId": data["id"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "11Gi"}, {
        "access_modes": "ReadWriteOnce", "storage": "8Gi"}]}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_multiple_clones():
    value_sc = {"volBackendFs": data["localFs"], "clusterId": data["id"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}],
                          "number_of_clones": 5}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_chain_clones():
    value_sc = {"volBackendFs": data["localFs"], "clusterId": data["id"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}],
                          "clone_chain": 1}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_expand_before():
    value_sc = {"volBackendFs": data["localFs"],
                "clusterId": data["id"], "allow_volume_expansion": True}
    value_pvc = [{"access_modes": "ReadWriteMany",
                  "storage": "9Gi", "volume_expansion_storage": ["11Gi"]}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "11Gi"},
                                        {"access_modes": "ReadWriteMany", "storage": "9Gi", "reason": "new PVC request must be greater than or equal in size"}]}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_with_subpath():
    value_sc = {"volBackendFs": data["localFs"], "clusterId": data["id"], "permissions": "777",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}, {
        "access_modes": "ReadWriteOnce", "storage": "1Gi"}]}
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False",
                  "sub_path": ["sub_path_mnt", "sub_path_mnt_2", "sub_path_mnt3"], "volumemount_readonly":[False, False, True], "runAsUser": "2000", "runAsGroup": "5000"}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_pod_passed=value_pod, value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_fail_1():
    value_sc = {"volBackendFs": data["localFs"], "clusterId": data["id"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi", "reason": "same storage class for cloning"}],
                          "clone_sc": {"volBackendFs": data["localFs"], "clusterId": data["id"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_fail_2():
    value_sc = {"volBackendFs": data["localFs"], "clusterId": data["id"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi", "reason": "same storage class for cloning"}],
                          "clone_sc": {"volBackendFs": data["localFs"], "filesetType": "dependent",
                                       "clusterId": data["id"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_fail_3():
    value_sc = {"volBackendFs": data["localFs"], "clusterId": data["id"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi", "reason": "same storage class for cloning"}],
                          "clone_sc": {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_fail_4():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "clusterId": data["id"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi", "reason": "same storage class for cloning"}],
                          "clone_sc": {"volBackendFs": data["localFs"], "clusterId": data["id"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_fail_5():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "clusterId": data["id"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc":  [{"access_modes": "ReadWriteMany", "storage": "1Gi", "reason": "same storage class for cloning"}],
                          "clone_sc": {"volBackendFs": data["localFs"], "filesetType": "dependent",
                                       "clusterId": data["id"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_fail_6():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "clusterId": data["id"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi", "reason": "same storage class for cloning"}],
                          "clone_sc": {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_fail_7():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi", "reason": "same storage class for cloning"}],
                          "clone_sc": {"volBackendFs": data["localFs"], "clusterId": data["id"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_fail_8():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi", "reason": "same storage class for cloning"}],
                          "clone_sc": {"volBackendFs": data["localFs"], "filesetType": "dependent",
                                       "clusterId": data["id"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_fail_9():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi", "reason": "same storage class for cloning"}],
                          "clone_sc": {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_fail_10():
    value_sc = {"volBackendFs": data["localFs"], "clusterId": data["id"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "5Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi",
                                         "reason": "new PVC request must be greater than or equal in size"}]}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_cg_cloning_1():
    value_sc = {"volBackendFs": data["localFs"], "version": "2"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}] * 2
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}, {
        "access_modes": "ReadWriteOnce", "storage": "1Gi"}]}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.regression
def test_driver_cg_cloning_2():
    value_sc = {"volBackendFs": data["localFs"], "version": "2"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}] * 2
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "11Gi"}, {
        "access_modes": "ReadWriteOnce", "storage": "8Gi"}]}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_snapshot_dynamic_volume_cloning_1():
    value_sc = {"volBackendFs": data["localFs"], "clusterId": data["id"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}, {
        "access_modes": "ReadWriteOnce", "storage": "1Gi"}]}
    snapshot_object.test_dynamic(value_sc, test_restore=True,
                                 value_pvc=value_pvc, value_clone_passed=value_clone_passed)