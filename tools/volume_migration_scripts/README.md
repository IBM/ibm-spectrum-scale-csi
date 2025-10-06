# IBM Storage Scale CSI → Volume Migration Scripts

This provides scripts to migrate existing IBM Storage Scale CSI PersistentVolumes (PVs) to the new compatible volumeHandles.

## Scripts

1. [`migration_csi_to_cnsa.bash`](README_migration_csi_to_cnsa.md)
    Migrates IBM Storage Scale CSI PersistentVolumes to the IBM Storage Scale container native format by updating the volumeHandle path with the specified prefix.

2. [`migration_csi_primary_removal.bash`](README_migration_csi_primary_removal.md)
    Migrates IBM Storage Scale CSI PersistentVolumes to a new format by updating the volumeHandle path with the specified prefix after primary filesystem removal.

3.  [`migration_cnsa_primary_removal.bash`](README_migration_cnsa_primary_removal.md)
    Migrates IBM Storage Scale CSI PersistentVolumes to IBM Storage Scale container native format by updating the volumeHandle path with the specified prefix after primary filesystem removal.


**Note:**
Each script has its **own README** with usage, inputs, outputs, and validation notes.
