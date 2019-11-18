# ADVISORY - Please rebase instead of merge all branches created before 11/12/19

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
export OPERATOR_DIR="$IBM_DIR/ibm-spectrum-scale-csi-operator"

# Ensure the dir is present then clone.
mkdir -p ${IBM_DIR}
cd ${IBM_DIR}
git clone https://github.com/IBM/ibm-spectrum-scale-csi-operator.git
```

### Development environment setup

The development environment dependencies are managed using an ansible playbook for the IBM Spectrum Scale CSI Operator. If ansible is installed in your environment simply run the following command:

``` bash
ansible-playbook $GOPATH/src/github.com/IBM/ibm-spectrum-scale-csi-operator/ansible/dev-env-playbook.yaml
```

This script will do the following:
1. Install `python3`
2. Install `python3` requirements (`sphinx`, `operator-courier`, `docker`)
3. Install `operator-sdk`
4. Ensure `go-1.13` is installed.


### Building the image

To build the image the user must navigate to the operator directory (This directory structure is an artifact of the IBM Cloud Pak certification process). 

``` bash
cd stable/ibm-spectrum-scale-csi-operator-bundle/operators/ibm-spectrum-scale-csi-operator
export GO111MODULE="on"
operator-sdk build csi-scale-operator

docker tag csi-scale-operator quay.io/mew2057/ibm-spectrum-scale-csi-operator:v0.9.1
```

>**NOTE** This requires `docker`.

### Using the image
>**NOTE** If you're using the quay image, this step can be skipped.

In order to use the image in your environment you will need to push the image to a [docker registry](https://docs.docker.com/registry/). You may setup your own image, or push to a repository such  as [quay.io](quay.io).

Deploying your own registry is an [involved process](https://docs.docker.com/registry/deploying/), and outside of the scope of this document. 

If you're using quay, we recommend doing the [Quay Tutorial](https://quay.io/tutorial/).


Once you have a repository ready and you've logged you can tag and push your image:
``` bash
docker tag csi-scale-operator <your-repo>/ibm-spectrum-scale-csi-operator:v0.9.1
docker push <your-repo>/ibm-spectrum-scale-csi-operator:v0.9.1

# This will update your deployment to point at your image.
hacks/change_deploy_image.py -i <your-repo>/ibm-spectrum-scale-csi-operator:v0.9.1
```

## Deploying the Operator

>**WARNING** If you are using your own image you must, complete (#using-the-image)!

### Option A: Manually

If you've built the image as outlined above and tagged it, you can easily run the following to deploy the operator manually:

``` bash
kubectl apply -f deploy/namespace.yaml
kubectl apply -f deploy/service_account.yaml
kubectl apply -f deploy/role.yaml
kubectl apply -f deploy/role_binding.yaml
kubectl apply -f deploy/crds/ibm_v1alpha1_csiscaleoperator_crd.yaml
kubectl apply -f deploy/operator.yaml
```

At this point the operator is running and ready for use!

### Option B: Using OLM

> **NOTE** : This will be the prefered method, however, work is ongoing.

The following will subsrcibe the [quay.io](quay.io) version of the operator assuming OLM is installed.

``` bash
kubectl apply -f deploy/olm-test/operator-source.yaml
```

## Starting the CSI Driver

Once the operator is running the user needs to access the operator's API and request a deployment. This is done through
use of the `CSIScaleOperator` Custom Resource. This resource will be tuned to your environment. A sample of the file is given:

``` YAML
# spectrum_scale.yaml

apiVersion: scale.ibm.com/v1alpha1
kind: 'CSIScaleOperator'
metadata:
    name: 'csi-scale-operator'
status: {}
spec:
  # Optional
  # ----
  # Attacher image for csi (actually attaches to the storage).
  attacher: "quay.io/k8scsi/csi-attacher:v1.0.0"
  
  # Provisioner image for csi (actually issues provision requests).
  provisioner:"quay.io/k8scsi/csi-provisioner:v1.0.0"
  
  # Sidecar container image for the csi spectrum scale plugin pods.
  driverRegistrar: "quay.io/k8scsi/csi-node-driver-registrar:v1.0.1"
  
  # Image name for the csi spectrum scale plugin container.
  spectrumScale: "quay.io/mew2057/ibm-spectrum-scale-csi-driver:v0.9.0"

  # Node selector for attacher sidecar, can have multiple key value.
  attacherNodeSelector:
    - key: "scale"
      value: "true"
    - key: "infranode"
      value: "2"

  # Node selector for provisioner sidecar, can have multiple key value.
  provisionerNodeSelector:
    - key: "scale"
      value: "true"
    - key: "infranode"
      value: "2"

  # Node selector for SpectrumScale CSI Plugin, can have multiple key value.
  pluginNodeSelector:
    - key: "scale"
      value: "true"

  # Node mapping between K8s node and SpectrumScale node, can have multiple
  # values.
  nodeMapping:
    - k8sNode: "node1"
      spectrumscaleNode: "scaleNode1"

  # ----
  
  # Required
  # ----
  # The path to the gpfs file system mounted on the host machine.
  scaleHostpath: "/ibm/fs1"

  # A collection of gpfs cluster properties for the csi driver to mount.
  clusters:
    # The cluster id of the gpfs cluster specified (mandatory).
    - id: "2120508922778391120"
      
      # A string specifying a secret resource name.
      secrets: "secret1"
      
      # Require a secure SSL connection to connect to GPFS.
      secureSslMode: false
      
      # A string specifying a cacert resource name.
      # cacert: <>
      
      # The primary file system for the GPFS cluster
      primary:
        # The name of the primary filesystem.
        primaryFs: "fs1"
        # The name of the primary fileset, created in primaryFS.
        primaryFset: "csiFset2"
        # Inode Limit for Primary Fileset
        inodeLimit: "1024"
        # Remote cluster ID
        remoteCluster: "2120508922778391121"
        # Filesystem name on remote cluster.
        remoteFs: "gpfs2"
        
      # A collection of targets for REST calls.
      restApi:
        # The hostname of the REST server.
        - guiHost: "GUI_HOST"
        
          # The port number running the REST server.
          # guiPort
        
  # ----

```
> **NOTE** : Work is ongoing to reduce the amount end users need to populate.

Before starting the pluging be sure to add any secrets to the appropriate namespace, the default
namespace is `ibm-spectrum-scale-csi-driver`:

``` bash
kubectl apply -f secrets.yaml -n ibm-spectrum-scale-csi-driver
```

> **ATTENTION** : If the driver pod doesn't start, it's generally because the secrets haven't been created.


To acutally start the CSI Plugin run the following command

``` bash
kubectl apply -f spectrum_scale.yaml
```

To stop the CSI plugin you can run:

``` bash
kubectl delete -f spectrum_scale.yaml
```

## Uninstalling the CSI Operator

To remove the operator:
``` bash
kubectl delete -f deploy/spectrum_scale.yaml
kubectl delete -f deploy/operator.yaml
kubectl delete -f deploy/role.yaml
kubectl delete -f deploy/role_binding.yaml
kubectl delete -f deploy/service_account.yaml
kubectl delete -f deploy/crds/ibm_v1alpha1_csiscaleoperator_crd.yaml
kubectl delete -f deploy/namespace.yaml
```

Please note, this will completely destroy the operator and all associated resources.


