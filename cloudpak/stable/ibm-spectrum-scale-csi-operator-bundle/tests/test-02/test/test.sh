#!/bin/bash
#
# Test script REQUIRED to test your operator
#
set -o errexit
set -o nounset
set -o pipefail

testDir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
echo "testDir is"
echo $testDir

command -v kubectl > /dev/null 2>&1 || { echo "kubectl pre-req is missing."; exit 1; }

echo "Check the ibm-spectrum-scale-csi-operator status"
output=$(kubectl get po -n "${CV_TEST_NAMESPACE:-default}")
echo $output

