#!/bin/bash
#
# Copyright 2022 IBM Corp.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

usage() {
  echo "Usage: $0
                -f|--filesystem <Name of Volume's Source Filesystem>
                -l|--path <full Path of Volume in Primary Filesystem>
                -F|--fileset <name of source fileset>
                -s|--size <size in GB>
                -u|--username <Username of IBM Storage Scale GUI user account.>
                -p|--password <Password of IBM Storage Scale GUI user account.>
                -r|--guihost <Route host name used to route traffic to the IBM Storage Scale GUI service.>
                [-P|--pvname <name for pv>]
                [-c|--storageclass <StorageClass for pv>]
                [-a|--accessmode <AccessMode for pv>]
                [-h|--help] " 1>&2
  exit 1
}

fullUsage() {
  echo "Usage: $0
                -f|--filesystem <Name of Volume's Source Filesystem>
                -l|--path <full Path of Volume in Primary Filesystem>
                -F|--fileset <name of source fileset>
                -s|--size <size in GB>
                -u|--username <Username of IBM Storage Scale GUI user account.>
                -p|--password <Password of IBM Storage Scale GUI user account.>
                -r|--guihost <HostName(or route) used to access IBM Storage Scale GUI service running on Primary Cluster.>
                [-P|--pvname <name for pv>]
                [-c|--storageclass <StorageClass for pv>]
                [-a|--accessmode <AccessMode for pv>]
                [-h|--help]


Example 1:  Directory based static volume 
  This example shows how to create a volume from a directory '/mnt/fs1/staticpv' within the filesystem 'fs1'.

  $0 --filesystem fs1 --path /mnt/fs1/staticpv --size 10 --pvname mystaticpv --guihost ibm-spectrum-scale-gui-ibm-spectrum-scale.apps.cluster.cp.fyre.ibm.com
	
Example 2: Fileset based volume
	This example shows how to create a volume from a fileset 'fileset1' within the filesystem 'fs1'.

	$0 --filesystem fs1 --fileset f1 --size 10 --pvname mystaticpv --guihost ibm-spectrum-scale-gui-ibm-spectrum-scale.apps.cluster.cp.fyre.ibm.com

	Note: The Path specified for option --path must be valid gpfs path from primary filesystem." 1>&2
  exit 1
}

# Generate Yaml
generate_pv_yaml() {
  volhandle=$1
  volname=$2
  volsize=$3
  accessmode=$4
  if [[ -f "${volname}.yaml" ]]; then
    echo "ERROR: File ${volname}.yaml already exist"
    exit 2
  fi

  cat >"${volname}".yaml <<EOL
# -- ${volname}.yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: ${volname}
spec:
  capacity:
    storage: ${volsize}Gi
  accessModes:
    - ${accessmode}
  csi:
    driver: spectrumscale.csi.ibm.com
    volumeHandle: ${volhandle}
  ${STORAGECLASS}
EOL
  echo "INFO: volumeHandle: ${volhandle}"
  echo "INFO: Successfully created ${volname}.yaml"
}

# Generate PVC manifest
generate_pvc_yaml() {
  volname=$1
  volsize=$2
  accessmode=$3
  if [[ -f "pvc-${volname}.yaml" ]]; then
    echo "ERROR: File pvc-${volname}.yaml already exist"
    exit 2
  fi

  cat >pvc-"${volname}".yaml <<EOL
# -- pvc-${volname}.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: pvc-${volname}
  namespace: <PVC namespace> 
spec:
  volumeName: ${volname}
  accessModes:
    - ${accessmode}
  resources:
    requests:
      storage: ${volsize}Gi
  ${STORAGECLASS}
EOL
  echo "INFO: Successfully created pvc-${volname}.yaml"
}

SHORT=hf:l:F:s:P:c:a:u:p:r:
LONG=help,filesystem:,path:,fileset:,size:,pvname:,storageclass:,accessmode:,username:,password:,guihost:
ERROROUT="/tmp/csierror.out"
OPTS=$(getopt --options $SHORT --long $LONG --name "$0" -- "$@")

