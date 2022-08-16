@story_registry @darwin @linux @windows
Feature: Local image to image-registry

    User creates a local container image with an app inside it. They then
    push it to the OpenShift image-registry in their project/namespace.
    They deploy and expose the app and check its accessibility.

    @startstop
    Scenario: Start CRC
        Given executing crc setup command succeeds
        When starting CRC with default bundle succeeds
        Then stdout should contain "Started the OpenShift cluster"

    Scenario: Login to cluster
        Given checking that CRC is running
        Then executing "eval $(crc oc-env)" succeeds
        And login to the oc cluster succeeds

    Scenario: Mirror image to OpenShift image registry
        Given executing "oc new-project testproj-img" succeeds
        When executing "oc registry login --insecure=true" succeeds
        Then stdout should contain "Saved credentials for default-route-openshift-image-registry.apps-crc.testing"
        And executing "oc image mirror registry.access.redhat.com/ubi8/httpd-24:latest=default-route-openshift-image-registry.apps-crc.testing/testproj-img/httpd-24:latest --insecure=true --filter-by-os=linux/amd64" succeeds
        And executing "oc set image-lookup httpd-24"


    Scenario: Deploy the image
        Given executing "oc new-app testproj-img/httpd-24:latest" succeeds
        When executing "oc rollout status deployment httpd-24" succeeds
        Then stdout should contain "successfully rolled out"
        When executing "oc get pods" succeeds
        Then stdout should contain "Running"
        When executing "oc logs deployment/httpd-24" succeeds
        Then stdout should contain "httpd"

    @startstop
    Scenario: Clean up
        When executing "crc delete -f" succeeds
        Then stdout should contain "Deleted the instance"
        When executing crc cleanup command succeeds
        Then stdout should contain "Cleanup finished"
