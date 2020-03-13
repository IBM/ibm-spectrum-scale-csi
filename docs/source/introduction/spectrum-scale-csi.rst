IBM Spectrum Scale CSI Driver
=============================

The IBM Spectrum Scale Container Storage Interface (CSI) driver allows IBM Spectrum Scale to be used as persistent storage 
for stateful application running in Kubernetes clusters. Through this CSI Driver, Kubernetes persistent volumes (PVs) can 
be provisioned from IBM Spectrum Scale. Thus, containers can be used with stateful microservices, such as database applications 
(MongoDB, PostgreSQL etc), web servers (nginx, apache), or any number of other containerized applications needing provisioned 
storage.


Features
--------

:Static provisioning: Ability to use existing directories as persistent volumes
:Lightweight dynamic provisioning: Ability to create directory-based volumes dynamically
:Fileset-based dynamic provisioning: Ability to create fileset-based volumes dynamically
:Multiple file systems support: Volumes can be created across multiple file systems
:Remote mount support: Volumes can be created on a remotely mounted file system


