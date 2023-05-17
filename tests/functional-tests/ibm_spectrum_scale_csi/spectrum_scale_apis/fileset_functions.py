import logging
import time
import re
import json
import requests
import urllib3
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
LOGGER = logging.getLogger()


def set_data(data):
    global test
    test = data


def set_scalevalidation(temp_scalevalidation):
    global scalevalidation
    scalevalidation = temp_scalevalidation


def delete_fileset(test_data):
    """
    Deletes the primaryFset provided in configuration file
    if primaryFset not deleted , asserts

    Args:
        param1: test_data : contents of configuration file

    Returns:
       None

    Raises:
       None

    """
    unlink_fileset(test_data)
    delete_link = f'https://{test_data["guiHost"]}:{test_data["port"]}/scalemgmt/v2/filesystems/{test_data["primaryFs"]}/filesets/{test_data["primaryFset"]}'
    response = requests.delete(
        delete_link, verify=False, auth=(test_data["username"], test_data["password"]))
    LOGGER.debug(response.text)
    time.sleep(10)
    get_link = f'https://{test_data["guiHost"]}:{test_data["port"]}/scalemgmt/v2/filesystems/{test_data["primaryFs"]}/filesets/'
    response = requests.get(get_link, verify=False, auth=(
        test_data["username"], test_data["password"]))
    LOGGER.debug(response.text)
    search_format = f'"{test_data["primaryFset"]}",'

    search_result = re.search(search_format, str(response.text))
    LOGGER.debug(search_result)
    if search_result is None:
        LOGGER.info(
            f'Success : Fileset {test_data["primaryFset"]} has been deleted')
    else:
        LOGGER.error(
            f'Failed : Fileset {test_data["primaryFset"]} has not been deleted')
        LOGGER.error(get_link)
        LOGGER.error(response.text)
        assert False


def unlink_fileset(test_data):
    """
    unlink primaryFset provided in configuration file

    Args:
        param1: test_data : contents of configuration file

    Returns:
       None

    Raises:
       None

    """
    time.sleep(10)
    unlink_link = f'https://{test_data["guiHost"]}:{test_data["port"]}/scalemgmt/v2/filesystems/{test_data["primaryFs"]}/filesets/{test_data["primaryFset"]}/link'
    response = requests.delete(
        unlink_link, verify=False, auth=(test_data["username"], test_data["password"]))
    LOGGER.debug(response.text)
    LOGGER.info(f'Fileset {test_data["primaryFset"]} unlinked')
    time.sleep(10)


def fileset_exists(test_data):
    """

    Checks primaryFset provided in configuration file exists or not

    Args:
        param1: test_data : contents of configuration file

    Returns:
       returns True , if primaryFset exists
       returns False , if primaryFset does not exists

    Raises:
       None

    """
    get_link = f'https://{test_data["guiHost"]}:{test_data["port"]}/scalemgmt/v2/filesystems/{test_data["primaryFs"]}/filesets/'
    time.sleep(10)
    response = requests.get(get_link, verify=False, auth=(
        test_data["username"], test_data["password"]))
    LOGGER.debug(response.text)
    search_format = f'"{test_data["primaryFset"]}",'

    search_result = re.search(search_format, str(response.text))
    if search_result is None:
        return False
    return True


def cred_check(test_data):
    """
    checks if given parameters in test_data are correct
    by calling API using that data

    if API gives any error , It asserts
    """
    if scalevalidation == "False":
        LOGGER.warning(f'cred check : skipped for guihost {test_data["guiHost"]} as scalevalidation = "False" in config file')
        return
    get_link = f'https://{test_data["guiHost"]}:{test_data["port"]}/scalemgmt/v2/cluster'
    try:
        response = requests.get(get_link, verify=False, auth=(
            test_data["username"], test_data["password"]))
        LOGGER.debug(response.text)
    except:
        LOGGER.error(
            f'not able to use Scale REST API , Check cr file (guiHost = {test_data["guiHost"]})')
        assert False

    if not(response.status_code == 200):
        LOGGER.error("API response is ")
        LOGGER.error(str(response))
        LOGGER.error("not able to use scale REST API")
        LOGGER.error("Recheck parameters of config/test.config file")
        LOGGER.error(f'must check username and password for {test_data["guiHost"]}')
        assert False


