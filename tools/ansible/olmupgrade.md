# Automation for Operator Lifecycle Manager, Operator installation and upgrade

## Following Steps are for openshift
### Pre-requisite: (openshift)

1. Install ansible package 
```
python3 -m pip install ansible
```
2. Install podman
```
yum install podman
```
3. Clone ibm-spectrum-scale-csi repository

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

1. Disable default operator sources on OCP platform
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
  - 2.2.0
  - 2.3.0
  - 2.3.1

# Quay username with write access to the application and Quay Password
QUAY_USERNAME: "QUAY_USERNAME"
QUAY_PASSWORD: "QUAY_PASSWORD"

# Check OPERATOR_DIR location is correct
OPERATOR_DIR:  /root/ibm-spectrum-scale-csi/operator/config/olm-catalog/ibm-spectrum-scale-csi-operator
```
3. Run OLM upgrade playbook using following command
```
ansible-playbook oc-olm-test-playbook.yaml
```
4. Go to operatorhub listing of your OCP cluster and install operator in ibm-spectrum-scale-csi-driver namespace and verify operator installation using 
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

1. Install ansible package 
```
python3 -m pip install ansible
```
2. Install docker
```
yum install docker
```
3. Clone ibm-spectrum-scale-csi repo
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

1. Login to quay.io using quay username which will be used for pushing the image in OLM upgrade run
```
docker login quay.io -u  <username>
```
2. Install OLM using following commands
```
kubectl apply -f https://raw.githubusercontent.com/operator-framework/operator-lifecycle-manager/v0.19.1/deploy/upstream/quickstart/crds.yaml
kubectl apply -f https://raw.githubusercontent.com/operator-framework/operator-lifecycle-manager/v0.19.1/deploy/upstream/quickstart/olm.yaml
```
3. Check OLM pods status 
```
kubectl get pods -n olm
```
4. Delete catalogsource of default  operators
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
  - 2.2.0
  - 2.3.0
  - 2.3.1

 QUAY_USERNAME: "QUAY_USERNAME"             # Quay username used for login to quay.io and have admin access to Quay Container image repository name
 QUAY_PASSWORD: "QUAY_PASSWORD"             # Quay username's token for login to quay.io to push the image to Quay Container image repository

# Check OPERATOR_DIR location is correct
OPERATOR_DIR:  /root/ibm-spectrum-scale-csi/operator/config/olm-catalog/ibm-spectrum-scale-csi-operator
```
6. Run following command
```
ansible-playbook k8s-olm-test-playbook.yaml
```
7. Verify operator installtion using  
```
kubectl get pods -n ibm-spectrum-scale-csi-driver
```
8. Cleanup
```
1. Delete CSI driver and operator

2. Run below commands for OLM cleanup 

kubectl delete sub ibm-spectrum-scale-csi-sub -n ibm-spectrum-scale-csi-driver
kubectl delete operatorgroup operatorgroup -n ibm-spectrum-scale-csi-driver
kubectl delete namespace ibm-spectrum-scale-csi-driver
kubectl delete catalogsource ibm-spectrum-scale-csi -n olm

kubectl delete -f https://raw.githubusercontent.com/operator-framework/operator-lifecycle-manager/v0.19.1/deploy/upstream/quickstart/crds.yaml
kubectl delete -f https://raw.githubusercontent.com/operator-framework/operator-lifecycle-manager/v0.19.1/deploy/upstream/quickstart/olm.yaml

```
9. Delete Repository from quay 
```
Repositories -> repository_name -> setting -> delete repository
```
