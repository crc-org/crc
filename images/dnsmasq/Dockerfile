FROM centos:8
MAINTAINER CodeReady Container <devtools-cdk@redhat.com>

RUN yum -y install dnsmasq && \
    yum clean all && \
    systemctl enable dnsmasq

COPY dnsmasq.conf /etc/
COPY resolv.dnsmasq.conf /etc/

EXPOSE 53/udp

CMD [ "/sbin/init" ]
