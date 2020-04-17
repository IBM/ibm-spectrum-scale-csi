# Functional Test Automation Suite

This Functional Test Automation Suite exercises and tests the IBM Spectrum Scale CSI functionality.

### Tested Testbed environment

- IBM Spectrum Scale Cluster - 5.0.4.1+ Version  (**IBM Spectrum Scale supported kernel version**)
- Kubernetes Cluster Version 1.14 - 1.18
- Openshift Version 4.2.x,4.3.x
- IBM Spectrum Scale Cluster CSI Version - 1.1.0+

### Pre-requesite for automation framework

Install Python (3.7.4 or higher) and use below command to install required python plugins from requirements.txt:

``` 
python3.7 -m pip install -r requirements.txt
```

### How to run IBM Spectrum Scale CSI test automation

- Configure parameters in [config.json](./tests/functional-test/config.json) file.

- Define kubernetes/Openshift Spectrum Scale node labels for `attacherNodeSelector`,`provisionerNodeSelector`,`pluginNodeSelector` in config.json file

- `--clusterconfig` parameter is mandatory to pass IBM Spectrum Scale API Credentials & CSI Operator/Driver image specific configuration as input.For example :
```
pytest  driver_test.py --clusterconfig config.json
```                   
- If kubeconfig file is at not at `~/.kube/config` location, pass the correct location with `--kubeconfig` (optional).For example :
```
pytest  driver_test.py --clusterconfig config.json --kubeconfig <kubeconfig_file_path>
```
- For generating the report in html format, pass the `--html` (optional).For example :
```
pytest  driver_test.py --clusterconfig config.json --html report.html
```

### Run driver tests using driver_test.py as shown below -
```
cd ./tests/functional-tests/

#Run all testcases in driver testsuite using:
pytest  driver_test.py --clusterconfig config.json

#Run any specific testcase in testsuite using:
pytest  driver_test.py::<test_name> --clusterconfig config.json

eg. pytest  driver_test.py::test_driver_pass_1 --clusterconfig config.json
```
                
### Run operator tests using operator_test.py as shown below -
```       
cd ./tests/functional-tests/

#Run all testcases in operator testsuite using:
pytest  operator_test.py --clusterconfig config.json

#Run any specific testcase in testsuite using:
pytest  operator_test.py::<test_name> --clusterconfig config.json

eg. pytest  operator_test.py::test_operator_deploy --clusterconfig config.json
```
