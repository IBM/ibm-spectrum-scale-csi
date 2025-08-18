# PV Migration Script – CSI → CNSA

This script helps migrate existing Kubernetes **PersistentVolumes (PVs)** created under the **standalone CSI driver** to the **CNSA-compatible CSI driver** format.
It ensures that workloads continue to access their data seamlessly after migration.

## Key Features

- Preserves all **original PV properties**, including:
  - Reclaim policy (e.g., `Retain`, `Delete`)
  - Access modes (`ReadWriteOnce`, `ReadWriteMany`, etc.)
  - Storage capacity
  - Filesystem type (`fsType`)
  - Labels, annotations, and other PV metadata

- Only the **volumeHandle path segment** is updated to CNSA’s required format.
- Supports migration of PVs created from **different volume types** (fileset-based, CG, cache, static, dependent, independent).
- Generates **backup YAML files** before applying changes.
- Can be safely re-run if required (**idempotent migration**).

## Why Migration is Required

In the standalone **CSI driver** setup, PVs are created with a `volumeHandle` format tied to paths under **primary filesets**, different **volumeMounts**, or **consistency groups**.

However, in a **CNSA setup**, this format becomes incompatible because:

- The **primary fileset path** is no longer available.
- CNSA expects a **different mount path hierarchy** for PV data.
- PVs must reference a **common base mount point** where all IBM Storage Scale remote filesystems are mounted across CNSA Kubernetes worker nodes.

- Without migration, existing workloads would **not be able to mount or access their data** after switching from standalone CSI to CNSA.

This script updates PV definitions to use the new path structure while **preserving all other PV properties**.


## What is `--new_path_prefix`?

The `--new_path_prefix` is the **base filesystem mount point** of the remotely mounted filesystems on the **local IBM Storage Scale instance running on your CNSA Kubernetes worker nodes**.

It defines the **root filesystem path** under which all migrated PVs will be remapped.

### Examples

- `/var/mnt` – often used in CNSA-based installations
- `/ibm` – default mount point (commonly used in classic setups)
- `/mnt` – alternative mount point used in some deployments

**Important:**
The prefix you provide must **exactly match how the filesystem is mounted** on all CNSA worker nodes.
If different nodes have different mount points, migration will **fail**.


## Example Transformation

### Before (CSI standalone `volumeHandle`):
```text
volumeHandle: 0;2;13009550825755318848;9A7B0B0A:68891B40;;pvc-26946b2b-b18a-4c0d-9f77-606a444094c1;/ibm/remotefs1/primary-fileset-remotefs1-475592072879187/.volumes/pvc-26946b2b-b18a-4c0d-9f77-606a444094c1
```

### After (CNSA-compatible `volumeHandle`):
```text
volumeHandle: 0;2;13009550825755318848;9A7B0B0A:68891B40;;pvc-26946b2b-b18a-4c0d-9f77-606a444094c1;/var/mnt/remotefs1/pvc-26946b2b-b18a-4c0d-9f77-606a444094c1/pvc-26946b2b-b18a-4c0d-9f77-606a444094c1-data
```

### Key Points

**Unchanged:**
- The identity portion of the handle (everything up to the last `;`).

**Rewritten:**
- Only the **path segment after the last `;`**.


## Volume Type Variations

The exact path suffix (e.g., `pvc-uuid-data`) may vary based on how the volume was originally created.
The script automatically detects and applies the correct mapping without user intervention.

## Prerequisites

Before running the migration script, ensure the following tools are installed and available in your `$PATH`:

- **kubectl** – to interact with the Kubernetes cluster and fetch/update PV/PVC objects
- **jq** – for JSON parsing and manipulation of Kubernetes API responses

You can verify installation with:

```bash
kubectl version --client
jq --version
```

## Migration Script Usage

A helper script is provided to automate PV migration:

```bash
# Usage
./pv_migration_csi_to_cnsa.bash <migration_name> --new_path_prefix /var/mnt
```

Where:

- `<migration_name>` is a user-defined identifier for this migration run.
- `--new_path_prefix` specifies the **base mount point** of IBM Storage Scale filesystems on CNSA worker nodes.
  Allowed values: `/ibm`, `/mnt`, `/var/mnt`


## Features of the Migration Script

- ✅ Filters only PVs created with the **spectrumscale.csi.ibm.com** driver.
- ✅ Skips PVs already migrated (with a CNSA-compatible path in `volumeHandle`).
- ✅ **Backs up PVs and PVCs** before modification into a structured directory:

```
csi_migration_data/
└── <migration-name>-<timestamp>/
    ├── migration.log
    ├── <namespace>/
    │   └── <pvc-name>/
    │       ├── pvc.yaml
    │       └── pv.yaml
    └── ...
```

- ✅ Logs all actions, successes, skips, and failures into:

```
csi_migration_data/<migration-name>-<timestamp>/migration.log
```

- ✅ Summarizes **success, skipped, and failed** migrations at the end.
- ✅ Idempotent – safe to re-run if needed.


## Preserved PV Properties

The script ensures that **all original PV configurations** are retained after migration.
The following fields are preserved:

- **Capacity** (`spec.capacity.storage`)
- **AccessModes** (`spec.accessModes`)
- **PersistentVolumeReclaimPolicy** (`spec.persistentVolumeReclaimPolicy`)
- **StorageClassName**
- **CSI driver details** (fsType, volumeAttributes, nodeStageSecrets, etc.)
- **PVC binding information** (safely re-created to preserve claim references)

- Only the **`volumeHandle` path** is modified to meet CNSA expectations.


## Notes and Limitations

- The **filesystem names** (e.g., `remotefs1`) must remain identical between standalone CSI and CNSA deployments.
- The provided `--new_path_prefix` must reflect the **actual base mount point** of IBM Storage Scale on all CNSA **worker nodes**.
- The script does **not delete or recreate volumes** on IBM Storage Scale; it only updates Kubernetes PV metadata.
- Existing workloads must be restarted to pick up new PV mount paths after migration.
