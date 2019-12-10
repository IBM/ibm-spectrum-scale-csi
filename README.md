
   * [IBM Spectrum Scale Container Storage Interface (CSI) Driver](#ibm-spectrum-scale-container-storage-interface-csi-driver)
      * [IBM Spectrum Scale Introduction](#ibm-spectrum-scale-introduction)
      * [IBM Spectrum Scale Container Storage Interface (CSI) driver](#ibm-spectrum-scale-container-storage-interface-csi-driver)
         * [Supported Features of the CSI driver](#supported-features-of-the-csi-driver)
         * [Limitations of the CSI driver](#limitations-of-the-csi-driver)
         * [Pre-requisites for installing and running the CSI driver](#pre-requisites-for-installing-and-running-the-csi-driver)
      * [Building the docker image](#building-the-docker-image)
      * [Install and Deploy the IBM Spectrum Scale CSI driver](#install-and-deploy-the-spectrum-scale-csi-driver)
      * [Static Provisioning](#static-provisioning)
      * [Dynamic Provisioning](#dynamic-provisioning)
         * [Storageclass](#storageClass)
      * [Environments in Test](TESTCONFIG.md#environments-in-test)
      * [Example Hardware Configs](TESTCONFIG.md#example-hardware-configs)
      * [Links](#links)

  

# IBM Spectrum Scale Container Storage Interface (CSI) Driver

Please refer to the IBM Spectrum Scale Container Storage Interface Driver documentation on knowledge center for detailed documentation. See the IBM Spectrum Scale Users Group links at the very bottom for a community to share and discuss test efforts.

  
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

Please refer to [IBM Spectrum Scale Knowledge Center](https://www.ibm.com/support/knowledgecenter/en/STXKQY/ibmspectrumscale_welcome.html) for limitations.

### Pre-requisites for installing and running the CSI driver

Please refer to [IBM Spectrum Scale Knowledge Center](https://www.ibm.com/support/knowledgecenter/en/STXKQY/ibmspectrumscale_welcome.html) for install pre-requisites.

## Building the docker image


**Using multi-stage build**

Pre-requisite: Docker 17.05 or higher is installed on local build machine.


1. Clone the code

   ```
   git clone https://github.com/IBM/ibm-spectrum-scale-csi-driver.git
   cd ibm-spectrum-scale-csi-driver
   ```

2. Invoke multi-stage build

   ```
   docker build -t ibm-spectrum-scale-csi:v1.0.0 -f Dockerfile.msb .
   ```

   On podman setup, use *podman* command instead of *docker*

3. save the docker image

   ```
   docker save ibm-spectrum-scale-csi:v1.0.0 -o ibm-spectrum-scale-csi_v1.0.0.tar
   ```

   On podman setup, use *podman* command instead of *docker*

      A tar file of docker image will be created.




## Install and Deploy the IBM Spectrum Scale CSI driver


1. Copy and load the docker image on all Kubernetes worker nodes

   ```
   docker image load -i ibm-spectrum-scale-csi_v1.0.0.tar
   ```

   On podman setup, use *podman* command instead of *docker*


2. Deploy CSI driver

   Follow the instructions from [ibm-spectrum-scale-csi-operator](https://github.com/IBM/ibm-spectrum-scale-csi-operator) for deployment of CSI driver

   For Advance configuration, Cleanup, Troubleshooting etc. refer [IBM Spectrum Scale Knowledge Center](https://www.ibm.com/support/knowledgecenter/en/STXKQY_5.0.4/com.ibm.spectrum.scale.csi.v5r04.doc/bl1csi_kc_landing.html)


## Static Provisioning

In static provisioning, the backend storage volumes and PVs are created by the administrator. Static provisioning can be used to provision a directory or fileset with existing data.

For static provisioning of existing directories perform the following steps:

- Generate static pv yaml file using helper script

   ```
   tools/generate_pv_yaml.sh --filesystem gpfs0 --size 10 \
   --linkpath /ibm/gpfs0/pvfileset/static-pv --pvname static-pv
   ```

- Use sample static_pvc and pod files for sanity test under `examples/static`

   ```
   kubectl apply -f examples/static/static_pv.yaml
   kubectl apply -f examples/static/static_pvc.yaml
   kubectl apply -f examples/static/static_pod.yaml
   ```
  

## Dynamic Provisioning

Dynamic provisioning is used to dynamically provision the storage backend volume based on the storageClass.

### Storageclass
Storageclass defines what type of backend volume should be created by dynamic provisioning. IBM Spectrum Scale CSI driver supports creation of directory based (also known as lightweight volumes) and fileset based (independent as well as dependent) volumes. Following parameters are supported by IBM Spectrum Scale CSI driver storageClass:

 - **volBackendFs**: Filesystem on which the volume should be created. This is a mandatory parameter.
 - **clusterId**: Cluster ID on which the volume should be created. 
 - **volDirBasePath**: Base directory path relative to the filesystem mount point under which directory based volumes should be created. If specified, the storageClass is used for directory based (lightweight) volume creation. If not specified, storageClass creates fileset based volumes.
 - **uid**: UID with which the volume should be created. Optional
 - **gid**: UID with which the volume should be created. Optional
 - **filesetType**: Type of fileset. Valid values are "independent" or "dependent". Default is "independent". 
 - **parentFileset**: Specifies the parent fileset under which dependent fileset should be created. Mandatory if "filesetType" is specified.
 - **inodeLimit**: Inode limit for fileset based volumes. If not specified, default IBM Spectrum Scale inode limit of 1 million is used.
 
For dynamic provisioning, use sample storageClass, pvc and pod files for sanity test under examples/dynamic

Example:

   ```
   kubectl apply -f examples/dynamic/fileset/storageclassfileset.yaml
   kubectl apply -f examples/dynamic/fileset/pvcfset.yaml
   kubectl apply -f examples/dynamic/fileset/podfset.yaml
   ```


## Links

[IBM Spectrum Scale Knowledge Center Welcome Page](https://www.ibm.com/support/knowledgecenter/en/STXKQY/ibmspectrumscale_welcome.html)
The Knowledge Center contains all official IBM Spectrum Scale information and guidance.

[IBM Spectrum Scale FAQ](https://www.ibm.com/support/knowledgecenter/en/STXKQY/gpfsclustersfaq.html)
Main starting page for all IBM Spectrum Scale compatibility information.

[IBM Spectrum Scale Protocols Quick Overview](https://www.ibm.com/developerworks/community/wikis/home?lang=en#!/wiki/fa32927c-e904-49cc-a4cc-870bcc8e307c/page/Protocols%20Quick%20Overview%20for%20IBM%20Spectrum%20Scale)
Guide showing how to quickly install a IBM Spectrum Scale cluster. Information similar to the above Install Toolkit example.

[IBM Block CSI driver](https://github.com/IBM/ibm-block-csi-driver)
CSI driver supporting multiple IBM storage systems.

[Installing kubeadm](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/)
Main Kubernetes site detailing how to install kubeadm and create a cluster.

[OpenShift Container Platform 4.x Tested Integrations](https://access.redhat.com/articles/4128421)
Red Hat's test matrix for OpenShift 4.x.

[IBM Storage Enabler for Containers Welcome Page](https://www.ibm.com/support/knowledgecenter/en/SSCKLT/landing/IBM_Storage_Enabler_for_Containers_welcome_page.html)
Flex Volume driver released in late 2018 with a HELM update in early 2019, providing compatibility with IBM Spectrum Scale for file storage and multiple IBM storage systems for block storage. Future development efforts have shifted to CSI.

[IBM Spectrum Scale Users Group](http://www.gpfsug.org/)
A group of both IBM and non-IBM users, interested in IBM Spectrum Scale

[IBM Spectrum Scale Users Group Mailing List and Slack Channel](https://www.spectrumscaleug.org/join/)
Join everyone and let the team know about your experience with the CSI driver
