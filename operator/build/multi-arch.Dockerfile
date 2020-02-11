# Multi-arch build for IBM Spectrum Scale CSI Operator
# usage: docker buildx build --platform=linux/amd64 -t my_image_tag .

FROM quay.io/mew2057/ansible-operator:$TARGETARCH
MAINTAINER jdunham@us.ibm.com

LABEL name="IBM Spectrum Scale CSI Operator" \
      vendor="ibm" \
      version="1.0.1" \
      release="1" \
      run='docker run ibm-spectrum-scale-csi-operator' \
      summary="An Ansible based operator to run and manage the deployment of the IBM Spectrum Scale CSI Driver." \
      description="An Ansible based operator to run and manage the deployment of the IBM Spectrum Scale CSI Driver." 

COPY hacks/health_check.sh .
COPY licenses /licenses
COPY watches.yaml ${HOME}/watches.yaml
COPY roles/ ${HOME}/roles/

ENV CSI_ATTACHER_IMAGE=registry.scale-peach.fyre.ibm.com/external-attacher:latest
ENV CSI_PROVISIONER_IMAGE=registry.scale-peach.fyre.ibm.com/external-provisioner:latest
ENV CSI_NODE_REGISTRAR_IMAGE=registry.scale-peach.fyre.ibm.com/node-driver-registrar:latest
ENV CSI_DRIVER_IMAGE=registry.scale-peach.fyre.ibm.com/ibm-spectrum-scale-csi-driver:latest
