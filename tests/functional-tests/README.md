# IBM Spectrum Scale CSI - Automation

The IBM Spectrum Scale Container Storage Interface (CSI) project enables container orchestrators, such as Kubernetes and OpenShift, to manage the life-cycle of persistent storage.

This project contains an automation framework to exercise the IBM Spectrum Scale CSI Driver and Operator functionality .


### Testbed environment 

- IBM Spectrum Scale Cluster - 5.0.4.1+ Version  (**IBM Spectrum Scale supported kernel version**)
- Kubernetes Cluster Version 1.14 - 1.18
- Openshift Version 4.2.x,4.3.x
- IBM Spectrum Scale Cluster CSI Version - 1.1.0+

### Pre-requesite for automation framework

Install Python (3.7.4 or higher) and below mentioned pip modules:

  ``` 
        python3.7 -m pip install kubernetes
        python3.7 -m pip install pytest
        python3.7 -m pip install pytest-html
        python3.7 -m pip install jsmin
  ```
       

### How to run IBM Spectrum Scale CSI test automation
- Clone the source code from [git](https://github.com/IBM/ibm-spectrum-scale-csi/) repository.

- Configure parameters in [csi-automation/scale_operator/config.json](https://github.com/IBM/ibm-spectrum-scale-csi/tests/functional-tests/config.json) file. 

- Set node labels for "attacherNodeSelector","provisionerNodeSelector","pluginNodeSelector" in config.json file


- cluster-config parameter is mandatory to pass the SpectrumScale & CSI Operator/Driver images configuration as input.

                For example :
                   pytest  driver_test.py --clusterconfig config.json
                   
- If kubeconfig file is at not at ~/.kube/config location, pass the correct location with --kubeconfig (optional)               
                
                For example :
                   pytest  driver_test.py --clusterconfig config.json --kubeconfig <kubeconfig_file_path>

- For generating the report in html format, pass the --html (optional)               
                
                For example :
                   pytest  driver_test.py --clusterconfig config.json --html report.html
                   
                
### Run driver tests using driver_test.py as shown below -
       
                cd csi-automation/scale_operator

                #Run all testcases in driver testsuite using:
                pytest  driver_test.py --clusterconfig config.json

                #Run any specific testcase in testsuite using:
                pytest  driver_test.py::<test_name> --clusterconfig config.json

                eg. pytest  driver_test.py::test_driver_pass_1 --clusterconfig config.json
                
### Run operator tests using operator_test.py as shown below -
       
                cd csi-automation/scale_operator
                
                #Run all testcases in operator testsuite using:
                pytest  operator_test.py --clusterconfig config.json
                
                #cluster-config parameter is mandatory to pass the spectrum scale configuration as input.
                pytest  operator_test.py --clusterconfig config.json

                #Run any specific testcase in testsuite using:
                pytest  operator_test.py::<test_name> --clusterconfig config.json

                eg. pytest  operator_test.py::test_operator_deploy --clusterconfig config.json