def link_fileset(test_data):
    """
    link primaryFset provided in configuration file

    Args:
        param1: test_data : contents of configuration file

    Returns:
       None

    Raises:
       None

    """
    fileset_link = f'https://{test_data["guiHost"]}:{test_data["port"]}/scalemgmt/v2/filesystems/{test_data["primaryFs"]}/filesets/{test_data["primaryFset"]}/link'
    response = requests.post(fileset_link, verify=False,
                             auth=(test_data["username"], test_data["password"]))
    LOGGER.debug(response.text)
    LOGGER.info(f'Fileset {test_data["primaryFset"]} linked')
    time.sleep(5)


def create_fileset(test_data):
    """
    create primaryFset provided in configuration file

    Args:
        param1: test_data : contents of configuration file

    Returns:
       None

    Raises:
       None

    """
    get_link = f'https://{test_data["guiHost"]}:{test_data["port"]}/scalemgmt/v2/filesystems/{test_data["primaryFs"]}/'
    time.sleep(2)
    response = requests.get(get_link, verify=False, auth=(
        test_data["username"], test_data["password"]))
    LOGGER.debug(response.text)
    response_dict = json.loads(response.text)
    scalehostpath = response_dict["filesystems"][0]["mount"]["mountPoint"]
    create_fileset_link = f'https://{test_data["guiHost"]}:{test_data["port"]}/scalemgmt/v2/filesystems/{test_data["primaryFs"]}/filesets'
    headers = {
        'content-type': 'application/json',
        'accept': 'application/json',
    }
    data = f'{{"filesetName":"{test_data["primaryFset"]}","path":"{scalehostpath}/{test_data["primaryFset"]}"}}'

    response = requests.post(create_fileset_link, headers=headers,
                             data=data, verify=False, auth=(test_data["username"], test_data["password"]))
    LOGGER.debug(response.text)
    LOGGER.info(f'Fileset {test_data["primaryFset"]} created and linked')
    time.sleep(5)


def unmount_fs(test_data):
    """
    unmount primaryFs provided in configuration file

    Args:
        param1: test_data : contents of configuration file

    Returns:
       None

    Raises:
       None

    """
    unmount_link = f'https://{test_data["guiHost"]}:{test_data["port"]}/scalemgmt/v2/filesystems/{test_data["primaryFs"]}/unmount'
    LOGGER.debug(unmount_link)
    headers = {
        'content-type': 'application/json',
        'accept': 'application/json',
    }
    data = f'{{"nodes":["{test_data["guiHost"]}"],"force": false}}'
    response = requests.put(unmount_link, headers=headers,
                            data=data, verify=False, auth=(test_data["username"], test_data["password"]))
    LOGGER.debug(response.text)
    time.sleep(5)

    get_link = f'https://{test_data["guiHost"]}:{test_data["port"]}/scalemgmt/v2/filesystems/{test_data["primaryFs"]}'
    response = requests.get(get_link, verify=False, auth=(
        test_data["username"], test_data["password"]))
    response_dict = json.loads(response.text)
    mounted_on = response_dict["filesystems"][0]["mount"]["nodesMountedReadWrite"]

    if test_data["guiHost"] in mounted_on:
        LOGGER.error(f'Unable to unmount {test_data["primaryFs"]} from {test_data["guiHost"]}')
        assert False

    LOGGER.info(f'primaryFS {test_data["primaryFs"]} unmounted from {test_data["guiHost"]}')


def mount_fs(test_data):
    """
    mount primaryFs provided in configuration file

    Args:
        param1: test_data : contents of configuration file

    Returns:
       None

    Raises:
       None

    """
    mount_link = f'https://{test_data["guiHost"]}:{test_data["port"]}/scalemgmt/v2/filesystems/{test_data["primaryFs"]}/mount'
    LOGGER.debug(mount_link)
    headers = {
        'content-type': 'application/json',
        'accept': 'application/json',
    }
    data = f'{{"nodes":["{test_data["guiHost"]}"]}}'
    response = requests.put(mount_link, headers=headers,
                            data=data, verify=False, auth=(test_data["username"], test_data["password"]))
    LOGGER.debug(response.text)
    time.sleep(5)

    get_link = f'https://{test_data["guiHost"]}:{test_data["port"]}/scalemgmt/v2/filesystems/{test_data["primaryFs"]}'
    response = requests.get(get_link, verify=False, auth=(
        test_data["username"], test_data["password"]))
    response_dict = json.loads(response.text)
    mounted_on = response_dict["filesystems"][0]["mount"]["nodesMountedReadWrite"]

    LOGGER.info(f'primaryFS {test_data["primaryFs"]} is mounted on {mounted_on}')


