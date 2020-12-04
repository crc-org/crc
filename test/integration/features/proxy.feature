@proxy @linux
Feature: Behind proxy test

    User starts CRC behind a proxy. They expect a successful start 
    and to be able to deploy an app and check its accessibility.

    Scenario: Setup the proxy container using podman
        Given executing "sudo podman run --name squid -d -p 3128:3128 quay.io/crcont/squid" succeeds

    Scenario: Start CRC
        Given executing "crc setup" succeeds
        And  executing "crc config set http-proxy http://192.168.130.1:3128" succeeds
        Then executing "crc config set https-proxy http://192.168.130.1:3128" succeeds
        When starting CRC with default bundle succeeds
        Then stdout should contain "Started the OpenShift cluster"
        And executing "eval $(crc oc-env)" succeeds
        When checking that CRC is running
        Then login to the oc cluster succeeds

    Scenario: Remove the proxy container and host proxy env (which set because of oc-env)
        Given executing "sudo podman stop squid" succeeds
        And executing "sudo podman rm squid" succeeds
        And executing "unset HTTP_PROXY HTTPS_PROXY NO_PROXY" succeeds

    Scenario: CRC delete and remove proxy settings from config
        When executing "crc delete -f" succeeds
        Then stdout should contain "Deleted the OpenShift cluster"
        And  executing "crc config unset http-proxy" succeeds
        And executing "crc config unset https-proxy" succeeds
        And executing "crc cleanup" succeeds

