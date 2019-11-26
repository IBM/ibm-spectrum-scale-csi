
Deployment
==========

Container Repository
--------------------

In order to use the container image that you just built in the previous step, the image needs to be pushed to some container repository.

* **Quay.io (recommended)**

  Follow this tutorial to configure `quay.io <https://quay.io/tutorial/>`_.
  
  Create a repository called ``ibm-spectrum-scale-csi-operator``.

* **Docker** 

  Deploying your own Docker registry is an `involved process <https://docs.docker.com/registry/deploying/>`_ and outside of the scope of this document. 

The documentation will assume that the quay.io path is being used. 

Pushing the image
`````````````````

Once you have a repository ready:

.. code-block:: bash

  # Authenticate to quay.io
  docker login <credentials> quay.io

  # Tag the build 
  docker tag csi-scale-operator quay.io/<your-user>/ibm-spectrum-scale-csi-operator:v0.9.1

  # push the image
  docker push quay.io/<your-user>/ibm-spectrum-scale-csi-operator:v0.9.1

  # Update your deployment to point at your image.
  hacks/change_deploy_image.py -i quay.io/<your-user>/ibm-spectrum-scale-csi-operator:v0.9.1
  

Installing Operator
```````````````````

.. note:: For OpenShift environments, replace ``kubectl`` with  ``oc``.

Run the following to deploy the operator manually:

.. code-block:: bash

  cd ${OPERATOR_DIR}/stable/ibm-spectrum-scale-csi-operator-bundle/operators/ibm-spectrum-scale-csi-operator

  kubectl apply -f deploy/namespace.yaml
  kubectl apply -f deploy/operator.yaml
  kubectl apply -f deploy/role.yaml
  kubectl apply -f deploy/role_binding.yaml
  kubectl apply -f deploy/service_account.yaml
  kubectl apply -f deploy/crds/ibm_v1alpha1_csiscaleoperator_crd.yaml
  
  
Starting the CSI Driver
```````````````````````

.. note:: Before starting the plugin, add any GUI secrets to the appropriate namespace. 

A sample of the file is provided `examples/spectrum_scale.yaml <https://raw.githubusercontent.com/IBM/ibm-spectrum-scale-csi-operator/master/stable/ibm-spectrum-scale-csi-operator-bundle/operators/ibm-spectrum-scale-csi-operator/example/spectrum_scale.yaml>`_.

Modify this file to match the properties in your environment, then:

To start: 

.. code-block:: bash

  kubectl apply -f spectrum_scale.yaml


To stop:

.. code-block:: bash

  kubectl delete -f spectrum_scale.yaml

Removing the CSI Operator
`````````````````````````

To remove the operator:

.. code-block:: bash

  # The following removes the csi-driver
  kubectl delete -f deploy/spectrum_scale.yaml

  # The following removes the csi-operator
  kubectl delete -f deploy/operator.yaml
  kubectl delete -f deploy/role.yaml
  kubectl delete -f deploy/role_binding.yaml
  kubectl delete -f deploy/service_account.yaml
  kubectl delete -f deploy/crds/ibm_v1alpha1_csiscaleoperator_crd.yaml

  # The following removes the namespace 
  kubectl delete -f deploy/namespace.yaml
```

This will completely destroy the operator and all associated resources.
