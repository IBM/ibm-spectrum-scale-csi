import logging
import pytest
import ibm_spectrum_scale_csi.base_class as baseclass
import ibm_spectrum_scale_csi.common_utils.input_data_functions as inputfunc
LOGGER = logging.getLogger()
pytestmark = [pytest.mark.volumecloning, pytest.mark.remotecluster]


@pytest.fixture(autouse=True)
def values(data_fixture, check_csi_operator, remote_cluster_fixture):
    global data, remote_data, driver_object, kubeconfig_value  # are required in every testcase
    data = data_fixture["driver_data"]
    remote_data = data_fixture["remote_data"]
    kubeconfig_value = data_fixture["cmd_values"]["kubeconfig_value"]
    driver_object = data_fixture["remote_driver_object"]


#: Testcase that are expected to pass:


@pytest.mark.regression
def test_get_version():
    LOGGER.info("Remote Cluster Details:")
    LOGGER.info("-----------------------")
    baseclass.filesetfunc.get_scale_version(remote_data)
    LOGGER.info("Local Cluster Details:")
    LOGGER.info("-----------------------")
    baseclass.filesetfunc.get_scale_version(data)
    baseclass.kubeobjectfunc.get_kubernetes_version(kubeconfig_value)
    baseclass.kubeobjectfunc.get_operator_image()
    baseclass.kubeobjectfunc.get_driver_image()


def test_driver_volume_cloning_pass_1():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}, {
        "access_modes": "ReadWriteOnce", "storage": "1Gi"}]}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_pass_2():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "clusterId": data["remoteid"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}, {
        "access_modes": "ReadWriteOnce", "storage": "1Gi"}]}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_pass_3():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "11Gi"}, {
        "access_modes": "ReadWriteOnce", "storage": "8Gi"}]}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_multiple_clones():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}],
                          "number_of_clones": 5}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_chain_clones():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}],
                          "clone_chain": 1}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_expand_before():
    value_sc = {"volBackendFs": data["remoteFs"],
                "clusterId": data["remoteid"], "allow_volume_expansion": True}
    value_pvc = [{"access_modes": "ReadWriteMany",
                  "storage": "9Gi", "volume_expansion_storage": ["11Gi"]}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "11Gi"},
                                        {"access_modes": "ReadWriteMany", "storage": "9Gi", "reason": "new PVC request must be greater than or equal in size"}]}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_with_subpath():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"], "permissions": "777",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}, {
        "access_modes": "ReadWriteOnce", "storage": "1Gi"}]}
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False",
                  "sub_path": ["sub_path_mnt", "sub_path_mnt_2", "sub_path_mnt3"], "volumemount_readonly":[False, False, True], "runAsUser": "2000", "runAsGroup": "5000"}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_pod_passed=value_pod, value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_fail_1():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "5Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi",
                                         "reason": "new PVC request must be greater than or equal in size"}]}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_fail_2():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi", "reason": "Volume cloning for directories for remote file system is not supported"},
                                        {"access_modes": "ReadWriteOnce", "storage": "1Gi", "reason": "Volume cloning for directories for remote file system is not supported"}]}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_cg_cloning_1():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}] * 2
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}, {
        "access_modes": "ReadWriteOnce", "storage": "1Gi"}]}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_cg_cloning_2():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}] * 2
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "11Gi"}, {
        "access_modes": "ReadWriteOnce", "storage": "8Gi"}]}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


