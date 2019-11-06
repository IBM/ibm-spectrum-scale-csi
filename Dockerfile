FROM registry.access.redhat.com/ubi7-minimal:latest
LABEL name="IBM Spectrum Scale CSI driver" \
      vendor="ibm" \
      version="0.9.0" \
      release="1" \
      run='docker run ibm-spectrum-scale-csi-driver' \
      summary="An implementation of CSI Plugin for the IBM Spectrum Scale product."\
      description="CSI Plugin for IBM Spectrum Scale"\
      maintainers="IBM Spectrum Scale"
COPY licenses /licenses

COPY _output/csi-spectrum-scale /csi-spectrum-scale
RUN chmod +x /csi-spectrum-scale
ENTRYPOINT ["/csi-spectrum-scale"]
