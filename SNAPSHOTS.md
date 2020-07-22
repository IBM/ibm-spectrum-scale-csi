# IBM Spectrum Scale CSI driver volume snapshots
Min Scale version required: 5.0.5.1

## Installing the external snapshotter
Note: Kubernetes distributions should provide the external snapshotter by default. OpenShift 4.4+ has the snapshot controller installed by default and below steps are not needed. Perform below two steps for Kubernetes cluster.

### Install external snapshotter CRDs

These are snapshotter beta CRDs. Do this once per cluster

   ```
   kubectl apply -f https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/v2.1.1/config/crd/snapshot.storage.k8s.io_volumesnapshotclasses.yaml

   kubectl apply -f https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/v2.1.1/config/crd/snapshot.storage.k8s.io_volumesnapshotcontents.yaml

   kubectl apply -f https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/v2.1.1/config/crd/snapshot.storage.k8s.io_volumesnapshots.yaml
   ```

### Install snapshot controller

   ```
   kubectl apply -f https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/v2.1.1/deploy/kubernetes/snapshot-controller/rbac-snapshot-controller.yaml

   kubectl apply -f https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/v2.1.1/deploy/kubernetes/snapshot-controller/setup-snapshot-controller.yaml
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
   apiVersion: snapshot.storage.k8s.io/v1beta1
   kind: VolumeSnapshotClass
   metadata:
     name: snapclass1
   driver: spectrumscale.csi.ibm.com
   deletionPolicy: Delete
   ```

### Create VolumeSnapshot
Specify the source volume to be used for creating snapshot here. Source PVC should be in the same namespace in which the snapshot is being created.

   ```
   apiVersion: snapshot.storage.k8s.io/v1beta1
   kind: VolumeSnapshot
   metadata:
     name: snap1
   spec:
     volumeSnapshotClassName: snapclass1
     source:
       persistentVolumeClaimName: pvcfset1

   ```

### Verify that snapshot is created
Snapshot should be in "readytouse" state and a corresponding fileset snapshot should be seen on Spectrum Scale

   ```
   # kubectl get volumesnapshot
   NAME    READYTOUSE   SOURCEPVC   SOURCESNAPSHOTCONTENT   RESTORESIZE   SNAPSHOTCLASS    SNAPSHOTCONTENT                                    CREATIONTIME   AGE
snap1   true         pvcfset1                            208Ki             snapclass1       snapcontent-2b478910-28d1-4c29-8e12-556149095094   2d23h          2d23h

   # mmlssnapshot fs1 -j pvc-d60f90f2-53ed-4f0e-b7be-4587fbcd0234
   Snapshots in file system fs1:
   Directory                SnapId    Status  Created                   Fileset
   snapshot-2b478910-28d1-4c29-8e12-556149095094 14        Valid   Fri Mar 27 05:35:35 2020  pvc-d60f90f2-53ed-4f0e-b7be-4587fbcd0234

   ```

### Create Volume from a source Snapshot
Source snapshot should be in the same namespace as the volume being created. Volume capacity should be less than or equal to the source snapshot's restore size.

   ```
   apiVersion: v1
   kind: PersistentVolumeClaim
   metadata:
      name: pvcfrmsnap1
   spec:
      accessModes:
      - ReadWriteMany
      resources:
         requests:
            storage: 1Gi
   storageClassName: scfilesetinode
   dataSource:
      name: snap1
      kind: VolumeSnapshot
      apiGroup: snapshot.storage.k8s.io

    ```

Resultant PVC should contain data from snap1.