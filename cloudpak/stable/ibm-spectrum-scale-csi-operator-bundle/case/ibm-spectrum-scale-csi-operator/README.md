
# ibm-spectrum-scale-csi-operator
The [IBM Storage Scale Container Storage Interface](https://github.com/IBM/ibm-spectrum-scale-csi) (CSI) project enables container orchestrators, such as Kubernetes and OpenShift, to manage the life-cycle of persistent storage.

This project contains an ansible-based operator to run and manage the deployment of the IBM Storage Scale CSI Driver.

# Introduction
This operator installs IBM Storage Scale CSI driver in kubernetes or Redhat Openshift Container platfrom cluster, consisting of Attacher, Provisioner, Resizer and Snapshotter deployments, and CSI driver DaemonSet.

## Details
The standard deployment of this operator consists of one pod for the Attacher, Provisioner, Resizer and Snapshotter each, and a pod on each node for the driver is to be used on.

### Configuration
Please refer to the [IBM Documentation](https://www.ibm.com/docs/en/spectrum-scale-csi?topic=260-configurations).

## Installing
Please refer to the [IBM Documentation](https://www.ibm.com/docs/en/spectrum-scale-csi?topic=260-installation).

# Limitations
Please refer to the [IBM Documentation](https://www.ibm.com/docs/en/spectrum-scale-csi?topic=260-limitations).

## Prerequisites
Please refer to the [IBM Documentation](https://www.ibm.com/docs/en/spectrum-scale-csi?topic=installation-performing-pre-tasks).

### Resources Required
Please refer to the [IBM Documentation](https://www.ibm.com/docs/en/spectrum-scale-csi?topic=260-planning).

# PodSecurityPolicy Requirements
This operator does not require any pod  security requirements.

# SecurityContextConstraints Requirements
The operator creates one custom SecurityContextConstraint (SCC) to control the access of various parts of the IBM Storage Scale CSI driver deployment. The SCC is created automatically by the operator if it does not exist, however it is not removed when the operator is uninstalled.

IBM Storage Scale CSI driver custom SecurityContextConstraints definition:

``` YAML
  kind: SecurityContextConstraints
  apiVersion: security.openshift.io/v1
  metadata:
    annotations:
      kubernetes.io/description: allow hostpath and host network to be accessible
    generation: 1
    name: spectrum-scale-csiaccess
    selfLink: /apis/security.openshift.io/v1/securitycontextconstraints/spectrum-scale-csiaccess
  readOnlyRootFilesystem: false
  requiredDropCapabilities:
  - KILL
  - MKNOD
  - SETUID
  - SETGID
  runAsUser:
    type: RunAsAny
  seLinuxContext:
    type: RunAsAny
  supplementalGroups:
    type: RunAsAny
  volumes:
  - configMap
  - downwardAPI
  - emptyDir
  - hostPath
  - persistentVolumeClaim
  - projected
  - secret
  allowHostDirVolumePlugin: true
  allowHostIPC: false
  allowHostNetwork: true
  allowHostPID: false
  allowHostPorts: false
  allowPrivilegeEscalation: true
  allowPrivilegedContainer: true
  allowedCapabilities: []
  defaultAddCapabilities: null
  fsGroup:
    type: MustRunAs
  {% if csiaccess_users|length > 0 %}
  users:
  {% for user in csiaccess_users %}
    - "{{user}}"
  {% endfor %}
  {% endif %}

```
