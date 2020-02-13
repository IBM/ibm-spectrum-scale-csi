 

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
 - **volDirBasePath**: Base directory path relative to the filesystem mount point under which directory based volumes should be created. If specified, the storageClass is used for directory based (lightweight) volume creation.
 - **uid**: UID with which the volume should be created. Optional
 - **gid**: UID with which the volume should be created. Optional
 - **filesetType**: Type of fileset. Valid values are "independent" or "dependent". Default: independent
 - **parentFileset**: Specifies the parent fileset under which dependent fileset should be created.
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
IBM Block CSI driver supporting multiple IBM storage systems.

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
