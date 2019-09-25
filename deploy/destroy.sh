set -e

kubectl delete -f deploy/csi-attacher-rbac.yaml
kubectl delete -f deploy/csi-nodeplugin-rbac.yaml
kubectl delete -f deploy/csi-provisioner-rbac.yaml

kubectl delete -f deploy/spectrum-scale-secret.json
kubectl delete configmap spectrum-scale-config 

kubectl delete -f deploy/csi-plugin-attacher.yaml
kubectl delete -f deploy/csi-plugin-provisioner.yaml
kubectl delete -f deploy/csi-plugin.yaml
