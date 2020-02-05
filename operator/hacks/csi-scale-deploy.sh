#! /bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd ${DIR}/..
set -x

if [[ $1 == "-b" ]]
then
    docker build . -t ibm-spectrum-scale-csi-operator
    shift

    export REPO="$(hostname -f):5000/"
    if [ ! -z "$1" ]
    then
        export REPO="$1/"
    fi 

    docker tag ibm-spectrum-scale-csi-operator ${REPO}csi-scale-operator
    docker push ${REPO}csi-scale-operator:latest

    #operator-sdk generate k8s
    hacks/change_deploy_image.py -i ${REPO}csi-scale-operator:latest
fi 

kubectl apply -f deploy/namespace.yaml
kubectl apply -f deploy/role.yaml
kubectl apply -f deploy/service_account.yaml
kubectl apply -f deploy/role_binding.yaml
kubectl apply -f deploy/crds/ibm-spectrum-scale-csi-operator-crd.yaml
kubectl apply -f deploy/operator.yaml

