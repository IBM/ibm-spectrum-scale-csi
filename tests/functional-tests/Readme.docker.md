# Functional Test Automation Suite

This Functional Test Automation Suite exercises and tests the IBM Spectrum Scale CSI functionality.

### Tested Testbed environment

- IBM Spectrum Scale Cluster - 5.0.4.2+ Version  (**IBM Spectrum Scale supported kernel version**)
- Kubernetes Cluster Version 1.14 - 1.18
- Openshift Version 4.3.x, 4.4.x , 4.5.x
- IBM Spectrum Scale Cluster CSI Version - 2.1.0+


### How to run IBM Spectrum Scale CSI test automation using container

- Configure parameters in [csiscaleoperators.csi.ibm.com_cr.yaml](./operator/deploy/crds/csiscaleoperators.csi.ibm.com_cr.yaml) file.

- Configuring parameters such as uid/gid, secret username/password, cacert path & remote/local filesystem name, you must modify relevant fields in [test.config](./tests/functional-tests/config/test.config) file.

Note: Use `oc` command instead of `kubectl` in case of Openshift Container Platform 

- Create config map using [test.config](./tests/functional-tests/config/test.config),[csiscaleoperators.csi.ibm.com_cr.yaml](./operator/deploy/crds/csiscaleoperators.csi.ibm.com_cr.yaml) & kubeconfig file (default location `~/.kube/config`)

```
kubectl create configmap  test-config  --from-file=test.config=<test.config file path>  --from-file=csiscaleoperators.csi.ibm.com_cr.yaml=<csiscaleoperators.csi.ibm.com_cr.yaml file path> --from-file=config=<kubeconfig file path>

```

- Configure sample [csi-test-pod.yaml](./tests/functional-tests/csi-test-pod.yaml) file 

```
kind: Pod
apiVersion: v1
metadata:
  name: ibm-spectrum-scale-csi-test-pod  
spec:
  containers:
  - name: csi-test
    image: quay.io/jainbrt/ibm-spectrum-scale-csi-test:x86 #container image for csi tests
    securityContext:
      privileged: true
    command: [ "/bin/sh", "-c", "--" ]
    args: [ "while true; do sleep 30; done;" ]
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

