@story_custom_developer_password
Feature: Custom Developer Password Test

    User provides configuration property to override default developer user password

    Background:
        Given ensuring CRC cluster is running

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
