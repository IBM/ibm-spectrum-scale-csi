#!/bin/bash

echo "==Cluster Role=="
kubectl get -n ibm-spectrum-scale-csi-driver ClusterRole  | grep -i csi
echo "==Cluster Role Binding=="
kubectl get -n ibm-spectrum-scale-csi-driver ClusterRoleBinding | grep -i csi
echo "==Stateful Set=="
kubectl get -n ibm-spectrum-scale-csi-driver StatefulSet | grep -i csi
echo "==Daemon Set=="
kubectl get -n ibm-spectrum-scale-csi-driver DaemonSet | grep -i csi
echo "==Service Account=="
kubectl get -n ibm-spectrum-scale-csi-driver ServiceAccount | grep -i csi
echo "==Config Map=="
kubectl get -n ibm-spectrum-scale-csi-driver ConfigMap | grep -i csi
