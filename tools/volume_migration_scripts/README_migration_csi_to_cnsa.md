# Migration Script - IBM Storage Scale CSI to IBM Storage Scale container native

The script helps to migrate existing Kubernetes **PersistentVolumes (PVs)** created under the **IBM Storage Scale CSI** to the **IBM Storage Scale container native-compatible** format.
It ensures that workloads continue to access their data seamlessly after migration.

## 1. Key Features

- Preserves all **original PV properties**, including:
    - Reclaim policy (e.g., `Retain`, `Delete`)
    - Access modes (`ReadWriteOnce`, `ReadWriteMany`, etc.)
    - Storage capacity
    - Filesystem type (`fsType`)
    - Labels, annotations, and other PV metadata

- Only the **volumeHandle path segment** is updated to IBM Storage Scale container native’s required format.
- Supports migration of PVs created from **different volume types** (fileset-based, CG, cache, static, dependent, independent).
- Generates **backup YAML files** before applying changes.
- Can be safely re-run if required (**idempotent migration**).

## 2. Why Migration is Required

In the standalone **IBM Storage Scale CSI** setup, PVs are created with a `volumeHandle` format tied to paths under **primary filesets**, different **volumeMounts**, or **consistency groups**.

However, in a **IBM Storage Scale container native setup**, this format becomes incompatible because:

- The **primary fileset path** is no longer available for the volumes, after the upgrade to IBM Storage Scale container native 5.2.3.x
- Also, the **primary fileset path** concept has been removed from IBM Storage Scale container native 6.0.0.x
- IBM Storage Scale container native expects a **different mount path hierarchy** for PV data.
- PVs must reference a **common base mount point** where all IBM Storage Scale remote filesystems are mounted across IBM Storage Scale container native Kubernetes worker nodes.

- Without migration, existing workloads would **not be able to mount or access their data** after switching from IBM Storage Scale CSI to IBM Storage Scale container native.

This script updates PV definitions to use the new path structure while **preserving all other PV properties**.


## 3. Prerequisites

- **Delete any existing workloads or application pods attached to the PVCs that will be migrated** from the existing IBM Storage Scale CSI cluster **before removing IBM Storage Scale CSI.**

Before running the migration script, ensure the following tools are installed and available in your `$PATH`:

- The script should be run in the context of the cluster where IBM Storage Scale CSI is deployed
- **kubectl** – to interact with the Kubernetes cluster and fetch/update PV/PVC objects
- **jq** – for JSON parsing and manipulation of Kubernetes API responses

You can verify installation with:

```bash
kubectl version --client
jq --version
```

## 4. What is `--new_path_prefix`?

The `--new_path_prefix` is the **base filesystem mount point** of the remotely mounted filesystems on the **local IBM Storage Scale instance running on your IBM Storage Scale container native Kubernetes worker nodes**.

It defines the **root filesystem path** under which all migrated PVs will be remapped.

### Examples

- `/var/mnt` – often used in IBM Storage Scale container native-based installations
- `/ibm` – default mount point (commonly used in IBM Storage Scale setups)
- `/mnt` – alternative mount point used in some deployments

**Important:**
The prefix you provide must **exactly match how the filesystem is mounted** on all IBM Storage Scale container native worker nodes.
If different nodes have different mount points, migration will **fail**.


## 5. Example Transformation

### Before (IBM Storage Scale CSI `volumeHandle`):
```text
volumeHandle: 0;2;13009550825755318848;9A7B0B0A:68891B40;;pvc-26946b2b-b18a-4c0d-9f77-606a444094c1;/ibm/remotefs1/primary-fileset-remotefs1-475592072879187/.volumes/pvc-26946b2b-b18a-4c0d-9f77-606a444094c1
```

### After (IBM Storage Scale container native-compatible `volumeHandle`):
```text
volumeHandle: 0;2;13009550825755318848;9A7B0B0A:68891B40;;pvc-26946b2b-b18a-4c0d-9f77-606a444094c1;/var/mnt/remotefs1/pvc-26946b2b-b18a-4c0d-9f77-606a444094c1/pvc-26946b2b-b18a-4c0d-9f77-606a444094c1-data
```

### Key Points

**Unchanged:**
- The identity portion of the handle (everything up to the last `;`).

**Rewritten:**
- Only the **path segment after the last `;`**.


## 6. Volume Type Variations

The exact path suffix (e.g., `pvc-uuid-data`) may vary based on how the volume was originally created.
The script automatically detects and applies the correct mapping without user intervention.


## 7. Migration Script Usage

A helper script is provided to automate PV migration:

```bash
# Usage
./migration_csi_to_cnsa.bash --new_path_prefix /var/mnt
```

Where:

- `--new_path_prefix` specifies the **base mount point** of IBM Storage Scale filesystems on IBM Storage Scale container native worker nodes. Allowed values: `/ibm`, `/mnt`, `/var/mnt`

## 8. Validate the migration.

Validate the migration once the migration script finishes. The migration summary should have only successful or skipped PVs. There shouldn't be any Failed PVs in the summary.
```bash
Migration Summary:
------------------------------------------------------------------------------------------------------------------------
Successful PVs: 1
    PVC Name                                                                  | PV Name
    --------------------------------------------------------------------------------------------------------------------
    ibm-spectrum-scale-pvc-clone-from-pvc-advanced                            | pvc-7981775f-08b1-4a53-ae4d-6740e2ec9a89

Skipped PVs (already migrated): 2
    PVC Name                                                                  | PV Name
    --------------------------------------------------------------------------------------------------------------------
    scale-advance-pvc                                                         | pvc-dd8e015c-382c-4487-a215-91dd22a01d45
    ibm-spectrum-scale-pvc-advance-from-snapshot                              | pvc-e920921e-7a25-4017-9c8d-6469750a4772
```

## 9. Features of the Migration Script

- ✅ Filters only PVs created with the **spectrumscale.csi.ibm.com** driver.
- ✅ Skips PVs already migrated (with a IBM Storage Scale container native-compatible path in `volumeHandle`).
- ✅ **Backs up PVs and PVCs** before modification into a structured directory:

```text
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

```text
csi_migration_data/<migration-name>-<timestamp>/migration.log
```

- ✅ Summarizes **success, skipped, and failed** migrations at the end.
- ✅ Idempotent – safe to re-run if needed.


## 10. Preserved PV Properties

The script ensures that **all original PV configurations** are retained after migration.
The following fields are preserved:

- **Capacity** (`spec.capacity.storage`)
- **AccessModes** (`spec.accessModes`)
- **PersistentVolumeReclaimPolicy** (`spec.persistentVolumeReclaimPolicy`)
- **StorageClassName**
- **CSI driver details** (fsType, volumeAttributes, nodeStageSecrets, etc.)
- **PVC binding information** (safely re-created to preserve claim references)

- Only the **`volumeHandle` path** is modified to meet IBM Storage Scale container native expectations.


## 11. Notes and Limitations

- The **filesystem names** (e.g., `remotefs1`) must remain identical between IBM Storage Scale CSI and IBM Storage Scale container native deployments.
- The provided `--new_path_prefix` must reflect the **actual base mount point** of IBM Storage Scale on all IBM Storage Scale container native **worker nodes**.
- The script does **not delete or recreate volumes** on IBM Storage Scale; it only updates Kubernetes PV metadata.
- Existing workloads must be restarted to pick up new PV mount paths after migration.
