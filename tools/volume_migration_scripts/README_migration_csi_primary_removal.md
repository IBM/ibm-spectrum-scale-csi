
# Migration Script – IBM Storage Scale CSI (Primary filesystem mount path to Actual fileset mount path)

This script helps to migrate existing Kubernetes **PersistentVolumes (PVs)** that were originally created when the **primary filesystem and fileset** was enabled, to a format that uses the **actual fileset mount path** after the **primary filesystem/fileset has been removed**.

It ensures that workloads continue to access their data seamlessly after migration.

## 1. Key Features

- Preserves all **original PV properties**, including:
    - Reclaim policy (e.g., `Retain`, `Delete`)
    - Access modes (`ReadWriteOnce`, `ReadWriteMany`, etc.)
    - Storage capacity
    - Filesystem type (`fsType`)
    - Labels, annotations, and other PV metadata

- Only the **volumeHandle path segment** is updated to a new IBM Storage Scale CSI required format.
- Supports migration of PVs across **different fileset types** (independent, dependent, static, cache, CG, etc.).
- Generates **backup YAML files** before applying changes.
- Can be safely re-run if required (**idempotent migration**).

## 2. Why Migration is Required

Till IBM Storage Scale CSI 2.14.x , the **primary filesystem and fileset** hierachy were being used, and the PVs were created under paths tied to the **primary filesystem and fileset hierarchy**.

Going forward IBM Storage Scale CSI 3.0.0, the **primary filesystem/fileset has been removed**, these paths are invalid:

- PVs definitions needs to be updated according to the **actual fileset mount path** in the IBM Storage Scale filesystem.
- Without migration, existing workloads would **not be able to mount or access their data**  after switching to IBM Storage Scale CSI 6.0.0

This script updates PV definitions to use the new path structure while **preserving all other PV properties**.

## 3. Prerequisites

- **Delete any existing workloads or application pods attached to the PVCs that will be migrated** from the existing IBM Storage Scale CSI cluster.

Before running the migration script, ensure the following tools are installed and available in your `$PATH`:

- The script should be run in the context of the cluster where IBM Storage Scale CSI is deployed
- **kubectl** – to interact with the Kubernetes cluster and fetch/update PV/PVC objects
- **jq** – for JSON parsing and manipulation of Kubernetes API responses

You can verify installation with:

```bash
kubectl version --client
jq --version
```


## 4. Example Transformation

### Before (PV created with **primary filesystem/fileset** enabled):
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

## 5. Volume Type Variations

The exact path suffix (e.g., `pvc-uuid-data`) may vary based on how the volume was originally created.
The script automatically detects and applies the correct mapping without user intervention.


## 6. Migration Script Usage

A helper script is provided to automate PV migration:

```bash
# Usage
./migration_csi_primary_removal.bash
```

- No arguments are required.
- The script automatically detects the **actual fileset mount path** for each PV and updates accordingly.

## 7. Validate the migration.

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

## 8. Features of the Migration Script

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
csi_migration_data/<timestamp>/migration.log
```

- ✅ Summarizes **success, skipped, and failed** migrations at the end.
- ✅ Idempotent – safe to re-run if needed.

## 9. Preserved PV Properties

The script ensures that **all original PV configurations** are retained after migration.
The following fields are preserved:

- **Capacity** (`spec.capacity.storage`)
- **AccessModes** (`spec.accessModes`)
- **PersistentVolumeReclaimPolicy** (`spec.persistentVolumeReclaimPolicy`)
- **StorageClassName**
- **CSI driver details** (fsType, volumeAttributes, nodeStageSecrets, etc.)
- **PVC binding information** (safely re-created to preserve claim references)

- Only the **`volumeHandle` path** is modified to reflect the actual fileset mount.

## 10. Notes and Limitations

- The **filesystem names** (e.g., `remotefs1`) must remain identical between pre-primary-filesystem and post-primary-filesystem removal deployments.
- The script does **not delete or recreate volumes** on IBM Storage Scale; it only updates Kubernetes PV metadata.
- Existing workloads must be restarted to pick up new PV mount paths after migration.
