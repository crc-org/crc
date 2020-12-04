@story_health
Feature: End-to-end health check

    User goes through a "day in the life of" CRC. They set up
    and start CRC. They then create a project and deploy an app.
    They check on the app and delete the project. They stop CRC
    and delete the CRC VM.

    @linux @darwin
    Scenario: Start CRC
        Given executing "crc setup" succeeds
        When setting config property "memory" to value "12000" succeeds
        When starting CRC with default bundle succeeds
        Then stdout should contain "Started the OpenShift cluster"
        And executing "eval $(crc oc-env)" succeeds
        When checking that CRC is running
        Then login to the oc cluster succeeds

    @windows
    Scenario: Start CRC on Windows
        Given executing "crc setup" succeeds
        When setting config property "memory" to value "12000" succeeds
        When starting CRC with default bundle and nameserver "10.75.5.25" succeeds
        Then stdout should contain "Started the OpenShift cluster"
        And executing "crc oc-env | Invoke-Expression" succeeds
        When checking that CRC is running
        Then login to the oc cluster succeeds

    @linux @darwin @windows    
    Scenario: Check cluster health
        When executing "oc get nodes"
        Then stdout contains "Ready" 
        And stdout does not contain "Not ready"
        # next line checks similar things as `crc status` except gives more informative output
        And with up to "5" retries with wait period of "1m" all cluster operators are running

    @darwin @linux @windows
    Scenario: Create project
        When executing "oc new-project testproj" succeeds
        Then stdout should contain
            """
            Now using project "testproj" on server "https://api.crc.testing:6443".
            """
        And stdout should contain "You can add applications to this project with the 'new-app' command."

    @darwin @linux @windows
    Scenario: Create and test app
        When executing "oc new-app httpd-example" succeeds
        Then stdout should contain "Creating resources"
        And stdout should contain
            """
            service "httpd-example" created
            """
        When executing "oc rollout status dc/httpd-example || oc rollout status deployment httpd-example" succeeds
        Then stdout should contain "successfully rolled out"

    @darwin @linux @windows
    Scenario: Stop and start CRC, then check app still runs
        Given with up to "2" retries with wait period of "60s" http response from "http://httpd-example-testproj.apps-crc.testing" has status code "200"
        When executing "crc stop -f" succeeds
        Then checking that CRC is stopped
        When starting CRC with default bundle succeeds
        Then checking that CRC is running
        And with up to "2" retries with wait period of "1m" http response from "http://httpd-example-testproj.apps-crc.testing" has status code "200"


    @darwin @linux @windows
    Scenario: Switch off CRC
        When executing "oc delete project testproj" succeeds
        Then executing "crc stop -f" succeeds
        And executing "crc delete -f" succeeds
        And executing "crc cleanup" succeeds
