@story_health
Feature: 
    End-to-end health check. Set-up and start CRC. Then create a
    project and deploy an app. Check on the app and delete the
    project. Stop and delete CRC.

    @linux @darwin
    Scenario: Start CRC
        Given executing "crc setup" succeeds
        When starting CRC with default bundle succeeds
        Then stdout should contain "Started the OpenShift cluster"
        When with up to "15" retries with wait period of "1m" command "crc status --log-level debug" output matches ".*Running \(v\d+\.\d+\.\d+.*\).*"
        And executing "eval $(crc oc-env)" succeeds
        Then login to the oc cluster succeeds

    @windows
    Scenario: Start CRC on Windows
        Given executing "crc setup" succeeds
        When starting CRC with default bundle and nameserver "10.75.5.25" succeeds
        Then stdout should contain "Started the OpenShift cluster"
        When with up to "15" retries with wait period of "1m" command "crc status --log-level debug" output matches ".*Running \(v\d+\.\d+\.\d+.*\).*"
        And executing "crc oc-env | Invoke-Expression" succeeds
        Then login to the oc cluster succeeds

    @linux @darwin @windows    
    Scenario: Check cluster health
        Given executing "crc status" succeeds
        And stdout should match ".*Running \(v\d+\.\d+\.\d+.*\).*"
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
        When executing "oc new-app centos/httpd-24-centos7~https://github.com/sclorg/httpd-ex" succeeds
        Then stdout should contain "Creating resources"
        And stdout should contain
            """
            service "httpd-ex" created
            """
        When executing "oc rollout status dc/httpd-ex" succeeds
        Then stdout should contain "successfully rolled out"
        And executing "oc expose svc/httpd-ex" succeeds
        And with up to "2" retries with wait period of "60s" http response from "http://httpd-ex-testproj.apps-crc.testing" has status code "200"

    @darwin @linux @windows
    Scenario: Stop and start CRC, then check app still runs
        Given with up to "2" retries with wait period of "60s" http response from "http://httpd-ex-testproj.apps-crc.testing" has status code "200"
        When executing "crc stop -f" succeeds
        Then with up to "4" retries with wait period of "2m" command "crc status" output should contain "Stopped"
        When starting CRC with default bundle succeeds
        Then with up to "15" retries with wait period of "1m" command "crc status" output should match ".*Running \(v\d+\.\d+\.\d+.*\).*"
        And with up to "2" retries with wait period of "60s" http response from "http://httpd-ex-testproj.apps-crc.testing" has status code "200"

    @darwin @linux @windows
    Scenario: Clean up
        Given executing "oc delete project testproj" succeeds
        When executing "crc stop -f" succeeds
        Then stdout should match "(.*)[Ss]topped the OpenShift cluster"
        When executing "crc delete -f" succeeds
        Then stdout should contain "Deleted the OpenShift cluster"
