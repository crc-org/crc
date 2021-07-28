
FROM registry.access.redhat.com/ubi8/go-toolset:1.15.13 as builder

USER root
WORKDIR /workspace
COPY . .
RUN make build_e2e

FROM registry.access.redhat.com/ubi8/ubi-minimal 

LABEL MAINTAINER "CodeReady Containers <devtools-cdk@redhat.com>"

COPY --from=builder /workspace/images/build-e2e/entrypoint.sh /usr/local/bin/
COPY --from=builder /workspace/out /opt/crc/bin
# Review this when go 1.16 with embed support
COPY --from=builder /workspace/test/e2e/features /opt/crc/features
COPY --from=builder /workspace/test/testdata /opt/crc/testdata
COPY --from=builder /workspace/test/extended/crc/ux/installer/applescripts /opt/crc/ux/installer/applescripts
COPY --from=builder /workspace/test/extended/crc/ux/notification/applescripts /opt/crc/ux/notification/applescripts
COPY --from=builder /workspace/test/extended/crc/ux/tray/applescripts /opt/crc/ux/tray/applescripts


ENV EPEL https://dl.fedoraproject.org/pub/epel/epel-release-latest-8.noarch.rpm

RUN rpm -ivh ${EPEL} \ 
    && microdnf --enablerepo=epel install -y openssh-clients sshpass \
    && microdnf clean all

ENTRYPOINT ["entrypoint.sh"]
