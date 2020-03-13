
# Manual Build


## Using local golang build environment to build the CSI driver code

	This method involves installation of golang and dep package on local build machine

1. Install the latest version of Go and add it to PATH. Refer https://golang.org/

   ```
   export PATH=$PATH:<go_install_dir>/bin
   ```

2. Set your GOPATH to a directory where you want to clone the repo. This examples uses `/root/go`.

   ```
   export GOPATH="/root/go"
   export IBM_DIR="$GOPATH/src/github.com/IBM"
   ```

3. Clone the code

   ```
   mkdir -p ${IBM_DIR}
   cd ${IBM_DIR}/
   git clone https://github.com/IBM/ibm-spectrum-scale-csi.git
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
  
     4.2 Compile the driver and build the docker images:

     ```
     # IBM_DIR is defined in the previous step
     export DRIVER_DIR="$IBM_DIR/ibm-spectrum-scale-csi/driver"
     cd ${DRIVER_DIR}

     # Compile the driver
     make

     # Build the docker image
     make build-image
     ```