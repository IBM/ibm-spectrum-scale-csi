 

## Environments in Test

  

  

**Software components**

|Component | Level |Comments |
|--|--|--|
|Spectrum Scale | 5.0.3.3 | x86_64 |
|OS | RHEL 7.6 | |
|Kubernetes | 1.15 | |
|Docker | 18.03, 18.06 | |
|Calico | 3.7, 3.8.2 | |
|Weave | 2.5.2 | |
|OpenShift | 4.1 | *not fully qualified (see listed limitations*) |

  

  

## Example Hardware Configs

  

All examples below consider a non-production Beta test environment:

  

### Small virtual cluster testbed

This environment assumes a minimum number of nodes to perform goodpath testing. Because this assumes a virtual environment, understand the [Spectrum Scale support statement for KVM / VMware clusters](https://www.ibm.com/support/knowledgecenter/STXKQY/gpfsclustersfaq.html?view=kc#virtual)

  

| Node Type | # | Cores | RAM | HD | Spectrum Scale | K8s |
|--|--|--|--|--|--|--|
| **Master node** | 1x | 2core | 8GB | 50GB | do not install | required |
| **Worker node** | 2x | 2core | 16GB | 50GB | required | required |
| **GUI node** | 1x | 2core | 16GB | 50GB | required | do not install |
| **NSD node** | 2x | 2core | 8GB | 50GB | required | optional |

  

- gpfs pagepool set to 2GB (*mmchconfig pagepool=2G*)

- No other Spectrum Scale functions installed

- Do not install Spectrum Scale on the Master node

- Do not install Kubernetes on the GUI node

- NSD nodes with shared disks (8x 10GB) for NSD and file system creation.  Boost memory and cpu cores on NSD nodes if stacking K8s / Scale functionality

  

### Baremetal cluster testbed

This environment assumes a minimum number of nodes to perform both goodpath and error inject testing.

  

| Node Type | # | Cores | RAM | HD | Spectrum Scale | K8s |
|--|--|--|--|--|--|--|
| **Master node** | 3x | 8core | 64GB | 300GB | do not install | required |
| **Worker node** | 3x | 8core | 128GB | 300GB | required | required |
| **GUI node** | 1x | 4core | 32GB | 300GB | required | do not install |
| **NSD node** | 2x | 4core | 32GB | 300GB | required | optional |

  

- pagepool set to 32GB on workers, 8GB on NSD and GUI

- No other Spectrum Scale functions installed

- Do not install Spectrum Scale on the Master nodes

- Do not install Kubernetes on the GUI node

- NSD nodes with shared disks (8x 100GB) for NSD and file system creation. Boost memory and cpu cores on NSD nodes if stacking K8s / Scale functionality

  

### OpenShift VMware testbed

This environment assumes a minimum number of nodes to perform goodpath testing. Because this assumes a virtual environment, understand the [Spectrum Scale support statement for KVM / VMware clusters](https://www.ibm.com/support/knowledgecenter/STXKQY/gpfsclustersfaq.html?view=kc#virtual)

  

| Node Type | # | Cores | RAM | HD | Spectrum Scale | K8s |
|--|--|--|--|--|--|--|
| **Master node** | 3x | 2core | 32GB | 50GB | do not install | required |
| **Worker node** | 3x | 2core | 32GB | 50GB | required | required |
| **GUI node** | 1x | 2core | 16GB | 50GB | required | do not install |
| **NSD node** | 2x | 2core | 16GB | 50GB | required | optional |

  

- pagepool set to 8GB on workers, 4GB on NSD and GUI

- No other Spectrum Scale functions installed

- Master nodes will run RHCOS. Spectrum Scale is not supported on RHCOS nodes. Do not install Spectrum Scale on the master nodes.

- Do not install Kubernetes on the GUI node

- NSD nodes with shared disks (8x 10GB) for NSD and file system creation. Boost memory on NSD nodes if stacking K8s / Scale functionality



  

### OpenShift baremetal testbed

  

| Node Type | # | Cores | RAM | HD | Spectrum Scale | K8s |
|--|--|--|--|--|--|--|
| **Master node** | 3x | 2core | 64GB | 50GB | do not install | required |
| **Worker node** | 3x | 2core | 128GB | 50GB | required | required |
| **GUI node** | 1x | 2core | 32GB | 50GB | required | do not install |
| **NSD node** | 2x | 2core | 32GB | 50GB | required | optional

  

- pagepool set to 32GB on workers, 8GB on NSD and GUI

- no other Spectrum Scale functions installed

- Master nodes will run RHCOS. Spectrum Scale is not supported on RHCOS nodes. Do not install Spectrum Scale on the master nodes.

- Do not install Kubernetes on the GUI node

- NSD nodes with shared disks (8x 100GB) for NSD and file system creation.  Boost memory and cpu cores on NSD nodes if stacking K8s / Scale functionality



  

  

## Example of using the Install Toolkit to build a Spectrum Scale cluster for testing the CSI driver

  

This example assumes 7 nodes. Kubernetes will be installed after the Spectrum Scale cluster is installed, although this can be reversed if desired.

  

| Node Type | Name | Spectrum Scale | K8s |
|--|--|--|--|
| **Master node** | csi-demo-master | no | yes
| **Worker node** | csi-demo-worker-1 | yes | yes
| **Worker node** | csi-demo-worker-2 | yes | yes
| **Worker node** | csi-demo-worker-3 | yes | yes
| **GUI node** | csi-demo-gui | yes | no
| **NSD node** | csi-demo-nsd-1 | yes | no
| **NSD node** | csi-demo-nsd-1 | yes | no

  

  

1) Verify the OS file system has ftype=1. This is important for when a container overlay FS is installed. If ftype=0, stop here. The file system ftype cannot be changed without recreating the file system.

  

```

xfs_info /

  

meta-data=/dev/mapper/rhel-root isize=512 agcount=4, agsize=15269632 blks

= sectsz=512 attr=2, projid32bit=1

= crc=1 finobt=0 spinodes=0

data = bsize=4096 blocks=61078528, imaxpct=25

= sunit=0 swidth=0 blks

naming =version 2 bsize=4096 ascii-ci=0 ftype=1

log =internal bsize=4096 blocks=29823, version=2

= sectsz=512 sunit=0 blks, lazy-count=1

realtime =none extsz=4096 blocks=0, rtextents=0

```

  

2) Turn off the firewall and selinux

  

