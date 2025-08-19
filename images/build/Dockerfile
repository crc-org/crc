FROM registry.ci.openshift.org/openshift/release:rhel-9-release-golang-1.23-openshift-4.19 AS builder
MAINTAINER CRC <devtools-cdk@redhat.com>

WORKDIR /opt/src
COPY . .
RUN make cross

FROM registry.access.redhat.com/ubi9/ubi
COPY --from=builder /opt/src/out/ /opt/
