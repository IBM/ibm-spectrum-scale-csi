
# PV Migration Script – CSI (Primary Filesystem → Actual Fileset Mount)

This script helps migrate existing Kubernetes **PersistentVolumes (PVs)** that were originally created when the **primary filesystem and fileset** was enabled, to a format that uses the **actual fileset mount path** after the **primary filesystem has been removed**.
It ensures that workloads continue to access their data seamlessly after migration.

## Key Features

- Preserves all **original PV properties**, including:
  - Reclaim policy (e.g., `Retain`, `Delete`)
  - Access modes (`ReadWriteOnce`, `ReadWriteMany`, etc.)
  - Storage capacity
  - Filesystem type (`fsType`)
  - Labels, annotations, and other PV metadata

- Only the **volumeHandle path segment** is updated to reflect the **actual fileset mount path**.
- Supports migration of PVs across **different fileset types** (independent, dependent, static, cache, CG, etc.).
- Generates **backup YAML files** before applying changes.
- Can be safely re-run if required (**idempotent migration**).

## Why Migration is Required

When the **primary filesystem and fileset** was enabled, PVs were created under paths tied to the **primary filesystem and fileset hierarchy**.
Now that **primary filesystem has been removed**, these paths are invalid:

- PVs must be mounted at the **actual fileset mount path** in the Storage Scale filesystem.
- Without migration, existing workloads would **not be able to mount or access their data**.

This script updates PV definitions to point to the correct fileset paths while **preserving all other PV properties**.

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

### Before (PV created with **primary filesystem** enabled):
```text
volumeHandle: 0;2;13009550825755318848;A3D56F10:9BC12E30;;pvc-3b1a-49d3-89e1-51f607b91234;/ibm/remotefs1/primary-remotefs1-123456789/.volumes/pvc-3b1a-49d3-89e1-51f607b91234
```

### After (Migrated to **actual fileset mount path**):
```text
volumeHandle: 0;2;13009550825755318848;A3D56F10:9BC12E30;;pvc-3b1a-49d3-89e1-51f607b91234;/var/mnt/remotefs1/fs1-pvc-3b1a-49d3-89e1-51f607b91234/pvc-3b1a-49d3-89e1-51f607b91234-data
```

### Key Points

**Unchanged:**
- The identity portion of the handle (everything up to the last `;`).

**Rewritten:**
- Only the **path segment after the last `;`**, now pointing to the actual fileset mount path.

## Volume Type Variations

The exact path suffix (e.g., `pvc-uuid-data`) may vary depending on how the volume was originally created.
The script automatically detects and applies the correct mapping without user intervention.


## Migration Script Usage

A helper script is provided to automate PV migration:

```bash
# Usage
./migration_csi_primary_removal.bash
```

- No arguments are required.
- The script automatically detects the **actual fileset mount path** for each PV and updates accordingly.

## Features of the Migration Script

- ✅ Filters only PVs created with the **spectrumscale.csi.ibm.com** driver.
- ✅ Skips PVs already migrated (with an actual fileset mount path in `volumeHandle`).
- ✅ **Backs up PVs and PVCs** before modification into a structured directory:

```
csi_migration_data/
└── migration-<timestamp>/
    ├── migration.log
    ├── <namespace>/
    │   └── <pvc-name>/
    │       ├── pvc.yaml
    │       └── pv.yaml
    └── ...
```

- ✅ Logs all actions, successes, skips, and failures into:

```
csi_migration_data/migration-<timestamp>/migration.log
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

- Only the **`volumeHandle` path** is modified to reflect the actual fileset mount.

## Notes and Limitations

- The **filesystem names** (e.g., `remotefs1`) must remain identical between pre-primary-filesystem and post-primary-filesystem removal deployments.
- The script does **not delete or recreate volumes** on IBM Storage Scale; it only updates Kubernetes PV metadata.
- Existing workloads must be restarted to pick up new PV mount paths after migration.
