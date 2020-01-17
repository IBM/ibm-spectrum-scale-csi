# IBM Spectrum Scale CSI Operator

[![Documentation Status](https://readthedocs.org/projects/ibm-spectrum-scale-csi-operator/badge/?version=latest)](https://ibm-spectrum-scale-csi-operator.readthedocs.io/en/latest/?badge=latest)

An Ansible based operator to run and manage the deployment of the 
[IBM Spectrum Scale CSI Driver](https://github.com/IBM/ibm-spectrum-scale-csi-driver)

This project was originally generated using [operator-sdk](https://github.com/operator-framework/operator-sdk).

> **WARNING**: This repository undergoing active development! If you encounter issues with the following instructions, [_please open an issue_](https://github.com/IBM/ibm-spectrum-scale-csi-operator/issues).

## Setup from scratch

### Cloning the repository

> **WARNING**: This repository needs to be accessible in your `GOPATH`. In testing, the root user was used and set to: `GOPATH=/root/go`.

> **NOTE**: Due to current constraints in golang (relative paths are not supported in golang), you **_MUST_** clone this repository under your gopath. If not, the `operator-sdk` build operation will fail.

``` bash
# Set up some helpful variables
export GOPATH="/root/go"
export IBM_DIR="$GOPATH/src/github.com/IBM"

# Ensure the dir is present then clone.
mkdir -p ${IBM_DIR}
cd ${IBM_DIR}
git clone https://github.com/IBM/ibm-spectrum-scale-csi.git
```

### Development environment setup

To help configure and resolve dependencies to build the csi-operator, a ansible playbook is provided.  You can run the following to invoke the playbook:

``` bash
ansible-playbook $GOPATH/src/github.com/IBM/ibm-spectrum-scale-csi/tools/ansible/dev-env-playbook.yaml
```

### Building the image

To build the image the user must navigate to the operator directory (This directory structure is an artifact of the IBM Cloud Pak certification process). 

``` bash
# IBM_DIR is defined in the previous step
export OPERATOR_DIR="$IBM_DIR/ibm-spectrum-scale-csi/operator"
cd ${OPERATOR_DIR}

export GO111MODULE="on"
operator-sdk build csi-scale-operator
```

>**NOTE** This requires `docker`.

### Using the image

In order to use the images that you just built, the image needs to be pushed to some container repository.

* **Quay.io (recommended)**

  Follow this tutorial to configure [quay.io](https://quay.io/tutorial/) and then create a repository named: `ibm-spectrum-scale-csi-operator`.

* **Docker** 

  Deploying your own Docker registry is an [involved process](https://docs.docker.com/registry/deploying/), and outside of the scope of this document. 

The documentation will assume that the quay.io path is being used. 

Once you have a repository ready:

``` bash
# Authenticate to quay.io
docker login <credentials> quay.io

# Tag the build 
docker tag csi-scale-operator quay.io/<your-user>/ibm-spectrum-scale-csi-operator:v0.9.1

# push the image
docker push quay.io/<your-user>/ibm-spectrum-scale-csi-operator:v0.9.1

# Update your deployment to point at your image.
hacks/change_deploy_image.py -i quay.io/<your-user>/ibm-spectrum-scale-csi-operator:v0.9.1
```

## Deploying the Operator

> **WARNING** If you are using your own image you must, complete [using the image](#using-the-image)!

### Option A: Manually

If you've built the image as outlined above and tagged it, you can easily run the following to deploy the operator manually, for openshift use "oc" instead of "kubectl"

``` bash
cd ${OPERATOR_DIR}/

kubectl apply -f deploy/namespace.yaml
kubectl apply -f deploy/service_account.yaml
kubectl apply -f deploy/role.yaml
kubectl apply -f deploy/role_binding.yaml
kubectl apply -f deploy/crds/ibm-spectrum-scale-csi-operator-crd.yaml
kubectl apply -f deploy/operator.yaml
```


> **NOTE**: Kubernetes uses `kubectl` the command, replace with `oc` if deploying in OpenShift.

At this point the operator is running and ready for use!

### Option B: Using Operator Lifecycle Manager (OLM)

> **NOTE**: This will be the prefered method.  However, work is ongoing.


> **NOTE**: Installing OLM is out of the scope of this document, please refer to [the official documentation](https://github.com/operator-framework/operator-lifecycle-manager/blob/master/doc/install/install.md). If you're still having trouble, [this guide goes even deeper](https://github.com/operator-framework/community-operators/blob/master/docs/testing-operators.md).

The following will subscribe the [quay.io](quay.io) version of the operator assuming OLM is installed.

``` bash
cd ${OPERATOR_DIR}/

kubectl apply -f deploy/olm-scripts/operator-source.yaml
```
> **NOTE**: Kubernetes use `kubectl` command, replace with `oc` if deploying in OpenShift.
```
cd ${OPERATOR_DIR}/

oc apply -f deploy/olm-scripts/operator-source-oc.yaml
```


## Starting the CSI Driver

Once the operator is running the user needs to access the operator's API and request a deployment. This is done through use of the `CSIScaleOperator` Custom Resource. 

> **ATTENTION** : If the driver pod does not start, it is generally due to missing secrets. 

Before starting the plugin, add any secrets to the appropriate namespace.  The Spectrum Scale namespace is `ibm-spectrum-scale-csi-driver`:

``` bash
kubectl apply -f secrets.yaml -n ibm-spectrum-scale-csi-driver
```

A sample of the file is provided [deploy/crds/ibm-spectrum-scale-csi-operator-cr.yaml](stable/ibm-spectrum-scale-csi-operator-bundle/operators/ibm-spectrum-scale-csi-operator/deploy/crds/ibm-spectrum-scale-csi-operator-cr.yaml). 

Modify this file to match the properties in your environment, then:

  * To start the CSI plugin, run: `kubectl apply -f deploy/crds/ibm-spectrum-scale-csi-operator-cr.yaml` 
  * To stop the CSI plugin, run: `kubectl delete -f deploy/crds/ibm-spectrum-scale-csi-operator-cr.yaml` 

## Uninstalling the CSI Operator

To remove the operator:

``` bash
kubectl delete -f deploy/crds/ibm-spectrum-scale-csi-operator-cr.yaml
kubectl delete -f deploy/operator.yaml
kubectl delete -f deploy/role.yaml
kubectl delete -f deploy/role_binding.yaml
kubectl delete -f deploy/service_account.yaml
kubectl delete -f deploy/crds/ibm-spectrum-scale-csi-operator-crd.yaml
kubectl delete -f deploy/namespace.yaml
```

> **NOTE**: Kubernetes use `kubectl` command, replace with `oc` if deploying in OpenShift.

This will completely destroy the operator and all associated resources.


### Open Shift Considerations

When uninstalling on OpenShift the operator creates a `SecurityContextConstraint`  named `csiaccess`.
This allows the driver to mount files in non default namespaces. 

To verify the `SecurityContextConstraint` is gone:

``` bash
kubectl get SecurityContextConstraints csiaccess

# If you get a result:
kubectl delete SecurityContextConstraints csiaccess
```

### Stuck Operator
In cases where deleting the operator `Custom Resource` fails the following recipe can be executed:

``` bash
# You need the proxy ro be running for this command.
kubectl proxy &
# This may need to be customized in OLM environments:
NAMESPACE=ibm-spectrum-scale-csi-driver
kubectl get csiscaleoperators -n ${NAMESPACE} -o json | jq '.spec = {"finalizers":[]}' >temp.json
curl -k -H "Content-Type: application/json" -X PUT --data-binary @temp.json 127.0.0.1:8001/api/v1/namespaces/$NAMESPACE/finalize
rm -f temp.json
```

Typically this happens when deleting the `Custom Resource Definition` before removing all of the `Custom Resources`.
For more details on this check the following [GitHub Issue](https://github.com/operator-framework/operator-sdk/issues/2094).

> **NOTE**: If the operator stops processing CR CRUD after applying this fix it's recommended that the user restart the operator pod.

To restart the operator pod, the following process must be followed:

``` bash
POD_NAME="ibm-spectrum-scale-csi-driver ibm-spectrum-scale-csi-operator-"
NAMESPACE=ibm-spectrum-scale-csi-driver
kubectl delete -n $NAMESPACE $POD_NAME
```
