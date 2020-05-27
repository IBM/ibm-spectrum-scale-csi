#!/bin/bash
#
#
# Install script to install the operator
#
set -o errexit
set -o nounset
set -o pipefail

operator=ibm-spectrum-scale-csi-operator
installDir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

$APP_TEST_LIBRARY_FUNCTIONS/operatorInstall.sh \
	--cr $CV_TEST_BUNDLE_DIR/operators/${operator}/deploy/crds/csiscaleoperators.csi.ibm.com_cr.yaml
