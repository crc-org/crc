Feature: Basic story - day 1
Nikola tries Code Ready Containers for the first time.
They start the CRC, deploy sample application and then they delete the CRC.

    Scenario: Nikola starts VM with CRC
    When executing "crc start" succeeds
    Then CRC should be in state "Running"

    Scenario: Nikola checks the OpenShift webconsole

    Scenario: Nikola sets the `oc` for the recently started OpenShift instance
 
    Scenario: Nikola logs into the cluster

    Scenario: Nikola creates new project

    Scenario: Nikola deploys sample application

    Scenario: Nikola checks application's route

    Scenario: Nikola deletes VM with CRC
    When executing "crc delete" succeeds
    Then CRC should be in state "Does Not Exist"
