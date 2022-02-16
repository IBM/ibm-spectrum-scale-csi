import logging
import pytest
import ibm_spectrum_scale_csi.scale_operator as scaleop
import ibm_spectrum_scale_csi.common_utils.input_data_functions as inputfunc
LOGGER = logging.getLogger()
pytestmark = [pytest.mark.volumeprovisioning, pytest.mark.localcluster, pytest.mark.cg]


@pytest.fixture(scope='session', autouse=True)
def values(request, check_csi_operator):
    global data, driver_object, kubeconfig_value  # are required in every testcase
    kubeconfig_value, clusterconfig_value, operator_namespace, test_namespace, runslow_val, operator_yaml = inputfunc.get_cmd_values(request)

    data = inputfunc.read_driver_data(clusterconfig_value, test_namespace, operator_namespace, kubeconfig_value)
    keep_objects = data["keepobjects"]
    if not(data["volBackendFs"] == ""):
        data["primaryFs"] = data["volBackendFs"]

    if runslow_val:
        value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"},
                     {"access_modes": "ReadWriteOnce", "storage": "1Gi"},
                     {"access_modes": "ReadOnlyMany", "storage": "1Gi",
                      "reason": "ReadOnlyMany is not supported"}
                     ]
        value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False"},
                     {"mount_path": "/usr/share/nginx/html/scale",
 
                     "read_only": "True", "reason": "Read-only file system"}
                     ]
    else:
        value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
        value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False"}]

    driver_object = scaleop.Driver(kubeconfig_value, value_pvc, value_pod, data["id"],
                                   test_namespace, keep_objects, data["image_name"], data["pluginNodeSelector"])


#: Testcase that are expected to pass:
@pytest.mark.regression
def test_get_version():
    LOGGER.info("Cluster Details:")
    LOGGER.info("----------------")
    scaleop.filesetfunc.get_scale_version(data)
    scaleop.kubeobjectfunc.get_kubernetes_version(kubeconfig_value)
    scaleop.kubeobjectfunc.get_operator_image()
    scaleop.kubeobjectfunc.get_driver_image()


def test_driver_cg_pass_1():
    value_sc = {"volBackendFs": data["primaryFs"], "version": "2"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}] * 5
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


@pytest.mark.regression
def test_driver_cg_pass_2():
    value_sc = {"volBackendFs": data["primaryFs"], "version": "2"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}, {"access_modes": "ReadWriteMany", "storage": "8Gi"}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_cg_pass_3():
    value_sc = {"volBackendFs": data["primaryFs"], "version": "2", "inodeLimit": data["inodeLimit"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}] * 2
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_cg_pass_4():
    value_sc = {"volBackendFs": data["primaryFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}] * 2
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_cg_pass_5():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"], "version": "2"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}] * 2
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_cg_fail_1():
    value_sc = {"volBackendFs": data["primaryFs"], "version": "2", "volDirBasePath": data["volDirBasePath"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi", "reason": "InvalidArgument desc = volDirBasePath and version=2"}] * 2
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_cg_fail_2():
    value_sc = {"volBackendFs": data["primaryFs"], "version": "2", "filesetType": "dependent"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi", "reason": "InvalidArgument desc = filesetType and version=2"}] * 2
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_cg_fail_3():
    value_sc = {"volBackendFs": data["primaryFs"], "version": "2", "filesetType": "independent"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi", "reason": "InvalidArgument desc = filesetType and version=2"}] * 2
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_cg_fail_4():
    value_sc = {"version": "2"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi", "reason": "InvalidArgument desc = volBackendFs must be specified"}] * 2
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_cg_fail_5():
    value_sc = {"volBackendFs": data["primaryFs"], "version": "2", "parentFileset": data["parentFileset"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi", "reason": "InvalidArgument desc = parentFileset and version=2"}] * 2
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_cg_fail_6():
    value_sc = {"volBackendFs": data["primaryFs"], "version": "3"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi", "reason": 'parameter "version" can have values only "1" or "2"'}] * 2
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_cg_expansion_1():
    value_sc = {"volBackendFs": data["primaryFs"], "version": "2", "allow_volume_expansion": True}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi", "volume_expansion_storage": ["4Gi", "16Gi"]}] * 2
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


@pytest.mark.regression
def test_driver_cg_expansion_2():
    value_sc = {"volBackendFs": data["primaryFs"], "version": "2", "allow_volume_expansion": True}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi", "volume_expansion_storage": ["100Gi", "250Gi"]}] * 2
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_cg_cloning_1():
    value_sc = {"volBackendFs": data["primaryFs"], "version": "2"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}] * 2
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}, {"access_modes": "ReadWriteOnce", "storage": "1Gi"}]}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_clone_passed=value_clone_passed)


@pytest.mark.regression
def test_driver_cg_cloning_2():
    value_sc = {"volBackendFs": data["primaryFs"], "version": "2"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}] * 2
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "11Gi"}, {"access_modes": "ReadWriteOnce", "storage": "8Gi"}]}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_clone_passed=value_clone_passed)


