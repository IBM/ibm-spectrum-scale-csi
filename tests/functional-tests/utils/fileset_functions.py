import logging
import time
import re
import json
import urllib3
import requests
LOGGER = logging.getLogger()



def delete_fileset(test):
    """
    Deletes the primaryFset provided in configuration file
    if primaryFset not deleted , asserts

    Args:
        param1: test : contents of configuration file

    Returns:
       None

    Raises:
       None

    """
    urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
    unlink_fileset(test)
    delete_link = "https://"+test["guiHost"]+":"+test["port"] + \
        "/scalemgmt/v2/filesystems/" + \
        test["primaryFs"]+"/filesets/"+test["primaryFset"]
    response = requests.delete(
        delete_link, verify=False, auth=(test["username"], test["password"]))
    LOGGER.debug(response.text)
    time.sleep(10)
    get_link = "https://"+test["guiHost"]+":"+test["port"] + \
        "/scalemgmt/v2/filesystems/"+test["primaryFs"]+"/filesets/"
    response = requests.get(get_link, verify=False, auth=(test["username"], test["password"]))
    LOGGER.debug(response.text)
    search_format = '"'+test["primaryFset"]+'",'
    search_result = re.search(search_format, str(response.text))
    LOGGER.debug(search_result)
    if search_result is None:
        LOGGER.info(
            f'Success : Fileset {test["primaryFset"]} has been deleted')
    else:
        LOGGER.error(
            f'Failed : Fileset {test["primaryFset"]} has not been deleted')
        assert False


def unlink_fileset(test):
    """
    unlink primaryFset provided in configuration file

    Args:
        param1: test : contents of configuration file

    Returns:
       None

    Raises:
       None

    """
    time.sleep(10)
    urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
    unlink_link = "https://"+test["guiHost"]+":"+test["port"] + \
        "/scalemgmt/v2/filesystems/" + \
        test["primaryFs"]+"/filesets/"+test["primaryFset"]+"/link"
    response = requests.delete(
        unlink_link, verify=False, auth=(test["username"], test["password"]))
    LOGGER.debug(response.text)
    LOGGER.info(f'Fileset {test["primaryFset"]} unlinked')
    time.sleep(10)


def fileset_exists(test):
    """

    Checks primaryFset provided in configuration file exists or not

    Args:
        param1: test : contents of configuration file

    Returns:
       returns True , if primaryFset exists
       returns False , if primaryFset does not exists

    Raises:
       None

    """
    urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
    get_link = "https://"+test["guiHost"]+":"+test["port"] + \
        "/scalemgmt/v2/filesystems/"+test["primaryFs"]+"/filesets/"
    time.sleep(10)
    response = requests.get(get_link, verify=False, auth=(test["username"], test["password"]))
    LOGGER.debug(response.text)
    if str(response) != "<Response [200]>":
        LOGGER.error("API response is ")
        LOGGER.error(str(response))
        LOGGER.error("API resonpse is not 200")
        LOGGER.error("Recheck parameters of config file")
        assert False
    search_format = '"'+test["primaryFset"]+'",'
    search_result = re.search(search_format, str(response.text))
    if search_result is None:
        return False
    return True


def link_fileset(test):
    """
    link primaryFset provided in configuration file

    Args:
        param1: test : contents of configuration file

    Returns:
       None

    Raises:
       None

    """
    urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
    fileset_link = "https://"+test["guiHost"]+":"+test["port"] + \
        "/scalemgmt/v2/filesystems/" + \
        test["primaryFs"]+"/filesets/"+test["primaryFset"]+"/link"
    response = requests.post(fileset_link, verify=False,
                             auth=(test["username"], test["password"]))
    LOGGER.debug(response.text)
    LOGGER.info(f'Fileset {test["primaryFset"]} linked')
    time.sleep(5)


def create_fileset(test):
    """
    create primaryFset provided in configuration file

    Args:
        param1: test : contents of configuration file

    Returns:
       None

    Raises:
       None

    """
    urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
    create_fileset_link = "https://"+test["guiHost"]+":"+test["port"] + \
        "/scalemgmt/v2/filesystems/"+test["primaryFs"]+"/filesets"
    headers = {
        'content-type': 'application/json',
        'accept': 'application/json',
    }
    data = '{"filesetName":"'+test["primaryFset"]+'","path":"' + \
        test["scaleHostpath"]+'/'+test["primaryFset"]+'"}'
    response = requests.post(create_fileset_link, headers=headers,
                             data=data, verify=False, auth=(test["username"], test["password"]))
    LOGGER.debug(response.text)
    LOGGER.info(f'Fileset {test["primaryFset"]} created and linked')
    time.sleep(5)


def unmount_fs(test):
    """
    unmount primaryFs provided in configuration file

    Args:
        param1: test : contents of configuration file

    Returns:
       None

    Raises:
       None

    """
    urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
    unmount_link = "https://"+test["guiHost"]+":"+test["port"] + \
        "/scalemgmt/v2/filesystems/"+test["primaryFs"]+"/unmount"
    LOGGER.debug(unmount_link)
    headers = {
        'content-type': 'application/json',
        'accept': 'application/json',
    }
    data = '{"nodes":["'+test["guiHost"]+'"],"force": false}'
    response = requests.put(unmount_link, headers=headers,
                            data=data, verify=False, auth=(test["username"], test["password"]))
    LOGGER.info(response.text)
    LOGGER.info(f'primaryFS {test["primaryFs"]} unmounted')
    time.sleep(5)


