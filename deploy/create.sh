set -e

kubectl apply -f deploy/csi-attacher-rbac.yaml
kubectl apply -f deploy/csi-nodeplugin-rbac.yaml
kubectl apply -f deploy/csi-provisioner-rbac.yaml

kubectl apply -f deploy/spectrum-scale-secret.json
kubectl create configmap spectrum-scale-config --from-file=spectrum-scale-config.json=deploy/spectrum-scale-config.json

kubectl apply -f deploy/csi-plugin-attacher.yaml
kubectl apply -f deploy/csi-plugin-provisioner.yaml
kubectl apply -f deploy/csi-plugin.yaml
