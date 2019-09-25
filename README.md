
   * [Welcome to the public Beta of IBM Spectrum Scale Container Storage Interface (CSI) Driver](#welcome-to-the-public-beta-of-ibm-spectrum-scale-container-storage-interface-csi-driver)
      * [IBM Spectrum Scale Introduction](#ibm-spectrum-scale-introduction)
      * [IBM Spectrum Scale Container Storage Interface (CSI) driver](#ibm-spectrum-scale-container-storage-interface-csi-driver)
         * [Supported Features of the CSI driver](#supported-features-of-the-csi-driver)
         * [Limitations of the CSI driver](#limitations-of-the-csi-driver)
         * [Pre-requisites for installing and running the CSI driver](#pre-requisites-for-installing-and-running-the-csi-driver)
      * [Building the docker image](#building-the-docker-image)
      * [Install and Deploy the Spectrum Scale CSI driver](#install-and-deploy-the-spectrum-scale-csi-driver)
      * [Static Provisioning](#static-provisioning)
      * [Dynamic Provisioning](#dynamic-provisioning)
         * [Storageclass](#storageclass)
      * [Advanced Configuration](#advanced-configuration)
         * [Remote mount support](#remote-mount-support)
         * [Node Selector](#node-selector)
         * [Kubernetes node to Spectrum Scale node mapping](#kubernetes-node-to-spectrum-scale-node-mapping)
      * [Cleanup](#cleanup)
      * [Environments in Test](TESTCONFIG.md#environments-in-test)
      * [Example Hardware Configs](TESTCONFIG.md#example-hardware-configs)
      * [Example of using the Install Toolkit to build a Spectrum Scale cluster for testing the CSI driver](TESTCONFIG.md#example-of-using-the-install-toolkit-to-build-a-spectrum-scale-cluster-for-testing-the-csi-driver)
      * [Links](#links)

  

# Welcome to the public Beta of IBM Spectrum Scale Container Storage Interface (CSI) Driver

DISCLAIMER: This Beta driver is provided as is, without warranty. Any issue will be handled on a best-effort basis. See the Spectrum Scale Users Group links at the very bottom for a community to share and discuss test efforts.

  
## IBM Spectrum Scale Introduction

IBM Spectrum Scale is a clustered file system providing concurrent access to a single file system or set of file systems from multiple nodes. The nodes can be SAN attached, network attached, a mixture of SAN attached and network attached, or in a shared nothing cluster configuration. This enables high performance access to this common set of data to support a scale-out solution or to provide a high availability platform.

IBM Spectrum Scale has many features beyond common data access including data replication, policy based storage management, and multi-site operations. You can create a cluster of AIX® nodes, Linux nodes, Windows server nodes, or a mix of all three. IBM Spectrum Scale can run on virtualized instances providing common data access in environments, leverage logical partitioning, or other hypervisors. Multiple IBM Spectrum Scale clusters can share data within a location or across wide area network (WAN) connections. For more information on IBM Spectrum Scale features, see the Product overview section in the IBM Spectrum Scale: Concepts, Planning, and Installation Guide.

Please refer to the [IBM Spectrum Scale Knowledge Center](https://www.ibm.com/support/knowledgecenter/en/STXKQY/ibmspectrumscale_welcome.html) for more information.
  

## IBM Spectrum Scale Container Storage Interface (CSI) driver

The IBM Spectrum Scale Container Storage Interface (CSI) driver allows IBM Spectrum Scale to be used as persistent storage for stateful application running in Kubernetes clusters. Through this CSI Driver, Kubernetes persistent volumes (PVs) can be provisioned from IBM Spectrum Scale. Thus, containers can be used with stateful microservices, such as database applications (MongoDB, PostgreSQL etc), web servers (nginx, apache), or any number of other containerized applications needing provisioned storage.

### Supported Features of the CSI driver

IBM Spectrum Scale Container Storage Interface (CSI) driver supports the following features:

- **Static provisioning:** Ability to use existing directories as persistent volumes
- **Lightweight dynamic provisioning:** Ability to create directory-based volumes dynamically
- **Fileset-based dynamic provisioning:** Ability to create fileset-based volumes dynamically
- **Multiple file systems support:** Volumes can be created across multiple file systems
- **Remote mount support:** Volumes can be created on a remotely mounted file system
  
### Limitations of the CSI driver

The IBM Spectrum Scale Container Storage Interface (CSI) driver has the following limitations:

- The size specified in PersistentVolumeClaim for lightweight volume and dependent fileset volume, is not honored.
- Volumes cannot be mounted in read-only mode.
- Maximum number of supported volumes that can be created using independent fileset storage class is 1000. This is based upon the [fileset maximums for IBM Spectrum Scale](https://www.ibm.com/support/knowledgecenter/STXKQY/gpfsclustersfaq.html#filesets)
- The IBM Spectrum Scale GUI server is relied upon for performing file system and cluster operations. If the GUI password or CA certificate expires, manual intervention is needed by the admin to reset the GUI password or generate a new certificate and update the configuration of the CSI driver. In this case, a restart of the CSI driver will be necessary.
- Rest API status, used by the CSI driver, may lag from actual state, causing PVC mount or unmount failures.
- Although multiple instances of the Spectrum Scale GUI are allowed, the CSI driver is currently limited to point to a single GUI node.
- OpenShift 4.x has not been fully qualified with the IBM Spectrum Scale CSI driver. Additionally, [Red Hat considers the Container Storage Interface as a Technology Preview feature, within OpenShift 4.1.](https://docs.openshift.com/container-platform/4.1/storage/persistent-storage/persistent-storage-csi.html)
- CRI-O, within OpenShift 4.x, currently issues an SELinux relabel upon mount of a PVC. This will fail if the underlying Spectrum Scale directory structure contains a .snapshots directory. Disable SELinux within CRI-O to avoid this (*see pre-requisites below*).

### Pre-requisites for installing and running the CSI driver

- IBM Spectrum Scale / Kubernetes overlap should be as follows

  | Node Type | Spectrum Scale | Kubernetes |
  |--|--|--|
  | **Master node(s)** | do not install | required |
  | **Worker node(s)** | required | required |
  | **GUI node** | required | do not install |
  | **NSD node** | required | optional |

- Red Hat 7.6 (**kernel 3.10.0-957 or higher**) on Spectrum Scale nodes

- IBM Spectrum Scale version 5.0.3.3 is installed.

- An IBM Spectrum Scale GUI is up and running on a Spectrum Scale node and a user is created and part of the `CsiAdmin` group

  ```
  /usr/lpp/mmfs/gui/cli/mkuser <__username__> -p <__password__> -g CsiAdmin
  ```

- Kubernetes ver 1.13+ cluster is created

- If using OpenShift (or CRI-O), ensure that ver 4.1+ is installed and SELinux is disabled in CRI-O

  a) On each worker node, edit `/etc/crio/crio.conf` and disable selinux
     ```
     selinux = false
     ```

  b) Reload and restart CRI-O 
     ```
     systemctl daemon-reload
     systemctl restart crio
     ```
  
- All Kubernetes worker nodes must also be Spectrum Scale client nodes. Install the Spectrum Scale client on all Kubernetes worker nodes and ensure they are added to the Spectrum Scale cluster. (To install Spectrum Scale and CSI driver only on selected nodes, perform the steps from [Node Selector](#node-selector)

- The Filesystem to be used for persistent storage must be mounted on the Spectrum Scale GUI node as well as all Kubernetes worker nodes. (*If multiple filesystems are to be used as persistent storage for containers, then all need to be mounted*)

- Quota must be enabled on the filesystem (*required for fileset based dynamic provisioning*)
  ```
  mmchfs <__filesystem_name__> -Q yes
  ```

## Building the docker image

1. Install the latest version of Go. Refer https://golang.org/
   Add Go to PATH

   ```
   export PATH=$PATH:/go/install/dir/bin
   ```

2. Set your GOPATH to a directory where you want to clone the repo

   ```
   export GOPATH=/path/to/repo/base
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

     4.4 Compile/build/save the docker image:

     ```
     cd $GOPATH/src/github.com/IBM/ibm-spectrum-scale-csi-driver
     make save-image
     ```

      A tar file of docker image will be stored under the _output directory.

## Install and Deploy the Spectrum Scale CSI driver

1. Load the docker image to all Kubernetes worker nodes

   ```
   docker image load -i csi-spectrum-scale_v0.9.0.tar
   ```

   *On OpenShift, use this command instead:*

   ```
   podman image load -i csi-spectrum-scale_v0.9.0.tar
   ```

2. Update `deploy/spectrum-scale-driver.conf` with your cluster and environment details.

3. Set the environment variable CSI_SCALE_PATH to `<repo_base_path>/ibm-spectrum-scale-csi-driver`

   ```
   export CSI_SCALE_PATH=<repo_base_path>/ibm-spectrum-scale-csi-driver
   ```

4. Run the install helper script:

   ```
   tools/spectrum-scale-driver.py $CSI_SCALE_PATH/deploy/spectrum-scale-driver.conf
   ```

   Review the generated configuration files in deploy.

5. Run the `deploy/create.sh` script to deploy the plugin

6. Check that the csi pods are up and running

   ```
   % kubectl get pod
   NAME READY STATUS RESTARTS AGE
   csi-spectrum-scale-7d8jg 2/2 Running 0 7s
   csi-spectrum-scale-attacher-0 1/1 Running 0 8s
   csi-spectrum-scale-provisioner-0 1/1 Running 0 8s
   ```


## Static Provisioning

In static provisioning, the backend storage volumes and PVs are created by the administrator. Static provisioning can be used to provision a directory or fileset with existing data.

For static provisioning of existing directories perform the following steps:

- Generate static pv yaml file using helper script

   ```
   tools/generate_pv_yaml.sh --filesystem rgpfs2 --size 10 \
   --linkpath /ibm/rgpfs2/static-pv-from-vmi-146/static-pv-1 --pvname static-pv
   ```

- Use sample static_pvc and pod files for sanity test under `examples/static`

   ```
   kubectl apply -f examples/static/static_pv.yaml
   kubectl apply -f examples/static/static_pvc.yaml
   kubectl apply -f examples/static/static_pod.yaml
   ```
  

## Dynamic Provisioning

Dynamic provisioning is used to dynamically provision the storage backend volume based on the storageclass.

### Storageclass
Storageclass defines what type of backend volume should be created by dynamic provisioning. IBM Spectrum Scale CSI driver supports creation of directory based (also known as lightweight volumes) and fileset based (independent as well as dependent) volumes. Following parameters are supported by BM Spectrum Scale CSI driver storageclass:

 - **volBackendFs**: Filesystem on which the volume should be created. This is a mandatory parameter.
 - **clusterId**: Cluster ID on which the volume should be created. 
 - **volDirBasePath**: Base directory path relative to the filesystem mount point under which directory based volumes should be created. If specified, the storageclass is used for directory based (lightweight) volume creation. If not specified, storageclass creates fileset based volumes.
 - **uid**: UID with which the volume should be created. Optional
 - **gid**: UID with which the volume should be created. Optional
 - **fileset-type**: Type of fileset. Valid values are "independent" or "dependent". Default is "independent". 
 - **parentFileset**: Specifies the parent fileset under which dependent fileset should be created. Mandatory if "fileset-type" is specified.
 - **inode-limit**: Inode limit for fileset based volumes. If not specified, default Spectrum Scale inode limit of 1million is used.
 
For dynamic provisioning, use sample storageclass, pvc and pod files for sanity test under examples/dynamic

Example:

   ```
   kubectl apply -f examples/dynamic/fileset/storageclassfileset.yaml
   kubectl apply -f examples/dynamic/fileset/pvcfset.yaml
   kubectl apply -f examples/dynamic/fileset/podfset.yaml
   ```

## Advanced Configuration

Following is advanced configuration of IBM Spectrum Scale CSI driver and is not supported through the installer "spectrum-scale-driver.py".

### Remote mount support

IBM Spectrum Scale provides a feature to mount a Spectrum Scale file system that belongs to another IBM Spectrum Scale cluster. Consider the case where Kubernetes worker nodes are part of a "primary" Spectrum Scale cluster. This primary cluster has filesystems mounted from a "remote" Spectrum Scale cluster. 

In order to deploy CSI driver on such a configuration, following steps should be performed after running the installer "spectrum-scale-driver.py":

- Update `deploy/spectrum-scale-config.json` file with remote cluster and filesystem name information under "primary" section by adding the two parameters as:

   * "**remoteCluster**":"<remote cluster ID>",  
   * "**remoteFS**":"remote filesystem name" (Required only if remote filesystem name is different than the locally mounted filesystem name)

- Make another entry for this cluster under the "clusters" section of `deploy/spectrum-scale-config.json` as:

   ```
   {"id":"2954738785946888888",
    "secrets":"secret2",
     "restApi": [
        {"guiHost":"172.16.1.33"
        }
     ]
   }
   ```

- Ensure that a new entry is created in secrets list in `deploy/spectrum-scale-secret.json` for "secret2" as:

   ```
   {
      "kind": "Secret",
      "apiVersion": "v1",
      "metadata": {
         "name": "secret2"
      },
      "data": {
         "username": "YWRtaW4=",
         "password": "MWYyZDFlMmU2N2Rm"
      }
   }
   ```
   **Note:** username and passoword are base64 encoded.

- Add an entry for secret2 in deploy/csi-plugin.yaml file under "volumes":

   ```
   - name: secret2
     secret:
       secretName: secret2
   ```

   Add corresponding entry under "containers -> csi-spectrum-scale -> volumeMounts" section:

   ```
   - name: secret2
     mountPath: /var/lib/ibm/secret2
     readOnly: true
   ```

- Deploy the driver by running `deploy/create.sh`

- For lightweight dynamic provisioning, no change in storageclass is needed.

- For fileset based dynamic provisioning, use the storageclass parameters as below:

   * **volBackendFs**: Filesystem on which the volume should be created. Use the remote cluster filesystem name here.
   * **clusterId**: Remote Cluster ID on which the volume (fileset) should be created. 
   * **localFs**: Name of the locally mounted filesystem. This is required only if the local name and remote filesystem names are different.
	Rest of the storageclass parameteres remain valid.

### Node Selector

Node selector is used to control on which Kubernetes worker nodes the IBM Spectrum Scale CSI driver should be running. Node selector also helps in cases where new worker nodes are added to Kubernetes cluster but does not have IBM Spectrum Scale installed, in this case we would not want the CSI driver to be deployed on these nodes. If node selector is not used, CSI driver gets deployed on all worker nodes.

To use this feature, perform the following steps after running the installer "spectrum-scale-driver.py":

- Label the Kubernetes worker nodes where IBM Spectrum Scale is running. Example:
	```
	kubectl label node node7 spectrumscalenode=yes --overwrite=true
	```
- Uncomment the lines from the following files: 
    * `deploy/csi-plugin-attacher.yaml`
    * `deploy/csi-plugin-provisioner.yaml`
    * `deploy/csi-plugin.yaml`

    ```
    #      nodeSelector:  
    #        spectrumscalenode: "yes"
    ```

- Deploy the driver by running `deploy/create.sh`

**Note:** If you choose to run csi plugin on selective nodes using the node selector then make sure pod using scale csi pvc are getting scheduled on nodes where csi driver is running.

### Kubernetes node to Spectrum Scale node mapping

In an environment where Kubernetes node names are different than the Spectrum Scale node names, this mapping feature must be used for application pods with Spectrum Scale as persistent storage to be successfully mounted.

To use this feature, perform the following steps after running the installer "spectrum-scale-driver.py":

- Add new environment variable in `deploy/csi-plugin.yaml` under container "*- name: csi-spectrum-scale*", where name of the environment variable is Kubernetes node name and value is the Spectrum Scale node name. 

  ```
  env:
     - name: k8snodename1  
       value: "scalenodename1"
     - name: k8snodename2  
       value: "scalenodename2"
  ```

  **Note:** Only add those nodes whose name is different in Kubernetes (`kubectl get nodes`) and Spectrum Scale (`mmlscluster/mmlsnode`)
	
- Deploy the driver by running `deploy/create.sh`

## Cleanup

1. Delete the resources that were created (pod, pvc, pv, storageclass)

2. Run deploy/destroy.sh script to cleanup the plugin resources

3. Find the CSI driver docker image and remove from all Kubernetes worker nodes

   ```
   % docker images -a | grep csi-spectrum-scale
   csi-spectrum-scale v0.9.0 465ca978127a 18 minutes ago 109MB

   % docker rmi 465ca978127a
   ```

## Links

[IBM Spectrum Scale Knowledge Center Welcome Page](https://www.ibm.com/support/knowledgecenter/en/STXKQY/ibmspectrumscale_welcome.html)
The Knowledge Center contains all official Spectrum Scale information and guidance.

[IBM Spectrum Scale FAQ](https://www.ibm.com/support/knowledgecenter/en/STXKQY/gpfsclustersfaq.html)
Main starting page for all Spectrum Scale compatibility information.

[IBM Spectrum Scale Protocols Quick Overview](https://www.ibm.com/developerworks/community/wikis/home?lang=en#!/wiki/fa32927c-e904-49cc-a4cc-870bcc8e307c/page/Protocols%20Quick%20Overview%20for%20IBM%20Spectrum%20Scale)
Guide showing how to quickly install a Spectrum Scale cluster. Information similar to the above Install Toolkit example.

[IBM Block CSI driver](https://github.com/IBM/ibm-block-csi-driver)
CSI driver supporting multiple IBM storage systems.

[Installing kubeadm](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/)
Main Kubernetes site detailing how to install kubeadm and create a cluster.

[OpenShift Container Platform 4.x Tested Integrations](https://access.redhat.com/articles/4128421)
Red Hat's test matrix for OpenShift 4.x.

[IBM Storage Enabler for Containers Welcome Page](https://www.ibm.com/support/knowledgecenter/en/SSCKLT/landing/IBM_Storage_Enabler_for_Containers_welcome_page.html)
Flex Volume driver released in late 2018 with a HELM update in early 2019, providing compatibility with IBM Spectrum Scale for file storage and multiple IBM storage systems for block storage. Future development efforts have shifted to CSI.

[Spectrum Scale Users Group](http://www.gpfsug.org/)
A group of both IBM and non-IBM users, interested in Spectrum Scale

[Spectrum Scale Users Group Mailing List and Slack Channel](https://www.spectrumscaleug.org/join/)
Join everyone and let the team know about your experience with the CSI driver
