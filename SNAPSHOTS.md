# IBM Spectrum Scale CSI driver volume snapshots
Min Scale version required: 5.1.0.1

## Installing the external snapshotter
Note: Kubernetes distributions should provide the external snapshotter by default. OpenShift 4.4+ has the snapshot controller installed by default and below steps are not needed. Perform below two steps for Kubernetes cluster.

### Install external snapshotter CRDs

These are snapshotter beta CRDs. Do this once per cluster

   ```
   kubectl apply -f https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/v4.0.0/client/config/crd/snapshot.storage.k8s.io_volumesnapshotclasses.yaml

   kubectl apply -f https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/v4.0.0/client/config/crd/snapshot.storage.k8s.io_volumesnapshotcontents.yaml

   kubectl apply -f https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/v4.0.0/client/config/crd/snapshot.storage.k8s.io_volumesnapshots.yaml
   ```

### Install snapshot controller

   ```
   kubectl apply -f https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/v4.0.0/deploy/kubernetes/snapshot-controller/rbac-snapshot-controller.yaml

   curl -O https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/v4.0.0/deploy/kubernetes/snapshot-controller/setup-snapshot-controller.yaml
   ```

Edit setup-snapshot-controller.yaml to ensure the image being used is us.gcr.io/k8s-artifacts-prod/sig-storage/snapshot-controller:v4.0.0. Then apply the manifest.

   ```
   kubectl apply -f setup-snapshot-controller.yaml
   ```

Do this once per cluster

## Install IBM Spectrum Scale CSI operator and driver

Install the operator and driver images with snapshot features. Following resources should be running-

   ```
   # kubectl -n ibm-spectrum-scale-csi-driver get all
   NAME                                                   READY   STATUS    RESTARTS   AGE
   pod/ibm-spectrum-scale-csi-5zvgj                       2/2     Running   0          2m58s
   pod/ibm-spectrum-scale-csi-65nz9                       2/2     Running   0          3m14s
   pod/ibm-spectrum-scale-csi-7jczg                       2/2     Running   0          3m18s
   pod/ibm-spectrum-scale-csi-attacher-0                  1/1     Running   4          10d
   pod/ibm-spectrum-scale-csi-operator-84dbd6f8f7-87hqj   2/2     Running   0          5m43s
   pod/ibm-spectrum-scale-csi-provisioner-0               1/1     Running   4          10d
   pod/ibm-spectrum-scale-csi-snapshotter-0               1/1     Running   4          10d
   pod/ibm-spectrum-scale-csi-wgtbn                       2/2     Running   0          3m38s

   NAME                                              TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE
   service/ibm-spectrum-scale-csi-operator-metrics   ClusterIP   10.98.116.142   <none>        8383/TCP,8686/TCP   14d

   NAME                                    DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR   AGE
   daemonset.apps/ibm-spectrum-scale-csi   4         4         4       4            4           <none>          10d

   NAME                                              READY   UP-TO-DATE   AVAILABLE   AGE
   deployment.apps/ibm-spectrum-scale-csi-operator   1/1     1            1           14d

   NAME                                                         DESIRED   CURRENT   READY   AGE
   replicaset.apps/ibm-spectrum-scale-csi-operator-84dbd6f8f7   1         1         1       5m43s

   NAME                                                  READY   AGE
   statefulset.apps/ibm-spectrum-scale-csi-attacher      1/1     10d
   statefulset.apps/ibm-spectrum-scale-csi-provisioner   1/1     10d
   statefulset.apps/ibm-spectrum-scale-csi-snapshotter   1/1     10d

   ```

## Using the volume snapshot feature

### Create a VolumeSnapshotClass
This is like a StorageClass that defines driver specific attributes for the snapshot to be created

   ```
   apiVersion: snapshot.storage.k8s.io/v1
   kind: VolumeSnapshotClass
   metadata:
     name: ibm-spectrum-scale-snapshot-class
   driver: spectrumscale.csi.ibm.com
   deletionPolicy: Delete
   ```

### Create VolumeSnapshot
Specify the source volume to be used for creating snapshot here. Source PVC should be in the same namespace in which the snapshot is being created. Snapshots can be created only from independent fileset based PVCs.

   ```
   apiVersion: snapshot.storage.k8s.io/v1
   kind: VolumeSnapshot
   metadata:
     name: ibm-spectrum-scale-snapshot
     namespace: default
   spec:
     volumeSnapshotClassName: ibm-spectrum-scale-snapshot-class
     source:
       persistentVolumeClaimName: ibm-spectrum-scale-pvc
   ```

