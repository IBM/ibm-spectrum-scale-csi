import copy
import logging
import os.path
import yaml
import string
import random
from kubernetes import config
import ibm_spectrum_scale_csi.kubernetes_apis.csi_object_function as csiobjectfunc
import ibm_spectrum_scale_csi.spectrum_scale_apis.fileset_functions as filesetfunc
LOGGER = logging.getLogger()


def get_test_data(test_config):
    filepath = test_config
    try:
        with open(filepath, "r") as f:
            data = yaml.full_load(f.read())
    except yaml.YAMLError as exc:
        print(f"Error in configuration file {filepath} :", exc)
        assert False

    if data['keepobjects'] == "True" or data['keepobjects'] == "true":
        data['keepobjects'] = True
    else:
        data['keepobjects'] = False

    if data.get('remote_username') is None:
        data['remote_username'] = {}
    if data.get('remote_password') is None:
        data['remote_password'] = {}
    if data.get('remote_cacert_path') is None:
        data['remote_cacert_path'] = {}

    return data


def read_driver_data(cmd_values):

    data = get_test_data(cmd_values["test_config"])

    data["namespace"] = cmd_values["test_namespace"]
    
    config.load_kube_config(config_file=cmd_values["kubeconfig_value"])
    loadcr_yaml = csiobjectfunc.get_scaleoperatorobject_values(cmd_values["operator_namespace"], data["csiscaleoperator_name"])

    if loadcr_yaml is False:
        try:
            with open(cmd_values["clusterconfig_value"], "r") as f:
                loadcr_yaml = yaml.full_load(f.read())
        except yaml.YAMLError as exc:
            LOGGER.error(f'Error in parsing the cr file {cmd_values["clusterconfig_value"]} : {exc}')
            assert False

    else:
        auto_fetch_gui_creds_and_remote_filesystem(loadcr_yaml, data)


    for cluster in loadcr_yaml["spec"]["clusters"]:
        if "primary" in cluster and "primaryFs" in cluster["primary"] and cluster["primary"]["primaryFs"] is not '':
            data["primaryFs"] = cluster["primary"]["primaryFs"]
            data["guiHost"] = cluster["restApi"][0]["guiHost"]
            if "primaryFset" in cluster:
                data["primaryFset"] = cluster["primary"]["primaryFset"]
            else:
                data["primaryFset"] = "spectrum-scale-csi-volume-store"
            data["id"] = cluster["id"]

    data["clusters"] = loadcr_yaml["spec"]["clusters"]
    if len(loadcr_yaml["spec"]["clusters"]) > 1:
        data["remote"] = True

    if "pluginNodeSelector" in loadcr_yaml["spec"]:
        data["pluginNodeSelector"] = loadcr_yaml["spec"]["pluginNodeSelector"]
    else:
        data["pluginNodeSelector"] = []

    return data


def read_operator_data(clusterconfig, namespace, testconfig, kubeconfig=None):

    data = get_test_data(testconfig)

    data["namespace"] = namespace

    if kubeconfig is not None:
        config.load_kube_config(config_file=kubeconfig)
        loadcr_yaml = csiobjectfunc.get_scaleoperatorobject_values(namespace, data["csiscaleoperator_name"])
    else:
        loadcr_yaml = False

    if loadcr_yaml is False:
        try:
            with open(clusterconfig, "r") as f:
                loadcr_yaml = yaml.full_load(f.read())
        except yaml.YAMLError as exc:
            LOGGER.error(f"Error in parsing the cr file {clusterconfig} : {exc}")
            assert False
    else:
        auto_fetch_gui_creds_and_remote_filesystem(loadcr_yaml, data)


    data["custom_object_body"] = copy.deepcopy(loadcr_yaml)
    data["custom_object_body"]["metadata"]["namespace"] = namespace
    data["remote_secret_names"] = []
    data["remote_cacert_names"] = []
    for cluster in loadcr_yaml["spec"]["clusters"]:
        if "primary" in cluster and "primaryFs" in cluster["primary"] and cluster["primary"]["primaryFs"] is not '':
            data["primaryFs"] = cluster["primary"]["primaryFs"]
            data["guiHost"] = cluster["restApi"][0]["guiHost"]
            data["local_secret_name"] = cluster["secrets"]
            if "primaryFset" in cluster["primary"]:
                data["primaryFset"] = cluster["primary"]["primaryFset"]
            else:
                data["primaryFset"] = "spectrum-scale-csi-volume-store"
            if "cacert" in cluster:
                data["local_cacert_name"] = cluster["cacert"]
        else:
            data["remote_secret_names"].append(cluster["secrets"])
            if "cacert" in cluster:
                data["remote_cacert_names"].append(cluster["cacert"])

    if "attacherNodeSelector" in loadcr_yaml["spec"]:
        data["attacherNodeSelector"] = loadcr_yaml["spec"]["attacherNodeSelector"]
    else:
        data["attacherNodeSelector"] = []

    if "provisionerNodeSelector" in loadcr_yaml["spec"]:
        data["provisionerNodeSelector"] = loadcr_yaml["spec"]["provisionerNodeSelector"]
    else:
        data["provisionerNodeSelector"] = []

    if "pluginNodeSelector" in loadcr_yaml["spec"]:
        data["pluginNodeSelector"] = loadcr_yaml["spec"]["pluginNodeSelector"]
    else:
        data["pluginNodeSelector"] = []

    if "resizerNodeSelector" in loadcr_yaml["spec"]:
        data["resizerNodeSelector"] = loadcr_yaml["spec"]["resizerNodeSelector"]
    else:
        data["resizerNodeSelector"] = []

    if "snapshotterNodeSelector" in loadcr_yaml["spec"]:
        data["snapshotterNodeSelector"] = loadcr_yaml["spec"]["snapshotterNodeSelector"]
    else:
        data["snapshotterNodeSelector"] = []

    if "local_cacert_name" in data:
        if data["cacert_path"] == "":
            LOGGER.error("if using cacert , MUST include cacert path in test.config")
            assert False

    for remote_secret_name in data["remote_secret_names"]:
        if not(remote_secret_name in data["remote_username"].keys()):
            LOGGER.error(f"Need username for {remote_secret_name} secret in test.config")
            assert False
        if not(remote_secret_name in data["remote_password"].keys()):
            LOGGER.error(f"Need password for {remote_secret_name} secret in test.config")
            assert False

    for remote_cacert_name in data["remote_cacert_names"]:
        if not(remote_cacert_name in data["remote_cacert_path"].keys()):
            LOGGER.error(f"Need cacert path for {remote_cacert_name} in test.config")
            assert False

    return data


