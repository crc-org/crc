
FROM registry.access.redhat.com/ubi8/go-toolset:1.15.13 as builder

USER root
WORKDIR /workspace
COPY . .
RUN make build_integration

FROM registry.access.redhat.com/ubi8/ubi-minimal 

LABEL MAINTAINER "CodeReady Containers <devtools-cdk@redhat.com>"

COPY --from=builder /workspace/images/build-integration/entrypoint.sh /usr/local/bin/
COPY --from=builder /workspace/out /opt/crc/bin

ENV EPEL https://dl.fedoraproject.org/pub/epel/epel-release-latest-8.noarch.rpm

RUN rpm -ivh ${EPEL} \ 
    && microdnf --enablerepo=epel install -y openssh-clients sshpass \
    && microdnf clean all

ENTRYPOINT ["entrypoint.sh"]