@pytest.mark.xfail
def test_driver_cg_permissions_777_1():
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False", "sub_path": ["sub_path_mnt"], "volumemount_readonly":[False]},
                 {"mount_path": "/usr/share/nginx/html/scale", "read_only": "False", "sub_path": ["sub_path_mnt"], "volumemount_readonly":[True],
                  "reason": "Read-only file system"},
                 {"mount_path": "/usr/share/nginx/html/scale", "read_only": "True",
                  "sub_path": ["sub_path_mnt"], "volumemount_readonly":[False], "reason": "Read-only file system"}
                 ]
    value_sc = {"volBackendFs": data["primaryFs"], "version": "2", "permissions": "777",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    driver_object.test_dynamic(value_sc, value_pod_passed=value_pod)


def test_driver_cg_compression_1():
    value_sc = {"volBackendFs": data["primaryFs"], "version": "2",  "compression": "z", "clusterId": data["id"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}] * 2
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


@pytest.mark.regression
def test_driver_cg_compression_2():
    value_sc = {"volBackendFs": data["primaryFs"], "version": "2",  "compression": "true"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}] * 2
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_cg_compression_3():
    value_sc = {"volBackendFs": data["primaryFs"], "version": "2",  "compression": "false"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}] * 2
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_cg_compression_4():
    value_sc = {"volBackendFs": data["primaryFs"], "version": "2",  "compression": "alphah"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}] * 2
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_cg_compression_5():
    value_sc = {"volBackendFs": data["primaryFs"], "version": "2",  "compression": "lz4"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}] * 2
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_cg_compression_6():
    value_sc = {"volBackendFs": data["primaryFs"], "version": "2",  "compression": "zfast"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}] * 2
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_cg_compression_7():
    value_sc = {"volBackendFs": data["primaryFs"], "version": "2",  "compression": "alphae"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}] * 2
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_cg_compression_fail_1():
    value_sc = {"volBackendFs": data["primaryFs"], "version": "2",  "compression": "wrongalgo"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi", "reason": "InvalidArgument desc = invalid compression algorithm"}] * 2
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_cg_compression_clone_1():
    value_sc = {"volBackendFs": data["primaryFs"], "version": "2",  "compression": "z"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}] * 2
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "11Gi"}, {"access_modes": "ReadWriteOnce", "storage": "8Gi"}]}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_clone_passed=value_clone_passed)


@pytest.mark.regression
def test_driver_cg_tier_pass_1():
    value_sc = {"volBackendFs": data["primaryFs"], "version": "2", "tier": data["tier"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}, {"access_modes": "ReadWriteMany", "storage": "8Gi"}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_cg_tier_fail_1():
    value_sc = {"volBackendFs": data["primaryFs"], "version": "2", "tier": "wrongtier", "reason":"400 Invalid value in 'storagePool'"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}, {"access_modes": "ReadWriteMany", "storage": "8Gi"}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_cg_tier_fail_2():
    value_sc = {"volBackendFs": data["primaryFs"], "version": "3", "tier": data["tier"], "reason": 'parameter "version" can have values only "1" or "2"'}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}, {"access_modes": "ReadWriteMany", "storage": "8Gi"}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_cg_tier_clone_1():
    value_sc = {"volBackendFs": data["primaryFs"], "version": "2",  "tier": data["tier"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}] * 2
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "11Gi"}, {"access_modes": "ReadWriteOnce", "storage": "8Gi"}]}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_clone_passed=value_clone_passed)


def test_driver_cg_tier_compression():
    value_sc = {"volBackendFs": data["primaryFs"], "version": "2", "tier": data["tier"], "compression": "z"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}, {"access_modes": "ReadWriteMany", "storage": "8Gi"}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


@pytest.mark.regression
def test_driver_cg_tier_compression_clone():
    value_sc = {"volBackendFs": data["primaryFs"], "version": "2",  "tier": data["tier"], "compression": "z"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}] * 2
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "11Gi"}, {"access_modes": "ReadWriteOnce", "storage": "8Gi"}]}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_clone_passed=value_clone_passed)