def get_remote_data(data_passed):
    remote_data = copy.deepcopy(data_passed)
    remote_data["remoteFs_remote_name"] = filesetfunc.get_remoteFs_remotename(copy.deepcopy(remote_data))
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


def get_pytest_cmd_values(request):
    """
    Get pytest commmand line parameters and convert them to dict
    """

    cmd_value_dict = {}

    kubeconfig_value = request.config.option.kubeconfig
    if kubeconfig_value is None:
        if os.path.isfile('config/kubeconfig'):
            kubeconfig_value = 'config/kubeconfig'
        else:
            kubeconfig_value = '~/.kube/config'

    clusterconfig_value = request.config.option.clusterconfig
    if clusterconfig_value is None:
        if os.path.isfile('config/csiscaleoperators.csi.ibm.com_cr.yaml'):
            clusterconfig_value = 'config/csiscaleoperators.csi.ibm.com_cr.yaml'
        else:
            clusterconfig_value = '../../operator/config/samples/csiscaleoperators.csi.ibm.com_cr.yaml'

    test_namespace = request.config.option.testnamespace
    if test_namespace is None:
        test_namespace = 'ibm-spectrum-scale-csi-driver'

    operator_namespace = request.config.option.operatornamespace
    if operator_namespace is None:
        operator_namespace = 'ibm-spectrum-scale-csi-driver'

    runslow_val = request.config.option.runslow

    operator_file = request.config.option.operatoryaml
    if operator_file is None:
        operator_file = '../../generated/installer/ibm-spectrum-scale-csi-operator-dev.yaml'

    test_config = request.config.option.testconfig
    if test_config is None:
        test_config = "config/test.config"

    cmd_value_dict = {"kubeconfig_value": kubeconfig_value,
                      "clusterconfig_value":clusterconfig_value, 
                      "test_namespace": test_namespace,
                      "operator_namespace":operator_namespace,
                      "runslow_val":runslow_val,
                      "operator_file":operator_file, 
                      "test_config":test_config
                     }

    return cmd_value_dict


def randomStringDigits(stringLength=6):
    """Generate a random string of letters and digits """
    lettersAndDigits = string.ascii_letters + string.digits
    return ''.join(random.choice(lettersAndDigits) for i in range(stringLength))


def randomString(stringLength=10):
    """Generate a random string of fixed length """
    letters = string.ascii_lowercase
    return ''.join(random.choice(letters) for i in range(stringLength))


def auto_fetch_gui_creds_and_remote_filesystem(loadcr_yaml, data):
    for cluster in loadcr_yaml["spec"]["clusters"]:
        if "primary" in cluster and "primaryFs" in cluster["primary"] and cluster["primary"]["primaryFs"] is not '':
            local_secret_name=cluster["secrets"]
            data["username"],data["password"]= \
                csiobjectfunc.get_gui_creds_for_username_password(data["namespace"],local_secret_name)
            if "remoteCluster" in cluster["primary"] and cluster["primary"]["remoteCluster"] is not '':
                data["remoteFs"] = cluster["primary"]["primaryFs"]
                data["remoteid"] = cluster["primary"]["remoteCluster"]
        else:
            remote_secret_name= cluster["secrets"]
            data["remote_username"][remote_secret_name],data["remote_password"][remote_secret_name]= \
                csiobjectfunc.get_gui_creds_for_username_password(data["namespace"],remote_secret_name)
