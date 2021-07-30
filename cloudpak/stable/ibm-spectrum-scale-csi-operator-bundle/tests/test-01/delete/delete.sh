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
kubectl delete -f $CV_TEST_BUNDLE_DIR/operators/${operator}/config/samples/csi_v1_csiscaleoperator.yaml &
kubectl patch CSIScaleOperator -n ibm-spectrum-scale-csi-driver  ibm-spectrum-scale-csi -p '{"metadata":{"finalizers":[]}}' --type=merge
set -o errexit
kubectl delete -f $CV_TEST_BUNDLE_DIR/operators/${operator}/config/crd/bases/csi.ibm.com_csiscaleoperators.yaml
kubectl delete -f $CV_TEST_BUNDLE_DIR/operators/${operator}/config/rbac/role_binding.yaml
kubectl delete -f $CV_TEST_BUNDLE_DIR/operators/${operator}/config/rbac/role.yaml
kubectl delete -f $CV_TEST_BUNDLE_DIR/operators/${operator}/config/rbac/service_account.yaml 
kubectl delete namespace ${CV_TEST_NAMESPACE}
set +o errexit
kubectl patch namespace ibm-spectrum-scale-csi-driver -p '{"metadata":{"finalizers":[]}}' --type=merge
set -o errexit

