Feature: Basic story - day 1
Nikola tries Code Ready Containers for the first time.
They check the CRC version.

    Scenario: Nikola checks CRC version
    When executing "crc version" succeeds
    Then stdout should contain "crc - Local OpenShift 4.0 clusters"
