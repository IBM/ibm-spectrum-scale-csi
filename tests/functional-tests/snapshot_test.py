import logging
import pytest
import scale_operator as scaleop
import utils.fileset_functions as ff
LOGGER = logging.getLogger()


@pytest.fixture(scope='session', autouse=True)
def values(request):
    global data, snapshot_object, kubeconfig_value  # are required in every testcase
    kubeconfig_value, clusterconfig_value, namespace_value, runslow_val = scaleop.get_cmd_values(request)

    data = scaleop.read_driver_data(clusterconfig_value, namespace_value)
    operator_data = scaleop.read_operator_data(clusterconfig_value, namespace_value)
    keep_objects = data["keepobjects"]
    test_namespace = namespace_value
    ff.cred_check(data)
    ff.set_data(data)

    operator = scaleop.Scaleoperator(kubeconfig_value, namespace_value)
    operator_object = scaleop.Scaleoperatorobject(operator_data, kubeconfig_value)
    condition = scaleop.check_ns_exists(kubeconfig_value, namespace_value)
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
    snapshot_object = scaleop.Snapshot(kubeconfig_value, test_namespace, keep_objects, value_pvc, value_vs_class, number_of_snapshots, data["image_name"], data["id"])
    if not(data["volBackendFs"] == ""):
        data["primaryFs"] = data["volBackendFs"]
    ff.create_dir(data["volDirBasePath"])

    yield
    if condition is False and not(keep_objects):
        operator_object.delete()
        operator.delete()
        if(ff.fileset_exists(data)):
            ff.delete_fileset(data)

def test_get_version():
    LOGGER.info("Cluster Details:")
    LOGGER.info("----------------")
    ff.get_scale_version(data)
    scaleop.get_kubernetes_version(kubeconfig_value)
    scaleop.scale_function.get_operator_image()
    scaleop.ob.get_driver_image()


def test_snapshot_static_pass_1():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    snapshot_object.test_static(value_sc, test_restore=True)

def test_snapshot_static_multiple_snapshots():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    snapshot_object.test_static(value_sc, test_restore=True, number_of_snapshots=3)

