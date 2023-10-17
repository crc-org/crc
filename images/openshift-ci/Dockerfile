# This Dockerfile is used by openshift CI
# It builds an image containing crc and nss-wrapper for remote deployments, as well as the google cloud-sdk for nested GCE environments.

FROM registry.ci.openshift.org/openshift/release:rhel-8-release-golang-1.20-openshift-4.15 AS builder
WORKDIR /go/src/github.com/crc-org/crc
COPY . .
RUN make release

FROM registry.access.redhat.com/ubi9/ubi
COPY --from=builder /go/src/github.com/crc-org/crc /opt/crc
COPY --from=builder /go/src/github.com/crc-org/crc/out/linux-amd64/crc /bin/crc
COPY --from=builder /go/src/github.com/crc-org/crc/out/linux-amd64/e2e.test /bin/e2e.test
COPY --from=builder /go/src/github.com/crc-org/crc/out/linux-amd64/integration.test /bin/integration.test
COPY --from=builder /go/src/github.com/crc-org/crc/out/windows-amd64/crc.exe /opt/crc.exe
COPY --from=builder /go/src/github.com/crc-org/crc/out/windows-amd64/e2e.test.exe /opt/e2e.test.exe
COPY --from=builder /go/src/github.com/crc-org/crc/out/windows-amd64/integration.test.exe /opt/integration.test.exe
COPY --from=builder /go/src/github.com/crc-org/crc/images/openshift-ci/mock-nss.sh /bin/mock-nss.sh
COPY --from=builder /go/src/github.com/crc-org/crc/images/openshift-ci/google-cloud-sdk.repo /etc/yum.repos.d/google-cloud-sdk.repo
COPY --from=builder /go/src/github.com/crc-org/crc/images/openshift-ci/azure-cli.repo /etc/yum.repos.d/azure-cli.repo

RUN yum update -y && \
    yum install --setopt=tsflags=nodocs -y \
    google-cloud-sdk \
    nss_wrapper \
    unzip \
    sshpass \
    jq \
    azure-cli \
    openssh-clients && \
    yum clean all && rm -rf /var/cache/yum/*
RUN mkdir /output && chown 1000:1000 /output
USER 1000:1000
ENV PATH /bin
ENV HOME /output
WORKDIR /output
