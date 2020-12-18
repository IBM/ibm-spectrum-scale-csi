# Automation for Operator Lifecycle Manager, Operator installation and upgrade

## Following Steps are for openshift
### Pre-requisite: (openshift)

1. ansible
```
python3 -m pip install ansible
```
2. podman
```
yum install podman
```
3. Get ibm-spectrum-scale-csi repo
```
cd /root
git clone https://github.com/IBM/ibm-spectrum-scale-csi.git

```
4. Create public Container Image Repository on quay
```
i.   Go to quay.io and login.
ii.  Click the + icon in the top right of the header on any quay.io page and choose 'New Repository'
iii. Select 'Container Image Repository' on the next page
iv.  Enter repository name , click on public and and then click the 'Create Public Repository' button.
```

### Steps to Follow (openshift)

1. Disable default operator sources.
```
oc patch OperatorHub cluster --type json -p '[{"op": "add", "path": "/spec/disableAllDefaultSources", "value": true}]'
```
2. Edit required values in oc-olm-test-playbook.yaml
```
QUAY_NAMESPACE: "QUAY_NAMESPACE"         # Quay username
PACKAGE_NAME: "PACKAGE_NAME"             # Quay Container image repository name

# Versions you want to test. Playbook will upload  in order and run  tests.
OPERATOR_VERSIONS:
  - 1.0.0
  - 1.1.0
  - 2.0.0
  - 2.1.0

# Quay username with write access to the application and Quay Password
QUAY_USERNAME: "QUAY_USERNAME"
QUAY_PASSWORD: "QUAY_PASSWORD"

# Check OPERATOR_DIR location is correct
OPERATOR_DIR:  /root/ibm-spectrum-scale-csi/operator/deploy/olm-catalog/ibm-spectrum-scale-csi-operator
```
3. Run following command
```
ansible-playbook oc-olm-test-playbook.yaml
```
4. Go to operatorhub listing of your Openshift cluster and install operator.
   verify operator installtion using  
```
oc get pod -n ibm-spectrum-scale-csi-driver
```
5. Enable default operator sources
```
oc patch OperatorHub cluster --type json -p '[{"op": "add", "path": "/spec/disableAllDefaultSources", "value": false}]'
```
6. Delete Repository from quay 
```
Repositories -> repository_name -> setting -> delete repository
```

## Following Steps are for kubernetes
### Pre-requisite: (kubernetes)

1. ansible
```
python3 -m pip install ansible
```
2. docker
```
yum install docker
```
3. Get ibm-spectrum-scale-csi repo
```
cd /root
git clone https://github.com/IBM/ibm-spectrum-scale-csi.git

```
4. Create public Container Image Repository on quay
```
i.   Go to quay.io and login.
ii.  Click the + icon in the top right of the header on any quay.io page and choose 'New Repository'
iii. Select 'Container Image Repository' on the next page
iv.  Enter repository name , click on public and and then click the 'Create Public Repository' button.
```

### Steps to Follow (kubernetes)

1. Login to quay.io
```
docker login quay.io
```
2. Run following commands
```
kubectl apply -f https://github.com/operator-framework/operator-lifecycle-manager/releases/download/0.13.0/crds.yaml
kubectl apply -f https://github.com/operator-framework/operator-lifecycle-manager/releases/download/0.13.0/olm.yaml
```
3. Check for OLM pods 
```
kubectl get pods -n olm
```
4. Delete catalogsource of all operators
```
kubectl delete catalogsource operatorhubio-catalog -n olm
```
5. Edit required values in k8s-olm-test-playbook.yaml
```
QUAY_NAMESPACE: "QUAY_NAMESPACE"         # Quay username
PACKAGE_NAME: "PACKAGE_NAME"             # Quay Container image repository name


# Versions you want to test. Playbook will upload  in order and run  tests.
OPERATOR_VERSIONS:
  - 1.0.0
  - 1.1.0
  - 2.0.0
  - 2.1.0

# Check OPERATOR_DIR location is correct
OPERATOR_DIR:  /root/ibm-spectrum-scale-csi/operator/deploy/olm-catalog/ibm-spectrum-scale-csi-operator
```
6. Run following command
```
ansible-playbook k8s-olm-test-playbook.yaml
```
7. Verify operator installtion using  
```
kubectl get pods -n ibm-spectrum-scale-csi-driver
kubectl get csv -n ibm-spectrum-scale-csi-driver
kubectl get sub -n ibm-spectrum-scale-csi-driver
kubectl get ip -n ibm-spectrum-scale-csi-driver
```
8. Cleanup
```
kubectl delete sub ibm-spectrum-scale-csi-sub -n ibm-spectrum-scale-csi-driver
kubectl delete operatorgroup operatorgroup -n ibm-spectrum-scale-csi-driver
kubectl delete namespace ibm-spectrum-scale-csi-driver #( before this delete the operator and driver)
kubectl delete catalogsource ibm-spectrum-scale-csi -n olm

kubectl delete -f https://github.com/operator-framework/operator-lifecycle-manager/releases/download/0.13.0/crds.yaml
kubectl delete -f https://github.com/operator-framework/operator-lifecycle-manager/releases/download/0.13.0/olm.yaml

```
9. Delete Repository from quay 
```
Repositories -> repository_name -> setting -> delete repository
```
