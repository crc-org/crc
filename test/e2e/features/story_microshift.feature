@story_microshift
Feature: Microshift test stories

	Background:
		Given setting config property "preset" to value "microshift" succeeds
		And ensuring network mode user
		And executing single crc setup command succeeds
		And starting CRC with default bundle succeeds
		And ensuring oc command is available
		And ensuring microshift cluster is fully operational
		
	# End-to-end health check

	@microshift @testdata @linux @windows @darwin @cleanup
	Scenario: Start and expose a basic HTTP service and check after restart
		Given executing "oc create namespace testproj" succeeds
		And executing "oc config set-context --current --namespace=testproj" succeeds
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
		When with up to "20" retries with wait period of "5s" http response from "http://httpd-example-testproj.apps.crc.testing" has status code "200"
		Then executing "curl -s http://httpd-example-testproj.apps.crc.testing" succeeds
		And stdout should contain "Hello CRC!"
		When executing "crc stop" succeeds
		And starting CRC with default bundle succeeds
		And checking that CRC is running
		And with up to "4" retries with wait period of "1m" http response from "http://httpd-example-testproj.apps.crc.testing" has status code "200"
		Then executing "curl -s http://httpd-example-testproj.apps.crc.testing" succeeds
		And stdout should contain "Hello CRC!"
		Then with up to "4" retries with wait period of "1m" http response from "http://httpd-example-testproj.apps.crc.testing" has status code "200"