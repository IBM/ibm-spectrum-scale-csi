Installing the Operator
=======================

The recommended method of deploying/managing the IBM Spectrum Scale CSI Plugin is through the use of Operators.  The 
IBM Spectrum Scale CSI Operator can be installed from the `OperatorHub <https://operatorhub.io>`_ with 
`Operator Lifecycle Manager <https://github.com/operator-framework/operator-lifecycle-manager>`_ (OLM).  OLM is part of the 
`Operator Framework <https://github.com/operator-framework/getting-started#manage-the-operator-using-the-operator-lifecycle-manager>`_.
For more information, see: `How to install an Operator from OperatorHub <https://operatorhub.io/how-to-install-an-operator>`_


OpenShift 
---------

1. Log into the OpenShift Console.  On the right sidebar, under **"Operators"**, click **"OperatorHub"**

    .. image:: images/openshift-menu.png
        :width: 400
        :alt: OpenShift OperatorHub Menu

2. Search **"IBM Spectrum Scale CSI"** and click to **"Install"** to install the Operator.  

.. tip:: Some operators may have multiple icons appear in OperatorHub.  We recommend to filter on "Certified" Operators.

3. Validate the options for the operator and click **"Subscribe"** to complete the install of the Operator. 

Kubernetes
----------

1.  Navigate to `ibm-spectrum-scale-csi-operator <https://operatorhub.io/operator/ibm-spectrum-scale-csi-operator>`_ and follow 
the instructions that appear when you click on the **"Install"** button. 

