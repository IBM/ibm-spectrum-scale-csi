import copy
import logging
import pytest
import utils.fileset_functions as ff
import scale_operator as scaleop
LOGGER = logging.getLogger()


@pytest.fixture(scope='session', autouse=True)
def values(request):
    global data, remote_data, snapshot_object, kubeconfig_value  # are required in every testcase
    kubeconfig_value, clusterconfig_value, operator_namespace, test_namespace, runslow_val = scaleop.get_cmd_values(request)

    data = scaleop.read_driver_data(clusterconfig_value, test_namespace)
    operator_data = scaleop.read_operator_data(clusterconfig_value, operator_namespace)
    keep_objects = data["keepobjects"]
    if not("remote" in data):
        LOGGER.error("remote data is not provided in cr file")
        assert False

    remote_data = get_remote_data(data)
    ff.cred_check(data)
    ff.cred_check(remote_data)
    ff.set_data(remote_data)

    operator = scaleop.Scaleoperator(kubeconfig_value, operator_namespace)
    operator_object = scaleop.Scaleoperatorobject(operator_data, kubeconfig_value)
    condition = scaleop.check_ns_exists(kubeconfig_value, operator_namespace)
    if condition is True:
        if not(operator_object.check()):
            LOGGER.error("Operator custom object is not deployed succesfully")
            assert False
    else:
        operator.create()
        operator.check()
        scaleop.check_nodes_available(operator_data["pluginNodeSelector"], "pluginNodeSelector")
        scaleop.check_nodes_available(
            operator_data["provisionerNodeSelector"], "provisionerNodeSelector")
        scaleop.check_nodes_available(
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
    snapshot_object = scaleop.Snapshot(kubeconfig_value, test_namespace, keep_objects, value_pvc, value_vs_class,
                               number_of_snapshots, data["image_name"], remote_data["id"])
    ff.create_dir(remote_data["volDirBasePath"])
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
    remote_data["port"] = remote_data["remote_port"]
    for cluster in remote_data["clusters"]:
        if cluster["id"] == remote_data["remoteid"]:
            remote_data["guiHost"] = cluster["restApi"][0]["guiHost"]
            remote_sec_name = cluster["secrets"]
            remote_data["username"] = remote_data["remote_username"][remote_sec_name]
            remote_data["password"] = remote_data["remote_password"][remote_sec_name]

    remote_data["volDirBasePath"] = remote_data["r_volDirBasePath"]
    remote_data["parentFileset"] = remote_data["r_parentFileset"]
    remote_data["gid_name"] = remote_data["r_gid_name"]
    remote_data["uid_name"] = remote_data["r_uid_name"]
    remote_data["gid_number"] = remote_data["r_gid_number"]
    remote_data["uid_number"] = remote_data["r_uid_number"]
    remote_data["inodeLimit"] = remote_data["r_inodeLimit"]
    # for get_mount_point function
    remote_data["type_remote"] = {"username": data_passed["username"],
                                  "password": data_passed["password"],
                                  "port": data_passed["port"],
                                  "guiHost": data_passed["guiHost"]}

    return remote_data


def test_get_version():
    LOGGER.info("Remote Cluster Details:")
    LOGGER.info("-----------------------")
    ff.get_scale_version(remote_data)
    LOGGER.info("Local Cluster Details:")
    LOGGER.info("-----------------------")
    ff.get_scale_version(data)
    scaleop.get_kubernetes_version(kubeconfig_value)
    scaleop.scale_function.get_operator_image()
    scaleop.ob.get_driver_image()


def test_snapshot_static_pass_1():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_multiple_snapshots():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    snapshot_object.test_static(value_sc, test_restore=True, number_of_snapshots=3)


def test_snapshot_static_pass_3():
    value_sc = {"volBackendFs": data["remoteFs"],
                "clusterId": data["remoteid"], "gid": data["r_gid_number"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_4():
    value_sc = {"volBackendFs": data["remoteFs"],
                "clusterId": data["remoteid"], "inodeLimit": data["r_inodeLimit"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_5():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"],
                "inodeLimit": data["r_inodeLimit"], "uid": data["r_uid_number"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_6():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_7():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"],
                "inodeLimit": data["r_inodeLimit"], "gid": data["r_gid_number"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_8():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"],
                "inodeLimit": data["r_inodeLimit"], "uid": data["r_uid_number"],
                "gid": data["r_gid_number"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_9():
    value_sc = {"volBackendFs": data["remoteFs"],
                "clusterId": data["remoteid"], "uid": data["r_uid_number"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_10():
    value_sc = {"volBackendFs": data["remoteFs"],
                "inodeLimit": data["r_inodeLimit"],
                "clusterId": data["remoteid"], "filesetType": "independent"}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_11():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "clusterId": data["remoteid"],
                "filesetType": "independent", "inodeLimit": data["r_inodeLimit"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_12():
    value_sc = {"clusterId": data["remoteid"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "volBackendFs": data["remoteFs"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_13():
    value_sc = {"clusterId": data["remoteid"], "uid": data["r_uid_number"],
                "inodeLimit": data["r_inodeLimit"],
                "volBackendFs": data["remoteFs"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_14():
    value_sc = {"clusterId": data["remoteid"], "gid": data["r_gid_number"],
                "inodeLimit": data["r_inodeLimit"],
                "volBackendFs": data["remoteFs"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_15():
    value_sc = {"clusterId": data["remoteid"], "volBackendFs": data["remoteFs"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"],
                "inodeLimit": data["r_inodeLimit"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_16():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    snapshot_object.test_static(value_sc, test_restore=False)


def test_snapshot_static_pass_17():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "clusterId": data["remoteid"],
                "filesetType": "independent", "inodeLimit": data["r_inodeLimit"]}
    snapshot_object.test_static(value_sc, test_restore=False)


def test_snapshot_static_pass_18():
    value_sc = {"clusterId": data["remoteid"], "gid": data["r_gid_number"],
                "inodeLimit": data["r_inodeLimit"],
                "volBackendFs": data["remoteFs"]}
    snapshot_object.test_static(value_sc, test_restore=False)


def test_snapshot_dynamic_pass_1():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_2():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, value_vs_class={"deletionPolicy": "Retain"})


def test_snapshot_dynamic_expected_fail_1():
    value_sc = {"volBackendFs": data["remoteFs"],
                "filesetType": "dependent", "clusterId": data["remoteid"]}
    snapshot_object.test_dynamic(value_sc, test_restore=False, reason="Volume snapshot can only be created when source volume is independent fileset")


def test_snapshot_dynamic_expected_fail_2():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r_volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc, test_restore=False, reason="Volume snapshot can only be created when source volume is independent fileset")


def test_snapshot_dynamic_multiple_snapshots():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, number_of_snapshots=3)


@pytest.mark.slow
def test_snapshot_dynamic_multiple_snapshots_256():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, number_of_snapshots=256)


@pytest.mark.slow
def test_snapshot_dynamic_multiple_snapshots_257():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, number_of_snapshots=257)


def test_snapshot_dynamic_pass_3():
    value_sc = {"volBackendFs": data["remoteFs"],
                "clusterId": data["remoteid"], "gid": data["r_gid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_4():
    value_sc = {"volBackendFs": data["remoteFs"],
                "clusterId": data["remoteid"], "inodeLimit": data["r_inodeLimit"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_5():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"],
                "inodeLimit": data["r_inodeLimit"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_6():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_7():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"],
                "inodeLimit": data["r_inodeLimit"], "gid": data["r_gid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_8():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"],
                "inodeLimit": data["r_inodeLimit"], "uid": data["r_uid_number"],
                "gid": data["r_gid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_9():
    value_sc = {"volBackendFs": data["remoteFs"],
                "clusterId": data["remoteid"], "uid": data["r_uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_10():
    value_sc = {"volBackendFs": data["remoteFs"],
                "inodeLimit": data["r_inodeLimit"],
                "clusterId": data["remoteid"], "filesetType": "independent"}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_11():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "clusterId": data["remoteid"],
                "filesetType": "independent", "inodeLimit": data["r_inodeLimit"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_12():
    value_sc = {"clusterId": data["remoteid"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "volBackendFs": data["remoteFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_13():
    value_sc = {"clusterId": data["remoteid"], "uid": data["r_uid_number"],
                "inodeLimit": data["r_inodeLimit"],
                "volBackendFs": data["remoteFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_14():
    value_sc = {"clusterId": data["remoteid"], "gid": data["r_gid_number"],
                "inodeLimit": data["r_inodeLimit"],
                "volBackendFs": data["remoteFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_15():
    value_sc = {"clusterId": data["remoteid"], "volBackendFs": data["remoteFs"],
                "gid": data["r_gid_number"], "uid": data["r_uid_number"],
                "inodeLimit": data["r_inodeLimit"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_16():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    snapshot_object.test_dynamic(value_sc, test_restore=False)


def test_snapshot_dynamic_pass_17():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r_gid_number"],
                "uid": data["r_uid_number"], "clusterId": data["remoteid"],
                "filesetType": "independent", "inodeLimit": data["r_inodeLimit"]}
    snapshot_object.test_dynamic(value_sc, test_restore=False)


def test_snapshot_dynamic_pass_18():
    value_sc = {"clusterId": data["remoteid"], "gid": data["r_gid_number"],
                "inodeLimit": data["r_inodeLimit"],
                "volBackendFs": data["remoteFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=False)


def test_snapshot_dynamic_different_sc_1():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    restore_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True,restore_sc=restore_sc)


def test_snapshot_dynamic_different_sc_2():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    restore_sc = {"volBackendFs": data["remoteFs"],
                "filesetType": "dependent", "clusterId": data["remoteid"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True,restore_sc=restore_sc)


def test_snapshot_dynamic_different_sc_3():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"],
                "inodeLimit": data["r_inodeLimit"], "uid": data["r_uid_number"],
                "gid": data["r_gid_number"]}
    restore_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True,restore_sc=restore_sc)


def test_snapshot_dynamic_different_sc_4():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"],
                "inodeLimit": data["r_inodeLimit"], "uid": data["r_uid_number"],
                "gid": data["r_gid_number"]}
    restore_sc = {"volBackendFs": data["remoteFs"],
                "filesetType": "dependent", "clusterId": data["remoteid"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True,restore_sc=restore_sc)


def test_snapshot_dynamic_different_sc_5():
    value_sc = {"volBackendFs": data["remoteFs"],
                "inodeLimit": data["r_inodeLimit"],
                "clusterId": data["remoteid"], "filesetType": "independent"}
    restore_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True,restore_sc=restore_sc)


def test_snapshot_dynamic_different_sc_6():
    value_sc = {"volBackendFs": data["remoteFs"],
                "inodeLimit": data["r_inodeLimit"],
                "clusterId": data["remoteid"], "filesetType": "independent"}
    restore_sc = {"volBackendFs": data["remoteFs"],
                "filesetType": "dependent", "clusterId": data["remoteid"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True,restore_sc=restore_sc)


def test_snapshot_static_different_sc_1():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    restore_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"]}
    snapshot_object.test_static(value_sc, test_restore=True,restore_sc=restore_sc)


def test_snapshot_static_different_sc_2():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    restore_sc = {"volBackendFs": data["remoteFs"],
                "filesetType": "dependent", "clusterId": data["remoteid"]}
    snapshot_object.test_static(value_sc, test_restore=True,restore_sc=restore_sc)


def test_snapshot_static_different_sc_3():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"],
                "inodeLimit": data["r_inodeLimit"], "uid": data["r_uid_number"],
                "gid": data["r_gid_number"]}
    restore_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"]}
    snapshot_object.test_static(value_sc, test_restore=True,restore_sc=restore_sc)


def test_snapshot_static_different_sc_4():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"],
                "inodeLimit": data["r_inodeLimit"], "uid": data["r_uid_number"],
                "gid": data["r_gid_number"]}
    restore_sc = {"volBackendFs": data["remoteFs"],
                "filesetType": "dependent", "clusterId": data["remoteid"]}
    snapshot_object.test_static(value_sc, test_restore=True,restore_sc=restore_sc)


def test_snapshot_static_different_sc_5():
    value_sc = {"volBackendFs": data["remoteFs"],
                "inodeLimit": data["r_inodeLimit"],
                "clusterId": data["remoteid"], "filesetType": "independent"}
    restore_sc = {"volBackendFs": data["remoteFs"], "volDirBasePath": data["r_volDirBasePath"]}
    snapshot_object.test_static(value_sc, test_restore=True,restore_sc=restore_sc)


def test_snapshot_static_different_sc_6():
    value_sc = {"volBackendFs": data["remoteFs"],
                "inodeLimit": data["r_inodeLimit"],
                "clusterId": data["remoteid"], "filesetType": "independent"}
    restore_sc = {"volBackendFs": data["remoteFs"],
                "filesetType": "dependent", "clusterId": data["remoteid"]}
    snapshot_object.test_static(value_sc, test_restore=True,restore_sc=restore_sc)


def test_snapshot_dynamic_nodeclass_1():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    restore_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"], "nodeClass": "GUI_MGMT_SERVERS"}
    snapshot_object.test_dynamic(value_sc, test_restore=True,restore_sc=restore_sc)


def test_snapshot_dynamic_nodeclass_2():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    restore_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"], "nodeClass": "GUI_SERVERS"}
    snapshot_object.test_dynamic(value_sc, test_restore=True,restore_sc=restore_sc)


def test_snapshot_dynamic_nodeclass_3():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    restore_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"], "nodeClass": "randomnodeclassx"}
    restore_pvc = {"access_modes": "ReadWriteMany", "storage": "1Gi", "reason": "NotFound desc = nodeclass"}
    snapshot_object.test_dynamic(value_sc, test_restore=True,restore_sc=restore_sc,restore_pvc=restore_pvc)


def test_snapshot_static_nodeclass_1():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    restore_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"], "nodeClass": "GUI_MGMT_SERVERS"}
    snapshot_object.test_static(value_sc, test_restore=True,restore_sc=restore_sc)


def test_snapshot_static_nodeclass_2():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    restore_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"], "nodeClass": "GUI_SERVERS"}
    snapshot_object.test_static(value_sc, test_restore=True,restore_sc=restore_sc)


def test_snapshot_static_nodeclass_3():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    restore_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"], "nodeClass": "randomnodeclassx"}
    restore_pvc = {"access_modes": "ReadWriteMany", "storage": "1Gi", "reason": "NotFound desc = nodeclass"}
    snapshot_object.test_static(value_sc, test_restore=True,restore_sc=restore_sc,restore_pvc=restore_pvc)

