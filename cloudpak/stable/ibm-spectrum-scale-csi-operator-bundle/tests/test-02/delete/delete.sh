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

set +o errexit
kubectl delete -f $CV_TEST_BUNDLE_DIR/operators/${operator}/deploy/crds/csiscaleoperators.csi.ibm.com_cr.yaml &
kubectl patch CSIScaleOperator -n ibm-spectrum-scale-csi-driver  ibm-spectrum-scale-csi -p '{"metadata":{"finalizers":[]}}' --type=merge
set -o errexit
kubectl delete -f $CV_TEST_BUNDLE_DIR/operators/${operator}/deploy/crds/csiscaleoperators.csi.ibm.com.crd.yaml
kubectl delete -f $CV_TEST_BUNDLE_DIR/operators/${operator}/deploy/role_binding.yaml
kubectl delete -f $CV_TEST_BUNDLE_DIR/operators/${operator}/deploy/role.yaml
kubectl delete -f $CV_TEST_BUNDLE_DIR/operators/${operator}/deploy/service_account.yaml 
kubectl delete namespace ${CV_TEST_NAMESPACE}
set +o errexit
kubectl patch namespace ibm-spectrum-scale-csi-driver -p '{"metadata":{"finalizers":[]}}' --type=merge
set -o errexit

#$APP_TEST_LIBRARY_FUNCTIONS/operatorDelete.sh \
#    --serviceaccount $CV_TEST_BUNDLE_DIR/operators/${operator}/deploy/service_account.yaml \
#    --crd $CV_TEST_BUNDLE_DIR/operators/${operator}/deploy/crds/csiscaleoperators.csi.ibm.com.crd.yaml \
#    --cr $CV_TEST_BUNDLE_DIR/operators/${operator}/deploy/crds/csiscaleoperators.csi.ibm.com_cr.yaml \
#    --role $CV_TEST_BUNDLE_DIR/operators/${operator}/deploy/role.yaml \
#    --rolebinding $CV_TEST_BUNDLE_DIR/operators/${operator}/deploy/role_binding.yaml \
#    --operator $CV_TEST_BUNDLE_DIR/operators/${operator}/deploy/operator.yaml \

#deleteNamespace ${CV_TEST_NAMESPACE}

