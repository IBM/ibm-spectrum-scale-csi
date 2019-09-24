FROM registry.access.redhat.com/ubi7-minimal:latest
LABEL maintainers="IBM Spectrum Scale"
LABEL description="CSI Plugin for IBM Spectrum Scale"

COPY _output/csi-spectrum-scale /csi-spectrum-scale
RUN chmod +x /csi-spectrum-scale
ENTRYPOINT ["/csi-spectrum-scale"]
