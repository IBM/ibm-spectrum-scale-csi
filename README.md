[![Build Status](https://travis.ibm.com/FSaaS/csi-gpfs.svg?token=sfEsUpvxtZ9kpqpJBFp8&branch=master)](https://travis.ibm.com/FSaaS/csi-gpfs)

# CSI Plugin for GPFS

## Development

* Set your GOPATH
```
export $GOPATH=/path/to/src
```

* Clone
```
mkdir -p $GOPATH/github.ibm.com/FSaaS
cd $GOPATH/github.ibm.com/FSaaS
git clone git@github.ibm.com:FSaaS/csi-gpfs.git
```

* Build
  * Dep (within Makefile) will try to use https. Enforce ssh by running:
  ```
  git config --global url."git@github.ibm.com:".insteadOf "https://github.ibm.com/"
  ```
  Source: https://github.com/golang/dep/blob/master/docs/FAQ.md
  
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
    * To deploy the plugin along with the external-attacher, external-provisioner, and set the RBAC settings run:
    ```
    ./deploy/create.sh
    ```
  * Delete
  ```
  ./deploy/delete.sh
  ```
