Feature: 
    End-to-end health check. Set-up and start CRC. Then create a
    project and deploy an app. Check on the app and delete the
    project. Stop and delete CRC.

    Scenario Outline: Start CRC
        Given executing "crc setup" succeeds
        When starting CRC with default bundle and hypervisor "<vm-driver>" succeeds
        Then stdout should contain "CodeReady Containers instance is running"
        And executing "eval $(crc oc-env)" succeeds
        And executing "oc login --insecure-skip-tls-verify -u kubeadmin -p ehbg7-zu5i6-JKt7V-PvJsm https://api.crc.testing:6443" succeeds
        
    @darwin
        Examples:
            | vm-driver  |
            | virtualbox |
            |            |
    @linux
        Examples:
            | vm-driver |
            | libvirt   |

    @darwin @linux
    Scenario: Create new project
        When executing "oc new-project testproj" succeeds
        Then stdout should contain
            """
            Now using project "testproj" on server "https://api.crc.testing:6443".
            """
        And stdout should contain "You can add applications to this project with the 'new-app' command."

    @darwin @linux
    Scenario: Create new app
        When executing "oc new-app cakephp-mysql-example" succeeds
        Then stdout should contain "Creating resources"
        And stdout should contain
            """
            service "cakephp-mysql-example" created
            """
        And check at most "20" times with delay of "60s" that pod "cakephp-mysql-example-1-build" is initialized
        And with up to "10" retries with wait period of "60s" command "curl -kI http://cakephp-mysql-example-testproj.apps-crc.testing" output should contain "HTTP/1.1 200 OK"

    Scenario: Clean up project
        Given executing "oc delete project testproj" succeeds
        When executing "crc stop -f" succeeds
        Then stdout should contain "CodeReady Containers instance stopped"
        When executing "crc delete" succeeds
        Then stdout should contain "CodeReady Containers instance deleted"

