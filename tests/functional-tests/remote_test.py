import logging
import copy
import pytest
from scale_operator import read_scale_config_file, Scaleoperator, check_ns_exists,\
    check_ds_exists, check_nodes_available, Scaleoperatorobject, Driver, check_key
import utils.fileset_functions as ff
LOGGER = logging.getLogger()


@pytest.fixture(scope='session', autouse=True)
def values(request):
    global data, driver_object  # are required in every testcase
    kubeconfig_value = request.config.option.kubeconfig
    if kubeconfig_value is None:
        kubeconfig_value = "~/.kube/config"
    clusterconfig_value = request.config.option.clusterconfig
    if clusterconfig_value is None:
        clusterconfig_value = "../../operator/deploy/crds/csiscaleoperators.csi.ibm.com_cr.yaml"
    namespace_value = request.config.option.namespace
    if namespace_value is None:
        namespace_value = "ibm-spectrum-scale-csi-driver"
    data = read_scale_config_file(clusterconfig_value, namespace_value)
    if not(check_key(data,"remote")):
        LOGGER.error("remote data is not provided in cr file")
        assert False
    test_namespace = namespace_value

    ff.fileset_exists(data)
    operator = Scaleoperator(kubeconfig_value)
    condition = check_ns_exists(kubeconfig_value, namespace_value)
    if condition is True:
        check_ds_exists(kubeconfig_value, namespace_value)
    else:
        operator.create(namespace_value, data)
        operator.check()
        check_nodes_available(data["pluginNodeSelector"], "pluginNodeSelector")
        check_nodes_available(
            data["provisionerNodeSelector"], "provisionerNodeSelector")
        check_nodes_available(
            data["attacherNodeSelector"], "attacherNodeSelector")
        operator_object = Scaleoperatorobject(data)
        operator_object.create(kubeconfig_value)
        val = operator_object.check(kubeconfig_value)
        if val is True:
            LOGGER.info("Operator custom object is deployed succesfully")
        else:
            LOGGER.error("Operator custom object is not deployed succesfully")
            assert False
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"},
                 {"access_modes": "ReadWriteOnce", "storage": "1Gi"},
                 {"access_modes": "ReadOnlyMany", "storage": "1Gi",
                     "reason": "ReadOnlyMany is not supported"}
                 ]
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False"},
                 {"mount_path": "/usr/share/nginx/html/scale",
                     "read_only": "True", "reason": "Read-only file system"}
                 ]
    data["remoteFs"] = ff.get_remoteFs(data)
    if data["remoteFs"] is None:
        LOGGER.error(" No remote filesystem is mounted")
        assert False
    remote_data = get_remote_data(data)
 
    driver_object = Driver(value_pvc, value_pod, remote_data, test_namespace)
    ff.create_dir(remote_data, remote_data["volDirBasePath"])
    # driver_object.create_test_ns(kubeconfig_value)
    yield
    # driver_object.delete_test_ns(kubeconfig_value)
    #ff.delete_dir(remote_data, remote_data["volDirBasePath"])
    if condition is False:
        operator_object.delete(kubeconfig_value)
        operator.delete()
        if(ff.fileset_exists(data)):
            ff.delete_fileset(data)



def get_remote_data(data_passed):
    remote_data = copy.deepcopy(data_passed)
    remote_data["username"] = remote_data["remote-username"]
    remote_data["password"] = remote_data["remote-password"]
    remote_data["port"] = remote_data["remote-port"]
    remote_data["guiHost"] = remote_data["remoteguiHost"]
    remote_data["primaryFs"] = remote_data["remoteFs"]    
    remote_data["id"] = remote_data["remoteid"]
    remote_data["volDirBasePath"] = remote_data["r-volDirBasePath"]
    remote_data["parentFileset"] = remote_data["r-parentFileset"]
    remote_data["gid_name"] = remote_data["r-gid_name"]
    remote_data["uid_name"] = remote_data["r-uid_name"]
    remote_data["gid_number"] = remote_data["r-gid_number"]
    remote_data["uid_number"] = remote_data["r-uid_number"]
    remote_data["inodeLimit"] = remote_data["r-inodeLimit"]
    return remote_data

#: Testcase that are expected to pass:

def test_driver_static_1():
    value_pvc_custom = [{"access_modes": "ReadWriteMany", "storage": "1Gi"},
                        {"access_modes": "ReadWriteOnce", "storage": "1Gi",
                            "reason": "incompatible accessMode"},
                        {"access_modes": "ReadOnlyMany", "storage": "1Gi",
                            "reason": "incompatible accessMode"}
                        ]
    value_pv = {"access_modes": "ReadWriteMany",
                "storage": "1Gi", "reclaim_policy": "Retain"}
    driver_object.test_static(value_pv, value_pvc_custom)


def test_driver_static_2():
    value_pvc_custom = [{"access_modes": "ReadWriteMany", "storage": "1Gi"},
                        {"access_modes": "ReadWriteOnce", "storage": "1Gi",
                            "reason": "incompatible accessMode"},
                        {"access_modes": "ReadOnlyMany", "storage": "1Gi",
                            "reason": "incompatible accessMode"}
                        ]
    value_pv = {"access_modes": "ReadWriteMany",
                "storage": "1Gi", "reclaim_policy": "Delete"}
    driver_object.test_static(value_pv, value_pvc_custom)


def test_driver_static_3():
    value_pvc_custom = [{"access_modes": "ReadWriteMany", "storage": "1Gi"},
                        {"access_modes": "ReadWriteOnce", "storage": "1Gi",
                            "reason": "incompatible accessMode"},
                        {"access_modes": "ReadOnlyMany", "storage": "1Gi",
                            "reason": "incompatible accessMode"}
                        ]
    value_pv = {"access_modes": "ReadWriteMany",
                "storage": "1Gi", "reclaim_policy": "Recycle"}
    driver_object.test_static(value_pv, value_pvc_custom)


def test_driver_static_4():
    value_pvc_custom = [{"access_modes": "ReadWriteMany", "storage": "1Gi",
                         "reason": "incompatible accessMode"},
                        {"access_modes": "ReadWriteOnce", "storage": "1Gi"},
                        {"access_modes": "ReadOnlyMany", "storage": "1Gi",
                            "reason": "incompatible accessMode"}
                        ]
    value_pv = {"access_modes": "ReadWriteOnce",
                "storage": "1Gi", "reclaim_policy": "Retain"}
    driver_object.test_static(value_pv, value_pvc_custom)


def test_driver_static_5():
    value_pvc_custom = [{"access_modes": "ReadWriteMany", "storage": "1Gi",
                         "reason": "incompatible accessMode"},
                        {"access_modes": "ReadWriteOnce", "storage": "1Gi"},
                        {"access_modes": "ReadOnlyMany", "storage": "1Gi",
                            "reason": "incompatible accessMode"}
                        ]
    value_pv = {"access_modes": "ReadWriteOnce",
                "storage": "1Gi", "reclaim_policy": "Delete"}
    driver_object.test_static(value_pv, value_pvc_custom)


def test_driver_static_6():
    value_pvc_custom = [{"access_modes": "ReadWriteMany", "storage": "1Gi",
                         "reason": "incompatible accessMode"},
                        {"access_modes": "ReadWriteOnce", "storage": "1Gi"},
                        {"access_modes": "ReadOnlyMany", "storage": "1Gi",
                            "reason": "incompatible accessMode"}
                        ]
    value_pv = {"access_modes": "ReadWriteOnce",
                "storage": "1Gi", "reclaim_policy": "Recycle"}
    driver_object.test_static(value_pv, value_pvc_custom)


def test_driver_static_7():
    value_pvc_custom = [{"access_modes": "ReadWriteMany", "storage": "1Gi",
                         "reason": "incompatible accessMode"},
                        {"access_modes": "ReadWriteOnce", "storage": "1Gi",
                            "reason": "incompatible accessMode"},
                        {"access_modes": "ReadOnlyMany", "storage": "1Gi"}
                        ]
    value_pv = {"access_modes": "ReadOnlyMany",
                "storage": "1Gi", "reclaim_policy": "Retain"}
    driver_object.test_static(value_pv, value_pvc_custom)


def test_driver_static_8():
    value_pvc_custom = [{"access_modes": "ReadWriteMany", "storage": "1Gi",
                         "reason": "incompatible accessMode"},
                        {"access_modes": "ReadWriteOnce", "storage": "1Gi",
                            "reason": "incompatible accessMode"},
                        {"access_modes": "ReadOnlyMany", "storage": "1Gi"}
                        ]
    value_pv = {"access_modes": "ReadOnlyMany",
                "storage": "1Gi", "reclaim_policy": "Delete"}
    driver_object.test_static(value_pv, value_pvc_custom)


def test_driver_static_9():
    value_pvc_custom = [{"access_modes": "ReadWriteMany", "storage": "1Gi",
                         "reason": "incompatible accessMode"},
                        {"access_modes": "ReadWriteOnce", "storage": "1Gi",
                            "reason": "incompatible accessMode"},
                        {"access_modes": "ReadOnlyMany", "storage": "1Gi"}
                        ]
    value_pv = {"access_modes": "ReadOnlyMany",
                "storage": "1Gi", "reclaim_policy": "Recycle"}
    driver_object.test_static(value_pv, value_pvc_custom)


