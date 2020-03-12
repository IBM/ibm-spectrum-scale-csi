Deploying the Driver
====================

Before you can deploy the driver, you need to modify the Custom Resource (CR) and set the properties matching your IBM Spectrum Scale install.

OpenShift
---------

1. To deploy the driver, select the **"IBM Spectrum Scale CSI Driver"** tab and click **"Create CSIScale Operator"**

    .. image:: images/operatorhub-driver-tab.png
        :alt: IBM Spectrum Scale CSI Operator Tabs

2. Modify the Custom Resource (CR) to match your running IBM Spectrum Scale properties, then click **"Create"**. 

    .. image:: images/operatorhub-custom-resource.png
        :alt: IBM Spectrum Scale CSI Operator Tabs

   For a complete sample of valid CR options, see `csiscaleoperators.csi.ibm.com.cr.yaml <https://raw.githubusercontent.com/IBM/ibm-spectrum-scale-csi/master/operator/deploy/crds/csiscaleoperators.csi.ibm.com.cr.yaml>`_

Kubernetes
----------

TODO