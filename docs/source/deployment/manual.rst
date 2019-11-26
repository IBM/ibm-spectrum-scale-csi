Manual
------

.. note:: For OpenShift environments, replace ``kubectl`` with  ``oc``.

The following ``.yaml`` files needs to be applied to your cluster 


namespace.yaml
    This configuration file creates the ``ibm-spectrum-scale-csi-driver`` namespace

ibm-spectrum-scale-csi-operator.yaml
    This starts up the operator 

example/spectrum_scale.yaml
    This is a custom file that the user must modify to match their Spectrum Scale install, to load the csi-driver


1. Download the ``.yaml`` files from the code repository

.. code-block:: bash

    curl https://raw.githubusercontent.com/IBM/ibm-spectrum-scale-csi-operator/master/generated/installer/namespace.yaml > namespace.yaml
    curl https://raw.githubusercontent.com/IBM/ibm-spectrum-scale-csi-operator/master/generated/installer/ibm-spectrum-scale-csi-operator.yaml > ibm-spectrum-scale-csi-operator.yaml
    curl https://raw.githubusercontent.com/IBM/ibm-spectrum-scale-csi-operator/master/generated/installer/examples/spectrum_scale.yaml > spectrum_scale.yaml

2. Apply the operator ``.yaml`` files 

.. code-block:: bash

    kubectl apply -f namespace.yaml
    kubectl apply -f ibm-spectrum-scale-csi-operator.yaml

3. Create and apply the secret for the Spectrum Scale GUI

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
                "name": "{{ gui_secret_name }}",
                "label": {
                    "app.kubernetes.io/name": "ibm-spectrum-scale-csi-operator"
                }
            },
            "data": {
                "username": "{{ gui_user | b64encode }}",
                "password": "{{ gui_pass | b64encode }}"
            }
        }]
    }

Then apply with the following command:

.. code-block:: bash

    kubectl apply -f secret.json 

4. Edit and apply the ``spectrum_scale`` yaml file to start the csi-driver 

.. code-block:: bash

    # Modify this file to match your environment properties
    kubectl apply -f spectrum_scale.yaml
