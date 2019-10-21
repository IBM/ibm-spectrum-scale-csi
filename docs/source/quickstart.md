# IBM Spectrum Scale CSI Operator Quickstart

The IBM Spectrum Scale CSI Operator runs within a Kubernetes cluster providing a means to 
deploy and manage the CSI plugin for spectrum scale.

This operator should be used to deploy the CSI plugin.

The configuration process is as follows:

1. (#spectrum-scale-gui-setup)
2.


## Spectrum Scale GUI Setup
1. Ensure the Spectrum Scale GUI is running by pointing your browser to the IP hosting the GUI:
  
  ![Spectrum Scale GUI login](https://user-images.githubusercontent.com/1195452/67230992-6d2d9700-f40c-11e9-96d5-3f0e5bcb2d9a.png)


and an account in the CsiAdmin group has been created

```
  export USERNAME="SomeUser"
  export PASSWORD="SomePassword"
  /usr/lpp/mmfs/gui/cli/mkuser ${USERNAME} -p ${PASSWORD} -g CsiAdmin
```
