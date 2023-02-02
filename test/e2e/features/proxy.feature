@proxy @linux
Feature: Behind proxy test

    Check CRC use behind proxy

    # inherits @proxy tag from Feature
    @cleanup
    Scenario: Start CRC behind proxy under openshift preset
        Given executing single crc setup command succeeds
        When starting CRC with default bundle succeeds
        Then stdout should contain "Started the OpenShift cluster"
        And executing "eval $(crc oc-env)" succeeds
        When checking that CRC is running
        Then login to the oc cluster succeeds

    # inherits @proxy tag from Feature
    @cleanup @podman-preset
    Scenario: Cache podman bundle behind proxy under podman preset
        * removing podman bundle from cache succeeds
        * executing "crc config set preset podman" succeeds
        When executing single crc setup command succeeds
        Then executing "crc start" succeeds
