#!/bin/bash

# Usage: ./migration_csi_primary_removal.bash
# Migrate existing Kubernetes PersistentVolumes (PVs) that were originally created when the primary filesystem and fileset was enabled, to a format that uses the actual fileset mount path after the primary filesystem has been removed.

set -euo pipefail

# --- Help Function ---
help() {
  echo ""
  echo "Usage: $0"
  echo ""
  echo "Description:"
  echo "  This script migrate existing Kubernetes PersistentVolumes (PVs) that were originally created when the primary filesystem and fileset was enabled, to a format that uses the actual fileset mount path after the primary filesystem has been removed."
  echo ""
  exit 1
}

# --- Argument Parsing ---
while [[ $# -gt 0 ]]; do
  case "$1" in
    -h|--help)
      help
      ;;
    *)
      echo "Unknown argument: $1"
      help
      ;;
  esac
  shift
done

# --- Initialize Counters and Lists ---
success_list=()
fail_list=()
skip_list=()

success_count=0
fail_count=0
skip_count=0

main() {
  init
  print_start_banner
  check_prerequisites
  collect_fs_prefixes
  echo "Starting migration of IBM Storage Scale CSI PersistentVolumes that were originally created when the primary filesystem and fileset was enabled, to a format that uses the actual fileset mount path after the primary filesystem has been removed."
  echo ""
  read -rp "Proceed with migration? (yes/y/Y to continue): " CONFIRM
  if [[ "$CONFIRM" != "yes" && "$CONFIRM" != "y" && "$CONFIRM" != "Y" ]]; then
    echo "Aborting migration."
    exit 0
  fi
  get_pv_list
  TOTAL_PVS=$(echo "$ALL_PVS" | wc -l | xargs)
  COUNT=1
  for PV in $ALL_PVS; do
    echo ""
    echo "[$COUNT/$TOTAL_PVS]  --------------------------------------------------------------------------------"
    echo "Processing PV: $PV    "
    migrate_each "$PV"
    COUNT=$((COUNT + 1))
  done
  final_summary
}

print_start_banner() {
  echo "======================================================================================================"
  echo "Starting Migration Script – IBM Storage Scale CSI (Primary filesystem mount path to Actual fileset mount path)"
  echo "======================================================================================================"
  echo ""
  echo "This script will:"
  echo "  • Collect all PersistentVolumes (PVs) with Storage Scale CSI"
  echo "  • Identify unique backend filesystems (backendFs)"
  echo "  • Ask you for NEW_PATH_PREFIX for each backend FS"
  echo "  • Rewrite volumeHandle paths accordingly"
  echo ""
  echo "Please ensure you have 'kubectl' and 'jq' installed and"
  echo "that you are logged into the correct Kubernetes cluster."
  echo "======================================================================================================"
  echo ""
}

check_prerequisites() {
  echo "Checking prerequisites..."
  for cmd in kubectl jq; do
    if ! command -v $cmd &>/dev/null; then
        echo "Error: '$cmd' is required but not installed or not in PATH." >&2
        exit 1
    fi
  done
  echo "All prerequisites met."
}

collect_fs_prefixes() {
  while true; do
    VOL_BACKEND_FS_LIST=()
    VOL_BACKEND_PREFIX_LIST=()

    echo "Collecting all unique backend filesystems from PVs..."
    ALL_FS=$(kubectl get pv -o json \
      | jq -r '.items[]
              | select(.spec.csi.driver == "spectrumscale.csi.ibm.com")
              | .spec.csi.volumeAttributes.volBackendFs // empty' \
      | grep -v '^$' | sort -u)

    for fs in $ALL_FS; do
      echo ""
      echo "Backend FS: $fs"
      # Show all PVs belonging to this FS
      kubectl get pv -o json | jq -r --arg FS "$fs" \
        '.items[] | select(.spec.csi.volumeAttributes.backendFs==$FS)
         | "  PV: \(.metadata.name)\tPVC: \(.spec.claimRef.name)"'

      echo ""
      read -rp "Enter NEW_PATH_PREFIX for backend FS '$fs': " prefix
      VOL_BACKEND_FS_LIST+=("$fs")
      VOL_BACKEND_PREFIX_LIST+=("$prefix")
    done

    echo ""
    echo "Collected FS prefix mappings:"
    for i in "${!VOL_BACKEND_FS_LIST[@]}"; do
      echo "  ${VOL_BACKEND_FS_LIST[$i]} => ${VOL_BACKEND_PREFIX_LIST[$i]}"
    done
    echo ""

    read -rp "Are these mappings of volumeBackendFS and it's NEW_PATH_PREFIX correct? (y/n): " confirm
    case "$confirm" in
      y|Y|yes|YES)
        echo "Prefix mappings confirmed."
        break
        ;;
      *)
        echo "Let's try again..."
        ;;
    esac
  done
}

