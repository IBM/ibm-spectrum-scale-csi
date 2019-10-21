IBM Spectrum Scale CSI Operator Quickstart
==========================================

The IBM Spectrum Scale CSI Operator runs within a Kubernetes cluster providing a means to 
deploy and manage the CSI plugin for spectrum scale.

This operator should be used to deploy the CSI plugin.

The configuration process is as follows:

1. [Spectrum Scale GUI Setup](#spectrum-scale-gui-setup)

Spectrum Scale GUI Setup 
------------------------
> **NOTE:** This step only needs to be preformed once per GUI.

1. Ensure the Spectrum Scale GUI is running by pointing your browser to the IP hosting the GUI:
    ![Spectrum Scale GUI login](https://user-images.githubusercontent.com/1195452/67230992-6d2d9700-f40c-11e9-96d5-3f0e5bcb2d9a.png)
    > If you do not see a login follow on screen instructions, or review the [GUI Documentation](https://www.ibm.com/support/knowledgecenter/en/STXKQY_5.0.3/com.ibm.spectrum.scale.v5r03.doc/bl1ins_quickrefforgui.htm)


2. Create a CsiAdmin group account on in the GUI (currently requires a CLI call):
   ``` bash
   export USERNAME="SomeUser"
   export PASSWORD="SomePassword"
   /usr/lpp/mmfs/gui/cli/mkuser ${USERNAME} -p ${PASSWORD} -g CsiAdmin
   ```

3. Create a Kubernetes secret for the `CsiAdmin` user:
  ``` bash
  export USERNAME_B64=$(echo $USERNAME | base64)
  export PASSWORD_B64=$(echo $PASSWORD | base64)
    
  cat << EOF > /tmp/csisecret.yaml
  apiVersion: v1
  data:
    password: ${PASSWORD_B64}
    username: ${USERNAME_B64}
  kind: Secret
  type: Opaque
  metadata:
    name: csisecret    # This should be in your CSIScaleOperator definition
    namespace: default # Set this to your operator namespace
  EOF
  
  kubectl create -f /tmp/csisecret.yaml
  rm -f /tmp/csisecret.yaml
  
  ```