'''
def test_driver_volume_cloning_Dependent_1_to_Dependent_1():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_1_to_Independent_1():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [
        {"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {"volBackendFs": data["remoteFs"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_1_to_LW_1():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_1_to_Version2_1():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_1_to_Dependent_2():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_1_to_Independent_2():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_1_to_LW_2():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_1_to_Version2_2():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_1_to_Dependent_3():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_1_to_Independent_3():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_1_to_LW_3():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_1_to_Version2_3():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_1_to_Dependent_4():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_1_to_Independent_4():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_1_to_LW_4():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_1_to_Version2_4():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_1_to_Dependent_5():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_1_to_Independent_5():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_1_to_LW_5():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_1_to_Version2_5():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_1_to_Dependent_1():
    value_sc = {"volBackendFs": data["remoteFs"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_1_to_Independent_1():
    value_sc = {"volBackendFs": data["remoteFs"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [
        {"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {"volBackendFs": data["remoteFs"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_1_to_LW_1():
    value_sc = {"volBackendFs": data["remoteFs"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_1_to_Version2_1():
    value_sc = {"volBackendFs": data["remoteFs"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_1_to_Dependent_2():
    value_sc = {"volBackendFs": data["remoteFs"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_1_to_Independent_2():
    value_sc = {"volBackendFs": data["remoteFs"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_1_to_LW_2():
    value_sc = {"volBackendFs": data["remoteFs"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_1_to_Version2_2():
    value_sc = {"volBackendFs": data["remoteFs"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_1_to_Dependent_3():
    value_sc = {"volBackendFs": data["remoteFs"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_1_to_Independent_3():
    value_sc = {"volBackendFs": data["remoteFs"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_1_to_LW_3():
    value_sc = {"volBackendFs": data["remoteFs"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_1_to_Version2_3():
    value_sc = {"volBackendFs": data["remoteFs"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_1_to_Dependent_4():
    value_sc = {"volBackendFs": data["remoteFs"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_1_to_Independent_4():
    value_sc = {"volBackendFs": data["remoteFs"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_1_to_LW_4():
    value_sc = {"volBackendFs": data["remoteFs"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_1_to_Version2_4():
    value_sc = {"volBackendFs": data["remoteFs"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_1_to_Dependent_5():
    value_sc = {"volBackendFs": data["remoteFs"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_1_to_Independent_5():
    value_sc = {"volBackendFs": data["remoteFs"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_1_to_LW_5():
    value_sc = {"volBackendFs": data["remoteFs"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_1_to_Version2_5():
    value_sc = {"volBackendFs": data["remoteFs"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_1_to_Dependent_1():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_1_to_Independent_1():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [
        {"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {"volBackendFs": data["remoteFs"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_1_to_LW_1():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_1_to_Version2_1():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_1_to_Dependent_2():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_1_to_Independent_2():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_1_to_LW_2():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_1_to_Version2_2():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_1_to_Dependent_3():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_1_to_Independent_3():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_1_to_LW_3():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_1_to_Version2_3():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_1_to_Dependent_4():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_1_to_Independent_4():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_1_to_LW_4():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_1_to_Version2_4():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_1_to_Dependent_5():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_1_to_Independent_5():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_1_to_LW_5():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_1_to_Version2_5():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_1_to_Dependent_1():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_1_to_Independent_1():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [
        {"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {"volBackendFs": data["remoteFs"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_1_to_LW_1():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_1_to_Version2_1():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_1_to_Dependent_2():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_1_to_Independent_2():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_1_to_LW_2():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_1_to_Version2_2():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_1_to_Dependent_3():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_1_to_Independent_3():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_1_to_LW_3():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_1_to_Version2_3():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_1_to_Dependent_4():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_1_to_Independent_4():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_1_to_LW_4():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_1_to_Version2_4():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_1_to_Dependent_5():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_1_to_Independent_5():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_1_to_LW_5():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_1_to_Version2_5():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_2_to_Dependent_1():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_2_to_Independent_1():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [
        {"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {"volBackendFs": data["remoteFs"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_2_to_LW_1():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_2_to_Version2_1():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_2_to_Dependent_2():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_2_to_Independent_2():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_2_to_LW_2():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_2_to_Version2_2():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_2_to_Dependent_3():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_2_to_Independent_3():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_2_to_LW_3():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_2_to_Version2_3():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_2_to_Dependent_4():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_2_to_Independent_4():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_2_to_LW_4():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_2_to_Version2_4():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_2_to_Dependent_5():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_2_to_Independent_5():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_2_to_LW_5():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_2_to_Version2_5():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_2_to_Dependent_1():
    value_sc = {"volBackendFs": data["remoteFs"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_2_to_Independent_1():
    value_sc = {"volBackendFs": data["remoteFs"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [
        {"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {"volBackendFs": data["remoteFs"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_2_to_LW_1():
    value_sc = {"volBackendFs": data["remoteFs"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_2_to_Version2_1():
    value_sc = {"volBackendFs": data["remoteFs"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_2_to_Dependent_2():
    value_sc = {"volBackendFs": data["remoteFs"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_2_to_Independent_2():
    value_sc = {"volBackendFs": data["remoteFs"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_2_to_LW_2():
    value_sc = {"volBackendFs": data["remoteFs"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_2_to_Version2_2():
    value_sc = {"volBackendFs": data["remoteFs"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_2_to_Dependent_3():
    value_sc = {"volBackendFs": data["remoteFs"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_2_to_Independent_3():
    value_sc = {"volBackendFs": data["remoteFs"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_2_to_LW_3():
    value_sc = {"volBackendFs": data["remoteFs"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_2_to_Version2_3():
    value_sc = {"volBackendFs": data["remoteFs"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_2_to_Dependent_4():
    value_sc = {"volBackendFs": data["remoteFs"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_2_to_Independent_4():
    value_sc = {"volBackendFs": data["remoteFs"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_2_to_LW_4():
    value_sc = {"volBackendFs": data["remoteFs"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_2_to_Version2_4():
    value_sc = {"volBackendFs": data["remoteFs"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_2_to_Dependent_5():
    value_sc = {"volBackendFs": data["remoteFs"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_2_to_Independent_5():
    value_sc = {"volBackendFs": data["remoteFs"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_2_to_LW_5():
    value_sc = {"volBackendFs": data["remoteFs"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_2_to_Version2_5():
    value_sc = {"volBackendFs": data["remoteFs"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_2_to_Dependent_1():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_2_to_Independent_1():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [
        {"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {"volBackendFs": data["remoteFs"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_2_to_LW_1():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_2_to_Version2_1():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_2_to_Dependent_2():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_2_to_Independent_2():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_2_to_LW_2():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_2_to_Version2_2():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_2_to_Dependent_3():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_2_to_Independent_3():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_2_to_LW_3():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_2_to_Version2_3():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_2_to_Dependent_4():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_2_to_Independent_4():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_2_to_LW_4():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_2_to_Version2_4():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_2_to_Dependent_5():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_2_to_Independent_5():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_2_to_LW_5():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_2_to_Version2_5():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_2_to_Dependent_1():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_2_to_Independent_1():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [
        {"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {"volBackendFs": data["remoteFs"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_2_to_LW_1():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_2_to_Version2_1():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_2_to_Dependent_2():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_2_to_Independent_2():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_2_to_LW_2():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_2_to_Version2_2():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_2_to_Dependent_3():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_2_to_Independent_3():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_2_to_LW_3():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_2_to_Version2_3():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_2_to_Dependent_4():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_2_to_Independent_4():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_2_to_LW_4():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_2_to_Version2_4():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_2_to_Dependent_5():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_2_to_Independent_5():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_2_to_LW_5():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_2_to_Version2_5():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_3_to_Dependent_1():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_3_to_Independent_1():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [
        {"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {"volBackendFs": data["remoteFs"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_3_to_LW_1():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_3_to_Version2_1():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_3_to_Dependent_2():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_3_to_Independent_2():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_3_to_LW_2():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_3_to_Version2_2():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_3_to_Dependent_3():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_3_to_Independent_3():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_3_to_LW_3():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_3_to_Version2_3():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_3_to_Dependent_4():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_3_to_Independent_4():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_3_to_LW_4():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_3_to_Version2_4():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_3_to_Dependent_5():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_3_to_Independent_5():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_3_to_LW_5():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_3_to_Version2_5():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_3_to_Dependent_1():
    value_sc = {"volBackendFs": data["remoteFs"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_3_to_Independent_1():
    value_sc = {"volBackendFs": data["remoteFs"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [
        {"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {"volBackendFs": data["remoteFs"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_3_to_LW_1():
    value_sc = {"volBackendFs": data["remoteFs"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_3_to_Version2_1():
    value_sc = {"volBackendFs": data["remoteFs"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_3_to_Dependent_2():
    value_sc = {"volBackendFs": data["remoteFs"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_3_to_Independent_2():
    value_sc = {"volBackendFs": data["remoteFs"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_3_to_LW_2():
    value_sc = {"volBackendFs": data["remoteFs"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_3_to_Version2_2():
    value_sc = {"volBackendFs": data["remoteFs"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_3_to_Dependent_3():
    value_sc = {"volBackendFs": data["remoteFs"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_3_to_Independent_3():
    value_sc = {"volBackendFs": data["remoteFs"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_3_to_LW_3():
    value_sc = {"volBackendFs": data["remoteFs"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_3_to_Version2_3():
    value_sc = {"volBackendFs": data["remoteFs"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_3_to_Dependent_4():
    value_sc = {"volBackendFs": data["remoteFs"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_3_to_Independent_4():
    value_sc = {"volBackendFs": data["remoteFs"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_3_to_LW_4():
    value_sc = {"volBackendFs": data["remoteFs"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_3_to_Version2_4():
    value_sc = {"volBackendFs": data["remoteFs"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_3_to_Dependent_5():
    value_sc = {"volBackendFs": data["remoteFs"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_3_to_Independent_5():
    value_sc = {"volBackendFs": data["remoteFs"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_3_to_LW_5():
    value_sc = {"volBackendFs": data["remoteFs"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_3_to_Version2_5():
    value_sc = {"volBackendFs": data["remoteFs"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_3_to_Dependent_1():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r_volDirBasePath"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_3_to_Independent_1():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r_volDirBasePath"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [
        {"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {"volBackendFs": data["remoteFs"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_3_to_LW_1():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r_volDirBasePath"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_3_to_Version2_1():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r_volDirBasePath"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_3_to_Dependent_2():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r_volDirBasePath"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_3_to_Independent_2():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r_volDirBasePath"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_3_to_LW_2():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r_volDirBasePath"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_3_to_Version2_2():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r_volDirBasePath"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_3_to_Dependent_3():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r_volDirBasePath"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_3_to_Independent_3():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r_volDirBasePath"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_3_to_LW_3():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r_volDirBasePath"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_3_to_Version2_3():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r_volDirBasePath"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_3_to_Dependent_4():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r_volDirBasePath"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_3_to_Independent_4():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r_volDirBasePath"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_3_to_LW_4():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r_volDirBasePath"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_3_to_Version2_4():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r_volDirBasePath"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_3_to_Dependent_5():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r_volDirBasePath"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_3_to_Independent_5():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r_volDirBasePath"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_3_to_LW_5():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r_volDirBasePath"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_3_to_Version2_5():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r_volDirBasePath"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_3_to_Dependent_1():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_3_to_Independent_1():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [
        {"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {"volBackendFs": data["remoteFs"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_3_to_LW_1():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_3_to_Version2_1():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_3_to_Dependent_2():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_3_to_Independent_2():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_3_to_LW_2():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_3_to_Version2_2():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_3_to_Dependent_3():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_3_to_Independent_3():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_3_to_LW_3():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_3_to_Version2_3():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_3_to_Dependent_4():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_3_to_Independent_4():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_3_to_LW_4():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_3_to_Version2_4():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_3_to_Dependent_5():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_3_to_Independent_5():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_3_to_LW_5():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_3_to_Version2_5():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_4_to_Dependent_1():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_4_to_Independent_1():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [
        {"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {"volBackendFs": data["remoteFs"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_4_to_LW_1():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_4_to_Version2_1():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_4_to_Dependent_2():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_4_to_Independent_2():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_4_to_LW_2():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_4_to_Version2_2():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_4_to_Dependent_3():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_4_to_Independent_3():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_4_to_LW_3():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_4_to_Version2_3():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_4_to_Dependent_4():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_4_to_Independent_4():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_4_to_LW_4():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_4_to_Version2_4():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_4_to_Dependent_5():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_4_to_Independent_5():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_4_to_LW_5():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_4_to_Version2_5():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_4_to_Dependent_1():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_4_to_Independent_1():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [
        {"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {"volBackendFs": data["remoteFs"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_4_to_LW_1():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_4_to_Version2_1():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_4_to_Dependent_2():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_4_to_Independent_2():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_4_to_LW_2():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_4_to_Version2_2():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_4_to_Dependent_3():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_4_to_Independent_3():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_4_to_LW_3():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_4_to_Version2_3():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_4_to_Dependent_4():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_4_to_Independent_4():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_4_to_LW_4():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_4_to_Version2_4():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_4_to_Dependent_5():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_4_to_Independent_5():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_4_to_LW_5():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_4_to_Version2_5():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_4_to_Dependent_1():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_4_to_Independent_1():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [
        {"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {"volBackendFs": data["remoteFs"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_4_to_LW_1():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_4_to_Version2_1():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_4_to_Dependent_2():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_4_to_Independent_2():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_4_to_LW_2():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_4_to_Version2_2():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_4_to_Dependent_3():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_4_to_Independent_3():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_4_to_LW_3():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_4_to_Version2_3():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_4_to_Dependent_4():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_4_to_Independent_4():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_4_to_LW_4():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_4_to_Version2_4():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_4_to_Dependent_5():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_4_to_Independent_5():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_4_to_LW_5():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_4_to_Version2_5():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_4_to_Dependent_1():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_4_to_Independent_1():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [
        {"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {"volBackendFs": data["remoteFs"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_4_to_LW_1():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_4_to_Version2_1():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_4_to_Dependent_2():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_4_to_Independent_2():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_4_to_LW_2():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_4_to_Version2_2():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_4_to_Dependent_3():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_4_to_Independent_3():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_4_to_LW_3():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_4_to_Version2_3():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_4_to_Dependent_4():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_4_to_Independent_4():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_4_to_LW_4():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_4_to_Version2_4():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_4_to_Dependent_5():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_4_to_Independent_5():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_4_to_LW_5():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_4_to_Version2_5():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_5_to_Dependent_1():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_5_to_Independent_1():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [
        {"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {"volBackendFs": data["remoteFs"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_5_to_LW_1():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_5_to_Version2_1():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_5_to_Dependent_2():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_5_to_Independent_2():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_5_to_LW_2():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_5_to_Version2_2():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_5_to_Dependent_3():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_5_to_Independent_3():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_5_to_LW_3():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_5_to_Version2_3():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_5_to_Dependent_4():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_5_to_Independent_4():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_5_to_LW_4():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_5_to_Version2_4():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_5_to_Dependent_5():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_5_to_Independent_5():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_5_to_LW_5():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Dependent_5_to_Version2_5():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_5_to_Dependent_1():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_5_to_Independent_1():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [
        {"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {"volBackendFs": data["remoteFs"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_5_to_LW_1():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_5_to_Version2_1():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_5_to_Dependent_2():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_5_to_Independent_2():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_5_to_LW_2():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_5_to_Version2_2():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_5_to_Dependent_3():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_5_to_Independent_3():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_5_to_LW_3():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_5_to_Version2_3():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_5_to_Dependent_4():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_5_to_Independent_4():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_5_to_LW_4():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_5_to_Version2_4():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_5_to_Dependent_5():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_5_to_Independent_5():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_5_to_LW_5():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Independent_5_to_Version2_5():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_5_to_Dependent_1():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_5_to_Independent_1():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [
        {"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {"volBackendFs": data["remoteFs"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_5_to_LW_1():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_5_to_Version2_1():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_5_to_Dependent_2():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_5_to_Independent_2():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_5_to_LW_2():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_5_to_Version2_2():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_5_to_Dependent_3():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_5_to_Independent_3():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_5_to_LW_3():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_5_to_Version2_3():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_5_to_Dependent_4():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_5_to_Independent_4():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_5_to_LW_4():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_5_to_Version2_4():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_5_to_Dependent_5():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_5_to_Independent_5():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_5_to_LW_5():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_volume_cloning_LW_5_to_Version2_5():
    value_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_5_to_Dependent_1():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_5_to_Independent_1():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [
        {"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {"volBackendFs": data["remoteFs"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_5_to_LW_1():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_5_to_Version2_1():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_5_to_Dependent_2():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_5_to_Independent_2():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_5_to_LW_2():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_5_to_Version2_2():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_5_to_Dependent_3():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_5_to_Independent_3():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_5_to_LW_3():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_5_to_Version2_3():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_5_to_Dependent_4():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_5_to_Independent_4():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_5_to_LW_4():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_5_to_Version2_4():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_5_to_Dependent_5():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_5_to_Independent_5():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_5_to_LW_5():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_Version2_5_to_Version2_5():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}], "clone_sc": {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)
'''