@running_cluster_tests
Feature: Test scenarios on a running cluster

    This feature contains scenarios that require a running cluster.
    1. Verify 9pfs mount of user's home directory is mounted, accessible
       and that basic file operations work. Only relevant on Windows.
    2. User provides configuration property to override default developer user password.

    Background:
        Given ensuring CRC cluster is running

    @fs9p @windows
    Scenario: Test mounted home directory functionality
            Given home directory mount exists in VM
            And filesystem is mounted
            Then listing files in mounted home directory should succeed
            And basic file operations in mounted home directory should succeed
            And basic directory operations in mounted home directory should succeed

    @linux @windows @darwin @cleanup
    Scenario: Override default developer password should be reflected during crc start
        Given executing "crc stop" succeeds
        And setting config property "developer-password" to value "secret-dev" succeeds
        When starting CRC with default bundle succeeds
        Then stdout should contain "Started the OpenShift cluster"
        And stdout should contain "Log in as administrator:"
        And stdout should contain "  Username: kubeadmin"
        And stdout should contain "Log in as user:"
        And stdout should contain "  Username: developer"
        And stdout should contain "  Password: secret-dev"
