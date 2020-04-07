# ibm-spectrum-scale-csi-operator

[Official Knowledge Center Documentation] (https://www.ibm.com/support/knowledgecenter/STXKQY_CSI_SHR/ibmspectrumscalecsi_welcome.html)

An operator for deploying and managing the IBM CSI Spectrum Scale Driver.

## Spectrum Scale GUI Setup

*NOTE:* This step only needs to be preformed once per GUI.

*WARNING:* If your daemonset pods (driver pods) do not come up, generally this means you have a secret that has not been defined in the correct namespace.

1. Ensure the Spectrum Scale GUI is running by pointing your browser to the IP hosting the GUI:

> If you do not see a login follow on screen instructions, or review the [GUI Documentation]( https://www.ibm.com/support/knowledgecenter/en/STXKQY_5.0.4/com.ibm.spectrum.scale.v5r04.doc/bl1ins_quickrefforgui.htm)

2. Create a CsiAdmin group account on in the GUI (currently requires a CLI call):

  ```
   export USERNAME="SomeUser"
   export PASSWORD="SomePassword"
   /usr/lpp/mmfs/gui/cli/mkuser ${USERNAME} -p ${PASSWORD} -g CsiAdmin
  ```

3. Create a Kubernetes secret for the CsiAdmin user:
 ```
  export USERNAME_B64=$(echo $USERNAME | base64)
  export PASSWORD_B64=$(echo $PASSWORD | base64)
  export OPERATOR_NAMESPACE="ibm-spectrum-scale-csi-driver"  # Set this to the namespace you deploy the operator in.


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

  ```
