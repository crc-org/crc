FROM registry.redhat.io/ubi8/ubi
MAINTAINER CodeReady Container <devtools-cdk@redhat.com>

RUN yum -y install dnsmasq && \
    yum clean all

COPY dnsmasq.conf /etc/
COPY resolv.dnsmasq.conf /etc/

EXPOSE 53/udp

ENTRYPOINT ["/usr/sbin/dnsmasq", "--keep-in-foreground"]
