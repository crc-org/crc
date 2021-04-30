FROM registry.access.redhat.com/ubi8/ubi-minimal

WORKDIR /tmp

RUN echo "hello" > index.html
RUN microdnf install python3

ENTRYPOINT ["python3", "-m", "http.server"]

CMD ["8080"]