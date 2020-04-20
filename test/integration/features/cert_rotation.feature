@cert_rotation @linux
Feature: Check the cert is rotation happen after it expire

    The user try to use crc after one month. he/she expect the
    cert rotation happen successfully and able to deploy the app and check its
    accessibility.

    Scenario: Set clock to 3 month ahead on the host
      Given executing "sudo timedatectl set-ntp off" succeeds
      Then executing "sudo date -s '3 month'" succeeds

    Scenario: Start CRC
      Given executing "crc setup" succeeds
      When starting CRC with default bundle along with stopped network time synchronization succeeds
      Then stdout should contain "Started the OpenShift cluster"
      And executing "eval $(crc oc-env)" succeeds
      When with up to "4" retries with wait period of "2m" command "crc status --log-level debug" output matches ".*Running \(v\d+\.\d+\.\d+.*\).*"
      Then login to the oc cluster succeeds

    Scenario: Set clock back to original time
      Given executing "sudo timedatectl set-ntp on" succeeds

    Scenario: CRC delete and cleanup
      When executing "crc delete -f" succeeds
      Then stdout should contain "Deleted the OpenShift cluster"
      When executing "crc cleanup" succeeds
