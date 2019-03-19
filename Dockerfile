FROM centos:7
LABEL maintainers="FSaaS Authors"
LABEL description="CSI Plugin for Scale"

RUN yum install -y openssh-clients openssh-server
RUN sshd-keygen; ssh-keygen -t rsa -N "" -f  /root/.ssh/id_rsa; cat ~/.ssh/id_rsa.pub > ~/.ssh/authorized_keys
RUN echo StrictHostKeyChecking no >> ~/.ssh/config

COPY _output/csi-scale /csi-scale
RUN chmod +x /csi-scale
ENTRYPOINT ["/csi-scale"]
