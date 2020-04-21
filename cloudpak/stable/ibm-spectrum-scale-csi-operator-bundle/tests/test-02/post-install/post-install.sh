#!/bin/bash
#
# Post-install script REQUIRED ONLY IF additional setup is required post to
# operator install for this test path.
#
set -o errexit
set -o nounset
set -o pipefail

postInstallDir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

# Verify pre-req environment
command -v kubectl > /dev/null 2>&1 || { echo "kubectl pre-req is missing."; exit 1; }
