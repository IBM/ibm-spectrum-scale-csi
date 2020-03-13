Prerequisites 
=============

1. Ensure the Spectrum Scale GUI is running by pointing your browser to the GUI IP address:

    .. image:: images/scale-gui-login.png
        :alt: Spectrum Scale GUI Login

If you do not see a login or on-screen instructions, review the `GUI Documentation <https://www.ibm.com/support/knowledgecenter/en/STXKQY_5.0.3/com.ibm.spectrum.scale.v5r03.doc/bl1ins_quickrefforgui.htm>`_ here.


2. Create a ``CsiAdmin`` group account.

.. code-block:: bash

   export USERNAME="SomeUser"
   export PASSWORD="SomePassword"
   /usr/lpp/mmfs/gui/cli/mkuser ${USERNAME} -p ${PASSWORD} -g CsiAdmin


.. tip:: If the user already exists, use ``chuser`` command to add the group to the existing user


3. Create a Kubernetes secret for the ``CsiAdmin`` user:

.. code-block:: bash

  export USERNAME_B64=$(echo $USERNAME | base64)
  export PASSWORD_B64=$(echo $PASSWORD | base64)

  # Set the following to the target namespace to deploy the operator in.
  export OPERATOR_NAMESPACE="SomeNamespace" 
  
  cat << EOF > /tmp/csisecret.yaml
  apiVersion: v1
  data:
    password: ${PASSWORD_B64}
    username: ${USERNAME_B64}
  kind: Secret
  type: Opaque
  metadata:
    name: csisecret    # This should be in your CSIScaleOperator definition
    namespace: ${OPERATOR_NAMESPACE} 
    labels:
      app.kubernetes.io/name: ibm-spectrum-scale-csi-operator # Used by the operator to detect changes, set on load of CR change if secret matches name in CR and namespace.
  EOF
  

  kubectl create -f /tmp/csisecret.yaml
  rm -f /tmp/csisecret.yaml
  