if [ $? != 0 ]; then
  echo "Failed to parse options...exiting." >&2
  usage
  exit 1
fi
[[ $# -lt 1 ]] && fullUsage

eval set -- "$OPTS"

while true; do
  case "$1" in
  -h | --help)
    fullUsage
    ;;
  -l | --path)
    VOLPATH="$2"
    shift 2
    ;;
  -f | --filesystem)
    FSNAME="$2"
    shift 2
    ;;
  -F | --fileset)
    FSETNAME="$2"
    shift 2
    ;;
  -s | --size)
    VOLSIZE="$2"
    shift 2
    ;;
  -P | --pvname)
    VOLNAME="$2"
    shift 2
    ;;
  -c | --storageclass)
    CLASS="$2"
    shift 2
    ;;
  -a | --accessmode)
    ACCESSMODE="$2"
    shift 2
    ;;
  -u | --username)
    USERNAME="$2"
    shift 2
    ;;
  -p | --password)
    PASSWORD="$2"
    shift 2
    ;;
  -r | --guihost)
    URL="$2"
    shift 2
    ;;
  --)
    shift
    break
    ;;
  *)
    usage
    exit 1
    ;;
  esac
done

# Secure username/password prompt if not passed with flag
if [ -z "$USERNAME" ]; then read -r -p "GUI Username: " USERNAME ; fi
if [ -z "$PASSWORD" ]; then read -r -p "GUI Password: " -s PASSWORD ; echo ; fi

# Pre-requisite check
if ! python3 --version 1>/dev/null 2>${ERROROUT};
then
  echo "ERROR: Pre-requisite check failed. Python3 not found."
  exit 2
fi

# Check for mandatory Params
MPARAM=""
[[ -z "${FSNAME}" ]] && MPARAM="${MPARAM}--filesystem "
[[ -z "${VOLSIZE}" ]] && MPARAM="${MPARAM}--size "

if [ ! -z "$MPARAM" ]; then
  echo "ERROR: Mandatory parameter missing : $MPARAM"
  usage
fi

if [[ -z "${VOLPATH}" && -z "${FSETNAME}" ]]; then
  echo "ERROR: Missing parameter. Either 'path' or 'fileset' is mandatory."
  usage
fi

if [[ ! ${VOLSIZE} =~ ^[1-9][0-9]*$ ]]; then
  echo "ERROR: Provided value for --size=${VOLSIZE} is not valid number."
  exit 2
fi

