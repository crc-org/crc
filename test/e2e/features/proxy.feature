@proxy @linux
Feature: Behind proxy test

    Check CRC use behind proxy

    Background: Setup the proxy container using podman
        * executing "podman run --name squid -d -p 3128:3128 quay.io/crcont/squid" succeeds

    @cleanup
    Scenario: Start CRC behind proxy under openshift preset
        Given executing single crc setup command succeeds
        And  executing "crc config set http-proxy http://192.168.130.1:3128" succeeds
        Then executing "crc config set https-proxy http://192.168.130.1:3128" succeeds
        When starting CRC with default bundle succeeds
        Then stdout should contain "Started the OpenShift cluster"
        And executing "eval $(crc oc-env)" succeeds
        When checking that CRC is running
        Then login to the oc cluster succeeds
        # Remove the proxy container and host proxy env (which set because of oc-env)
        Given executing "podman stop squid" succeeds
        And executing "podman rm squid" succeeds
        And executing "unset HTTP_PROXY HTTPS_PROXY NO_PROXY" succeeds
        # CRC delete and remove proxy settings from config
        When executing "crc delete -f" succeeds
        Then stdout should contain "Deleted the instance"
        And  executing "crc config unset http-proxy" succeeds
        And executing "crc config unset https-proxy" succeeds
        And executing crc cleanup command succeeds


    Scenario: Cache podman bundle behind proxy under podman preset
        * executing "crc config set http-proxy http://192.168.130.1:3128" succeeds
        * executing "crc config set https-proxy http://192.168.130.1:3128" succeeds
        * removing podman bundle from cache succeeds
<<<<<<< HEAD
        Given executing "crc config set preset podman" succeeds
        Then executing single crc setup command succeeds
        And podman bundle is cached
        # cleanup proxy and preset settings from config
        * executing "crc config unset http-proxy" succeeds
        * executing "crc config unset https-proxy" succeeds
        * executing "crc config unset preset" succeeds
        * executing "podman stop squid" succeeds
        * executing "podman rm squid" succeeds
=======
        * executing "crc config set preset podman" succeeds
        When executing single crc setup command succeeds
        Then executing "crc start" succeeds
>>>>>>> 5c6d0b94 (Trying a pure start for podman preset)
