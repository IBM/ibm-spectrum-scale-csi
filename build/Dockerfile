#FROM registry.access.redhat.com/ubi7/ubi-minimal:latest
FROM quay.io/operator-framework/ansible-operator:v0.9.0

#ENV OPERATOR=/usr/local/bin/ibm-spectrum-scale-csi-operator \
#    USER_UID=1001 \
#    USER_NAME=ibm-spectrum-scale-csi-operator

# install operator binary
COPY build/_output/bin/ibm-spectrum-scale-csi-operator ${OPERATOR}
COPY build/_output/bin/ibm-spectrum-scale-csi-operator /usr/local/bin/ibm-spectrum-scale-csi-operator

COPY build/bin /usr/local/bin
#RUN  /usr/local/bin/user_setup

# Copy the Roles and watches
COPY watches.yaml ${HOME}/watches.yaml
COPY roles/ ${HOME}/roles/

ENTRYPOINT ["/usr/local/bin/entrypoint"]

#USER ${USER_UID}
