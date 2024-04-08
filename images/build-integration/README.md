# Overview

This image contains the integration suite of tests, and it is intended to run them against a target host, the logic to run the tests remotely is inherit from its base image: [deliverest](https://github.com/adrianriobo/deliverest), each platform and os has its own image.

## Build

We should build an image per platform and arch we want to test:

```bash
CRC_INTEGRATION_IMG_VERSION=v2.35.0-dev OS=windows ARCH=amd64 make containerized_integration
CRC_INTEGRATION_IMG_VERSION=v2.35.0-dev OS=linux ARCH=amd64 make containerized_integration
CRC_INTEGRATION_IMG_VERSION=v2.35.0-dev OS=darwin ARCH=amd64 make containerized_integration
CRC_INTEGRATION_IMG_VERSION=v2.35.0-dev OS=darwin ARCH=arm64 make containerized_integration
```

## Run

The image version contains the specs about the plaftorm and the arch; then the test customization is made by the command executed within the image; as we can see the cmd should be defined depending on the platform:

* crc-integration/run.ps1 ... (windows)
* crc-integration/run.sh ... (linux/darwin)

And the execution is customized by the params addded, available params:

* bundleLocation When testing a custom bundle we should pass the path on the target host
* targetFolder Name of the folder on the target host under $HOME where all the content will be copied
* junitFilename Name for the junit file with the tests results

### Windows amd64

```bash
podman run --rm -d --name crc-integration-windows \
    -e TARGET_HOST=XXXX \
    -e TARGET_HOST_USERNAME=XXXX \
    -e TARGET_HOST_KEY_PATH=/data/id_rsa \
    -e TARGET_FOLDER=crc-integration \
    -e TARGET_RESULTS=results \
    -e OUTPUT_FOLDER=/data \
    -e DEBUG=true \
    -v $PWD/pull-secret:/opt/crc/pull-secret:z \
    -v $PWD:/data:z \
    quay.io/crcont/crc-integration:v2.34.0-windows-amd64  \
        crc-integration/run.ps1 -junitFilename crc-integration-junit.xml
```

### Mac arm64

```bash
podman run --rm -d --name crc-integration-darwin \
    -e TARGET_HOST=XXXX \
    -e TARGET_HOST_USERNAME=XXXX \
    -e TARGET_HOST_KEY_PATH=/data/id_rsa \
    -e TARGET_FOLDER=crc-integration \
    -e TARGET_RESULTS=results \
    -e OUTPUT_FOLDER=/data \
    -e DEBUG=true \
    -v $PWD/pull-secret:/opt/crc/pull-secret:z \
    -v $PWD:/data:z \
    quay.io/crcont/crc-integration:v2.34.0-darwin-arm64  \
        crc-integration/run.sh -junitFilename crc-integration-junit.xml 
```

### Mac amd64

```bash
podman run --rm -d --name crc-integration-darwin \
    -e TARGET_HOST=XXXX \
    -e TARGET_HOST_USERNAME=XXXX \
    -e TARGET_HOST_KEY_PATH=/data/id_rsa \
    -e TARGET_FOLDER=crc-integration \
    -e TARGET_RESULTS=results \
    -e OUTPUT_FOLDER=/data \
    -e DEBUG=true \
    -v $PWD/pull-secret:/opt/crc/pull-secret:z \
    -v $PWD:/data:z \
    quay.io/crcont/crc-integration:v2.34.0-darwin-amd64  \
        crc-integration/run.sh -junitFilename crc-integration-junit.xml 
```


### Linux amd64

```bash
podman run --rm -d --name crc-integration-linux \
    -e TARGET_HOST=XXXX \
    -e TARGET_HOST_USERNAME=XXXX \
    -e TARGET_HOST_KEY_PATH=/data/id_rsa \
    -e TARGET_FOLDER=crc-integration \
    -e TARGET_RESULTS=results \
    -e OUTPUT_FOLDER=/data \
    -e DEBUG=true \
    -v $PWD/pull-secret:/opt/crc/pull-secret:z \
    -v $PWD:/data:z \
    quay.io/crcont/crc-integration:v2.34.0-linux-amd64  \
        crc-integration/run.sh -junitFilename crc-integration-junit.xml 
```