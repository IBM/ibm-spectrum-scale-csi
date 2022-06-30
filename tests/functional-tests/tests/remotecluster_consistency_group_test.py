import logging
import pytest
import ibm_spectrum_scale_csi.base_class as baseclass
import ibm_spectrum_scale_csi.common_utils.input_data_functions as inputfunc
LOGGER = logging.getLogger()
pytestmark = [pytest.mark.volumeprovisioning, pytest.mark.remotecluster, pytest.mark.cg]


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


def test_driver_cg_pass_1():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "consistencyGroup": "remote-test_driver_cg_pass_1-cg"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}] * 5
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


@pytest.mark.regression
def test_driver_cg_pass_2():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "consistencyGroup": None}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"},
                 {"access_modes": "ReadWriteMany", "storage": "8Gi"}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_cg_pass_3():
    value_sc = {"volBackendFs": data["remoteFs"],
                "version": "2", "inodeLimit": data["r_inodeLimit"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}] * 2
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_cg_pass_4():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}] * 2
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_cg_pass_5():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"], "version": "2"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}] * 2
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_cg_fail_1():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "volDirBasePath": data["r_volDirBasePath"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi",
                  "reason": "InvalidArgument desc = volDirBasePath and version=2"}] * 2
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_cg_fail_2():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "filesetType": "dependent"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi",
                  "reason": "InvalidArgument desc = filesetType and version=2"}] * 2
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_cg_fail_3():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "filesetType": "independent"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi",
                  "reason": "InvalidArgument desc = filesetType and version=2"}] * 2
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_cg_fail_4():
    value_sc = {"version": "2"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi",
                  "reason": "InvalidArgument desc = volBackendFs must be specified"}] * 2
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_cg_fail_5():
    value_sc = {"volBackendFs": data["remoteFs"],
                "version": "2", "parentFileset": data["r_parentFileset"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi",
                  "reason": "InvalidArgument desc = parentFileset and version=2"}] * 2
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_cg_fail_6():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "3"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi",
                  "reason": 'parameter "version" can have values only "1" or "2"'}] * 2
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_cg_expansion_1():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "allow_volume_expansion": True}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi",
                  "volume_expansion_storage": ["4Gi", "16Gi"]}] * 2
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False"}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_pod_passed=value_pod)


def test_driver_cg_expansion_2():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "allow_volume_expansion": True}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi",
                  "volume_expansion_storage": ["100Gi", "250Gi"]}] * 2
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False"}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_pod_passed=value_pod)


@pytest.mark.xfail
def test_driver_cg_permissions_777_1():
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False", "sub_path": ["sub_path_mnt"], "volumemount_readonly":[False], "runAsUser": "2000", "runAsGroup": "5000"},
                 {"mount_path": "/usr/share/nginx/html/scale", "read_only": "False", "sub_path": ["sub_path_mnt"], "volumemount_readonly":[True],
                  "reason": "Read-only file system", "runAsUser": "2000", "runAsGroup": "5000"},
                 {"mount_path": "/usr/share/nginx/html/scale", "read_only": "True",
                  "sub_path": ["sub_path_mnt"], "volumemount_readonly":[False], "reason": "Read-only file system", "runAsUser": "2000", "runAsGroup": "5000"}
                 ]
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "permissions": "777",
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    driver_object.test_dynamic(value_sc, value_pod_passed=value_pod)


def test_driver_cg_compression_1():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "compression": "z", "clusterId": data["remoteid"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}] * 2
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_cg_compression_2():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",  "compression": "true"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}] * 2
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_cg_compression_3():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",  "compression": "false"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}] * 2
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_cg_compression_4():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",  "compression": "alphah"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}] * 2
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_cg_compression_5():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",  "compression": "lz4"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}] * 2
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_cg_compression_6():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",  "compression": "zfast"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}] * 2
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_cg_compression_7():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",  "compression": "alphae"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}] * 2
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_cg_compression_fail_1():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",  "compression": "wrongalgo"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi",
                  "reason": "InvalidArgument desc = invalid compression algorithm"}] * 2
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_cg_compression_clone_1():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",  "compression": "z"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}] * 2
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "11Gi"}, {
        "access_modes": "ReadWriteOnce", "storage": "8Gi"}]}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


