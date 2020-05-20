# ibm-spectrum-scale-csi-operator
The [IBM Spectrum Scale Container Storage Interface](https://github.com/IBM/ibm-spectrum-scale-csi) (CSI) project enables container orchestrators, such as Kubernetes and OpenShift, to manage the life-cycle of persistent storage.

This project contains an ansible-based operator to run and manage the deployment of the IBM Spectrum Scale CSI Driver.

# Introduction
This operator installs IBM Spectrum Scale CSI in the cluster, consisting of Attacher and Provisioner StatefulSets, and Driver DaemonSet.

## Details
The standard deployment of this operator consists of one pod for the Attacher and Provisioner each, and a pod on each node the driver is to be used on.

### Configuration
Please refer to the [Knowledge Center Documentation](https://www.ibm.com/support/knowledgecenter/en/STXKQY_5.0.4/com.ibm.spectrum.scale.csi.v5r04.doc/bl1csi_configurations.html).

## Installing
Please refer to the [Knowledge Center Documentation](https://www.ibm.com/support/knowledgecenter/en/STXKQY_5.0.4/com.ibm.spectrum.scale.csi.v5r04.doc/bl1csi_instal_intro.html).

# Limitations
Please refer to the [Knowledge Center Documentation](https://www.ibm.com/support/knowledgecenter/en/STXKQY_5.0.4/com.ibm.spectrum.scale.csi.v5r04.doc/bl1csi_limitations.html).

## Prerequisites
Please refer to the [Knowledge Center Documentation](https://www.ibm.com/support/knowledgecenter/en/STXKQY_5.0.4/com.ibm.spectrum.scale.csi.v5r04.doc/bl1csi_instal_prereq.html).

### Resources Required
Please refer to the [Knowledge Center Documentation](https://www.ibm.com/support/knowledgecenter/en/STXKQY_5.0.4/com.ibm.spectrum.scale.csi.v5r04.doc/bl1csi_planning.html).

# PodSecurityPolicy Requirements
This operator does not require any pod  security requirements.

# SecurityContextConstraints Requirements
The operator maintains the Security Context Constraints, removing the required restraints when the operator is uninstalled.

The installed SCC is as follows, please note this is a jinja2 template applied by the operator:

``` YAML
  kind: SecurityContextConstraints
  apiVersion: security.openshift.io/v1
  metadata:
    annotations:
      kubernetes.io/description: allow hostpath and host network to be accessible
    generation: 1
    name: csiaccess
    selfLink: /apis/security.openshift.io/v1/securitycontextconstraints/csiaccess
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
  groups:
  - system:authenticated
  {% if csiaccess_users|length > 0 %}
  users:
  {% for user in csiaccess_users %}
    - "{{user}}"
  {% endfor %}
  {% endif %}

```