def cleanup():
    """
    deletes all primaryFsets that matches pattern pvc-
    used for cleanup in case of parallel pvc.

    Args:
       None

    Returns:
       None

    Raises:
       None

    """
    get_link = f'https://{test["guiHost"]}:{test["port"]}/scalemgmt/v2/filesystems/{test["primaryFs"]}/filesets/'
    response = requests.get(get_link, verify=False, auth=(test["username"], test["password"]))
    lst = re.findall(r'\S+pvc-\S+', response.text)
    lst2 = []
    for la in lst:
        lst2.append(la[1:-2])
    for res in lst2:
        volume_name = res
        unlink_link = f'https://{test["guiHost"]}:{test["port"]}/scalemgmt/v2/filesystems/{test["primaryFs"]}/filesets/{volume_name}/link'
        response = requests.delete(
            unlink_link, verify=False, auth=(test["username"], test["password"]))
        LOGGER.debug(response.text)
        LOGGER.info(f"Fileset {volume_name} unlinked")
        time.sleep(5)
        delete_link = f'https://{test["guiHost"]}:{test["port"]}/scalemgmt/v2/filesystems/{test["primaryFs"]}/filesets/{volume_name}/link'
        response = requests.delete(
            delete_link, verify=False, auth=(test["username"], test["password"]))
        LOGGER.debug(response.text)
        time.sleep(10)
        get_link = f'https://{test["guiHost"]}:{test["port"]}/scalemgmt/v2/filesystems/{test["primaryFs"]}/filesets/'
        response = requests.get(get_link, verify=False,
                                auth=(test["username"], test["password"]))
        LOGGER.debug(response.text)
        search_result = re.search(volume_name, str(response.text))
        if search_result is None:
            LOGGER.info(f'Fileset {volume_name} deleted successfully')
        else:
            LOGGER.error(f'Fileset {volume_name} is not deleted')
            assert False


def delete_created_fileset(volume_name):
    """
    deletes primaryFset created by pvc

    Args:
        param1: volume_name : fileset to be deleted

    Returns:
       None

    Raises:
       None

    """
    if volume_name is None:
        return
    if scalevalidation == "False":
        LOGGER.warning(f'Delete created fileset : skipped for fileset {volume_name} as scalevalidation = "False" in config file')
        return
    get_link = f'https://{test["guiHost"]}:{test["port"]}/scalemgmt/v2/filesystems/{test["primaryFs"]}/filesets/'
    response = requests.get(get_link, verify=False, auth=(test["username"], test["password"]))
    LOGGER.debug(response.text)
    search_result = re.search(volume_name, str(response.text))
    LOGGER.debug(search_result)
    if search_result is None:
        LOGGER.info(f'Fileset Delete : Fileset {volume_name} has already been deleted')
    else:
        unlink_link = f'https://{test["guiHost"]}:{test["port"]}/scalemgmt/v2/filesystems/{test["primaryFs"]}/filesets/{volume_name}/link'
        response = requests.delete(
            unlink_link, verify=False, auth=(test["username"], test["password"]))
        LOGGER.debug(response.text)
        LOGGER.info(f'Fileset {volume_name} has been unlinked')
        time.sleep(5)
        delete_link = f'https://{test["guiHost"]}:{test["port"]}/scalemgmt/v2/filesystems/{test["primaryFs"]}/filesets/{volume_name}'
        response = requests.delete(
            delete_link, verify=False, auth=(test["username"], test["password"]))
        LOGGER.debug(response.text)

        for _ in range(0, 12):
            get_link = f'https://{test["guiHost"]}:{test["port"]}/scalemgmt/v2/filesystems/{test["primaryFs"]}/filesets/'
            response = requests.get(get_link, verify=False,
                                    auth=(test["username"], test["password"]))
            LOGGER.debug(response.text)
            search_result = re.search(volume_name, str(response.text))
            LOGGER.debug(search_result)
            if search_result is None:
                LOGGER.info(f'Fileset Delete : Fileset {volume_name} has been deleted successfully')
                return
            time.sleep(15)
            LOGGER.info(f'Fileset Check : Checking for Fileset {volume_name}')
        LOGGER.error(f'Fileset Delete : Fileset {volume_name} deletion operation failed')
        assert False


def check_fileset_deleted(volume_name):
    """
    Checks fileset volume_name deleted or not
    """
    if scalevalidation == "False":
        LOGGER.warning(f'Check fileset deleted : skipped for fileset {volume_name} as scalevalidation = "False" in config file')
        return True
    get_link = f'https://{test["guiHost"]}:{test["port"]}/scalemgmt/v2/filesystems/{test["primaryFs"]}/filesets/{volume_name}'
    time.sleep(10)
    response = requests.get(get_link, verify=False, auth=(test["username"], test["password"]))
    LOGGER.debug(response.text)
    search_format = f'"{volume_name}",'
    search_result = re.search(search_format, str(response.text))
    if search_result is None:
        return True
    return False