@pytest.mark.regression
def test_driver_cg_tier_pass_1():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "tier": data["r_tier"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"},
                 {"access_modes": "ReadWriteMany", "storage": "8Gi"}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_cg_tier_fail_1():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "tier": "wrongtier",
                "reason": "invalid tier 'wrongtier' specified for filesystem"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"},
                 {"access_modes": "ReadWriteMany", "storage": "8Gi"}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_cg_tier_fail_2():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "3", "tier": data["r_tier"],
                "reason": 'parameter "version" can have values only "1" or "2"'}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"},
                 {"access_modes": "ReadWriteMany", "storage": "8Gi"}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_cg_tier_clone_1():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",  "tier": data["r_tier"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}] * 2
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "11Gi"}, {
        "access_modes": "ReadWriteOnce", "storage": "8Gi"}]}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_cg_tier_compression():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "tier": data["r_tier"], "compression": "z"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"},
                 {"access_modes": "ReadWriteMany", "storage": "8Gi"}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


@pytest.mark.regression
def test_driver_cg_tier_compression_clone():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2",
                "tier": data["r_tier"], "compression": "z"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}] * 2
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "11Gi"}, {
        "access_modes": "ReadWriteOnce", "storage": "8Gi"}]}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc,
                               value_clone_passed=value_clone_passed)


def test_driver_shared_fsgroup_rwo():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteOnce", "storage": "1Gi"}]
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False", "fsgroup": "3000",
                 "runasnonroot": False}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_pod_passed=value_pod)


def test_driver_shared_fsgroup_rwo_subpath():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteOnce", "storage": "1Gi"}]
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False", "sub_path": ["sub_path_mnt"], "volumemount_readonly":[False],
                 "fsgroup": "3000", "runasnonroot": False}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_pod_passed=value_pod)


def test_driver_shared_fsgroup_nonroot_rwo():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteOnce", "storage": "1Gi"}]
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False", "fsgroup": "3000",
                 "runAsUser": "2000", "runAsGroup": "5000", "runasnonroot": True}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_pod_passed=value_pod)


def test_driver_shared_fsgroup_nonroot_rwo_subpath():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteOnce", "storage": "1Gi"}]
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False", "sub_path": ["sub_path_mnt"], "volumemount_readonly":[False],
                 "fsgroup": "3000", "runAsUser": "2000", "runAsGroup": "5000", "runasnonroot": True}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_pod_passed=value_pod)


def test_driver_shared_rwo():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteOnce", "storage": "1Gi"}]
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False",
                 "runasnonroot": False}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_pod_passed=value_pod)


def test_driver_shared_rwo_subpath():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteOnce", "storage": "1Gi"}]
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False", "sub_path": ["sub_path_mnt"], "volumemount_readonly":[False],
                 "runasnonroot": False}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_pod_passed=value_pod)


def test_driver_shared_nonroot_rwo():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteOnce", "storage": "1Gi"}]
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False",
                 "runAsUser": "2000", "runAsGroup": "5000", "runasnonroot": True}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_pod_passed=value_pod)


def test_driver_shared_nonroot_rwo_subpath():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteOnce", "storage": "1Gi"}]
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False", "sub_path": ["sub_path_mnt"], "volumemount_readonly":[False],
                 "runAsUser": "2000", "runAsGroup": "5000", "runasnonroot": True}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_pod_passed=value_pod)


def test_driver_fsgroup_rwo():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "shared": "False"}
    value_pvc = [{"access_modes": "ReadWriteOnce", "storage": "1Gi"}]
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False", "fsgroup": "3000",
                 "runasnonroot": False}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_pod_passed=value_pod)


def test_driver_fsgroup_rwo_subpath():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "shared": "False"}
    value_pvc = [{"access_modes": "ReadWriteOnce", "storage": "1Gi"}]
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False", "sub_path": ["sub_path_mnt"], "volumemount_readonly":[False],
                 "fsgroup": "3000", "runasnonroot": False}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_pod_passed=value_pod)


def test_driver_fsgroup_nonroot_rwo():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2"}
    value_pvc = [{"access_modes": "ReadWriteOnce", "storage": "1Gi"}]
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False", "fsgroup": "3000",
                 "runAsUser": "2000", "runAsGroup": "5000", "runasnonroot": True}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_pod_passed=value_pod)


