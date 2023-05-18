# Functional Test Automation Suite

This Functional Test Automation Suite exercises and tests the IBM Storage Scale CSI functionality.

### Tested Testbed environment

- IBM Storage Scale Cluster - 5.1.0.1+ Version  (**IBM Storage Scale supported kernel version**)
- Kubernetes Cluster Version 1.19 - 1.22
- Openshift Version 4.8.x, 4.9.x
- IBM Storage Scale Cluster CSI Version - 2.4.0+


### How to run IBM Storage Scale CSI test automation using container

Note: Use `oc` command instead of `kubectl` in case of Openshift Container Platform 

#### In case IBM Storage Scale CSI operator and driver is already deployed

- Configuring parameters such as remote/local filesystem name, uid/gid, you must modify relevant fields in [test.config](./config/test.config) file.

- Create config map using [test.config](./config/test.config)

```
kubectl create configmap  test-config  --from-file=test.config=<test.config file path>
```
Example
```
kubectl create configmap  test-config  --from-file=test.config=/root/ibm-spectrum-scale-csi/tests/functional-tests/config/test.config 
```

#### In case IBM Storage Scale CSI operator and driver is not deployed
- Configure parameters in [csiscaleoperators.csi.ibm.com_cr.yaml](../../operator/config/samples/csiscaleoperators.csi.ibm.com_cr.yaml) file.

- Configuring parameters such as secret username/password, cacert path, uid/gid  & remote/local filesystem name, you must modify relevant fields in [test.config](./config/test.config) file.