def mount_fs(test):
    """
    mount primaryFs provided in configuration file

    Args:
        param1: test : contents of configuration file

    Returns:
       None

    Raises:
       None

    """
    urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
    mount_link = "https://"+test["guiHost"]+":"+test["port"] + \
        "/scalemgmt/v2/filesystems/"+test["primaryFs"]+"/mount"
    LOGGER.debug(mount_link)
    headers = {
        'content-type': 'application/json',
        'accept': 'application/json',
    }
    data = '{"nodes":["'+test["guiHost"]+'"]}'
    response = requests.put(mount_link, headers=headers,
                            data=data, verify=False, auth=(test["username"], test["password"]))
    LOGGER.debug(response.text)
    LOGGER.info(f'primaryFS {test["primaryFs"]} mounted')
    time.sleep(5)


def cleanup(test):
    """
    deletes all primaryFsets that matches pattern pvc-
    used for cleanup in case of parallel pvc.

    Args:
        param1: test : contents of configuration file

    Returns:
       None

    Raises:
       None

    """
    urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
    get_link = "https://"+test["guiHost"]+":"+test["port"] + \
        "/scalemgmt/v2/filesystems/"+test["primaryFs"]+"/filesets/"
    response = requests.get(get_link, verify=False, auth=(test["username"], test["password"]))
    lst = re.findall(r'\S+pvc-\S+', response.text)
    lst2 = []
    for l in lst:
        lst2.append(l[1:-2])
    for res in lst2:
        volume_name = res
        unlink_link = "https://"+test["guiHost"]+":"+test["port"] + \
            "/scalemgmt/v2/filesystems/" + \
            test["primaryFs"]+"/filesets/"+volume_name+"/link"
        response = requests.delete(
            unlink_link, verify=False, auth=(test["username"], test["password"]))
        LOGGER.debug(response.text)
        LOGGER.info(f"Fileset {volume_name} unlinked")
        time.sleep(5)
        delete_link = "https://"+test["guiHost"]+":"+test["port"] + \
            "/scalemgmt/v2/filesystems/" + \
            test["primaryFs"]+"/filesets/"+volume_name
        response = requests.delete(
            delete_link, verify=False, auth=(test["username"], test["password"]))
        LOGGER.debug(response.text)
        time.sleep(10)
        get_link = "https://"+test["guiHost"]+":"+test["port"] + \
            "/scalemgmt/v2/filesystems/"+test["primaryFs"]+"/filesets/"
        response = requests.get(get_link, verify=False,
                                auth=(test["username"], test["password"]))
        LOGGER.debug(response.text)
        search_result = re.search(volume_name, str(response.text))
        if search_result is None:
            LOGGER.info(f'Fileset {volume_name} deleted successfully')
        else:
            LOGGER.error(f'Fileset {volume_name} is not deleted')
            assert False


def delete_created_fileset(test, volume_name):
    """
    deletes primaryFset created by pvc

    Args:
        param1: test : contents of configuration file
        param2: volume_name : fileset to be deleted

    Returns:
       None

    Raises:
       None

    """
    urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
    get_link = "https://"+test["guiHost"]+":"+test["port"] + \
        "/scalemgmt/v2/filesystems/"+test["primaryFs"]+"/filesets/"
    response = requests.get(get_link, verify=False, auth=(test["username"], test["password"]))
    LOGGER.debug(response.text) 
    search_result = re.search(volume_name, str(response.text))
    LOGGER.debug(search_result)
    if search_result is None:
        LOGGER.info(f'Fileset {volume_name} has already been deleted')
    else:
        unlink_link = "https://"+test["guiHost"]+":"+test["port"] + \
            "/scalemgmt/v2/filesystems/" + \
            test["primaryFs"]+"/filesets/"+volume_name+"/link"
        response = requests.delete(
            unlink_link, verify=False, auth=(test["username"], test["password"]))
        LOGGER.debug(response.text)
        LOGGER.info(f'Fileset {volume_name} has been unlinked')
        time.sleep(5)
        delete_link = "https://"+test["guiHost"]+":"+test["port"] + \
            "/scalemgmt/v2/filesystems/" + \
            test["primaryFs"]+"/filesets/"+volume_name
        response = requests.delete(
            delete_link, verify=False, auth=(test["username"], test["password"]))
        LOGGER.debug(response.text)

        for _ in range(0,24):
            get_link = "https://"+test["guiHost"]+":"+test["port"] + \
            "/scalemgmt/v2/filesystems/"+test["primaryFs"]+"/filesets/"
            response = requests.get(get_link, verify=False,
                                auth=(test["username"], test["password"]))
            LOGGER.debug(response.text)
            search_result = re.search(volume_name, str(response.text))
            LOGGER.debug(search_result)
            if search_result is None:
                LOGGER.info(f'Fileset {volume_name} has been deleted successfully')
                return
            else:
                time.sleep(5)
        LOGGER.error(f'Fileset {volume_name} deletion operation failed')
        assert False


