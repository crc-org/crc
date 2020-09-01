FROM registry.redhat.io/ubi8
MAINTAINER CodeReady Containers <devtools-cdk@redhat.com>

RUN yum update -y && \
    yum install --setopt=tsflags=nodocs -y \
    ruby rubygems python38 && \
    yum clean all && rm -rf /var/cache/yum/*

RUN gem install asciidoctor asciidoctor-pdf --pre

COPY entrypoint.sh /sbin/entrypoint.sh
RUN chmod 755 /sbin/entrypoint.sh

VOLUME /docs
WORKDIR /docs

USER root

ENTRYPOINT [ "entrypoint.sh" ]
