#! /bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd ${DIR}/..

if [[ $1 == "-b" ]]
then
    operator-sdk build csi-scale-operator
    shift
    export REPO="$(hostname -f)/"
    if [ ! -z "$1" ]
    then
        export REPO="$1/"
    fi 

    docker tag csi-scale-operator ${REPO}csi-scale-operator
    docker push ${REPO}csi-scale-operator
fi 

# This is the user config
kubectl create -f deploy/spectrum_scale.yaml

kubectl create -f deploy/service_account.yaml
kubectl create -f deploy/role.yaml
kubectl create -f deploy/role_binding.yaml
kubectl create -f deploy/crds/cache_v1alpha1_podset_crd.yaml
kubectl create -f deploy/operator.yaml
kubectl create -f deploy/crds/cache_v1alpha1_podset_cr.yaml


