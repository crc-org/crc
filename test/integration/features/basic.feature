@basic
Feature: Basic test
    Checks whether CRC top-level commands behave correctly.

    @darwin @linux @windows
    Scenario: CRC version
        When executing "crc version" succeeds
        Then stderr should be empty
        And stdout should contain "version:"

    @darwin @linux @windows
    Scenario: CRC help
        When executing "crc --help" succeeds
        Then stdout should contain "Usage:"
        And stdout should contain "Available Commands:"
        And stdout should contain "help"
        And stdout should contain "version"
        And stdout should contain "setup"
        And stdout should contain "start"
        And stdout should contain "stop"
        And stdout should contain "delete"
        And stdout should contain "status"
        And stdout should contain "Flags:"
        And stdout should contain 
            """
            Use "crc [command] --help" for more information about a command.
            """

    @darwin @linux @windows
    Scenario: CRC status
        When executing "crc status" fails
        Then stderr should contain 
        """
        Machine "crc" does not exist. Use "crc start" to add a new one.
        """

    @linux
    Scenario: CRC setup on Linux
        When executing "crc setup" succeeds
        Then stdout should contain "Checking if running as non-root" 
        And stdout should contain "Caching oc binary" 
        And stdout should contain "Setting up virtualization" 
        And stdout should contain "Setting up KVM"
        And stdout should contain "Installing libvirt service and dependencies" 
        And stdout should contain "Adding user to libvirt group"
        And stdout should contain "Will use root access: add user to libvirt group"
        And stdout should contain "Enabling libvirt"
        And stdout should contain "Starting libvirt service"
        And stdout should contain "Will use root access: start libvirtd service"
        And stdout should contain "Installing crc-driver-libvirt"
        And stdout should contain "Removing older system-wide crc-driver-libvirt"
        And stdout should contain "Setting up libvirt 'crc' network"
        And stdout should contain "Starting libvirt 'crc' network"
        And stdout should contain "Checking if NetworkManager is installed"
        And stdout should contain "Checking if NetworkManager service is running"
        And stdout should contain "Writing Network Manager config for crc"
        And stdout should contain "Will use root access: write NetworkManager config in /etc/NetworkManager/conf.d/crc-nm-dnsmasq.conf"
        And stdout should contain "Will use root access: execute systemctl daemon-reload command"
        And stdout should contain "Will use root access: execute systemctl stop/start command"
        And stdout should contain "Writing dnsmasq config for crc"
        And stdout should contain "Will use root access: write dnsmasq configuration in /etc/NetworkManager/dnsmasq.d/crc.conf"
        And stdout should contain "Will use root access: execute systemctl daemon-reload command"
        And stdout should contain "Will use root access: execute systemctl stop/start command"
        And stdout should contain "Unpacking bundle from the CRC binary"
        And stdout should contain "Setup is complete"

    @darwin
    Scenario: CRC setup on Mac
        When executing "crc setup" succeeds
        Then stdout should contain "Checking if running as non-root" 
        And stdout should contain "Caching oc binary"
        And stdout should contain "Setting up virtualization"
        And stdout should contain "Will use root access: change ownership"
        And stdout should contain "Will use root access: set suid"
        And stdout should contain "Setting file permissions"

    @linux
    Scenario: CRC start on Linux
        When starting CRC with default bundle and default hypervisor succeeds
        Then stdout should contain "CodeReady Containers instance is running"

    @darwin
    Scenario: CRC start on Mac
        When starting CRC with default bundle and hypervisor "hyperkit" succeeds
        Then stdout should contain "CodeReady Containers instance is running"
    
    @darwin @linux @windows
    Scenario: CRC status check
        When with up to "15" retries with wait period of "1m" command "crc status" output should not contain "Stopped"
        And stdout should contain "Running"

    @darwin @linux @windows
    Scenario: CRC IP check
        When executing "crc ip" succeeds
        Then stdout should match "\d+\.\d+\.\d+\.\d+"

    @darwin @linux @windows
    Scenario: CRC console URL
        When executing "crc console --url" succeeds
        Then stdout should contain "https://console-openshift-console.apps-crc.testing"

    @darwin @linux @windows
    Scenario: CRC console credentials
        When executing "crc console --credentials" succeeds
        Then stdout should contain "To login as a normal user, username is 'developer' and password is 'developer'."
        And stdout should contain "To login as an admin, username is 'kubeadmin' and password is "

    @darwin @linux @windows
    Scenario: CRC forcible stop
        When executing "crc stop -f"
        Then stdout should match "CodeReady Containers instance(.*)stopped"

    @darwin @linux @windows
    Scenario: CRC status check
        When with up to "2" retries with wait period of "1m" command "crc status" output should not contain "Running"
        And stdout should contain "Stopped"

    @darwin @linux @windows
    Scenario: CRC console check
        Given executing "crc status" succeeds
        And stdout contains "Stopped"
        When executing "crc console"
        Then stderr should contain "CodeReady Containers instance is not running, cannot open the OpenShift Web Console."

    @darwin @linux @windows
    Scenario: CRC delete
        When executing "crc delete -f" succeeds
        Then stdout should contain "CodeReady Containers instance deleted"
