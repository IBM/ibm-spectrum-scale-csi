FROM centos:7
LABEL maintainers="FSaaS Authors"
LABEL description="CSI Plugin for GPFS"

COPY csi-gpfs /csi-gpfs
RUN chmod +x /csi-gpfs
ENTRYPOINT ["/csi-gpfs"]
