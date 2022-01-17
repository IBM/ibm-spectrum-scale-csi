# Functional Test Automation Suite

This Functional Test Automation Suite exercises and tests the IBM Spectrum Scale CSI functionality.

### Tested Testbed environment

- IBM Spectrum Scale Cluster - 5.1.0.1+ Version  (**IBM Spectrum Scale supported kernel version**)
- Kubernetes Cluster Version 1.20 - 1.22
- Openshift Version 4.8.x, 4.9.x
- IBM Spectrum Scale Cluster CSI Version - 2.4.0+

### Pre-requesite for automation framework

Install Python (3.7.4 or higher) and use below command to install required python plugins from requirements.txt:

``` 
python3.7 -m pip install -r requirements.txt
```

### How to run IBM Spectrum Scale CSI test automation

- Configure parameters in [csiscaleoperators.csi.ibm.com_cr.yaml](../../operator/config/samples/csiscaleoperators.csi.ibm.com_cr.yaml) file.


  If changed `csiscaleoperators.csi.ibm.com_cr.yaml` configuration file is not present at default path `./operator/config/samples/`, `csiscaleoperators.csi.ibm.com_cr.yaml` file path location must be passwd with `--clusterconfig` parameter. For example :
```
pytest  driver_test.py --clusterconfig /root/csiscaleoperators.csi.ibm.com_cr.yaml
```
- For configuring parameters such as uid/gid, secret username/password, cacert path & remote/local filesystem name, you must modify relevant fields in [test.config](./config/test.config) file.

- If kubeconfig file is at not at `~/.kube/config` location, pass the correct location with `--kubeconfig` (optional).For example :
```
pytest  driver_test.py --kubeconfig <kubeconfig_file_path>
```
- To run the tests in namespace other than ibm-spectrum-scale-csi-driver (default) , please use --testnamespace parameter 
- If operator is not running in namespace ibm-spectrum-scale-csi-driver (default) , please use --operatornamespace with value where operator is already running 
- If operator yaml file is not at ../../generated/installer/ibm-spectrum-scale-csi-operator-dev.yaml (default), please use --operatoryaml with value of operator yaml file path
- For changing the name of html report, pass the `--html` with remote file name (optional).For example :
```
pytest  driver_test.py --html report.html
```
- For running full testsuite with the tests which take long time, use `--runslow` parameter ( these tests are being marked with @pytest.mark.slow).For example :
```
pytest driver_test.py --runslow  #This will run all testcases including those marked with slow
pytest driver_test.py::test_driver_sequential_pvc --runslow
```
### Run driver tests on primary cluster using driver_test.py as shown below -
```
cd ./tests/functional-tests/

#Run all testcases in driver testsuite using:
pytest  driver_test.py 

#Run any specific testcase in testsuite using:
pytest  driver_test.py::<test_name> 

eg. pytest  driver_test.py::test_driver_dynamic_pass_1 

#Run specific testcases using marker 
pytest driver_test.py -m marker_name

eg. pytest driver_test.py -m regression
```
                
### Run operator tests using operator_test.py as shown below -
```       
cd ./tests/functional-tests/

#Run all testcases in operator testsuite using:
pytest  operator_test.py 

#Run any specific testcase in testsuite using:
pytest  operator_test.py::<test_name> 

eg. pytest  operator_test.py::test_operator_deploy 
```

### Run driver tests on remote cluster using remote_test.py as shown below -
```
cd ./tests/functional-tests/

#Run all testcases in driver testsuite using:
pytest  remote_test.py 

#Run any specific testcase in testsuite using:
pytest  remote_test.py::<test_name> 

eg. pytest  remote_test.py::test_driver_dynamic_pass_1
```

### Run Stress/Load testcases using stresstest marker and runslow
```
pytest driver_test.py -m "stresstest" --runslow --workers 6 #(ensure driver is already deployed before using --workers)
```

### List available Driver & Operator tests 
Available functional tests list for driver & operator can be collected using following command
```
pytest --collect-only
```

### Running Testcases using markers
1. Running all testcases except operator testcases
```
pytest -m "volumeprovisioning or volumesnapshot"
```
2. Running only local cluster testcases
```
pytest -m "localcluster"
```
3. Running only remove cluster testcases
```
pytest -m "remotecluster"
```
4. Running only operator testcases
```
pytest -m "csioperator"
```
5. Running only localcluster volumeprovisioning testcases
```
pytest -m "volumeprovisioning and localcluster"
```
like above format you can combine volumeprovisioning, volumesnapshot, localcluster and remotecluster markers.
Few examples,
```
pytest -m "volumeprovisioning and remotecluster"
pytest -m "volumesnapshot and localcluster"
pytest -m "volumesnapshot and remotecluster"
```
