@basic
Feature: Basic test

    User explores some of the top-level CRC commands while going
    through the lifecycle of CRC.

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
            Machine does not exist. Use 'crc start' to create it
            """

    @linux
    Scenario: CRC setup on Linux
        When executing "crc setup --check-only" fails
        And executing "crc setup" succeeds
        And stderr should contain "Checking if CRC bundle is extracted in '$HOME/.crc'"
        And stderr should contain "Checking if running as non-root"
        And stderr should contain "Checking if Virtualization is enabled"
        And stderr should contain "Checking if KVM is enabled"
        And stderr should contain "Checking if libvirt is installed"
        And stderr should contain "Checking if user is part of libvirt group"
        And stderr should contain "Checking if libvirt daemon is running"
        And stderr should contain "Checking if a supported libvirt version is installed"
        And stderr should contain "Checking if libvirt 'crc' network is available"
        And stderr should contain "Checking if libvirt 'crc' network is active"
        And stderr should contain "Checking if NetworkManager is installed"
        And stderr should contain "Checking if NetworkManager service is running"
        And stderr should contain "Using root access: Executing systemctl daemon-reload command"
        And stderr should contain "Using root access: Executing systemctl reload NetworkManager"
        And stdout should contain "Your system is correctly setup for using CodeReady Containers, you can now run 'crc start -b $bundlename' to start the OpenShift cluster" if bundle is not embedded
        And stdout should contain "Your system is correctly setup for using CodeReady Containers, you can now run 'crc start' to start the OpenShift cluster" if bundle is embedded

    @darwin
    Scenario: CRC setup on Mac
        When executing "crc setup --check-only" fails
        And executing "crc setup" succeeds
        And stderr should contain "Checking if running as non-root"
        And stderr should contain "Checking if HyperKit is installed"
        And stderr should contain "Checking if crc-driver-hyperkit is installed"
        And stderr should contain "Installing crc-machine-hyperkit"
        And stderr should contain "Using root access: Changing ownership"
        And stderr should contain "Using root access: Setting suid"
        And stderr should contain "Checking file permissions"

    @windows
    Scenario: CRC setup on Windows
        When executing "crc setup" succeeds
        Then stderr should contain "Extracting bundle from the CRC executable" if bundle is embedded
        Then stderr should contain "Checking Windows 10 release"
        Then stderr should contain "Checking if Hyper-V is installed"
        Then stderr should contain "Checking if user is a member of the Hyper-V Administrators group"
        Then stderr should contain "Checking if the Hyper-V virtual switch exist"

    @darwin @linux @windows
    Scenario: Request start with monitoring stack
        When setting config property "enable-cluster-monitoring" to value "true" succeeds
        And setting config property "memory" to value "14000" succeeds
        Then starting CRC with default bundle fails
        And stderr should contain "Too little memory"
        And setting config property "memory" to value "16000" succeeds

    @darwin @linux @windows
    Scenario: CRC start
        When starting CRC with default bundle succeeds
        Then stdout should contain "Started the OpenShift cluster"
        # Check if user can copy-paste login details for developer and kubeadmin users
        And stdout should match "(?s)(.*)oc login -u developer https:\/\/api\.crc\.testing:6443(.*)$"
        And stdout should match "(?s)(.*)https:\/\/console-openshift-console\.apps-crc\.testing(.*)$"

    @darwin @linux @windows
    Scenario: CRC status and disk space check
        When checking that CRC is running
        And stdout should match ".*Disk Usage: *\d+[\.\d]*GB of 32.\d+GB.*"

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

    @darwin @linux
    Scenario: Monitoring stack check
        Given checking that CRC is running
        When executing "eval $(crc oc-env)" succeeds
        And login to the oc cluster succeeds
        And executing "oc get pods -n openshift-monitoring" succeeds
        Then stdout matches ".*cluster-monitoring-operator-\w+-\w+\ *2/2\ *Running.*"
        And unsetting config property "enable-cluster-monitoring" succeeds
        And unsetting config property "memory" succeeds

    @windows
    Scenario: Monitoring stack check
        Given checking that CRC is running
        And executing "crc oc-env | Invoke-Expression" succeeds
        And login to the oc cluster succeeds
        And executing "oc get pods -n openshift-monitoring" succeeds
        Then stdout matches ".*cluster-monitoring-operator-\w+-\w+\ *2/2\ *Running.*"
        And unsetting config property "enable-cluster-monitoring" succeeds
        And unsetting config property "memory" succeeds

    @linux
    Scenario: Bundle generation check
        # This will remove the pull secret from the instance and from the cluster
        # You need to provide pull secret file again if you want to start this cluster
        # from a stopped state.
        Given executing "crc bundle generate" succeeds

    @darwin @windows
    Scenario: CRC forcible stop
        When executing "crc stop -f"
        Then stdout should match "(.*)[Ss]topped the OpenShift cluster"
        And executing "oc whoami" fails

    @darwin @linux @windows
    Scenario: CRC status check
        When checking that CRC is stopped
        And stdout should not contain "Running"

    @darwin @linux @windows
    Scenario: CRC console check
        When executing "crc console"
        Then stderr should contain "The OpenShift cluster is not running, cannot open the OpenShift Web Console"

    @darwin @linux @windows
    Scenario: CRC delete
        When executing "crc delete -f" succeeds
        Then stdout should contain "Deleted the OpenShift cluster"

    @linux
    Scenario: CRC starts with generated bundle
        Given starting CRC with custom bundle succeeds
        Then stderr should contain "Using custom bundle"

    @darwin
    Scenario Outline: CRC clean-up
        When executing "crc cleanup" succeeds
        Then stderr should contain "Removing /etc/resolver/testing file"
        And stderr should contain "Unload CodeReady Containers daemon"
        And stderr should contain "Removing pull secret from the keyring"
        And stdout should contain "Cleanup finished"


    @linux
    Scenario Outline: CRC clean-up
        When executing "crc cleanup" succeeds
        Then stderr should contain "Removing the crc VM if exists"
        And stderr should contain "Removing 'crc' network from libvirt"
        And stderr should contain "Using root access: Executing systemctl daemon-reload command"
        And stderr should contain "Using root access: Executing systemctl reload NetworkManager"
        And stderr should contain "Removing pull secret from the keyring"
        And stderr should contain "Removing older logs"
        And stderr should contain "Removing CRC Machine Instance directory"
        And stdout should contain "Cleanup finished"

    @windows
    Scenario Outline: CRC clean-up
        When executing "crc cleanup" succeeds
        Then stderr should contain "Uninstalling tray if installed"
        Then stderr should contain "Uninstalling daemon if installed"
        And stderr should contain "Removing the crc VM if exists"
        And stderr should contain "Removing dns server from interface"
        And stderr should contain "Will run as admin: Remove dns entry for default switch"
        And stdout should contain "Cleanup finished"