def created_fileset_exists(volume_name):
    """
    Checks fileset volume_name exists or not

    Args:
        param1: volume_name : fileset to be checked

    Returns:
       returns True , if volume_name exists
       returns False , if volume_name does not exists

    Raises:
       None

    """
    if scalevalidation == "False":
        LOGGER.warning(f'Check fileset exists : skipped for fileset {volume_name} as scalevalidation = "False" in config file')
        return True
    get_link = f'https://{test["guiHost"]}:{test["port"]}/scalemgmt/v2/filesystems/{test["primaryFs"]}/filesets/{volume_name}'
    time.sleep(10)
    response = requests.get(get_link, verify=False, auth=(test["username"], test["password"]))
    LOGGER.debug(response.text)
    search_format = f'"{volume_name}",'
    search_result = re.search(search_format, str(response.text))
    if search_result is None:
        return False
    return True


def create_dir(dir_name):
    """
    creates directory named dir_name

    Args:
       param1: dir_name : name of directory to be created

    Returns:
       None

    Raises:
       None

    """
    if scalevalidation == "False":
        LOGGER.warning(f'Create Directory : skipped for directory {dir_name} as scalevalidation = "False" in config file')
        return

    if check_dir(dir_name) is True:
        return

    dir_link = f'https://{test["guiHost"]}:{test["port"]}/scalemgmt/v2/filesystems/{test["primaryFs"]}/directory/{dir_name}'
    LOGGER.debug(dir_link)
    headers = {
        'content-type': 'application/json',
        'accept': 'application/json',
    }
    data = f'{{ "user":"{test["uid_name"]}", "uid":{test["uid_number"]}, "group":"{test["gid_name"]}", "gid":{test["gid_number"]} }}'
    response = requests.post(dir_link, headers=headers,
                             data=data, verify=False, auth=(test["username"], test["password"]))
    LOGGER.debug(response.text)
    LOGGER.info(f'Directory Create : Creating directory {dir_name}')
    if check_dir(dir_name) is True:
        LOGGER.info(f'Directory Check : directory {dir_name} created successfully')
        return
    LOGGER.error(f'directory {dir_name} not created successfully')
    LOGGER.error(str(response))
    LOGGER.error(str(response.text))
    assert False


def check_dir(dir_name):
    """
    checks directory dir_name is present or not
    asserts  if not present
    """
    if scalevalidation == "False":
        LOGGER.warning(f'Check Directory : skipped for directory {dir_name} as scalevalidation = "False" in config file')
        return
    val = 0
    while val < 12:
        check_dir_link = f'https://{test["guiHost"]}:{test["port"]}/scalemgmt/v2/filesystems/{test["primaryFs"]}/owner/{dir_name}'
        LOGGER.debug(check_dir_link)
        headers = {
            'accept': 'application/json',
        }
        response = requests.get(check_dir_link, headers=headers,
                                verify=False, auth=(test["username"], test["password"]))
        LOGGER.debug(response.text)
        if response.status_code == 200:
            return True
        time.sleep(15)
        LOGGER.info(f'Directory Check : Checking for directory {dir_name}')
        val += 1
    return False


def delete_dir(dir_name):
    """
    deleted directory named dir_name

    Args:
       param1: dir_name : name of directory to be deleted

    Returns:
       None

    Raises:
       None

    """
    if scalevalidation == "False":
        LOGGER.warning(f'Delete Directory : skipped for directory {dir_name} as scalevalidation = "False" in config file')
        return
    dir_link = f'https://{test["guiHost"]}:{test["port"]}/scalemgmt/v2/filesystems/{test["primaryFs"]}/directory/{dir_name}'
    LOGGER.debug(dir_link)
    headers = {
        'content-type': 'application/json',
    }
    response = requests.delete(
        dir_link, headers=headers, verify=False, auth=(test["username"], test["password"]))
    LOGGER.debug(response.text)
    LOGGER.info(f'Deleted directory {dir_name}')
    time.sleep(5)


def get_FSUID():
    """
    return the UID of primaryFs

    Args:
       None

    Returns:
       FSUID

    Raises:
       None

    """
    info_filesystem = f'https://{test["guiHost"]}:{test["port"]}/scalemgmt/v2/filesystems/{test["primaryFs"]}?fields=:all:'
    response = requests.get(info_filesystem, verify=False,
                            auth=(test["username"], test["password"]))
    LOGGER.debug(response.text)
    lst = re.findall(r'\S+uuid[\S.\s]+', response.text)
    FSUID = str(lst[0][10:27])
    LOGGER.debug(FSUID)
    return FSUID