```

systemctl status firewalld

systemctl stop firewalld

systemctl disable firewalld

  

setenforce 0

vim /etc/selinux/config

----- change to disabled or permissive

SELINUX=disabled

```

  

3) Extract the Spectrum Scale code on the desired installer node

```

./Spectrum_Scale_Data_Management-5.0.3.3-x86_64-Linux-install

```

  

4) Change to the installer location

```

cd /usr/lpp/mmfs/5.0.3.3/installer

```

  

- Refer to the [Spectrum Scale Protocols Quick Overview Guide](https://www.ibm.com/developerworks/community/wikis/home?lang=en#!/wiki/fa32927c-e904-49cc-a4cc-870bcc8e307c/page/Protocols%20Quick%20Overview%20for%20IBM%20Spectrum%20Scale) for more details and links on the steps that will follow.

  

5) Setup the installer node. Indicate an IP on the installer node that all other nodes can see.

```

./spectrumscale setup -s 172.16.180.70

```

  

6) Add nodes to be part of the Spectrum Scale cluster. The installer node does not need to be one of these nodes

```

./spectrumscale node add csi-demo-gui.ibm.com -a -g

./spectrumscale node add csi-demo-worker-1.ibm.com

./spectrumscale node add csi-demo-worker-2.ibm.com

./spectrumscale node add csi-demo-worker-3.ibm.com

./spectrumscale node add csi-demo-nsd-1.ibm.com -n

./spectrumscale node add csi-demo-nsd-2.ibm.com -n

```

  

7) Check the disk devices available on the NSD nodes. In this example, there are 4 extra disks on csi-demo-nsd-1 and csi-demo-nsd-2. For ease of explanation, these are non-shared disks, meaning that both nodes csi-demo-nsd-1 and csi-demo-nsd-2 see a different set of 4 disks.

```

ssh csi-demo-nsd-1 lsblk

NAME MAJ:MIN RM SIZE RO TYPE MOUNTPOINT

vda 252:0 0 250G 0 disk

├─vda1 252:1 0 1G 0 part /boot

└─vda2 252:2 0 249G 0 part

├─rhel-root 253:0 0 233G 0 lvm /

└─rhel-swap 253:1 0 16G 0 lvm [SWAP]

vdb 252:16 0 50G 0 disk <--- will use with Spectrum Scale

vdc 252:32 0 50G 0 disk <--- will use with Spectrum Scale

vdd 252:48 0 50G 0 disk <--- will use with Spectrum Scale

vde 252:64 0 50G 0 disk <--- will use with Spectrum Scale

```

Repeat for any other NSD nodes and identify the block devices to use with Spectrum Scale.

  

