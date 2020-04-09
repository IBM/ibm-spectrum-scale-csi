#!/bin/bash
#
# Delete script for resources tested
#
set -o errexit
set -o nounset
set -o pipefail

operator=ibm-spectrum-scale-csi-operator
deleteDir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
echo "deleteDir is "
echo $deleteDir

$APP_TEST_LIBRARY_FUNCTIONS/operatorDelete.sh \
    --serviceaccount $CV_TEST_BUNDLE_DIR/operators/${operator}/deploy/service_account.yaml \
    --crd $CV_TEST_BUNDLE_DIR/operators/${operator}/deploy/crds/csiscaleoperators.csi.ibm.com.crd.yaml \
    --cr $CV_TEST_BUNDLE_DIR/operators/${operator}/deploy/crds/csiscaleoperators.csi.ibm.com_cr.yaml \
    --role $CV_TEST_BUNDLE_DIR/operators/${operator}/deploy/role.yaml \
    --rolebinding $CV_TEST_BUNDLE_DIR/operators/${operator}/deploy/role_binding.yaml \
    --operator $CV_TEST_BUNDLE_DIR/operators/${operator}/deploy/operator.yaml \

#deleteNamespace ${CV_TEST_NAMESPACE}

