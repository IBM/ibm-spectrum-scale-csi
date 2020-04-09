#!/bin/bash
#
# Clean-up script REQUIRED ONLY IF 'helm delete <releasename> --purge' for
# this test path will result in orphaned components.
#
# For example, if PersistantVolumes (PVs) are created as pre-requisite to chart installation
# they will need to be deleted post helm delete.
#
# Parameters :
#   -c <chartReleaseName>, the name of the release used to install the helm chart
#
# Pre-req environment: authenticated to cluster & kubectl cli install / setup complete

# Exit when failures occur (including unset variables)
set -o errexit
set -o nounset
set -o pipefail

# Just try and see what's happening in  the log.
output=$(kubectl get po -n "${CV_TEST_NAMESPACE:-default}")
echo $output

pod="$(cut -d' ' -f6 <<< $output)"

kubectl logs -n "${CV_TEST_NAMESPACE:-default}" $pod operator

[[ `dirname $0 | cut -c1` = '/' ]] && preinstallDir=`dirname $0`/ || preinstallDir=`pwd`/`dirname $0`/

# Optional - set tool repo and source library for creating/configuring and removing namespace
. $APP_TEST_LIBRARY_FUNCTIONS/createNamespace.sh

# Process parameters notify of any unexpected
while test $# -gt 0; do
        [[ $1 =~ ^-c|--chartrelease$ ]] && { chartRelease="$2"; shift 2; continue; };
    echo "Parameter not recognized: $1, ignored"
    shift
done
: "${chartRelease:="default"}"

# Verify pre-req environment of kubectl exists
command -v kubectl > /dev/null 2>&1 || { echo "kubectl pre-req is missing."; exit 1; }

# Execute clean-up kubectl commands
# For example, delete PV/PVCs created by pre-install.sh script
#kubectl delete pvc/$chartRelease-pvc
#kubectl delete pv/$chartRelease-pv

#deleteNamespace ${CV_TEST_NAMESPACE}
