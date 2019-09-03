Feature: 
    End-to-end health check. Set-up and start CRC. Then create a
    project and deploy an app. Check on the app and delete the
    project. Stop and delete CRC.

    Scenario Outline: Start CRC
        Given executing "crc setup" succeeds
        When starting CRC with default bundle and hypervisor "<vm-driver>" succeeds
        Then stdout should contain "CodeReady Containers instance is running"
        And executing "eval $(crc oc-env)" succeeds
        When with up to "4" retries with wait period of "2m" command "crc status" output should contain "Running (v4."
        Then login to the oc cluster succeeds

    @darwin
        Examples:
            | vm-driver  |
            | hyperkit   |

    @linux
        Examples:
            | vm-driver |
            | libvirt   |

    @darwin @linux 
    Scenario: Check cluster health
        Given executing "crc status" succeeds
        And stdout contains "Running (v4."
        When executing "oc get nodes"
        Then stdout contains "Ready" 
        And stdout does not contain "Not ready"
        # next line checks similar things as `crc status` except gives more informative output
        And with up to "5" retries with wait period of "1m" all cluster operators are running

    @darwin @linux
    Scenario: Create project
        When executing "oc new-project testproj" succeeds
        Then stdout should contain
            """
            Now using project "testproj" on server "https://api.crc.testing:6443".
            """
        And stdout should contain "You can add applications to this project with the 'new-app' command."

    @darwin @linux
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
        And with up to "2" retries with wait period of "60s" http response from "http://httpd-ex-testproj.apps-crc.testing" should have status code "200"

    Scenario: Clean up
        Given executing "oc delete project testproj" succeeds
        When executing "crc stop -f" succeeds
        Then stdout should contain "CodeReady Containers instance stopped"
        When executing "crc delete" succeeds
        Then stdout should contain "CodeReady Containers instance deleted"
