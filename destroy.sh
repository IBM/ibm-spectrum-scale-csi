#
kubectl delete -f deploy/gpfs-ds.yaml
kubectl delete -f csi/examples/gpfs/pvc.yaml
kubectl delete -f csi/examples/gpfs/storageclass.yaml

kubectl delete -f flex/deploy/gpfs/kubernetes/rbac.yaml

kubectl delete -f csi/deploy/gpfs/kubernetes/csi-gpfsplugin.yaml
kubectl delete -f csi/deploy/gpfs/kubernetes/csi-gpfsplugin-attacher.yaml
kubectl delete -f csi/deploy/gpfs/kubernetes/csi-gpfsplugin-provisioner.yaml
#
#
kubectl delete -f csi/examples/gpfs/secret.yaml
kubectl delete -f csi/deploy/gpfs/kubernetes/csi-attacher-rbac.yaml
kubectl delete -f csi/deploy/gpfs/kubernetes/csi-nodeplugin-rbac.yaml
kubectl delete -f csi/deploy/gpfs/kubernetes/csi-provisioner-rbac.yaml

kubectl delete configmap cmap-config
