kubectl delete -f deploy/common/csi-plugin.yaml
kubectl delete -f deploy/common/csi-plugin-attacher.yaml
kubectl delete -f deploy/common/csi-plugin-provisioner.yaml

kubectl delete -f deploy/common/csi-attacher-rbac.yaml
kubectl delete -f deploy/common/csi-nodeplugin-rbac.yaml
kubectl delete -f deploy/common/csi-provisioner-rbac.yaml

kubectl delete configmap cmap-config

####
#kubectl delete -f examples//pvc.yaml
#kubectl delete -f examples/storageclass.yaml
####