### Verify that snapshot is created
Snapshot should be in "readytouse" state and a corresponding fileset snapshot should be seen on Spectrum Scale

   ```
   # kubectl get volumesnapshot
   NAME    READYTOUSE   SOURCEPVC   SOURCESNAPSHOTCONTENT   RESTORESIZE   SNAPSHOTCLASS    SNAPSHOTCONTENT                                    CREATIONTIME   AGE
ibm-spectrum-scale-snapshot   true         ibm-spectrum-scale-pvc                            1Gi             ibm-spectrum-scale-snapshot-class       snapcontent-2b478910-28d1-4c29-8e12-556149095094   2d23h          2d23h

   # mmlssnapshot fs1 -j pvc-d60f90f2-53ed-4f0e-b7be-4587fbcd0234
   Snapshots in file system fs1:
   Directory                SnapId    Status  Created                   Fileset
   snapshot-2b478910-28d1-4c29-8e12-556149095094 14        Valid   Fri Mar 27 05:35:35 2020  pvc-d60f90f2-53ed-4f0e-b7be-4587fbcd0234

   ```

Note: Volume size of the source PVC is used as the restore size of snapshot. Any volume created from this snapshot must be of the same or larger capacity.

### Create Volume from a source Snapshot
Source snapshot should be in the same namespace as the volume being created. Volume capacity should be greater than or equal to the source snapshot's restore size. Resultant PVC should contain data from ibm-spectrum-scale-snapshot.

   ```
   apiVersion: v1
   kind: PersistentVolumeClaim
   metadata:
      name: ibm-spectrum-scale-pvc-from-snap
      namespace: default
   spec:
      accessModes:
      - ReadWriteMany
      resources:
         requests:
            storage: 1Gi
      storageClassName: ibm-spectrum-scale-storageclass
      dataSource:
         name: ibm-spectrum-scale-snapshot
         kind: VolumeSnapshot
         apiGroup: snapshot.storage.k8s.io
    
   ```

When creating a volume from a volume snapshot, data from source snapshot is copied to the newly created volume. This copy operation uses multiple Spectrum Scale nodes using mmapplypolicy. By default nodes on which the source filesystem is mounted are used. Users can define their own nodeclass to control the nodes where this copy operation should run based on their current workloads. Nodeclass must be defined on Spectrum Scale (ref. mmcrnodeclass). It can then be specified in the storageclass being used for creating PVC using parameter-

   ```
   nodeClass: <nodeclass_name>
   ```

### Static VolumeSnapshot
You can expose a pre-existing Spectrum Scale fileset snapshot in Kubernetes by manually creating a VolumeSnapshotContent as below-

   ```
   apiVersion: snapshot.storage.k8s.io/v1
   kind: VolumeSnapshotContent
   metadata:
      name: ibm-spectrum-scale-snapshot-content
   spec:
      deletionPolicy: Delete
      driver: spectrumscale.csi.ibm.com
      source:
         snapshotHandle: 18133600329030594550;0A1501E9:5F02F150;pvc-79e82f45-1b25-4d83-8ed0-c0d370cad5bd;mysnap
      volumeSnapshotRef:
         name: ibm-spectrum-scale-snapshot
         namespace: default
   ```

Here snapshotHandle is of the format "clusterID;filesystem_UUID;filesetname;snapshotname;relative_path" where "relative_path" is optional.
volumeSnapshotRef is the pointer to the VolumeSnapshot object this content should bind to.

Once the VolumeSnapshotContent is created, create the VolumeSnapshot pointing to the same VolumeSnapshotContent

   ```
   apiVersion: snapshot.storage.k8s.io/v1
   kind: VolumeSnapshot
   metadata:
      name: ibm-spectrum-scale-snapshot
      namespace: default
   spec:
      source:
         volumeSnapshotContentName: ibm-spectrum-scale-snapshot-content
   ```

The VolumeSnapshot listing shows as below-

   ```
   # kubectl get volumesnapshot
   NAME          READYTOUSE   SOURCEPVC   SOURCESNAPSHOTCONTENT   RESTORESIZE   SNAPSHOTCLASS   SNAPSHOTCONTENT                                    CREATIONTIME   AGE
   ibm-spectrum-scale-snapshot    true                     ibm-spectrum-scale-snapshot-content           0                             ibm-spectrum-scale-snapshot-content                                      73m            72m
   ```

Note that here SOURCEPVC is empty and SOURCESNAPSHOTCONTENT points to the VolumeSnapshotContent. This snapshot can be normally used as a source while creating a PVC.

Convenience script tools/generate_volsnapcontent_yaml.sh can be used to generate the volumeSnapshotContent yaml file.

   ```
   Usage: ./generate_volsnapcontent_yaml.sh
                -f|--filesystem <Name of Snapshot's Source Filesystem>
                -F|--fileset <Name of Snapshot's Source Fileset>
                -s|--snapshot <Name of the Snapshot>
                [-p|--path <Relative path within the snapshot>]
                [-c|--snapshotcontentname <name for VolumeSnapshotContent>]
                [-v|--snapshotname <name for VolumeSnapshot>]
                [-n|--namespace <namespace for VolumeSnapshot>]
                [-h|--help] 
   ```
