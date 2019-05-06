@basic @quick
Feature: Basic test
Checks whether CRC top-level commands behave correctly.
	
  Scenario: CRC version
    When executing "crc version" succeeds
    Then stderr should be empty
    And stdout should contain "version:"
    
  Scenario: CRC help
    When executing "crc --help" succeeds
    Then stdout should contain "Usage:"
    And stdout should contain "Available Commands:"
    And stdout should contain "Flags:"
    And stdout should contain 
    """
    Use "crc [command] --help" for more information about a command.
    """

  Scenario: CRC setup
    When executing "crc setup" succeeds
    Then stdout should contain "Caching oc binary"
    And stdout should contain "Starting Libvirt crc network"
    And stdout should contain "Setting up virtualization"
    And stdout should contain "Setting up KVM"
    And stdout should contain "Installing Libvirt"
    And stdout should contain "Adding user to libvirt group"
    And stdout should contain "Enabling libvirt"
    And stdout should contain "Starting Libvirt service"
    And stdout should contain "Installing crc-driver-libvirt"
    And stdout should contain "Setting up Libvirt crc network"
    And stdout should contain "Starting Libvirt crc network"

  Scenario: CRC start
    When executing "crc start -b ~/Downloads/crc_libvirt_4.1.0-rc.5.tar.xz" succeeds
    Then stdout should contain "Running"
    
  Scenario: CRC kill
    When executing "crc stop -f"
    Then stdout should contain "true"
    
  Scenario: CRC delete
    When executing "crc delete" succeeds
    Then stdout should contain "true"
