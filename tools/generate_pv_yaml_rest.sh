#!/bin/bash
#
# Copyright 2021 IBM Corp.
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
                -l|--linkpath <full Path of Volume in Primary Filesystem>
                -F|--fileset <name of source fileset>
                -s|--size <size in GB>
                -u|--username <Username of spectrum scale GUI user account.>
                -t|--password <Password of spectrum scale GUI user account.>
                -r|--guihost <Route host name used to route traffic to the spectrum scale GUI service.>
                [-p|--pvname <name for pv>]
                [-c|--storageclass <StorageClass for pv>]
                [-a|--accessmode <AccessMode for pv>]
                [-h|--help] " 1>&2
  exit 1
}

fullUsage() {
  echo "Usage: $0
                -f|--filesystem <Name of Volume's Source Filesystem>
                -l|--linkpath <full Path of Volume in Primary Filesystem>
                -F|--fileset <name of source fileset>
                -s|--size <size in GB>
                -u|--username <Username of spectrum scale GUI user account.>
                -t|--password <Password of spectrum scale GUI user account.>
                -r|--guihost <Route host name used to route traffic to the spectrum scale GUI service.>
                [-p|--pvname <name for pv>]
                [-c|--storageclass <StorageClass for pv>]
                [-a|--accessmode <AccessMode for pv>]
                [-h|--help]


Example 1: Single Fileystem
	In this setup there is only one fileystem 'gpfs0' and directory from the same fileystem is being used as volume.

	$0 --filesystem gpfs0 --linkpath /ibm/gpfs0/fileset1/.volumes/staticpv --size 10 --pvname mystaticpv --guihost ibm-spectrum-scale-gui-ibm-spectrum-scale.apps.hci-cluster.cp.fyre.ibm.com


Example 2: Two or More Filesystem
	In this setup there are two filesystems 'gpfs0' and 'gpfs1'. gpfs0 is configured as primary fileystem in Spectrum-scale-csi setup. User want to create volume from the directory present in the gpfs1 filesystem. Say the directory in the gpfs1 is /ibm/gpfs1/dir1. As a first step user will create softlink  /ibm/gpfs1/dir1 --> /ibm/gpfs0/fileset1/.volumes/staticpv1 and then run following command to generate the pv.yaml.

	$0 --filesystem gpfs1 --linkpath /ibm/gpfs0/fileset1/.volumes/staticpv1 --size 10 --pvname mystaticpv1 --guihost ibm-spectrum-scale-gui-ibm-spectrum-scale.apps.hci-cluster.cp.fyre.ibm.com

Example 3: Fileset based volume
	This example shows how to create a volume from a fileset 'fileset1' within the filesyetem 'gpfs0'.

	$0 --filesystem gpfs0 --fileset fileset1 --size 10 --pvname mystaticpv --guihost ibm-spectrum-scale-gui-ibm-spectrum-scale.apps.hci-cluster.cp.fyre.ibm.com

	Note: This script does not validate if softlinks are correctly created.
	      The Path specified for option --linkpath must be valid gpfs path from primary fileystem." 1>&2
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
  claimRef:
    name: pvc-${volname}
    namespace: ibm-spectrum-scale-csi
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
  namespace: ibm-spectrum-scale-csi
spec:
  accessModes:
    - ${accessmode}
  resources:
    requests:
      storage: ${volsize}Gi
EOL
  echo "INFO: Successfully created pvc-${volname}.yaml"
}

SHORT=hf:l:F:s:p:c:a:u:t:r:
LONG=help,filesystem:,linkpath:,fileset:,size:,pvname:,storageclass:,accessmode:,username:,password:,guihost:
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
  -l | --linkpath)
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
  -p | --pvname)
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
  -t | --password)
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

if [[ ! -z "${VOLPATH}" && ! -z "${FSETNAME}" ]]; then
  echo "ERROR: Missing parameter. Either 'linkpath' or 'fileset' is mandatory"
  usage
fi

if [[ ! ${VOLSIZE} =~ ^[1-9][0-9]*$ ]]; then
  echo "ERROR: Provided value for --size=${VOLSIZE} is not valid number"
  exit 2
fi