def get_mount_point():
    """
    return the mount point of primaryFs

    Args:
       None

    Returns:
       mount point

    Raises:
       None

    """

    if "type_remote" in test:
        info_filesystem = f'https://{test["type_remote"]["guiHost"]}:{test["type_remote"]["port"]}/scalemgmt/v2/filesystems/{test["remoteFs"]}?fields=:all:'
        response = requests.get(info_filesystem, verify=False,
                                auth=(test["type_remote"]["username"], test["type_remote"]["password"]))
    else:
        info_filesystem = f'https://{test["guiHost"]}:{test["port"]}/scalemgmt/v2/filesystems/{test["primaryFs"]}?fields=:all:'
        response = requests.get(info_filesystem, verify=False,
                                auth=(test["username"], test["password"]))
    LOGGER.debug(response.text)
    response_dict = json.loads(response.text)
    mount_point = response_dict["filesystems"][0]["mount"]["mountPoint"]
    LOGGER.debug(mount_point)
    return mount_point


def get_remoteFs_remotename_and_remoteid(test_data):
    """ return name of remote filesystem's remote name """
    if scalevalidation == "False":
        LOGGER.warning('auto remoteFs_remotename_and_remoteid fetch : skipped as scalevalidation = "False" in config file')
        LOGGER.warning(f'Using remoteclusterid {test_data["remoteclusterid"]} provided in test.config file')
        return test_data["remoteFs"],test_data["remoteclusterid"]

    info_filesystem = f'https://{test_data["guiHost"]}:{test_data["port"]}/scalemgmt/v2/filesystems/{test_data["remoteFs"]}'
    response = requests.get(info_filesystem, verify=False,
                            auth=(test_data["username"], test_data["password"]))
    LOGGER.debug(response.text)
    if not(response.status_code == 200):
        LOGGER.error(response.text)
        LOGGER.error(
            f'Not able to find filesystem {test_data["remoteFs"]} on GUI {test_data["guiHost"]}')
        assert False

    response_dict = json.loads(response.text)
    fs_remote_name, remoteid = None, None
    for filesystem in response_dict["filesystems"]:
        if filesystem["name"] == test_data["remoteFs"]:
            device_name = filesystem["mount"]["remoteDeviceName"]
            LOGGER.debug(device_name)
            temp_split = device_name.split(":")
            fs_remote_name = temp_split[1]
            remote_cluster_name = temp_split[0]

    if fs_remote_name is None:
        return fs_remote_name, remoteid

    for cluster in test_data["clusters"]:
        if test_data["guiHost"] != cluster["restApi"][0]["guiHost"]:
            remote_sec_name = cluster["secrets"]
            remote_gui_host = cluster["restApi"][0]["guiHost"]
            remote_username = test_data['remote_username'][remote_sec_name]
            remote_password = test_data['remote_password'][remote_sec_name]
            get_link = f'https://{remote_gui_host}:{test_data["remote_port"]}/scalemgmt/v2/cluster'
            response = requests.get(get_link, verify=False, auth=(remote_username, remote_password))
            if not(response.status_code == 200):
                LOGGER.error(response.text)
                LOGGER.error(get_link)
                assert False
            response_dict = json.loads(response.text)
            this_cluster_name = response_dict["cluster"]["clusterSummary"]["clusterName"]
            if this_cluster_name == remote_cluster_name:
                remoteid = str(response_dict["cluster"]["clusterSummary"]["clusterId"])
                return fs_remote_name, remoteid
    return fs_remote_name, remoteid


def check_snapshot_exists(snapshot_name, volume_name):
    """
    checks if snapshot is snapshot_name created for volume_name

    if created returns True
    else return False
    """
    if scalevalidation == "False":
        LOGGER.warning(f'Check snapshot exists : skipped for snapshot {snapshot_name} as scalevalidation = "False" in config file')
        return True
    val = 0
    while val < 12:
        snap_link = f'https://{test["guiHost"]}:{test["port"]}/scalemgmt/v2/filesystems/{test["primaryFs"]}/filesets/{volume_name}/snapshots'
        response = requests.get(snap_link, verify=False,
                                auth=(test["username"], test["password"]))
        LOGGER.debug(response.text)
        LOGGER.info(
            f"Snapshot Fileset Check : Checking for Snapshot {snapshot_name} of Fileset {volume_name}")
        response_dict = json.loads(response.text)
        LOGGER.debug(response_dict)

        for snapshot in response_dict["snapshots"]:
            if snapshot["snapshotName"] == snapshot_name:
                return True
        val += 1
        time.sleep(5)
    return False


