# Operator Lifecycle Manager (OLM) Setup

## Sample Setup

From Operator Directory:

``` bash 
curl -sL https://github.com/operator-framework/operator-lifecycle-manager/releases/download/0.11.0/install.sh | bash -s 0.11.0
kubectl apply -f deploy/olm-test/operator-source.yaml
```
