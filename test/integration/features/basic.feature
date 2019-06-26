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
        And stdout should contain "Flags:"
        And stdout should contain 
            """
            Use "crc [command] --help" for more information about a command.
            """
    @linux
    Scenario: CRC setup on Linux
        When executing "crc setup" succeeds
        Then stdout should contain "Caching oc binary"
        And stdout should contain "Starting libvirt 'crc' network"
        And stdout should contain "Setting up virtualization"
        And stdout should contain "Setting up KVM"
        And stdout should contain "Installing libvirt"
        And stdout should contain "Adding user to libvirt group"
        And stdout should contain "Enabling libvirt"
        And stdout should contain "Starting libvirt service"
        And stdout should contain "Installing crc-driver-libvirt"
        And stdout should contain "Setting up libvirt 'crc' network"
        And stdout should contain "Starting libvirt 'crc' network"

    @darwin
    Scenario: CRC setup on Mac
        When executing "crc setup" succeeds
        Then stdout should contain "Caching oc binary"
        And stdout should contain "Setting up virtualization"
        And stdout should contain "Setting file permissions for resolver"

    @linux
    Scenario: CRC start on Linux
        When starting CRC with default bundle and default hypervisor succeeds
        Then stdout should contain "CodeReady Containers instance is running"

    @darwin
    Scenario: CRC start on Mac
        When starting CRC with default bundle and hypervisor "virtualbox" succeeds
        Then stdout should contain "CodeReady Containers instance is running"

    @darwin @linux @windows
    Scenario: CRC IP
        When executing "crc ip" succeeds
        Then stdout should match "\d+\.\d+\.\d+\.\d+"

    @darwin @linux @windows
    Scenario: CRC forcible stop
        When executing "crc stop -f"
        Then stdout should contain "CodeReady Containers instance stopped"
    
    @darwin @linux @windows
    Scenario: CRC delete
        When executing "crc delete" succeeds
        Then stdout should contain "CodeReady Containers instance deleted"
