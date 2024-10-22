@story_application_deployment
Feature: Application Deployment Test

    User deploys a basic java application onto CRC cluster and expects that it's
    deployed successfully and is accessible via route

    Background:
        Given ensuring CRC cluster is running
        And ensuring oc command is available
        And ensuring user is logged in succeeds

    @testdata @linux @windows @darwin @cleanup @needs_namespace
    Scenario: Deploy a java application using Eclipse JKube in pod and then verify it's health
        When executing "oc new-project testproj" succeeds
        And executing "oc create -f jkube-kubernetes-build-resources.yaml" succeeds
        And executing "oc start-build jkube-application-deploy-buildconfig --follow" succeeds
        And executing "oc rollout status -w dc/jkube-application-deploy-test --timeout=600s" succeeds
        Then stdout should contain "successfully rolled out"
        And executing "oc logs -lapp=jkube-application-deploy-test -f" succeeds
        And executing "oc rollout status -w dc/quarkus --timeout=600s" succeeds
        Then stdout should contain "successfully rolled out"
        And executing "oc get build -lapp=quarkus" succeeds
        Then stdout should contain "quarkus-s2i"
        And executing "oc get buildconfig -lapp=quarkus" succeeds
        Then stdout should contain "quarkus-s2i"
        And executing "oc get imagestream -lapp=quarkus" succeeds
        Then stdout should contain "quarkus"
        And executing "oc get pods -lapp=quarkus" succeeds
        Then stdout should contain "quarkus"
        And stdout should contain "1/1     Running"
        And executing "oc get svc -lapp=quarkus" succeeds
        Then stdout should contain "quarkus"
        And executing "oc get routes -lapp=quarkus" succeeds
        Then stdout should contain "quarkus"
        And with up to "4" retries with wait period of "1m" http response from "http://quarkus-testproj.apps-crc.testing" has status code "200"
        Then executing "curl -s http://quarkus-testproj.apps-crc.testing" succeeds
        And stdout should contain "{"applicationName":"JKube","message":"Subatomic JKube really whips the llama's ass!"}"
