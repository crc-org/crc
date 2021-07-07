@story_marketplace
Feature: Operator from marketplace

    User installs an OpenShift operator from OperatorHub and uses
    it to manage admin tasks.

    @linux @darwin
    Scenario: Start CRC and login to cluster
        Given executing "crc setup" succeeds
        When starting CRC with default bundle succeeds
        Then stdout should contain "Started the OpenShift cluster"
        When checking that CRC is running
        Then executing "eval $(crc oc-env)" succeeds
        And login to the oc cluster succeeds

    @windows
    Scenario: Start CRC on Windows
        Given executing "crc setup" succeeds
        When starting CRC with default bundle and nameserver "10.75.5.25" succeeds
        Then stdout should contain "Started the OpenShift cluster"
        And executing "crc oc-env | Invoke-Expression" succeeds
        When checking that CRC is running
        Then login to the oc cluster succeeds

    @darwin @linux @windows @testdata
    Scenario: Install new operator
        When executing "oc apply -f redis-sub.yaml" succeeds
        # check if cluster operator is running
        Then with up to "20" retries with wait period of "30s" command "oc get csv" output matches ".*redis-operator\.(.*)Succeeded$"
        
    @darwin @linux @windows @testdata
    Scenario: Install the redis instance
        When executing "oc apply -f redis-cluster.yaml" succeeds
        Then with up to "10" retries with wait period of "30s" command "oc get pods" output matches "redis-standalone-[a-z0-9]* .*Running.*"
    
    @darwin @linux
    Scenario: Failover
        # simulate failure of 1 pod, check that it was replaced
        When executing "POD=$(oc get pod -o jsonpath="{.items[0].metadata.name}")" succeeds
        And executing "echo $POD" succeeds
        And executing "oc delete pod $POD --now" succeeds
        Then stdout should match "^pod(.*)deleted$"
        # after a while 1 pods should be up & running again
        And with up to "10" retries with wait period of "30s" command "oc get pods" output matches "redis-standalone-[a-z0-9]* .*Running.*"

    @windows
    Scenario: Failover
        # simulate failure of 1 pod, check that it was replaced
        When executing "$Env:POD = $(oc get pod -o jsonpath="{.items[0].metadata.name}")" succeeds
        And executing "echo $Env:POD" succeeds
        And executing "oc delete pod $Env:POD --now" succeeds
        Then stdout should match "^pod(.*)deleted$"
        # after a while 1 pods should be up & running again
        And with up to "10" retries with wait period of "30s" command "oc get pods" output matches "redis-standalone-[a-z0-9]* .*Running.*"
        
    @darwin @linux @windows
    Scenario: Clean up
        When executing "crc delete -f" succeeds
        Then stdout should contain "Deleted the OpenShift cluster"
        When executing "crc cleanup" succeeds
        Then stdout should contain "Cleanup finished"
