# Functional Test Automation Suite

This Functional Test Automation Suite exercises and tests the IBM Spectrum Scale CSI functionality.

### Tested Testbed environment

- IBM Spectrum Scale Cluster - 5.1.0.1+ Version  (**IBM Spectrum Scale supported kernel version**)
- Kubernetes Cluster Version 1.19 - 1.22
- Openshift Version 4.6.x, 4.7.x, 4.8.x
- IBM Spectrum Scale Cluster CSI Version - 2.1.0+


### How to run IBM Spectrum Scale CSI test automation using container

- Configure parameters in [csiscaleoperators.csi.ibm.com_cr.yaml](../../operator/config/samples/csiscaleoperators.csi.ibm.com_cr.yaml) file.

- Configuring parameters such as uid/gid, secret username/password, cacert path & remote/local filesystem name, you must modify relevant fields in [test.config](./config/test.config) file.

Note: Use `oc` command instead of `kubectl` in case of Openshift Container Platform 

##### If IBM Spectrum Scale CSI operator and driver is already deployed.
- Create config map using [test.config](./config/test.config) & kubeconfig file (default location `~/.kube/config`)

```
kubectl create configmap  test-config  --from-file=test.config=<test.config file path>  --from-file=kubeconfig=<kubeconfig file path>

```

```
eg. kubectl create configmap  test-config  --from-file=test.config=/root/ibm-spectrum-scale-csi/tests/functional-tests/config/test.config  --from-file=kubeconfig=/root/.kube/config

```

#### If IBM Spectrum Scale CSI operator and driver is not deployed
- Create config map using [test.config](./config/test.config),[csiscaleoperators.csi.ibm.com_cr.yaml](../..//operator/config/samples/csiscaleoperators.csi.ibm.com_cr.yaml) & kubeconfig file (default location `~/.kube/config`)

```
kubectl create configmap  test-config  --from-file=test.config=<test.config file path>  --from-file=csiscaleoperators.csi.ibm.com_cr.yaml=<csiscaleoperators.csi.ibm.com_cr.yaml file path> --from-file=kubeconfig=<kubeconfig file path>

```

```
eg. kubectl create configmap  test-config  --from-file=test.config=/root/ibm-spectrum-scale-csi/tests/functional-tests/config/test.config  --from-file=csiscaleoperators.csi.ibm.com_cr.yaml=/root/ibm-spectrum-scale-csi/operator/config/samples/csiscaleoperators.csi.ibm.com_cr.yaml --from-file=kubeconfig=/root/.kube/config
 
```
- To run the tests in namespace other than ibm-spectrum-scale-csi-driver (default) , please use --testnamespace parameter
- If operator is not running in namespace ibm-spectrum-scale-csi-driver (default) , please use --operatornamespace with value where operator is already running

- if you want to use SSL=enable, for cacert configmap use following command and change the path in test.config file as `config/local.crt`
```
kubectl create configmap  test-config  --from-file=test.config=<test.config filepath>  --from-file=csiscaleoperators.csi.ibm.com_cr.yaml=<csiscaleoperators.csi.ibm.com_cr.yaml file path> --from-file=kubeconfig=<kubeconfig file path> --from-file=local.crt=<local.crt file path>
```
Note : for remote crt, pass remote.crt file in the same configmap and user in the test.config file

- Configure sample [csi-test-pod.yaml](./csi-test-pod.yaml) file 

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
    volumeMounts:
    - mountPath: /data
      name: report
    - mountPath: /ibm-spectrum-scale-csi/tests/functional-tests/config
      name: test-config
  volumes:
  - configMap:
      defaultMode: 420
      name: test-config  
    name: test-config
  - hostPath:
      path: /ibm/gpfs0    #local path for saving the test reports files on node of running pod
    name: report
  restartPolicy: "Never"
  nodeSelector:
     kubernetes.io/hostname: <Worker Node name> #node selector for scheduling the test pod

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
kubectl exec -it <csi-test-pod-name> -- pytest  driver_test.py --html=/data/<report-name>.html

#Run any specific testcase in testsuite using:
kubectl exec -it <csi-test-pod-name> -- pytest  driver_test.py::<test_name> --html=/data/<report-name>.html

eg. kubectl exec -it <csi-test-pod-name> -- pytest  driver_test.py::test_driver_dynamic_pass_1 --html=/data/<report-name>.html
```
                
### Run operator tests using operator_test.py as shown below -
```       

#Run all testcases in operator testsuite using:
kubectl exec -it <csi-test-pod-name> -- pytest  operator_test.py --html=/data/<report-name>.html

#Run any specific testcase in testsuite using:
kubectl exec -it <csi-test-pod-name> -- pytest  operator_test.py::<test_name>  --html=/data/<report-name>.html

eg. kubectl exec -it <csi-test-pod-name> -- pytest  operator_test.py::test_operator_deploy --html=/data/<report-name>.html
```

### Run driver tests on remote cluster using remote_test.py as shown below -
```

#Run all testcases in driver testsuite using:
kubectl exec -it <csi-test-pod-name> -- pytest  remote_test.py --html=/data/<report-name>.html

#Run any specific testcase in testsuite using:
kubectl exec -it <csi-test-pod-name> -- pytest  remote_test.py::<test_name> --html=/data/<report-name>.html

eg. kubectl exec -it <csi-test-pod-name> -- pytest  remote_test.py::test_driver_dynamic_pass_1 --html=/data/<report-name>.html
```

### List available Driver & Operator tests 
Available functional tests list for driver & operator can be collected using following command
```
kubectl exec -it <csi-test-pod-name> -- pytest --collect-only 
```
### For running full testsuite with the tests which take long time, use `--runslow` parameter ( these tests are being marked with @pytest.mark.slow).
For example :

```
kubectl exec -it <csi-test-pod-name> -- pytest snapshot_test.py --runslow --html=/data/<report-name>.html   #This will run all testcases including those marked with slow
kubectl exec -it <csi-test-pod-name> -- pytest snapshot_test.py::test_snapshot_dynamic_multiple_snapshots_256 --runslow --html=/data/<report-name>.html
```
### Run specific testcases using marker
```
pytest driver_test.py -m marker_name

eg. kubectl exec -it <csi-test-pod-name> -- pytest driver_test.py -m regression --html=/data/<report-name>.html
```
### Run Stress/Load testcases using stresstest marker and runslow
```
pytest driver_test.py -m "stresstest" --runslow --workers 6 #(ensure driver is already deployed before using --workers)

eg. kubectl exec -it <csi-test-pod-name> -- pytest driver_test.py -m "stresstest" --runslow --workers 6 --html=/data/<report-name>.html
```
