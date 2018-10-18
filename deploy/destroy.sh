kubectl delete -f deploy/csi-plugin.yaml
kubectl delete -f deploy/csi-plugin-attacher.yaml
kubectl delete -f deploy/csi-plugin-provisioner.yaml

kubectl delete -f deploy/csi-attacher-rbac.yaml
kubectl delete -f deploy/csi-nodeplugin-rbac.yaml
kubectl delete -f deploy/csi-provisioner-rbac.yaml

kubectl delete configmap cmap-config

####
#kubectl delete -f examples//pvc.yaml
#kubectl delete -f examples/storageclass.yaml
####