# Helper to retrieve prefix for given FS
get_prefix_for_fs() {
  local search_fs=$1
  for i in "${!VOL_BACKEND_FS_LIST[@]}"; do
    if [[ "${VOL_BACKEND_FS_LIST[$i]}" == "$search_fs" ]]; then
      echo "${VOL_BACKEND_PREFIX_LIST[$i]}"
      return
    fi
  done
  echo ""  # fallback if not found
}

init() {
  RUN_TIMESTAMP=$(date +%Y%m%d_%H%M%S)

  BACKUP_PARENT_DIR="csi_migration_data"
  mkdir -p "$BACKUP_PARENT_DIR"

  BACKUP_BASE_DIR="$BACKUP_PARENT_DIR/$RUN_TIMESTAMP"
  mkdir -p "$BACKUP_BASE_DIR"

  LOG_FILE="$BACKUP_BASE_DIR/migration.log"
  exec > >(tee "$LOG_FILE") 2>&1

  echo "Logging to $LOG_FILE"
  echo "Starting migration at: $(date)"
  echo ""
}

get_pv_list() {
  ALL_PVS=$(kubectl get pv -o json | jq -r '.items[] | select(.spec.csi.driver == "spectrumscale.csi.ibm.com") | .metadata.name')
  echo "Found $(echo "$ALL_PVS" | wc -l) PVs using spectrumscale.csi.ibm.com driver"
}

