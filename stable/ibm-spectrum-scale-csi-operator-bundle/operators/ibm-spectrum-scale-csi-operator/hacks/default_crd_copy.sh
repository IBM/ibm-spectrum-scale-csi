#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd ${DIR}/..

OP_CRD="deploy/crds/ibm_v1alpha1_csiscaleoperator_crd.yaml"
OP_CSV="deploy/olm-catalog/ibm-spectrum-scale-csi-operator/0.0.1/ibm-spectrum-scale-csi-operator.v0.0.1.clusterserviceversion.yaml"

hacks/copy_crd_descriptions.py --crd ${OP_CRD} --csv ${OP_CSV}
hacks/copy_docs_csv.py

