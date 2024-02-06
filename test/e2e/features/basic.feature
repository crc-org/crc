@basic
Feature: Basic test

    User explores some of the top-level CRC commands while going
    through the lifecycle of CRC.

    @darwin @linux @windows
    Scenario: CRC version
        When executing crc version command
        Then stdout should contain correct version

    @darwin @linux @windows
    Scenario: CRC help
        When executing crc help command succeeds
        Then stdout should contain "Usage:"
        And stdout should contain "Available Commands:"
        And stdout should contain "help"
        And stdout should contain "version"
        And stdout should contain "setup"
        And stdout should contain "start"
        And stdout should contain "stop"
        And stdout should contain "delete"
        And stdout should contain "status"
        And stdout should contain "Flags:"

    @darwin @linux @windows
    Scenario: CRC status
        When executing crc status command fails
        Then stderr should contain "crc does not seem to be setup correctly, have you run 'crc setup'?"

    @darwin @linux @windows @cleanup
    Scenario: CRC start usecase
        Given executing "crc setup --check-only" fails
        # Request start with monitoring stack
        * setting config property "enable-cluster-monitoring" to value "true" succeeds
        * setting config property "memory" to value "16000" succeeds
        Given executing single crc setup command succeeds
        When starting CRC with default bundle succeeds
        Then stdout should contain "Started the OpenShift cluster"
        # Check if user can copy-paste login details for developer and kubeadmin users
        * stdout should match "(?s)(.*)oc login -u developer https:\/\/api\.crc\.testing:6443(.*)$"
        * stdout should match "(?s)(.*)https:\/\/console-openshift-console\.apps-crc\.testing(.*)$"
        # status
        When checking that CRC is running
        # ip
        When executing crc ip command succeeds
        Then stdout should match "\d+\.\d+\.\d+\.\d+"
        # console url
        When executing "crc console --url" succeeds
        Then stdout should contain "https://console-openshift-console.apps-crc.testing"
        # console credentials
        When executing "crc console --credentials" succeeds
        Then stdout should contain "To login as a regular user, run 'oc login -u developer -p developer"
        And stdout should contain "To login as an admin, run 'oc login -u kubeadmin -p "
        # monitoring stack check
        When checking that CRC is running
        And ensuring user is logged in succeeds
        Then with up to "12" retries with wait period of "10s" command "oc get pods -n openshift-monitoring" output matches ".*cluster-monitoring-operator-\w+-\w+\ *1/1\ *Running.*"
        # stop
        When executing "crc stop"
        Then stdout should match "(.*)[Ss]topped the instance"
        And executing "oc whoami" fails
        # status check
        When checking that CRC is stopped
        And stdout should not contain "Running"
        # console check
        When executing crc console command
        Then stderr should contain "The OpenShift cluster is not running, cannot open the OpenShift Web Console"
        # delete
        When executing "crc delete -f" succeeds
        Then stdout should contain "Deleted the instance"
        # cleanup
        When executing crc cleanup command succeeds
