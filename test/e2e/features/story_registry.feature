@story_registry @linux
Feature: Local image to image-registry

    User creates a local container image with an app inside it. They then
    push it to the OpenShift image-registry in their project/namespace.
    They deploy and expose the app and check its accessibility.

    Scenario: Start CRC and login to cluster
        Given execute crc setup command succeeds
        When starting CRC with default bundle succeeds
        Then stdout should contain "Started the OpenShift cluster"
        When checking that CRC is running
        Then executing "eval $(crc oc-env)" succeeds
        And login to the oc cluster succeeds

    Scenario: Create local image
        Given executing "podman pull quay.io/bitnami/nginx" succeeds
        When executing "podman images" succeeds
        Then stdout should contain "quay.io/bitnami/nginx"
        
    Scenario: Push local image to OpenShift image registry
        Given executing "oc new-project testproj-img" succeeds
        When executing "podman login -u kubeadmin -p $(oc whoami -t) default-route-openshift-image-registry.apps-crc.testing --tls-verify=false" succeeds
        Then stdout should contain "Login Succeeded!"
         And executing "podman tag quay.io/bitnami/nginx:latest default-route-openshift-image-registry.apps-crc.testing/testproj-img/hello:test" succeeds
        When executing "podman push default-route-openshift-image-registry.apps-crc.testing/testproj-img/hello:test --tls-verify=false" succeeds

    Scenario: Deploy the image
        Given executing "oc new-app testproj-img/hello:test" succeeds
        When executing "oc rollout status deployment hello" succeeds
        Then stdout should contain "successfully rolled out"
        When executing "oc get pods" succeeds
        Then stdout should contain "Running"
        When executing "oc logs deployment/hello" succeeds
        Then stdout should contain "Starting NGINX"

    Scenario: Clean up
        Given executing "podman images" succeeds
        When stdout contains "quay.io/bitnami/nginx"
        Then executing "podman image rm quay.io/bitnami/nginx" succeeds
        And executing "oc delete project testproj-img" succeeds
        And executing "crc delete -f" succeeds
        Then stdout should contain "Deleted the OpenShift cluster"
        When execute crc cleanup command succeeds
        Then stdout should contain "Cleanup finished"
