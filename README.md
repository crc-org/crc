CRC - Runs Containers
=====================

- [Introduction](https://github.com/crc-org/crc#intro-to-crc)
- [Usage data](https://github.com/crc-org/crc#usage-data)
- [Documentation](https://github.com/crc-org/crc#documentation)
- [Community](https://github.com/crc-org/crc#community)

[![main](https://github.com/crc-org/crc/actions/workflows/make-check.yml/badge.svg?branch=main)](https://github.com/crc-org/crc/actions/workflows/make-check.yml) [![macos pkg](https://github.com/crc-org/crc/actions/workflows/macos-installer.yml/badge.svg)](https://github.com/crc-org/crc/actions/workflows/macos-installer.yml) [![rpm](https://github.com/crc-org/crc/actions/workflows/make-rpm.yml/badge.svg)](https://github.com/crc-org/crc/actions/workflows/make-rpm.yml) [![win](https://github.com/crc-org/crc/actions/workflows/make-check-win.yml/badge.svg)](https://github.com/crc-org/crc/actions/workflows/make-check-win.yml)


## Introduction
`crc` is a tool to run containers. It manages a local [OpenShift 4.x](https://github.com/openshift/origin) cluster, or an [OKD](https://github.com/openshift/okd) cluster VM optimized for testing and development purposes.

If you are looking for a solution for running OpenShift 3.x, you will need tools such as `oc cluster up`, [Minishift](http://github.com/minishift/minishift) or [CDK](https://developers.redhat.com/products/cdk/overview/).


## Usage data
The first time CRC is run, you will be asked to opt-in to Red Hat’s telemetry collection program.

With your approval, CRC collects pseudonymized usage data and sends it to Red Hat servers to help improve our products and services. Read our [privacy statement](https://developers.redhat.com/article/tool-data-collection) to learn more about it. For the specific data points being collected, see [Usage data](https://github.com/crc-org/crc/blob/main/usage-data.adoc#data-table).


### Manually configuring usage data collection
You can manually change your preference about usage data collection by running `crc config set consent-telemetry <yes/no>` before the next `crc start`.


## Documentation

### Getting CRC
CRC binaries with an embedded OpenShift disk image can be downloaded from [this page](https://console.redhat.com/openshift/create/local).


### Using CRC
The documentation for CRC is currently hosted by GitHub Pages.

See the [CRC Getting Started Guide](https://crc-org.github.io/crc/).

### Building the documentation
You can find the source files for the documentation in the [docs](https://github.com/crc-org/crc/blob/main/docs) directory.

To build the formatted documentation, [install podman](https://github.com/containers/libpod/blob/master/install.md) then use the following:

```shell
$ git clone https://github.com/crc-org/crc
$ cd crc
$ make build_docs
```

This will create a `docs/build/master.html` file which you can view in your browser.

### Developing CRC
Developers who want to work on CRC should visit the [Developing CRC](https://github.com/crc-org/crc/blob/main/developing.adoc) document.

## Community
Contributions, questions, and comments are all welcomed and encouraged!

You can reach the community by:

- Joining the #codeready channel on [Freenode IRC](https://freenode.net/)
    

If you want to contribute, make sure to follow the [contribution guidelines](https://github.com/crc-org/crc/blob/main/CONTRIBUTING.md) when you open issues or submit pull requests.