#!/bin/bash

kubectl delete csiscaleoperators --all
kubectl delete -f deploy/operator.yaml
kubectl delete -f deploy/role.yaml
kubectl delete -f deploy/service_account.yaml
kubectl delete -f deploy/role_binding.yaml
kubectl delete -f deploy/crds/ibm-spectrum-scale-csi-operator-crd.yaml
kubectl delete -f deploy/namespace.yaml