- Create config map using [test.config](./config/test.config),[csiscaleoperators.csi.ibm.com_cr.yaml](../..//operator/config/samples/csiscaleoperators.csi.ibm.com_cr.yaml)

```
kubectl create configmap  test-config  --from-file=test.config=<test.config file path>  --from-file=csiscaleoperators.csi.ibm.com_cr.yaml=<csiscaleoperators.csi.ibm.com_cr.yaml file path>

```
Example
```
kubectl create configmap  test-config  --from-file=test.config=/root/ibm-spectrum-scale-csi/tests/functional-tests/config/test.config  --from-file=csiscaleoperators.csi.ibm.com_cr.yaml=/root/ibm-spectrum-scale-csi/operator/config/samples/csiscaleoperators.csi.ibm.com_cr.yaml
 
```
- To run the tests in namespace other than ibm-spectrum-scale-csi-test namespace (default) , please use --testnamespace parameter
- To run each testcase in its own namespace, please use --createnamespace parameter
- If operator is not running in namespace ibm-spectrum-scale-csi-driver namespace (default) , please use --operatornamespace with value where operator is already running
- If operator yaml file is not at ../../generated/installer/ibm-spectrum-scale-csi-operator-dev.yaml (default), please use --operatoryaml with value of operator yaml file path

- if you want to use SSL=enable, for cacert configmap use following command and change the path in test.config file as `config/local.crt`
```
kubectl create configmap  test-config  --from-file=test.config=<test.config filepath>  --from-file=csiscaleoperators.csi.ibm.com_cr.yaml=<csiscaleoperators.csi.ibm.com_cr.yaml file path> --from-file=kubeconfig=<kubeconfig file path> --from-file=local.crt=<local.crt file path>
```
Note : for remote crt, pass remote.crt file in the same configmap and user in the test.config file

- Configure sample [csi-test-pod.yaml](./csi-test-pod.yaml) file 

- How to get "APISERVER VALUE" and "TOKEN VALUE"
on Kubernetes or Openshift cluster where kubernetes objects will be created
```
export APISERVER=$(kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}')
echo "APISERVER is " $APISERVER
```
```
kubectl create serviceaccount test-automation -n default
kubectl create clusterrolebinding test-automation-crb --clusterrole=cluster-admin --serviceaccount=default:test-automation

export TOKEN=$(kubectl get secret $(kubectl get secrets -n default | grep test-automation-token | awk 'NR==1{print $1}')  -o jsonpath='{.data.token}' -n default | base64 --decode)
echo "TOKEN is " $TOKEN
```
Optional: ca.crt value
```
export CACRT=$(kubectl get secret $(kubectl get secrets -n default | grep test-automation-token | awk 'NR==1{print $1}')  -o jsonpath='{.data.ca\.crt}' -n default)
echo "CACRT is " $CACRT
```

```
kind: Pod
apiVersion: v1
metadata:
  name: ibm-spectrum-scale-csi-test-pod
spec:
  containers:
  - name: csi-test
    image: quay.io/jainbrt/ibm-spectrum-scale-csi-test:x86 #container iamge for x86
#    image: quay.io/jainbrt/ibm-spectrum-scale-csi-test:ppcle64 #container iamge for ppcle64
#    image: quay.io/jainbrt/ibm-spectrum-scale-csi-test:s390x  #container iamge for s390x
    securityContext:
      privileged: true
    command: [ "/bin/sh", "-c", "--" ]
    args: [ "while true; do sleep 120; done;" ]
    imagePullPolicy: "Always"
    env:
    - name: APISERVER
      value: "<APISERVER VALUE>"
    - name: TOKEN
      value: "<TOKEN VALUE>"
#    - name: CACRT
#      value: "<CACRT VALUE>"
    volumeMounts:
    - mountPath: /data
      name: report
    - mountPath: /ibm-spectrum-scale-csi/tests/functional-tests/config
      name: test-config
  volumes:
  - configMap:
      defaultMode: 420
      name: test-config  #test-config configmap for configuration files
    name: test-config
  - hostPath:
      path: /ibm/gpfs0    #local path for saving the test reports files on node of running pod
    name: report
  restartPolicy: "Never"
  nodeSelector:
     kubernetes.io/hostname: "<Worker Node name>" #node selector for scheduling the test pod
```

- Create test pod 

```
kubectl apply -f <test-pod-yaml-file>

eg. kubectl apply -f csi-test-pod.yaml 
```

- For changing the name of html report, pass the `--html` with remote file name (optional).For example :

### Run driver tests on primary cluster using driver_test.py as shown below -
```

#Run all testcases in driver testsuite using:
kubectl exec -it <csi-test-pod-name> -- pytest  tests/volume_provisioning.py --html=/data/<report-name>.html

#Run any specific testcase in testsuite using:
kubectl exec -it <csi-test-pod-name> -- pytest  tests/volume_provisioning.py::<test_name> --html=/data/<report-name>.html

eg. kubectl exec -it <csi-test-pod-name> -- pytest  tests/volume_provisioning.py::test_driver_dynamic_pass_1 --html=/data/<report-name>.html
```
                
### Run operator tests using operator_test.py as shown below -
```       

#Run all testcases in operator testsuite using:
kubectl exec -it <csi-test-pod-name> -- pytest  tests/operator_test.py --html=/data/<report-name>.html

#Run any specific testcase in testsuite using:
kubectl exec -it <csi-test-pod-name> -- pytest  tests/operator_test.py::<test_name>  --html=/data/<report-name>.html

eg. kubectl exec -it <csi-test-pod-name> -- pytest  tests/operator_test.py::test_operator_deploy --html=/data/<report-name>.html
```

### Run driver tests on remote cluster using remote_test.py as shown below -
```

#Run all testcases in driver testsuite using:
kubectl exec -it <csi-test-pod-name> -- pytest  tests/remotecluster_volume_provisioning.py --html=/data/<report-name>.html

#Run any specific testcase in testsuite using:
kubectl exec -it <csi-test-pod-name> -- pytest  tests/remotecluster_volume_provisioning.py::<test_name> --html=/data/<report-name>.html

eg. kubectl exec -it <csi-test-pod-name> -- pytest  tests/remotecluster_volume_provisioning.py::test_driver_dynamic_pass_1 --html=/data/<report-name>.html
```

### List available Driver & Operator tests 
Available functional tests list for driver & operator can be collected using following command
```
kubectl exec -it <csi-test-pod-name> -- pytest --collect-only 
```
### For running full testsuite with the tests which take long time, use `--runslow` parameter ( these tests are being marked with @pytest.mark.slow).
For example :

```
kubectl exec -it <csi-test-pod-name> -- pytest tests/volume_snapshot.py --runslow --html=/data/<report-name>.html   #This will run all testcases including those marked with slow
kubectl exec -it <csi-test-pod-name> -- pytest tests/volume_snapshot.py::test_snapshot_dynamic_multiple_snapshots_256 --runslow --html=/data/<report-name>.html
```

### Running Testcases using markers
1. Running all testcases except operator testcases
```
kubectl exec -it <csi-test-pod-name> -- pytest -m "volumeprovisioning or volumesnapshot"
```
2. Running only local cluster testcases
```
kubectl exec -it <csi-test-pod-name> -- pytest -m "localcluster"
```
3. Running only remove cluster testcases
```
kubectl exec -it <csi-test-pod-name> -- pytest -m "remotecluster"
```
4. Running only operator testcases
```
kubectl exec -it <csi-test-pod-name> -- pytest -m "csioperator"
```
5. Running only localcluster volumeprovisioning testcases
```
kubectl exec -it <csi-test-pod-name> -- pytest -m "volumeprovisioning and localcluster"
```
like above format you can combine volumeprovisioning, volumesnapshot, localcluster and remotecluster markers.
Few examples,
```
eg. kubectl exec -it <csi-test-pod-name> -- pytest -m "volumeprovisioning and remotecluster"
eg. kubectl exec -it <csi-test-pod-name> -- pytest -m "volumesnapshot and localcluster"
eg. kubectl exec -it <csi-test-pod-name> -- pytest -m "volumesnapshot and remotecluster"
