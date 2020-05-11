@monitoring @linux
Feature: 

    Run for 24h and record vital data about memory, cpu, and other
    resources.

    Scenario: Start CRC
        Given executing "crc setup" succeeds
        When starting CRC with default bundle succeeds
        And executing "eval $(crc oc-env)" succeeds
        And with up to "4" retries with wait period of "2m" command "crc status --log-level debug" output matches ".*Running \(v\d+\.\d+\.\d+.*\).*"
        Then login to the oc cluster succeeds

    Scenario: Monitoring
        Given preparing and recording the environment succeeds
        When taking snapshot of the node every "60m" exactly "22" times succeeds
        Then packaging and uploading data succeeds

    Scenario: Closing shop
        When executing "crc stop -f" succeeds
        And executing "crc delete -f" succeeds
        And executing "crc cleanup" succeeds