def test_driver_static_10():
    value_pvc_custom = [{"access_modes": "ReadWriteMany", "storage": "1Gi"},
                        {"access_modes": "ReadWriteOnce", "storage": "1Gi",
                            "reason": "incompatible accessMode"},
                        {"access_modes": "ReadOnlyMany", "storage": "1Gi",
                            "reason": "incompatible accessMode"}
                        ]
    value_pv = {"access_modes": "ReadWriteMany",
                "storage": "1Gi", "reclaim_policy": "Default"}
    driver_object.test_static(value_pv, value_pvc_custom)


def test_driver_static_11():
    value_pvc_custom = [{"access_modes": "ReadWriteMany", "storage": "1Gi",
                         "reason": "incompatible accessMode"},
                        {"access_modes": "ReadWriteOnce", "storage": "1Gi"},
                        {"access_modes": "ReadOnlyMany", "storage": "1Gi",
                            "reason": "incompatible accessMode"}
                        ]
    value_pv = {"access_modes": "ReadWriteOnce",
                "storage": "1Gi", "reclaim_policy": "Default"}
    driver_object.test_static(value_pv, value_pvc_custom)


def test_driver_static_12():
    value_pvc_custom = [{"access_modes": "ReadWriteMany", "storage": "1Gi",
                         "reason": "incompatible accessMode"},
                        {"access_modes": "ReadWriteOnce", "storage": "1Gi",
                            "reason": "incompatible accessMode"},
                        {"access_modes": "ReadOnlyMany", "storage": "1Gi"}
                        ]
    value_pv = {"access_modes": "ReadOnlyMany",
                "storage": "1Gi", "reclaim_policy": "Default"}
    driver_object.test_static(value_pv, value_pvc_custom)


def test_driver_static_sc_13():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    value_pv = {"access_modes": "ReadWriteMany", "storage": "1Gi",
                "reclaim_policy": "Retain"}
    value_pvc_custom = [{"access_modes": "ReadWriteMany", "storage": "1Gi"},
                        {"access_modes": "ReadWriteOnce", "storage": "1Gi",
                            "reason": "incompatible accessMode"},
                        {"access_modes": "ReadOnlyMany", "storage": "1Gi",
                            "reason": "incompatible accessMode"}
                        ]
    driver_object.test_static(value_pv, value_pvc_custom, value_sc)


def test_driver_static_sc_14():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    value_pv = {"access_modes":"ReadWriteMany", "storage":"1Gi",
                "reclaim_policy":"Delete"}
    value_pvc_custom = [{"access_modes": "ReadWriteMany", "storage": "1Gi"},
                        {"access_modes": "ReadWriteOnce", "storage": "1Gi",
                            "reason": "incompatible accessMode"},
                        {"access_modes": "ReadOnlyMany", "storage": "1Gi",
                            "reason": "incompatible accessMode"}
                        ]
    driver_object.test_static(value_pv, value_pvc_custom, value_sc)


def test_driver_static_sc_15():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    value_pv = {"access_modes": "ReadWriteMany", "storage": "1Gi",
                "reclaim_policy": "Recycle"}
    value_pvc_custom = [{"access_modes": "ReadWriteMany", "storage": "1Gi"},
                        {"access_modes": "ReadWriteOnce", "storage": "1Gi",
                            "reason": "incompatible accessMode"},
                        {"access_modes": "ReadOnlyMany", "storage": "1Gi",
                            "reason": "incompatible accessMode"}
                        ]
    driver_object.test_static(value_pv, value_pvc_custom, value_sc)


def test_driver_static_sc_16():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    value_pv = {"access_modes": "ReadWriteOnce", "storage": "1Gi",
                "reclaim_policy": "Retain"}
    value_pvc_custom = [{"access_modes": "ReadWriteMany", "storage": "1Gi",
                         "reason": "incompatible accessMode"},
                        {"access_modes": "ReadWriteOnce", "storage": "1Gi"},
                        {"access_modes": "ReadOnlyMany", "storage": "1Gi",
                            "reason": "incompatible accessMode"}
                        ]
    driver_object.test_static(value_pv, value_pvc_custom, value_sc)


def test_driver_static_sc_17():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    value_pv = {"access_modes": "ReadWriteOnce", "storage": "1Gi",
                "reclaim_policy": "Delete"}
    value_pvc_custom = [{"access_modes": "ReadWriteMany", "storage": "1Gi",
                         "reason": "incompatible accessMode"},
                        {"access_modes": "ReadWriteOnce", "storage": "1Gi"},
                        {"access_modes": "ReadOnlyMany", "storage": "1Gi",
                            "reason": "incompatible accessMode"}
                        ]
    driver_object.test_static(value_pv, value_pvc_custom, value_sc)

def test_driver_static_sc_18():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    value_pv = {"access_modes": "ReadWriteOnce", "storage": "1Gi",
                "reclaim_policy": "Recycle"}
    value_pvc_custom = [{"access_modes": "ReadWriteMany", "storage": "1Gi",
                         "reason": "incompatible accessMode"},
                        {"access_modes": "ReadWriteOnce", "storage": "1Gi"},
                        {"access_modes": "ReadOnlyMany", "storage": "1Gi",
                            "reason": "incompatible accessMode"}
                        ]
    driver_object.test_static(value_pv, value_pvc_custom, value_sc)


def test_driver_static_sc_19():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    value_pv = {"access_modes": "ReadOnlyMany", "storage": "1Gi",
                "reclaim_policy": "Retain"}
    value_pvc_custom = [{"access_modes": "ReadWriteMany", "storage": "1Gi",
                         "reason": "incompatible accessMode"},
                        {"access_modes": "ReadWriteOnce", "storage": "1Gi",
                            "reason": "incompatible accessMode"},
                        {"access_modes": "ReadOnlyMany", "storage": "1Gi"}
                        ]
    driver_object.test_static(value_pv,value_pvc_custom,value_sc)


def test_driver_static_sc_20():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    value_pv = {"access_modes": "ReadOnlyMany", "storage": "1Gi",
                "reclaim_policy":"Delete"}
    value_pvc_custom = [{"access_modes": "ReadWriteMany", "storage": "1Gi",
                         "reason": "incompatible accessMode"},
                        {"access_modes": "ReadWriteOnce", "storage": "1Gi",
                            "reason": "incompatible accessMode"},
                        {"access_modes": "ReadOnlyMany", "storage": "1Gi"}
                        ]
    driver_object.test_static(value_pv, value_pvc_custom, value_sc)


def test_driver_static_sc_21():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    value_pv = {"access_modes": "ReadOnlyMany", "storage": "1Gi",
                "reclaim_policy": "Recycle"}
    value_pvc_custom = [{"access_modes": "ReadWriteMany", "storage": "1Gi",
                         "reason": "incompatible accessMode"},
                        {"access_modes": "ReadWriteOnce", "storage": "1Gi",
                            "reason": "incompatible accessMode"},
                        {"access_modes": "ReadOnlyMany", "storage": "1Gi"}
                        ]
    driver_object.test_static(value_pv, value_pvc_custom, value_sc)


def test_driver_static_sc_22():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    value_pv = {"access_modes": "ReadWriteMany", "storage": "1Gi",
                "reclaim_policy": "Default"}
    value_pvc_custom = [{"access_modes": "ReadWriteMany", "storage": "1Gi"},
                        {"access_modes": "ReadWriteOnce", "storage": "1Gi",
                            "reason": "incompatible accessMode"},
                        {"access_modes": "ReadOnlyMany", "storage": "1Gi",
                            "reason": "incompatible accessMode"}
                        ]
    driver_object.test_static(value_pv, value_pvc_custom, value_sc)


def test_driver_static_sc_23():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    value_pv = {"access_modes":"ReadWriteOnce", "storage":"1Gi",
                "reclaim_policy":"Default"}
    value_pvc_custom = [{"access_modes": "ReadWriteMany", "storage": "1Gi",
                            "reason": "incompatible accessMode"},
                        {"access_modes": "ReadWriteOnce", "storage": "1Gi"},
                        {"access_modes": "ReadOnlyMany", "storage": "1Gi",
                            "reason": "incompatible accessMode"}
                        ]
    driver_object.test_static(value_pv, value_pvc_custom, value_sc)


def test_driver_static_sc_24():
    value_sc = {"volBackendFs":data["remoteFs"],"clusterId" : data["remoteid"]}
    value_pv = {"access_modes":"ReadOnlyMany","storage":"1Gi","reclaim_policy":"Default"}
    value_pvc_custom = [{"access_modes": "ReadWriteMany", "storage": "1Gi",
                         "reason": "incompatible accessMode"},
                        {"access_modes": "ReadWriteOnce", "storage": "1Gi",
                            "reason": "incompatible accessMode"},
                        {"access_modes": "ReadOnlyMany", "storage": "1Gi"}
                        ]
    driver_object.test_static(value_pv, value_pvc_custom, value_sc)


"""
def test_driver_static_26():
    LOGGER.info("wrong VolumeHandel -> FSUID")
    LOGGER.info("EXPECTED TO FAIL")
    value_pvc_custom = [{"access_modes":"ReadWriteMany","storage":"1Gi"}]
    value_pv = {"access_modes":"ReadWriteMany","storage":"1Gi","reclaim_policy":"Default"}
    wrong={"id_wrong":False,"FSUID_wrong":True}
    driver_object.test_static(value_pv,value_pvc_custom,wrong=wrong)
"""

