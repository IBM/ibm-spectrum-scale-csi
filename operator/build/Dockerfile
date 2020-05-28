# Multi-arch build for IBM Spectrum Scale CSI Operator
# usage: docker buildx build -f build/Dockerfile --platform=linux/amd64 -t my_image_tag .

FROM quay.io/operator-framework/ansible-operator:v0.17.0
MAINTAINER jdunham@us.ibm.com

ARG CSI_ATTACHER_IMAGE
ARG CSI_PROVISIONER_IMAGE
ARG CSI_NODE_REGISTRAR_IMAGE
ARG CSI_DRIVER_IMAGE
ARG GIT_COMMIT


ENV CSI_ATTACHER_IMAGE $CSI_ATTACHER_IMAGE
ENV CSI_PROVISIONER_IMAGE $CSI_PROVISIONER_IMAGE
ENV CSI_NODE_REGISTRAR_IMAGE $CSI_NODE_REGISTRAR_IMAGE
ENV CSI_DRIVER_IMAGE $CSI_DRIVER_IMAGE

LABEL name="IBM Spectrum Scale CSI Operator" \
      vendor="ibm" \
      version="2.0.0" \
      release="1" \
      run='docker run ibm-spectrum-scale-csi-operator' \
      summary="An Ansible based operator to run and manage the deployment of the IBM Spectrum Scale CSI Driver." \
      description="An Ansible based operator to run and manage the deployment of the IBM Spectrum Scale CSI Driver." \
      git_commit=$GIT_COMMIT

COPY hacks/health_check.sh .
COPY licenses /licenses
COPY watches.yaml ${HOME}/watches.yaml
COPY roles/ ${HOME}/roles/
