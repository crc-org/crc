@cert_rotation @linux
Feature: Certificate rotation test

    User starts CRC more than one month after the release. They expect
    certificate rotation to happen successfully and to be able to deploy
    an app and check its accessibility.

    Background: Setup CRC and rewind the clock forward
        When executing crc setup command succeeds
        And executing "sudo timedatectl set-ntp off" succeeds
        Then executing "sudo date -s '13 month'" succeeds
        And with up to "10" retries with wait period of "1s" command "virsh --readonly -c qemu:///system capabilities" output matches "^<capabilities>"

    Scenario: Start CRC "in the future" and clean up
        When starting CRC with default bundle along with stopped network time synchronization succeeds
        Then stdout should contain "Started the OpenShift cluster"
        And executing "eval $(crc oc-env)" succeeds
        When checking that CRC is running
        Then login to the oc cluster succeeds
        Then executing "oc whoami" succeeds
        And stdout should contain "kubeadmin"
        # Set clock back to the original time
        When executing "sudo date -s '-13 month'" succeeds
        And executing "sudo timedatectl set-ntp on" succeeds
        # CRC delete and cleanup
        When executing "crc delete -f" succeeds
        Then stdout should contain "Deleted the instance"
        When executing crc cleanup command succeeds
