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
                -f|--filesystem <Name of Snapshot's Source Filesystem>
                -F|--fileset <Name of Snapshot's Source Fileset>
                -s|--snapshot <Name of the Snapshot>
                [-p|--path <Relative path within the snapshot>]
                [-c|--snapshotcontentname <name for VolumeSnapshotContent>]
                [-v|--snapshotname <name for VolumeSnapshot>]
                [-n|--namespace <namespace for VolumeSnapshot>]
                [-h|--help] " 1>&2; exit 1; }

# Generate Yaml
generate_yaml()
{
snaphandle=$1
snapcontentname=$2
snapname=$3
namespace=$4
if [[ -f "${snapname}.yaml" ]]; then
    echo "ERROR: File ${snapname}.yaml already exist"
    exit 2
fi

/usr/bin/cat > ${snapcontentname}.yaml  <<EOL
# -- ${snapcontentname}.yaml
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshotContent
metadata:
  name: ${snapcontentname}
spec:
  deletionPolicy: Delete
  driver: spectrumscale.csi.ibm.com
  source:
    snapshotHandle: ${snaphandle}
  volumeSnapshotRef:
    name: ${snapname}
    namespace: ${namespace}
EOL
echo "INFO: snapshotHandle: ${snaphandle}"
echo "INFO: Successfully created ${snapcontentname}.yaml"
}


SHORT=hf:F:s:p:v:c:n:
LONG=help,filesystem:,fileset:,snapshot:,path:,volumesnapname:,snapshotcontentname:,namespace:
ERROROUT="/tmp/csierror.out"
OPTS=$(getopt --options $SHORT --long $LONG --name "$0" -- "$@")

if [ $? != 0 ]; then echo "Failed to parse options...exiting." >&2; usage ; exit 1 ; fi
[[ $# -lt 1 ]] && usage

eval set -- "$OPTS"

while true ; do
  case "$1" in
    -h | --help )
      usage
      ;;
    -F | --fileset )
      FSETNAME="$2"
      shift 2 
      ;;
    -f | --filesystem )
      FSNAME="$2"
      shift 2
      ;;
    -s | --snapshot )
      SNAPSHOT="$2"
      shift 2
      ;;
    -p | --path )
      SNAPPATH="$2"
      shift 2
      ;;
    -c | --snapshotcontentname )
      SNAPCONNAME="$2"
      shift 2
      ;;
    -v | --volumesnapname )
      SNAPNAME="$2"
      shift 2
      ;;
    -n | --namespace )
      NAMESPACE="$2"
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
[[ -z "${FSETNAME}" ]] && MPARAM="${MPARAM}--fileset "
[[ -z "${SNAPSHOT}" ]] && MPARAM="${MPARAM}--snapshot "

if [ ! -z "$MPARAM" ]; then
   echo "ERROR: Mandatory parameter missing : $MPARAM"
   usage
fi

if [ -z "${SNAPCONNAME}" ] ; then
    SNAPCONNAME="snapshotcontent-${SNAPSHOT}"
    if [[ ${#SNAPCONNAME} -ge 254 ]]; then
       echo "ERROR: Specify name for volumeSnapshotContent using option --snapshot"
       exit 2
   fi
elif [[ ${#SNAPCONNAME} -ge 254 ]]; then
    echo "ERROR: volumeSnapshotContent name specified against option --snapshotcontentname must be less than 254 characters"
    exit 2
fi

if ! [[ "${SNAPCONNAME}" =~ ^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$ ]]; then
        echo "ERROR: Invalid volumeSnapshotContent name ${SNAPCONNAME}. volumeSnapshotContent name must satisfy DNS-1123 label requirement."
        exit 2
fi

if [ -z "${SNAPNAME}" ] ; then
   SNAPNAME="snapshot-${SNAPSHOT}"
   if [[ ${#SNAPNAME} -ge 254 ]]; then
       echo "ERROR: Specify name for volumeSnapshot using option --snapshot"
       exit 2
   fi
fi

if ! [[ "${SNAPNAME}" =~ ^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$ ]]; then
	echo "ERROR: Invalid volumeSnapshot name specified. volumeSnapshot name must satisfy DNS-1123 label requirement."
        exit 2
fi

[[ -z "${NAMESPACE}" ]] && NAMESPACE="default"

# Check if this is IBM Storage Scale node
if [[ ! -f /usr/lpp/mmfs/bin/mmlscluster ]] ; then
    echo "ERROR: IBM Storage Scale cli's are not present on this node"
    exit 2
fi

echo > ${ERROROUT}

# Get the IBM Storage Scale cluster ID 
clusterID=`/usr/lpp/mmfs/bin/mmlscluster -Y 2>${ERROROUT} | /usr/bin/grep clusterSummary | /usr/bin/grep -v HEADER | /usr/bin/awk '{split($0,a,":"); print a[8]}'` 
if [[ $? -ne 0 ]] || [[ -z "$clusterID" ]]; then
     echo "ERROR: Failed to get the IBM Storage Scale cluster ID"
     /usr/bin/cat ${ERROROUT}
     exit 2
fi

# Get the Fileystem ID 
fileSystemID=`/usr/lpp/mmfs/bin/mmlsfs ${FSNAME} --uid 2>${ERROROUT}  | /usr/bin/tail -1 | /usr/bin/awk '{split($0,a," "); print a[2]}'`
if [[ $? -ne 0 ]] || [[ -z "$fileSystemID" ]]; then
     echo "ERROR: Failed to get the Fileystem ID of ${FSNAME}"
     /usr/bin/cat ${ERROROUT}
     exit 2
fi

# Check if fileset exists
/usr/lpp/mmfs/bin/mmlsfileset ${FSNAME} ${FSETNAME} 1>/dev/null 2>${ERROROUT}
if [[ $? -ne 0 ]]; then
     echo "ERROR: Fileset ${FSETNAME} could not be found in filesystem ${FSNAME}"
     cat ${ERROROUT}
     exit 2
fi

# Check if snapshot exists
/usr/lpp/mmfs/bin/mmlssnapshot ${FSNAME} -s ${FSETNAME}:${SNAPSHOT} 1>/dev/null 2>${ERROROUT}
if [[ $? -ne 0 ]]; then
     echo "ERROR: Snapshot ${FSETNAME}:${SNAPSHOT} could not be found in filesystem ${FSNAME}"
     cat ${ERROROUT}
     exit 2
fi

if [ -z "${SNAPPATH}" ] ; then
    # Generate Volume Handle
    SnapshotHandle="${clusterID};${fileSystemID};${FSETNAME};${SNAPSHOT}"
else
    SnapshotHandle="${clusterID};${fileSystemID};${FSETNAME};${SNAPSHOT};${SNAPPATH}"
fi

# Gererate yaml file
generate_yaml "${SnapshotHandle}" "${SNAPCONNAME}" "${SNAPNAME}" "${NAMESPACE}"

/usr/bin/rm -f ${ERROROUT}
exit 0
