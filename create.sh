kubectl apply -f examples/storageclass.yaml
#
kubectl apply -f examples/secret.yaml
#
kubectl apply -f deploy/kubernetes/csi-attacher-rbac.yaml
kubectl apply -f deploy/kubernetes/csi-nodeplugin-rbac.yaml
kubectl apply -f deploy/kubernetes/csi-provisioner-rbac.yaml
#
kubectl create configmap cmap-config --from-file=cmap-devices.json=deploy/cmap-devices.json --from-file=cmap-settings.json=deploy/cmap-settings.json
#

kubectl apply -f deploy/kubernetes/csi-plugin-attacher.yaml
kubectl apply -f deploy/kubernetes/csi-plugin-provisioner.yaml
kubectl apply -f deploy/kubernetes/csi-plugin.yaml

#######
#kubectl apply -f examples/pvc.yaml
