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

function find_versions()
{

  ns=$1
  cmd=$2

  operator_image="null"
  csi_version="null"
  driver_image="null"
  registrar_image="null"
  attacher_image="null"
  provisioner_image="null"
  snapshotter_image="null"

  driver_pod=`$cmd -n $ns get pod -l app=ibm-spectrum-scale-csi -o jsonpath="{.items[0].metadata.name}"`
  operator_pod=`$cmd -n $ns get pod -l name=ibm-spectrum-scale-csi-operator -o jsonpath="{.items[0].metadata.name}"`
  attacher_pod=`$cmd -n $ns get pod -l app=ibm-spectrum-scale-csi-attacher -o jsonpath="{.items[0].metadata.name}"`
  provisioner_pod=`$cmd -n $ns get pod -l app=ibm-spectrum-scale-csi-provisioner -o jsonpath="{.items[0].metadata.name}"`
  snapshotter_pod=`$cmd -n $ns get pod -l app=ibm-spectrum-scale-csi-snapshotter -o jsonpath="{.items[0].metadata.name}"`

  #get operator image
  if [[ $operator_pod != ibm-spectrum-scale-csi-operator* ]]
  then
    echo "ibm-spectrum-scale-csi operator pod is not running in namespace $ns. Can't extract operator version."
  else
    operator_image=`$cmd -n $ns get pod $operator_pod -o jsonpath='{.status.containerStatuses[?(@.name=="operator")].image}'`
  fi

  #get driver image, registrar image and csi version from driver image
  if [[ $driver_pod != ibm-spectrum-scale-csi-* ]]
  then
    echo "ibm-spectrum-scale-csi driver pod is not running in namespace $ns. Can't extract driver and registrar version."
  else
    driver_image=`$cmd -n $ns get pod $driver_pod -o jsonpath='{.status.containerStatuses[?(@.name=="ibm-spectrum-scale-csi")].image}'`
    registrar_image=`$cmd -n $ns get pod $driver_pod -o jsonpath='{.status.containerStatuses[?(@.name=="driver-registrar")].image}'`
    csi_version=`echo $driver_image | cut -f2 -d:`
  fi

  #get attacher image
  if [[ $attacher_pod != ibm-spectrum-scale-csi-attacher* ]]
  then
    echo "ibm-spectrum-scale-csi attacher pod is not running in namespace $ns. Can't extract attacher version."
  else
    attacher_image=`$cmd -n $ns get pod $attacher_pod -o jsonpath='{.status.containerStatuses[?(@.name=="ibm-spectrum-scale-csi-attacher")].image}'`
  fi

  #get provisioner image
  if [[ $provisioner_pod != ibm-spectrum-scale-csi-provisioner* ]]
  then
    echo "ibm-spectrum-scale-csi provisioner pod is not running in namespace $ns. Can't extract provisioner version."
  else
    provisioner_image=`$cmd -n $ns get pod $provisioner_pod -o jsonpath='{.status.containerStatuses[?(@.name=="csi-provisioner")].image}'`
  fi

  #get snapshotter image
  if [[ $snapshotter_pod != ibm-spectrum-scale-csi-snapshotter* ]]
  then
    echo "ibm-spectrum-scale-csi snapshotter pod is not running in namespace $ns. Can't extract snapshotter version."
  else
    snapshotter_image=`$cmd -n $ns get pod $snapshotter_pod -o jsonpath='{.status.containerStatuses[?(@.name=="csi-snapshotter")].image}'`
  fi

  #print collected data
  echo "IBM Spectrum Scale CSI driver : $csi_version"
  echo "Operator Image                : $operator_image"
  echo "Driver Image                  : $driver_image"
  echo "Node Registrar Image          : $registrar_image"
  echo "Attacher Image                : $attacher_image"
  echo "Provisioner Image             : $provisioner_image"
  echo "Snapshotter Image             : $snapshotter_image"

}

ns="ibm-spectrum-scale-csi-driver"
outdir="."
cmd="kubectl"
version_flag=0

while getopts 'n:o:vh' OPTION; do
  case "$OPTION" in
    n)
      ns="$OPTARG"
      ;;

    o)
      outdir="$OPTARG"
      ;;

    v)
      version_flag=1
      ;;

    h)
      echo "USAGE: spectrum-scale-driver-snap.sh [-n namespace] [-o output-dir] [-v] [-h]"
      exit 0
      ;;

    ?)
      echo "USAGE: spectrum-scale-driver-snap.sh [-n namespace] [-o output-dir] [-v] [-h]"
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
if !($cmd get namespace | grep -q "$ns\s*Active")
then
  echo "Namespace $ns is invalid or not active. Please provide a valid namespace"
  exit 1
fi

if [[ $version_flag -eq 1 ]]
then
 find_versions $ns $cmd
 exit 0
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
get_version_images=${logdir}/ibm-spectrum-scale-csi-versions

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

describe_label_cmd="$cmd describe all,cm,secret,storageclass,pvc,ds,serviceaccount,clusterroles,clusterrolebindings -l product=${CSI_SPECTRUM_SCALE_LABEL} --namespace $ns"
echo "$describe_label_cmd"
$describe_label_cmd > "$describe_all_per_label" 2>&1 || :

get_label_cmd="$cmd get all,cm,secret,storageclass,pvc,serviceaccount,clusterroles,clusterrolebindings --namespace $ns -l product=${CSI_SPECTRUM_SCALE_LABEL}"
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

find_versions $ns $cmd >> "$get_version_images" 2>&1 || :

get_clusterinfo_cmd="$cmd cluster-info dump --namespaces kube-system --output-directory=$logdir"
echo "$get_clusterinfo_cmd"
$get_clusterinfo_cmd &>/dev/null

echo "Finished collecting \"$PRODUCT_NAME\" logs in the folder -> $logdir"
