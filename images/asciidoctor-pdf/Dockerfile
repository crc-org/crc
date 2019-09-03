FROM registry.redhat.io/ubi8/ruby-25
MAINTAINER CodeReady Containers <devtools-cdk@redhat.com>

RUN gem install asciidoctor-pdf --pre

VOLUME /docs
WORKDIR /docs

USER root

ENTRYPOINT [ "asciidoctor-pdf" ]
