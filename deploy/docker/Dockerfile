FROM centos:7
LABEL maintainers="FSaaS Authors"
LABEL description="GPFS CSI Plugin"

#ENV GPFS_VERSION "bla"
#RUN yum  install -y centos-release-ceph && \
#    yum install -y ceph-common e2fsprogs && \ 
#    yum clean all

COPY gpfsplugin /gpfsplugin
COPY fake_client /fake_client
RUN chmod +x /gpfsplugin
ENTRYPOINT ["/gpfsplugin"]
