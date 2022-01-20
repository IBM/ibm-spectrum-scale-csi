import logging
import pytest
import ibm_spectrum_scale_csi.scale_operator as scaleop
import ibm_spectrum_scale_csi.spectrum_scale_apis.fileset_functions as ff
LOGGER = logging.getLogger()
pytestmark = [pytest.mark.volumeprovisioning, pytest.mark.localcluster]


@pytest.fixture(scope='session', autouse=True)
def values(request, check_csi_operator):
    global data, driver_object, kubeconfig_value  # are required in every testcase
    kubeconfig_value, clusterconfig_value, operator_namespace, test_namespace, runslow_val, operator_yaml = scaleop.get_cmd_values(request)

    data = scaleop.read_driver_data(clusterconfig_value, test_namespace, operator_namespace, kubeconfig_value)
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
    ff.get_scale_version(data)
    scaleop.get_kubernetes_version(kubeconfig_value)
    scaleop.scale_function.get_operator_image()
    scaleop.ob.get_driver_image()


@pytest.mark.regression
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
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    value_pv = {"access_modes": "ReadWriteMany", "storage": "1Gi",
                "reclaim_policy": "Retain"}
    value_pvc_custom = [{"access_modes": "ReadWriteMany", "storage": "1Gi"},
                        {"access_modes": "ReadWriteOnce", "storage": "1Gi",
                            "reason": "incompatible accessMode"},
                        {"access_modes": "ReadOnlyMany", "storage": "1Gi",
                            "reason": "incompatible accessMode"}
                        ]
    driver_object.test_static(value_pv, value_pvc_custom, value_sc)


@pytest.mark.regression
def test_driver_static_sc_14():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    value_pv = {"access_modes": "ReadWriteMany", "storage": "1Gi",
                "reclaim_policy": "Delete"}
    value_pvc_custom = [{"access_modes": "ReadWriteMany", "storage": "1Gi"},
                        {"access_modes": "ReadWriteOnce", "storage": "1Gi",
                            "reason": "incompatible accessMode"},
                        {"access_modes": "ReadOnlyMany", "storage": "1Gi",
                            "reason": "incompatible accessMode"}
                        ]
    driver_object.test_static(value_pv, value_pvc_custom, value_sc)


def test_driver_static_sc_15():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
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
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
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
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
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
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
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
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    value_pv = {"access_modes": "ReadOnlyMany", "storage": "1Gi",
                "reclaim_policy": "Retain"}
    value_pvc_custom = [{"access_modes": "ReadWriteMany", "storage": "1Gi",
                         "reason": "incompatible accessMode"},
                        {"access_modes": "ReadWriteOnce", "storage": "1Gi",
                            "reason": "incompatible accessMode"},
                        {"access_modes": "ReadOnlyMany", "storage": "1Gi"}
                        ]
    driver_object.test_static(value_pv, value_pvc_custom, value_sc)


def test_driver_static_sc_20():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    value_pv = {"access_modes": "ReadOnlyMany", "storage": "1Gi",
                "reclaim_policy": "Delete"}
    value_pvc_custom = [{"access_modes": "ReadWriteMany", "storage": "1Gi",
                         "reason": "incompatible accessMode"},
                        {"access_modes": "ReadWriteOnce", "storage": "1Gi",
                            "reason": "incompatible accessMode"},
                        {"access_modes": "ReadOnlyMany", "storage": "1Gi"}
                        ]
    driver_object.test_static(value_pv, value_pvc_custom, value_sc)


def test_driver_static_sc_21():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
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
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
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
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    value_pv = {"access_modes": "ReadWriteOnce", "storage": "1Gi",
                "reclaim_policy": "Default"}
    value_pvc_custom = [{"access_modes": "ReadWriteMany", "storage": "1Gi",
                         "reason": "incompatible accessMode"},
                        {"access_modes": "ReadWriteOnce", "storage": "1Gi"},
                        {"access_modes": "ReadOnlyMany", "storage": "1Gi",
                            "reason": "incompatible accessMode"}
                        ]
    driver_object.test_static(value_pv, value_pvc_custom, value_sc)


def test_driver_static_sc_24():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    value_pv = {"access_modes": "ReadOnlyMany", "storage": "1Gi", "reclaim_policy": "Default"}
    value_pvc_custom = [{"access_modes": "ReadWriteMany", "storage": "1Gi",
                         "reason": "incompatible accessMode"},
                        {"access_modes": "ReadWriteOnce", "storage": "1Gi",
                            "reason": "incompatible accessMode"},
                        {"access_modes": "ReadOnlyMany", "storage": "1Gi"}
                        ]
    driver_object.test_static(value_pv, value_pvc_custom, value_sc)


