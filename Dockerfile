FROM centos:7
LABEL maintainers="FSaaS Authors"
LABEL description="CSI Plugin for Scale"

COPY _output/csi-scale /csi-scale
RUN chmod +x /csi-scale
ENTRYPOINT ["/csi-scale"]
