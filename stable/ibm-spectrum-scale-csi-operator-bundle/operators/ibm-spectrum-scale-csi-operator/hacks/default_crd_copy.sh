#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd ${DIR}/..

CRD="ibm-spectrum-scale-csi-operator-crd.yaml"
OP_CRD="deploy/crds/${CRD}"
CAT="deploy/olm-catalog/ibm-spectrum-scale-csi-operator/1.0.0"
OP_CSV="${CAT}/ibm-spectrum-scale-csi-operator.v1.0.0.clusterserviceversion.yaml"

hacks/copy_crd_descriptions.py --crd ${OP_CRD} --csv ${OP_CSV}
hacks/copy_docs_csv.py
cp ${OP_CRD} "${CAT}/${CRD}"
