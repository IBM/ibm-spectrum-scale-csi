#!/bin/bash

# Usage: ./pv_migration_csi_to cnsa.bash /var/mnt/remote-sample

# Migrates Spectrum Scale CSI PVs to updated volumeHandle paths

set -euo pipefail

# --- Help Function ---
help() {
  echo ""
  echo "Usage: $0 <NEW_PATH_PREFIX>"
  echo ""
  echo "Example:"
  echo "  $0 /var/mnt/remote-sample"
  echo ""
  echo "Description:"
  echo "  This script migrates Storage Scale CSI PVs to use the new path prefix based on the new volumeHandle."
  echo "  It backs up PV/PVC definitions, updates the volumeHandle path, and recreates PVs/PVCs."
  echo ""
  exit 1
}

if [[ $# -ne 1 || "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  help
fi

# OLD_PATH_PREFIX="$1"
NEW_PATH_PREFIX="$1"

# --- Initialize Counters and Lists ---
success_list=()
fail_list=()
skip_list=()

success_count=0
fail_count=0
skip_count=0

main() {
  # echo "Using old path prefix: $OLD_PATH_PREFIX"
  echo "Using new path prefix: $NEW_PATH_PREFIX"
  init
  get_pv_list
  for PV in $ALL_PVS; do
    migrate_each "$PV"
  done
  final_summary
}

init() {
  LOG_FILE_DIR="migration_logs"
  mkdir -p "$LOG_FILE_DIR"
  LOG_FILE="$LOG_FILE_DIR/migration_$(date +%Y%m%d_%H%M%S).log"

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
  echo "--------------------------------------------------------------------------------"
  echo "Processing PV: $PV"

  VOLUME_HANDLE=$(kubectl get pv "$PV" -o jsonpath='{.spec.csi.volumeHandle}')
  echo "   Current volumeHandle: $VOLUME_HANDLE"

  if [[ "$VOLUME_HANDLE" == *"$NEW_PATH_PREFIX"* ]]; then
    echo "Already migrated (volumeHandle contains new path). Skipping $PV"
    ((skip_count++))
    skip_list+=("$PV")
    return
  fi

  PVC_NAME=$(kubectl get pv "$PV" -o jsonpath='{.spec.claimRef.name}')
  PVC_NAMESPACE=$(kubectl get pv "$PV" -o jsonpath='{.spec.claimRef.namespace}')

  echo "Preparing to migrate PVC: $PVC_NAME in namespace $PVC_NAMESPACE"

  if ! kubectl get pvc "$PVC_NAME" -n "$PVC_NAMESPACE" >/dev/null 2>&1; then
    echo "PVC $PVC_NAME not found in namespace $PVC_NAMESPACE, skipping."
    fail_list+=("$PV")
    ((fail_count++))
    return
  fi

  # Extract key PV/PVC attributes
  ACCESS_MODES=$(kubectl get pvc "$PVC_NAME" -n "$PVC_NAMESPACE" -o json | jq '.spec.accessModes')
  STORAGE=$(kubectl get pvc "$PVC_NAME" -n "$PVC_NAMESPACE" -o json | jq -r '.spec.resources.requests.storage')
  STORAGE_CLASS=$(kubectl get pvc "$PVC_NAME" -n "$PVC_NAMESPACE" -o jsonpath='{.spec.storageClassName}')
  VOLUME_MODE=$(kubectl get pvc "$PVC_NAME" -n "$PVC_NAMESPACE" -o jsonpath='{.spec.volumeMode}')
  ATTRS=$(kubectl get pv "$PV" -o json | jq '.spec.csi.volumeAttributes')


  DRIVER=$(kubectl get pv "$PV" -o jsonpath='{.spec.csi.driver}')
  FSTYPE=$(kubectl get pv "$PV" -o jsonpath='{.spec.csi.fsType}')

  # Construct new volumeHandle with updated path
  OLD_PATH=$(echo "$VOLUME_HANDLE" | awk -F';' '{print $NF}')
  # NEW_PATH=$(echo "$OLD_PATH" | sed "s|$OLD_PATH_PREFIX|$NEW_PATH_PREFIX|")
  NEW_PATH="$NEW_PATH_PREFIX/$PV/$PV-data"
  NEW_VOLUME_HANDLE=$(echo "$VOLUME_HANDLE" | sed "s|$OLD_PATH|$NEW_PATH|")

  echo "Updating volumeHandle to: $NEW_VOLUME_HANDLE"

  # Set PV reclaim policy to Retain
  kubectl patch pv "$PV" --type=merge -p '{"spec": {"persistentVolumeReclaimPolicy": "Retain"}}'

  # Backup original PV/PVC
  BACKUP_DIR="backup_pv_pvc/${PVC_NAME}"
  mkdir -p "$BACKUP_DIR"

  echo "Taking backup of PV ${PV}"
  kubectl get pv "$PV" -o yaml > "$BACKUP_DIR/$PV.yaml"

  echo "Taking backup of PVC ${PVC_NAME}"
  kubectl get pvc "$PVC_NAME" -n "$PVC_NAMESPACE" -o yaml > "$BACKUP_DIR/$PVC_NAME.yaml"

  echo "Backups stored under: $BACKUP_DIR"

  echo "Deleting PVC and PV..."
  kubectl delete pvc "$PVC_NAME" -n "$PVC_NAMESPACE"
  kubectl delete pv "$PV"

  echo "Recreating PV $PV with updated volumeHandle..."

  cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: PersistentVolume
metadata:
  name: $PV
spec:
  capacity:
    storage: $STORAGE
  accessModes: $(echo "$ACCESS_MODES" | jq '.')
  persistentVolumeReclaimPolicy: Delete
  storageClassName: $STORAGE_CLASS
  csi:
    driver: $DRIVER
    fsType: $FSTYPE
    volumeHandle: "$NEW_VOLUME_HANDLE"
    volumeAttributes: $(echo "$ATTRS" | jq '.')
  volumeMode: $VOLUME_MODE
EOF

  echo "Recreating PVC $PVC_NAME in $PVC_NAMESPACE"

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
  echo "--------------------------------------------------------------------------------"
  ((success_count++))
  success_list+=("$PV")
}

final_summary() {
  echo ""
  echo "Migration Summary:"
  echo "----------------------------"

  if (( ${#success_list[@]} > 0 )); then
    echo "Successful PVs: ${#success_list[@]}"
    for pv in "${success_list[@]}"; do echo "   - $pv"; done
    echo ""
  fi

  if (( ${#fail_list[@]} > 0 )); then
    echo "Failed PVs: ${#fail_list[@]}"
    for pv in "${fail_list[@]}"; do echo "   - $pv"; done
    echo ""
  fi

  if (( ${#skip_list[@]} > 0 )); then
    echo "Skipped PVs (already migrated): ${#skip_list[@]}"
    for pv in "${skip_list[@]}"; do echo "   - $pv"; done
    echo ""
  fi

  echo "Completed migration at: $(date)"
}

main

exit 0
