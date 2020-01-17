
# Manual Build


## Using local golang build environment to build the CSI driver code

	This method involves installation of golang and dep package on local build machine

1. Install the latest version of Go and add it to PATH. Refer https://golang.org/

   ```
   export PATH=$PATH:<go_install_dir>/bin
   ```

2. Set your GOPATH to a directory where you want to clone the repo

   ```
   export GOPATH=<path_to_repo_base>
   ```

3. Clone the code

   ```
   mkdir -p $GOPATH/src/github.com/IBM
   cd $GOPATH/src/github.com/IBM
   git clone https://github.com/IBM/ibm-spectrum-scale-csi-driver.git
   ```

4. Build

     4.1 Get dep (*Go's dependency manager*), which will be used by Makefile:

     ```
     wget --directory-prefix=/tmp/ https://github.com/golang/dep/releases/download/v0.5.1/dep-linux-amd64
     mkdir -p $GOPATH/bin
     mv /tmp/dep-linux-amd64 $GOPATH/bin/dep
     chmod +x $GOPATH/bin/dep
     export PATH=$PATH:$GOPATH/bin
     ```
  
     4.2 Compile:

     ```
     cd $GOPATH/src/github.com/IBM/ibm-spectrum-scale-csi-driver
     make
     ```
  
     4.3 Compile/build the docker image:

     ```
     cd $GOPATH/src/github.com/IBM/ibm-spectrum-scale-csi-driver
     make build-image
     ```

