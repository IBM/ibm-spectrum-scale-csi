#!/bin/bash
#==================================================================================================
# Get jq, operator-courier, and helm
which jq > /dev/null
if [ $? -ne 0 ]
then
  wget /usr/bin/jq https://github.com/stedolan/jq/releases/download/jq-1.6/jq-linux64 
  chmod +x jq-linux64
  mv jq-linux64 /usr/bin/jq
fi

helm registry -h >/dev/null
if [ $? -ne 0 ]
then
  curl -L https://git.io/get_helm.sh | bash
  mkdi r-p ~/.helm/plugins/
  cd ~/.helm/plugins/ && git clone https://github.com/app-registry/appr-helm-plugin.git registry
fi

operator-courier -h >>/dev/null
if [ $? -ne  0 ]
then
  pip3 install operator-courier
fi
#==================================================================================================
# This script takes OPERATOR_DIR, QUAY_NAMESPACE, and PACKAGE_NAME as input to the script.

# Set values
export OPERATOR_DIR=${OPERATOR_DIR:-operator/deploy/olm-catalog/ibm-spectrum-scale-csi-operator}
export QUAY_NAMESPACE=${QUAY_NAMESPACE:-ibm-spectrum-scale-dev}
export PACKAGE_NAME=${PACKAGE_NAME:-ibm-spectrum-scale-csi-app}
export PACKAGE_VERSION=$(helm registry list quay.io -o ${QUAY_NAMESPACE} --output json | jq --arg NAME "${QUAY_NAMESPACE}/${PACKAGE_NAME}" '.[] | select(.name == $NAME) |.default' | awk -F'[ .\"]' '{print $2"."$3"."$4+1""}')
export TOKEN=$(curl -sH "Content-Type: application/json" -XPOST https://quay.io/cnr/api/v1/users/login -d '
{"user": {"username": "'"${QUAY_USERNAME}"'","password": "'"${QUAY_PASSWORD}"'"}}' | cut -d'"' -f4)
export PACKAGE_FILE=${OPERATOR_DIR}/ibm-spectrum-scale-csi-operator.package.yaml

# Rename  the package 
sed -i  "s/packageName: ibm-spectrum-scale-csi-operator/packageName: ${PACKAGE_NAME}/g" ${PACKAGE_FILE}

operator-courier push "$OPERATOR_DIR" "$QUAY_NAMESPACE" "$PACKAGE_NAME" "$PACKAGE_VERSION" "$TOKEN"

# Reset the package.
sed -i  "s/packageName: ${PACKAGE_NAME}/packageName: ibm-spectrum-scale-csi-operator/g" ${PACKAGE_FILE}
