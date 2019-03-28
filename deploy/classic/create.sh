kubectl apply -f deploy/common/csi-attacher-rbac.yaml
kubectl apply -f deploy/common/csi-nodeplugin-rbac.yaml
kubectl apply -f deploy/common/csi-provisioner-rbac.yaml

kubectl create configmap cmap-config --from-file=cmap-devices.json=deploy/classic/cmap-devices.json --from-file=cmap-settings.json=deploy/classic/cmap-settings.json

kubectl apply -f deploy/common/csi-plugin-attacher.yaml
kubectl apply -f deploy/common/csi-plugin-provisioner.yaml
kubectl apply -f deploy/common/csi-plugin.yaml
