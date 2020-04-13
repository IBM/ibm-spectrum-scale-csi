#!/bin/bash
#
# Wait script for resouces to become available in the cluster
#
set -o errexit
set -o nounset
set -o pipefail

waitDir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

echo "Start wait.sh file...."

dep="ibm-spectrum-scale-csi-operator"
retries=20 # 10 minute timeout
kubectl get deployment -A 
echo "here"
while ! kubectl rollout status -w "deployment/${dep}" --namespace=$CV_TEST_NAMESPACE; do
    sleep 30
    echo "here"
    kubectl get deployment -A 
    kubectl describe deployment -n $CV_TEST_NAMESPACE $dep
    kubectl rollout status -w "deployment/${dep}" --namespace=$CV_TEST_NAMESPACE
    retries=$((retries - 1))
    if [[ $retries == 0 ]]; then
      echo "FAIL: Failed to rollout deployloyment $dep"
      exit 1
    fi
    echo "retrying check rollout status for deployment $dep..."
done

# kubectl get deployments -l app.kubernetes.io/instance=$labelname -n $namespace
echo "Successfully rolled out deployment \"$dep\" in namespace \"$CV_TEST_NAMESPACE\""

echo "Checking the ibm-spectrum-scale-csi-operator status..."
  retries=20 # 10 minute timeout
  while [[ "$(kubectl get po -n "${CV_TEST_NAMESPACE:-default}" | grep "ibm-spectrum-scale-csi-operator*"  | awk '{ print $3 }')" != "Running" ]]; do
    sleep 30
    retries=$((retries - 1))
    if [[ $retries == 0 ]]; then
      echo "FAIL: Failed to check the status for ibm-spectrum-scale-csi-operator"
      exit 1
    fi
    echo "retrying check the status for ibm-spectrum-scale-csi-operator"
  done
echo "Successfully check the status for ibm-spectrum-scale-csi-operator in namespace \"$CV_TEST_NAMESPACE\""

echo "Checking the ibm-spectrum-scale-csi-operator status..."
  retries=20 # 10 minute timeout
  while [[ "$(kubectl get po -n "${CV_TEST_NAMESPACE:-default}" | grep "ibm-spectrum-scale-csi-operator*"  | awk '{ print $3 }' | sed -n '1p')" != "Running" ]]; do
    sleep 30
    retries=$((retries - 1))
    if [[ $retries == 0 ]]; then
      echo "FAIL: Failed to check the status for ibm-spectrum-scale-csi-operator"
      exit 1
    fi
    echo "retrying check the status for ibm-spectrum-scale-csi-operator"
  done
echo "Successfully check the status for ibm-spectrum-scale-csi-operator in namespace \"$CV_TEST_NAMESPACE\""
