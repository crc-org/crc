# This Dockerfile is used by openshift CI
# It builds an image containing crc and nss-wrapper for remote deployments, as well as the google cloud-sdk for nested GCE environments.

FROM registry.svc.ci.openshift.org/openshift/release:golang-1.15 AS builder
WORKDIR /go/src/github.com/code-ready/crc
COPY . .
RUN make release

FROM registry.centos.org/centos/centos:7
COPY --from=builder /go/src/github.com/code-ready/crc /opt/crc
COPY --from=builder /go/src/github.com/code-ready/crc/out/linux-amd64/crc /bin/crc
COPY --from=builder /go/src/github.com/code-ready/crc/images/openshift-ci/mock-nss.sh /bin/mock-nss.sh
COPY --from=builder /go/src/github.com/code-ready/crc/images/openshift-ci/google-cloud-sdk.repo /etc/yum.repos.d/google-cloud-sdk.repo

RUN yum update -y && \
    yum install --setopt=tsflags=nodocs -y \
    genisoimage \
    gettext \
    google-cloud-sdk \
    libvirt-client \
    libvirt-libs \
    nss_wrapper \
    unzip \
    openssh-clients && \
    yum clean all && rm -rf /var/cache/yum/*
RUN mkdir /output && chown 1000:1000 /output
USER 1000:1000
ENV PATH /bin
ENV HOME /output
WORKDIR /output
