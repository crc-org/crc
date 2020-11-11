@basic
Feature: Basic test

    User explores some of the top-level CRC commands while going
    through the lifecycle of CRC.

    @darwin @linux @windows
    Scenario: CRC lifecycle

        Given user checks that CRC is not running already        
        And user sets up their environment 
        When starting CRC with default bundle succeeds
        Then user observes a running cluster
        And user checks cluster IP
        When user stops the cluster
        And user deletes the cluster
        And user cleans up the cluster
        Then stdout should contain "Cleanup finished"