def test_driver_static_25():
    LOGGER.info("wrong VolumeHandel -> cluster id")
    value_pvc_custom = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_pv = {"access_modes": "ReadWriteMany",
                "storage": "1Gi", "reclaim_policy": "Default"}
    wrong = {"id_wrong": True, "FSUID_wrong": False}
    driver_object.test_static(value_pv, value_pvc_custom, wrong=wrong)


def test_driver_static_27():
    LOGGER.info("PV creation with fileset root path as lightweight volume")
    value_pvc_custom = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_pv = {"access_modes": "ReadWriteMany",
                "storage": "1Gi", "reclaim_policy": "Default"}
    driver_object.test_static(value_pv, value_pvc_custom, root_volume=True)


def test_driver_pass_1():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"]}
    driver_object.test_dynamic(value_sc)


def test_driver_pass_2():
    value_sc = {"volBackendFs": data["remoteFs"],
                "clusterId": data["remoteid"], "uid": data["r-uid_number"]}
    driver_object.test_dynamic(value_sc)


def test_driver_pass_3():
    value_sc = {"volBackendFs": data["remoteFs"],
                "clusterId": data["remoteid"], "gid": data["r-gid_number"]}
    driver_object.test_dynamic(value_sc)


def test_driver_pass_4():
    value_sc = {"volBackendFs": data["remoteFs"],
                "clusterId": data["remoteid"], "inodeLimit": data["r-inodeLimit"]}
    driver_object.test_dynamic(value_sc)


def test_driver_pass_5():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"],
                "inodeLimit": data["r-inodeLimit"], "uid": data["r-uid_number"]}
    driver_object.test_dynamic(value_sc)


def test_driver_pass_6():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"],
                "gid": data["r-gid_number"], "uid": data["r-uid_number"]}
    driver_object.test_dynamic(value_sc)


def test_driver_pass_7():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"],
                "inodeLimit": data["r-inodeLimit"], "gid": data["r-gid_number"]}
    driver_object.test_dynamic(value_sc)


def test_driver_pass_8():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"],
                "inodeLimit": data["r-inodeLimit"], "uid": data["r-uid_number"],
                "gid": data["r-gid_number"]}
    driver_object.test_dynamic(value_sc)


def test_driver_pass_9():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"]}
    driver_object.test_dynamic(value_sc)


def test_driver_pass_10():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "clusterId": data["remoteid"]}
    driver_object.test_dynamic(value_sc)


def test_driver_pass_11():
    value_sc = {"volBackendFs": data["remoteFs"],
                "filesetType": "dependent", "clusterId": data["remoteid"]}
    driver_object.test_dynamic(value_sc)


def test_driver_pass_12():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "uid": data["r-uid_number"]}
    driver_object.test_dynamic(value_sc)


def test_driver_pass_13():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "gid": data["r-gid_number"]}
    driver_object.test_dynamic(value_sc)


def test_driver_pass_14():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "uid": data["r-uid_number"], "clusterId": data["remoteid"]}
    driver_object.test_dynamic(value_sc)


def test_driver_pass_15():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "gid": data["r-gid_number"], "clusterId": data["remoteid"]}
    driver_object.test_dynamic(value_sc)


def test_driver_pass_16():
    value_sc = {"volBackendFs": data["remoteFs"], "uid": data["r-uid_number"],
                "clusterId": data["remoteid"], "filesetType": "dependent"}
    driver_object.test_dynamic(value_sc)


def test_driver_pass_17():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r-gid_number"],
                "clusterId": data["remoteid"], "filesetType": "dependent"}
    driver_object.test_dynamic(value_sc)


def test_driver_pass_18():
    value_sc = {"volBackendFs": data["remoteFs"],
                "parentFileset": data["r-parentFileset"],
                "clusterId": data["remoteid"], "filesetType": "dependent"}
    driver_object.test_dynamic(value_sc)


def test_driver_pass_19():
    value_sc = {"volBackendFs": data["remoteFs"],
                "inodeLimit": data["r-inodeLimit"],
                "clusterId": data["remoteid"], "filesetType": "independent"}
    driver_object.test_dynamic(value_sc)


def test_driver_pass_20():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "gid": data["r-gid_number"], "uid": data["r-uid_number"]}
    driver_object.test_dynamic(value_sc)


def test_driver_pass_21():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "gid": data["r-gid_number"], "uid": data["r-uid_number"],
                "clusterId": data["remoteid"]}
    driver_object.test_dynamic(value_sc)


def test_driver_pass_22():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r-gid_number"],
                "uid": data["r-uid_number"], "clusterId": data["remoteid"],
                "filesetType": "dependent"}
    driver_object.test_dynamic(value_sc)


def test_driver_pass_23():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r-gid_name"],
                "uid": data["r-uid_number"], "clusterId": data["remoteid"],
                "filesetType": "dependent"}
    driver_object.test_dynamic(value_sc)


def test_driver_pass_24():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r-gid_number"],
                "uid": data["r-uid_name"], "clusterId": data["remoteid"],
                "filesetType": "dependent"}
    driver_object.test_dynamic(value_sc)


def test_driver_pass_25():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r-gid_name"],
                "uid": data["r-uid_name"], "clusterId": data["remoteid"],
                "filesetType": "dependent"}
    driver_object.test_dynamic(value_sc)


def test_driver_pass_26():
    value_sc = {"volBackendFs": data["remoteFs"], "uid": data["r-uid_number"],
                "clusterId": data["remoteid"], "filesetType": "dependent",
                "parentFileset": data["r-parentFileset"]}
    driver_object.test_dynamic(value_sc)


def test_driver_pass_27():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r-gid_number"],
                "clusterId": data["remoteid"], "filesetType": "dependent",
                "parentFileset": data["r-parentFileset"]}
    driver_object.test_dynamic(value_sc)


def test_driver_pass_28():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r-gid_number"],
                "uid": data["r-uid_number"], "clusterId": data["remoteid"],
                "filesetType": "dependent",
                "parentFileset": data["r-parentFileset"]}
    driver_object.test_dynamic(value_sc)


def test_driver_pass_29():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r-gid_number"],
                "uid": data["r-uid_number"], "clusterId": data["remoteid"],
                "filesetType": "independent", "inodeLimit": data["r-inodeLimit"]}
    driver_object.test_dynamic(value_sc)

#   Testcases expected to fail with valid values of parameters


def test_driver_30():
    value_sc = {"volBackendFs": data["remoteFs"],
                "reason": "desc = Unable to create fileset"}
    driver_object.test_dynamic(value_sc)


