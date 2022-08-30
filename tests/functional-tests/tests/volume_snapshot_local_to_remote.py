import logging
import pytest
import ibm_spectrum_scale_csi.base_class as baseclass
import ibm_spectrum_scale_csi.common_utils.input_data_functions as inputfunc

LOGGER = logging.getLogger()
pytestmark = [pytest.mark.volumesnapshot, pytest.mark.crosscluster]


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

'''
def test_driver_volume_snapshot_Independent_1_to_Dependent_1():
    value_sc = {"volBackendFs": data["localFs"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_1_to_Independent_1():
    value_sc = {"volBackendFs": data["localFs"]}
    restore_sc = {"volBackendFs": data["remoteFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_1_to_LW_1():
    value_sc = {"volBackendFs": data["localFs"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_1_to_Version2_1():
    value_sc = {"volBackendFs": data["localFs"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_1_to_Dependent_2():
    value_sc = {"volBackendFs": data["localFs"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_1_to_Independent_2():
    value_sc = {"volBackendFs": data["localFs"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_1_to_LW_2():
    value_sc = {"volBackendFs": data["localFs"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_1_to_Version2_2():
    value_sc = {"volBackendFs": data["localFs"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_1_to_Dependent_3():
    value_sc = {"volBackendFs": data["localFs"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_1_to_Independent_3():
    value_sc = {"volBackendFs": data["localFs"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_1_to_LW_3():
    value_sc = {"volBackendFs": data["localFs"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_1_to_Version2_3():
    value_sc = {"volBackendFs": data["localFs"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_1_to_Dependent_4():
    value_sc = {"volBackendFs": data["localFs"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_1_to_Independent_4():
    value_sc = {"volBackendFs": data["localFs"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_1_to_LW_4():
    value_sc = {"volBackendFs": data["localFs"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_1_to_Version2_4():
    value_sc = {"volBackendFs": data["localFs"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_1_to_Dependent_5():
    value_sc = {"volBackendFs": data["localFs"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_1_to_Independent_5():
    value_sc = {"volBackendFs": data["localFs"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_1_to_LW_5():
    value_sc = {"volBackendFs": data["localFs"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_1_to_Version2_5():
    value_sc = {"volBackendFs": data["localFs"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_1_to_Dependent_1():
    value_sc = {"volBackendFs": data["localFs"], "version": "2"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_1_to_Independent_1():
    value_sc = {"volBackendFs": data["localFs"], "version": "2"}
    restore_sc = {"volBackendFs": data["remoteFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_1_to_LW_1():
    value_sc = {"volBackendFs": data["localFs"], "version": "2"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_1_to_Version2_1():
    value_sc = {"volBackendFs": data["localFs"], "version": "2"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_1_to_Dependent_2():
    value_sc = {"volBackendFs": data["localFs"], "version": "2"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_1_to_Independent_2():
    value_sc = {"volBackendFs": data["localFs"], "version": "2"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_1_to_LW_2():
    value_sc = {"volBackendFs": data["localFs"], "version": "2"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_1_to_Version2_2():
    value_sc = {"volBackendFs": data["localFs"], "version": "2"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_1_to_Dependent_3():
    value_sc = {"volBackendFs": data["localFs"], "version": "2"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_1_to_Independent_3():
    value_sc = {"volBackendFs": data["localFs"], "version": "2"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_1_to_LW_3():
    value_sc = {"volBackendFs": data["localFs"], "version": "2"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_1_to_Version2_3():
    value_sc = {"volBackendFs": data["localFs"], "version": "2"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_1_to_Dependent_4():
    value_sc = {"volBackendFs": data["localFs"], "version": "2"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_1_to_Independent_4():
    value_sc = {"volBackendFs": data["localFs"], "version": "2"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_1_to_LW_4():
    value_sc = {"volBackendFs": data["localFs"], "version": "2"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_1_to_Version2_4():
    value_sc = {"volBackendFs": data["localFs"], "version": "2"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_1_to_Dependent_5():
    value_sc = {"volBackendFs": data["localFs"], "version": "2"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_1_to_Independent_5():
    value_sc = {"volBackendFs": data["localFs"], "version": "2"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_1_to_LW_5():
    value_sc = {"volBackendFs": data["localFs"], "version": "2"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_1_to_Version2_5():
    value_sc = {"volBackendFs": data["localFs"], "version": "2"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_2_to_Dependent_1():
    value_sc = {"volBackendFs": data["localFs"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_2_to_Independent_1():
    value_sc = {"volBackendFs": data["localFs"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {"volBackendFs": data["remoteFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_2_to_LW_1():
    value_sc = {"volBackendFs": data["localFs"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_2_to_Version2_1():
    value_sc = {"volBackendFs": data["localFs"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_2_to_Dependent_2():
    value_sc = {"volBackendFs": data["localFs"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_2_to_Independent_2():
    value_sc = {"volBackendFs": data["localFs"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_2_to_LW_2():
    value_sc = {"volBackendFs": data["localFs"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_2_to_Version2_2():
    value_sc = {"volBackendFs": data["localFs"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_2_to_Dependent_3():
    value_sc = {"volBackendFs": data["localFs"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_2_to_Independent_3():
    value_sc = {"volBackendFs": data["localFs"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_2_to_LW_3():
    value_sc = {"volBackendFs": data["localFs"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_2_to_Version2_3():
    value_sc = {"volBackendFs": data["localFs"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_2_to_Dependent_4():
    value_sc = {"volBackendFs": data["localFs"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_2_to_Independent_4():
    value_sc = {"volBackendFs": data["localFs"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_2_to_LW_4():
    value_sc = {"volBackendFs": data["localFs"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_2_to_Version2_4():
    value_sc = {"volBackendFs": data["localFs"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_2_to_Dependent_5():
    value_sc = {"volBackendFs": data["localFs"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_2_to_Independent_5():
    value_sc = {"volBackendFs": data["localFs"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_2_to_LW_5():
    value_sc = {"volBackendFs": data["localFs"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_2_to_Version2_5():
    value_sc = {"volBackendFs": data["localFs"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_2_to_Dependent_1():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_2_to_Independent_1():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {"volBackendFs": data["remoteFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_2_to_LW_1():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_2_to_Version2_1():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_2_to_Dependent_2():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_2_to_Independent_2():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_2_to_LW_2():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_2_to_Version2_2():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_2_to_Dependent_3():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_2_to_Independent_3():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_2_to_LW_3():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_2_to_Version2_3():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_2_to_Dependent_4():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_2_to_Independent_4():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_2_to_LW_4():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_2_to_Version2_4():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_2_to_Dependent_5():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_2_to_Independent_5():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_2_to_LW_5():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_2_to_Version2_5():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_3_to_Dependent_1():
    value_sc = {"volBackendFs": data["localFs"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_3_to_Independent_1():
    value_sc = {"volBackendFs": data["localFs"], "shared": "True"}
    restore_sc = {"volBackendFs": data["remoteFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_3_to_LW_1():
    value_sc = {"volBackendFs": data["localFs"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_3_to_Version2_1():
    value_sc = {"volBackendFs": data["localFs"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_3_to_Dependent_2():
    value_sc = {"volBackendFs": data["localFs"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_3_to_Independent_2():
    value_sc = {"volBackendFs": data["localFs"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_3_to_LW_2():
    value_sc = {"volBackendFs": data["localFs"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_3_to_Version2_2():
    value_sc = {"volBackendFs": data["localFs"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_3_to_Dependent_3():
    value_sc = {"volBackendFs": data["localFs"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_3_to_Independent_3():
    value_sc = {"volBackendFs": data["localFs"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_3_to_LW_3():
    value_sc = {"volBackendFs": data["localFs"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_3_to_Version2_3():
    value_sc = {"volBackendFs": data["localFs"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_3_to_Dependent_4():
    value_sc = {"volBackendFs": data["localFs"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_3_to_Independent_4():
    value_sc = {"volBackendFs": data["localFs"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_3_to_LW_4():
    value_sc = {"volBackendFs": data["localFs"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_3_to_Version2_4():
    value_sc = {"volBackendFs": data["localFs"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_3_to_Dependent_5():
    value_sc = {"volBackendFs": data["localFs"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_3_to_Independent_5():
    value_sc = {"volBackendFs": data["localFs"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_3_to_LW_5():
    value_sc = {"volBackendFs": data["localFs"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_3_to_Version2_5():
    value_sc = {"volBackendFs": data["localFs"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_3_to_Dependent_1():
    value_sc = {"volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_3_to_Independent_1():
    value_sc = {"volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    restore_sc = {"volBackendFs": data["remoteFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_3_to_LW_1():
    value_sc = {"volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_3_to_Version2_1():
    value_sc = {"volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_3_to_Dependent_2():
    value_sc = {"volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_3_to_Independent_2():
    value_sc = {"volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_3_to_LW_2():
    value_sc = {"volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_3_to_Version2_2():
    value_sc = {"volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_3_to_Dependent_3():
    value_sc = {"volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_3_to_Independent_3():
    value_sc = {"volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_3_to_LW_3():
    value_sc = {"volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_3_to_Version2_3():
    value_sc = {"volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_3_to_Dependent_4():
    value_sc = {"volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_3_to_Independent_4():
    value_sc = {"volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_3_to_LW_4():
    value_sc = {"volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_3_to_Version2_4():
    value_sc = {"volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_3_to_Dependent_5():
    value_sc = {"volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_3_to_Independent_5():
    value_sc = {"volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_3_to_LW_5():
    value_sc = {"volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_3_to_Version2_5():
    value_sc = {"volBackendFs": data["localFs"], "version": "2", "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_4_to_Dependent_1():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_4_to_Independent_1():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "shared": "True"}
    restore_sc = {"volBackendFs": data["remoteFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_4_to_LW_1():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_4_to_Version2_1():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_4_to_Dependent_2():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_4_to_Independent_2():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_4_to_LW_2():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_4_to_Version2_2():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_4_to_Dependent_3():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_4_to_Independent_3():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_4_to_LW_3():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_4_to_Version2_3():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_4_to_Dependent_4():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_4_to_Independent_4():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_4_to_LW_4():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_4_to_Version2_4():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_4_to_Dependent_5():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_4_to_Independent_5():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_4_to_LW_5():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_4_to_Version2_5():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_4_to_Dependent_1():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_4_to_Independent_1():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {"volBackendFs": data["remoteFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_4_to_LW_1():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_4_to_Version2_1():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_4_to_Dependent_2():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_4_to_Independent_2():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_4_to_LW_2():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_4_to_Version2_2():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_4_to_Dependent_3():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_4_to_Independent_3():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_4_to_LW_3():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_4_to_Version2_3():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_4_to_Dependent_4():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_4_to_Independent_4():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_4_to_LW_4():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_4_to_Version2_4():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_4_to_Dependent_5():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_4_to_Independent_5():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_4_to_LW_5():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_4_to_Version2_5():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "shared": "True"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_5_to_Dependent_1():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_5_to_Independent_1():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {"volBackendFs": data["remoteFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_5_to_LW_1():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_5_to_Version2_1():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_5_to_Dependent_2():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_5_to_Independent_2():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_5_to_LW_2():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_5_to_Version2_2():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_5_to_Dependent_3():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_5_to_Independent_3():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_5_to_LW_3():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_5_to_Version2_3():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_5_to_Dependent_4():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_5_to_Independent_4():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_5_to_LW_4():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_5_to_Version2_4():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_5_to_Dependent_5():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_5_to_Independent_5():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_5_to_LW_5():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Independent_5_to_Version2_5():
    value_sc = {"volBackendFs": data["localFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_5_to_Dependent_1():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_5_to_Independent_1():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {"volBackendFs": data["remoteFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_5_to_LW_1():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_5_to_Version2_1():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_5_to_Dependent_2():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_5_to_Independent_2():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_5_to_LW_2():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_5_to_Version2_2():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_5_to_Dependent_3():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_5_to_Independent_3():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_5_to_LW_3():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_5_to_Version2_3():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2", "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_5_to_Dependent_4():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_5_to_Independent_4():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_5_to_LW_4():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_5_to_Version2_4():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "shared": "True"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_5_to_Dependent_5():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "filesetType": "dependent", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_5_to_Independent_5():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_5_to_LW_5():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "r_volDirBasePath": data["r_volDirBasePath"], "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)


def test_driver_volume_snapshot_Version2_5_to_Version2_5():
    value_sc = {"volBackendFs": data["localFs"], "version": "2",
                "gid": data["gid_number"], "uid": data["uid_number"], "permissions": "755"}
    restore_sc = {
        "volBackendFs": data["remoteFs"], "version": "2", "gid": data["r_gid_number"], "uid": data["r_uid_number"], "permissions": "755"}
    snapshot_object.test_dynamic(value_sc, test_restore=True, restore_sc=restore_sc)
'''