def created_fileset_exists(test, volume_name):
    """
    Checks fileset volume_name exists or not

    Args:
        param1: test : contents of configuration file
        param2: volume_name : fileset to be checked

    Returns:
       returns True , if volume_name exists
       returns False , if volume_name does not exists

    Raises:
       None

    """
    urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
    get_link = "https://"+test["guiHost"]+":"+test["port"] + \
        "/scalemgmt/v2/filesystems/"+test["primaryFs"]+"/filesets/"
    time.sleep(10)
    response = requests.get(get_link, verify=False, auth=(test["username"], test["password"]))
    LOGGER.debug(response.text)
    search_format = '"'+volume_name+'",'
    search_result = re.search(search_format, str(response.text))
    if search_result is None:
        return False
    return True


def create_dir(test, dir_name):
    """
    creates directory named dir_name

    Args:
       param1: test : contents of configuration file
       param2: dir_name : name of directory to be created

    Returns:
       None

    Raises:
       None

    """
    urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
    dir_link = "https://"+test["guiHost"]+":"+test["port"] + \
        "/scalemgmt/v2/filesystems/"+test["primaryFs"]+"/directory/"+dir_name
    LOGGER.debug(dir_link)
    headers = {
        'content-type': 'application/json',
        'accept': 'application/json',
    }
    data = '{ "user":"' + test["uid_name"]+'", "uid":'+test["uid_number"] + \
        ', "group":"'+test["gid_name"]+'", "gid":' + test["gid_number"]+' }'
    response = requests.post(dir_link, headers=headers,
                             data=data, verify=False, auth=(test["username"], test["password"]))
    LOGGER.debug(response.text)
    LOGGER.info(f'Created directory {dir_name}')
    time.sleep(5)


def delete_dir(test, dir_name):
    """
    deleted directory named dir_name

    Args:
       param1: test : contents of configuration file
       param2: dir_name : name of directory to be deleted

    Returns:
       None

    Raises:
       None

    """
    urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
    dir_link = "https://"+test["guiHost"]+":"+test["port"] + \
        "/scalemgmt/v2/filesystems/"+test["primaryFs"]+"/directory/"+dir_name
    LOGGER.debug(dir_link)
    headers = {
        'content-type': 'application/json',
    }
    response = requests.delete(
        dir_link, headers=headers, verify=False, auth=(test["username"], test["password"]))
    LOGGER.debug(response.text)
    LOGGER.info(f'Deleted directory {dir_name}')
    time.sleep(5)


def get_FSUID(test):
    """
    return the UID of primaryFs

    Args:
       param1: test : contents of configuration file

    Returns:
       FSUID

    Raises:
       None

    """
    urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
    info_filesystem = "https://"+test["guiHost"]+":"+test["port"] + \
        "/scalemgmt/v2/filesystems/"+test["primaryFs"]+"?fields=:all:"
    response = requests.get(info_filesystem, verify=False,
                            auth=(test["username"], test["password"]))
    LOGGER.debug(response.text)
    lst = re.findall(r'\S+uuid[\S.\s]+', response.text)
    FSUID = str(lst[0][10:27])
    LOGGER.debug(FSUID)
    return FSUID

def get_mount_point(test):
    """
    return th mount point of primaryFs

    Args:
       param1: test : contents of configuration file

    Returns:
       mount point

    Raises:
       None

    """


    urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
    if "type_remote" in test:
        info_filesystem = "https://"+test["type_remote"]["guiHost"]+":"+test["type_remote"]["port"] + \
        "/scalemgmt/v2/filesystems/"+test["remoteFs"]+"?fields=:all:"
        response = requests.get(info_filesystem, verify=False,
                            auth=(test["type_remote"]["username"], test["type_remote"]["password"]))
    else:
        info_filesystem = "https://"+test["guiHost"]+":"+test["port"] + \
        "/scalemgmt/v2/filesystems/"+test["primaryFs"]+"?fields=:all:"
        response = requests.get(info_filesystem, verify=False,
                            auth=(test["username"], test["password"]))
    LOGGER.debug(response.text)
    response_dict = json.loads(response.text) 
    mount_point = response_dict["filesystems"][0]["mount"]["mountPoint"]
    LOGGER.debug(mount_point)
    return mount_point


def get_remoteFs_remotename(test):
    """ return name of remote filesystem's remote name """
    urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

    info_filesystem = "https://"+test["guiHost"]+":"+test["port"] + \
        "/scalemgmt/v2/filesystems/"+test["remoteFs"]
    response = requests.get(info_filesystem, verify=False,
                            auth=(test["username"], test["password"]))
    LOGGER.debug(response.text)
    
    response_dict = json.loads(response.text)
    for filesystem in response_dict["filesystems"]:
        if filesystem["name"] == test["remoteFs"]:
            device_name = filesystem["mount"]["remoteDeviceName"]
            LOGGER.debug(device_name)
            temp_split = device_name.split(":")
            return temp_split[1]
    return None