"""
def test_driver_static_27():
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


def test_driver_static_26():
    LOGGER.info("PV creation with fileset root path as lightweight volume")
    value_pvc_custom = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_pv = {"access_modes": "ReadWriteMany",
                "storage": "1Gi", "reclaim_policy": "Default"}
    driver_object.test_static(value_pv, value_pvc_custom, root_volume=True)


@pytest.mark.regression
def test_driver_dynamic_pass_1():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_2():
    value_sc = {"volBackendFs": data["primaryFs"],
                "clusterId": data["id"], "uid": data["uid_number"]}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_3():
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"},
                 {"access_modes": "ReadWriteOnce", "storage": "1Gi"},
                 {"access_modes": "ReadOnlyMany", "storage": "1Gi",
                     "reason": "ReadOnlyMany is not supported"}
                 ]
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False"},
                 {"mount_path": "/usr/share/nginx/html/scale",
                     "read_only": "True", "reason": "Read-only file system"}
                 ]

    value_sc = {"volBackendFs": data["primaryFs"],
                "clusterId": data["id"], "gid": data["gid_number"]}
    driver_object.test_dynamic(value_sc, value_pvc, value_pod)


def test_driver_dynamic_pass_4():
    value_sc = {"volBackendFs": data["primaryFs"],
                "clusterId": data["id"], "inodeLimit": data["inodeLimit"]}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_5():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"],
                "inodeLimit": data["inodeLimit"], "uid": data["uid_number"]}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_6():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_7():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"],
                "inodeLimit": data["inodeLimit"], "gid": data["gid_number"]}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_8():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"],
                "inodeLimit": data["inodeLimit"], "uid": data["uid_number"],
                "gid": data["gid_number"]}
    driver_object.test_dynamic(value_sc)


@pytest.mark.regression
def test_driver_dynamic_pass_9():
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"},
                 {"access_modes": "ReadWriteOnce", "storage": "1Gi"},
                 {"access_modes": "ReadOnlyMany", "storage": "1Gi",
                     "reason": "ReadOnlyMany is not supported"}
                 ]
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False"},
                 {"mount_path": "/usr/share/nginx/html/scale",
                     "read_only": "True", "reason": "Read-only file system"}
                 ]

    value_sc = {"volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"]}
    driver_object.test_dynamic(value_sc, value_pvc, value_pod)


def test_driver_dynamic_pass_10():
    value_sc = {"volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "clusterId": data["id"]}
    driver_object.test_dynamic(value_sc)


@pytest.mark.regression
def test_driver_dynamic_pass_11():
    value_sc = {"volBackendFs": data["primaryFs"],
                "filesetType": "dependent", "clusterId": data["id"]}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_12():
    value_sc = {"volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "uid": data["uid_number"]}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_13():
    value_sc = {"volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"]}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_14():
    value_sc = {"volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "uid": data["uid_number"], "clusterId": data["id"]}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_15():
    value_sc = {"volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "clusterId": data["id"]}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_16():
    value_sc = {"volBackendFs": data["primaryFs"], "uid": data["uid_number"],
                "clusterId": data["id"], "filesetType": "dependent"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_17():
    value_sc = {"volBackendFs": data["primaryFs"], "gid": data["gid_number"],
                "clusterId": data["id"], "filesetType": "dependent"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_18():
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"},
                 {"access_modes": "ReadWriteOnce", "storage": "1Gi"},
                 {"access_modes": "ReadOnlyMany", "storage": "1Gi",
                     "reason": "ReadOnlyMany is not supported"}
                 ]
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False"},
                 {"mount_path": "/usr/share/nginx/html/scale",
                     "read_only": "True", "reason": "Read-only file system"}
                 ]

    value_sc = {"volBackendFs": data["primaryFs"],
                "parentFileset": data["parentFileset"],
                "clusterId": data["id"], "filesetType": "dependent"}
    driver_object.test_dynamic(value_sc, value_pvc, value_pod)


def test_driver_dynamic_pass_19():
    value_sc = {"volBackendFs": data["primaryFs"],
                "inodeLimit": data["inodeLimit"],
                "clusterId": data["id"], "filesetType": "independent"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_20():
    value_sc = {"volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"]}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_21():
    value_sc = {"volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"],
                "clusterId": data["id"]}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_22():
    value_sc = {"volBackendFs": data["primaryFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "clusterId": data["id"],
                "filesetType": "dependent"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_23():
    value_sc = {"volBackendFs": data["primaryFs"], "gid": data["gid_name"],
                "uid": data["uid_number"], "clusterId": data["id"],
                "filesetType": "dependent"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_24():
    value_sc = {"volBackendFs": data["primaryFs"], "gid": data["gid_number"],
                "uid": data["uid_name"], "clusterId": data["id"],
                "filesetType": "dependent"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_25():
    value_sc = {"volBackendFs": data["primaryFs"], "gid": data["gid_name"],
                "uid": data["uid_name"], "clusterId": data["id"],
                "filesetType": "dependent"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_26():
    value_sc = {"volBackendFs": data["primaryFs"], "uid": data["uid_number"],
                "clusterId": data["id"], "filesetType": "dependent",
                "parentFileset": data["parentFileset"]}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_27():
    value_sc = {"volBackendFs": data["primaryFs"], "gid": data["gid_number"],
                "clusterId": data["id"], "filesetType": "dependent",
                "parentFileset": data["parentFileset"]}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_28():
    value_sc = {"volBackendFs": data["primaryFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "clusterId": data["id"],
                "filesetType": "dependent",
                "parentFileset": data["parentFileset"]}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_29():
    value_sc = {"volBackendFs": data["primaryFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "clusterId": data["id"],
                "filesetType": "independent", "inodeLimit": data["inodeLimit"]}
    driver_object.test_dynamic(value_sc)

#   Testcases expected to fail with valid values of parameters


def test_driver_dynamic_pass_30():
    value_sc = {"volBackendFs": data["primaryFs"]}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_31():
    value_sc = {"clusterId": data["id"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_32():
    value_sc = {"gid": data["gid_number"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_33():
    value_sc = {"uid": data["uid_number"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_34():
    value_sc = {"filesetType": "dependent", "volBackendFs": data["primaryFs"]}
    driver_object.test_dynamic(value_sc)


@pytest.mark.regression
def test_driver_dynamic_fail_35():
    value_sc = {"inodeLimit": data["inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_36():
    value_sc = {"volDirBasePath": data["volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_37():
    value_sc = {"parentFileset": data["parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_38():
    value_sc = {"uid": data["uid_number"], "volBackendFs": data["primaryFs"]}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_39():
    value_sc = {"gid": data["gid_number"], "volBackendFs": data["primaryFs"]}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_40():
    value_sc = {"filesetType": "dependent", "volBackendFs": data["primaryFs"]}
    driver_object.test_dynamic(value_sc)


@pytest.mark.regression
def test_driver_dynamic_fail_41():
    value_sc = {"parentFileset": data["parentFileset"],
                "volBackendFs": data["primaryFs"],
                "reason": "InvalidArgument desc = parentFileset and filesetType=independent"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_42():
    value_sc = {"inodeLimit": data["inodeLimit"],
                "volBackendFs": data["primaryFs"]}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_43():
    value_sc = {"clusterId": data["id"],
                "volDirBasePath": data["volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_44():
    value_sc = {"clusterId": data["id"], "uid": data["uid_number"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_45():
    value_sc = {"clusterId": data["id"], "gid": data["gid_number"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_46():
    value_sc = {"clusterId": data["id"], "filesetType": "dependent",
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_47():
    value_sc = {"clusterId": data["id"], "parentFileset": data["parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_48():
    value_sc = {"clusterId": data["id"], "inodeLimit": data["inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_49():
    value_sc = {"volDirBasePath": data["volDirBasePath"],
                "uid": data["uid_number"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_50():
    value_sc = {"volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_51():
    value_sc = {"volDirBasePath": data["volDirBasePath"],
                "filesetType": "dependent",
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_52():
    value_sc = {"volDirBasePath": data["volDirBasePath"],
                "parentFileset": data["parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_53():
    value_sc = {"volDirBasePath": data["volDirBasePath"],
                "inodeLimit": data["inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_54():
    value_sc = {"uid": data["uid_number"], "gid": data["gid_number"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_55():
    value_sc = {"uid": data["uid_number"], "filesetType": "dependent",
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_56():
    value_sc = {"uid": data["uid_number"],
                "parentFileset": data["parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_57():
    value_sc = {"uid": data["uid_number"], "inodeLimit": data["inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_58():
    value_sc = {"gid": data["gid_number"], "filesetType": "dependent",
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_59():
    value_sc = {"gid": data["gid_number"],
                "parentFileset": data["parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_60():
    value_sc = {"gid": data["gid_number"], "inodeLimit": data["inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_61():
    value_sc = {"filesetType": "dependent",
                "parentFileset": data["parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_62():
    value_sc = {"filesetType": "dependent",
                "inodeLimit": data["inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_63():
    value_sc = {"inodeLimit": data["inodeLimit"],
                "parentFileset": data["parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_64():
    value_sc = {"clusterId": data["id"], "volBackendFs": data["primaryFs"],
                "parentFileset": data["parentFileset"],
                "reason": "InvalidArgument desc = parentFileset and filesetType=independent"}
    driver_object.test_dynamic(value_sc)


@pytest.mark.regression
def test_driver_dynamic_fail_65():
    value_sc = {"volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "filesetType": "dependent",
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_66():
    value_sc = {"volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "parentFileset": data["parentFileset"],
                "reason": "parentFileset and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_67():
    value_sc = {"volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "inodeLimit": data["inodeLimit"],
                "reason": "inodeLimit and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_68():
    value_sc = {"volBackendFs": data["primaryFs"], "uid": data["uid_number"],
                "gid": data["gid_number"]}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_69():
    value_sc = {"volBackendFs": data["primaryFs"], "uid": data["uid_number"],
                "filesetType": "dependent"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_70():
    value_sc = {"volBackendFs": data["primaryFs"], "uid": data["uid_number"],
                "parentFileset": data["parentFileset"],
                "reason": "InvalidArgument desc = parentFileset and filesetType=independent"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_71():
    value_sc = {"volBackendFs": data["primaryFs"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"]}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_72():
    value_sc = {"volBackendFs": data["primaryFs"], "gid": data["gid_number"],
                "filesetType": "dependent"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_73():
    value_sc = {"volBackendFs": data["primaryFs"], "gid": data["gid_number"],
                "parentFileset": data["parentFileset"],
                "reason": "InvalidArgument desc = parentFileset and filesetType=independent"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_74():
    value_sc = {"volBackendFs": data["primaryFs"], "gid": data["gid_number"],
                "inodeLimit": data["inodeLimit"]}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_75():
    value_sc = {"volBackendFs": data["primaryFs"], "filesetType": "dependent",
                "parentFileset": data["parentFileset"]}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_76():
    value_sc = {"volBackendFs": data["primaryFs"], "filesetType": "dependent",
                "inodeLimit": data["inodeLimit"],
                "reason": "inodeLimit and filesetType=dependent must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_77():
    value_sc = {"volBackendFs": data["primaryFs"],
                "parentFileset": data["parentFileset"],
                "inodeLimit": data["inodeLimit"],
                "reason": "InvalidArgument desc = parentFileset and filesetType=independent"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_78():
    value_sc = {"clusterId": data["id"],
                "volDirBasePath": data["volDirBasePath"],
                "uid": data["uid_number"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_79():
    value_sc = {"clusterId": data["id"],
                "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_80():
    value_sc = {"clusterId": data["id"],
                "volDirBasePath": data["volDirBasePath"],
                "filesetType": "dependent", "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_81():
    value_sc = {"clusterId": data["id"],
                "volDirBasePath": data["volDirBasePath"],
                "parentFileset": data["parentFileset"], "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_82():
    value_sc = {"clusterId": data["id"],
                "volDirBasePath": data["volDirBasePath"],
                "inodeLimit": data["inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_83():
    value_sc = {"clusterId": data["id"], "uid": data["uid_number"],
                "gid": data["gid_number"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_84():
    value_sc = {"clusterId": data["id"], "uid": data["uid_number"],
                "filesetType": "dependent",
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_85():
    value_sc = {"clusterId": data["id"], "uid": data["uid_number"],
                "parentFileset": data["parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_86():
    value_sc = {"clusterId": data["id"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_87():
    value_sc = {"clusterId": data["id"], "gid": data["gid_number"],
                "filesetType": "dependent",
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_88():
    value_sc = {"clusterId": data["id"], "gid": data["gid_number"],
                "parentFileset": data["parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_89():
    value_sc = {"clusterId": data["id"], "gid": data["gid_number"],
                "inodeLimit": data["inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_90():
    value_sc = {"clusterId": data["id"], "filesetType": "dependent",
                "parentFileset": data["parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_91():
    value_sc = {"clusterId": data["id"], "inodeLimit": data["inodeLimit"],
                "filesetType": "dependent",
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_92():
    value_sc = {"clusterId": data["id"], "inodeLimit": data["inodeLimit"],
                "parentFileset": data["parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_93():
    value_sc = {"volDirBasePath": data["volDirBasePath"],
                "uid": data["uid_number"], "gid": data["gid_number"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_94():
    value_sc = {"volDirBasePath": data["volDirBasePath"],
                "uid": data["uid_number"], "filesetType": "dependent",
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_95():
    value_sc = {"volDirBasePath": data["volDirBasePath"],
                "uid": data["uid_number"],
                "parentFileset": data["parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_96():
    value_sc = {"volDirBasePath": data["volDirBasePath"],
                "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_97():
    value_sc = {"volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"],
                "filesetType": "dependent",
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_98():
    value_sc = {"volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"],
                "parentFileset": data["parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_99():
    value_sc = {"volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"],
                "inodeLimit": data["inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_100():
    value_sc = {"volDirBasePath": data["volDirBasePath"],
                "filesetType": "dependent",
                "parentFileset": data["parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_101():
    value_sc = {"volDirBasePath": data["volDirBasePath"],
                "filesetType": "dependent",
                "inodeLimit": data["inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_102():
    value_sc = {"volDirBasePath": data["volDirBasePath"],
                "parentFileset": data["parentFileset"],
                "inodeLimit": data["inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_103():
    value_sc = {"uid": data["uid_number"], "gid": data["gid_number"],
                "filesetType": "dependent",
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_104():
    value_sc = {"uid": data["uid_number"], "gid": data["gid_number"],
                "parentFileset": data["parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_105():
    value_sc = {"uid": data["uid_number"], "gid": data["gid_number"],
                "inodeLimit": data["inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_106():
    value_sc = {"uid": data["uid_number"], "filesetType": "dependent",
                "parentFileset": data["parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_107():
    value_sc = {"uid": data["uid_number"], "filesetType": "dependent",
                "inodeLimit": data["inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_108():
    value_sc = {"uid": data["uid_number"],
                "parentFileset": data["parentFileset"],
                "inodeLimit": data["inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_109():
    value_sc = {"gid": data["gid_number"], "filesetType": "dependent",
                "parentFileset": data["parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_110():
    value_sc = {"gid": data["gid_number"], "filesetType": "dependent",
                "inodeLimit": data["inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_111():
    value_sc = {"gid": data["gid_number"],
                "parentFileset": data["parentFileset"],
                "inodeLimit": data["inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_112():
    value_sc = {"filesetType": "dependent",
                "parentFileset": data["parentFileset"],
                "inodeLimit": data["inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_113():
    value_sc = {"clusterId": data["id"],  "filesetType":  "dependent",
                "volBackendFs":  data["primaryFs"],
                "volDirBasePath":  data["volDirBasePath"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_114():
    value_sc = {"clusterId": data["id"], "volBackendFs": data["primaryFs"],
                "parentFileset": data["parentFileset"],
                "volDirBasePath": data["volDirBasePath"],
                "reason": "parentFileset and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_115():
    value_sc = {"clusterId": data["id"], "inodeLimit": data["inodeLimit"],
                "volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "reason": "inodeLimit and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_116():
    value_sc = {"clusterId": data["id"], "gid": data["gid_number"],
                "uid": data["uid_number"], "volBackendFs": data["primaryFs"]}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_117():
    value_sc = {"clusterId": data["id"], "uid": data["uid_number"],
                "volBackendFs": data["primaryFs"],
                "parentFileset": data["parentFileset"],
                "reason": "InvalidArgument desc = parentFileset and filesetType=independent"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_118():
    value_sc = {"clusterId": data["id"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"],
                "volBackendFs": data["primaryFs"]}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_119():
    value_sc = {"clusterId": data["id"], "gid": data["gid_number"],
                "volBackendFs": data["primaryFs"],
                "parentFileset": data["parentFileset"],
                "reason": "InvalidArgument desc = parentFileset and filesetType=independent"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_120():
    value_sc = {"clusterId": data["id"], "gid": data["gid_number"],
                "inodeLimit": data["inodeLimit"],
                "volBackendFs": data["primaryFs"]}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_121():
    value_sc = {"clusterId": data["id"], "filesetType": "independent",
                "volBackendFs": data["primaryFs"],
                "parentFileset": data["parentFileset"],
                "reason": "InvalidArgument desc = parentFileset and filesetType=independent"}
    driver_object.test_dynamic(value_sc)


@pytest.mark.regression
def test_driver_dynamic_fail_122():
    value_sc = {"clusterId": data["id"], "inodeLimit": data["inodeLimit"],
                "filesetType": "dependent", "volBackendFs": data["primaryFs"],
                "reason": "inodeLimit and filesetType=dependent must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_123():
    value_sc = {"clusterId": data["id"], "inodeLimit": data["inodeLimit"],
                "volBackendFs": data["primaryFs"],
                "parentFileset": data["parentFileset"],
                "reason": "InvalidArgument desc = parentFileset and filesetType=independent"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_124():
    value_sc = {"uid": data["uid_number"], "filesetType": "dependent",
                "volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_125():
    value_sc = {"uid": data["uid_number"], "volBackendFs": data["primaryFs"],
                "parentFileset": data["parentFileset"],
                "volDirBasePath": data["volDirBasePath"],
                "reason": "parentFileset and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_126():
    value_sc = {"uid": data["uid_number"], "inodeLimit": data["inodeLimit"],
                "volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "reason": "inodeLimit and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_127():
    value_sc = {"gid": data["gid_number"], "filesetType": "dependent",
                "volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_128():
    value_sc = {"gid": data["gid_number"], "volBackendFs": data["primaryFs"],
                "parentFileset": data["parentFileset"],
                "volDirBasePath": data["volDirBasePath"],
                "reason": "parentFileset and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_129():
    value_sc = {"gid": data["gid_number"], "inodeLimit": data["inodeLimit"],
                "volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "reason": "inodeLimit and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_130():
    value_sc = {"filesetType": "dependent", "volBackendFs": data["primaryFs"],
                "parentFileset": data["parentFileset"],
                "volDirBasePath": data["volDirBasePath"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_131():
    value_sc = {"inodeLimit": data["inodeLimit"], "filesetType": "dependent",
                "volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_132():
    value_sc = {"inodeLimit": data["inodeLimit"],
                "volBackendFs": data["primaryFs"],
                "parentFileset": data["parentFileset"],
                "volDirBasePath": data["volDirBasePath"],
                "reason": "parentFileset and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_133():
    value_sc = {"gid": data["gid_number"], "uid": data["uid_number"],
                "filesetType": "dependent",
                "volBackendFs": data["primaryFs"]}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_134():
    value_sc = {"gid": data["gid_number"], "uid": data["uid_number"],
                "volBackendFs": data["primaryFs"],
                "parentFileset": data["parentFileset"],
                "reason": "InvalidArgument desc = parentFileset and filesetType=independent"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_135():
    value_sc = {"gid": data["gid_number"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"],
                "volBackendFs": data["primaryFs"]}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_136():
    value_sc = {"uid": data["uid_number"], "filesetType": "dependent",
                "volBackendFs": data["primaryFs"],
                "parentFileset": data["parentFileset"]}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_137():
    value_sc = {"uid": data["uid_number"], "inodeLimit": data["inodeLimit"],
                "filesetType": "dependent", "volBackendFs": data["primaryFs"],
                "reason": "inodeLimit and filesetType=dependent must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_138():
    value_sc = {"uid": data["uid_number"], "inodeLimit": data["inodeLimit"],
                "volBackendFs": data["primaryFs"],
                "parentFileset": data["parentFileset"],
                "reason": "InvalidArgument desc = parentFileset and filesetType=independent"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_139():
    value_sc = {"gid": data["gid_number"], "filesetType": "dependent",
                "volBackendFs": data["primaryFs"],
                "parentFileset": data["parentFileset"]}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_140():
    value_sc = {"gid": data["gid_number"], "inodeLimit": data["inodeLimit"],
                "filesetType": "dependent", "volBackendFs": data["primaryFs"],
                "reason": "inodeLimit and filesetType=dependent must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_141():
    value_sc = {"gid": data["gid_number"], "inodeLimit": data["inodeLimit"],
                "volBackendFs": data["primaryFs"],
                "parentFileset": data["parentFileset"],
                "reason": "InvalidArgument desc = parentFileset and filesetType=independent"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_142():
    value_sc = {"inodeLimit": data["inodeLimit"], "filesetType": "dependent",
                "volBackendFs": data["primaryFs"],
                "parentFileset": data["parentFileset"],
                "reason": "inodeLimit and filesetType=dependent must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_143():
    value_sc = {"clusterId": data["id"], "gid": data["gid_number"],
                "uid": data["uid_number"],
                "volDirBasePath": data["volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_144():
    value_sc = {"clusterId": data["id"], "uid": data["uid_number"],
                "filesetType": "dependent",
                "volDirBasePath": data["volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_145():
    value_sc = {"clusterId": data["id"], "uid": data["uid_number"],
                "parentFileset": data["parentFileset"],
                "volDirBasePath": data["volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_146():
    value_sc = {"clusterId": data["id"], "uid": data["uid_number"],
                "volDirBasePath": data["volDirBasePath"],
                "inodeLimit": data["inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_147():
    value_sc = {"clusterId": data["id"], "gid": data["gid_number"],
                "filesetType": "dependent", "volDirBasePath": data["volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_148():
    value_sc = {"clusterId": data["id"], "gid": data["gid_number"],
                "parentFileset": data["parentFileset"],
                "volDirBasePath": data["volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_149():
    value_sc = {"clusterId": data["id"], "gid": data["gid_number"],
                "inodeLimit": data["inodeLimit"],
                "volDirBasePath": data["volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_150():
    value_sc = {"clusterId": data["id"], "filesetType": "dependent",
                "parentFileset": data["parentFileset"],
                "volDirBasePath": data["volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_151():
    value_sc = {"clusterId": data["id"], "inodeLimit": data["inodeLimit"],
                "filesetType": "dependent",
                "volDirBasePath": data["volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_152():
    value_sc = {"clusterId": data["id"], "inodeLimit": data["inodeLimit"],
                "parentFileset": data["parentFileset"],
                "volDirBasePath": data["volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_153():
    value_sc = {"clusterId": data["id"], "gid": data["gid_number"],
                "uid": data["uid_number"], "filesetType": "dependent",
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_154():
    value_sc = {"clusterId": data["id"], "gid": data["gid_number"],
                "uid": data["uid_number"],
                "parentFileset": data["parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_155():
    value_sc = {"clusterId": data["id"], "gid": data["gid_number"],
                "uid": data["uid_number"], "inodeLimit": data["inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_156():
    value_sc = {"clusterId": data["id"], "uid": data["uid_number"],
                "filesetType": "dependent",
                "parentFileset": data["parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_157():
    value_sc = {"clusterId": data["id"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"],
                "filesetType": "dependent",
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_158():
    value_sc = {"clusterId": data["id"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"],
                "parentFileset": data["parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_159():
    value_sc = {"clusterId": data["id"], "gid": data["gid_number"],
                "filesetType": "dependent",
                "parentFileset": data["parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_160():
    value_sc = {"clusterId": data["id"], "gid": data["gid_number"],
                "inodeLimit": data["inodeLimit"],
                "filesetType": "dependent",
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_161():
    value_sc = {"clusterId": data["id"],
                "gid": data["gid_number"], "inodeLimit": data["inodeLimit"],
                "parentFileset": data["parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_162():
    value_sc = {"clusterId": data["id"], "inodeLimit": data["inodeLimit"],
                "filesetType": "dependent",
                "parentFileset": data["parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_163():
    value_sc = {"gid": data["gid_number"], "uid": data["uid_number"],
                "filesetType": "dependent",
                "volDirBasePath": data["volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_164():
    value_sc = {"gid": data["gid_number"], "uid": data["uid_number"],
                "parentFileset": data["parentFileset"],
                "volDirBasePath": data["volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_165():
    value_sc = {"gid": data["gid_number"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"],
                "volDirBasePath": data["volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_166():
    value_sc = {"uid": data["uid_number"], "filesetType": "dependent",
                "parentFileset": data["parentFileset"],
                "volDirBasePath": data["volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_167():
    value_sc = {"uid": data["uid_number"], "inodeLimit": data["inodeLimit"],
                "filesetType": "dependent",
                "volDirBasePath": data["volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_168():
    value_sc = {"uid": data["uid_number"], "inodeLimit": data["inodeLimit"],
                "parentFileset": data["parentFileset"],
                "volDirBasePath": data["volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_169():
    value_sc = {"gid": data["gid_number"], "filesetType": "dependent",
                "parentFileset": data["parentFileset"],
                "volDirBasePath": data["volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_170():
    value_sc = {"gid": data["gid_number"], "inodeLimit": data["inodeLimit"],
                "filesetType": "dependent",
                "volDirBasePath": data["volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_171():
    value_sc = {"gid": data["gid_number"], "inodeLimit": data["inodeLimit"],
                "parentFileset": data["parentFileset"],
                "volDirBasePath": data["volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_172():
    value_sc = {"inodeLimit": data["inodeLimit"], "filesetType": "dependent",
                "parentFileset": data["parentFileset"],
                "volDirBasePath": data["volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_173():
    value_sc = {"gid": data["gid_number"], "uid": data["uid_number"],
                "filesetType": "dependent",
                "parentFileset": data["parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_174():
    value_sc = {"gid": data["gid_number"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"],
                "filesetType": "dependent",
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_175():
    value_sc = {"gid": data["gid_number"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"],
                "parentFileset": data["parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_176():
    value_sc = {"uid": data["uid_number"], "inodeLimit": data["inodeLimit"],
                "filesetType": "dependent",
                "parentFileset": data["parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_177():
    value_sc = {"gid": data["gid_number"], "inodeLimit": data["inodeLimit"],
                "filesetType": "dependent",
                "parentFileset": data["parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_178():
    value_sc = {"clusterId": data["id"], "volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "uid": data["uid_number"],
                "filesetType": "dependent",
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_179():
    value_sc = {"clusterId": data["id"], "volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "uid": data["uid_number"],
                "parentFileset": data["parentFileset"],
                "reason": "parentFileset and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_180():
    value_sc = {"clusterId": data["id"], "volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"],
                "reason": "inodeLimit and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_181():
    value_sc = {"clusterId": data["id"], "volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"],
                "filesetType": "dependent",
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_182():
    value_sc = {"clusterId": data["id"], "volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"],
                "parentFileset": data["parentFileset"],
                "reason": "parentFileset and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_183():
    value_sc = {"clusterId": data["id"], "volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"],
                "inodeLimit": data["inodeLimit"],
                "reason": "inodeLimit and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_184():
    value_sc = {"clusterId": data["id"], "volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "filesetType": "dependent",
                "parentFileset": data["parentFileset"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_185():
    value_sc = {"clusterId": data["id"], "volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "inodeLimit": data["inodeLimit"],
                "filesetType": "dependent",
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_186():
    value_sc = {"clusterId": data["id"], "volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "inodeLimit": data["inodeLimit"],
                "parentFileset": data["parentFileset"],
                "reason": "parentFileset and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_187():
    value_sc = {"clusterId": data["id"], "volBackendFs": data["primaryFs"],
                "gid": data["gid_number"], "uid": data["uid_number"],
                "parentFileset": data["parentFileset"],
                "reason": "InvalidArgument desc = parentFileset and filesetType=independent"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_188():
    value_sc = {"clusterId": data["id"], "volBackendFs": data["primaryFs"],
                "gid": data["gid_number"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"]}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_189():
    value_sc = {"clusterId": data["id"], "volBackendFs": data["primaryFs"],
                "uid": data["uid_number"], "filesetType": "independent",
                "parentFileset": data["parentFileset"],
                "reason": "InvalidArgument desc = parentFileset and filesetType=independent"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_190():
    value_sc = {"clusterId": data["id"], "volBackendFs": data["primaryFs"],
                "uid": data["uid_number"],
                "parentFileset": data["parentFileset"],
                "inodeLimit": data["inodeLimit"],
                "reason": "InvalidArgument desc = parentFileset and filesetType=independent"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_191():
    value_sc = {"clusterId": data["id"], "volBackendFs": data["primaryFs"],
                "gid": data["gid_number"],
                "parentFileset": data["parentFileset"],
                "inodeLimit": data["inodeLimit"], "reason": "InvalidArgument desc = parentFileset and filesetType=independent"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_192():
    value_sc = {"volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"],
                "filesetType": "dependent",
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_193():
    value_sc = {"volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"],
                "parentFileset": data["parentFileset"],
                "reason": "parentFileset and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_194():
    value_sc = {"volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"],
                "reason": "inodeLimit and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_195():
    value_sc = {"volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "uid": data["uid_number"], "filesetType": "dependent",
                "parentFileset": data["parentFileset"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_196():
    value_sc = {"volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "uid": data["uid_number"], "inodeLimit": data["inodeLimit"],
                "filesetType": "dependent",
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_197():
    value_sc = {"volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "uid": data["uid_number"],
                "parentFileset": data["parentFileset"],
                "inodeLimit": data["inodeLimit"],
                "reason": "parentFileset and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_198():
    value_sc = {"volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "filesetType": "dependent",
                "parentFileset": data["parentFileset"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_199():
    value_sc = {"volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"], "inodeLimit": data["inodeLimit"],
                "filesetType": "dependent",
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_200():
    value_sc = {"volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "gid": data["gid_number"],
                "parentFileset": data["parentFileset"],
                "inodeLimit": data["inodeLimit"],
                "reason": "parentFileset and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_201():
    value_sc = {"volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "parentFileset": data["parentFileset"],
                "inodeLimit": data["inodeLimit"], "filesetType": "dependent",
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_pass_202():
    value_sc = {"volBackendFs": data["primaryFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "filesetType": "dependent",
                "parentFileset": data["parentFileset"]}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_203():
    value_sc = {"volBackendFs": data["primaryFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "filesetType": "dependent",
                "inodeLimit": data["inodeLimit"],
                "reason": "inodeLimit and filesetType=dependent must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_204():
    value_sc = {"volBackendFs": data["primaryFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "parentFileset": data["parentFileset"],
                "inodeLimit": data["inodeLimit"],
                "reason": "InvalidArgument desc = parentFileset and filesetType=independent"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_205():
    value_sc = {"volBackendFs": data["primaryFs"], "uid": data["uid_number"],
                "filesetType": "dependent",
                "parentFileset": data["parentFileset"],
                "inodeLimit": data["inodeLimit"],
                "reason": "inodeLimit and filesetType=dependent must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_206():
    value_sc = {"volBackendFs": data["primaryFs"], "gid": data["gid_number"],
                "filesetType": "dependent",
                "parentFileset": data["parentFileset"],
                "inodeLimit": data["inodeLimit"],
                "reason": "inodeLimit and filesetType=dependent must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_207():
    value_sc = {"clusterId": data["id"], "gid": data["gid_number"],
                "uid": data["uid_number"], "filesetType": "dependent",
                "volDirBasePath": data["volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_208():
    value_sc = {"clusterId": data["id"], "gid": data["gid_number"],
                "uid": data["uid_number"],
                "volDirBasePath": data["volDirBasePath"],
                "parentFileset": data["parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_209():
    value_sc = {"clusterId": data["id"], "gid": data["gid_number"],
                "uid": data["uid_number"],
                "volDirBasePath": data["volDirBasePath"],
                "inodeLimit": data["inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_210():
    value_sc = {"clusterId": data["id"], "uid": data["uid_number"],
                "volDirBasePath": data["volDirBasePath"],
                "filesetType": "dependent",
                "parentFileset": data["parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_211():
    value_sc = {"clusterId": data["id"], "uid": data["uid_number"],
                "volDirBasePath": data["volDirBasePath"],
                "filesetType": "dependent", "inodeLimit": data["inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_212():
    value_sc = {"clusterId": data["id"], "uid": data["uid_number"],
                "volDirBasePath": data["volDirBasePath"],
                "inodeLimit": data["inodeLimit"],
                "parentFileset": data["parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_213():
    value_sc = {"clusterId": data["id"], "gid": data["gid_number"],
                "volDirBasePath": data["volDirBasePath"],
                "filesetType": "dependent",
                "parentFileset": data["parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_214():
    value_sc = {"clusterId": data["id"], "gid": data["gid_number"],
                "volDirBasePath": data["volDirBasePath"],
                "filesetType": "dependent", "inodeLimit": data["inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_215():
    value_sc = {"clusterId": data["id"], "gid": data["gid_number"],
                "volDirBasePath": data["volDirBasePath"],
                "inodeLimit": data["inodeLimit"],
                "parentFileset": data["parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_216():
    value_sc = {"clusterId": data["id"], "filesetType": "dependent",
                "volDirBasePath": data["volDirBasePath"],
                "inodeLimit": data["inodeLimit"],
                "parentFileset": data["parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_217():
    value_sc = {"clusterId": data["id"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"],
                "parentFileset": data["parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_218():
    value_sc = {"clusterId": data["id"], "filesetType": "dependent",
                "gid": data["gid_number"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_219():
    value_sc = {"clusterId": data["id"], "gid": data["gid_number"],
                "uid": data["uid_number"], "inodeLimit": data["inodeLimit"],
                "parentFileset": data["parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_220():
    value_sc = {"clusterId": data["id"], "filesetType": "dependent",
                "uid": data["uid_number"],
                "parentFileset": data["parentFileset"],
                "inodeLimit": data["inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_221():
    value_sc = {"clusterId": data["id"], "gid": data["gid_number"],
                "filesetType": "dependent",
                "parentFileset": data["parentFileset"],
                "inodeLimit": data["inodeLimit"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_222():
    value_sc = {"filesetType": "dependent", "gid": data["gid_number"],
                "uid": data["uid_number"],
                "volDirBasePath": data["volDirBasePath"],
                "parentFileset": data["parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_223():
    value_sc = {"filesetType": "dependent", "gid": data["gid_number"],
                "uid": data["uid_number"], "inodeLimit": data["inodeLimit"],
                "volDirBasePath": data["volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_224():
    value_sc = {"gid": data["gid_number"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"],
                "volDirBasePath": data["volDirBasePath"],
                "parentFileset": data["parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_225():
    value_sc = {"filesetType": "dependent", "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"],
                "volDirBasePath": data["volDirBasePath"],
                "parentFileset": data["parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_226():
    value_sc = {"filesetType": "dependent", "gid": data["gid_number"],
                "inodeLimit": data["inodeLimit"],
                "volDirBasePath": data["volDirBasePath"],
                "parentFileset": data["parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_227():
    value_sc = {"filesetType": "dependent", "gid": data["gid_number"],
                "uid": data["uid_number"], "inodeLimit": data["inodeLimit"],
                "parentFileset": data["parentFileset"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_228():
    value_sc = {"gid": data["gid_number"], "uid": data["uid_number"],
                "volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "clusterId":  data["id"], "filesetType": "dependent",
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_229():
    value_sc = {"gid": data["gid_number"], "uid": data["uid_number"],
                "volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "clusterId": data["id"], "parentFileset": data["parentFileset"],
                "reason": "parentFileset and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_230():
    value_sc = {"gid": data["gid_number"], "uid": data["uid_number"],
                "volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "clusterId": data["id"], "inodeLimit": data["inodeLimit"],
                "reason": "inodeLimit and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_231():
    value_sc = {"uid": data["uid_number"], "volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "clusterId": data["id"], "filesetType": "dependent",
                "parentFileset": data["parentFileset"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_232():
    value_sc = {"uid": data["uid_number"], "volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "clusterId": data["id"],
                "filesetType": "dependent", "inodeLimit": data["inodeLimit"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_233():
    value_sc = {"uid": data["uid_number"], "volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "clusterId": data["id"], "inodeLimit": data["inodeLimit"],
                "parentFileset": data["parentFileset"],
                "reason": "parentFileset and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_234():
    value_sc = {"gid": data["gid_number"], "volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "clusterId": data["id"], "filesetType": "dependent",
                "parentFileset": data["parentFileset"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_235():
    value_sc = {"gid": data["gid_number"], "volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "clusterId": data["id"],
                "filesetType": "dependent", "inodeLimit": data["inodeLimit"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_236():
    value_sc = {"gid": data["gid_number"], "volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"],
                "clusterId": data["id"], "parentFileset": data["parentFileset"],
                "inodeLimit": data["inodeLimit"],
                "reason": "parentFileset and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_237():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"],
                "volDirBasePath": data["volDirBasePath"],
                "filesetType": "dependent", "inodeLimit": data["inodeLimit"],
                "parentFileset": data["parentFileset"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_238():
    value_sc = {"clusterId": data["id"], "volBackendFs": data["primaryFs"],
                "uid": data["uid_number"], "inodeLimit": data["inodeLimit"],
                "parentFileset": data["parentFileset"],
                "gid": data["gid_number"],
                "reason": "InvalidArgument desc = parentFileset and filesetType=independent"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_239():
    value_sc = {"clusterId": data["id"], "volBackendFs": data["primaryFs"],
                "uid": data["uid_number"], "inodeLimit": data["inodeLimit"],
                "parentFileset": data["parentFileset"],
                "filesetType": "dependent",
                "reason": "inodeLimit and filesetType=dependent must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_240():
    value_sc = {"clusterId": data["id"], "volBackendFs": data["primaryFs"],
                "uid": data["uid_number"], "inodeLimit": data["inodeLimit"],
                "parentFileset": data["parentFileset"],
                "filesetType": "independent",
                "reason": "InvalidArgument desc = parentFileset and filesetType=independent"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_241():
    value_sc = {"volBackendFs": data["primaryFs"], "uid": data["uid_number"],
                "gid": data["gid_number"], "filesetType": "dependent",
                "parentFileset": data["parentFileset"],
                "volDirBasePath": data["volDirBasePath"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_242():
    value_sc = {"volBackendFs": data["primaryFs"], "uid": data["uid_number"],
                "gid": data["gid_number"], "inodeLimit": data["inodeLimit"],
                "filesetType": "dependent",
                "volDirBasePath": data["volDirBasePath"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_243():
    value_sc = {"volBackendFs": data["primaryFs"], "uid": data["uid_number"],
                "gid": data["gid_number"], "inodeLimit": data["inodeLimit"],
                "parentFileset": data["parentFileset"],
                "volDirBasePath": data["volDirBasePath"],
                "reason": "parentFileset and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_244():
    value_sc = {"volBackendFs": data["primaryFs"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"],
                "parentFileset": data["parentFileset"],
                "filesetType": "dependent",
                "volDirBasePath": data["volDirBasePath"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_245():
    value_sc = {"volBackendFs": data["primaryFs"], "gid": data["gid_number"],
                "inodeLimit": data["inodeLimit"],
                "parentFileset": data["parentFileset"],
                "filesetType": "dependent",
                "volDirBasePath": data["volDirBasePath"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_246():
    value_sc = {"volBackendFs": data["primaryFs"], "uid": data["uid_number"],
                "gid": data["gid_number"], "inodeLimit": data["inodeLimit"],
                "parentFileset": data["parentFileset"],
                "filesetType": "dependent",
                "reason": "inodeLimit and filesetType=dependent must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_247():
    value_sc = {"clusterId":  data["id"], "uid": data["uid_number"],
                "gid": data["gid_number"],
                "parentFileset": data["parentFileset"],
                "filesetType": "dependent",
                "volDirBasePath": data["volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_248():
    value_sc = {"clusterId": data["id"], "uid": data["uid_number"],
                "gid": data["gid_number"], "inodeLimit": data["inodeLimit"],
                "filesetType": "dependent",
                "volDirBasePath": data["volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_249():
    value_sc = {"clusterId": data["id"], "uid": data["uid_number"],
                "gid": data["gid_number"], "inodeLimit": data["inodeLimit"],
                "parentFileset": data["parentFileset"],
                "volDirBasePath": data["volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_250():
    value_sc = {"clusterId": data["id"], "uid": data["uid_number"],
                "inodeLimit": data["inodeLimit"],
                "parentFileset": data["parentFileset"],
                "filesetType": "dependent",
                "volDirBasePath": data["volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_251():
    value_sc = {"clusterId": data["id"], "gid": data["gid_number"],
                "inodeLimit": data["inodeLimit"],
                "parentFileset": data["parentFileset"],
                "filesetType": "dependent",
                "volDirBasePath": data["volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_252():
    value_sc = {"clusterId": data["id"], "uid": data["uid_number"],
                "gid": data["gid_number"], "inodeLimit": data["inodeLimit"],
                "parentFileset": data["parentFileset"],
                "filesetType": "dependent",
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_253():
    value_sc = {"uid": data["uid_number"], "gid": data["gid_number"],
                "inodeLimit": data["inodeLimit"],
                "parentFileset": data["parentFileset"],
                "filesetType": "dependent",
                "volDirBasePath": data["volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_254():
    value_sc = {"clusterId": data["id"], "volBackendFs": data["primaryFs"],
                "uid": data["uid_number"], "gid": data["gid_number"],
                "parentFileset": data["parentFileset"],
                "filesetType": "dependent",
                "volDirBasePath": data["volDirBasePath"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_255():
    value_sc = {"clusterId": data["id"], "volBackendFs": data["primaryFs"],
                "uid": data["uid_number"], "gid": data["gid_number"],
                "inodeLimit": data["inodeLimit"],
                "filesetType": "dependent",
                "volDirBasePath": data["volDirBasePath"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_256():
    value_sc = {"clusterId": data["id"], "volBackendFs": data["primaryFs"],
                "uid": data["uid_number"], "gid": data["gid_number"],
                "inodeLimit": data["inodeLimit"],
                "parentFileset": data["parentFileset"],
                "volDirBasePath": data["volDirBasePath"],
                "reason": "parentFileset and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_257():
    value_sc = {"clusterId": data["id"], "volBackendFs": data["primaryFs"],
                "uid": data["uid_number"], "inodeLimit": data["inodeLimit"],
                "parentFileset": data["parentFileset"],
                "filesetType": "dependent",
                "volDirBasePath": data["volDirBasePath"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_258():
    value_sc = {"clusterId": data["id"], "volBackendFs": data["primaryFs"],
                "gid": data["gid_number"], "inodeLimit": data["inodeLimit"],
                "parentFileset": data["parentFileset"],
                "filesetType": "dependent",
                "volDirBasePath": data["volDirBasePath"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_259():
    value_sc = {"volBackendFs": data["primaryFs"], "uid": data["uid_number"],
                "gid": data["gid_number"], "inodeLimit": data["inodeLimit"],
                "parentFileset": data["parentFileset"],
                "filesetType": "dependent",
                "volDirBasePath": data["volDirBasePath"],
                "reason": "filesetType and volDirBasePath must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_260():
    value_sc = {"clusterId": data["id"], "uid": data["uid_number"],
                "gid": data["gid_number"], "inodeLimit": data["inodeLimit"],
                "parentFileset": data["parentFileset"],
                "filesetType": "dependent",
                "volDirBasePath": data["volDirBasePath"],
                "reason": "volBackendFs must be specified in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_261():
    value_sc = {"clusterId": data["id"], "volBackendFs": data["primaryFs"],
                "uid": data["uid_number"], "gid": data["gid_number"],
                "inodeLimit": data["inodeLimit"],
                "parentFileset": data["parentFileset"],
                "filesetType": "dependent",
                "reason": "inodeLimit and filesetType=dependent must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_262():
    value_sc = {"clusterId": data["id"], "filesetType": "dependent",
                "uid": data["uid_number"], "inodeLimit": data["inodeLimit"],
                "volBackendFs": data["primaryFs"],
                "reason": "inodeLimit and filesetType=dependent must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_263():
    value_sc = {"clusterId": data["id"], "filesetType": "dependent",
                "gid": data["gid_number"], "inodeLimit": data["inodeLimit"],
                "volBackendFs": data["primaryFs"],
                "reason": "inodeLimit and filesetType=dependent must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_264():
    value_sc = {"clusterId": data["id"], "filesetType": "independent",
                "inodeLimit": data["inodeLimit"],
                "volBackendFs": data["primaryFs"],
                "parentFileset": data["parentFileset"],
                "reason": "InvalidArgument desc = parentFileset and filesetType=independent"}
    driver_object.test_dynamic(value_sc)

# invalid input test cases


invaliddata = {
    "volDirBasePath": "/invalid",
    "primaryFs": "invalid",
    "parentFileset": "invalid",
    "inodeLimit": "1023",
    "gid_number": "9999",
    "uid_number": "9999",
    "gid_name": "invalid_name",
                "uid_name": "invalid_name"
}


def test_driver_dynamic_fail_invalid_input_265():
    value_sc = {"volBackendFs": data["primaryFs"],
                "volDirBasePath": invaliddata["volDirBasePath"],
                "reason": "directory base path /invalid not present in"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_invalid_input_266():
    value_sc = {"clusterId":  data["id"], "filesetType": "dependent",
                "volBackendFs": invaliddata["primaryFs"],
                "reason": "filesystem invalid in not known to primary cluster"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_invalid_input_267():
    value_sc = {"clusterId":  data["id"], "filesetType": "dependent",
                "volBackendFs": data["primaryFs"],
                "parentFileset": invaliddata["parentFileset"],
                "reason": "unable to create fileset"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_invalid_input_268():
    value_sc = {"clusterId":  data["id"], "filesetType": "dependent",
                "volBackendFs": data["primaryFs"],
                "inodeLimit": invaliddata["inodeLimit"],
                "reason": "inodeLimit and filesetType=dependent must not be specified together in storageClass"}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_invalid_input_269():
    value_sc = {"clusterId":  data["id"], "filesetType": "independent",
                "volBackendFs": data["primaryFs"],
                "inodeLimit": invaliddata["inodeLimit"],
                "reason": "inodeLimit specified in storageClass must be equal to or greater than 1024"}
    driver_object.test_dynamic(value_sc)


@pytest.mark.skip
def test_driver_dynamic_fail_invalid_input_270():
    value_sc = {"clusterId":  data["id"], "volBackendFs": data["primaryFs"],
                "filesetType": "dependent", "gid": invaliddata["gid_number"],
                "uid": invaliddata["uid_number"],
                "reason": 'The object "9999" specified for "user" does not exist'}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_invalid_input_271():
    value_sc = {"clusterId":  data["id"], "volBackendFs": data["primaryFs"],
                "filesetType": "dependent", "gid": data["gid_number"],
                "uid": invaliddata["uid_name"],
                "reason": 'The object "invalid_name" specified for "user" does not exist'}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_invalid_input_272():
    value_sc = {"clusterId":  data["id"], "volBackendFs": data["primaryFs"],
                "filesetType": "dependent", "gid": invaliddata["gid_name"],
                "uid": invaliddata["uid_name"],
                "reason": 'The object "invalid_name" specified for "user" does not exist'}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_invalid_input_273():
    value_sc = {"clusterId":  data["id"], "volBackendFs": data["primaryFs"],
                "filesetType": "dependent", "gid": invaliddata["gid_name"],
                "uid": data["uid_number"],
                "reason": 'The object "invalid_name" specified for "group" does not exist'}
    driver_object.test_dynamic(value_sc)


def test_driver_dynamic_fail_invalid_input_274():
    value_sc = {"clusterId":  data["id"], "volBackendFs": data["primaryFs"],
                "gid": data["gid_number"], "uid": data["uid_number"],
                "parentFileset": invaliddata["parentFileset"],
                "reason": "parentFileset and filesetType=independent"}
    driver_object.test_dynamic(value_sc)


@pytest.mark.slow
def test_driver_one_pvc_two_pod_pass_1():
    value_pvc = {"access_modes": "ReadWriteMany", "storage": "1Gi"}
    value_ds = {"mount_path": "/usr/share/nginx/html/scale", "read_only": "False"}
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId":  data["id"]}
    driver_object.one_pvc_two_pod(value_sc, value_pvc, value_ds)


@pytest.mark.slow
def test_driver_one_pvc_two_pod_fail_1():
    value_pvc = {"access_modes": "ReadWriteOnce", "storage": "1Gi"}
    value_ds = {"mount_path": "/usr/share/nginx/html/scale", "read_only": "False",
                "reason": "Multi-Attach error for volume"}
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId":  data["id"]}
    driver_object.one_pvc_two_pod(value_sc, value_pvc, value_ds)


@pytest.mark.slow
def test_driver_one_pvc_two_pod_pass_2():
    value_pvc = {"access_modes": "ReadWriteMany", "storage": "1Gi"}
    value_ds = {"mount_path": "/usr/share/nginx/html/scale", "read_only": "False"}
    value_sc = {"volBackendFs": data["primaryFs"], "volDirBasePath": data["volDirBasePath"]}
    driver_object.one_pvc_two_pod(value_sc, value_pvc, value_ds)


@pytest.mark.slow
def test_driver_one_pvc_two_pod_fail_2():
    value_pvc = {"access_modes": "ReadWriteOnce", "storage": "1Gi"}
    value_ds = {"mount_path": "/usr/share/nginx/html/scale", "read_only": "False",
                "reason": "Multi-Attach error for volume"}
    value_sc = {"volBackendFs": data["primaryFs"], "volDirBasePath": data["volDirBasePath"]}
    driver_object.one_pvc_two_pod(value_sc, value_pvc, value_ds)


@pytest.mark.slow
def test_driver_one_pvc_two_pod_pass_3():
    value_pvc = {"access_modes": "ReadWriteMany", "storage": "1Gi"}
    value_ds = {"mount_path": "/usr/share/nginx/html/scale", "read_only": "False"}
    value_sc = {"volBackendFs": data["primaryFs"], "gid": data["gid_number"],
                "clusterId": data["id"], "filesetType": "dependent"}
    driver_object.one_pvc_two_pod(value_sc, value_pvc, value_ds)


@pytest.mark.slow
def test_driver_one_pvc_two_pod_fail_3():
    value_pvc = {"access_modes": "ReadWriteOnce", "storage": "1Gi"}
    value_ds = {"mount_path": "/usr/share/nginx/html/scale", "read_only": "False",
                "reason": "Multi-Attach error for volume"}
    value_sc = {"volBackendFs": data["primaryFs"], "gid": data["gid_number"],
                "clusterId": data["id"], "filesetType": "dependent"}
    driver_object.one_pvc_two_pod(value_sc, value_pvc, value_ds)


@pytest.mark.slow
def test_driver_one_pvc_two_pod_pass_4():
    value_pvc = {"access_modes": "ReadWriteMany", "storage": "1Gi"}
    value_ds = {"mount_path": "/usr/share/nginx/html/scale", "read_only": "False"}
    value_sc = {"volBackendFs": data["primaryFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "clusterId": data["id"],
                "filesetType": "dependent",
                "parentFileset": data["parentFileset"]}
    driver_object.one_pvc_two_pod(value_sc, value_pvc, value_ds)


@pytest.mark.slow
def test_driver_one_pvc_two_pod_fail_4():
    value_pvc = {"access_modes": "ReadWriteOnce", "storage": "1Gi"}
    value_ds = {"mount_path": "/usr/share/nginx/html/scale", "read_only": "False",
                "reason": "Multi-Attach error for volume"}
    value_sc = {"volBackendFs": data["primaryFs"], "gid": data["gid_number"],
                "uid": data["uid_number"], "clusterId": data["id"],
                "filesetType": "dependent",
                "parentFileset": data["parentFileset"]}
    driver_object.one_pvc_two_pod(value_sc, value_pvc, value_ds)


@pytest.mark.slow
@pytest.mark.stresstest
def test_driver_parallel_pvc_1():
    # this testcase contains sequential pod creation after all pvc are bound
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId":  data["id"], "inodeLimit": "1024"}
    driver_object.parallel_pvc(value_sc, data["number_of_parallel_pvc"], pod_creation=True)


@pytest.mark.slow
@pytest.mark.stresstest
def test_driver_parallel_pvc_2():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId":  data["id"], "inodeLimit": "1024"}
    driver_object.parallel_pvc(value_sc, data["number_of_parallel_pvc"], pod_creation=False)


@pytest.mark.slow
@pytest.mark.stresstest
def test_driver_parallel_pvc_3():
    # this testcase contains sequential pod creation after all pvc are bound
    value_sc = {"volBackendFs": data["primaryFs"],
                "filesetType": "dependent", "clusterId": data["id"]}
    driver_object.parallel_pvc(value_sc, data["number_of_parallel_pvc"], pod_creation=True)


@pytest.mark.slow
@pytest.mark.stresstest
def test_driver_parallel_pvc_4():
    value_sc = {"volBackendFs": data["primaryFs"],
                "filesetType": "dependent", "clusterId": data["id"]}
    driver_object.parallel_pvc(value_sc, data["number_of_parallel_pvc"], pod_creation=False)


@pytest.mark.slow
@pytest.mark.stresstest
def test_driver_parallel_pvc_5():
    # this testcase contains sequential pod creation after all pvc are bound
    value_sc = {"volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"]}
    driver_object.parallel_pvc(value_sc, data["number_of_parallel_pvc"], pod_creation=True)


@pytest.mark.slow
@pytest.mark.stresstest
def test_driver_parallel_pvc_6():
    value_sc = {"volBackendFs": data["primaryFs"],
                "volDirBasePath": data["volDirBasePath"]}
    driver_object.parallel_pvc(value_sc, data["number_of_parallel_pvc"], pod_creation=False)


@pytest.mark.regression
def test_driver_sc_permissions_empty_independent_pass_1():
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False",
                 "sub_path": ["sub_path_mnt"], "volumemount_readonly":[False], "reason": "Permission denied"},
                 {"mount_path": "/usr/share/nginx/html/scale",
                  "sub_path": ["sub_path_mnt"], "volumemount_readonly":[False],
                  "read_only": "True", "reason": "Read-only file system"}
                 ]
    # to test default behavior i.e. directory should be created with 771 permissions
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"], "permissions": "",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    driver_object.test_dynamic(value_sc, value_pod_passed=value_pod)


@pytest.mark.regression
def test_driver_sc_permissions_777_independent_pass_2():
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False", "sub_path": ["sub_path_mnt"], "volumemount_readonly":[False]},
                 {"mount_path": "/usr/share/nginx/html/scale", "read_only": "False", "sub_path": ["sub_path_mnt"], "volumemount_readonly":[True],
                  "reason": "Read-only file system"},
                 {"mount_path": "/usr/share/nginx/html/scale", "read_only": "True",
                  "sub_path": ["sub_path_mnt"], "volumemount_readonly":[False], "reason": "Read-only file system"}
                 ]
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"], "permissions": "777",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    driver_object.test_dynamic(value_sc, value_pod_passed=value_pod)


def test_driver_sc_permissions_666_independent_pass_3():
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False",
                  "sub_path": ["sub_path_mnt"], "volumemount_readonly":[False], "reason": "Permission denied"},
                 {"mount_path": "/usr/share/nginx/html/scale",
                  "sub_path": ["sub_path_mnt"], "volumemount_readonly":[False],
                  "read_only": "True", "reason": "Read-only file system"}
                 ]
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"], "permissions": "666",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    driver_object.test_dynamic(value_sc, value_pod_passed=value_pod)


def test_driver_sc_permissions_777_independent_pass_4():
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False",
                  "sub_path": ["sub_path_mnt", "sub_path_mnt_2", "sub_path_mnt3"], "volumemount_readonly":[False, True, True]},
                 {"mount_path": "/usr/share/nginx/html/scale", "read_only": "False",
                  "sub_path": ["sub_path_mnt", "sub_path_mnt_2", "sub_path_mnt3"], "volumemount_readonly":[True, False, False], "reason": "Read-only file system"},
                 {"mount_path": "/usr/share/nginx/html/scale", "read_only": "True",
                  "sub_path": ["sub_path_mnt", "sub_path_mnt_2", "sub_path_mnt3"], "volumemount_readonly":[False, False, False], "reason": "Read-only file system"}
                 ]
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"], "permissions": "777",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    driver_object.test_dynamic(value_sc, value_pod_passed=value_pod)


def test_driver_sc_permissions_777_dependent_pass_1():
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False", "sub_path": ["sub_path_mnt"], "volumemount_readonly":[False]},
                 {"mount_path": "/usr/share/nginx/html/scale", "read_only": "False", "sub_path": ["sub_path_mnt"], "volumemount_readonly":[True],
                  "reason": "Read-only file system"},
                 {"mount_path": "/usr/share/nginx/html/scale", "read_only": "True",
                  "sub_path": ["sub_path_mnt"], "volumemount_readonly":[False], "reason": "Read-only file system"}
                 ]
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"], "filesetType": "dependent", "permissions": "777",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    driver_object.test_dynamic(value_sc, value_pod_passed=value_pod)


def test_driver_sc_permissions_666_dependent_pass_2():
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False",
                  "sub_path": ["sub_path_mnt"], "volumemount_readonly":[False], "reason": "Permission denied"},
                 {"mount_path": "/usr/share/nginx/html/scale",
                  "sub_path": ["sub_path_mnt"], "volumemount_readonly":[False],
                  "read_only": "True", "reason": "Read-only file system"}
                 ]
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"], "filesetType": "dependent", "permissions": "666",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    driver_object.test_dynamic(value_sc, value_pod_passed=value_pod)


def test_driver_sc_permissions_777_dependent_pass_3():
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False",
                  "sub_path": ["sub_path_mnt", "sub_path_mnt_2", "sub_path_mnt3"], "volumemount_readonly":[False, True, True]},
                 {"mount_path": "/usr/share/nginx/html/scale", "read_only": "False",
                  "sub_path": ["sub_path_mnt", "sub_path_mnt_2", "sub_path_mnt3"], "volumemount_readonly":[True, False, False], "reason": "Read-only file system"},
                 {"mount_path": "/usr/share/nginx/html/scale", "read_only": "True",
                  "sub_path": ["sub_path_mnt", "sub_path_mnt_2", "sub_path_mnt3"], "volumemount_readonly":[False, False, False], "reason": "Read-only file system"}
                 ]
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"], "filesetType": "dependent", "permissions": "777",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    driver_object.test_dynamic(value_sc, value_pod_passed=value_pod)


def test_driver_sc_permissions_invalid_input_1():
    value_sc = {"clusterId":  data["id"], "volBackendFs": data["primaryFs"],
                "permissions": "   ",
                "reason": 'invalid value specified for permissions'}
    driver_object.test_dynamic(value_sc)


def test_driver_sc_permissions_invalid_input_2():
    value_sc = {"clusterId":  data["id"], "volBackendFs": data["primaryFs"],
                "permissions": "abcd",
                "reason": 'invalid value specified for permissions'}
    driver_object.test_dynamic(value_sc)


def test_driver_sc_permissions_invalid_input_3():
    value_sc = {"clusterId":  data["id"], "volBackendFs": data["primaryFs"],
                "permissions": "0x309",
                "reason": 'invalid value specified for permissions'}
    driver_object.test_dynamic(value_sc)


def test_driver_sc_permissions_invalid_input_4():
    value_sc = {"clusterId":  data["id"], "volBackendFs": data["primaryFs"],
                "permissions": "o777",
                "reason": 'invalid value specified for permissions'}
    driver_object.test_dynamic(value_sc)


def test_driver_sc_permissions_invalid_input_5():
    value_sc = {"clusterId":  data["id"], "volBackendFs": data["primaryFs"],
                "permissions": "77",
                "reason": 'invalid value specified for permissions'}
    driver_object.test_dynamic(value_sc)


def test_driver_sc_permissions_invalid_input_6():
    value_sc = {"clusterId":  data["id"], "volBackendFs": data["primaryFs"],
                "permissions": "778",
                "reason": 'invalid value specified for permissions'}
    driver_object.test_dynamic(value_sc)


def test_driver_volume_expansion_1():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"], "allow_volume_expansion": True}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi", "volume_expansion_storage": ["5Gi", "15Gi", "25Gi"]},
                 {"access_modes": "ReadWriteOnce", "storage": "1Gi", "volume_expansion_storage": ["4Gi", "12Gi"]},
                 {"access_modes": "ReadOnlyMany", "storage": "1Gi",
                  "reason": "ReadOnlyMany is not supported"}
                 ]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_volume_expansion_2():
    value_sc = {"volBackendFs": data["primaryFs"], "volDirBasePath": data["volDirBasePath"], "allow_volume_expansion": True}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi", "volume_expansion_storage": ["5Gi", "15Gi", "25Gi"]},
                 {"access_modes": "ReadWriteOnce", "storage": "1Gi", "volume_expansion_storage": ["4Gi", "12Gi"]},
                 {"access_modes": "ReadOnlyMany", "storage": "1Gi",
                  "reason": "ReadOnlyMany is not supported"}
                 ]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_volume_expansion_3():
    value_sc = {"volBackendFs": data["primaryFs"], "filesetType": "dependent",
                "clusterId": data["id"], "allow_volume_expansion": True}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi", "volume_expansion_storage": ["4Gi", "16Gi"]},
                 {"access_modes": "ReadWriteOnce", "storage": "1Gi", "volume_expansion_storage": ["5Gi", "15Gi", "25Gi"]},
                 {"access_modes": "ReadOnlyMany", "storage": "1Gi",
                  "reason": "ReadOnlyMany is not supported"}
                 ]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)


def test_driver_volume_cloning_pass_1():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}, {"access_modes": "ReadWriteOnce", "storage": "1Gi"}]}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_pass_2():
    value_sc = {"volBackendFs": data["primaryFs"], "filesetType": "dependent",
                "clusterId": data["id"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}, {"access_modes": "ReadWriteOnce", "storage": "1Gi"}]}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_pass_3():
    value_sc = {"volBackendFs": data["primaryFs"], "volDirBasePath": data["volDirBasePath"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}, {"access_modes": "ReadWriteOnce", "storage": "1Gi"}]}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_pass_4():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "11Gi"}, {"access_modes": "ReadWriteOnce", "storage": "8Gi"}]}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_multiple_clones():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}],
                          "number_of_clones": 5}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_chain_clones():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}],
                          "clone_chain": 1}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_expand_before():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"], "allow_volume_expansion": True}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "9Gi", "volume_expansion_storage": ["11Gi"]}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "11Gi"},
                                        {"access_modes": "ReadWriteMany", "storage": "9Gi", "reason": "new PVC request must be greater than or equal in size"}]}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_with_subpath():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"], "permissions": "777",
                "gid": data["gid_number"], "uid": data["uid_number"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi"}, {"access_modes": "ReadWriteOnce", "storage": "1Gi"}]}
    value_pod = [{"mount_path": "/usr/share/nginx/html/scale", "read_only": "False",
                  "sub_path": ["sub_path_mnt", "sub_path_mnt_2", "sub_path_mnt3"], "volumemount_readonly":[False, False, True]}]
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_pod_passed=value_pod, value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_fail_1():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi", "reason": "same storage class for cloning"}],
                          "clone_sc": {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_fail_2():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi", "reason": "same storage class for cloning"}],
                          "clone_sc": {"volBackendFs": data["primaryFs"], "filesetType": "dependent",
                                       "clusterId": data["id"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_fail_3():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi", "reason": "same storage class for cloning"}],
                          "clone_sc": {"volBackendFs": data["primaryFs"], "volDirBasePath": data["volDirBasePath"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_fail_4():
    value_sc = {"volBackendFs": data["primaryFs"], "filesetType": "dependent",
                "clusterId": data["id"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi", "reason": "same storage class for cloning"}],
                          "clone_sc": {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_fail_5():
    value_sc = {"volBackendFs": data["primaryFs"], "filesetType": "dependent",
                "clusterId": data["id"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc":  [{"access_modes": "ReadWriteMany", "storage": "1Gi", "reason": "same storage class for cloning"}],
                          "clone_sc": {"volBackendFs": data["primaryFs"], "filesetType": "dependent",
                                       "clusterId": data["id"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_fail_6():
    value_sc = {"volBackendFs": data["primaryFs"], "filesetType": "dependent",
                "clusterId": data["id"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi", "reason": "same storage class for cloning"}],
                          "clone_sc": {"volBackendFs": data["primaryFs"], "volDirBasePath": data["volDirBasePath"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_fail_7():
    value_sc = {"volBackendFs": data["primaryFs"], "volDirBasePath": data["volDirBasePath"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi", "reason": "same storage class for cloning"}],
                          "clone_sc": {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_fail_8():
    value_sc = {"volBackendFs": data["primaryFs"], "volDirBasePath": data["volDirBasePath"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi", "reason": "same storage class for cloning"}],
                          "clone_sc": {"volBackendFs": data["primaryFs"], "filesetType": "dependent",
                                       "clusterId": data["id"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_fail_9():
    value_sc = {"volBackendFs": data["primaryFs"], "volDirBasePath": data["volDirBasePath"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi", "reason": "same storage class for cloning"}],
                          "clone_sc": {"volBackendFs": data["primaryFs"], "volDirBasePath": data["volDirBasePath"]}}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_clone_passed=value_clone_passed)


def test_driver_volume_cloning_fail_10():
    value_sc = {"volBackendFs": data["primaryFs"], "clusterId": data["id"]}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "5Gi"}]
    value_clone_passed = {"clone_pvc": [{"access_modes": "ReadWriteMany", "storage": "1Gi", "reason": "new PVC request must be greater than or equal in size"}]}
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc, value_clone_passed=value_clone_passed)


@pytest.mark.regression
def test_driver_cg_pass_1():
    value_sc = {"volBackendFs": data["primaryFs"], "type": "advanced"}
    value_pvc = [{"access_modes": "ReadWriteMany", "storage": "1Gi"}] * 10
    driver_object.test_dynamic(value_sc, value_pvc_passed=value_pvc)
