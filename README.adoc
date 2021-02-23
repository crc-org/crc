= Red Hat CodeReady Containers - OpenShift 4 on your Laptop
:icons:
:toc: macro
:toc-title:
:toclevels:

toc::[]

image:https://travis-ci.org/code-ready/crc.svg?branch=master["Travis CI Build Status", link="https://travis-ci.org/code-ready/crc"]
image:https://circleci.com/gh/code-ready/crc/tree/master.svg?style=svg["CircleCI Build Status", link="https://circleci.com/gh/code-ready/crc"]
image:https://ci.centos.org/buildStatus/icon?job=codeready-crc-master["CentOS CI Build Status", link="https://ci.centos.org/job/codeready-crc-master"]

[[intro-to-crc]]
== Introduction

This project is focused on bringing a minimal http://github.com/openshift/origin[OpenShift 4.x] cluster to your local laptop or desktop computer.

If you are looking for a solution for running OpenShift 3.x, you will need tools such as `oc cluster up`, http://github.com/minishift/minishift[Minishift] or https://developers.redhat.com/products/cdk/overview/[CDK].

[[usage-data]]
== Usage data

The first time CodeReady Containers is run, you will be asked to opt-in to Red Hat's telemetry collection program.

With your approval, CodeReady Containers collects pseudonymized usage data and sends it to Red Hat servers to help improve our products and services. Read our https://developers.redhat.com/article/tool-data-collection[privacy statement] to learn more about it. For the specific data points being collected, see xref:usage-data.adoc#data-table[Usage data].

=== Manually configuring usage data collection

You can manually change your preference about usage data collection by running `crc config set consent-telemetry <yes/no>` before the next `crc start`. 


[[documentation]]
== Documentation

=== Getting CodeReady Containers

CodeReady Containers binaries with an embedded OpenShift disk image can be downloaded from link:https://cloud.redhat.com/openshift/create/local[this page].

=== Using CodeReady Containers

The documentation for CodeReady Containers is currently hosted by GitHub Pages.

See the link:https://code-ready.github.io/crc/[Red Hat CodeReady Containers Getting Started Guide].

=== Building the documentation

You can find the source files for the documentation in the link:./docs/source[docs/source] directory.

To build the formatted documentation, link:https://github.com/containers/libpod/blob/master/install.md[install podman] then use the following:

```bash
$ git clone https://github.com/code-ready/crc
$ cd crc
$ make build_docs
```

This will create a [filename]`docs/build/master.html` file which you can view in your browser.

=== Developing CodeReady Containers

Developers who want to work on CodeReady Containers should visit the link:./developing.adoc[Developing CodeReady Containers] document.

[[community]]
== Community

Contributions, questions, and comments are all welcomed and encouraged!

You can reach the community by:

- Joining the #codeready channel on https://freenode.net/[Freenode IRC]

If you want to contribute, make sure to follow the link:CONTRIBUTING.adoc[contribution guidelines]
when you open issues or submit pull requests.
