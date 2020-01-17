 

## Environments in Test

  

  

**Software components**

|Component | Level |Comments |
|--|--|--|
|Spectrum Scale | 5.0.4.1 | x86_64 |
|OS | RHEL 7.5, RHEL 7.6, RHEL 7.7 worker nodes | |
|Kubernetes | 1.13, 1.14, 1.15, 1.16 | |
|Docker | 18.03, 18.06 | |
|OpenShift | 4.2 | |

  

  

## Example Hardware Configs

  

All examples below consider a non-production test environment:

  

### Small virtual cluster testbed

This environment assumes a minimum number of nodes to perform goodpath testing. Because this assumes a virtual environment, understand the [Spectrum Scale support statement for KVM / VMware clusters](https://www.ibm.com/support/knowledgecenter/STXKQY/gpfsclustersfaq.html?view=kc#virtual)

  

| Node Type | # | Cores | RAM | HD | Spectrum Scale | K8s |
|--|--|--|--|--|--|--|
| **Master node** | 1x | 2core | 8GB | 50GB | do not install | required |
| **Worker node** | 2x | 2core | 16GB | 50GB | required | required |
| **GUI node** | 1x | 2core | 16GB | 50GB | required | do not install |
| **NSD node** | 2x | 2core | 8GB | 50GB | required | not recommended  |

  

- gpfs pagepool set to 2GB (*mmchconfig pagepool=2G*)

- No other Spectrum Scale functions installed

- Do not install Spectrum Scale on the Master node

- Do not install Kubernetes on the GUI node

- NSD nodes with shared disks (8x 10GB) for NSD and file system creation.  Boost memory and cpu cores on NSD nodes if stacking K8s / Scale functionality


### OpenShift VMware testbed

This environment assumes a minimum number of nodes to perform goodpath testing. Because this assumes a virtual environment, understand the [Spectrum Scale support statement for KVM / VMware clusters](https://www.ibm.com/support/knowledgecenter/STXKQY/gpfsclustersfaq.html?view=kc#virtual)

  

| Node Type | # | Cores | RAM | HD | Spectrum Scale | K8s |
|--|--|--|--|--|--|--|
| **Master node** | 3x | 4core | 16GB | 120GB | do not install | required |
| **Infrastructure node** | 3x | 4core | 16GB | 120GB | required | required |
| **Worker node** | 3x | 2core | 16GB | 120GB | required | required |
| **GUI node** | 1x | 2core | 16GB | 50GB | required | do not install |
| **NSD node** | 2x | 2core | 16GB | 50GB | required | not recommended |

  

- pagepool set to 8GB on workers, 4GB on NSD and GUI

- No other Spectrum Scale functions installed

- Master nodes will run RHCOS. Spectrum Scale is not supported on RHCOS nodes. Do not install Spectrum Scale on the master nodes.

- Do not add Spectrum Scale GUI node to Openshift cluster 

- NSD nodes with shared disks (8x 10GB) for NSD and file system creation. Boost memory on NSD nodes if stacking K8s / Scale functionality



  

### OpenShift baremetal testbed

  

| Node Type | # | Cores | RAM | HD | Spectrum Scale | K8s |
|--|--|--|--|--|--|--|
| **Master node** | 3x | 2core | 64GB | 120GB | do not install | required |
| **Infrastructure node** | 3x | 4core | 64GB | 120GB | required | required |
| **Worker node** | 3x | 2core | 128GB | 120GB | required | required |
| **GUI node** | 1x | 2core | 32GB | 50GB | required | do not install |
| **NSD node** | 2x | 2core | 32GB | 50GB | required | not recommended |

  

- pagepool set to 32GB on workers, 8GB on NSD and GUI

- no other Spectrum Scale functions installed

- Master nodes will run RHCOS. Spectrum Scale is not supported on RHCOS nodes. Do not install Spectrum Scale on the master nodes.

- Do not add Spectrum Scale GUI node to Openshift cluster

- NSD nodes with shared disks (8x 100GB) for NSD and file system creation.  Boost memory and cpu cores on NSD nodes if stacking K8s / Scale functionality