def create_snapshot(snapshot_name, volume_name, created_objects):
    """
    create snapshot snapshot_name for volume_name
    """
    if scalevalidation == "False":
        LOGGER.warning(f'Create snapshot : skipped for snapshot {snapshot_name} as scalevalidation = "False" in config file'
)
        return

    headers = {
        'content-type': 'application/json',
        'accept': 'application/json',
    }

    data = f'{{ "snapshotName": "{snapshot_name}" }}'
    snap_link = f'https://{test["guiHost"]}:{test["port"]}/scalemgmt/v2/filesystems/{test["primaryFs"]}/filesets/{volume_name}/snapshots'
    response = requests.post(snap_link, headers=headers, data=data, verify=False,
                             auth=(test["username"], test["password"]))
    LOGGER.debug(response.text)
    created_objects["scalesnapshot"].append([snapshot_name, volume_name])
    LOGGER.info(
        f"Static Snapshot Create :snapshot {snapshot_name} created for volume {volume_name}")


def delete_snapshot(snapshot_name, volume_name, created_objects):
    if scalevalidation == "False":
        LOGGER.warning(f'Delete snapshot : skipped for snapshot {snapshot_name} as scalevalidation = "False" in config file'
)
        return
    snap_link = f'https://{test["guiHost"]}:{test["port"]}/scalemgmt/v2/filesystems/{test["primaryFs"]}/filesets/{volume_name}/snapshots/{snapshot_name}'
    response = requests.delete(snap_link, verify=False, auth=(test["username"], test["password"]))
    LOGGER.debug(response.text)
    created_objects["scalesnapshot"].remove([snapshot_name, volume_name])
    LOGGER.info(
        f"Scale Snapshot Delete :snapshot {snapshot_name} of volume {volume_name} is deleted")


def check_snapshot_deleted(snapshot_name, volume_name):
    if scalevalidation == "False":
        LOGGER.warning(f'Check snapshot deleted : skipped for snapshot {snapshot_name} as scalevalidation = "False" in config file')
        return True
    val = 0
    while val < 12:
        snap_link = f'https://{test["guiHost"]}:{test["port"]}/scalemgmt/v2/filesystems/{test["primaryFs"]}/filesets/{volume_name}/snapshots'
        response = requests.get(snap_link, verify=False,
                                auth=(test["username"], test["password"]))
        LOGGER.debug(response.text)
        LOGGER.info(
            f"Snapshot Check : Checking for deletion of snapshot {snapshot_name} of volume {volume_name}")
        response_dict = json.loads(response.text)
        LOGGER.debug(response_dict)

        snapshots_list = []
        for snapshot in response_dict["snapshots"]:
            snapshots_list.append(snapshot["snapshotName"])

        if snapshot_name in snapshots_list:
            val += 1
            time.sleep(10)
        else:
            return True
    return False


def feature_available(feature_name):
    """
    returns True , if passed feature_name available in scale version
    else , returns False
    """
    if scalevalidation == "False":
        LOGGER.warning(f'Check feature available : skipped for feature {feature_name} as scalevalidation = "False" in config file')
        return True
    features = {"snapshot": 5110, "permissions": 5112}
    scale_version = return_scale_version()
    if int(scale_version) >= features[feature_name]:
        return True
    return False


def return_scale_version():
    """
    get IBM Storage Scale version and return it
    """
    if scalevalidation == "False":
        LOGGER.warning(f'Get scale version : skipped for gui {test["guiHost"]} as scalevalidation = "False" in config file'
)
        return "9999"
    get_link = f'https://{test["guiHost"]}:{test["port"]}/scalemgmt/v2/info'
    response = requests.get(get_link, verify=False, auth=(test["username"], test["password"]))
    LOGGER.debug(response.text)

    response_dict = json.loads(response.text)
    LOGGER.info(f'scale version is {response_dict["info"]["serverVersion"]}')
    scale_version = response_dict["info"]["serverVersion"]
    scale_version = scale_version[0] + scale_version[2] + scale_version[4] + scale_version[6]
    return scale_version


def get_scale_version(test_data):
    """
    get IBM Storage Scale version and display it
    """
    if scalevalidation == "False":
        LOGGER.warning(f'Get scale version : skipped for gui {test["guiHost"]} as scalevalidation = "False" in config file')
        return
    get_link = f'https://{test_data["guiHost"]}:{test_data["port"]}/scalemgmt/v2/info'
    response = requests.get(get_link, verify=False, auth=(
        test_data["username"], test_data["password"]))
    LOGGER.debug(response.text)

    response_dict = json.loads(response.text)
    LOGGER.info(f'scale version is {response_dict["info"]["serverVersion"]}')


