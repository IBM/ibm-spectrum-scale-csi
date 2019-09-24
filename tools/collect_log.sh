#!/bin/bash
#
# Copyright 2019 IBM Corp.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
#USAGE collect_log.sh <namespace> 

ns=${1-default}

time=`date +"%m-%d-%Y-%T"`
logdir=./spectrum_scale_csi_collect_logs_$time
klog="kubectl logs --namespace $ns"
mkdir $logdir
CSI_SPECTRUM_SCALE_LABEL="ibm-spectrum-scale-csi"
PRODUCT_NAME="ibm-spectrum-scale-csi"

echo "Collecting \"$PRODUCT_NAME\" logs..."
echo "The log files will be saved in the folder [$logdir]"

csi_spectrum_scale_attacher_log_name=${logdir}/csi_spectrum_scale_attacher.log
csi_spectrum_scale_provisioner_log_name=${logdir}/csi_spectrum_scale_provisioner.log

describe_all_per_label=${logdir}/csi_spectrum_scale_describe_all_by_label
get_all_per_label=${logdir}/csi_spectrum_scale_get_all_by_label
get_configmap=${logdir}/csi_spectrum_scale_configmap
get_k8snodes=${logdir}/csi_spectrums_scale_k8snodes

echo "$klog StatefulSet/csi-spectrum-scale-attacher"
$klog StatefulSet/csi-spectrum-scale-attacher > ${csi_spectrum_scale_attacher_log_name} 2>&1 || :
echo "$klog StatefulSet/csi-spectrum-scale-provisioner"
$klog StatefulSet/csi-spectrum-scale-provisioner > ${csi_spectrum_scale_provisioner_log_name} 2>&1 || :

# kubectl logs on csi pods
for csi_pod in `kubectl get pod -l app=csi-spectrum-scale --namespace $ns | grep -v NAME | awk '{print $1}'`; do
   echo "$klog pod/${csi_pod}"
   $klog pod/${csi_pod} -c csi-spectrum-scale > ${logdir}/${csi_pod}.log 2>&1 || :
   $klog pod/${csi_pod} -c driver-registrar > ${logdir}/${csi_pod}_driver_registrar.log 2>&1 || :
done

describe_label_cmd="kubectl describe all,cm,secret,storageclass,pvc,ds,serviceaccount -l product=${CSI_SPECTRUM_SCALE_LABEL} --namespace $ns"
echo "$describe_label_cmd"
$describe_label_cmd > $describe_all_per_label 2>&1 || :

describe_clusterroles="kubectl describe clusterroles/external-provisioner-runner clusterrolebindings/csi-provisioner-role clusterroles/external-attacher-runner clusterrolebindings/csi-provisioner-role clusterroles/csi-nodeplugin clusterrolebindings/csi-nodeplugin --namespace $ns"
echo "$describe_clusterroles"
$describe_clusterroles >> $describe_all_per_label 2>&1 || :

get_label_cmd="kubectl get all,cm,secret,storageclass,pvc,ds,serviceaccount --namespace $ns -l product=${CSI_SPECTRUM_SCALE_LABEL}"
echo "$get_label_cmd"
$get_label_cmd > $get_all_per_label 2>&1 || :

get_label_cmd="kubectl get pod --namespace $ns -o wide  -l product=${CSI_SPECTRUM_SCALE_LABEL}"
echo "$get_label_cmd"
$get_label_cmd >> $get_all_per_label 2>&1 || :

get_configmap_cmd="kubectl get configmap spectrum-scale-config --namespace $ns -o yaml"
echo "$get_configmap_cmd"
$get_configmap_cmd > $get_configmap 2>&1 || :

get_k8snodes_cmd="kubectl get nodes"
echo "$get_k8snodes_cmd"
$get_k8snodes_cmd > $get_k8snodes 2>&1 || :

get_k8snodes_cmd="kubectl describe nodes"
echo "$get_k8snodes_cmd"
$get_k8snodes_cmd >> $get_k8snodes 2>&1 || :

echo "Finished collecting \"$PRODUCT_NAME\" logs in the folder -> $logdir"