def test_driver_fsgroup_nonroot_rwo_subpath():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2"}
    value_pvc = [{"access_modes": "ReadWriteOnce", "storage": "1Gi"}]
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False", "sub_path": ["sub_path_mnt"], "volumemount_readonly":[False],
                 "fsgroup": "3000", "runAsUser": "2000", "runAsGroup": "5000", "runasnonroot": True}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_pod_passed=value_pod)


def test_driver_nonroot_rwo():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2"}
    value_pvc = [{"access_modes": "ReadWriteOnce", "storage": "1Gi"}]
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False", "runAsUser": "2000", "runAsGroup": "5000",
                 "runasnonroot": True, "reason": "Permission denied"}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_pod_passed=value_pod)


def test_driver_nonroot_rwo_subpath():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2"}
    value_pvc = [{"access_modes": "ReadWriteOnce", "storage": "1Gi"}]
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False", "sub_path": ["sub_path_mnt"], "volumemount_readonly":[False],
                 "runAsUser": "2000", "runAsGroup": "5000", "runasnonroot": True,  "reason": "Permission denied"}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_pod_passed=value_pod)


def test_driver_shared_fsgroup_rwx():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False", "fsgroup": "3000",
                 "runasnonroot": False}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_pod_passed=value_pod)


def test_driver_shared_fsgroup_rwx_subpath():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False", "sub_path": ["sub_path_mnt"], "volumemount_readonly":[False],
                 "fsgroup": "3000", "runasnonroot": False}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_pod_passed=value_pod)


def test_driver_shared_fsgroup_nonroot_rwx():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False", "fsgroup": "3000",
                 "runAsUser": "2000", "runAsGroup": "5000", "runasnonroot": True}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_pod_passed=value_pod)


def test_driver_shared_fsgroup_nonroot_rwx_subpath():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False", "sub_path": ["sub_path_mnt"], "volumemount_readonly":[False],
                 "fsgroup": "3000", "runAsUser": "2000", "runAsGroup": "5000", "runasnonroot": True}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_pod_passed=value_pod)


def test_driver_shared_rwx():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False",
                 "runasnonroot": False}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_pod_passed=value_pod)


def test_driver_shared_rwx_subpath():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False", "sub_path": ["sub_path_mnt"], "volumemount_readonly":[False],
                 "runasnonroot": False}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_pod_passed=value_pod)


def test_driver_shared_nonroot_rwx():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False",
                 "runAsUser": "2000", "runAsGroup": "5000", "runasnonroot": True}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_pod_passed=value_pod)


def test_driver_shared_nonroot_rwx_subpath():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False", "sub_path": ["sub_path_mnt"], "volumemount_readonly":[False],
                 "runAsUser": "2000", "runAsGroup": "5000", "runasnonroot": True}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_pod_passed=value_pod)


def test_driver_fsgroup_rwx():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False", "fsgroup": "3000",
                 "runasnonroot": False}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_pod_passed=value_pod)


def test_driver_fsgroup_rwx_subpath():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False", "sub_path": ["sub_path_mnt"], "volumemount_readonly":[False],
                 "fsgroup": "3000", "runasnonroot": False}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_pod_passed=value_pod)


def test_driver_fsgroup_nonroot_rwx():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False", "fsgroup": "3000",
                 "runAsUser": "2000", "runAsGroup": "5000", "runasnonroot": True, "reason": "Permission denied"}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_pod_passed=value_pod)


def test_driver_fsgroup_nonroot_rwx_subpath():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False", "sub_path": ["sub_path_mnt"], "volumemount_readonly":[False],
                 "fsgroup": "3000", "runAsUser": "2000", "runAsGroup": "5000", "runasnonroot": True, "reason": "Permission denied"}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_pod_passed=value_pod)


def test_driver_nonroot_rwx():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False", "runAsUser": "2000", "runAsGroup": "5000",
                 "runasnonroot": True, "reason": "Permission denied"}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_pod_passed=value_pod)


def test_driver_nonroot_rwx_subpath():
    value_sc = {"volBackendFs": data["remoteFs"], "version": "2"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False", "sub_path": ["sub_path_mnt"], "volumemount_readonly":[False],
                 "runAsUser": "2000", "runAsGroup": "5000", "runasnonroot": True,  "reason": "Permission denied"}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_pod_passed=value_pod)
