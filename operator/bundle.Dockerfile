FROM scratch

LABEL operators.operatorframework.io.bundle.mediatype.v1=registry+v1
LABEL operators.operatorframework.io.bundle.manifests.v1=manifests/
LABEL operators.operatorframework.io.bundle.metadata.v1=metadata/
LABEL operators.operatorframework.io.bundle.package.v1=operator
LABEL operators.operatorframework.io.bundle.channels.v1=stable
LABEL operators.operatorframework.io.bundle.channel.default.v1=stable

COPY config/olm-catalog/ibm-spectrum-scale-csi-operator/manifests /manifests/
COPY config/olm-catalog/ibm-spectrum-scale-csi-operator/metadata/annotations.yaml /metadata/annotations.yaml
