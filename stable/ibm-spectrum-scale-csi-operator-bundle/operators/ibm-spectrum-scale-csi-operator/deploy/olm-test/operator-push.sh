#!/bin/bash
set -x
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd ${DIR}/..

export OPERATOR_DIR=olm-catalog/ibm-spectrum-scale-csi-operator
export QUAY_NAMESPACE=mew2057
export PACKAGE_NAME=ibm-spectrum-scale-csi-operator-app
export PACKAGE_VERSION=0.1.3
export TOKEN=$(curl -sH "Content-Type: application/json" -XPOST https://quay.io/cnr/api/v1/users/login -d '
{"user": {"username": "'"${QUAY_USERNAME}"'","password": "'"${QUAY_PASSWORD}"'"}}' | cut -d'"' -f4)

operator-courier verify --ui_validate_io "$OPERATOR_DIR"

if [ $? -eq 0 ] 
then
  operator-courier push "$OPERATOR_DIR" "$QUAY_NAMESPACE" "$PACKAGE_NAME" "$PACKAGE_VERSION" "$TOKEN"
fi

  