def test_driver_31():
    value_sc = {"clusterId": data["remoteid"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_32():
    value_sc = {"gid": data["r-gid_number"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_33():
    value_sc = {"uid": data["r-uid_number"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_34():
    value_sc = {"filesetType": "dependent", "volBackendFs": data["remoteFs"],
                "reason": "desc = Unable to create fileset"}
    driver_object.test_dynamic(value_sc)


def test_driver_35():
    value_sc = {"inodeLimit": data["r-inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_36():
    value_sc = {"volDirBasePath": data["r-volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_37():
    value_sc = {"parentFileset": data["r-parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_38():
    value_sc = {"uid": data["r-uid_number"], "volBackendFs": data["remoteFs"],
                "reason": "desc = Unable to create fileset"}
    driver_object.test_dynamic(value_sc)


def test_driver_39():
    value_sc = {"gid": data["r-gid_number"], "volBackendFs": data["remoteFs"],
                "reason": "desc = Unable to create fileset"}
    driver_object.test_dynamic(value_sc)


def test_driver_40():
    value_sc = {"filesetType": "dependent", "volBackendFs": data["remoteFs"],
                "reason": "desc = Unable to create fileset"}
    driver_object.test_dynamic(value_sc)


def test_driver_41():
    value_sc = {"parentFileset": data["r-parentFileset"],
                "volBackendFs": data["remoteFs"],
                "reason": "InvalidArgument desc = parentFileset and filesetType=independent"}
    driver_object.test_dynamic(value_sc)


def test_driver_42():
    value_sc = {"inodeLimit": data["r-inodeLimit"],
                "volBackendFs": data["remoteFs"],
                "reason": "desc = Unable to create fileset"}
    driver_object.test_dynamic(value_sc)


def test_driver_43():
    value_sc = {"clusterId": data["remoteid"],
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_44():
    value_sc = {"clusterId": data["remoteid"], "uid": data["r-uid_number"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_45():
    value_sc = {"clusterId": data["remoteid"], "gid": data["r-gid_number"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_46():
    value_sc = {"clusterId": data["remoteid"], "filesetType": "dependent",
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_47():
    value_sc = {"clusterId": data["remoteid"], "parentFileset": data["r-parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_48():
    value_sc = {"clusterId": data["remoteid"], "inodeLimit": data["r-inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_49():
    value_sc = {"volDirBasePath": data["r-volDirBasePath"],
                "uid": data["r-uid_number"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_50():
    value_sc = {"volDirBasePath": data["r-volDirBasePath"],
                "gid": data["r-gid_number"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_51():
    value_sc = {"volDirBasePath": data["r-volDirBasePath"],
                "filesetType": "dependent",
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_52():
    value_sc = {"volDirBasePath": data["r-volDirBasePath"],
                "parentFileset": data["r-parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_53():
    value_sc = {"volDirBasePath": data["r-volDirBasePath"],
                "inodeLimit": data["r-inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_54():
    value_sc = {"uid": data["r-uid_number"], "gid": data["r-gid_number"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_55():
    value_sc = {"uid": data["r-uid_number"], "filesetType": "dependent",
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_56():
    value_sc = {"uid": data["r-uid_number"],
                "parentFileset": data["r-parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_57():
    value_sc = {"uid": data["r-uid_number"], "inodeLimit": data["r-inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_58():
    value_sc = {"gid": data["r-gid_number"], "filesetType": "dependent",
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_59():
    value_sc = {"gid": data["r-gid_number"],
                "parentFileset": data["r-parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_60():
    value_sc = {"gid": data["r-gid_number"], "inodeLimit": data["r-inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_61():
    value_sc = {"filesetType": "dependent",
                "parentFileset": data["r-parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_62():
    value_sc = {"filesetType": "dependent",
                "inodeLimit": data["r-inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_63():
    value_sc = {"inodeLimit": data["r-inodeLimit"],
                "parentFileset": data["r-parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_64():
    value_sc = {"clusterId": data["remoteid"], "volBackendFs": data["remoteFs"],
                "parentFileset": data["r-parentFileset"],
                "reason": "InvalidArgument desc = parentFileset and filesetType=independent"}
    driver_object.test_dynamic(value_sc)


def test_driver_65():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "filesetType": "dependent",
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_66():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "parentFileset": data["r-parentFileset"],
                "reason": "parentFileset and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_67():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "inodeLimit": data["r-inodeLimit"],
                "reason": "inodeLimit and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_68():
    value_sc = {"volBackendFs": data["remoteFs"], "uid": data["r-uid_number"],
                "gid": data["r-gid_number"],
                "reason": "desc = Unable to create fileset"}
    driver_object.test_dynamic(value_sc)


def test_driver_69():
    value_sc = {"volBackendFs": data["remoteFs"], "uid": data["r-uid_number"],
                "filesetType": "dependent",
                "reason": "desc = Unable to create fileset"}
    driver_object.test_dynamic(value_sc)


def test_driver_70():
    value_sc = {"volBackendFs": data["remoteFs"], "uid": data["r-uid_number"],
                "parentFileset": data["r-parentFileset"],
                "reason": "InvalidArgument desc = parentFileset and filesetType=independent"}
    driver_object.test_dynamic(value_sc)


def test_driver_71():
    value_sc = {"volBackendFs": data["remoteFs"], "uid": data["r-uid_number"],
                "inodeLimit": data["r-inodeLimit"],
                "reason": "desc = Unable to create fileset"}
    driver_object.test_dynamic(value_sc)


def test_driver_72():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r-gid_number"],
                "filesetType": "dependent",
                "reason": "desc = Unable to create fileset"}
    driver_object.test_dynamic(value_sc)


def test_driver_73():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r-gid_number"],
                "parentFileset": data["r-parentFileset"],
                "reason": "InvalidArgument desc = parentFileset and filesetType=independent"}
    driver_object.test_dynamic(value_sc)


def test_driver_74():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r-gid_number"],
                "inodeLimit": data["r-inodeLimit"],
                "reason": "desc = Unable to create fileset"}
    driver_object.test_dynamic(value_sc)


def test_driver_75():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "parentFileset": data["r-parentFileset"],
                "reason": "desc = Unable to create fileset"}
    driver_object.test_dynamic(value_sc)


def test_driver_76():
    value_sc = {"volBackendFs": data["remoteFs"], "filesetType": "dependent",
                "inodeLimit": data["r-inodeLimit"],
                "reason": "inodeLimit and filesetType=dependent must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_77():
    value_sc = {"volBackendFs": data["remoteFs"],
                "parentFileset": data["r-parentFileset"],
                "inodeLimit": data["r-inodeLimit"],
                "reason": "InvalidArgument desc = parentFileset and filesetType=independent"}
    driver_object.test_dynamic(value_sc)


def test_driver_78():
    value_sc = {"clusterId": data["remoteid"],
                "volDirBasePath": data["r-volDirBasePath"],
                "uid": data["r-uid_number"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_79():
    value_sc = {"clusterId": data["remoteid"],
                "volDirBasePath": data["r-volDirBasePath"],
                "gid": data["r-gid_number"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_80():
    value_sc = {"clusterId": data["remoteid"],
                "volDirBasePath": data["r-volDirBasePath"],
                "filesetType": "dependent", "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_81():
    value_sc = {"clusterId": data["remoteid"],
                "volDirBasePath": data["r-volDirBasePath"],
                "parentFileset": data["r-parentFileset"], "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_82():
    value_sc = {"clusterId": data["remoteid"],
                "volDirBasePath": data["r-volDirBasePath"],
                "inodeLimit": data["r-inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_83():
    value_sc = {"clusterId": data["remoteid"], "uid": data["r-uid_number"],
                "gid": data["r-gid_number"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_84():
    value_sc = {"clusterId": data["remoteid"], "uid": data["r-uid_number"],
                "filesetType": "dependent",
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_85():
    value_sc = {"clusterId": data["remoteid"], "uid": data["r-uid_number"],
                "parentFileset": data["r-parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_86():
    value_sc = {"clusterId": data["remoteid"], "uid": data["r-uid_number"],
                "inodeLimit": data["r-inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_87():
    value_sc = {"clusterId": data["remoteid"], "gid": data["r-gid_number"],
                "filesetType": "dependent",
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_88():
    value_sc = {"clusterId": data["remoteid"], "gid": data["r-gid_number"],
                "parentFileset": data["r-parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_89():
    value_sc = {"clusterId": data["remoteid"], "gid": data["r-gid_number"],
                "inodeLimit": data["r-inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_90():
    value_sc = {"clusterId": data["remoteid"], "filesetType": "dependent",
                "parentFileset": data["r-parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_91():
    value_sc = {"clusterId": data["remoteid"], "inodeLimit": data["r-inodeLimit"],
                "filesetType": "dependent",
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_92():
    value_sc = {"clusterId": data["remoteid"], "inodeLimit": data["r-inodeLimit"],
                "parentFileset": data["r-parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_93():
    value_sc = {"volDirBasePath": data["r-volDirBasePath"],
                "uid": data["r-uid_number"], "gid": data["r-gid_number"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_94():
    value_sc = {"volDirBasePath": data["r-volDirBasePath"],
                "uid": data["r-uid_number"], "filesetType": "dependent",
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_95():
    value_sc = {"volDirBasePath": data["r-volDirBasePath"],
                "uid": data["r-uid_number"],
                "parentFileset": data["r-parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_96():
    value_sc = {"volDirBasePath": data["r-volDirBasePath"],
                "uid": data["r-uid_number"],
                "inodeLimit": data["r-inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_97():
    value_sc = {"volDirBasePath": data["r-volDirBasePath"],
                "gid": data["r-gid_number"],
                "filesetType": "dependent",
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_98():
    value_sc = {"volDirBasePath": data["r-volDirBasePath"],
                "gid": data["r-gid_number"],
                "parentFileset": data["r-parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_99():
    value_sc = {"volDirBasePath": data["r-volDirBasePath"],
                "gid": data["r-gid_number"],
                "inodeLimit": data["r-inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_100():
    value_sc = {"volDirBasePath": data["r-volDirBasePath"],
                "filesetType": "dependent",
                "parentFileset": data["r-parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_101():
    value_sc = {"volDirBasePath": data["r-volDirBasePath"],
                "filesetType": "dependent",
                "inodeLimit": data["r-inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_102():
    value_sc = {"volDirBasePath": data["r-volDirBasePath"],
                "parentFileset": data["r-parentFileset"],
                "inodeLimit": data["r-inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_103():
    value_sc = {"uid": data["r-uid_number"], "gid": data["r-gid_number"],
                "filesetType": "dependent",
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_104():
    value_sc = {"uid": data["r-uid_number"], "gid": data["r-gid_number"],
                "parentFileset": data["r-parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_105():
    value_sc = {"uid": data["r-uid_number"], "gid": data["r-gid_number"],
                "inodeLimit": data["r-inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_106():
    value_sc = {"uid": data["r-uid_number"], "filesetType": "dependent",
                "parentFileset": data["r-parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_107():
    value_sc = {"uid": data["r-uid_number"], "filesetType": "dependent",
                "inodeLimit": data["r-inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_108():
    value_sc = {"uid": data["r-uid_number"],
                "parentFileset": data["r-parentFileset"],
                "inodeLimit": data["r-inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_109():
    value_sc = {"gid": data["r-gid_number"], "filesetType": "dependent",
                "parentFileset": data["r-parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_110():
    value_sc = {"gid": data["r-gid_number"], "filesetType": "dependent",
                "inodeLimit": data["r-inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_111():
    value_sc = {"gid": data["r-gid_number"],
                "parentFileset": data["r-parentFileset"],
                "inodeLimit": data["r-inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_112():
    value_sc = {"filesetType": "dependent",
                "parentFileset": data["r-parentFileset"],
                "inodeLimit": data["r-inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_113():
    value_sc = {"clusterId": data["remoteid"],  "filesetType":  "dependent",
                "volBackendFs":  data["remoteFs"],
                "volDirBasePath":  data["r-volDirBasePath"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_114():
    value_sc = {"clusterId": data["remoteid"], "volBackendFs": data["remoteFs"],
                "parentFileset": data["r-parentFileset"],
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "parentFileset and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_115():
    value_sc = {"clusterId": data["remoteid"], "inodeLimit": data["r-inodeLimit"],
                "volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "inodeLimit and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_pass_116():
    value_sc = {"clusterId": data["remoteid"], "gid": data["r-gid_number"],
                "uid": data["r-uid_number"], "volBackendFs": data["remoteFs"]}
    driver_object.test_dynamic(value_sc)


def test_driver_117():
    value_sc = {"clusterId": data["remoteid"], "uid": data["r-uid_number"],
                "volBackendFs": data["remoteFs"],
                "parentFileset": data["r-parentFileset"],
                "reason": "InvalidArgument desc = parentFileset and filesetType=independent"}
    driver_object.test_dynamic(value_sc)


def test_driver_pass_118():
    value_sc = {"clusterId": data["remoteid"], "uid": data["r-uid_number"],
                "inodeLimit": data["r-inodeLimit"],
                "volBackendFs": data["remoteFs"]}
    driver_object.test_dynamic(value_sc)


def test_driver_119():
    value_sc = {"clusterId": data["remoteid"], "gid": data["r-gid_number"],
                "volBackendFs": data["remoteFs"],
                "parentFileset": data["r-parentFileset"],
                "reason": "InvalidArgument desc = parentFileset and filesetType=independent"}
    driver_object.test_dynamic(value_sc)


def test_driver_pass_120():
    value_sc = {"clusterId": data["remoteid"], "gid": data["r-gid_number"],
                "inodeLimit": data["r-inodeLimit"],
                "volBackendFs": data["remoteFs"]}
    driver_object.test_dynamic(value_sc)


def test_driver_121():
    value_sc = {"clusterId": data["remoteid"], "filesetType": "independent",
                "volBackendFs": data["remoteFs"],
                "parentFileset": data["r-parentFileset"],
                "reason": "InvalidArgument desc = parentFileset and filesetType=independent"}
    driver_object.test_dynamic(value_sc)


def test_driver_122():
    value_sc = {"clusterId": data["remoteid"], "inodeLimit": data["r-inodeLimit"],
                "filesetType": "dependent", "volBackendFs": data["remoteFs"],
                "reason": "inodeLimit and filesetType=dependent must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_123():
    value_sc = {"clusterId": data["remoteid"], "inodeLimit": data["r-inodeLimit"],
                "volBackendFs": data["remoteFs"],
                "parentFileset": data["r-parentFileset"],
                "reason": "InvalidArgument desc = parentFileset and filesetType=independent"}
    driver_object.test_dynamic(value_sc)


def test_driver_124():
    value_sc = {"uid": data["r-uid_number"], "filesetType": "dependent",
                "volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_125():
    value_sc = {"uid": data["r-uid_number"], "volBackendFs": data["remoteFs"],
                "parentFileset": data["r-parentFileset"],
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "parentFileset and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_126():
    value_sc = {"uid": data["r-uid_number"], "inodeLimit": data["r-inodeLimit"],
                "volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "inodeLimit and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_127():
    value_sc = {"gid": data["r-gid_number"], "filesetType": "dependent",
                "volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_128():
    value_sc = {"gid": data["r-gid_number"], "volBackendFs": data["remoteFs"],
                "parentFileset": data["r-parentFileset"],
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "parentFileset and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_129():
    value_sc = {"gid": data["r-gid_number"], "inodeLimit": data["r-inodeLimit"],
                "volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "inodeLimit and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_130():
    value_sc = {"filesetType": "dependent", "volBackendFs": data["remoteFs"],
                "parentFileset": data["r-parentFileset"],
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_131():
    value_sc = {"inodeLimit": data["r-inodeLimit"], "filesetType": "dependent",
                "volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_132():
    value_sc = {"inodeLimit": data["r-inodeLimit"],
                "volBackendFs": data["remoteFs"],
                "parentFileset": data["r-parentFileset"],
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "parentFileset and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_133():
    value_sc = {"gid": data["r-gid_number"], "uid": data["r-uid_number"],
                "filesetType": "dependent",
                "volBackendFs": data["remoteFs"],
                "reason": "desc = Unable to create fileset"}
    driver_object.test_dynamic(value_sc)


def test_driver_134():
    value_sc = {"gid": data["r-gid_number"], "uid": data["r-uid_number"],
                "volBackendFs": data["remoteFs"],
                "parentFileset": data["r-parentFileset"],
                "reason": "InvalidArgument desc = parentFileset and filesetType=independent"}
    driver_object.test_dynamic(value_sc)


def test_driver_135():
    value_sc = {"gid": data["r-gid_number"], "uid": data["r-uid_number"],
                "inodeLimit": data["r-inodeLimit"],
                "volBackendFs": data["remoteFs"],
                "reason": "desc = Unable to create fileset"}
    driver_object.test_dynamic(value_sc)


def test_driver_136():
    value_sc = {"uid": data["r-uid_number"], "filesetType": "dependent",
                "volBackendFs": data["remoteFs"],
                "parentFileset": data["r-parentFileset"],
                "reason": "desc = Unable to create fileset"}
    driver_object.test_dynamic(value_sc)


def test_driver_137():
    value_sc = {"uid": data["r-uid_number"], "inodeLimit": data["r-inodeLimit"],
                "filesetType": "dependent", "volBackendFs": data["remoteFs"],
                "reason": "inodeLimit and filesetType=dependent must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_138():
    value_sc = {"uid": data["r-uid_number"], "inodeLimit": data["r-inodeLimit"],
                "volBackendFs": data["remoteFs"],
                "parentFileset": data["r-parentFileset"],
                "reason": "InvalidArgument desc = parentFileset and filesetType=independent"}
    driver_object.test_dynamic(value_sc)


def test_driver_139():
    value_sc = {"gid": data["r-gid_number"], "filesetType": "dependent",
                "volBackendFs": data["remoteFs"],
                "parentFileset": data["r-parentFileset"],
                "reason": "desc = Unable to create fileset"}
    driver_object.test_dynamic(value_sc)


def test_driver_140():
    value_sc = {"gid": data["r-gid_number"], "inodeLimit": data["r-inodeLimit"],
                "filesetType": "dependent", "volBackendFs": data["remoteFs"],
                "reason": "inodeLimit and filesetType=dependent must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_141():
    value_sc = {"gid": data["r-gid_number"], "inodeLimit": data["r-inodeLimit"],
                "volBackendFs": data["remoteFs"],
                "parentFileset": data["r-parentFileset"],
                "reason": "InvalidArgument desc = parentFileset and filesetType=independent"}
    driver_object.test_dynamic(value_sc)


def test_driver_142():
    value_sc = {"inodeLimit": data["r-inodeLimit"], "filesetType": "dependent",
                "volBackendFs": data["remoteFs"],
                "parentFileset": data["r-parentFileset"],
                "reason": "inodeLimit and filesetType=dependent must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_143():
    value_sc = {"clusterId": data["remoteid"], "gid": data["r-gid_number"],
                "uid": data["r-uid_number"],
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_144():
    value_sc = {"clusterId": data["remoteid"], "uid": data["r-uid_number"],
                "filesetType": "dependent",
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_145():
    value_sc = {"clusterId": data["remoteid"], "uid": data["r-uid_number"],
                "parentFileset": data["r-parentFileset"],
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_146():
    value_sc = {"clusterId": data["remoteid"], "uid": data["r-uid_number"],
                "volDirBasePath": data["r-volDirBasePath"],
                "inodeLimit": data["r-inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_147():
    value_sc = {"clusterId": data["remoteid"], "gid": data["r-gid_number"],
                "filesetType": "dependent", "volDirBasePath": data["r-volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_148():
    value_sc = {"clusterId": data["remoteid"], "gid": data["r-gid_number"],
                "parentFileset": data["r-parentFileset"],
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_149():
    value_sc = {"clusterId": data["remoteid"], "gid": data["r-gid_number"],
                "inodeLimit": data["r-inodeLimit"],
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_150():
    value_sc = {"clusterId": data["remoteid"], "filesetType": "dependent",
                "parentFileset": data["r-parentFileset"],
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_151():
    value_sc = {"clusterId": data["remoteid"], "inodeLimit": data["r-inodeLimit"],
                "filesetType": "dependent",
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_152():
    value_sc = {"clusterId": data["remoteid"], "inodeLimit": data["r-inodeLimit"],
                "parentFileset": data["r-parentFileset"],
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_153():
    value_sc = {"clusterId": data["remoteid"], "gid": data["r-gid_number"],
                "uid": data["r-uid_number"], "filesetType": "dependent",
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_154():
    value_sc = {"clusterId": data["remoteid"], "gid": data["r-gid_number"],
                "uid": data["r-uid_number"],
                "parentFileset": data["r-parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_155():
    value_sc = {"clusterId": data["remoteid"], "gid": data["r-gid_number"],
                "uid": data["r-uid_number"], "inodeLimit": data["r-inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_156():
    value_sc = {"clusterId": data["remoteid"], "uid": data["r-uid_number"],
                "filesetType": "dependent",
                "parentFileset": data["r-parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_157():
    value_sc = {"clusterId": data["remoteid"], "uid": data["r-uid_number"],
                "inodeLimit": data["r-inodeLimit"],
                "filesetType": "dependent",
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_158():
    value_sc = {"clusterId": data["remoteid"], "uid": data["r-uid_number"],
                "inodeLimit": data["r-inodeLimit"],
                "parentFileset": data["r-parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_159():
    value_sc = {"clusterId": data["remoteid"], "gid": data["r-gid_number"],
                "filesetType": "dependent",
                "parentFileset": data["r-parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_160():
    value_sc = {"clusterId": data["remoteid"], "gid": data["r-gid_number"],
                "inodeLimit": data["r-inodeLimit"],
                "filesetType": "dependent",
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_161():
    value_sc = {"clusterId": data["remoteid"],
                "gid": data["r-gid_number"], "inodeLimit": data["r-inodeLimit"],
                "parentFileset": data["r-parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_162():
    value_sc = {"clusterId": data["remoteid"], "inodeLimit": data["r-inodeLimit"],
                "filesetType": "dependent",
                "parentFileset": data["r-parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_163():
    value_sc = {"gid": data["r-gid_number"], "uid": data["r-uid_number"],
                "filesetType": "dependent",
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_164():
    value_sc = {"gid": data["r-gid_number"], "uid": data["r-uid_number"],
                "parentFileset": data["r-parentFileset"],
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_165():
    value_sc = {"gid": data["r-gid_number"], "uid": data["r-uid_number"],
                "inodeLimit": data["r-inodeLimit"],
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_166():
    value_sc = {"uid": data["r-uid_number"], "filesetType": "dependent",
                "parentFileset": data["r-parentFileset"],
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_167():
    value_sc = {"uid": data["r-uid_number"], "inodeLimit": data["r-inodeLimit"],
                "filesetType": "dependent",
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_168():
    value_sc = {"uid": data["r-uid_number"], "inodeLimit": data["r-inodeLimit"],
                "parentFileset": data["r-parentFileset"],
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_169():
    value_sc = {"gid": data["r-gid_number"], "filesetType": "dependent",
                "parentFileset": data["r-parentFileset"],
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_170():
    value_sc = {"gid": data["r-gid_number"], "inodeLimit": data["r-inodeLimit"],
                "filesetType": "dependent",
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_171():
    value_sc = {"gid": data["r-gid_number"], "inodeLimit": data["r-inodeLimit"],
                "parentFileset": data["r-parentFileset"],
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_172():
    value_sc = {"inodeLimit": data["r-inodeLimit"], "filesetType": "dependent",
                "parentFileset": data["r-parentFileset"],
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_173():
    value_sc = {"gid": data["r-gid_number"], "uid": data["r-uid_number"],
                "filesetType": "dependent",
                "parentFileset": data["r-parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_174():
    value_sc = {"gid": data["r-gid_number"], "uid": data["r-uid_number"],
                "inodeLimit": data["r-inodeLimit"],
                "filesetType": "dependent",
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_175():
    value_sc = {"gid": data["r-gid_number"], "uid": data["r-uid_number"],
                "inodeLimit": data["r-inodeLimit"],
                "parentFileset": data["r-parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_176():
    value_sc = {"uid": data["r-uid_number"], "inodeLimit": data["r-inodeLimit"],
                "filesetType": "dependent",
                "parentFileset": data["r-parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_177():
    value_sc = {"gid": data["r-gid_number"], "inodeLimit": data["r-inodeLimit"],
                "filesetType": "dependent",
                "parentFileset": data["r-parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_178():
    value_sc = {"clusterId": data["remoteid"], "volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "uid": data["r-uid_number"],
                "filesetType": "dependent",
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_179():
    value_sc = {"clusterId": data["remoteid"], "volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "uid": data["r-uid_number"],
                "parentFileset": data["r-parentFileset"],
                "reason": "parentFileset and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_180():
    value_sc = {"clusterId": data["remoteid"], "volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "uid": data["r-uid_number"],
                "inodeLimit": data["r-inodeLimit"],
                "reason": "inodeLimit and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_181():
    value_sc = {"clusterId": data["remoteid"], "volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "gid": data["r-gid_number"],
                "filesetType": "dependent",
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_182():
    value_sc = {"clusterId": data["remoteid"], "volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "gid": data["r-gid_number"],
                "parentFileset": data["r-parentFileset"],
                "reason": "parentFileset and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_183():
    value_sc = {"clusterId": data["remoteid"], "volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "gid": data["r-gid_number"],
                "inodeLimit": data["r-inodeLimit"],
                "reason": "inodeLimit and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_184():
    value_sc = {"clusterId": data["remoteid"], "volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "filesetType": "dependent",
                "parentFileset": data["r-parentFileset"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_185():
    value_sc = {"clusterId": data["remoteid"], "volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "inodeLimit": data["r-inodeLimit"],
                "filesetType": "dependent",
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_186():
    value_sc = {"clusterId": data["remoteid"], "volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "inodeLimit": data["r-inodeLimit"],
                "parentFileset": data["r-parentFileset"],
                "reason": "parentFileset and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_187():
    value_sc = {"clusterId": data["remoteid"], "volBackendFs": data["remoteFs"],
                "gid": data["r-gid_number"], "uid": data["r-uid_number"],
                "parentFileset": data["r-parentFileset"],
                "reason": "InvalidArgument desc = parentFileset and filesetType=independent"}
    driver_object.test_dynamic(value_sc)


def test_driver_pass_188():
    value_sc = {"clusterId": data["remoteid"], "volBackendFs": data["remoteFs"],
                "gid": data["r-gid_number"], "uid": data["r-uid_number"],
                "inodeLimit": data["r-inodeLimit"]}
    driver_object.test_dynamic(value_sc)


def test_driver_189():
    value_sc = {"clusterId": data["remoteid"], "volBackendFs": data["remoteFs"],
                "uid": data["r-uid_number"], "filesetType": "independent",
                "parentFileset": data["r-parentFileset"],
                "reason": "InvalidArgument desc = parentFileset and filesetType=independent"}
    driver_object.test_dynamic(value_sc)


def test_driver_190():
    value_sc = {"clusterId": data["remoteid"], "volBackendFs": data["remoteFs"],
                "uid": data["r-uid_number"],
                "parentFileset": data["r-parentFileset"],
                "inodeLimit": data["r-inodeLimit"],
                "reason": "InvalidArgument desc = parentFileset and filesetType=independent"}
    driver_object.test_dynamic(value_sc)


def test_driver_191():
    value_sc = {"clusterId": data["remoteid"], "volBackendFs": data["remoteFs"],
                "gid": data["r-gid_number"],
                "parentFileset": data["r-parentFileset"],
                "inodeLimit": data["r-inodeLimit"], "reason": "InvalidArgument desc = parentFileset and filesetType=independent"}
    driver_object.test_dynamic(value_sc)


def test_driver_192():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "gid": data["r-gid_number"], "uid": data["r-uid_number"],
                "filesetType": "dependent",
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_193():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "gid": data["r-gid_number"], "uid": data["r-uid_number"],
                "parentFileset": data["r-parentFileset"],
                "reason": "parentFileset and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_194():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "gid": data["r-gid_number"], "uid": data["r-uid_number"],
                "inodeLimit": data["r-inodeLimit"],
                "reason": "inodeLimit and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_195():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "uid": data["r-uid_number"], "filesetType": "dependent",
                "parentFileset": data["r-parentFileset"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_196():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "uid": data["r-uid_number"], "inodeLimit": data["r-inodeLimit"],
                "filesetType": "dependent",
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_197():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "uid": data["r-uid_number"],
                "parentFileset": data["r-parentFileset"],
                "inodeLimit": data["r-inodeLimit"],
                "reason": "parentFileset and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_198():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "gid": data["r-gid_number"], "filesetType": "dependent",
                "parentFileset": data["r-parentFileset"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_199():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "gid": data["r-gid_number"], "inodeLimit": data["r-inodeLimit"],
                "filesetType": "dependent",
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_200():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "gid": data["r-gid_number"],
                "parentFileset": data["r-parentFileset"],
                "inodeLimit": data["r-inodeLimit"],
                "reason": "parentFileset and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_201():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "parentFileset": data["r-parentFileset"],
                "inodeLimit": data["r-inodeLimit"], "filesetType": "dependent",
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_202():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r-gid_number"],
                "uid": data["r-uid_number"], "filesetType": "dependent",
                "parentFileset": data["r-parentFileset"],
                "reason": "desc = Unable to create fileset"} 
    driver_object.test_dynamic(value_sc)


def test_driver_203():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r-gid_number"],
                "uid": data["r-uid_number"], "filesetType": "dependent",
                "inodeLimit": data["r-inodeLimit"],
                "reason": "inodeLimit and filesetType=dependent must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_204():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r-gid_number"],
                "uid": data["r-uid_number"], "parentFileset": data["r-parentFileset"],
                "inodeLimit": data["r-inodeLimit"],
                "reason": "InvalidArgument desc = parentFileset and filesetType=independent"}
    driver_object.test_dynamic(value_sc)


def test_driver_205():
    value_sc = {"volBackendFs": data["remoteFs"], "uid": data["r-uid_number"],
                "filesetType": "dependent",
                "parentFileset": data["r-parentFileset"],
                "inodeLimit": data["r-inodeLimit"],
                "reason": "inodeLimit and filesetType=dependent must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_206():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r-gid_number"],
                "filesetType": "dependent",
                "parentFileset": data["r-parentFileset"],
                "inodeLimit": data["r-inodeLimit"],
                "reason": "inodeLimit and filesetType=dependent must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_207():
    value_sc = {"clusterId": data["remoteid"], "gid": data["r-gid_number"],
                "uid": data["r-uid_number"], "filesetType": "dependent",
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_208():
    value_sc = {"clusterId": data["remoteid"], "gid": data["r-gid_number"],
                "uid": data["r-uid_number"],
                "volDirBasePath": data["r-volDirBasePath"],
                "parentFileset": data["r-parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_209():
    value_sc = {"clusterId": data["remoteid"], "gid": data["r-gid_number"],
                "uid": data["r-uid_number"],
                "volDirBasePath": data["r-volDirBasePath"],
                "inodeLimit": data["r-inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_210():
    value_sc = {"clusterId": data["remoteid"], "uid": data["r-uid_number"],
                "volDirBasePath": data["r-volDirBasePath"],
                "filesetType": "dependent",
                "parentFileset": data["r-parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_211():
    value_sc = {"clusterId": data["remoteid"], "uid": data["r-uid_number"],
                "volDirBasePath": data["r-volDirBasePath"],
                "filesetType": "dependent", "r-inodeLimit": data["r-inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_212():
    value_sc = {"clusterId": data["remoteid"], "uid": data["r-uid_number"],
                "volDirBasePath": data["r-volDirBasePath"],
                "inodeLimit": data["r-inodeLimit"],
                "parentFileset": data["r-parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_213():
    value_sc = {"clusterId": data["remoteid"], "gid": data["r-gid_number"],
                "volDirBasePath": data["r-volDirBasePath"],
                "filesetType": "dependent",
                "parentFileset": data["r-parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_214():
    value_sc = {"clusterId": data["remoteid"], "gid": data["r-gid_number"],
                "volDirBasePath": data["r-volDirBasePath"],
                "filesetType": "dependent", "inodeLimit": data["r-inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_215():
    value_sc = {"clusterId": data["remoteid"], "gid": data["r-gid_number"],
                "volDirBasePath": data["r-volDirBasePath"],
                "inodeLimit": data["r-inodeLimit"],
                "parentFileset": data["r-parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_216():
    value_sc = {"clusterId": data["remoteid"], "filesetType": "dependent",
                "volDirBasePath": data["r-volDirBasePath"],
                "inodeLimit": data["r-inodeLimit"],
                "parentFileset": data["r-parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_217():
    value_sc = {"clusterId": data["remoteid"], "filesetType": "dependent",
                "gid": data["r-gid_number"], "uid": data["r-uid_number"],
                "parentFileset": data["r-parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_218():
    value_sc = {"clusterId": data["remoteid"], "filesetType": "dependent",
                "gid": data["r-gid_number"], "uid": data["r-uid_number"],
                "inodeLimit": data["r-inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_219():
    value_sc = {"clusterId": data["remoteid"], "gid": data["r-gid_number"],
                "uid": data["r-uid_number"], "inodeLimit": data["r-inodeLimit"],
                "parentFileset": data["r-parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_220():
    value_sc = {"clusterId": data["remoteid"], "filesetType": "dependent",
                "uid": data["r-uid_number"],
                "parentFileset": data["r-parentFileset"],
                "inodeLimit": data["r-inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_221():
    value_sc = {"clusterId": data["remoteid"], "gid": data["r-gid_number"],
                "filesetType": "dependent",
                "parentFileset": data["r-parentFileset"],
                "inodeLimit": data["r-inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_222():
    value_sc = {"filesetType": "dependent", "gid": data["r-gid_number"],
                "uid": data["r-uid_number"],
                "volDirBasePath": data["r-volDirBasePath"],
                "parentFileset": data["r-parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_223():
    value_sc = {"filesetType": "dependent", "gid": data["r-gid_number"],
                "uid": data["r-uid_number"], "inodeLimit": data["r-inodeLimit"],
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_224():
    value_sc = {"gid": data["r-gid_number"], "uid": data["r-uid_number"],
                "inodeLimit": data["r-inodeLimit"],
                "volDirBasePath": data["r-volDirBasePath"],
                "parentFileset": data["r-parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_225():
    value_sc = {"filesetType": "dependent", "uid": data["r-uid_number"],
                "inodeLimit": data["r-inodeLimit"],
                "volDirBasePath": data["r-volDirBasePath"],
                "parentFileset": data["r-parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_226():
    value_sc = {"filesetType": "dependent", "gid": data["r-gid_number"],
                "inodeLimit": data["r-inodeLimit"],
                "volDirBasePath": data["r-volDirBasePath"],
                "parentFileset": data["r-parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_227():
    value_sc = {"filesetType": "dependent", "gid": data["r-gid_number"],
                "uid": data["r-uid_number"], "inodeLimit": data["r-inodeLimit"],
                "parentFileset": data["r-parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_228():
    value_sc = {"gid": data["r-gid_number"], "uid": data["r-uid_number"],
                "volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "clusterId":  data["remoteid"], "filesetType": "dependent",
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_229():
    value_sc = {"gid": data["r-gid_number"], "uid": data["r-uid_number"],
                "volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "clusterId": data["remoteid"], "parentFileset": data["r-parentFileset"],
                "reason": "parentFileset and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_230():
    value_sc = {"gid": data["r-gid_number"], "uid": data["r-uid_number"],
                "volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "clusterId": data["remoteid"], "inodeLimit": data["r-inodeLimit"],
                "reason": "inodeLimit and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_231():
    value_sc = {"uid": data["r-uid_number"], "volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "clusterId": data["remoteid"], "filesetType": "dependent",
                "parentFileset": data["r-parentFileset"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_232():
    value_sc = {"uid": data["r-uid_number"], "volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "clusterId": data["remoteid"],
                "filesetType": "dependent", "inodeLimit": data["r-inodeLimit"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_233():
    value_sc = {"uid": data["r-uid_number"], "volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "clusterId": data["remoteid"], "inodeLimit": data["r-inodeLimit"],
                "parentFileset": data["r-parentFileset"],
                "reason": "parentFileset and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_234():
    value_sc = {"gid": data["r-gid_number"], "volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "clusterId": data["remoteid"], "filesetType": "dependent",
                "parentFileset": data["r-parentFileset"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_235():
    value_sc = {"gid": data["r-gid_number"], "volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "clusterId": data["remoteid"],
                "filesetType": "dependent", "inodeLimit": data["r-inodeLimit"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_236():
    value_sc = {"gid": data["r-gid_number"], "volBackendFs": data["remoteFs"],
                "volDirBasePath": data["r-volDirBasePath"],
                "clusterId": data["remoteid"], "parentFileset": data["r-parentFileset"],
                "inodeLimit": data["r-inodeLimit"],
                "reason": "parentFileset and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_237():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId": data["remoteid"],
                "volDirBasePath": data["r-volDirBasePath"],
                "filesetType": "dependent", "inodeLimit": data["r-inodeLimit"],
                "parentFileset": data["r-parentFileset"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_238():
    value_sc = {"clusterId": data["remoteid"], "volBackendFs": data["remoteFs"],
                "uid": data["r-uid_number"], "inodeLimit": data["r-inodeLimit"],
                "parentFileset": data["r-parentFileset"],
                "gid": data["r-gid_number"],
                "reason": "InvalidArgument desc = parentFileset and filesetType=independent"}
    driver_object.test_dynamic(value_sc)


def test_driver_239():
    value_sc = {"clusterId": data["remoteid"], "volBackendFs": data["remoteFs"],
                "uid": data["r-uid_number"], "inodeLimit": data["r-inodeLimit"],
                "parentFileset": data["r-parentFileset"],
                "filesetType": "dependent",
                "reason": "inodeLimit and filesetType=dependent must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_240():
    value_sc = {"clusterId": data["remoteid"], "volBackendFs": data["remoteFs"],
                "uid": data["r-uid_number"], "inodeLimit": data["r-inodeLimit"],
                "parentFileset": data["r-parentFileset"],
                "filesetType": "independent",
                "reason": "InvalidArgument desc = parentFileset and filesetType=independent"}
    driver_object.test_dynamic(value_sc)


def test_driver_241():
    value_sc = {"volBackendFs": data["remoteFs"], "uid": data["r-uid_number"],
                "gid": data["r-gid_number"], "filesetType": "dependent",
                "parentFileset": data["r-parentFileset"],
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_242():
    value_sc = {"volBackendFs": data["remoteFs"], "uid": data["r-uid_number"],
                "gid": data["r-gid_number"], "inodeLimit": data["r-inodeLimit"],
                "filesetType": "dependent",
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_243():
    value_sc = {"volBackendFs": data["remoteFs"], "uid": data["r-uid_number"],
                "gid": data["r-gid_number"], "inodeLimit": data["r-inodeLimit"],
                "parentFileset": data["r-parentFileset"],
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "parentFileset and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_244():
    value_sc = {"volBackendFs": data["remoteFs"], "uid": data["r-uid_number"],
                "inodeLimit": data["r-inodeLimit"],
                "parentFileset": data["r-parentFileset"],
                "filesetType": "dependent",
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_245():
    value_sc = {"volBackendFs": data["remoteFs"], "gid": data["r-gid_number"],
                "inodeLimit": data["r-inodeLimit"],
                "parentFileset": data["r-parentFileset"],
                "filesetType": "dependent",
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_246():
    value_sc = {"volBackendFs": data["remoteFs"], "uid": data["r-uid_number"],
                "gid": data["r-gid_number"], "inodeLimit": data["r-inodeLimit"],
                "parentFileset": data["r-parentFileset"],
                "filesetType": "dependent",
                "reason": "inodeLimit and filesetType=dependent must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_247():
    value_sc = {"clusterId":  data["remoteid"], "uid": data["r-uid_number"],
                "gid": data["r-gid_number"],
                "parentFileset": data["r-parentFileset"],
                "filesetType": "dependent",
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_248():
    value_sc = {"clusterId": data["remoteid"], "uid": data["r-uid_number"],
                "gid": data["r-gid_number"], "inodeLimit": data["r-inodeLimit"],
                "filesetType": "dependent",
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_249():
    value_sc = {"clusterId": data["remoteid"], "uid": data["r-uid_number"],
                "gid": data["r-gid_number"], "inodeLimit": data["r-inodeLimit"],
                "parentFileset": data["r-parentFileset"],
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_250():
    value_sc = {"clusterId": data["remoteid"], "uid": data["r-uid_number"],
                "inodeLimit": data["r-inodeLimit"],
                "parentFileset": data["r-parentFileset"],
                "filesetType": "dependent",
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_251():
    value_sc = {"clusterId": data["remoteid"], "gid": data["r-gid_number"],
                "inodeLimit": data["r-inodeLimit"],
                "parentFileset": data["r-parentFileset"],
                "filesetType": "dependent",
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_252():
    value_sc = {"clusterId": data["remoteid"], "uid": data["r-uid_number"],
                "gid": data["r-gid_number"], "inodeLimit": data["r-inodeLimit"],
                "parentFileset": data["r-parentFileset"],
                "filesetType": "dependent",
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_253():
    value_sc = {"uid": data["r-uid_number"], "gid": data["r-gid_number"],
                "inodeLimit": data["r-inodeLimit"],
                "parentFileset": data["r-parentFileset"],
                "filesetType": "dependent",
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_254():
    value_sc = {"clusterId": data["remoteid"], "volBackendFs": data["remoteFs"],
                "uid": data["r-uid_number"], "gid": data["r-gid_number"],
                "parentFileset": data["r-parentFileset"],
                "filesetType": "dependent",
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_255():
    value_sc = {"clusterId": data["remoteid"], "volBackendFs": data["remoteFs"],
                "uid": data["r-uid_number"], "gid": data["r-gid_number"],
                "inodeLimit": data["r-inodeLimit"],
                "filesetType": "dependent",
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_256():
    value_sc = {"clusterId": data["remoteid"], "volBackendFs": data["remoteFs"],
                "uid": data["r-uid_number"], "gid": data["r-gid_number"],
                "inodeLimit": data["r-inodeLimit"],
                "parentFileset": data["r-parentFileset"],
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "parentFileset and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_257():
    value_sc = {"clusterId": data["remoteid"], "volBackendFs": data["remoteFs"],
                "uid": data["r-uid_number"], "inodeLimit": data["r-inodeLimit"],
                "parentFileset": data["r-parentFileset"],
                "filesetType": "dependent",
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_258():
    value_sc = {"clusterId": data["remoteid"], "volBackendFs": data["remoteFs"],
                "gid": data["r-gid_number"], "inodeLimit": data["r-inodeLimit"],
                "parentFileset": data["r-parentFileset"],
                "filesetType": "dependent",
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_259():
    value_sc = {"volBackendFs": data["remoteFs"], "uid": data["r-uid_number"],
                "gid": data["r-gid_number"], "inodeLimit": data["r-inodeLimit"],
                "parentFileset": data["r-parentFileset"],
                "filesetType": "dependent",
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_260():
    value_sc = {"clusterId": data["remoteid"], "uid": data["r-uid_number"],
                "gid": data["r-gid_number"], "inodeLimit": data["r-inodeLimit"],
                "parentFileset": data["r-parentFileset"],
                "filesetType": "dependent",
                "volDirBasePath": data["r-volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_261():
    value_sc = {"clusterId": data["remoteid"], "volBackendFs": data["remoteFs"],
                "uid": data["r-uid_number"], "gid": data["r-gid_number"],
                "inodeLimit": data["r-inodeLimit"],
                "parentFileset": data["r-parentFileset"],
                "filesetType": "dependent",
                "reason": "inodeLimit and filesetType=dependent must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_262():
    value_sc = {"clusterId": data["remoteid"], "filesetType": "dependent",
                "uid": data["r-uid_number"], "inodeLimit": data["r-inodeLimit"],
                "volBackendFs": data["remoteFs"],
                "reason": "inodeLimit and filesetType=dependent must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_263():
    value_sc = {"clusterId": data["remoteid"], "filesetType": "dependent",
                "gid": data["r-gid_number"], "inodeLimit": data["r-inodeLimit"],
                "volBackendFs": data["remoteFs"],
                "reason": "inodeLimit and filesetType=dependent must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_264():
    value_sc = {"clusterId": data["remoteid"], "filesetType": "independent",
                "inodeLimit": data["r-inodeLimit"],
                "volBackendFs": data["remoteFs"],
                "parentFileset": data["r-parentFileset"],
                "reason": "InvalidArgument desc = parentFileset and filesetType=independent"}
    driver_object.test_dynamic(value_sc)

# invalid input test cases


invaliddata = {
    "r-volDirBasePath": "/invalid",
    "remoteFs": "invalid",
    "r-parentFileset": "invalid",
    "r-inodeLimit": "1023",
    "r-gid_number": "9999",
    "r-uid_number": "9999",
    "r-gid_name": "invalid_name",
    "r-uid_name": "invalid_name"
}


def test_driver_invalid_input_265():
    value_sc = {"volBackendFs": data["remoteFs"],
                "volDirBasePath": invaliddata["r-volDirBasePath"],
                "reason": "Directory base path /invalid not present in FS"}
    driver_object.test_dynamic(value_sc)


def test_driver_invalid_input_266():
    value_sc = {"clusterId":  data["remoteid"], "filesetType": "dependent",
                "volBackendFs": invaliddata["remoteFs"],
                "reason": "Unable to get Mount Details for FS"}
    driver_object.test_dynamic(value_sc)


def test_driver_invalid_input_267():
    value_sc = {"clusterId":  data["remoteid"], "filesetType": "dependent",
                "volBackendFs": data["remoteFs"],
                "parentFileset": invaliddata["r-parentFileset"],
                "reason": "Unable to create fileset"}
    driver_object.test_dynamic(value_sc)


def test_driver_invalid_input_268():
    value_sc = {"clusterId":  data["remoteid"], "filesetType": "dependent",
                "volBackendFs": data["remoteFs"],
                "inodeLimit": invaliddata["r-inodeLimit"],
                "reason": "inodeLimit and filesetType=dependent must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_invalid_input_269():
    value_sc = {"clusterId":  data["remoteid"], "filesetType": "independent",
                "volBackendFs": data["remoteFs"],
                "inodeLimit": invaliddata["r-inodeLimit"],
                "reason": "inodeLimit specified in storageClass must be equal to or greater than 1024"}
    driver_object.test_dynamic(value_sc)


def test_driver_invalid_input_270():
    value_sc = {"clusterId":  data["remoteid"], "volBackendFs": data["remoteFs"],
                "filesetType": "dependent", "gid": invaliddata["r-gid_number"],
                "uid": invaliddata["r-uid_number"],
                "reason": 'The object "9999" specified for "user" does not exist'}
    driver_object.test_dynamic(value_sc)


def test_driver_invalid_input_271():
    value_sc = {"clusterId":  data["remoteid"], "volBackendFs": data["remoteFs"],
                "filesetType": "dependent", "gid": data["r-gid_number"],
                "uid": invaliddata["r-uid_name"],
                "reason": 'The object "invalid_name" specified for "user" does not exist'}
    driver_object.test_dynamic(value_sc)


def test_driver_invalid_input_272():
    value_sc = {"clusterId":  data["remoteid"], "volBackendFs": data["remoteFs"],
                "filesetType": "dependent", "gid": invaliddata["r-gid_name"],
                "uid": invaliddata["r-uid_name"],
                "reason": 'The object "invalid_name" specified for "user" does not exist'}
    driver_object.test_dynamic(value_sc)


def test_driver_invalid_input_273():
    value_sc = {"clusterId":  data["remoteid"], "volBackendFs": data["remoteFs"],
                "filesetType": "dependent", "gid": invaliddata["r-gid_name"],
                "uid": data["r-uid_number"],
                "reason": 'The object "invalid_name" specified for "group" does not exist'}
    driver_object.test_dynamic(value_sc)


def test_driver_invalid_input_274():
    value_sc = {"clusterId":  data["remoteid"], "volBackendFs": data["remoteFs"],
                "gid": data["r-gid_number"], "uid": data["r-uid_number"],
                "parentFileset": invaliddata["r-parentFileset"],
                "reason": "parentFileset and filesetType=independent"}
    driver_object.test_dynamic(value_sc)


def test_driver_one_pvc_two_pod():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId":  data["remoteid"]}
    driver_object.one_pvc_two_pod(value_sc)


def test_driver_parallel_pvc():
    value_sc = {"volBackendFs": data["remoteFs"], "clusterId":  data["remoteid"]}
    driver_object.parallel_pvc(value_sc)