def get_and_verify_fileset_permissions(volume_name, mode, cg_fileset_name):
    """

    Get permissions for pv path and ensure they match with parameter mode.

    Args:
        volume_name: name of fileset
        mode: expected permissions for persistent volume

    Returns:
       returns True , if persistent volume permissions match with mode
       returns False , if persistent volume permissions and mode mismatch

    Raises:
       None

    Brief:
       An extra parameter "permissions" is added to Storageclass. This helps
       create pv with user specified permissions e.g. permissions=777.
       We want to ensure that pv path is really created with input permissions.

       However, there is no REST API to get mode bits aka permissions (777)
       for a filesystem path. But, another REST API "acl" exists which returns
       acls for a path. We can translate the acls to mode bits.

       e.g. "acl" API returns "rwmxDaAnNcCos"; here acl to mode bit translation
       "r" == 4 , "w" == 2 , "x" == 1
       so for permissions=777, we need to ensure that "rwx" is present for all
       i.e. for owner, group and everyone
    """
    # get acl for a path
    if scalevalidation == "False":
        LOGGER.warning(f'Check fileset permissions : skipped for fileset {volume_name} as scalevalidation = "False" in config file')
        return True

    if cg_fileset_name is not None:
        get_link = f'https://{test["guiHost"]}:{test["port"]}/scalemgmt/v2/filesystems/{test["primaryFs"]}/acl/{cg_fileset_name}%2F{volume_name}'
    else:
        get_link = f'https://{test["guiHost"]}:{test["port"]}/scalemgmt/v2/filesystems/{test["primaryFs"]}/acl/{volume_name}%2F{volume_name}-data'

    response = requests.get(get_link, verify=False, auth=(test["username"], test["password"]))
    LOGGER.debug(response.text)

    # store acl string as python dictionary
    response_dict = json.loads(response.text)

    possible_aces = ["x", "w", "r"]
    mode_to_ace_bits = {
        "1": ["x"],
        "2": ["w"],
        "4": ["r"],
        "5": ["x", "r"],
        "6": ["w", "r"],
        "7": ["x", "w", "r"],
    }

    if "acl" not in response_dict:
        LOGGER.error(f"not able to find acl in {response.text}")
        return False

    # retrieve actual aces set for owner, group, everyone
    owner_aces = response_dict["acl"]["entries"][0]["permissions"]
    group_aces = response_dict["acl"]["entries"][1]["permissions"]
    everyone_aces = response_dict["acl"]["entries"][2]["permissions"]
    # LOGGER.info(owner_aces) #, group_aces, everyone_aces)
    LOGGER.info(
        f'Permissions Check : acebits for owner : {owner_aces} , group : {group_aces} , everyone : {everyone_aces}')

    # expected permissions for owner, group, everyone
    owner, group, everyone = [c for c in mode]
    # LOGGER.info(owner) #, group, everyone)
    LOGGER.info(
        f'Permissions Check : modebits for owner : {owner} , group : {group}, everyone : {everyone}')

    # verify requred ace bits are set for owner, group, everyone
    owner_aces_matched = all([c in owner_aces for c in mode_to_ace_bits[owner]])
    group_aces_matched = all([c in group_aces for c in mode_to_ace_bits[group]])
    everyone_aces_matched = all([c in everyone_aces for c in mode_to_ace_bits[everyone]])
    # LOGGER.info(owner_aces_matched) #, group_aces_matched, everyone_aces_matched)
    LOGGER.info(
        f'Permissions Check : required acebits matched result for owner : {owner_aces_matched}, group : {group_aces_matched} , everyone : {everyone_aces_matched}')

    # retrieve acebits that should not be set for owner, group, everyone
    owner_excluded_aces = [ace for ace in possible_aces if ace not in mode_to_ace_bits[owner]]
    group_excluded_aces = [ace for ace in possible_aces if ace not in mode_to_ace_bits[group]]
    everyone_excluded_aces = [ace for ace in possible_aces if ace not in mode_to_ace_bits[everyone]]
    # LOGGER.info(owner_excluded_aces) #, group_excluded_aces, everyone_excluded_aces)
    LOGGER.info(
        f'Permissions Check : acebits missing for owner : {owner_excluded_aces} , group : {group_excluded_aces} , everyone : {everyone_excluded_aces}')

    # ensure expected and actual missing aces should match
    owner_excluded_aces_missing = all([c not in owner_aces for c in owner_excluded_aces])
    group_excluded_aces_missing = all([c not in group_aces for c in group_excluded_aces])
    everyone_excluded_aces_missing = all([c not in everyone_aces for c in everyone_excluded_aces])
    # LOGGER.info(owner_excluded_aces_missing) #, group_excluded_aces_missing, everyone_excluded_aces_missing)
    LOGGER.info(
        f'Permissions Check : missing acebits result for owner : {owner_excluded_aces_missing} , group : {group_excluded_aces_missing} , everyone : {everyone_excluded_aces_missing}')

    # return True only if required ace bits are set and excluded ace bits are missing for all (owner, group, everyone)
    status = all([owner_aces_matched, group_aces_matched, everyone_aces_matched,
                 owner_excluded_aces_missing, group_excluded_aces_missing, everyone_excluded_aces_missing])
    LOGGER.info(f'Permissions Check : final result : {status}')

    return status


