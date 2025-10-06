# IBM Storage Scale CSI/IBM Storage Scale container native → Volume Migration Scripts

This provides scripts to migrate existing PersistentVolumes (PVs) to the new compatible volumeHandles.

## Scripts

1. [`migration_csi_to_cnsa.bash`](README_migration_csi_to_cnsa.md)
    Migrates IBM Storage Scale CSI PersistentVolumes to the IBM Storage Scale container native format by updating the volumeHandle path with the specified filesystem mount path prefix.

2. [`migration_csi_primary_removal.bash`](README_migration_csi_primary_removal.md)
    Migrates IBM Storage Scale CSI PersistentVolumes to a new format by updating the volumeHandle path with the specified filesystem mount path prefix after primary filesystem removal.

3.  [`migration_cnsa_primary_removal.bash`](README_migration_cnsa_primary_removal.md)
    Migrates IBM Storage Scale container native PersistentVolumes details by updating the volumeHandle path with the specified filesystem mount path prefix after primary filesystem/fileset removal.


**Note:**
Each script has its **own README** with usage, inputs, outputs, and validation notes.
