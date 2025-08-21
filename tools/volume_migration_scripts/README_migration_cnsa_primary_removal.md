
# PV Migration Script – CNSA (Primary → Actual Fileset Mount)

This script helps migrate existing Kubernetes **PersistentVolumes (PVs)** in a **CNSA environment** that were originally created when the **primary fileset option** was enabled.
It updates PVs to use the **actual fileset mount path** on CNSA worker nodes after the primary has been removed, ensuring workloads continue to access their data seamlessly.

## Key Features

- Preserves all **original PV properties**, including:
  - Reclaim policy (e.g., `Retain`, `Delete`)
  - Access modes (`ReadWriteOnce`, `ReadWriteMany`, etc.)
  - Storage capacity
  - Filesystem type (`fsType`)
  - Labels, annotations, and other PV metadata

- Only the **volumeHandle path segment** is updated based on the **fileset type** and **prefix**.
- Supports migration of PVs across **different fileset types** (independent, dependent, static, cache, CG, etc.).
- Generates **backup YAML files** before applying changes.
- Can be safely re-run if required (**idempotent migration**).

## Why Migration is Required

When the **primary fileset** was enabled, PVs were created under paths tied to the **primary fileset hierarchy**.
Now that **primary is removed**, these paths are invalid:

- PVs must be mounted at the **actual fileset mount path** on CNSA worker nodes.
- Without migration, workloads would **not be able to mount or access their data**.

This script updates PV definitions to point to the correct CNSA paths while **preserving all other PV properties**.

## Prerequisites

- **Delete any existing workloads or application pods attached to the PVCs that will be migrated** from the existing CSI cluster **before installing CNSA.**

Before running the migration script, ensure the following tools are installed and available in your `$PATH`:

- The script should be run in the context of the cluster where Spectrum Scale CSI is deployed
- **kubectl** – to interact with the Kubernetes cluster and fetch/update PV/PVC objects
- **jq** – for JSON parsing and manipulation of Kubernetes API responses

You can verify installation with:

```bash
kubectl version --client
jq --version
```

## Example Transformation

### Before (PV created with primary enabled):
```text
volumeHandle: 0;2;13009550825755318848;9A7B0B0A:68891B40;;pvc-26946b2b-b18a-4c0d-9f77-606a444094c1;/ibm/remotefs1/primary-fileset-remotefs1-475592072879187/.volumes/pvc-26946b2b-b18a-4c0d-9f77-606a444094c1
```

### After (PV updated to actual fileset mount path with prefix `/var/mnt`):
```text
volumeHandle: 0;2;13009550825755318848;9A7B0B0A:68891B40;;pvc-26946b2b-b18a-4c0d-9f77-606a444094c1;/var/mnt/remotefs1/pvc-26946b2b-b18a-4c0d-9f77-606a444094c1/pvc-26946b2b-4c0d-9f77-606a444094c1-data
```

### Key Points

**Unchanged:**
- The identity portion of the handle (everything up to the last `;`).

**Rewritten:**
- Only the **path segment after the last `;`**, now using the **prefix** and correct fileset mapping for CNSA.

## Volume Type Variations

The exact path suffix (e.g., `pvc-uuid-data`) may vary depending on how the volume was originally created.
The script automatically detects and applies the correct mapping for each fileset type.


## Migration Script Usage

Run the script to automatically migrate PVs:

```bash
# Usage
./migration_cnsa_primary_removal.bash --new_path_prefix /var/mnt
```

Where:

- `--new_path_prefix` specifies the **base mount point** for all IBM Storage Scale filesystems on CNSA worker nodes (e.g., `/var/mnt`, `/ibm`, `/mnt`).

## Features of the Migration Script

- ✅ Filters only PVs created with the **spectrumscale.csi.ibm.com** driver.
- ✅ Skips PVs already migrated (with an actual fileset mount path in `volumeHandle`).
- ✅ **Backs up PVs and PVCs** before modification into a structured directory:

```
csi_migration_data/
└── <timestamp>/
    ├── migration.log
    ├── <namespace>/
    │   └── <pvc-name>/
    │       ├── pvc.yaml
    │       └── pv.yaml
    └── ...
```

- ✅ Logs all actions, successes, skips, and failures into:

```
csi_migration_data/<timestamp>/migration.log
```

- ✅ Summarizes **success, skipped, and failed** migrations at the end.
- ✅ Idempotent – safe to re-run if needed.

## Preserved PV Properties

The script ensures that **all original PV configurations** are retained after migration:

- **Capacity** (`spec.capacity.storage`)
- **AccessModes** (`spec.accessModes`)
- **PersistentVolumeReclaimPolicy** (`spec.persistentVolumeReclaimPolicy`)
- **StorageClassName**
- **CSI driver details** (fsType, volumeAttributes, nodeStageSecrets, etc.)
- **PVC binding information** (safely re-created to preserve claim references)

- Only the **`volumeHandle` path** is modified to reflect the actual fileset mount with the provided **`--new_path_prefix`**.

## Notes and Limitations

- The **filesystem names** (e.g., `remotefs1`) must remain identical between pre-primary and post-primary removal deployments.
- The provided `--new_path_prefix` must reflect the **actual base mount point** of IBM Storage Scale on all CNSA worker nodes.
- The script does **not delete or recreate volumes** on IBM Storage Scale; it only updates Kubernetes PV metadata.
- Existing workloads must be restarted to pick up new PV mount paths after migration.
