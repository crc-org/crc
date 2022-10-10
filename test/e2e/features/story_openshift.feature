@story_openshift
Feature: 3 Openshift stories

	Background:
		Given ensuring CRC cluster is running succeeds
		And ensuring user is logged in succeeds

	# Old: End-to-end health check

	@darwin @linux @windows @startstop @testdata
	Scenario: Overall cluster health
		#Given ensuring CRC cluster is running succeeds
		#And ensuring user is logged in succeeds
		Given checking that CRC is running
		Then executing "oc new-project testproj" succeeds
		And executing "oc create deployment httpd-example --image=registry.access.redhat.com/ubi8/httpd-24 --port=8080" succeeds
		When executing "oc rollout status deployment httpd-example" succeeds
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
		And executing "oc delete project testproj" succeeds

	# Old: Local image to image-registry feature

	@linux
	Scenario: Create local image, push to registry, deploy
		Given checking that CRC is running
		And executing "podman pull quay.io/centos7/httpd-24-centos7" succeeds
		When executing "oc new-project testproj-img" succeeds
		And executing "podman login -u kubeadmin -p $(oc whoami -t) default-route-openshift-image-registry.apps-crc.testing --tls-verify=false" succeeds
		And executing "podman tag quay.io/centos7/httpd-24-centos7 default-route-openshift-image-registry.apps-crc.testing/testproj-img/hello:test" succeeds
		And executing "podman push default-route-openshift-image-registry.apps-crc.testing/testproj-img/hello:test --tls-verify=false" succeeds
		And executing "oc new-app testproj-img/hello:test" succeeds
		When executing "oc rollout status deployment hello" succeeds
		Then stdout should contain "successfully rolled out"
		And executing "oc get pods" succeeds
		Then stdout should contain "Running"
		When executing "oc logs deployment/hello" succeeds
		Then stdout should contain "httpd"
		#And executing "podman image rm quay.io/bitnami/nginx" succeeds
		And executing "oc delete project testproj-img" succeeds

	# Old: Operator from marketplace

	@darwin @linux @windows @testdata
	Scenario: Install new operator
		Given checking that CRC is running
		When executing "oc apply -f redis-sub.yaml" succeeds
		Then with up to "20" retries with wait period of "30s" command "oc get csv" output matches ".*redis-operator\.(.*)Succeeded$"
		# install redis operator
		When executing "oc apply -f redis-cluster.yaml" succeeds
		Then with up to "10" retries with wait period of "30s" command "oc get pods" output matches "redis-standalone-[a-z0-9]* .*Running.*"
		When deleting a pod succeeds
		Then stdout should match "^pod(.*)deleted$"
		# after a while 1 pods should be up & running again
		And with up to "10" retries with wait period of "30s" command "oc get pods" output matches "redis-standalone-[a-z0-9]* .*Running.*"
		# cleanup
		When executing "crc delete -f" succeeds
		And executing crc cleanup command succeeds