def test_snapshot_static_pass_3():
    value_sc = {"volBackendFs": data["primaryFs"],
                "clusterId": data["id"], "gid": data["gid_number"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_4():
    value_sc = {"volBackendFs": data["primaryFs"],
                "clusterId": data["id"], "inodeLimit": data["inodeLimit"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_5():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"],
                "inodeLimit": data["inodeLimit"], "uid": data["uid_number"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_6():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_7():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"],
                "inodeLimit": data["inodeLimit"], "gid": data["gid_number"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_8():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"],
                "inodeLimit": data["inodeLimit"], "uid": data["uid_number"],
                "gid": data["gid_number"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_9():
    value_sc = {"volBackendFs": data["primaryFs"],
                "clusterId": data["id"], "uid": data["uid_number"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_10():
    value_sc = {"volBackendFs": data["primaryFs"],
                "inodeLimit": data["inodeLimit"],
                "clusterId": data["id"], "filesetType": "independent"}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_11():
    value_sc = {"volBackendFs": data["primaryFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "clusterId": data["id"],
                "filesetType": "independent", "inodeLimit": data["inodeLimit"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_12():
    value_sc = {"volBackendFs": data["primaryFs"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_13():
    value_sc = {"uid": data["uid_number"], "volBackendFs": data["primaryFs"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_14():
    value_sc = {"gid": data["gid_number"], "volBackendFs": data["primaryFs"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_15():
    value_sc = {"inodeLimit": data["inodeLimit"],
                "volBackendFs": data["primaryFs"]}
    snapshot_object.test_static(value_sc, test_restore=True)

def test_snapshot_static_pass_16():
    value_sc = {"volBackendFs": data["primaryFs"], "uid": data["uid_number"],
                "gid": data["gid_number"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_17():
    value_sc = {"volBackendFs": data["primaryFs"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_18():
    value_sc = {"volBackendFs": data["primaryFs"], "gid": data["gid_number"],
                "inodeLimit": data["inodeLimit"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_19():
    value_sc = {"clusterId": data["id"], "gid": data["gid_number"],
                "uid": data["uid_number"], "volBackendFs": data["primaryFs"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_20():
    value_sc = {"clusterId": data["id"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"],
                "volBackendFs": data["primaryFs"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_21():
    value_sc = {"clusterId": data["id"], "gid": data["gid_number"],
                "inodeLimit": data["inodeLimit"],
                "volBackendFs": data["primaryFs"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_22():
    value_sc = {"gid": data["gid_number"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"],
                "volBackendFs": data["primaryFs"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_23():
    value_sc = {"clusterId": data["id"], "volBackendFs": data["primaryFs"],
                "gid": data["gid_number"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"]}
    snapshot_object.test_static(value_sc, test_restore=True)


def test_snapshot_static_pass_24():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    snapshot_object.test_static(value_sc, test_restore=False)


def test_snapshot_static_pass_25():
    value_sc = {"volBackendFs": data["primaryFs"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"]}
    snapshot_object.test_static(value_sc, test_restore=False)


def test_snapshot_static_pass_26():
    value_sc = {"gid": data["gid_number"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"],
                "volBackendFs": data["primaryFs"]}
    snapshot_object.test_static(value_sc, test_restore=False)


def test_snapshot_dynamic_pass_1():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_2():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, value_vs_class={"deletionPolicy": "Retain"})


def test_snapshot_dynamic_expected_fail_1():
    value_sc = {"volBackendFs": data["primaryFs"],
                "filesetType": "dependent", "clusterId": data["id"]}
    snapshot_object.test_dynamic(value_sc, test_restore=False, reason="Volume snapshot can only be created when source volume is independent fileset")


def test_snapshot_dynamic_expected_fail_2():
    value_sc = {"volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc, test_restore=False, reason="Volume snapshot can only be created when source volume is independent fileset")


def test_snapshot_dynamic_multiple_snapshots():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, number_of_snapshots=3)


@pytest.mark.slow
def test_snapshot_dynamic_multiple_snapshots_256():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, number_of_snapshots=256)


@pytest.mark.slow
def test_snapshot_dynamic_multiple_snapshots_257():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True, number_of_snapshots=257)


def test_snapshot_dynamic_pass_3():
    value_sc = {"volBackendFs": data["primaryFs"],
                "clusterId": data["id"], "gid": data["gid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_4():
    value_sc = {"volBackendFs": data["primaryFs"],
                "clusterId": data["id"], "inodeLimit": data["inodeLimit"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_5():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"],
                "inodeLimit": data["inodeLimit"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_6():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_7():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"],
                "inodeLimit": data["inodeLimit"], "gid": data["gid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_8():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"],
                "inodeLimit": data["inodeLimit"], "uid": data["uid_number"],
                "gid": data["gid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_9():
    value_sc = {"volBackendFs": data["primaryFs"],
                "clusterId": data["id"], "uid": data["uid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_10():
    value_sc = {"volBackendFs": data["primaryFs"],
                "inodeLimit": data["inodeLimit"],
                "clusterId": data["id"], "filesetType": "independent"}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_11():
    value_sc = {"volBackendFs": data["primaryFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "clusterId": data["id"],
                "filesetType": "independent", "inodeLimit": data["inodeLimit"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_12():
    value_sc = {"volBackendFs": data["primaryFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_13():
    value_sc = {"uid": data["uid_number"], "volBackendFs": data["primaryFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_14():
    value_sc = {"gid": data["gid_number"], "volBackendFs": data["primaryFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_15():
    value_sc = {"inodeLimit": data["inodeLimit"],
                "volBackendFs": data["primaryFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_16():
    value_sc = {"volBackendFs": data["primaryFs"], "uid": data["uid_number"],
                "gid": data["gid_number"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_17():
    value_sc = {"volBackendFs": data["primaryFs"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_18():
    value_sc = {"volBackendFs": data["primaryFs"], "gid": data["gid_number"],
                "inodeLimit": data["inodeLimit"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_19():
    value_sc = {"clusterId": data["id"], "gid": data["gid_number"],
                "uid": data["uid_number"], "volBackendFs": data["primaryFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_20():
    value_sc = {"clusterId": data["id"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"],
                "volBackendFs": data["primaryFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_21():
    value_sc = {"clusterId": data["id"], "gid": data["gid_number"],
                "inodeLimit": data["inodeLimit"],
                "volBackendFs": data["primaryFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_22():
    value_sc = {"gid": data["gid_number"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"],
                "volBackendFs": data["primaryFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_23():
    value_sc = {"clusterId": data["id"], "volBackendFs": data["primaryFs"],
                "gid": data["gid_number"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True)


def test_snapshot_dynamic_pass_24():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    snapshot_object.test_dynamic(value_sc, test_restore=False)


def test_snapshot_dynamic_pass_25():
    value_sc = {"volBackendFs": data["primaryFs"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"]}
    snapshot_object.test_dynamic(value_sc, test_restore=False)


def test_snapshot_dynamic_pass_26():
    value_sc = {"gid": data["gid_number"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"],
                "volBackendFs": data["primaryFs"]}
    snapshot_object.test_dynamic(value_sc, test_restore=False)


def test_snapshot_dynamic_different_sc_1():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    restore_sc = {"volBackendFs": data["primaryFs"], "volDirBasePath": data["volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True,restore_sc=restore_sc)


def test_snapshot_dynamic_different_sc_2():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    restore_sc = {"volBackendFs": data["primaryFs"],
                "filesetType": "dependent", "clusterId": data["id"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True,restore_sc=restore_sc)


def test_snapshot_dynamic_different_sc_3():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"],
                "inodeLimit": data["inodeLimit"], "uid": data["uid_number"],
                "gid": data["gid_number"]}
    restore_sc = {"volBackendFs": data["primaryFs"], "volDirBasePath": data["volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True,restore_sc=restore_sc)


def test_snapshot_dynamic_different_sc_4():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"],
                "inodeLimit": data["inodeLimit"], "uid": data["uid_number"],
                "gid": data["gid_number"]}
    restore_sc = {"volBackendFs": data["primaryFs"],
                "filesetType": "dependent", "clusterId": data["id"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True,restore_sc=restore_sc)


def test_snapshot_dynamic_different_sc_5():
    value_sc = {"volBackendFs": data["primaryFs"],
                "inodeLimit": data["inodeLimit"],
                "clusterId": data["id"], "filesetType": "independent"}
    restore_sc = {"volBackendFs": data["primaryFs"], "volDirBasePath": data["volDirBasePath"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True,restore_sc=restore_sc)


def test_snapshot_dynamic_different_sc_6():
    value_sc = {"volBackendFs": data["primaryFs"],
                "inodeLimit": data["inodeLimit"],
                "clusterId": data["id"], "filesetType": "independent"}
    restore_sc = {"volBackendFs": data["primaryFs"],
                "filesetType": "dependent", "clusterId": data["id"]}
    snapshot_object.test_dynamic(value_sc, test_restore=True,restore_sc=restore_sc)


def test_snapshot_static_different_sc_1():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    restore_sc = {"volBackendFs": data["primaryFs"], "volDirBasePath": data["volDirBasePath"]}
    snapshot_object.test_static(value_sc, test_restore=True,restore_sc=restore_sc)


def test_snapshot_static_different_sc_2():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    restore_sc = {"volBackendFs": data["primaryFs"],
                "filesetType": "dependent", "clusterId": data["id"]}
    snapshot_object.test_static(value_sc, test_restore=True,restore_sc=restore_sc)


def test_snapshot_static_different_sc_3():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"],
                "inodeLimit": data["inodeLimit"], "uid": data["uid_number"],
                "gid": data["gid_number"]}
    restore_sc = {"volBackendFs": data["primaryFs"], "volDirBasePath": data["volDirBasePath"]}
    snapshot_object.test_static(value_sc, test_restore=True,restore_sc=restore_sc)


def test_snapshot_static_different_sc_4():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"],
                "inodeLimit": data["inodeLimit"], "uid": data["uid_number"],
                "gid": data["gid_number"]}
    restore_sc = {"volBackendFs": data["primaryFs"],
                "filesetType": "dependent", "clusterId": data["id"]}
    snapshot_object.test_static(value_sc, test_restore=True,restore_sc=restore_sc)


def test_snapshot_static_different_sc_5():
    value_sc = {"volBackendFs": data["primaryFs"],
                "inodeLimit": data["inodeLimit"],
                "clusterId": data["id"], "filesetType": "independent"}
    restore_sc = {"volBackendFs": data["primaryFs"], "volDirBasePath": data["volDirBasePath"]}
    snapshot_object.test_static(value_sc, test_restore=True,restore_sc=restore_sc)


def test_snapshot_static_different_sc_6():
    value_sc = {"volBackendFs": data["primaryFs"],
                "inodeLimit": data["inodeLimit"],
                "clusterId": data["id"], "filesetType": "independent"}
    restore_sc = {"volBackendFs": data["primaryFs"],
                "filesetType": "dependent", "clusterId": data["id"]}
    snapshot_object.test_static(value_sc, test_restore=True,restore_sc=restore_sc)


def test_snapshot_dynamic_nodeclass_1():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    restore_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"], "nodeClass": "GUI_MGMT_SERVERS"}
    snapshot_object.test_dynamic(value_sc, test_restore=True,restore_sc=restore_sc)


def test_snapshot_dynamic_nodeclass_2():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    restore_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"], "nodeClass": "GUI_SERVERS"}
    snapshot_object.test_dynamic(value_sc, test_restore=True,restore_sc=restore_sc)


def test_snapshot_dynamic_nodeclass_3():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    restore_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"], "nodeClass": "randomnodeclassx"}
    restore_pvc = {"access_modes": "ReadWriteMany", "storage": "1Gi","reason":"NotFound desc = nodeclass"}
    snapshot_object.test_dynamic(value_sc, test_restore=True,restore_sc=restore_sc,restore_pvc=restore_pvc)


def test_snapshot_static_nodeclass_1():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    restore_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"], "nodeClass": "GUI_MGMT_SERVERS"}
    snapshot_object.test_static(value_sc, test_restore=True,restore_sc=restore_sc)


def test_snapshot_static_nodeclass_2():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    restore_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"], "nodeClass": "GUI_SERVERS"}
    snapshot_object.test_static(value_sc, test_restore=True,restore_sc=restore_sc)


def test_snapshot_static_nodeclass_3():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    restore_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"], "nodeClass": "randomnodeclassx"}
    restore_pvc = {"access_modes": "ReadWriteMany", "storage": "1Gi","reason":"NotFound desc = nodeclass"}
    snapshot_object.test_static(value_sc, test_restore=True,restore_sc=restore_sc,restore_pvc=restore_pvc)
