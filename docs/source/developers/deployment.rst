
Deployment
==========

Container Repository
--------------------

In order to consume the csi-driver and csi-operator container images built in the previous steps, the images should be pushed to a container repository.

* **Quay.io (recommended)**

  Follow this tutorial to configure `quay.io <https://quay.io/tutorial/>`_.
  
  Create two repositories: ``ibm-spectrum-scale-csi-operator`` and ``ibm-spectrum-scale-csi-driver``.

* **Docker** 

  Deploying your own Docker registry is an `involved process <https://docs.docker.com/registry/deploying/>`_ and outside of the scope of this document. 

The documentation will assume that the quay.io path is being used. 

Pushing the image
-----------------

Once you have a repository ready:

.. code-block:: bash

  #
  # Configure some variables
  #
  # VERSION - a tag version for your image
  VERSION="v0.0.1"
  # MYUSER  - A user or organization for your container registry
  MYUSER="<your-user>"

  # Authenticate to quay.io
  docker login <credentials> quay.io

  # Tag and push the operator image 
  docker tag ibm-spectrum-scale-csi-operator quay.io/${MYUSER}/ibm-spectrum-scale-csi-operator:${VERSION}
  docker push quay.io/${MYUSER}/ibm-spectrum-scale-csi-operator:${VERSION}

  # Tag and push the driver image
  docker tag ibm-spectrum-scale-csi-driver quay.io/${MYUSER}/ibm-spectrum-scale-csi-driver:${VERSION}
  docker push quay.io/${MYUSER}/ibm-spectrum-scale-csi-driver:${VERSION}

  # OPERATOR_DIR has been defined in previous steps
  cd ${OPERATOR_DIR}
  # Use a helper script to update your deployment to point at your operator image
  ansible-playbook hacks/change_deploy_image.yml --extra-vars "quay_operator_endpoint=quay.io/${MYUSER}/ibm-spectrum-scale-csi-operator:${VERSION}"
  

Installing the CSI Operator
---------------------------

.. note:: For OpenShift environments, replace ``kubectl`` with  ``oc``

Run the following to deploy the IBM Spectrum Scale CSI operator manually:

.. code-block:: bash

  # OPERATOR_DIR has been defined in the previous steps
  kubectl apply -f ${OPERATOR_DIR}/deploy/namespace.yaml
  kubectl apply -f ${OPERATOR_DIR}/deploy/operator.yaml
  kubectl apply -f ${OPERATOR_DIR}/deploy/role.yaml
  kubectl apply -f ${OPERATOR_DIR}/deploy/role_binding.yaml
  kubectl apply -f ${OPERATOR_DIR}/deploy/service_account.yaml
  kubectl apply -f ${OPERATOR_DIR}/deploy/crds/csiscaleoperators.csi.ibm.com.crd.yaml
  
  
Installing the CSI Driver
-------------------------

.. tip:: Before starting the plugin, ensure that any GUI secrets have been added to the appropriate namespace. 

A Custom Resource (CR) file is provided `csiscaleoperators.csi.ibm.com.cr.yaml <https://raw.githubusercontent.com/IBM/ibm-spectrum-scale-csi/master/operator/deploy/crds/csiscaleoperators.csi.ibm.com.cr.yaml>`_. Modify this file to match the properties in your environment.

To start: 

.. code-block:: bash

  kubectl apply -f ${OPERATOR_DIR}/deploy/crds/csiscaleoperators.csi.ibm.com.cr.yaml


To stop:

.. code-block:: bash

  kubectl delete -f ${OPERATOR_DIR}/deploy/crds/csiscaleoperators.csi.ibm.com.cr.yaml

Removing the CSI Operator and Driver
------------------------------------

To remove the IBM Spectrum Scale CSI Operator and Driver:

.. code-block:: bash

  # The following removes the csi-driver
  kubectl delete -f ${OPERATOR_DIR}/deploy/crds/csiscaleoperators.csi.ibm.com.cr.yaml

  # The following removes the csi-operator
  kubectl delete -f ${OPERATOR_DIR}/deploy/operator.yaml
  kubectl delete -f ${OPERATOR_DIR}/deploy/role.yaml
  kubectl delete -f ${OPERATOR_DIR}/deploy/role_binding.yaml
  kubectl delete -f ${OPERATOR_DIR}/deploy/service_account.yaml
  kubectl delete -f ${OPERATOR_DIR}/deploy/crds/csiscaleoperators.csi.ibm.com.crd.yaml

  # The following removes the namespace 
  kubectl delete -f ${OPERATOR_DIR}/deploy/namespace.yaml


This will completely destroy the operator, driver, and all associated resources.
