@cert_rotation @linux
Feature: Certificate rotation test

    User starts CRC more than 13 months after the release. They expect
    certificate rotation to happen successfully and to be able to deploy
    an app and check its accessibility.

    @timesync @cleanup
    Scenario: Start CRC "in the future" and clean up
        Given executing single crc setup command succeeds
        When starting CRC with default bundle along with stopped network time synchronization succeeds
        Then stdout should contain "Started the OpenShift cluster"
        And executing "eval $(crc oc-env)" succeeds
        When checking that CRC is running
        Then login to the oc cluster succeeds
        Then executing "oc whoami" succeeds
        And stdout should contain "kubeadmin"