def check_fileset_quota(volume_name, fileset_size, max_inode_from_sc):
    if scalevalidation == "False":
        LOGGER.warning(f'Check fileset quota: skipped for fileset {volume_name} as scalevalidation = "False" in config file')
        return True
    get_link = f'https://{test["guiHost"]}:{test["port"]}/scalemgmt/v2/filesystems/{test["primaryFs"]}/quotas?filter=objectName={volume_name}'
    LOGGER.debug(get_link)
    response = requests.get(get_link, verify=False, auth=(test["username"], test["password"]))
    LOGGER.debug(response.text)

    if not(response.status_code == 200):
        LOGGER.error(f"Response status code is not 200 for {get_link}")
        LOGGER.error(response)
        return False

    response_dict = json.loads(response.text)
    quota_from_api = int(response_dict["quotas"][0]["blockLimit"])
    quota_from_pvc = 0

    power_of_10 = {"M": int(1000**2 / 1024), "G": int(1000**3 / 1024), "T": int(1000**4 / 1024)}
    power_of_2 = {"Mi": int(1024), "Gi": int(1024**2), "Ti": int(1024**3)}

    if fileset_size[-1:] in power_of_10:
        quota_from_pvc = int(fileset_size[:-1]) * power_of_10[fileset_size[-1:]]
    if fileset_size[-2:] in power_of_2:
        quota_from_pvc = int(fileset_size[:-2]) * power_of_2[fileset_size[-2:]]
    if quota_from_pvc < int(1024**2):
        quota_from_pvc = int(1024**2)

    LOGGER.info(
        f"PVC Check : Minimum quota expected = {quota_from_pvc}   Actual quota set = {quota_from_api}")

    if quota_from_api >= quota_from_pvc:
        if max_inode_from_sc is None:
            expected_max_inode = 200000 if quota_from_pvc > int(1024*1024*10) else 100000
        else:
            expected_max_inode = int(max_inode_from_sc)
        return (check_fileset_max_inode(volume_name, expected_max_inode))
    return False


def check_fileset_max_inode(volume_name, expected_max_inode):
    if scalevalidation == "False":
        LOGGER.warning(f'Check fileset max inode: skipped for fileset {volume_name} as scalevalidation = "False" in config file')
        return True
    count = 15
    while count > 0:
        get_link = f'https://{test["guiHost"]}:{test["port"]}/scalemgmt/v2/filesystems/{test["primaryFs"]}/filesets/{volume_name}'
        response = requests.get(get_link, verify=False, auth=(test["username"], test["password"]))

        if not(response.status_code == 200):
            LOGGER.error(f"Response status code is not 200 for {get_link}")
            LOGGER.error(response)
            return False

        response_dict = json.loads(response.text)

        if "maxNumInodes" in response_dict["filesets"][0]["config"]:
            actual_max_inode = int(response_dict["filesets"][0]["config"]['maxNumInodes'])
            if actual_max_inode >= expected_max_inode:
                LOGGER.info(
                    f"PVC Check : Actual maximun number of inodes {actual_max_inode} is greater than expected maximum inodes {expected_max_inode}")
                return True
        count -= 1
        time.sleep(20)
        LOGGER.info(f"PVC Check : Checking maximun number of inodes for {volume_name} fileset")

    LOGGER.error(
        f"PVC Check : Either actual max inode number is smaller than expected max inodes {expected_max_inode} or response does not contain 'maxNumInodes' ( for more info STG Defect 285687)")
    LOGGER.error(response.text)
    return False


def get_scalevalidation():
    return scalevalidation
