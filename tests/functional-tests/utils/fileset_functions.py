import logging
import base64
import time
import re
import urllib3
import requests
LOGGER = logging.getLogger()


def username_password_setter(test):
    """
    Decodes username and password from base 64 and make them global.

    Args:
        param1: test : contents of configuration file

    Returns:
       None

    Raises:
       None

    """
    global username, password
    urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
    username = base64.b64decode(test["username"]).decode('utf-8')
    password = base64.b64decode(test["password"]).decode('utf-8')


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
    username_password_setter(test)
    unlink_fileset(test)
    delete_link = "https://"+test["guiHost"]+":"+test["port"] + \
        "/scalemgmt/v2/filesystems/" + \
        test["primaryFs"]+"/filesets/"+test["primaryFset"]
    response = requests.delete(
        delete_link, verify=False, auth=(username, password))
    LOGGER.debug(response.text)
    time.sleep(10)
    get_link = "https://"+test["guiHost"]+":"+test["port"] + \
        "/scalemgmt/v2/filesystems/"+test["primaryFs"]+"/filesets/"
    response = requests.get(get_link, verify=False, auth=(username, password))
    LOGGER.debug(response.text)
    search_format = '"'+test["primaryFset"]+'",'
    search_result = re.search(search_format, str(response.text))
    LOGGER.info(search_result)
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
    username_password_setter(test)
    unlink_link = "https://"+test["guiHost"]+":"+test["port"] + \
        "/scalemgmt/v2/filesystems/" + \
        test["primaryFs"]+"/filesets/"+test["primaryFset"]+"/link"
    response = requests.delete(
        unlink_link, verify=False, auth=(username, password))
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
    username_password_setter(test)
    get_link = "https://"+test["guiHost"]+":"+test["port"] + \
        "/scalemgmt/v2/filesystems/"+test["primaryFs"]+"/filesets/"
    time.sleep(10)
    response = requests.get(get_link, verify=False, auth=(username, password))
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
    username_password_setter(test)
    fileset_link = "https://"+test["guiHost"]+":"+test["port"] + \
        "/scalemgmt/v2/filesystems/" + \
        test["primaryFs"]+"/filesets/"+test["primaryFset"]+"/link"
    response = requests.post(fileset_link, verify=False,
                             auth=(username, password))
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
    username_password_setter(test)
    create_fileset_link = "https://"+test["guiHost"]+":"+test["port"] + \
        "/scalemgmt/v2/filesystems/"+test["primaryFs"]+"/filesets"
    headers = {
        'content-type': 'application/json',
        'accept': 'application/json',
    }
    data = '{"filesetName":"'+test["primaryFset"]+'","path":"' + \
        test["scaleHostpath"]+'/'+test["primaryFset"]+'"}'
    response = requests.post(create_fileset_link, headers=headers,
                             data=data, verify=False, auth=(username, password))
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
    username_password_setter(test)
    unmount_link = "https://"+test["guiHost"]+":"+test["port"] + \
        "/scalemgmt/v2/filesystems/"+test["primaryFs"]+"/unmount"
    LOGGER.debug(unmount_link)
    headers = {
        'content-type': 'application/json',
        'accept': 'application/json',
    }
    data = '{"nodes":["'+test["guiHost"]+'"],"force": false}'
    response = requests.put(unmount_link, headers=headers,
                            data=data, verify=False, auth=(username, password))
    LOGGER.debug(response.text)
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
    username_password_setter(test)
    mount_link = "https://"+test["guiHost"]+":"+test["port"] + \
        "/scalemgmt/v2/filesystems/"+test["primaryFs"]+"/mount"
    LOGGER.debug(mount_link)
    headers = {
        'content-type': 'application/json',
        'accept': 'application/json',
    }
    data = '{"nodes":["'+test["guiHost"]+'"]}'
    response = requests.put(mount_link, headers=headers,
                            data=data, verify=False, auth=(username, password))
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
    username_password_setter(test)
    get_link = "https://"+test["guiHost"]+":"+test["port"] + \
        "/scalemgmt/v2/filesystems/"+test["primaryFs"]+"/filesets/"
    response = requests.get(get_link, verify=False, auth=(username, password))
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
            unlink_link, verify=False, auth=(username, password))
        LOGGER.debug(response.text)
        LOGGER.info(f"Fileset {volume_name} unlinked")
        time.sleep(5)
        delete_link = "https://"+test["guiHost"]+":"+test["port"] + \
            "/scalemgmt/v2/filesystems/" + \
            test["primaryFs"]+"/filesets/"+volume_name
        response = requests.delete(
            delete_link, verify=False, auth=(username, password))
        LOGGER.debug(response.text)
        time.sleep(10)
        get_link = "https://"+test["guiHost"]+":"+test["port"] + \
            "/scalemgmt/v2/filesystems/"+test["primaryFs"]+"/filesets/"
        response = requests.get(get_link, verify=False,
                                auth=(username, password))
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
    username_password_setter(test)
    get_link = "https://"+test["guiHost"]+":"+test["port"] + \
        "/scalemgmt/v2/filesystems/"+test["primaryFs"]+"/filesets/"
    response = requests.get(get_link, verify=False, auth=(username, password))
    LOGGER.debug(response.text)
    search_result = re.search(volume_name, str(response.text))
    if search_result is None:
        LOGGER.info(f'Fileset {volume_name} has already been deleted')
    else:
        unlink_link = "https://"+test["guiHost"]+":"+test["port"] + \
            "/scalemgmt/v2/filesystems/" + \
            test["primaryFs"]+"/filesets/"+volume_name+"/link"
        response = requests.delete(
            unlink_link, verify=False, auth=(username, password))
        LOGGER.debug(response.text)
        LOGGER.info(f'Fileset {volume_name} has been unlinked')
        time.sleep(5)
        delete_link = "https://"+test["guiHost"]+":"+test["port"] + \
            "/scalemgmt/v2/filesystems/" + \
            test["primaryFs"]+"/filesets/"+volume_name
        response = requests.delete(
            delete_link, verify=False, auth=(username, password))
        LOGGER.debug(response.text)
        time.sleep(10)
        get_link = "https://"+test["guiHost"]+":"+test["port"] + \
            "/scalemgmt/v2/filesystems/"+test["primaryFs"]+"/filesets/"
        response = requests.get(get_link, verify=False,
                                auth=(username, password))
        LOGGER.debug(response.text)
        search_result = re.search(volume_name, str(response.text))
        if search_result is None:
            LOGGER.info(f'Fileset {volume_name} has been deleted successfully')
        else:
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
    username_password_setter(test)
    get_link = "https://"+test["guiHost"]+":"+test["port"] + \
        "/scalemgmt/v2/filesystems/"+test["primaryFs"]+"/filesets/"
    time.sleep(10)
    response = requests.get(get_link, verify=False, auth=(username, password))
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
    username_password_setter(test)
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
                             data=data, verify=False, auth=(username, password))
    LOGGER.debug(response.text)
    LOGGER.info(f'Created directory {dir_name} in {test["scaleHostpath"]}')
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
    username_password_setter(test)
    dir_link = "https://"+test["guiHost"]+":"+test["port"] + \
        "/scalemgmt/v2/filesystems/"+test["primaryFs"]+"/directory/"+dir_name
    LOGGER.debug(dir_link)
    headers = {
        'content-type': 'application/json',
    }
    response = requests.delete(
        dir_link, headers=headers, verify=False, auth=(username, password))
    LOGGER.debug(response.text)
    LOGGER.info(f'Deleted directory {dir_name} in {test["scaleHostpath"]}')
    time.sleep(5)


def get_FSUID(test):
    """
    return th UID of primaryFs

    Args:
       param1: test : contents of configuration file

    Returns:
       FSUID

    Raises:
       None

    """
    username_password_setter(test)
    info_filesystem = "https://"+test["guiHost"]+":"+test["port"] + \
        "/scalemgmt/v2/filesystems/"+test["primaryFs"]+"?fields=:all:"
    response = requests.get(info_filesystem, verify=False,
                            auth=(username, password))
    LOGGER.debug(response.text)
    lst = re.findall(r'\S+uuid[\S.\s]+', response.text)
    FSUID = str(lst[0][10:27])
    LOGGER.debug(FSUID)
    return FSUID
