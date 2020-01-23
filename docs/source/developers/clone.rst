Clone and Build
===============

Clone
-----


.. warning:: This repository needs to be accessible in your ``GOPATH``. The examples use the ``root`` user and ``GOPATH=/root/go``

.. warning:: Due to current constraints in golang, relative paths are not supported.  You **must** clone this repository under your ``GOPATH``.  If not, the ``operator-sdk`` build operation may fail.

.. code-block:: bash
  :linenos:

  # Set up some helpful variables
  export GOPATH="/root/go"
  export IBM_DIR="$GOPATH/src/github.com/IBM"

  # Ensure the dir is present then clone.
  mkdir -p ${IBM_DIR}
  cd ${IBM_DIR}
  git clone https://github.com/IBM/ibm-spectrum-scale-csi-operator.git

Build
-----

Environment
```````````

To assist in proper configuration of the build environment, a playbook is provided:

.. code-block:: bash

  ansible-playbook $GOPATH/src/github.com/IBM/ibm-spectrum-scale-csi-operator/ansible/dev-env-playbook.yaml


Create the the Image
````````````````````

Navigate to the operator directory and use ``operator-sdk`` to build the container image.

.. code-block:: bash

  # IBM_DIR is defined in the previous step
  export OPERATOR_DIR="$IBM_DIR/ibm-spectrum-scale-csi-operator"
  cd ${OPERATOR_DIR}/stable/ibm-spectrum-scale-csi-operator-bundle/operators/ibm-spectrum-scale-csi-operator

  export GO111MODULE="on"
  operator-sdk build ibm-spectrum-scale-csi-operator

.. note:: This requires ``docker``
