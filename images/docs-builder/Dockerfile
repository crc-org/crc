FROM registry.access.redhat.com/ubi8/ubi
MAINTAINER CodeReady Containers <devtools-cdk@redhat.com>

RUN yum update -y && \
    yum install --setopt=tsflags=nodocs -y \
    ruby rubygems python38 && \
    yum module install nodejs:14 -y && \
    yum clean all && rm -rf /var/cache/yum/*

RUN gem install asciidoctor asciidoctor-pdf --pre && \
    npm install -g asciidoc-link-check

COPY entrypoint.sh /sbin/entrypoint.sh
RUN chmod 755 /sbin/entrypoint.sh
COPY links_ignorelist.json /links_ignorelist.json

VOLUME /docs
WORKDIR /docs

USER root

ENTRYPOINT [ "entrypoint.sh" ]
