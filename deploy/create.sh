kubectl apply -f deploy/csi-attacher-rbac.yaml
kubectl apply -f deploy/csi-nodeplugin-rbac.yaml
kubectl apply -f deploy/csi-provisioner-rbac.yaml

kubectl create configmap cmap-config --from-file=cmap-devices.json=deploy/cmap-devices.json --from-file=cmap-settings.json=deploy/cmap-settings.json

kubectl apply -f deploy/csi-plugin-attacher.yaml
kubectl apply -f deploy/csi-plugin-provisioner.yaml
kubectl apply -f deploy/csi-plugin.yaml
