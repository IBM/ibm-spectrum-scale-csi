#!/bin/bash
#
# Pre-install script REQUIRED ONLY IF additional setup is required prior to
# operator install for this test path.
#
# For example, if PersistantVolumes (PVs) are required
#
set -o errexit
set -o nounset
set -o pipefail

preinstallDir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
operator=ibm-spectrum-scale-csi-operator

# Verify pre-req environment
command -v kubectl > /dev/null 2>&1 || { echo "kubectl pre-req is missing."; exit 1; }

# Optional - set tool repo and source library for creating/configuring namespace
# NOTE: toolrepositoryroot needed for setting Policy Security Policy
#. $APP_TEST_LIBRARY_FUNCTIONS/createNamespace.sh
#toolrepositoryroot=$APP_TEST_LIBRARY_FUNCTIONS/../../

set +o errexit
kubectl create namespace ${CV_TEST_NAMESPACE}
set -o errexit

kubectl apply -f $CV_TEST_BUNDLE_DIR/operators/${operator}/deploy/role.yaml
kubectl apply -f $CV_TEST_BUNDLE_DIR/operators/${operator}/deploy/service_account.yaml 
kubectl apply -f $CV_TEST_BUNDLE_DIR/operators/${operator}/deploy/role_binding.yaml
kubectl apply -f $CV_TEST_BUNDLE_DIR/operators/${operator}/deploy/crds/csiscaleoperators.csi.ibm.com.crd.yaml

#kubectl get CSIScaleOperator --namespace=ibm-spectrum-scale-csi-driver
#kubectl patch CSIScaleOperator ibm-spectrum-scale-csi-operator -p '{"metadata":{"finalizers":[]}}' --type=merge --namespace=ibm-spectrum-scale-csi-driver

kubectl apply -f  $CV_TEST_BUNDLE_DIR/operators/${operator}/deploy/operator.yaml

#$APP_TEST_LIBRARY_FUNCTIONS/operatorDeployment.sh \
#    --serviceaccount $CV_TEST_BUNDLE_DIR/operators/${operator}/deploy/service_account.yaml \
#    --crd $CV_TEST_BUNDLE_DIR/operators/${operator}/deploy/crds/csiscaleoperators.csi.ibm.com.crd.yaml \
#    --role $CV_TEST_BUNDLE_DIR/operators/${operator}/deploy/role.yaml \
#    --rolebinding $CV_TEST_BUNDLE_DIR/operators/${operator}/deploy/role_binding.yaml \
#    --operator $CV_TEST_BUNDLE_DIR/operators/${operator}/deploy/operator.yaml
#    # --secretname FIXME \
#    # --imagename FIXME

# Optional setup for hardcoded namespace(s) with specific Pod Security Policy
# NOTE: clean-up.sh need to contain matching removeNamespace
# removeNamespace testopr ibm-privileged-psp || true
# configureNamespace testopr ibm-privileged-psp
