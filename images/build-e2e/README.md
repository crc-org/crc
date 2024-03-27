# Overview

This image contains the e2e suite of tests, and it is intended to run them agaisnt a target host, the logic to run the tests remotely is inherit from its base image: [deliverest](https://github.com/adrianriobo/deliverest), each platform and os has its own image.

## Run

The image version contains the specs about the plaftorm and the arch; then the test customization is made by the command executed within the image; as we can see the cmd should be defined depending on the platform:

* crc-e2e/run.ps1 ... (windows)
* crc-e2e/run.sh ... (linux/darwin)

And the execution is customized by the params addded, available params:

* bundleLocation When testing a custom bundle we should pass the paht on the target host
* e2eTagExpression To set an specific set of tests based on annotations 
* targetFolder Name of the folder on the target host under $HOME where all the content will be copied
* junitFilename Name for the junit file with the tests results
* crcMemory Customize memory for the cluster to run the tests

### Windows amd64

```bash
podman run --rm -d --name crc-e2e-windows \
    -e TARGET_HOST=XXXX \
    -e TARGET_HOST_USERNAME=XXXX \
    -e TARGET_HOST_KEY_PATH=/data/id_rsa \
    -e TARGET_FOLDER=crc-e2e \
    -e TARGET_RESULTS=results \
    -e OUTPUT_FOLDER=/data \
    -e DEBUG=true \
    -v $PWD/pull-secret:/opt/crc/pull-secret:z \
    -v $PWD:/data:z \
    quay.io/crcont/crc-e2e:v2.34.0-windows-amd64  \
        crc-e2e/run.ps1 -junitFilename crc-e2e-junit.xml
```

### Mac arm64

```bash
podman run --rm -d --name crc-e2e-darwin \
    -e TARGET_HOST=XXXX \
    -e TARGET_HOST_USERNAME=XXXX \
    -e TARGET_HOST_KEY_PATH=/data/id_rsa \
    -e TARGET_FOLDER=crc-e2e \
    -e TARGET_RESULTS=results \
    -e OUTPUT_FOLDER=/data \
    -e DEBUG=true \
    -v $PWD/pull-secret:/opt/crc/pull-secret:z \
    -v $PWD:/data:z \
    quay.io/crcont/crc-e2e:v2.34.0-darwin-arm64  \
        crc-e2e/run.sh --junitFilename crc-e2e-junit.xml 
```

### Mac amd64

```bash
podman run --rm -d --name crc-e2e-darwin \
    -e TARGET_HOST=XXXX \
    -e TARGET_HOST_USERNAME=XXXX \
    -e TARGET_HOST_KEY_PATH=/data/id_rsa \
    -e TARGET_FOLDER=crc-e2e \
    -e TARGET_RESULTS=results \
    -e OUTPUT_FOLDER=/data \
    -e DEBUG=true \
    -v $PWD/pull-secret:/opt/crc/pull-secret:z \
    -v $PWD:/data:z \
    quay.io/crcont/crc-e2e:v2.34.0-darwin-amd64  \
        crc-e2e/run.sh --junitFilename crc-e2e-junit.xml 
```


### Linux amd64

```bash
podman run --rm -d --name crc-e2e-linux \
    -e TARGET_HOST=XXXX \
    -e TARGET_HOST_USERNAME=XXXX \
    -e TARGET_HOST_KEY_PATH=/data/id_rsa \
    -e TARGET_FOLDER=crc-e2e \
    -e TARGET_RESULTS=results \
    -e OUTPUT_FOLDER=/data \
    -e DEBUG=true \
    -v $PWD/pull-secret:/opt/crc/pull-secret:z \
    -v $PWD:/data:z \
    quay.io/crcont/crc-e2e:v2.34.0-linux-amd64  \
        crc-e2e/run.sh --junitFilename crc-e2e-junit.xml 
```