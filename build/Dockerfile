#FROM registry.access.redhat.com/ubi7/ubi-minimal:latest
FROM quay.io/operator-framework/ansible-operator:v0.9.0

#ENV OPERATOR=/usr/local/bin/ibm-spectrum-scale-csi-operator \
#    USER_UID=1001 \
#    USER_NAME=ibm-spectrum-scale-csi-operator\
#    HOME=/opt/ansible\
#    RUNNER_BASE_COMMAND=ansible-playbook

#COPY build/bin /usr/local/bin
COPY build/_output/bin/ibm-spectrum-scale-csi-operator ${OPERATOR}
COPY watches.yaml ${HOME}/watches.yaml
COPY roles/ ${HOME}/roles/