migrate_each() {
  PV="$1"

  VOLUME_HANDLE=$(kubectl get pv "$PV" -o jsonpath='{.spec.csi.volumeHandle}') || {
    echo "Failed to get volumeHandle for PV: $PV"; fail_count=$(expr $fail_count + 1); fail_list+=("$PVC_NAME|$PV"); return;
  }
  echo "Current volumeHandle: $VOLUME_HANDLE"

  PVC_NAME=$(kubectl get pv "$PV" -o jsonpath='{.spec.claimRef.name}') || {
    echo "Failed to get PVC name for PV: $PV"; fail_count=$(expr $fail_count + 1); fail_list+=("|$PV"); return;
  }
  PVC_NAMESPACE=$(kubectl get pv "$PV" -o jsonpath='{.spec.claimRef.namespace}') || {
    echo "Failed to get PVC namespace for PV: $PV"; fail_count=$(expr $fail_count + 1); fail_list+=("$PVC_NAME|$PV"); return;
  }

  echo "Preparing to migrate PVC: $PVC_NAME in namespace $PVC_NAMESPACE"

  if ! kubectl get pvc "$PVC_NAME" -n "$PVC_NAMESPACE" >/dev/null 2>&1; then
    echo "PVC $PVC_NAME not found in namespace $PVC_NAMESPACE, skipping."
    fail_list+=("$PVC_NAME|$PV")
    fail_count=$(expr $fail_count + 1)
    return
  fi

  ACCESS_MODES=$(kubectl get pvc "$PVC_NAME" -n "$PVC_NAMESPACE" -o json | jq '.spec.accessModes') || {
    echo "Failed to get access modes."; fail_count=$(expr $fail_count + 1); fail_list+=("$PVC_NAME|$PV"); return;
  }
  STORAGE=$(kubectl get pvc "$PVC_NAME" -n "$PVC_NAMESPACE" -o json | jq -r '.spec.resources.requests.storage') || {
    echo "Failed to get storage size."; fail_count=$(expr $fail_count + 1); fail_list+=("$PVC_NAME|$PV"); return;
  }
  STORAGE_CLASS=$(kubectl get pvc "$PVC_NAME" -n "$PVC_NAMESPACE" -o jsonpath='{.spec.storageClassName}') || {
    echo "Failed to get storage class."; fail_count=$(expr $fail_count + 1); fail_list+=("$PVC_NAME|$PV"); return;
  }
  VOLUME_MODE=$(kubectl get pvc "$PVC_NAME" -n "$PVC_NAMESPACE" -o jsonpath='{.spec.volumeMode}') || {
    echo "Failed to get volume mode."; fail_count=$(expr $fail_count + 1); fail_list+=("$PVC_NAME|$PV"); return;
  }
  ATTRS=$(kubectl get pv "$PV" -o json | jq '.spec.csi.volumeAttributes') || {
    echo "Failed to get volumeAttributes."; fail_count=$(expr $fail_count + 1); fail_list+=("$PVC_NAME|$PV"); return;
  }
  VOL_BACKEND_FS=$(kubectl get pv "$PV" -o json | jq -r '.spec.csi.volumeAttributes.volBackendFs') || {
    echo "Failed to get volBackendFs."; fail_count=$(expr $fail_count + 1); fail_list+=("$PVC_NAME|$PV"); return;
  }
  DRIVER=$(kubectl get pv "$PV" -o jsonpath='{.spec.csi.driver}') || {
    echo "Failed to get driver."; fail_count=$(expr $fail_count + 1); fail_list+=("$PVC_NAME|$PV"); return;
  }
  FSTYPE=$(kubectl get pv "$PV" -o jsonpath='{.spec.csi.fsType}') || {
    echo "Failed to get fsType."; fail_count=$(expr $fail_count + 1); fail_list+=("$PVC_NAME|$PV"); return;
  }
  RECLAIM_POLICY_ORIGINAL=$(kubectl get pv "$PV" -o jsonpath='{.spec.persistentVolumeReclaimPolicy}') || {
    echo "Failed to get reclaimPolicy for PV $PV."; fail_count=$(expr $fail_count + 1); fail_list+=("$PVC_NAME|$PV"); return;
  }

  # Get prefix from stored list for volume backend FS
  NEW_PATH_PREFIX=$(get_prefix_for_fs "$VOL_BACKEND_FS")



  if [[ -z "$NEW_PATH_PREFIX" ]]; then
    echo "No prefix defined for $VOL_BACKEND_FS, skipping migration"
    skip_count=$(expr $skip_count + 1)
    skip_list+=("$PVC_NAME|$PV")
    return
  else
    # Remove trailing slash if present
    NEW_PATH_PREFIX="${NEW_PATH_PREFIX%/}"
    echo "Using prefix for $VOL_BACKEND_FS --> $NEW_PATH_PREFIX"
  fi

  OLD_PATH=$(echo "$VOLUME_HANDLE" | awk -F';' '{print $NF}')

  IFS=';' read -ra parts <<< "$VOLUME_HANDLE"
  last_index=$(( ${#parts[@]} - 1 ))
  # Determine volume type based on volumeHandle format and frame NEW_PATH
  if [[ "${parts[0]}" == "0" && "${parts[1]}" == "0" ]]; then
    echo "Detected volumeHandle type: 0;0 (volDirBasePath) : Lightweight volume"
    if [[ "${parts[$last_index]}" == *"/.volumes/"* ]]; then
      VOL_DIR_BASE_PATH=$(echo "$ATTRS" | jq -r '."volDirBasePath" // empty')
      NEW_PATH="$NEW_PATH_PREFIX/$VOL_BACKEND_FS/$VOL_DIR_BASE_PATH/$PV"
    else
      echo "For this volume primary migration is not required. Skipping $PV"
      skip_count=$(expr $skip_count + 1)
      skip_list+=("$PVC_NAME|$PV")
      return
    fi

  elif [[ "${parts[0]}" == "0" && "${parts[1]}" == "1" ]]; then
    echo "Detected volumeHandle type: 0;1 (parentFileset) : Fileset volume and dependent fileset"
    if [[ "${parts[$last_index]}" == *"/.volumes/"* ]]; then

      PARENT_FILESET=$(echo "$ATTRS" | jq -r '."parentFileset" // empty')
      existingVolume=$(echo "$ATTRS" | jq -r '."existingVolume" // empty')
      VOL_DIR_BASE_PATH=$(echo "$ATTRS" | jq -r '."volDirBasePath" // empty')

      if [[ "$existingVolume" == "yes" ]]; then
        echo "Static Fileset volume and dependent fileset"
        if [[ -n "$PARENT_FILESET" && "$PARENT_FILESET" != "root" ]]; then
          NEW_PATH="$NEW_PATH_PREFIX/$VOL_BACKEND_FS/$PARENT_FILESET/$PVC_NAME"
        else
          echo "parentFileset is empty or 'root', using default path"
          NEW_PATH="$NEW_PATH_PREFIX/$VOL_BACKEND_FS/$PVC_NAME"
        fi
      else
        echo "Dynamic Fileset volume and dependent fileset"
        if [[ -n "$PARENT_FILESET" && "$PARENT_FILESET" != "root" && -n "$VOL_DIR_BASE_PATH" ]]; then
          echo "Using parentFileset with volDirBasePath"
          NEW_PATH="$NEW_PATH_PREFIX/$VOL_BACKEND_FS/$VOL_DIR_BASE_PATH/$PARENT_FILESET/$PV/$PV-data"
        elif [[ -n "$PARENT_FILESET" && "$PARENT_FILESET" != "root" ]]; then
          NEW_PATH="$NEW_PATH_PREFIX/$VOL_BACKEND_FS/$PARENT_FILESET/$PV/$PV-data"
        else
          echo "parentFileset is empty or 'root', using default path"
          NEW_PATH="$NEW_PATH_PREFIX/$VOL_BACKEND_FS/$PV/$PV-data"
        fi
      fi
    else
      echo "For this volume primary migration is not required. Skipping $PV"
      skip_count=$(expr $skip_count + 1)
      skip_list+=("$PVC_NAME|$PV")
      return
    fi

  elif [[ "${parts[0]}" == "0" && "${parts[1]}" == "2" ]]; then
    echo "Detected volumeHandle type: 0;2 (fileset) : Fileset volume and independent fileset"
    if [[ "${parts[$last_index]}" == *"/.volumes/"* ]]; then
      # Default path of dynamic fileset
      NEW_PATH="$NEW_PATH_PREFIX/$VOL_BACKEND_FS/$PV/$PV-data"

      VOL_DIR_BASE_PATH=$(echo "$ATTRS" | jq -r '."volDirBasePath" // empty')

      # Check if existingVolume is set to yes for static fileset volumes
      existingVolume=$(echo "$ATTRS" | jq -r '."existingVolume" // empty')
      if [[ "$existingVolume" == "yes" ]]; then
        echo "Static Fileset volume and independent fileset"
        NEW_PATH="$NEW_PATH_PREFIX/$VOL_BACKEND_FS/$PVC_NAME"
      elif [[ -n "$VOL_DIR_BASE_PATH" ]]; then
        echo "Dynamic Fileset volume with volDirBasePath and independent fileset"
        NEW_PATH="$NEW_PATH_PREFIX/$VOL_BACKEND_FS/$VOL_DIR_BASE_PATH/$PV/$PV-data"
      fi
    else
      echo "For this volume primary migration is not required. Skipping $PV"
      skip_count=$(expr $skip_count + 1)
      skip_list+=("$PVC_NAME|$PV")
      return
    fi

  elif [[ "${parts[0]}" == "1" && "${parts[1]}" == "1" ]]; then
    echo "Detected volumeHandle type: 1;1 (version2) : Consistency group fileset"
    echo "For Consistency group fileset migration is not required. Skipping $PV"
    skip_count=$(expr $skip_count + 1)
    skip_list+=("$PVC_NAME|$PV")
    return

  elif [[ "${parts[0]}" == "0" && "${parts[1]}" == "3" ]]; then
    echo "Detected volumeHandle type: 0;3 : Shallow copy fileset"
    echo "For Shallow copy fileset migration is not required. Skipping $PV"
    skip_count=$(expr $skip_count + 1)
    skip_list+=("$PVC_NAME|$PV")
    return

  elif [[ "${parts[0]}" == "1" && "${parts[1]}" == "3" ]]; then
    echo "Detected volumeHandle type: 1;3 version-2: Shallow copy fileset"
    echo "For Shallow copy fileset migration is not required. Skipping $PV"
    skip_count=$(expr $skip_count + 1)
    skip_list+=("$PVC_NAME|$PV")
    return

  else
    echo "Unknown volumeHandle type: ${parts[0]};${parts[1]} — skipping migration for PV: $PV"
    fail_list+=("$PVC_NAME|$PV")
    fail_count=$(expr $fail_count + 1)
    return
  fi

  NEW_VOLUME_HANDLE=$(echo "$VOLUME_HANDLE" | sed "s|$OLD_PATH|$NEW_PATH|")

  if [[ "$VOLUME_HANDLE" == "$NEW_VOLUME_HANDLE" ]]; then
    echo "Already migrated (volumeHandle matches target). Skipping $PV"
    skip_count=$(expr $skip_count + 1)
    skip_list+=("$PVC_NAME|$PV")
    return
  fi

  echo "Updated volumeHandle for $PV: $NEW_VOLUME_HANDLE"

  echo "Setting reclaim policy to Retain for PV: $PV"
  if ! kubectl patch pv "$PV" --type=merge -p '{"spec": {"persistentVolumeReclaimPolicy": "Retain"}}'; then
    echo "Failed to patch reclaim policy for PV: $PV"
    fail_count=$(expr $fail_count + 1)
    fail_list+=("$PVC_NAME|$PV")
    return
  fi

  BACKUP_DIR="$BACKUP_BASE_DIR/${PVC_NAMESPACE}/${PVC_NAME}"
  mkdir -p "$BACKUP_DIR"

  echo "Backing up PV and PVC..."
  kubectl get pv "$PV" -o yaml > "$BACKUP_DIR/pv.yaml"
  kubectl get pvc "$PVC_NAME" -n "$PVC_NAMESPACE" -o yaml > "$BACKUP_DIR/pvc.yaml"

  RECLAIM_POLICY=$(kubectl get pv "$PV" -o jsonpath='{.spec.persistentVolumeReclaimPolicy}')
  if [[ "$RECLAIM_POLICY" != "Retain" ]]; then
    echo "Reclaim policy is not Retain (got: $RECLAIM_POLICY), skipping."
    fail_count=$(expr $fail_count + 1)
    fail_list+=("$PVC_NAME|$PV")
    return
  fi

  echo "Deleting PVC and PV..."
  kubectl delete pvc "$PVC_NAME" -n "$PVC_NAMESPACE"

# Remove external-attacher finalizer from PV if it exists
  finalizer="external-attacher/spectrumscale-csi-ibm-com"
  index=$(kubectl get pv "${PV}" -o json | jq -r \
  ".metadata.finalizers | to_entries | map(select(.value==\"${finalizer}\")) | .[0].key")

  if [[ -n "$index" && "$index" != "null" ]]; then
    echo "Removing ${finalizer} finalizer from PV ${PV}..."
    kubectl patch pv "${PV}" --type=json \
      -p="[ { \"op\": \"remove\", \"path\": \"/metadata/finalizers/${index}\" } ]"
  fi
  kubectl delete pv "$PV"

  echo "Recreating PV and PVC..."
  cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: PersistentVolume
metadata:
  name: $PV
spec:
  capacity:
    storage: $STORAGE
  accessModes: $(echo "$ACCESS_MODES" | jq '.')
  persistentVolumeReclaimPolicy: $RECLAIM_POLICY_ORIGINAL
  storageClassName: $STORAGE_CLASS
  csi:
    driver: $DRIVER
    fsType: $FSTYPE
    volumeHandle: "$NEW_VOLUME_HANDLE"
    volumeAttributes: $(echo "$ATTRS" | jq '.')
  volumeMode: $VOLUME_MODE
EOF

  cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: $PVC_NAME
  namespace: $PVC_NAMESPACE
spec:
  accessModes: $(echo "$ACCESS_MODES" | jq '.')
  storageClassName: $STORAGE_CLASS
  volumeMode: $VOLUME_MODE
  resources:
    requests:
      storage: $STORAGE
  volumeName: $PV
EOF

  echo "Migration successful for PV: $PV and PVC: $PVC_NAME"
  echo "----------------------------------------------------------------------------------------"
  success_count=$(expr $success_count + 1)
  success_list+=("$PVC_NAME|$PV")
}

final_summary() {
  echo ""
  echo "Migration Summary:"
  echo "----------------------------------------------------------------------------------------------------------------------------"

  if (( ${#success_list[@]} > 0 )); then
    echo "Successful PVs: ${#success_list[@]}"
    printf "   %-80s | %s\n" "PVC Name" "PV Name"
    printf "   %s\n" "----------------------------------------------------------------------------------------------------------------------------"
    for entry in "${success_list[@]}"; do
      IFS='|' read -r pvc pv <<< "$entry"
      printf "   %-80s | %s\n" "$pvc" "$pv"
    done
    echo ""
  fi

  if (( ${#fail_list[@]} > 0 )); then
    echo "Failed PVs: ${#fail_list[@]}"
    printf "   %-80s | %s\n" "PVC Name" "PV Name"
    printf "   %s\n" "----------------------------------------------------------------------------------------------------------------------------"
    for entry in "${fail_list[@]}"; do
      IFS='|' read -r pvc pv <<< "$entry"
      printf "   %-80s | %s\n" "$pvc" "$pv"
    done
    echo ""
  fi

  if (( ${#skip_list[@]} > 0 )); then
    echo "Skipped PVs (already migrated): ${#skip_list[@]}"
    printf "   %-80s | %s\n" "PVC Name" "PV Name"
    printf "   %s\n" "----------------------------------------------------------------------------------------------------------------------------"
    for entry in "${skip_list[@]}"; do
      IFS='|' read -r pvc pv <<< "$entry"
      printf "   %-80s | %s\n" "$pvc" "$pv"
    done
    echo ""
  fi

  echo "Completed migration at: $(date)"
}

main

exit 0
