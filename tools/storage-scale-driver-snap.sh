#!/bin/bash
#
# Copyright 2023 IBM Corp.
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

function find_versions()
{

  ns=$1
  cmd=$2

  operator_image="null"
  driver_image="null"
  registrar_image="null"
  liveness_probe_image="null"
  attacher_image="null"
  provisioner_image="null"
  snapshotter_image="null"
  resizer_image="null"

  driver_pod=`$cmd -n $ns get pod -l app=ibm-spectrum-scale-csi -o jsonpath="{.items[0].metadata.name}"`
  operator_pod=`$cmd -n $ns get pod -l name=ibm-spectrum-scale-csi-operator -o jsonpath="{.items[0].metadata.name}"`
  attacher_pod=`$cmd -n $ns get pod -l app=ibm-spectrum-scale-csi-attacher -o jsonpath="{.items[0].metadata.name}"`
  provisioner_pod=`$cmd -n $ns get pod -l app=ibm-spectrum-scale-csi-provisioner -o jsonpath="{.items[0].metadata.name}"`
  snapshotter_pod=`$cmd -n $ns get pod -l app=ibm-spectrum-scale-csi-snapshotter -o jsonpath="{.items[0].metadata.name}"`
  resizer_pod=`$cmd -n $ns get pod -l app=ibm-spectrum-scale-csi-resizer -o jsonpath="{.items[0].metadata.name}"`

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
    liveness_probe_image=`$cmd -n $ns get pod $driver_pod -o jsonpath='{.status.containerStatuses[?(@.name=="liveness-probe")].image}'`
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
    provisioner_image=`$cmd -n $ns get pod $provisioner_pod -o jsonpath='{.status.containerStatuses[?(@.name=="ibm-spectrum-scale-csi-provisioner")].image}'`
  fi

  #get snapshotter image
  if [[ $snapshotter_pod != ibm-spectrum-scale-csi-snapshotter* ]]
  then
    echo "ibm-spectrum-scale-csi snapshotter pod is not running in namespace $ns. Can't extract snapshotter version."
  else
    snapshotter_image=`$cmd -n $ns get pod $snapshotter_pod -o jsonpath='{.status.containerStatuses[?(@.name=="ibm-spectrum-scale-csi-snapshotter")].image}'`
  fi

  #get resizer image
  if [[ $resizer_pod != ibm-spectrum-scale-csi-resizer* ]]
  then
    echo "ibm-spectrum-scale-csi resizer pod is not running in namespace $ns. Can't extract resizer version."
  else
    resizer_image=`$cmd -n $ns get pod $resizer_pod -o jsonpath='{.status.containerStatuses[?(@.name=="ibm-spectrum-scale-csi-resizer")].image}'`
  fi


  #print collected data
  echo "Operator Image                : $operator_image"
  echo "Driver Image                  : $driver_image"
  echo "Node Registrar Image          : $registrar_image"
  echo "Liveness Probe Image          : $liveness_probe_image"
  echo "Attacher Image                : $attacher_image"
  echo "Provisioner Image             : $provisioner_image"
  echo "Snapshotter Image             : $snapshotter_image"
  echo "Resizer Image                 : $resizer_image"

}

function collect_csi_pod_logs()
{
  ns=$1
  cmd=$2
  since=$3
  previous=$4
  specific_logs=$5
  csi_pod_logs=${logdir}/namespaces/${ns}/pod/
  klog="$cmd logs --namespace $ns"

  if [[ $specific_logs -eq 1 ]]
  then
    nodes=""
    sidecar_pods=""
    # Find nodes on which sidecar pods are running
    for pod in $($cmd get pods --no-headers --namespace "$ns" -l type=sidecar -o jsonpath='{range .items[*]}{.metadata.name},{.spec.nodeName}{"\n"}{end}' ); do
      sidecar_pods+="$(echo $pod | cut -d ',' -f 1) "
      node=$(echo $pod | cut -d ',' -f 2)
      if !(echo "$nodes" | grep -q -E "$node"); then 
        nodes+="$node "
      fi
      echo $nodes
    done 

    # Find driver pods scheduled on nodes on which sidecar pods are running
    driver_pods=""
    for pod in $($cmd get pods --no-headers --namespace "$ns" -l app=ibm-spectrum-scale-csi -o jsonpath='{range .items[*]}{.metadata.name},{.spec.nodeName}{"\n"}{end}' ); do
      node_name=$(echo "$pod" | cut -d "," -f 2)
      if (echo "$nodes" | grep -q -E "$node_name"); then
        driver_pod=$(echo "$pod" | cut -d "," -f 1)
        driver_pods+="$driver_pod " 
      fi
      echo $driver_pods
    done 

    # Get operator pod name 
    operator_pod=$($cmd get pods --no-headers --namespace "$ns" -l name=ibm-spectrum-scale-csi-operator | awk '{print $1}')

    # kubectl logs on specific csi driver, sidecar and operator pods
    opPodNames="$sidecar_pods $driver_pods $operator_pod"
  else
    # kubectl logs on all csi driver and operator pods
    opPodNames=$($cmd get pods --no-headers --namespace "$ns" -l app.kubernetes.io/name=ibm-spectrum-scale-csi-operator | awk '{print $1}')
  fi
  for opPodName in $opPodNames; do
    echo "Gather data for pod/${opPodName}"
    for containerName in $($cmd get pods "$opPodName" --namespace "$ns" -o jsonpath="{.spec.containers[*].name}"); do
      mkdir -p "$csi_pod_logs"/"${opPodName}"/"${containerName}"
      if [[ $since != "" ]]
      then
        $klog pod/"${opPodName}" -c ${containerName} --since "$since" > "$csi_pod_logs"/"${opPodName}"/"${containerName}"/"${opPodName}"-"${containerName}".log 2>&1 || :
      else
        $klog pod/"${opPodName}" -c ${containerName} > "$csi_pod_logs"/"${opPodName}"/"${containerName}"/"${opPodName}"-"${containerName}".log 2>&1 || :
      fi
      if [[ $previous != "False" ]]
      then
        echo "Gather data for pod/${opPodName} --previous "
        $klog pod/"${opPodName}" -c ${containerName} > "$csi_pod_logs"/"${opPodName}"/"${containerName}"/"${opPodName}"-"${containerName}"-previous.log 2>&1 || :
      fi
    done
  done
}


