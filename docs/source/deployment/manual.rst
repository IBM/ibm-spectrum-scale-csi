Manual
------

.. note:: For OpenShift environments, replace ``kubectl`` with  ``oc``

The following ``.yaml`` files needs to be applied to your cluster 


namespace.yaml
    This configuration file creates the ``ibm-spectrum-scale-csi-driver`` namespace.

ibm-spectrum-scale-csi-operator.yaml
    This is an auto-generated combined configuration file that starts the operator pods.

ibm-spectrum-scale-csi-operator-cr.yaml
    This is a custom resource file (CR) that the admin must modify to match their Spectrum Scale environment, which loads the csi-driver plugin.


Create the Operator
===================

1. Download and extract a ``.tar.gz`` file from `ibm-spectrum-scale-csi/releases <https://github.com/IBM/ibm-spectrum-scale-csi/releases/>`_ page.

2. Apply the namespace and operator configuration files.

  .. code-block:: bash

      kubectl apply -f namespace.yaml
      kubectl apply -f ibm-spectrum-scale-csi-operator.yaml

3. Create and apply the secret for the Spectrum Scale GUI.

  Create a file ``secret.json`` with the following, replacing the ``name|username|password`` fields. 

  .. code-block:: json
    
    {
        "apiVersion": "v1",
        "kind": "List",
        "items":
        [{
            "kind": "Secret",
            "apiVersion": "v1",
            "metadata": {
                "name": "<spectrum-scale-gui-secret>",
                "label": {
                    "app.kubernetes.io/name": "ibm-spectrum-scale-csi-operator"
                }
            },
            "data": {
                "username": "<base64_username>",
                "password": "<base64_password>"
            }
        }]
    }

  Then apply with the following command:

  .. code-block:: bash

    kubectl apply -f secret.json 

4. Edit and apply the ``ibm-spectrum-scale-csi-operator-cr.yaml`` file to start the csi-driver plugin.

  .. code-block:: bash

    # Modify this file to match your environment properties
    kubectl apply -f ibm-spectrum-scale-csi-operator-cr.yaml

Delete the Operator 
===================

1. To remove the operator, run ``delete`` of the yaml files in the following order: 

  .. code-block:: bash

      kubectl delete -f ibm-spectrum-scale-csi-operator-cr.yaml
      kubectl delete -f ibm-spectrum-scale-csi-operator.yaml
      kubectl delete -f namespace.yaml