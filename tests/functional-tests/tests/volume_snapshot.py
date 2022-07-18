import logging
import pytest
import ibm_spectrum_scale_csi.base_class as baseclass
import ibm_spectrum_scale_csi.common_utils.input_data_functions as inputfunc

LOGGER = logging.getLogger()
pytestmark = [pytest.mark.volumesnapshot, pytest.mark.localcluster]


@pytest.fixture(autouse=True)
def values(data_fixture, check_csi_operator, local_cluster_fixture):
    global data, snapshot_object, kubeconfig_value  # are required in every testcase
    data = data_fixture["driver_data"]
    kubeconfig_value = data_fixture["cmd_values"]["kubeconfig_value"]
    snapshot_object = data_fixture["local_snapshot_object"]


@pytest.mark.regression
def test_get_version():
    LOGGER.info("Cluster Details:")
    LOGGER.info("----------------")
    baseclass.filesetfunc.get_scale_version(data)
    baseclass.kubeobjectfunc.get_kubernetes_version(kubeconfig_value)
    baseclass.kubeobjectfunc.get_operator_image()
    baseclass.kubeobjectfunc.get_driver_image()


@pytest.mark.regression
def test_snapshot_static_pass_1():
    value_sc = {"volBackendFs": data["localFs"], "clusterId": data["id"]}
    snapshot_object.test_static(value_sc, test_restore=True)


@pytest.mark.regression
def test_snapshot_static_multiple_snapshots():
    value_sc = {"volBackendFs": data["localFs"], "clusterId": data["id"]}
    snapshot_object.test_static(value_sc, test_restore=True, number_of_snapshots=3)


