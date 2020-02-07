Secrets
=======

The IBM Spectrum Scale CSI Driver leverages secrets to store API authentication. In the event
of an authentication going stale the user will need to update the secret in kubernetes.

Updating a Secret
-----------------

.. note:: For OpenShift environments, replace ``kubectl`` with  ``oc``

Due to `ansible-operator` constraints when updating a secret `kubectl apply` and `kubectl edit` 
are not usable at this time. To update the secret and have the operator  apply it, please follow
the folowing steps:

1. Edit the  `json` or `yaml` defining your secret to have the updated authentication information.

   ... code-block:: bash
      export SECRET_NAME="mysecret"
      export NAMESPACE="ibm-spectrum-scale-csi-driver"

      # Note if you still have a json or yaml file you can just edit that.
      kubectl get secret -n ${NAMESPACE} ${SECRET_NAME} -o yaml > secret.yaml

      # Edit the contents of secret.yaml to be up to date.

2. Ensure the secret has the correct labelling. If the label is not set the operator will not trigger.

   ... code-block:: yaml

      metadata:
        labels:
          app.kubernetes.io/name: ibm-spectrum-scale-csi-operator

3. Delete the old secret and apply the updated secret configuration.

   ... code-block:: bash

      kubectl delete secret -n ${NAMESPACE} ${SECRET_NAME}
      kubectl apply -f secret.yaml


After running the fresh apply you should see the  `spec.trigger` field increment if the secret
was sucessfully created. The process may then be monitored in operator logs.

Additionally, if the operator's custom resource was deployed before the secrets were created the 
above process may be leveraged to start the operator without  deleting the Custom Resource.
