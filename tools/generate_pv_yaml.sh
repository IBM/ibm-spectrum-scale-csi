#!/bin/bash
#
# Copyright 2019 IBM Corp.
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

usage(){
echo "Usage: $0
                -f|--filesystem <Name of Volume's Source Filesystem>
                -l|--linkpath <full Path of Volume in Primary Filesystem>
                -s|--size <size in GB>
                [-p|--pvname <name for pv>]
                [-c|--storageclass <StorageClass for pv>]
                [-a|--accessmode <AccessMode for pv>]
                [-h|--help] " 1>&2; exit 1; }

fullUsage(){
echo "Usage: $0
		-f|--filesystem <Name of Volume's Source Filesystem>
		-l|--linkpath <full Path of Volume in Primary Filesystem>
		-s|--size <size in GB>
		[-p|--pvname <name for pv>]
                [-c|--storageclass <StorageClass for pv>]
                [-a|--accessmode <AccessMode for pv>]
		[-h|--help] 
		

Example 1: Single Fileystem
	In this setup there is only one fileystem 'gpfs0' and directory from the same fileystem is being used as volume.

	$0 --filesystem gpfs0 --linkpath /ibm/gpfs0/fileset1/.volumes/staticpv --size 10 --pvname mystaticpv


Example 2: Two or More Filesystem
	In this setup there are two filesystems 'gpfs0' and 'gpfs1'. gpfs0 is configured as primary fileystem in Spectrum-scale-csi setup. User want to create volume from the directory present in the gpfs1 filesystem. Say the directory in the gpfs1 is /ibm/gpfs1/dir1. As a first step user will create softlink  /ibm/gpfs1/dir1 --> /ibm/gpfs0/fileset1/.volumes/staticpv1 and then run following command to generate the pv.yaml.

	$0 --filesystem gpfs1 --linkpath /ibm/gpfs0/fileset1/.volumes/staticpv1 --size 10 --pvname mystaticpv1

	Note: This script does not validate if softlinks are correctly created.
	      The Path specified for option --linkpath must be valid gpfs path from primary fileystem." 1>&2; exit 1; }

# Generate Yaml
generate_yaml()
{
volhandle=$1
volname=$2
volsize=$3
accessmode=$4
if [[ -f "${volname}.yaml" ]]; then
    echo "ERROR: File ${volname}.yaml already exist"
    exit 2
fi

cat > ${volname}.yaml  <<EOL
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


SHORT=hf:l:s:p:c:a:
LONG=help,filesystem:,linkpath:,size:,pvname:,storageclass:,accessmode:
ERROROUT="/tmp/csierror.out"
OPTS=$(getopt --options $SHORT --long $LONG --name "$0" -- "$@")

if [ $? != 0 ]; then echo "Failed to parse options...exiting." >&2; usage ; exit 1 ; fi
[[ $# -lt 1 ]] && fullUsage

eval set -- "$OPTS"

while true ; do
  case "$1" in
    -h | --help )
      fullUsage
      ;;
    -l | --linkpath )
      VOLPATH="$2"
      shift 2 
      ;;
    -f | --filesystem )
      FSNAME="$2"
      shift 2
      ;;
    -s | --size )
      VOLSIZE="$2"
      shift 2
      ;;
    -p | --pvname )
      VOLNAME="$2"
      shift 2
      ;;
    -c | --storageclass )
      CLASS="$2"
      shift 2
      ;;
    -a | --accessmode )
      ACCESSMODE="$2"
      shift 2
      ;;
    -- )
      shift
      break
      ;;
    *)
      usage
      exit 1
      ;;
  esac
done


# Check for mandatory Params
MPARAM=""
[[ -z "${FSNAME}" ]] && MPARAM="${MPARAM}--filesystem "
[[ -z "${VOLSIZE}" ]] && MPARAM="${MPARAM}--size "
[[ -z "${VOLPATH}" ]] && MPARAM="${MPARAM}--linkpath "

if [ ! -z "$MPARAM" ]; then
   echo "ERROR: Mandatory parameter missing : $MPARAM"
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

if [ -z "${VOLNAME}" ] ; then
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

if ! [[ "$ACCESSMODE" == "ReadWriteMany" || "$ACCESSMODE" == "ReadWriteOnce" ]]
then
        echo "ERROR: Invalid access mode specified. Valid accessmode are ReadWriteMany and ReadWriteOnce."
        exit 2
fi

STORAGECLASS=""
if ! [[ -z "${CLASS}" ]] ; then
	if ! [[ "${CLASS}" =~ ^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$ ]]; then
		echo "ERROR: Invalid storageClass name specified. storageClass name must satisfy DNS-1123 label requirement."
		exit 2
	fi
	STORAGECLASS="storageClassName: ${CLASS}"
fi


# Check if this is spectrum scale node
if [[ ! -f /usr/lpp/mmfs/bin/mmlscluster ]] ; then
    echo "ERROR: Spectrum Scale cli's are not present on this node"
    exit 2
fi

echo > ${ERROROUT}

# Get the Spectrum Scale cluster ID 
clusterID=`/usr/lpp/mmfs/bin/mmlscluster -Y 2>${ERROROUT} | grep clusterSummary | grep -v HEADER | awk '{split($0,a,":"); print a[8]}'` 
if [[ $? -ne 0 ]] || [[ -z "$clusterID" ]]; then
     echo "ERROR: Failed to get the Spectrum Scale cluster ID"
     cat ${ERROROUT}
     exit 2
fi

# Get the Fileystem ID 
fileSystemID=`/usr/lpp/mmfs/bin/mmlsfs ${FSNAME} --uid 2>${ERROROUT}  | tail -1 | awk '{split($0,a," "); print a[2]}'`
if [[ $? -ne 0 ]] || [[ -z "$fileSystemID" ]]; then
     echo "ERROR: Failed to get the Fileystem ID of ${FSNAME}"
     cat ${ERROROUT}
     exit 2
fi

# TODO : Add check for kubernetes lable limit for value of VolumeHandle

# Verify if path exist. It should be either directory or softlink
if ! ( [  -d "${VOLPATH}" ] || [ -L "${VOLPATH}" ]); then
	echo "ERROR: Either Path (${VOLPATH}) does not exist or it is not a Directory/Softlink."
        exit 2
fi

# Check if given path is gpfs path
/usr/lpp/mmfs/bin/mmlsattr ${VOLPATH} &> /dev/null
if [[ $? -ne 0 ]]; then
	echo "ERROR: The Path (${VOLPATH}) is not gpfs path"
        exit 2
fi

# Generate Volume Handle
VolumeHandle="${clusterID};${fileSystemID};path=${VOLPATH}"

# Gererate yaml file
generate_yaml "${VolumeHandle}" "${VOLNAME}" "${VOLSIZE}" "${ACCESSMODE}"

rm -f ${ERROROUT}
exit 0