8) Tell the installer what block devices to use with Spectrum Scale, which NSD node they are attached to, desired usage, desired file system name, desired failure group settings. This example uses non-shared disks and will assign all disks on nsd-1 to file system fs1, and all disks on nsd-2 to file system fs2. This is not redundant but will work for the purposes of this example. If redundancy is desired, either create multi-writer/shared disks across both NSD servers + input a secondary server in each line below. Or create only a single file system with disks on nsd-1 in failure group 1 (fg 1) and disks on nsd-2 in failure group 2 (fg 2).

  

```

./spectrumscale nsd add -p csi-demo-nsd-1.ibm.com -u dataAndMetadata -fs fs1 -fg 1 "/dev/vdb"

./spectrumscale nsd add -p csi-demo-nsd-1.ibm.com -u dataAndMetadata -fs fs1 -fg 1 "/dev/vdc"

./spectrumscale nsd add -p csi-demo-nsd-1.ibm.com -u dataAndMetadata -fs fs1 -fg 1 "/dev/vdd"

./spectrumscale nsd add -p csi-demo-nsd-1.ibm.com -u dataAndMetadata -fs fs1 -fg 1 "/dev/vde"

  

./spectrumscale nsd add -p csi-demo-nsd-2.ibm.com -u dataAndMetadata -fs fs2 -fg 1 "/dev/vdb"

./spectrumscale nsd add -p csi-demo-nsd-2.ibm.com -u dataAndMetadata -fs fs2 -fg 1 "/dev/vdc"

./spectrumscale nsd add -p csi-demo-nsd-2.ibm.com -u dataAndMetadata -fs fs2 -fg 1 "/dev/vdd"

./spectrumscale nsd add -p csi-demo-nsd-2.ibm.com -u dataAndMetadata -fs fs2 -fg 1 "/dev/vde"

```

  

- If using non-standard device names, consult the following file for advice:

```

vim /usr/lpp/mmfs/samples/nsddevices.sample

```

  

9) Set the name of the cluster to csi-demo.ibm.com

```

./spectrumscale config gpfs -c csi-demo.ibm.com

```

  

10) Disable callhome for this test

```

./spectrumscale callhome disable

```

  

11) List the nodes, file systems, NSDs, and gpfs cluster config. Make sure this reflects the desired cluster.

```

./spectrumscale node list

./spectrumscale filesystem list

./spectrumscale nsd list

./spectrumscale gpfs config

```

  

12) Run an install precheck.

```

./spectrumscale install --precheck

```

  

13) Run the install. This will install Spectrum Scale on all nodes previously inputted into the Install Toolkit. It will set the appropriate licenses, create NSDs, install and configure performance monitoring, install the Spectrum Scale GUI, and start Spectrum Scale on each node. This will take roughly 10min depending upon number of nodes, bandwidth, cpu.

```

./spectrumscale install

```

  

14) Run a deploy precheck.

  

```

./spectrumscale deploy --precheck

```

  

15) Run the deploy. This will create the Spectrum Scale file system on the NSDs previously created by the install and previously inputted into the Install Toolkit. It will take roughly 15min depending upon the number of nodes, bandwidth, cpu.

```

./spectrumscale deploy

```

  

16) Configure quotas for the file systems which were just created. This is necessary for CSI.

```

mmchfs fs1 -Q yes

mmchfs fs2 -Q yes

```

  

17) Spectrum Scale is now installed, configured, and active on 6 nodes. A 7th node was reserved as the master node for k8s. If Kubernetes is not yet installed, proceed as follows:

  

- Install Docker on Master and Worker nodes

- Install K8s on Master and Worker nodes

- Initialize the K8s Master node

- Setup Calico networking on the Master node

- Join the Worker nodes to the K8s cluster

  

An intersection of Spectrum Scale and Kubernetes now exists. The Kubernetes Master node does not have Spectrum Scale installed upon it, thus giving it maximum cpu/memory resources for managing the K8s cluster. The Kubernetes Worker nodes are also Spectrum Scale nodes. The GUI and NSD nodes have Spectrum Scale installed and are part of the Spectrum Scale cluster, but are not part of the Kubernetes cluster. The CSI driver will run on the Worker nodes. Because all Worker nodes have Spectrum Scale installed, the pods on all worker nodes will have access to PVC mounts to the underlying Spectrum Scale file system. 

  

  

  

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

  

[Spectrum Scale Users Group](http://www.gpfsug.org/%29%5D%28http://www.gpfsug.org/)
A group of both IBM and non-IBM users, interested in Spectrum Scale

  
[Spectrum Scale Users Group Mailing List and Slack Channel](https://www.spectrumscaleug.org/join/%29%5D%28https://www.spectrumscaleug.org/join/)
Join everyone and let the team know about your experience with the CSI driver
