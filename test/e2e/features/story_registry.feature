@story_registry @linux
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

    Scenario: Create local image
        Given executing "podman pull registry.access.redhat.com/ubi8/httpd-24" succeeds
        When executing "podman images" succeeds
        Then stdout should contain "registry.access.redhat.com/ubi8/httpd-24"

    Scenario: Push local image to OpenShift image registry
        Given executing "oc new-project testproj-img" succeeds
        When executing "podman login -u kubeadmin -p $(oc whoami -t) default-route-openshift-image-registry.apps-crc.testing --tls-verify=false" succeeds
        Then stdout should contain "Login Succeeded!"
         And executing "podman tag registry.access.redhat.com/ubi8/httpd-24 default-route-openshift-image-registry.apps-crc.testing/testproj-img/hello:test" succeeds
        When executing "podman push default-route-openshift-image-registry.apps-crc.testing/testproj-img/hello:test --tls-verify=false" succeeds

    Scenario: Deploy the image
        Given executing "oc new-app testproj-img/hello:test" succeeds
        When executing "oc rollout status deployment hello" succeeds
        Then stdout should contain "successfully rolled out"
        When executing "oc get pods" succeeds
        Then stdout should contain "Running"
        When executing "oc logs deployment/hello" succeeds
        Then stdout should contain "Apache"

    Scenario: Clean up image and project
        Given executing "podman images" succeeds
        When stdout contains "registry.access.redhat.com/ubi8/httpd-24"
        Then executing "podman image rm registry.access.redhat.com/ubi8/httpd-24" succeeds
        And executing "oc delete project testproj-img" succeeds

    @startstop
    Scenario: Clean up
        When executing "crc delete -f" succeeds
        Then stdout should contain "Deleted the instance"
        When executing crc cleanup command succeeds
        Then stdout should contain "Cleanup finished"
