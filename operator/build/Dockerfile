FROM quay.io/operator-framework/ansible-operator:v0.14.0
MAINTAINER jdunham@us.ibm.com

LABEL name="IBM Spectrum Scale CSI Operator" \
      vendor="ibm" \
      version="1.1.0" \
      release="1" \
      run='docker run ibm-spectrum-scale-csi-operator' \
      summary="An Ansible based operator to run and manage the deployment of the IBM Spectrum Scale CSI Driver." \
      description="An Ansible based operator to run and manage the deployment of the IBM Spectrum Scale CSI Driver." 

COPY hacks/health_check.sh .
COPY licenses /licenses
COPY watches.yaml ${HOME}/watches.yaml
COPY roles/ ${HOME}/roles/
