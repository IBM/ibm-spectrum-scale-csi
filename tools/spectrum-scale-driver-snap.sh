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

ns="ibm-spectrum-scale-csi-driver"
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
if (which oc &>/dev/null)
then
  if (oc status &>/dev/null)
  then
    cmd="oc"
  fi
fi

# check if the namespace is valid and active
if !($cmd get namespace | grep "$ns\s*Active")
then
  echo "Namespace $ns is invalid or not active. Please provide a valid namespace"
  exit 1
fi

operator=`$cmd get deployment --field-selector=metadata.name==ibm-spectrum-scale-csi-operator --namespace $ns  |grep -v NAME |awk '{print $1}'`
if [[ "$operator" != "ibm-spectrum-scale-csi-operator" ]]; then
      echo "ibm-spectrum-scale-csi driver and operator is not running in namespace $ns. Please provide a valid namespace"
      exit 1
 fi

time=$(date +"%m-%d-%Y-%T")
logdir=${outdir%/}/ibm-spectrum-scale-csi-logs_$time

klog="$cmd logs --namespace $ns"
mkdir "$logdir"
CSI_SPECTRUM_SCALE_LABEL="ibm-spectrum-scale-csi"
PRODUCT_NAME="ibm-spectrum-scale-csi"

echo "Collecting \"$PRODUCT_NAME\" logs..."
echo "The log files will be saved in the folder [$logdir]"

describe_all_per_label=${logdir}/ibm-spectrum-scale-csi-describe-all-by-label
get_all_per_label=${logdir}/ibm-spectrum-scale-csi-get-all-by-label
get_configmap=${logdir}/ibm-spectrum-scale-csi-configmap
get_k8snodes=${logdir}/ibm-spectrum-scale-csi-k8snodes
get_daemonset=${logdir}/ibm-spectrum-scale-csi-daemonsets
describe_CSIScaleOperator=${logdir}/ibm-spectrum-scale-csi-describe-CSIScaleOperator

for statefulSetName in `$cmd -n $ns get StatefulSet --no-headers -l "app.kubernetes.io/name=ibm-spectrum-scale-csi-operator" |  awk '{print $1}'`; do
  echo "$klog StatefulSet/${statefulSetName}"
  $klog StatefulSet/"${statefulSetName}" > "${logdir}"/"${statefulSetName}".log 2>&1 || :
  $cmd describe --namespace "$ns" StatefulSet/"${statefulSetName}" > "${logdir}"/"${statefulSetName}" 2>&1 || :
done

# kubectl logs on operator pods
operatorName=$($cmd get deployment ibm-spectrum-scale-csi-operator  --namespace "$ns"  | grep -v NAME | awk '{print $1}')
if [[ "$operatorName" == "ibm-spectrum-scale-csi-operator" ]]; then
   describeCSIScaleOperator="$cmd describe CSIScaleOperator --namespace $ns"
   echo "$describeCSIScaleOperator"
   $describeCSIScaleOperator > "${describe_CSIScaleOperator}" 2>&1 || :
 fi

# kubectl logs on csi pods
for opPodName in $($cmd get pods --no-headers --namespace "$ns" -l app.kubernetes.io/name=ibm-spectrum-scale-csi-operator | awk '{print $1}'); do
  echo "$klog pod/${opPodName}"
  $klog pod/"${opPodName}" --all-containers  > "${logdir}"/"${opPodName}".log 2>&1 || :
  $klog pod/"${opPodName}" --all-containers  --previous > "${logdir}"/"${opPodName}"-previous.log 2>&1 || :
done

describe_label_cmd="$cmd describe all,cm,secret,storageclass,pvc,ds,serviceaccount -l product=${CSI_SPECTRUM_SCALE_LABEL} --namespace $ns"
echo "$describe_label_cmd"
$describe_label_cmd > "$describe_all_per_label" 2>&1 || :

describe_clusterroles="$cmd describe clusterroles/external-provisioner-runner clusterrolebindings/csi-provisioner-role clusterroles/external-attacher-runner clusterrolebindings/csi-provisioner-role clusterroles/csi-nodeplugin clusterrolebindings/csi-nodeplugin clusterroles/ibm-spectrum-scale-csi-snapshotter clusterrolebindings/ibm-spectrum-scale-csi-snapshotter clusterroles/snapshot-controller-runner clusterrolebindings/snapshot-controller-role --namespace $ns"
echo "$describe_clusterroles"
$describe_clusterroles >> "$describe_all_per_label" 2>&1 || :

get_label_cmd="$cmd get all,cm,secret,storageclass,pvc,ds,serviceaccount --namespace $ns -l product=${CSI_SPECTRUM_SCALE_LABEL}"
echo "$get_label_cmd"
$get_label_cmd > "$get_all_per_label" 2>&1 || :

get_label_cmd="$cmd get pod --namespace $ns -o wide  -l product=${CSI_SPECTRUM_SCALE_LABEL}"
echo "$get_label_cmd"
$get_label_cmd >> "$get_all_per_label" 2>&1 || :

get_configmap_cmd="$cmd get configmap spectrum-scale-config --namespace $ns -o yaml"
echo "$get_configmap_cmd"
$get_configmap_cmd > "$get_configmap" 2>&1 || :

get_k8snodes_cmd="$cmd get nodes"
echo "$get_k8snodes_cmd"
$get_k8snodes_cmd > "$get_k8snodes" 2>&1 || :

get_k8snodes_cmd="$cmd describe nodes"
echo "$get_k8snodes_cmd"
$get_k8snodes_cmd >> "$get_k8snodes" 2>&1 || :

get_spectrum_cmd="$cmd describe ds -l app.kubernetes.io/name=ibm-spectrum-scale-csi-operator -n $ns"
echo "$get_spectrum_cmd"
$get_spectrum_cmd >> "$get_daemonset" 2>&1 || :

if [[ "$cmd" == "oc" ]]
then
   get_scc_cmd="$cmd describe scc spectrum-scale-csiaccess"
   echo "$get_scc_cmd"
   $get_scc_cmd > "${logdir}"/${PRODUCT_NAME}-scc.log 2>&1 || :
fi

get_clusterinfo_cmd="$cmd cluster-info dump --namespaces kube-system --output-directory=$logdir"
echo "$get_clusterinfo_cmd"
$get_clusterinfo_cmd &>/dev/null

echo "Finished collecting \"$PRODUCT_NAME\" logs in the folder -> $logdir"
