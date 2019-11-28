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
        Machine 'crc' does not exist. Use 'crc start' to create it.
        """

    @linux
    Scenario: CRC setup on Linux
        When executing "crc setup" succeeds
        Then stdout should contain "Caching oc binary"
        And stdout should contain "Checking if CRC bundle is cached in '$HOME/.crc'"
        And stdout should contain "Checking if running as non-root"
        And stdout should contain "Checking if Virtualization is enabled"
        And stdout should contain "Checking if KVM is enabled"
        And stdout should contain "Checking if libvirt is installed"
        And stdout should contain "Checking if user is part of libvirt group"
        And stdout should contain "Checking if libvirt is enabled"
        And stdout should contain "Checking if libvirt daemon is running"
        And stdout should contain "Checking if a supported libvirt version is installed"
        And stdout should contain "Checking for obsolete crc-driver-libvirt"
        And stdout should contain "Checking if libvirt 'crc' network is available"
        And stdout should contain "Checking if libvirt 'crc' network is active"
        And stdout should contain "Checking if NetworkManager is installed"
        And stdout should contain "Checking if NetworkManager service is running"
        And stdout should contain "Checking if /etc/NetworkManager/conf.d/crc-nm-dnsmasq.conf exists"
        And stdout should contain "Writing Network Manager config for crc"
        And stdout should contain "Will use root access: write NetworkManager config in /etc/NetworkManager/conf.d/crc-nm-dnsmasq.conf"
        And stdout should contain "Will use root access: execute systemctl daemon-reload command"
        And stdout should contain "Will use root access: execute systemctl stop/start command"
        And stdout should contain "Checking if /etc/NetworkManager/dnsmasq.d/crc.conf exists"
        And stdout should contain "Writing dnsmasq config for crc"
        And stdout should contain "Will use root access: write dnsmasq configuration in /etc/NetworkManager/dnsmasq.d/crc.conf"
        And stdout should contain "Will use root access: execute systemctl daemon-reload command"
        And stdout should contain "Will use root access: execute systemctl stop/start command"
        And stdout should contain "Setup is complete, you can now run 'crc start -b $bundlename' to start the OpenShift cluster" if bundle is not embedded
        And stdout should contain "Setup is complete, you can now run 'crc start' to start the OpenShift cluster" if bundle is embedded

    @darwin
    Scenario: CRC setup on Mac
        When executing "crc setup" succeeds
        Then stdout should contain "Caching oc binary"
        And stdout should contain "Checking if running as non-root"
        And stdout should contain "Checking if HyperKit is installed"
        And stdout should contain "Checking if crc-driver-hyperkit is installed"
        And stdout should contain "Installing crc-machine-hyperkit"
        And stdout should contain "Will use root access: change ownership"
        And stdout should contain "Will use root access: set suid"
        And stdout should contain "Checking file permissions"

    @windows
    Scenario: CRC setup on Windows
        When executing "crc setup" succeeds
        Then stdout should contain "Caching oc binary"
        Then stdout should contain "Unpacking bundle from the CRC binary"
        Then stdout should contain "Checking Windows 10 release"
        Then stdout should contain "Checking if Hyper-V is installed"
        Then stdout should contain "Checking if user is a member of the Hyper-V Administrators group"
        Then stdout should contain "Checking if the Hyper-V virtual switch exist"

    @linux @windows
    Scenario: CRC start
        When starting CRC with default bundle and default hypervisor succeeds
        Then stdout should contain "Started the OpenShift cluster"

    @darwin
    Scenario: CRC start on Mac
        When starting CRC with default bundle and hypervisor "hyperkit" succeeds
        Then stdout should contain "Started the OpenShift cluster"
    
    @darwin @linux @windows
    Scenario: CRC status and disk space check
        When with up to "15" retries with wait period of "1m" command "crc status" output should not contain "Stopped"
        And stdout should contain "Running"
        And stdout should match ".*Disk Usage: *\d+\.\d+GB of 32.\d+GB.*"

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
        Then stdout should contain "To login as a regular user, run 'oc login -u developer -p developer"
        And stdout should contain "To login as an admin, run 'oc login -u kubeadmin -p "

    @darwin @linux @windows
    Scenario: CRC forcible stop
        When executing "crc stop -f"
        Then stdout should match "(.*)[Ss]topped the OpenShift cluster"

    @darwin @linux @windows
    Scenario: CRC status check
        When with up to "2" retries with wait period of "1m" command "crc status" output should not contain "Running"
        And stdout should contain "Stopped"

    @darwin @linux @windows
    Scenario: CRC console check
        Given executing "crc status" succeeds
        And stdout contains "Stopped"
        When executing "crc console"
        Then stderr should contain "The OpenShift cluster is not running, cannot open the OpenShift Web Console."

    @darwin @linux @windows
    Scenario: CRC delete
        When executing "crc delete -f" succeeds
        Then stdout should contain "Deleted the OpenShift cluster"