if [[ ${#VOLNAME} -ge 254 ]]; then
  echo "ERROR: pvname specified against option --pvname must be less than 254 characters"
  exit 2
fi

if [ -z "${VOLNAME}" ]; then
  VOLNAME=${VOLPATH%/}
  VOLNAME=${VOLNAME##*/}
  VOLNAME="pv-${FSNAME}-${VOLNAME}"
  VOLNAME=${VOLNAME,,}
  if [[ ${#VOLNAME} -ge 254 ]]; then
    echo "ERROR: Specify name for pv using option --pvname"
    exit 2
  fi

  if ! [[ "${VOLNAME}" =~ ^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$ ]]; then
    echo "ERROR: Specify name for pv using option --pvname"
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
fi

# Check if this is spectrum scale node
#if [[ ! -f /usr/lpp/mmfs/bin/mmlscluster ]]; then
#  echo "ERROR: Spectrum Scale cli's are not present on this node"
#  exit 2
#fi

echo >${ERROROUT}

# Get the Spectrum Scale cluster ID
clusterID=$(curl -k -u "${USERNAME}":"${PASSWORD}" -X GET \
  --header 'accept:application/json' \
  "https://${URL}:443/scalemgmt/v2/cluster" \
  2>${ERROROUT} | python3 -c "import sys, json; print(json.load(sys.stdin)['cluster']['clusterSummary']['clusterId'])")
if [[ $? -ne 0 ]] || [[ -z "$clusterID" ]]; then
  echo "ERROR: Failed to get the Spectrum Scale cluster ID"
  cat ${ERROROUT}
  exit 2
fi

# Get the Fileystem ID
fileSystemID=$(curl -k -u "${USERNAME}":"${PASSWORD}" -X GET \
  --header 'accept:application/json' \
  "https://${URL}:443/scalemgmt/v2/filesystems/${FSNAME}" \
  2>${ERROROUT} | python3 -c "import sys, json; print(json.load(sys.stdin)['filesystems'][0]['uuid'])")
if [[ $? -ne 0 ]] || [[ -z "$fileSystemID" ]]; then
  echo "ERROR: Failed to get the Fileystem ID of ${FSNAME}"
  cat ${ERROROUT}
  exit 2
fi

# TODO : Add check for kubernetes lable limit for value of VolumeHandle

echo "FSETNAME=${FSETNAME}"
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
  2>${ERROROUT} | python3 -c "import sys, json; print(json.dumps(json.load(sys.stdin)['status']))")

  responseCode=$(echo "$response" | python3 -c "import sys, json; print(json.load(sys.stdin)['code'])")
  responseMsg=$(echo "$response" | python3 -c "import sys, json; print(json.load(sys.stdin)['message'])")

  if [[ $responseCode != 200 ]]; then
    if [[ $responseMsg == "Path is not a valid GPFS path." ]]; then
      echo "ERROR: The Path (${VOLPATH}) is not gpfs path"
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
if [[ ! -z "${FSETNAME}" ]]; then
  fsetId=$(curl -k -u "${USERNAME}":"${PASSWORD}" -X GET \
    --header 'accept:application/json' \
    "https://${URL}:443/scalemgmt/v2/filesystems/${FSNAME}/filesets/${FSETNAME}" \
    2>${ERROROUT} | python3 -c "import sys, json; print(json.load(sys.stdin)['filesets'][0]['config']['id'])")
  
  fsetLinkPath=$(curl -k -u "${USERNAME}":"${PASSWORD}" -X GET \
    --header 'accept:application/json' \
    "https://${URL}:443/scalemgmt/v2/filesystems/${FSNAME}/filesets/${FSETNAME}" \
    2>${ERROROUT} | python3 -c "import sys, json; print(json.load(sys.stdin)['filesets'][0]['config']['path'])")

  if [[ "${fsetLinkPath}" == "--" ]]; then
    echo "ERROR: Fileset ${FSETNAME} is not linked."
    exit 2
  fi
  VolumeHandle="${clusterID};${fileSystemID};fileset=${fsetId};path=${fsetLinkPath}"
else
  VolumeHandle="${clusterID};${fileSystemID};path=${VOLPATH}"
fi

# Gererate yaml file
generate_pv_yaml "${VolumeHandle}" "${VOLNAME}" "${VOLSIZE}" "${ACCESSMODE}"
generate_pvc_yaml "${VOLNAME}" "${VOLSIZE}" "${ACCESSMODE}"

rm -f ${ERROROUT}
exit 0
