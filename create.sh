kubectl apply -f flex/deploy/gpfs/kubernetes/rbac.yaml
#
kubectl apply -f csi/examples/gpfs/storageclass.yaml
#
kubectl apply -f csi/examples/gpfs/secret.yaml
#
kubectl apply -f csi/deploy/gpfs/kubernetes/csi-attacher-rbac.yaml
kubectl apply -f csi/deploy/gpfs/kubernetes/csi-nodeplugin-rbac.yaml
kubectl apply -f csi/deploy/gpfs/kubernetes/csi-provisioner-rbac.yaml
#
kubectl create configmap cmap-config --from-file=cmap-devices.json=deploy/cmap-devices.json --from-file=cmap-settings.json=deploy/cmap-settings.json
#

#kubectl apply -f deploy/gpfs-ds.yaml
#kubectl apply -f csi/deploy/gpfs/kubernetes/csi-gpfsplugin-attacher.yaml
#kubectl apply -f csi/deploy/gpfs/kubernetes/csi-gpfsplugin-provisioner.yaml
#kubectl apply -f csi/deploy/gpfs/kubernetes/csi-gpfsplugin.yaml

#######
#
#kubectl apply -f csi/examples/gpfs/pvc.yaml
#
# kubectl create configmap cmap-settings --from-file=cmap-settings.json
# kubectl create configmap cmap-config --from-file=cmap-devices.json=k8storage/deploy/cmap-devices.json --from-file=cmap-settings.json=k8storage/deploy/cmap-settings.json
