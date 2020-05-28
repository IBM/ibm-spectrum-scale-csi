Clone and Build
===============

Clone
-----

Clone down the repository. This repository needs to be accessible in your ``GOPATH``. The examples below utilize the ``root`` user with ``GOPATH=/root/go``


.. code-block:: bash

  # Set up some helpful variables
  export GOPATH="/root/go"
  export IBM_DIR="$GOPATH/src/github.com/IBM"

  # Ensure the dir is present then clone.
  mkdir -p ${IBM_DIR}
  cd ${IBM_DIR}
  git clone https://github.com/IBM/ibm-spectrum-scale-csi.git

.. warning:: Due to current constraints in golang, relative paths are not supported.  You **must** clone this repository under your ``GOPATH``.


Build
-----

.. note:: Builds requires ``docker`` 17.05 and later. 


Operator
````````

The operator build requires ``operator-sdk``.  

.. tip:: To assist in proper configuration of the build environment, a playbook is provided.  ``ansible-playbook ${IBM_DIR}/ibm-spectrum-scale-csi/tools/ansible/dev-env-playbook.yaml``

1. Navigate to the ``operator`` directory and use ``operator-sdk`` to build the operator container image.

.. code-block:: bash

  # IBM_DIR is defined in the previous steps
  export REPO_DIR="${IBM_DIR}/ibm-spectrum-scale-csi"
  export OPERATOR_DIR="${REPO_DIR}/operator"
  cd ${OPERATOR_DIR}

  export GO111MODULE="on"

  # Build the container image
  operator-sdk build ibm-spectrum-scale-csi-operator


Driver
``````

1. Navigate to the ``driver`` directory and use ``docker`` to build the driver container image. 

.. code-block:: bash

  # IBM_DIR is defined in the previous steps
  export REPO_DIR="${IBM_DIR}/ibm-spectrum-scale-csi"
  export DRIVER_DIR="${REPO_DIR}/driver"
  cd ${DRIVER_DIR}

  # Build the container image 
  VERSION="v2.0.0"
  docker build -t ibm-spectrum-scale-csi:${VERSION} .

  # Save the image into a .tar file
  docker save ibm-spectrum-scale-csi:${VERSION} -o ibm-spectrum-scale-csi_${VERSION}.tar