def test_snapshot_static_pass_3():
    value_sc = {"volBackendFs": data["localFs"],
                "clusterId": data["id"], "gid": data["gid_number"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_4():
    value_sc = {"volBackendFs": data["localFs"],
                "clusterId": data["id"], "inodeLimit": data["inodeLimit"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_5():
    value_sc = {"volBackendFs": data["localFs"], "clusterId": data["id"],
                "inodeLimit": data["inodeLimit"], "uid": data["uid_number"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_6():
    value_sc = {"volBackendFs": data["localFs"], "clusterId": data["id"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_7():
    value_sc = {"volBackendFs": data["localFs"], "clusterId": data["id"],
                "inodeLimit": data["inodeLimit"], "gid": data["gid_number"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_8():
    value_sc = {"volBackendFs": data["localFs"], "clusterId": data["id"],
                "inodeLimit": data["inodeLimit"], "uid": data["uid_number"],
                "gid": data["gid_number"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_9():
    value_sc = {"volBackendFs": data["localFs"],
                "clusterId": data["id"], "uid": data["uid_number"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_10():
    value_sc = {"volBackendFs": data["localFs"],
                "inodeLimit": data["inodeLimit"],
                "clusterId": data["id"], "filesetType": "independent"}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_11():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "clusterId": data["id"],
                "filesetType": "independent", "inodeLimit": data["inodeLimit"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_12():
    value_sc = {"volBackendFs": data["localFs"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_13():
    value_sc = {"uid": data["uid_number"], "volBackendFs": data["localFs"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_14():
    value_sc = {"gid": data["gid_number"], "volBackendFs": data["localFs"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_15():
    value_sc = {"inodeLimit": data["inodeLimit"],
                "volBackendFs": data["localFs"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_16():
    value_sc = {"volBackendFs": data["localFs"], "uid": data["uid_number"],
                "gid": data["gid_number"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_17():
    value_sc = {"volBackendFs": data["localFs"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_18():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "inodeLimit": data["inodeLimit"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_19():
    value_sc = {"clusterId": data["id"], "gid": data["gid_number"],
                "uid": data["uid_number"], "volBackendFs": data["localFs"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_20():
    value_sc = {"clusterId": data["id"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"],
                "volBackendFs": data["localFs"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_21():
    value_sc = {"clusterId": data["id"], "gid": data["gid_number"],
                "inodeLimit": data["inodeLimit"],
                "volBackendFs": data["localFs"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_22():
    value_sc = {"gid": data["gid_number"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"],
                "volBackendFs": data["localFs"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_23():
    value_sc = {"clusterId": data["id"], "volBackendFs": data["localFs"],
                "gid": data["gid_number"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_24():
    value_sc = {"volBackendFs": data["localFs"], "clusterId": data["id"]}
    snapshot_object.test_static(value_sc, test_restore=False)


def test_snapshot_static_pass_25():
    value_sc = {"volBackendFs": data["localFs"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"]}
    snapshot_object.test_static(value_sc, test_restore=False)


def test_snapshot_static_pass_26():
    value_sc = {"gid": data["gid_number"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"],
                "volBackendFs": data["localFs"]}
    snapshot_object.test_static(value_sc, test_restore=False)


@pytest.mark.regression
def test_snapshot_dynamic_pass_1():
    value_sc = {"volBackendFs": data["localFs"], "clusterId": data["id"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_2():
    value_sc = {"volBackendFs": data["localFs"], "clusterId": data["id"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True,
                                 value_vs_class={"deletionPolicy": "Retain"})


@pytest.mark.regression
@pytest.mark.xfail
def test_snapshot_dynamic_expected_fail_1():
    value_sc = {"volBackendFs": data["localFs"],
                "filesetType": "dependent", "clusterId": data["id"]}
    snapshot_object.test_dynamic(value_sc, test_restore=False,
                                 reason="Volume snapshot can only be created when source volume is independent fileset")


@pytest.mark.regression
@pytest.mark.xfail
def test_snapshot_dynamic_expected_fail_2():
    value_sc = {"volBackendFs": data["localFs"],
                "volDirBasePath": data["volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc, test_restore=False,
                                 reason="Volume snapshot can only be created when source volume is independent fileset")


def test_snapshot_dynamic_multiple_snapshots():
    value_sc = {"volBackendFs": data["localFs"], "clusterId": data["id"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, number_of_snapshots=3)


@pytest.mark.slow
def test_snapshot_dynamic_multiple_snapshots_256():
    value_sc = {"volBackendFs": data["localFs"], "clusterId": data["id"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, number_of_snapshots=256)


@pytest.mark.xfail
@pytest.mark.slow
def test_snapshot_dynamic_multiple_snapshots_257():
    value_sc = {"volBackendFs": data["localFs"], "clusterId": data["id"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, number_of_snapshots=257)


def test_snapshot_dynamic_pass_3():
    value_sc = {"volBackendFs": data["localFs"],
                "clusterId": data["id"], "gid": data["gid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_4():
    value_sc = {"volBackendFs": data["localFs"],
                "clusterId": data["id"], "inodeLimit": data["inodeLimit"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_5():
    value_sc = {"volBackendFs": data["localFs"], "clusterId": data["id"],
                "inodeLimit": data["inodeLimit"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_6():
    value_sc = {"volBackendFs": data["localFs"], "clusterId": data["id"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_7():
    value_sc = {"volBackendFs": data["localFs"], "clusterId": data["id"],
                "inodeLimit": data["inodeLimit"], "gid": data["gid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_8():
    value_sc = {"volBackendFs": data["localFs"], "clusterId": data["id"],
                "inodeLimit": data["inodeLimit"], "uid": data["uid_number"],
                "gid": data["gid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_9():
    value_sc = {"volBackendFs": data["localFs"],
                "clusterId": data["id"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_10():
    value_sc = {"volBackendFs": data["localFs"],
                "inodeLimit": data["inodeLimit"],
                "clusterId": data["id"], "filesetType": "independent"}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_11():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "clusterId": data["id"],
                "filesetType": "independent", "inodeLimit": data["inodeLimit"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_12():
    value_sc = {"volBackendFs": data["localFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_13():
    value_sc = {"uid": data["uid_number"], "volBackendFs": data["localFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_14():
    value_sc = {"gid": data["gid_number"], "volBackendFs": data["localFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_15():
    value_sc = {"inodeLimit": data["inodeLimit"],
                "volBackendFs": data["localFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_16():
    value_sc = {"volBackendFs": data["localFs"], "uid": data["uid_number"],
                "gid": data["gid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_17():
    value_sc = {"volBackendFs": data["localFs"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_18():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "inodeLimit": data["inodeLimit"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_19():
    value_sc = {"clusterId": data["id"], "gid": data["gid_number"],
                "uid": data["uid_number"], "volBackendFs": data["localFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_20():
    value_sc = {"clusterId": data["id"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"],
                "volBackendFs": data["localFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_21():
    value_sc = {"clusterId": data["id"], "gid": data["gid_number"],
                "inodeLimit": data["inodeLimit"],
                "volBackendFs": data["localFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_22():
    value_sc = {"gid": data["gid_number"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"],
                "volBackendFs": data["localFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_23():
    value_sc = {"clusterId": data["id"], "volBackendFs": data["localFs"],
                "gid": data["gid_number"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_24():
    value_sc = {"volBackendFs": data["localFs"], "clusterId": data["id"]}
    snapshot_object.test_dynamic(value_sc, test_restore=False)


def test_snapshot_dynamic_pass_25():
    value_sc = {"volBackendFs": data["localFs"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"]}
    snapshot_object.test_dynamic(value_sc, test_restore=False)


def test_snapshot_dynamic_pass_26():
    value_sc = {"gid": data["gid_number"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"],
                "volBackendFs": data["localFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=False)


def test_snapshot_static_different_sc_1():
    value_sc = {"volBackendFs": data["localFs"], "clusterId": data["id"]}
    restore_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    snapshot_object.test_static(value_sc, test_restore=True, restore_sc=restore_sc)


def test_snapshot_static_different_sc_2():
    value_sc = {"volBackendFs": data["localFs"], "clusterId": data["id"]}
    restore_sc = {"volBackendFs": data["localFs"],
                  "filesetType": "dependent", "clusterId": data["id"]}
    snapshot_object.test_static(value_sc, test_restore=True, restore_sc=restore_sc)


def test_snapshot_static_different_sc_3():
    value_sc = {"volBackendFs": data["localFs"], "clusterId": data["id"],
                "inodeLimit": data["inodeLimit"], "uid": data["uid_number"],
                "gid": data["gid_number"]}
    restore_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    snapshot_object.test_static(value_sc, test_restore=True, restore_sc=restore_sc)


def test_snapshot_static_different_sc_4():
    value_sc = {"volBackendFs": data["localFs"], "clusterId": data["id"],
                "inodeLimit": data["inodeLimit"], "uid": data["uid_number"],
                "gid": data["gid_number"]}
    restore_sc = {"volBackendFs": data["localFs"],
                  "filesetType": "dependent", "clusterId": data["id"]}
    snapshot_object.test_static(value_sc, test_restore=True, restore_sc=restore_sc)


def test_snapshot_static_different_sc_5():
    value_sc = {"volBackendFs": data["localFs"],
                "inodeLimit": data["inodeLimit"],
                "clusterId": data["id"], "filesetType": "independent"}
    restore_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    snapshot_object.test_static(value_sc, test_restore=True, restore_sc=restore_sc)


def test_snapshot_static_different_sc_6():
    value_sc = {"volBackendFs": data["localFs"],
                "inodeLimit": data["inodeLimit"],
                "clusterId": data["id"], "filesetType": "independent"}
    restore_sc = {"volBackendFs": data["localFs"],
                  "filesetType": "dependent", "clusterId": data["id"]}
    snapshot_object.test_static(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.regression
def test_snapshot_dynamic_nodeclass_1():
    value_sc = {"volBackendFs": data["localFs"], "clusterId": data["id"]}
    restore_sc = {"volBackendFs": data["localFs"],
                  "clusterId": data["id"], "nodeClass": "GUI_MGMT_SERVERS"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_snapshot_dynamic_nodeclass_2():
    value_sc = {"volBackendFs": data["localFs"], "clusterId": data["id"]}
    restore_sc = {"volBackendFs": data["localFs"],
                  "clusterId": data["id"], "nodeClass": "GUI_SERVERS"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.regression
def test_snapshot_dynamic_nodeclass_3():
    value_sc = {"volBackendFs": data["localFs"], "clusterId": data["id"]}
    restore_sc = {"volBackendFs": data["localFs"],
                  "clusterId": data["id"], "nodeClass": "randomnodeclassx"}
    restore_pvc = {"access_modes": "ReadWriteMany",
                   "storage": "1Gi", "reason": "NotFound desc = nodeclass"}
    snapshot_object.test_dynamic(value_sc, test_restore=True,
                                 restore_sc=restore_sc, restore_pvc=restore_pvc)


def test_snapshot_static_nodeclass_1():
    value_sc = {"volBackendFs": data["localFs"], "clusterId": data["id"]}
    restore_sc = {"volBackendFs": data["localFs"],
                  "clusterId": data["id"], "nodeClass": "GUI_MGMT_SERVERS"}
    snapshot_object.test_static(value_sc, test_restore=True, restore_sc=restore_sc)


def test_snapshot_static_nodeclass_2():
    value_sc = {"volBackendFs": data["localFs"], "clusterId": data["id"]}
    restore_sc = {"volBackendFs": data["localFs"],
                  "clusterId": data["id"], "nodeClass": "GUI_SERVERS"}
    snapshot_object.test_static(value_sc, test_restore=True, restore_sc=restore_sc)


def test_snapshot_static_nodeclass_3():
    value_sc = {"volBackendFs": data["localFs"], "clusterId": data["id"]}
    restore_sc = {"volBackendFs": data["localFs"],
                  "clusterId": data["id"], "nodeClass": "randomnodeclassx"}
    restore_pvc = {"access_modes": "ReadWriteMany",
                   "storage": "1Gi", "reason": "NotFound desc = nodeclass"}
    snapshot_object.test_static(value_sc, test_restore=True,
                                restore_sc=restore_sc, restore_pvc=restore_pvc)


def test_snapshot_dynamic_permissions_777_independent():
    LOGGER.warning("Testcase will fail if scale version < 5.1.1-4")
    value_pod = {"mount_path": "/usr/share/nginx/html/scale", "read_only": "False", "sub_path": ["sub_path_mnt"],
                 "volumemount_readonly": [False], "runAsUser": "2000", "runAsGroup": "5000"}
    value_sc = {"volBackendFs": data["localFs"], "clusterId": data["id"], "permissions": "777",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, value_pod=value_pod)


def test_snapshot_dynamic_volume_expansion_1():
    value_sc = {"volBackendFs": data["localFs"],
                "clusterId": data["id"], "allow_volume_expansion": True}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi", "presnap_volume_expansion_storage": ["2Gi"],
                  "post_presnap_volume_expansion_storage": ["5Gi", "15Gi"], "postsnap_volume_expansion_storage": ["10Gi", "15Gi"]}]
    snapshot_object.test_dynamic(value_sc, test_restore=True, value_pvc=value_pvc)


def test_snapshot_dynamic_volume_expansion_2():
    value_sc = {"volBackendFs": data["localFs"],
                "clusterId": data["id"], "allow_volume_expansion": True}
    restore_sc = {"volBackendFs": data["localFs"],
                  "filesetType": "dependent", "clusterId": data["id"], "allow_volume_expansion": True}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi", "presnap_volume_expansion_storage": ["3Gi"],
                  "post_presnap_volume_expansion_storage": ["5Gi", "12Gi"], "postsnap_volume_expansion_storage": ["8Gi", "12Gi"]}]
    snapshot_object.test_dynamic(value_sc, test_restore=True,
                                 value_pvc=value_pvc, restore_sc=restore_sc)


def test_snapshot_dynamic_volume_expansion_3():
    value_sc = {"volBackendFs": data["localFs"],
                "clusterId": data["id"], "allow_volume_expansion": True}
    restore_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                  "allow_volume_expansion": True}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi", "presnap_volume_expansion_storage": ["2Gi"],
                  "post_presnap_volume_expansion_storage": ["5Gi", "15Gi"], "postsnap_volume_expansion_storage": ["10Gi", "15Gi"]}]
    snapshot_object.test_dynamic(value_sc, test_restore=True,
                                 value_pvc=value_pvc, restore_sc=restore_sc)


def test_snapshot_dynamic_volume_cloning_1():
    value_sc = {"volBackendFs": data["localFs"], "clusterId": data["id"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}, {
        "access_modes": "ReadWriteOnce", "storage": "1Gi"}]}
    snapshot_object.test_dynamic(value_sc, test_restore=True,
                                 value_pvc=value_pvc, value_clone_passed=value_clone_passed)


def test_driver_volume_snapshot_Dependent_1_to_Dependent_1():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_1_to_Independent_1():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent"}
    restore_sc = {"volBackendFs": data["localFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_1_to_LW_1():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_1_to_Version2_1():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_1_to_Dependent_2():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_1_to_Independent_2():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent"}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_1_to_LW_2():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_1_to_Version2_2():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent"}

    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_1_to_Dependent_3():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent"}

    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_1_to_Independent_3():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent"}

    restore_sc = {
        "volBackendFs": data["localFs"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_1_to_LW_3():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_1_to_Version2_3():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_1_to_Dependent_4():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_1_to_Independent_4():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent"}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_1_to_LW_4():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_1_to_Version2_4():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_1_to_Dependent_5():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_1_to_Independent_5():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent"}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_1_to_LW_5():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_1_to_Version2_5():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_1_to_Dependent_1():
    value_sc = {"volBackendFs": data["localFs"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_1_to_Independent_1():
    value_sc = {"volBackendFs": data["localFs"]}
    restore_sc = {"volBackendFs": data["localFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_1_to_LW_1():
    value_sc = {"volBackendFs": data["localFs"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_1_to_Version2_1():
    value_sc = {"volBackendFs": data["localFs"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_1_to_Dependent_2():
    value_sc = {"volBackendFs": data["localFs"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_1_to_Independent_2():
    value_sc = {"volBackendFs": data["localFs"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_1_to_LW_2():
    value_sc = {"volBackendFs": data["localFs"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_1_to_Version2_2():
    value_sc = {"volBackendFs": data["localFs"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_1_to_Dependent_3():
    value_sc = {"volBackendFs": data["localFs"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_1_to_Independent_3():
    value_sc = {"volBackendFs": data["localFs"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_1_to_LW_3():
    value_sc = {"volBackendFs": data["localFs"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_1_to_Version2_3():
    value_sc = {"volBackendFs": data["localFs"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_1_to_Dependent_4():
    value_sc = {"volBackendFs": data["localFs"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_1_to_Independent_4():
    value_sc = {"volBackendFs": data["localFs"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_1_to_LW_4():
    value_sc = {"volBackendFs": data["localFs"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_1_to_Version2_4():
    value_sc = {"volBackendFs": data["localFs"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_1_to_Dependent_5():
    value_sc = {"volBackendFs": data["localFs"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_1_to_Independent_5():
    value_sc = {"volBackendFs": data["localFs"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_1_to_LW_5():
    value_sc = {"volBackendFs": data["localFs"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_1_to_Version2_5():
    value_sc = {"volBackendFs": data["localFs"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_1_to_Dependent_1():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_1_to_Independent_1():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    restore_sc = {"volBackendFs": data["localFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_LW_1_to_LW_1():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_1_to_Version2_1():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_1_to_Dependent_2():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_1_to_Independent_2():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_LW_1_to_LW_2():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_1_to_Version2_2():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_1_to_Dependent_3():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_1_to_Independent_3():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_LW_1_to_LW_3():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_1_to_Version2_3():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_1_to_Dependent_4():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_1_to_Independent_4():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_LW_1_to_LW_4():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_1_to_Version2_4():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_1_to_Dependent_5():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_1_to_Independent_5():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_LW_1_to_LW_5():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_1_to_Version2_5():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_1_to_Dependent_1():
    value_sc = {"volBackendFs": data["localFs"], "version": "2"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_1_to_Independent_1():
    value_sc = {"volBackendFs": data["localFs"], "version": "2"}
    restore_sc = {"volBackendFs": data["localFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_1_to_LW_1():
    value_sc = {"volBackendFs": data["localFs"], "version": "2"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_1_to_Version2_1():
    value_sc = {"volBackendFs": data["localFs"], "version": "2"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_1_to_Dependent_2():
    value_sc = {"volBackendFs": data["localFs"], "version": "2"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_1_to_Independent_2():
    value_sc = {"volBackendFs": data["localFs"], "version": "2"}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_1_to_LW_2():
    value_sc = {"volBackendFs": data["localFs"], "version": "2"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_1_to_Version2_2():
    value_sc = {"volBackendFs": data["localFs"], "version": "2"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_1_to_Dependent_3():
    value_sc = {"volBackendFs": data["localFs"], "version": "2"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_1_to_Independent_3():
    value_sc = {"volBackendFs": data["localFs"], "version": "2"}
    restore_sc = {
        "volBackendFs": data["localFs"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_1_to_LW_3():
    value_sc = {"volBackendFs": data["localFs"], "version": "2"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_1_to_Version2_3():
    value_sc = {"volBackendFs": data["localFs"], "version": "2"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_1_to_Dependent_4():
    value_sc = {"volBackendFs": data["localFs"], "version": "2"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_1_to_Independent_4():
    value_sc = {"volBackendFs": data["localFs"], "version": "2"}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_1_to_LW_4():
    value_sc = {"volBackendFs": data["localFs"], "version": "2"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_1_to_Version2_4():
    value_sc = {"volBackendFs": data["localFs"], "version": "2"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_1_to_Dependent_5():
    value_sc = {"volBackendFs": data["localFs"], "version": "2"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_1_to_Independent_5():
    value_sc = {"volBackendFs": data["localFs"], "version": "2"}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_1_to_LW_5():
    value_sc = {"volBackendFs": data["localFs"], "version": "2"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_1_to_Version2_5():
    value_sc = {"volBackendFs": data["localFs"], "version": "2"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_2_to_Dependent_1():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_2_to_Independent_1():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {"volBackendFs": data["localFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_2_to_LW_1():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_2_to_Version2_1():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_2_to_Dependent_2():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_2_to_Independent_2():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_2_to_LW_2():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_2_to_Version2_2():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_2_to_Dependent_3():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_2_to_Independent_3():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_2_to_LW_3():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_2_to_Version2_3():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_2_to_Dependent_4():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_2_to_Independent_4():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_2_to_LW_4():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_2_to_Version2_4():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_2_to_Dependent_5():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_2_to_Independent_5():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_2_to_LW_5():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_2_to_Version2_5():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_2_to_Dependent_1():
    value_sc = {"volBackendFs": data["localFs"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_2_to_Independent_1():
    value_sc = {"volBackendFs": data["localFs"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {"volBackendFs": data["localFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_2_to_LW_1():
    value_sc = {"volBackendFs": data["localFs"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_2_to_Version2_1():
    value_sc = {"volBackendFs": data["localFs"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_2_to_Dependent_2():
    value_sc = {"volBackendFs": data["localFs"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_2_to_Independent_2():
    value_sc = {"volBackendFs": data["localFs"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_2_to_LW_2():
    value_sc = {"volBackendFs": data["localFs"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_2_to_Version2_2():
    value_sc = {"volBackendFs": data["localFs"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_2_to_Dependent_3():
    value_sc = {"volBackendFs": data["localFs"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_2_to_Independent_3():
    value_sc = {"volBackendFs": data["localFs"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_2_to_LW_3():
    value_sc = {"volBackendFs": data["localFs"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_2_to_Version2_3():
    value_sc = {"volBackendFs": data["localFs"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_2_to_Dependent_4():
    value_sc = {"volBackendFs": data["localFs"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_2_to_Independent_4():
    value_sc = {"volBackendFs": data["localFs"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_2_to_LW_4():
    value_sc = {"volBackendFs": data["localFs"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_2_to_Version2_4():
    value_sc = {"volBackendFs": data["localFs"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_2_to_Dependent_5():
    value_sc = {"volBackendFs": data["localFs"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_2_to_Independent_5():
    value_sc = {"volBackendFs": data["localFs"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_2_to_LW_5():
    value_sc = {"volBackendFs": data["localFs"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_2_to_Version2_5():
    value_sc = {"volBackendFs": data["localFs"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_2_to_Dependent_1():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_2_to_Independent_1():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {"volBackendFs": data["localFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_LW_2_to_LW_1():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_2_to_Version2_1():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_2_to_Dependent_2():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_2_to_Independent_2():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_LW_2_to_LW_2():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_2_to_Version2_2():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_2_to_Dependent_3():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_2_to_Independent_3():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_LW_2_to_LW_3():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_2_to_Version2_3():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_2_to_Dependent_4():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_2_to_Independent_4():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_LW_2_to_LW_4():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_2_to_Version2_4():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_2_to_Dependent_5():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_2_to_Independent_5():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_LW_2_to_LW_5():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_2_to_Version2_5():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_2_to_Dependent_1():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_2_to_Independent_1():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {"volBackendFs": data["localFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_2_to_LW_1():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_2_to_Version2_1():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_2_to_Dependent_2():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_2_to_Independent_2():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_2_to_LW_2():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_2_to_Version2_2():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_2_to_Dependent_3():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_2_to_Independent_3():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_2_to_LW_3():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_2_to_Version2_3():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_2_to_Dependent_4():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_2_to_Independent_4():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_2_to_LW_4():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_2_to_Version2_4():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_2_to_Dependent_5():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_2_to_Independent_5():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_2_to_LW_5():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_2_to_Version2_5():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_3_to_Dependent_1():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_3_to_Independent_1():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent", "shared": "True"}
    restore_sc = {"volBackendFs": data["localFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_3_to_LW_1():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_3_to_Version2_1():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_3_to_Dependent_2():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_3_to_Independent_2():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_3_to_LW_2():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_3_to_Version2_2():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_3_to_Dependent_3():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_3_to_Independent_3():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_3_to_LW_3():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_3_to_Version2_3():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_3_to_Dependent_4():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_3_to_Independent_4():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_3_to_LW_4():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_3_to_Version2_4():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_3_to_Dependent_5():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_3_to_Independent_5():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_3_to_LW_5():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_3_to_Version2_5():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_3_to_Dependent_1():
    value_sc = {"volBackendFs": data["localFs"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_3_to_Independent_1():
    value_sc = {"volBackendFs": data["localFs"], "shared": "True"}
    restore_sc = {"volBackendFs": data["localFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_3_to_LW_1():
    value_sc = {"volBackendFs": data["localFs"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_3_to_Version2_1():
    value_sc = {"volBackendFs": data["localFs"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_3_to_Dependent_2():
    value_sc = {"volBackendFs": data["localFs"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_3_to_Independent_2():
    value_sc = {"volBackendFs": data["localFs"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_3_to_LW_2():
    value_sc = {"volBackendFs": data["localFs"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_3_to_Version2_2():
    value_sc = {"volBackendFs": data["localFs"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_3_to_Dependent_3():
    value_sc = {"volBackendFs": data["localFs"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_3_to_Independent_3():
    value_sc = {"volBackendFs": data["localFs"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_3_to_LW_3():
    value_sc = {"volBackendFs": data["localFs"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_3_to_Version2_3():
    value_sc = {"volBackendFs": data["localFs"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_3_to_Dependent_4():
    value_sc = {"volBackendFs": data["localFs"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_3_to_Independent_4():
    value_sc = {"volBackendFs": data["localFs"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_3_to_LW_4():
    value_sc = {"volBackendFs": data["localFs"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_3_to_Version2_4():
    value_sc = {"volBackendFs": data["localFs"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_3_to_Dependent_5():
    value_sc = {"volBackendFs": data["localFs"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_3_to_Independent_5():
    value_sc = {"volBackendFs": data["localFs"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_3_to_LW_5():
    value_sc = {"volBackendFs": data["localFs"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_3_to_Version2_5():
    value_sc = {"volBackendFs": data["localFs"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_3_to_Dependent_1():
    value_sc = {"volBackendFs": data["localFs"],
                "volDirBasePath": data["volDirBasePath"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_3_to_Independent_1():
    value_sc = {"volBackendFs": data["localFs"],
                "volDirBasePath": data["volDirBasePath"], "shared": "True"}
    restore_sc = {"volBackendFs": data["localFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_LW_3_to_LW_1():
    value_sc = {"volBackendFs": data["localFs"],
                "volDirBasePath": data["volDirBasePath"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_3_to_Version2_1():
    value_sc = {"volBackendFs": data["localFs"],
                "volDirBasePath": data["volDirBasePath"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_3_to_Dependent_2():
    value_sc = {"volBackendFs": data["localFs"],
                "volDirBasePath": data["volDirBasePath"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_3_to_Independent_2():
    value_sc = {"volBackendFs": data["localFs"],
                "volDirBasePath": data["volDirBasePath"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_LW_3_to_LW_2():
    value_sc = {"volBackendFs": data["localFs"],
                "volDirBasePath": data["volDirBasePath"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_3_to_Version2_2():
    value_sc = {"volBackendFs": data["localFs"],
                "volDirBasePath": data["volDirBasePath"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_3_to_Dependent_3():
    value_sc = {"volBackendFs": data["localFs"],
                "volDirBasePath": data["volDirBasePath"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_3_to_Independent_3():
    value_sc = {"volBackendFs": data["localFs"],
                "volDirBasePath": data["volDirBasePath"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_LW_3_to_LW_3():
    value_sc = {"volBackendFs": data["localFs"],
                "volDirBasePath": data["volDirBasePath"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_3_to_Version2_3():
    value_sc = {"volBackendFs": data["localFs"],
                "volDirBasePath": data["volDirBasePath"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_3_to_Dependent_4():
    value_sc = {"volBackendFs": data["localFs"],
                "volDirBasePath": data["volDirBasePath"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_3_to_Independent_4():
    value_sc = {"volBackendFs": data["localFs"],
                "volDirBasePath": data["volDirBasePath"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_LW_3_to_LW_4():
    value_sc = {"volBackendFs": data["localFs"],
                "volDirBasePath": data["volDirBasePath"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_3_to_Version2_4():
    value_sc = {"volBackendFs": data["localFs"],
                "volDirBasePath": data["volDirBasePath"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_3_to_Dependent_5():
    value_sc = {"volBackendFs": data["localFs"],
                "volDirBasePath": data["volDirBasePath"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_3_to_Independent_5():
    value_sc = {"volBackendFs": data["localFs"],
                "volDirBasePath": data["volDirBasePath"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_LW_3_to_LW_5():
    value_sc = {"volBackendFs": data["localFs"],
                "volDirBasePath": data["volDirBasePath"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_3_to_Version2_5():
    value_sc = {"volBackendFs": data["localFs"],
                "volDirBasePath": data["volDirBasePath"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_3_to_Dependent_1():
    value_sc = {"volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_3_to_Independent_1():
    value_sc = {"volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    restore_sc = {"volBackendFs": data["localFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_3_to_LW_1():
    value_sc = {"volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_3_to_Version2_1():
    value_sc = {"volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_3_to_Dependent_2():
    value_sc = {"volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_3_to_Independent_2():
    value_sc = {"volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_3_to_LW_2():
    value_sc = {"volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_3_to_Version2_2():
    value_sc = {"volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_3_to_Dependent_3():
    value_sc = {"volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_3_to_Independent_3():
    value_sc = {"volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_3_to_LW_3():
    value_sc = {"volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_3_to_Version2_3():
    value_sc = {"volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_3_to_Dependent_4():
    value_sc = {"volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_3_to_Independent_4():
    value_sc = {"volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_3_to_LW_4():
    value_sc = {"volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_3_to_Version2_4():
    value_sc = {"volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_3_to_Dependent_5():
    value_sc = {"volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_3_to_Independent_5():
    value_sc = {"volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_3_to_LW_5():
    value_sc = {"volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_3_to_Version2_5():
    value_sc = {"volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_4_to_Dependent_1():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_4_to_Independent_1():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {"volBackendFs": data["localFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_4_to_LW_1():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_4_to_Version2_1():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}

    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_4_to_Dependent_2():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}

    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_4_to_Independent_2():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_4_to_LW_2():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_4_to_Version2_2():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_4_to_Dependent_3():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_4_to_Independent_3():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_4_to_LW_3():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_4_to_Version2_3():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_4_to_Dependent_4():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_4_to_Independent_4():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_4_to_LW_4():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_4_to_Version2_4():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_4_to_Dependent_5():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_4_to_Independent_5():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_4_to_LW_5():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_4_to_Version2_5():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_4_to_Dependent_1():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_4_to_Independent_1():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "shared": "True"}
    restore_sc = {"volBackendFs": data["localFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_4_to_LW_1():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_4_to_Version2_1():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_4_to_Dependent_2():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_4_to_Independent_2():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_4_to_LW_2():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_4_to_Version2_2():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_4_to_Dependent_3():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_4_to_Independent_3():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_4_to_LW_3():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_4_to_Version2_3():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_4_to_Dependent_4():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_4_to_Independent_4():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_4_to_LW_4():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_4_to_Version2_4():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_4_to_Dependent_5():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_4_to_Independent_5():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_4_to_LW_5():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_4_to_Version2_5():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_4_to_Dependent_1():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_4_to_Independent_1():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {"volBackendFs": data["localFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_LW_4_to_LW_1():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_4_to_Version2_1():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_4_to_Dependent_2():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_4_to_Independent_2():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_LW_4_to_LW_2():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_4_to_Version2_2():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_4_to_Dependent_3():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_4_to_Independent_3():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_LW_4_to_LW_3():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_4_to_Version2_3():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_4_to_Dependent_4():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_4_to_Independent_4():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_LW_4_to_LW_4():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_4_to_Version2_4():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_4_to_Dependent_5():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_4_to_Independent_5():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_LW_4_to_LW_5():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_4_to_Version2_5():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_4_to_Dependent_1():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_4_to_Independent_1():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {"volBackendFs": data["localFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_4_to_LW_1():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_4_to_Version2_1():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_4_to_Dependent_2():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_4_to_Independent_2():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_4_to_LW_2():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_4_to_Version2_2():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_4_to_Dependent_3():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_4_to_Independent_3():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_4_to_LW_3():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_4_to_Version2_3():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_4_to_Dependent_4():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_4_to_Independent_4():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_4_to_LW_4():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_4_to_Version2_4():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_4_to_Dependent_5():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_4_to_Independent_5():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_4_to_LW_5():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_4_to_Version2_5():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_5_to_Dependent_1():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_5_to_Independent_1():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {"volBackendFs": data["localFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_5_to_LW_1():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_5_to_Version2_1():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_5_to_Dependent_2():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_5_to_Independent_2():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_5_to_LW_2():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_5_to_Version2_2():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_5_to_Dependent_3():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_5_to_Independent_3():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_5_to_LW_3():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_5_to_Version2_3():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_5_to_Dependent_4():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_5_to_Independent_4():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_5_to_LW_4():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_5_to_Version2_4():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_5_to_Dependent_5():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_5_to_Independent_5():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_5_to_LW_5():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Dependent_5_to_Version2_5():
    value_sc = {"volBackendFs": data["localFs"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_5_to_Dependent_1():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_5_to_Independent_1():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {"volBackendFs": data["localFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_5_to_LW_1():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_5_to_Version2_1():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_5_to_Dependent_2():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_5_to_Independent_2():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_5_to_LW_2():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_5_to_Version2_2():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_5_to_Dependent_3():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_5_to_Independent_3():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_5_to_LW_3():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_5_to_Version2_3():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_5_to_Dependent_4():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_5_to_Independent_4():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_5_to_LW_4():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_5_to_Version2_4():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_5_to_Dependent_5():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_5_to_Independent_5():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_5_to_LW_5():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_5_to_Version2_5():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_5_to_Dependent_1():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_5_to_Independent_1():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {"volBackendFs": data["localFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_LW_5_to_LW_1():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_5_to_Version2_1():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_5_to_Dependent_2():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_5_to_Independent_2():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_LW_5_to_LW_2():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_5_to_Version2_2():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_5_to_Dependent_3():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_5_to_Independent_3():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_LW_5_to_LW_3():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_5_to_Version2_3():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_5_to_Dependent_4():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_5_to_Independent_4():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_LW_5_to_LW_4():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_5_to_Version2_4():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_5_to_Dependent_5():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_5_to_Independent_5():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_LW_5_to_LW_5():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


@pytest.mark.xfail
def test_driver_volume_snapshot_LW_5_to_Version2_5():
    value_sc = {"volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_5_to_Dependent_1():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_5_to_Independent_1():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {"volBackendFs": data["localFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_5_to_LW_1():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_5_to_Version2_1():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_5_to_Dependent_2():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_5_to_Independent_2():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_5_to_LW_2():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_5_to_Version2_2():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_5_to_Dependent_3():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_5_to_Independent_3():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_5_to_LW_3():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_5_to_Version2_3():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_5_to_Dependent_4():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_5_to_Independent_4():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_5_to_LW_4():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_5_to_Version2_4():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_5_to_Dependent_5():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "filesetType": "dependent", "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_5_to_Independent_5():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_5_to_LW_5():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "volDirBasePath": data["volDirBasePath"], "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_5_to_Version2_5():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["localFs"], "version": "2", "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)
