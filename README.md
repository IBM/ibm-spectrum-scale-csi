# IBM Spectrum Scale CSI Operator

[![Documentation Status](https://readthedocs.org/projects/ibm-spectrum-scale-csi-operator/badge/?version=latest)](https://ibm-spectrum-scale-csi-operator.readthedocs.io/en/latest/?badge=latest)

[![Docker Repository on Quay](https://quay.io/repository/mew2057/ibm-spectrum-scale-csi-operator/status "Docker Repository on Quay")](https://quay.io/repository/mew2057/ibm-spectrum-scale-csi-operator)

An Ansible based operator to run and manage the deployment of the 
[IBM Spectrum Scale CSI Driver](https://github.com/IBM/ibm-spectrum-scale-csi-driver)

This project was originally generated using [operator-sdk](https://github.com/operator-framework/operator-sdk).


## Quick Deploy

To deploy the operator:
```
kubectl create -f deploy/service_account.yaml
kubectl create -f deploy/role.yaml
kubectl create -f deploy/role_binding.yaml
kubectl create -f deploy/crds/csi-scale-operators_v1alpha1_podset_crd.yaml
kubectl create -f deploy/operator.yaml
```

This should start the operator, to then launch the CSI Spectrum Scale plugin the user must then
create a Custom Resource of kind `CSIScaleOperator`.

```
kubectl create -f deploy/spectrum_scale.yaml
```


## Uninstall

To remove the operator:
```
kubectl delete -f deploy/spectrum_scale.yaml
kubectl delete -f deploy/operator.yaml
kubectl delete -f deploy/role.yaml
kubectl delete -f deploy/role_binding.yaml
kubectl delete -f deploy/service_account.yaml
kubectl delete -f deploy/crds/csi-scale-operators_v1alpha1_podset_crd.yaml
```

Please note, this will completely destroy the operator and all associated resources..

## Building the Operator

In order to build the operator the [operator-sdk](https://github.com/operator-framework/operator-sdk) cli,
and docker are required. 
--- TODO This needs to be more precise -- John (9.19.19)

The following must be executed in the root directory of this repository:
--- TODO This needs to be replaced by a `make` command -- John (9.19.19)
```
operator-sdk build csi-scale-operator
```