if [[ ${#VOLNAME} -ge 254 ]]; then
  echo "ERROR: pvname specified against option --pvname must be less than 254 characters."
  exit 2
fi

if [ -z "${VOLNAME}" ]; then
  VOLNAME=${VOLPATH%/}
  VOLNAME=${VOLNAME##*/}
  VOLNAME="pv-${FSNAME}-${VOLNAME}"
  VOLNAME=${VOLNAME,,}
  if [[ ${#VOLNAME} -ge 254 ]]; then
    echo "ERROR: Specify name for pv using option --pvname."
    exit 2
  fi

  if ! [[ "${VOLNAME}" =~ ^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$ ]]; then
    echo "ERROR: Specify name for pv using option --pvname."
    exit 2
  fi
fi

if ! [[ "${VOLNAME}" =~ ^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$ ]]; then
  echo "ERROR: Invalid pv name specified. pv name must satisfy DNS-1123 label requirement."
  exit 2
fi

[[ -z "${ACCESSMODE}" ]] && ACCESSMODE="ReadWriteMany"

if ! [[ "$ACCESSMODE" == "ReadWriteMany" || "$ACCESSMODE" == "ReadWriteOnce" ]]; then
  echo "ERROR: Invalid access mode specified. Valid accessmode are ReadWriteMany and ReadWriteOnce."
  exit 2
fi

STORAGECLASS=""
if ! [[ -z "${CLASS}" ]]; then
  if ! [[ "${CLASS}" =~ ^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$ ]]; then
    echo "ERROR: Invalid storageClass name specified. storageClass name must satisfy DNS-1123 label requirement."
    exit 2
  fi
  STORAGECLASS="storageClassName: ${CLASS}"
else
  STORAGECLASS="storageClassName: \"\""
fi

# Check if this is IBM Storage Scale node
#if [[ ! -f /usr/lpp/mmfs/bin/mmlscluster ]]; then
#  echo "ERROR: IBM Storage Scale cli's are not present on this node"
#  exit 2
#fi

echo >${ERROROUT}

# Authentication and route validation
response=$(curl -kv -u "${USERNAME}":"${PASSWORD}" -X GET \
  --header 'accept:application/json' \
  "https://${URL}:443/scalemgmt/v2/cluster" \
  2>&1 | grep -i 'HTTP/1.1 ' | awk '{print $3}'| sed -e 's/^[ \t]*//')

if [[  ${response} == 401 ]]; then
  echo "ERROR: Unauthorized. Incorrect username or password."
  exit 2
elif [[  -z ${response} ]]; then
  echo "ERROR: Could not resolve host ${URL}."
  exit 2
fi

# Get the IBM Storage Scale cluster ID
clusterID=$(curl -k -u "${USERNAME}":"${PASSWORD}" -X GET \
  --header 'accept:application/json' \
  "https://${URL}:443/scalemgmt/v2/cluster" \
  2>${ERROROUT} | python3 -c "import sys, json; print(json.load(sys.stdin)['cluster']['clusterSummary']['clusterId'])" 2>>${ERROROUT})
if [[ $? -ne 0 ]] || [[ -z "$clusterID" ]]; then
  echo "ERROR: Failed to get the IBM Storage Scale cluster ID."
  #cat ${ERROROUT}
  exit 2
fi

# Get the Fileystem ID
fileSystemID=$(curl -k -u "${USERNAME}":"${PASSWORD}" -X GET \
  --header 'accept:application/json' \
  "https://${URL}:443/scalemgmt/v2/filesystems/${FSNAME}" \
  2>${ERROROUT} | python3 -c "import sys, json; print(json.load(sys.stdin)['filesystems'][0]['uuid'])" 2>>${ERROROUT})
if [[ $? -ne 0 ]] || [[ -z "$fileSystemID" ]]; then
  echo "ERROR: Failed to get the Fileystem ID of ${FSNAME}"
  #cat ${ERROROUT}
  exit 2
fi

# TODO : Add check for kubernetes lable limit for value of VolumeHandle

# echo "FSETNAME=${FSETNAME}"
if [[ -z "${FSETNAME}" ]]; then
  # Verify the path exists and is a GPFS path.
  mountpathDepth=$(curl -k -u "${USERNAME}":"${PASSWORD}" -X GET \
  --header 'accept:application/json' \
  "https://${URL}:443/scalemgmt/v2/filesystems/${FSNAME}" \
  2>${ERROROUT} | python3 -c "import sys, json; print(json.load(sys.stdin)['filesystems'][0]['mount']['mountPoint'])" | grep -o "[\/]" | wc -l) 

  relativePath=""
  for ((i=1;i<=mountpathDepth;i++)); do     relativePath+="../"; done

  relativePath+=$VOLPATH
  relativePath=${relativePath//\//%2F}

  response=$(curl -k -u "${USERNAME}":"${PASSWORD}" -X GET \
  --header 'accept:application/json' \
  "https://${URL}:443/scalemgmt/v2/filesystems/${FSNAME}/owner/${relativePath}" \
  2>${ERROROUT} | python3 -c "import sys, json; print(json.dumps(json.load(sys.stdin)['status']))" 2>>${ERROROUT})

  responseCode=$(echo "$response" | python3 -c "import sys, json; print(json.load(sys.stdin)['code'])")
  responseMsg=$(echo "$response" | python3 -c "import sys, json; print(json.load(sys.stdin)['message'])")

  if [[ $responseCode != 200 ]]; then
    if [[ $responseMsg == "Path is not a valid GPFS path." ]]; then
      echo "ERROR: The Path (${VOLPATH}) is not gpfs path."
      exit 2
    elif [[ $responseMsg == "File not found" ]]; then
      echo "ERROR: Either Path (${VOLPATH}) does not exist or it is not a Directory/Softlink."
      exit 2
    else 
      echo "ERROR: Failed to verify the path (${VOLPATH}). Check error log for details."
      echo "$responseMsg" > ${ERROROUT} 
      exit 2
    fi
  fi
fi


# Generate Volume Handle
# Volume handle format from CSI 2.5.0 onwards:
# <storageclass_type>;<volume_type>;<cluster_id>;<filesystem_uuid>;<consistency_group>;<fileset_name>;<path>
# For static volumes, storageclass_type=0 and consistency_group="" always.

scType="0"
cg=""
volType=""
filesetName=""
path=""
if [[ ! -z "${FSETNAME}" ]]; then
  fsetId=$(curl -k -u "${USERNAME}":"${PASSWORD}" -X GET \
    --header 'accept:application/json' \
    "https://${URL}:443/scalemgmt/v2/filesystems/${FSNAME}/filesets/${FSETNAME}" \
    2>${ERROROUT} | python3 -c "import sys, json; print(json.load(sys.stdin)['filesets'][0]['config']['id'])" 2>>${ERROROUT})
  if [[ $? -ne 0 ]] || [[ -z "$fsetId" ]]; then
    echo "ERROR: Failed to get the fileset ID of ${FSETNAME}."
    #cat ${ERROROUT}
    exit 2
  fi
  
  fsetLinkPath=$(curl -k -u "${USERNAME}":"${PASSWORD}" -X GET \
    --header 'accept:application/json' \
    "https://${URL}:443/scalemgmt/v2/filesystems/${FSNAME}/filesets/${FSETNAME}" \
    2>${ERROROUT} | python3 -c "import sys, json; print(json.load(sys.stdin)['filesets'][0]['config']['path'])" 2>>${ERROROUT})
  if [[ $? -ne 0 ]] || [[ -z "$fsetLinkPath" ]]; then
    echo "ERROR: Failed to get the fileset link path of ${FSETNAME}."
    exit 2
  fi

  if [[ "${fsetLinkPath}" == "--" ]]; then
    echo "ERROR: Fileset ${FSETNAME} is not linked."
    exit 2
  fi

  parentId=$(curl -k -u "${USERNAME}":"${PASSWORD}" -X GET \
    --header 'accept:application/json' \
    "https://${URL}:443/scalemgmt/v2/filesystems/${FSNAME}/filesets/${FSETNAME}" \
    2>${ERROROUT} | python3 -c "import sys, json; print(json.load(sys.stdin)['filesets'][0]['config']['parentId'])" 2>>${ERROROUT})
  if [[ $? -ne 0 ]] || [[ -z "$parentId" ]]; then
    echo "ERROR: Failed to get the parentId of fileset ${FSETNAME}."
    exit 2
  fi

  if [[ "${parentId}" == "0" ]]; then
    volType="2"
  else
    volType="1"
  fi

  filesetName="${FSETNAME}"
  path=${fsetLinkPath}
  
else
  volType="0"
  filesetName=""
  path=${VOLPATH}
fi
VolumeHandle="${scType};${volType};${clusterID};${fileSystemID};${cg};${filesetName};${path}"

# Gererate yaml file
generate_pv_yaml "${VolumeHandle}" "${VOLNAME}" "${VOLSIZE}" "${ACCESSMODE}"
generate_pvc_yaml "${VOLNAME}" "${VOLSIZE}" "${ACCESSMODE}"

rm -f ${ERROROUT}
exit 0
