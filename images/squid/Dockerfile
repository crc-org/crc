# First run 'generate-certs.sh' so that the TLS certificate are generated
# Then 'podman build .' can be run
# The resulting container image needs to be run with -p 3128:3128 -p 3129:3129
# Once the container is running, the https proxy can be tested with
# https_proxy=https://localhost:3129 curl --proxy-cacert ./pki/rootCA.crt -L https://gandi.net
#
# The generated TLS certificate is currently only valid for 192.168.122.1, you'll need to modify
# generate-certs.sh if you want to use it on a different host
#
# https Traffic in the VM can be blocked by running this in a VM:
# sudo firewall-cmd  --direct --add-rule ipv4 filter OUTPUT 0 -p tcp -m tcp --dport=443 -j REJECT
# sudo firewall-cmd  --direct --add-rule ipv4 filter OUTPUT 1 -j ACCEPT
#
# To allow traffic on port 3128/3129 from the VM to the host, run this on the host:
# sudo firewall-cmd --zone=libvirt --add-port=3128/tcp
# sudo firewall-cmd --zone=libvirt --add-port=3129/tcp
#
# After this, running this command in a VM hangs:
# curl -L -I https://gandi.net
# and this command succeds:
# https_proxy=https://192.168.122.1:3129 curl  --proxy-cacert ./rootCA.crt  -L  https://gandi.net
# (commonName must be set to 192.168.122.1 before running generate-certs.sh and building the container image)

FROM registry.centos.org/centos/centos:centos8
MAINTAINER CodeReady Containers <devtools-cdk@redhat.com>

RUN yum -y install squid && \
    yum clean all
RUN systemctl enable squid.service

# Workaround for https://github.com/moby/moby/issues/31243
RUN usermod -a -G tty squid

# Allow localnet to access proxy and enable access to squid over https
RUN sed -i "s/^#\+\(.*[acl|allow] localnet\)/\1/" /etc/squid/squid.conf && \
    sed -i "s!http_port 3128!http_port 3128\nhttps_port 3129 tls-cert=/etc/pki/squid/squid.pem!" /etc/squid/squid.conf

COPY pki/squid.pem /etc/pki/squid/squid.pem

EXPOSE 3128/tcp
EXPOSE 3129/tcp

CMD [ "/sbin/init" ]
