FROM registry.redhat.io/ubi8/python-36
MAINTAINER CodeReady Containers <devtools-cdk@redhat.com>

USER root

VOLUME /docs
WORKDIR /docs/build/

EXPOSE 8088

USER root

ENTRYPOINT [ "python" ]
CMD [ "-m", "http.server", "8088" ]
