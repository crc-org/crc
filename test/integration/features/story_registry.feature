@story_registry @linux
Feature: Local image to image-registry to deployment

    The user creates a local container image with an app. They then
    push it to the openshift image-registry in their
    project/namespace. They deploy and expose the app and check its
    accessibility.

    Scenario Outline: Start CRC
        Given executing "crc setup" succeeds
        When starting CRC with default bundle and hypervisor "<vm-driver>" succeeds
        Then stdout should contain "The OpenShift cluster is running"
        And executing "eval $(crc oc-env)" succeeds
        When with up to "4" retries with wait period of "2m" command "crc status" output matches ".*Running \(v\d+\.\d+\.\d+.*\).*"
        Then login to the oc cluster succeeds

        Examples:
            | vm-driver |
            | libvirt   |

    Scenario: Create local image
        Given executing "cd ../../../testdata" succeeds
        When executing "sudo podman build -t hello:test ." succeeds
        And executing "cd ../integration" succeeds
        Then executing "sudo podman images" succeeds
        And stdout should contain "localhost/hello"
        
    Scenario: Push local image to OpenShift image registry
        Given executing "oc new-project testproj-img" succeeds
        When executing "sudo podman login -u kubeadmin -p $(oc whoami -t) default-route-openshift-image-registry.apps-crc.testing --tls-verify=false" succeeds
        Then stdout should contain "Login Succeeded!"
        When executing "sudo podman push hello:test default-route-openshift-image-registry.apps-crc.testing/testproj-img/hello:test --tls-verify=false" succeeds

    Scenario: Deploy the image
        Given executing "oc new-app testproj-img/hello:test" succeeds
        When executing "oc rollout status dc/hello" succeeds
        Then stdout should contain "successfully rolled out"
        When executing "oc get pods" succeeds
        Then stdout should contain "Running"
        And stdout should contain "Completed"
        When executing "oc logs -f dc/hello" succeeds
        Then stdout should contain "Hello, it works!"

