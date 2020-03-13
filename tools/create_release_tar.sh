#!/bin/bash
#
# While not recommended, you can pass in the version instead of being on the tagged/branch
#
# Usage: ./create_release_tar.sh
#        ./create_release_tar.sh v1.0.0   # Not recommended
#
# 

if [[ `which md5sum >> /dev/null; echo $?` != 0 ]]; then
   echo "ERROR, this script requires md5sum."
   exit 1
fi

FILES=( './stable/ibm-spectrum-scale-csi-operator-bundle/operators/ibm-spectrum-scale-csi-operator/deploy/namespace.yaml'
'./generated/installer//ibm-spectrum-scale-csi-operator.yaml'
'./stable/ibm-spectrum-scale-csi-operator-bundle/operators/ibm-spectrum-scale-csi-operator/deploy/crds/csiscaleoperators.csi.ibm.com.cr.yaml'
)

TOPLEVEL=`git rev-parse --show-toplevel`
echo ${TOPLEVEL}

PROJ_NAME=`basename ${TOPLEVEL}`
if [[ -z ${1} ]]; then
   TAG_NAME=`git describe --tags --exact-match || git describe`
else
   TAG_NAME=${1}
fi

if [[ -z ${TAG_NAME} ]]; then
   echo "ERROR, could not determine the tag name, cannot continue."
   exit 1
fi

TARGET_NAME="${PROJ_NAME}-${TAG_NAME}"

echo "DEBUG: PROJ_NAME . . : ${PROJ_NAME}"
echo "DEBUG: TAG_NAME . . .: ${TAG_NAME}"
echo "DEBUG: TARGET_NAME  .: ${TARGET_NAME}"

TAR_FILE="${TARGET_NAME}.tar.gz"
TMP_DIR="${TARGET_NAME}"

echo "Creating ${TMP_DIR} ..."
mkdir -p ${TMP_DIR}

echo "Copying yaml files to ${TMP_DIR} ..."
for f in ${FILES[*]}; do 
   cp ${TOPLEVEL}/${f} ${TMP_DIR}/
done

# Create the md5sum for the yaml files 
#
cd ${TMP_DIR}
md5sum * >> md5sum
cd -

#
# Tar up and clean up working files 
echo "Tar up files ..."
tar -czvf ${TAR_FILE} ./${TMP_DIR}
echo "Cleanup ${TMP_DIR} ..."
rm -rf ${TMP_DIR}

echo "====== Generated Files ======="
ls -ltr ${TAR_FILE}
sum ${TAR_FILE}


