@cert_rotation @linux
Feature: Certificate rotation test

    User starts CRC more than one month after the release. They expect
    certificate rotation to happen successfully and to be able to deploy
    an app and check its accessibility.

    Scenario: Set clock to 3 months ahead on the host
        Given executing "sudo timedatectl set-ntp off" succeeds
        Then executing "sudo date -s '3 month'" succeeds

    Scenario: Start CRC
        Given executing "crc setup" succeeds
        When starting CRC with default bundle along with stopped network time synchronization succeeds
        Then stdout should contain "Started the OpenShift cluster"
        And executing "eval $(crc oc-env)" succeeds
        When checking that CRC is running
        Then login to the oc cluster succeeds

    Scenario: Set clock back to the original time
        When executing "sudo date -s '-3 month'" succeeds
        And executing "sudo timedatectl set-ntp on" succeeds

    Scenario: CRC delete and cleanup
        When executing "crc delete -f" succeeds
        Then stdout should contain "Deleted the OpenShift cluster"
        When executing "crc cleanup" succeeds
