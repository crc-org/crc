FROM registry.centos.org/centos/centos:latest

RUN groupadd -r crcek -g 1234 && useradd -u 1234 -r -g crcek -m -d /crcek -s /sbin/nologin -c "CRC user" crcek && chmod 755 /crcek

WORKDIR /crcek

USER crcek

ENTRYPOINT ["/bin/bash", "-c", "echo Hello, it works! && echo Sleeping for 10min && sleep 10m && echo Done."]
