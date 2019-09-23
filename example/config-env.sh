#!/bin/bash
export REPO="$(hostname -f)/"
if [ ! -z "$1" ]
then
    export REPO="$1/"
fi

git config --global url."git@github.ibm.com:".insteadOf "https://github.ibm.com/"
#eval $(minikube docker-env)
operator-sdk build csi-scale-operator
docker tag csi-scale-operator ${REPO}csi-scale-operator
docker push ${REPO}csi-scale-operator

cd ../../FSaaS/csi-scale
make build-image
cd -
docker tag "sys-scale-containers-csi-docker-local.artifactory.swg-devops.com/csi-spectrum-scale:v1.0.0" ${REPO}csi-scale
docker push ${REPO}csi-scale

