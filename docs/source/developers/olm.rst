OLM
===

Using Test Versions of CSV
--------------------------

Due to the nature of Operator Lifecycle Manager (OLM) it is necessary to maintain an application 
repository to host the most up to date Cluster Service Version (CSV). To assist, two application registries 
are maintained by the development team:  

* `Master - https://quay.io/application/ibm-spectrum-scale-dev/ibm-spectrum-scale-csi-master <https://quay.io/application/ibm-spectrum-scale-dev/ibm-spectrum-scale-csi-master>`_
* `Dev - https://quay.io/application/ibm-spectrum-scale-dev/ibm-spectrum-scale-csi-dev <https://quay.io/application/ibm-spectrum-scale-dev/ibm-spectrum-scale-csi-dev>`_

These subscriptions maintain the latest iteration of the CSV on the `dev <https://github.com/IBM/ibm-spectrum-scale-csi/tree/dev>`_ and `master <https://github.com/IBM/ibm-spectrum-scale-csi/tree/master>`_ branches respectively.
To subscribe to these applicaions via OLM, the code repository provides three YAML files:

``tools/olm/operator-source-openshift.yaml``

* Used for both applications on OpenShift.
* Created in the `openshift-marketplace` namespace.

``tools/olm/operator-source-k8s-master.yaml``

* Used for OLM subscription to the master stream in raw k8s.
* Created in the `marketplace` namespace.
* **WARNING** : Currently disabled, as master has some issues for upgrade.

``tools/olm/operator-source-k8s-dev.yaml``

* Used for OLM subscription to the master stream in raw k8s.
* Created in the `marketplace` namespace.

This yaml files should be applied against your Kubernetes or OpenShift cluster:

.. code-block:: bash
  
    kubectl apply -f <operator-source-____.yaml>

.. note:: For OpenShift environments, replace ``kubectl`` with  ``oc``

Testing an in development CSV
-----------------------------

While modifying a CSV it is conceivable that a developer would want to test their CSV in a local environment.
One method for achieving this is to host the CSV on `quay.io <https://quay.io>`_.

1. Create a new `Application Repository` in `quay.io/new <https://quay.io/new/>`_.

.. tip:: Save the name of this repository, because you'll need it in the next steps.

2. Install helm and helm registry:

  .. code-block::  bash
    
    curl -L https://git.io/get_helm.sh | bash
    helm init
    cd ~/.helm/plugins/ && git clone https://github.com/app-registry/appr-helm-plugin.git registry

3. Create a helm project for your application and push it to quay:

  .. code-block::  bash
  
    # Set your variables
    QUAY_REPO_NAME="<Your Repo Name>"
    QUAY_USER="<Your Quay Username>"
    CHANNEL_NAME="test"
    
    # Create the helm project
    cd ~
    helm create ${QUAY_REPO_NAME}
    cd ${QUAY_REPO_NAME}
    
    # Push to quay
    helm registry login quay.io
    helm registry push --namespace ${QUAY_USER} quay.io
    helm registry push --namespace ${QUAY_USER} --channel ${CHANNEL_NAME} quay.io

4. Edit the variables for the test playbook (which will push your csv):

  .. code-block:: bash
    
    vi tools/ansible/olm-test-playbook.yaml 
  

5. Deploy using `olm-test-playbook.yaml`, you'll need to set the user name and password:

  .. code-block:: bash
    cd tools/ansible/
    ansible-playbook olm-test-playbook.yaml --extra-vars '{"QUAY_PASSWORD":"A_TOKEN"}'

At this point your application is ready to be subscribed to.  Use the following templates for k8s and OpenShift respectively.

Kubernetes subscription template
++++++++++++++++++++++++++++++++

.. code-block:: yaml

  apiVersion: rbac.authorization.k8s.io/v1
  kind: ClusterRoleBinding
  metadata:
    name: olm-crb
  subjects:
  - kind: ServiceAccount
    name: default
    namespace: kube-system
  roleRef:
    kind: ClusterRole
    name: cluster-admin
    apiGroup: ""
  
  ---
  apiVersion: operators.coreos.com/v1
  kind: OperatorSource
  metadata:
    name: ibm-spectrum-scale-csi
    namespace: marketplace
  spec:
    type: appregistry
    endpoint: https://quay.io/cnr
    registryNamespace:  {{ QUAY_USER }}
  
  ---
  apiVersion: operators.coreos.com/v1
  kind: OperatorGroup
  metadata:
    name: operator-group
    namespace: marketplace
  spec:
    targetNamespaces:
    - marketplace
  
  ---
  apiVersion: operators.coreos.com/v1alpha1
  kind: Subscription
  metadata:
    name: oper-sub
    namespace: marketplace
  spec:
    channel: stable
    name: {{ REPO_NAME }}
    source: {{ REPO_NAME }}
    sourceNamespace: marketplace 

OpenShift subscription template
+++++++++++++++++++++++++++++++

.. code-block:: yaml

  apiVersion: operators.coreos.com/v1
  kind: OperatorSource
  metadata:
    name: ibm-spectrum-scale
    namespace: openshift-marketplace
  spec:
    type: appregistry
    endpoint: https://quay.io/cnr
    registryNamespace:  {{ QUAY_USER }}
    displayName: "CSI Scale Operator"
    publisher: "IBM"
