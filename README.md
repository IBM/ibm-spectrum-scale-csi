[![Build Status](https://travis.ibm.com/FSaaS/csi-scale.svg?token=sfEsUpvxtZ9kpqpJBFp8&branch=master)](https://travis.ibm.com/FSaaS/csi-scale)

# CSI Plugin for Scale

## Development

* Set your GOPATH
  ```
  export $GOPATH=/path/to/go-tree
  ```

* Clone
  ```
  mkdir -p $GOPATH/src/github.ibm.com/FSaaS
  cd $GOPATH/src/github.ibm.com/FSaaS
  git clone git@github.ibm.com:FSaaS/csi-scale.git
  ```

* Build
  * Get dep (Go's dependency manager) that is used by Makefile.  Dep will try to use https, do enforce ssh by running:
    ```
    git config --global url."git@github.ibm.com:".insteadOf "https://github.ibm.com/"
    ```
    Source: https://github.com/golang/dep/blob/master/docs/FAQ.md

    Install dep; there an install script you can run directly from the internet, but "just say no".  Instead, try something like this:
    ```
    % wget --directory-prefix=/tmp/ https://github.com/golang/dep/releases/download/v0.5.1/dep-linux-amd64
    % mkdir -p $GOPATH/bin
    % mv /tmp/dep-linux-amd64 $GOPATH/bin/dep
    % chmod +x $GOPATH/bin/dep
    % ls $GOPATH/bin/dep
    -rwxr-xr-x. 1 11052608 Mar 11 00:17 /home/ota/csi-scale/bin/dep*
    % export PATH=$PATH:$GOPATH/bin
    ```

  _( Which of these to do?  Every `make` appears to rebuild this large binary, which is very slow. )_
  * To compile:
    ```
    make
    ```
  * To compile/build image:
    ```
    make build-image
    ```
  * To compile/build/push image:
    ```
    make push-image
    ```
  
## Deployment

  * Create
    * Deploy the plugin along with the external-attacher, external-provisioner, configuration (Scale API, block devices), and RBAC:
    _( the **paths have changed**... )_
    ```
    ./deploy/create.sh
    ```
  * Delete
    ```
    ./deploy/destroy.sh
    ```

## Example usage
* First, deploy scale-image (see other repo). In particular, the API server must have started.
* Create
  * Create a Scale PVC - provision a Scale PV
    _( the **paths have changed**... )_
  ```
  ./examples/create.sh
  ```
  * Deploy POD using created PVC
  ```
  kubectl create -f ./examples/pod.yaml
  ```
  * Deploy additional POD using created PVC
  ```
  kubectl create -f ./examples/pod2.yaml
  ```
* Delete
  * Delete PODs
  ```
  kubectl delete -f ./examples/pod.yaml
  kubectl delete -f ./examples/pod2.yaml
  ```
  * Delete PVC (deprovision Scale PV)
  ```
  ./examples/destroy.sh
  ```
