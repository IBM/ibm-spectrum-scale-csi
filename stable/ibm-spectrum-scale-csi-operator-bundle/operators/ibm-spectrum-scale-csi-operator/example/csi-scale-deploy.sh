#! /bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd ${DIR}/..

if [[ $1 == "-b" ]]
then
    export GO111MODULE="on"
    #operator-sdk generate k8s
    operator-sdk build csi-scale-operator
    shift

    export REPO="$(hostname -f)/"
    if [ ! -z "$1" ]
    then
        export REPO="$1/"
    fi 

    sed -i "s|REPLACE_IMAGE|${REPO}csi-scale-operator|g" deploy/operator.yaml
    docker tag csi-scale-operator ${REPO}csi-scale-operator
    docker push ${REPO}csi-scale-operator
    echo ${REPO}csi-scale-operator
fi 

kubectl create -f deploy/role.yaml
kubectl create -f deploy/service_account.yaml
kubectl create -f deploy/role_binding.yaml
kubectl create -f deploy/crds/ibm_v1alpha1_csiscaleoperator_crd.yaml
#kubectl create -f deploy/crds/csi-scale-operators_v1alpha1_podset_cr.yaml --save-config
kubectl create -f example/spectrum_scale.yaml

kubectl create -f deploy/operator.yaml

