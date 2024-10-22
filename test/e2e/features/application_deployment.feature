@story_application_deployment
Feature: Application Deployment Test

    User deploys a basic java application onto CRC cluster and expects that it's
    deployed successfully and is accessible via route

    Background:
        Given ensuring CRC cluster is running
        And ensuring oc command is available
        And ensuring user is logged in succeeds

    @linux @windows @darwin
    Scenario: Deploy a java application using Eclipse JKube in podman container and then verify it's health
        When executing "podman build --tag localhost/redhat-ubi-openjdk17-with-oc:latest -f ../../../testdata/JKubeAppDeployment_Containerfile   ." succeeds
        And executing "podman run --rm --dns=8.8.8.8 --name application-deployment-build-e2e --add-host=oauth-openshift.apps-crc.testing:host-gateway localhost/redhat-ubi-openjdk17-with-oc:latest" succeeds
        And executing "oc rollout status -w dc/quarkus --namespace jkube-quarkus-app-deploy-flow-test --timeout=600s" succeeds
        Then stdout should contain "successfully rolled out"
        And executing "oc get build -lapp=quarkus --namespace jkube-quarkus-app-deploy-flow-test" succeeds
        Then stdout should contain "quarkus-s2i"
        And executing "oc get buildconfig -lapp=quarkus --namespace jkube-quarkus-app-deploy-flow-test" succeeds
        Then stdout should contain "quarkus-s2i"
        And executing "oc get imagestream -lapp=quarkus --namespace jkube-quarkus-app-deploy-flow-test" succeeds
        Then stdout should contain "quarkus"
        And executing "oc get pods -lapp=quarkus --namespace jkube-quarkus-app-deploy-flow-test" succeeds
        Then stdout should contain "quarkus"
        And stdout should contain "1/1     Running"
        And executing "oc get svc -lapp=quarkus --namespace jkube-quarkus-app-deploy-flow-test" succeeds
        Then stdout should contain "quarkus"
        And executing "oc get routes -lapp=quarkus --namespace jkube-quarkus-app-deploy-flow-test" succeeds
        Then stdout should contain "quarkus"
        And with up to "4" retries with wait period of "1m" http response from "http://quarkus-jkube-quarkus-app-deploy-flow-test.apps-crc.testing" has status code "200"
#        Then executing "curl -s http://quarkus-jkube-quarkus-app-deploy-flow-test.apps-crc.testing" succeeds
#        And stdout should contain "{\"applicationName\":\"JKube\",\"message\":\"Subatomic JKube really whips the llama's ass!\"}"
#        Workaround added for above two statements, they're not working as expected
        Then deployed application is running in cluster
        # cleanup
        When executing "oc delete project jkube-quarkus-app-deploy-flow-test" succeeds
        And executing "podman rmi -f localhost/redhat-ubi-openjdk17-with-oc:latest" succeeds
