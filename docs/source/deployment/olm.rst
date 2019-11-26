Operator Lifecycle Manager (OLM)
--------------------------------

1. Install OLM:

.. code-block:: bash

    curl -sL https://github.com/operator-framework/operator-lifecycle-manager/releases/download/0.11.0/install.sh | bash -s 0.11.0


2. Download the CSI Operator ``.yaml`` and apply

.. code-block:: bash

    curl https://raw.githubusercontent.com/IBM/ibm-spectrum-scale-csi-operator/master/stable/ibm-spectrum-scale-csi-operator-bundle/operators/ibm-spectrum-scale-csi-operator/deploy/olm-scripts/operator-source.yaml > operator-source.yaml

    kubectl apply -f operator-source.yaml

.. note:: For OpenShift environments, replace ``kubectl`` with  ``oc``.

