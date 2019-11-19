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
#USAGE spectrum-scale-driver-snap.sh [-n namespace] [-o output-dir] [-h]

ns="default"
node=""
outdir="."
cmd="kubectl"

while getopts 'n:o:h' OPTION; do
  case "$OPTION" in
    n)
      ns="$OPTARG"
      ;;

    o)
      outdir="$OPTARG"
      ;;

    h)
      echo "USAGE: spectrum-scale-driver-snap.sh [-n namespace] [-o output-dir] [-h]"
      exit 0
      ;;

    ?)
      echo "USAGE: spectrum-scale-driver-snap.sh [-n namespace] [-o output-dir] [-h]"
      exit 1
      ;;
  esac
done

if [[ $outdir != "." && ! -d $outdir ]]
then
  echo "Output directory $outdir does not exist. "
  exit 1
fi

# use oc commands on openshift cluster
out=$(which oc 2>&1) 
if [[ $? == 0 ]]
then
  out=$(oc status 2>&1)
  if [[ $? == 0 ]]
  then
    cmd="oc"
  fi
fi

# check if the namespace is valid and active
out=$($cmd get namespace | grep "$ns\s*Active" 2>&1)
if [[ $? != 0 ]]
then
  echo "Namespace $ns is invalid or not active. Please provide a valid namespace"
  exit 1
fi

# check if ibm-spectrum-scale-csi resources are running in specified namespace
out=$($cmd describe StatefulSet ibm-spectrum-scale-csi-attacher 2>&1 | grep 'Namespace' | awk 'BEGIN { FS="[[:space:]]+" } ; { print $2 }')
if [[ $out != $ns ]]
then
  echo "ibm-spectrum-scale-csi is not running in namespace $ns. Please provide a valid namespace"
  exit 1
fi

time=`date +"%m-%d-%Y-%T"`
logdir=${outdir%/}/ibm-spectrum-scale-csi-logs_$time


klog="$cmd logs --namespace $ns"
mkdir $logdir
CSI_SPECTRUM_SCALE_LABEL="ibm-spectrum-scale-csi"
PRODUCT_NAME="ibm-spectrum-scale-csi"

echo "Collecting \"$PRODUCT_NAME\" logs..."
echo "The log files will be saved in the folder [$logdir]"

csi_spectrum_scale_attacher_log_name=${logdir}/ibm-spectrum-scale-csi-attacher.log
csi_spectrum_scale_provisioner_log_name=${logdir}/ibm-spectrum-scale-csi-provisioner.log

describe_all_per_label=${logdir}/ibm-spectrum-scale-csi-describe-all-by-label
get_all_per_label=${logdir}/ibm-spectrum-scale-csi-get-all-by-label
get_configmap=${logdir}/ibm-spectrum-scale-csi-configmap
get_k8snodes=${logdir}/ibm-spectrum-scale-csi-k8snodes

echo "$klog StatefulSet/ibm-spectrum-scale-csi-attacher"
$klog StatefulSet/ibm-spectrum-scale-csi-attacher > ${csi_spectrum_scale_attacher_log_name} 2>&1 || :
echo "$klog StatefulSet/ibm-spectrum-scale-csi-provisioner"
$klog StatefulSet/ibm-spectrum-scale-csi-provisioner > ${csi_spectrum_scale_provisioner_log_name} 2>&1 || :

# kubectl logs on csi pods
for csi_pod in `$cmd get pod -l app=ibm-spectrum-scale-csi --namespace $ns | grep -v NAME | awk '{print $1}'`; do
   echo "$klog pod/${csi_pod}"
   $klog pod/${csi_pod} -c ibm-spectrum-scale-csi > ${logdir}/${csi_pod}.log 2>&1 || :
   $klog pod/${csi_pod} -c driver-registrar > ${logdir}/${csi_pod}-driver-registrar.log 2>&1 || :
   $klog pod/${csi_pod} -c ibm-spectrum-scale-csi --previous > ${logdir}/${csi_pod}-previous.log 2>&1 || :
   $klog pod/${csi_pod} -c driver-registrar --previous > ${logdir}/${csi_pod}-driver-registrar-previous.log 2>&1 || :
done

describe_label_cmd="$cmd describe all,cm,secret,storageclass,pvc,ds,serviceaccount -l product=${CSI_SPECTRUM_SCALE_LABEL} --namespace $ns"
echo "$describe_label_cmd"
$describe_label_cmd > $describe_all_per_label 2>&1 || :

describe_clusterroles="$cmd describe clusterroles/external-provisioner-runner clusterrolebindings/csi-provisioner-role clusterroles/external-attacher-runner clusterrolebindings/csi-provisioner-role clusterroles/csi-nodeplugin clusterrolebindings/csi-nodeplugin --namespace $ns"
echo "$describe_clusterroles"
$describe_clusterroles >> $describe_all_per_label 2>&1 || :

get_label_cmd="$cmd get all,cm,secret,storageclass,pvc,ds,serviceaccount --namespace $ns -l product=${CSI_SPECTRUM_SCALE_LABEL}"
echo "$get_label_cmd"
$get_label_cmd > $get_all_per_label 2>&1 || :

get_label_cmd="$cmd get pod --namespace $ns -o wide  -l product=${CSI_SPECTRUM_SCALE_LABEL}"
echo "$get_label_cmd"
$get_label_cmd >> $get_all_per_label 2>&1 || :

get_configmap_cmd="$cmd get configmap spectrum-scale-config --namespace $ns -o yaml"
echo "$get_configmap_cmd"
$get_configmap_cmd > $get_configmap 2>&1 || :

get_k8snodes_cmd="$cmd get nodes"
echo "$get_k8snodes_cmd"
$get_k8snodes_cmd > $get_k8snodes 2>&1 || :

get_k8snodes_cmd="$cmd describe nodes"
echo "$get_k8snodes_cmd"
$get_k8snodes_cmd >> $get_k8snodes 2>&1 || :

if [[ $cmd == "oc" ]]
then
   get_scc_cmd="$cmd describe scc csiaccess"
   echo "$get_scc_cmd"
   $get_scc_cmd > ${logdir}/${PRODUCT_NAME}-scc.log 2>&1 || :
fi

get_clusterinfo_cmd="$cmd cluster-info dump --namespaces kube-system --output-directory=$logdir"
echo "$get_clusterinfo_cmd"
out=$($get_clusterinfo_cmd 2>&1)

echo "Finished collecting \"$PRODUCT_NAME\" logs in the folder -> $logdir"