function get_kind()
{

  ns=$1
  cmd=$2

  cluster_scoped_kinds=(storageclass clusterroles clusterrolebindings nodes pv volumeattachment csinodes )
  namespace_kinds=(pod secret configmap daemonset serviceaccount deployment events CSIScaleOperator )
  namespace_kind_log=${logdir}/namespaces/${ns}

  for kind in ${namespace_kinds[@]}
  do
    echo "Gather data for kind $kind..."
    mkdir -p "${namespace_kind_log}"/"${kind}"
    $cmd get $kind --namespace $ns > "${namespace_kind_log}"/"${kind}"/"${kind}"  2>&1 || :
    $cmd describe $kind --namespace $ns  > "${namespace_kind_log}"/"${kind}"/"${kind}".yaml 2>&1 || :
  done

  cluster_scoped_kind_log=${logdir}/cluster-scoped-resources

  for kind in ${cluster_scoped_kinds[@]}
  do
    echo "Gather data for kind $kind..."
    mkdir -p "${cluster_scoped_kind_log}"/"${kind}"
    $cmd get $kind  > "${cluster_scoped_kind_log}"/"${kind}"/"${kind}"  2>&1 || :
    $cmd describe $kind  > "${cluster_scoped_kind_log}"/"${kind}"/"${kind}".yaml 2>&1 || :
  done

}

function help()
{
   # Display Help
   echo "USAGE: storage-scale-driver-snap.sh [-l|n|o|p|s|v|h]"
   echo "options:"
   echo "     l     Collect logs only from driver pods which are running along with sidecars"
   echo "     n     CSI driver plugin namespace"
   echo "     o     output-dir"
   echo "     p     previous[=True]: If False, does not collect the logs for the previous instance of the container in a pod"
   echo "     s     Only return logs newer than a relative duration like 2h, or 4d. Defaults to all logs"
   echo "     v     Print CSI version"
   echo "     h     Print Help"
}

ns="ibm-spectrum-scale-csi-driver"
outdir="."
cmd="kubectl"
version_flag=0
previous="True"
since="0s"

# use oc commands on openshift cluster
if (which oc &>/dev/null)
then
  if (oc status &>/dev/null)
  then
    cmd="oc"
    ns="ibm-spectrum-scale-csi"
  fi
fi

while getopts 'ln:o:p:s:vh' OPTION; do
  case "$OPTION" in
    l)
      specific_logs=1
      ;;
    n)
      ns="$OPTARG"
      ;;

    o)
      outdir="$OPTARG"
      ;;
    p)
      previous="$OPTARG"
      ;;

    s)
      since="$OPTARG"
      ;;

    v)
      version_flag=1
      ;;

    h)
      help
      exit 0
      ;;

    ?)
      help
      exit 1
      ;;
  esac
done

if [[ $outdir != "." && ! -d $outdir ]]
then
  echo "Output directory $outdir does not exist. "
  exit 1
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

time=$(date +"%m-%d-%Y-%T"| sed 's/://g')
logdir=${outdir%/}/ibm-spectrum-scale-csi-logs_$time

mkdir "$logdir"
CSI_SPECTRUM_SCALE_LABEL="ibm-spectrum-scale-csi"
PRODUCT_NAME="ibm-spectrum-scale-csi"

echo "Collecting \"$PRODUCT_NAME\" logs..."
echo "The log files will be saved in the folder [$logdir]"

get_version_images=${logdir}/version
find_versions $ns $cmd >> "$get_version_images" 2>&1 || :
get_kind $ns $cmd
collect_csi_pod_logs  $ns $cmd $since $previous $specific_logs

get_clusterinfo_cmd="$cmd cluster-info dump --namespaces kube-system --output-directory=$logdir"
echo "$get_clusterinfo_cmd"
$get_clusterinfo_cmd &>/dev/null

echo "Finished collecting \"$PRODUCT_NAME\" logs in the folder -> $logdir"
