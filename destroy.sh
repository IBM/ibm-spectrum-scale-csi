#
kubectl delete -f examples//pvc.yaml
kubectl delete -f examples/storageclass.yaml

kubectl delete -f deploy/kubernetes/csi-plugin.yaml
kubectl delete -f deploy/kubernetes/csi-plugin-attacher.yaml
kubectl delete -f deploy/kubernetes/csi-plugin-provisioner.yaml
#
#
kubectl delete -f examples/secret.yaml
kubectl delete -f deploy/kubernetes/csi-attacher-rbac.yaml
kubectl delete -f deploy/kubernetes/csi-nodeplugin-rbac.yaml
kubectl delete -f deploy/kubernetes/csi-provisioner-rbac.yaml

kubectl delete configmap cmap-config
