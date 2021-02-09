FROM registry.centos.org/centos/centos:centos8
MAINTAINER CodeReady Containers <devtools-cdk@redhat.com>

ENV SQUID_CACHE_DIR=/var/spool/squid \
    SQUID_LOG_DIR=/var/log/squid \
    SQUID_USER=squid

RUN yum -y install squid && \
    yum clean all

# Allow localnet to access proxy.
RUN sed -i "s/^#\+\(.*[acl|allow] localnet\)/\1/" /etc/squid/squid.conf

# Workaround for https://github.com/moby/moby/issues/31243
RUN usermod -a -G tty squid

COPY entrypoint.sh /sbin/entrypoint.sh
RUN chmod 755 /sbin/entrypoint.sh

EXPOSE 3128/tcp

ENTRYPOINT ["/sbin/entrypoint.sh"]
