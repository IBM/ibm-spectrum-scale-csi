#! /bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd ${DIR}/..
set -x

if [[ $1 == "-b" ]]
then
    operator-sdk build . -t ibm-spectrum-scale-csi-operator
    shift

    export REPO="$(hostname -f):5000/"
    if [ ! -z "$1" ]
    then
        export REPO="$1/"
    fi 

    docker tag ibm-spectrum-scale-csi-operator ${REPO}csi-scale-operator:latest
    docker push ${REPO}csi-scale-operator:latest

    #operator-sdk generate k8s
    ansible-playbook hacks/change_deploy_image.yml --extra-vars "quay_operator_endpoint=${REPO}csi-scale-operator:latest"
fi 

kubectl apply -f deploy/namespace.yaml
kubectl apply -f deploy/role.yaml
kubectl apply -f deploy/service_account.yaml
kubectl apply -f deploy/role_binding.yaml
kubectl apply -f deploy/crds/csiscaleoperators.csi.ibm.com.crd.yaml
kubectl apply -f deploy/operator.yaml

