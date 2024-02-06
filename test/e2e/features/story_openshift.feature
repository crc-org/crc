@story_openshift
Feature: 4 Openshift stories

	Background:
		Given ensuring CRC cluster is running
		And ensuring oc command is available
		And ensuring user is logged in succeeds

	# End-to-end health check

	@darwin @linux @windows @testdata @story_health @needs_namespace
	Scenario: Overall cluster health
		Given executing "oc new-project testproj" succeeds
		When executing "oc apply -f httpd-example.yaml" succeeds
		And executing "oc rollout status deployment httpd-example" succeeds
		Then stdout should contain "successfully rolled out"
		When executing "oc create configmap www-content --from-file=index.html=httpd-example-index.html" succeeds
		Then stdout should contain "configmap/www-content created"
		When executing "oc set volume deployment/httpd-example --add --type configmap --configmap-name www-content --name www --mount-path /var/www/html" succeeds
		Then stdout should contain "deployment.apps/httpd-example volume updated"
		When executing "oc expose deployment httpd-example --port 8080" succeeds
		Then stdout should contain "httpd-example exposed"
		When executing "oc expose svc httpd-example" succeeds
		Then stdout should contain "httpd-example exposed"
		When with up to "20" retries with wait period of "5s" http response from "http://httpd-example-testproj.apps-crc.testing" has status code "200"
		Then executing "curl -s http://httpd-example-testproj.apps-crc.testing" succeeds
		And stdout should contain "Hello CRC!"
		When executing "crc stop" succeeds
		And starting CRC with default bundle succeeds
		And checking that CRC is running
		And with up to "4" retries with wait period of "1m" http response from "http://httpd-example-testproj.apps-crc.testing" has status code "200"
		Then executing "curl -s http://httpd-example-testproj.apps-crc.testing" succeeds
		And stdout should contain "Hello CRC!"
		Then with up to "4" retries with wait period of "1m" http response from "http://httpd-example-testproj.apps-crc.testing" has status code "200"

	# Local image to image-registry feature

	@darwin @linux @windows @testdata @story_registry @mirror-registry @needs_namespace
	Scenario: Mirror image to OpenShift image registry
		Given executing "oc new-project testproj" succeeds
		# mirror
		When executing "oc registry login --insecure=true" succeeds
		Then executing "oc image mirror quay.io/centos7/httpd-24-centos7:latest=default-route-openshift-image-registry.apps-crc.testing/testproj/hello:test --insecure=true --filter-by-os=linux/amd64" succeeds
		And executing "oc set image-lookup hello" succeeds
		# deploy
		When executing "oc apply -f hello.yaml" succeeds
		When executing "oc rollout status deployment hello" succeeds
		Then stdout should contain "successfully rolled out"
		When executing "oc get pods" succeeds
		Then stdout should contain "Running"
		When executing "oc logs deployment/hello" succeeds
		Then stdout should contain "httpd"

	@darwin @linux @windows @testdata @story_registry @local-registry @needs_namespace
	Scenario: Pull image locally, push to registry, deploy
		Given podman command is available
		And executing "oc new-project testproj" succeeds
		When pulling image "quay.io/centos7/httpd-24-centos7", logging in, and pushing local image to internal registry succeeds
		And executing "oc apply -f hello.yaml" succeeds
		When executing "oc rollout status deployment hello" succeeds
		Then stdout should contain "successfully rolled out"
		And executing "oc get pods" succeeds
		Then stdout should contain "Running"
		When executing "oc logs deployment/hello" succeeds
		Then stdout should contain "httpd"

	# Operator from marketplace

	@darwin @linux @windows @testdata @story_marketplace @cleanup @needs_namespace
	Scenario: Install new operator
		Given executing "oc new-project testproj" succeeds
		When executing "oc apply -f pipeline-sub.yaml" succeeds
		Then with up to "10" retries with wait period of "30s" command "oc get csv" output matches ".*pipelines-operator(.*)Succeeded$"
		Then with up to "20" retries with wait period of "30s" command "oc get pods -n openshift-pipelines" output matches ".*tekton-pipelines-webhook(.*)1/1(.*)Running.*"
		When with up to "60" retries with wait period of "5s" command "oc explain task" output matches ".*KIND(.*)Task.*"
		When with up to "60" retries with wait period of "5s" command "oc explain taskruns" output matches ".*KIND(.*)TaskRun.*"
		# install tekton task
		When executing "oc apply -f tekton-task.yaml" succeeds
		Then with up to "60" retries with wait period of "5s" command "oc get taskruns" output matches "echo-task-run(.*)Succeeded.*"
		When executing "oc logs --selector=tekton.dev/taskRun=echo-task-run" succeeds
		Then stdout should contain "Hello World"