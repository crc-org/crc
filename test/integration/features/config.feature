@quick
Feature: Test configuration settings
Checks whether CRC `config set` command works as expected in conjunction with `crc setup` and `crc start`.

# WARN

    Scenario Outline: CRC config checks (warnings)
        When executing "<check>" succeeds
        Then file "crc.json" exists in CRC home folder
        And "JSON" config file "crc.json" in CRC home folder contains key "<property>" with value matching "true"
        When executing "<nocheck>" succeeds
        Then "JSON" config file "crc.json" in CRC home folder contains key "<property>" with value matching "false"

        Examples: Config options
            | check                                                | nocheck                                               | property                         |
            | crc config set warn-check-user-in-libvirt-group true | crc config set warn-check-user-in-libvirt-group false | warn-check-user-in-libvirt-group |
            | crc config set warn-check-crc-dnsmasq-file true      | crc config set warn-check-crc-dnsmasq-file false      | warn-check-crc-dnsmasq-file      |
            | crc config set warn-check-virt-enabled true          | crc config set warn-check-virt-enabled false          | warn-check-virt-enabled          |
            | crc config set warn-check-kvm-enabled true           | crc config set warn-check-kvm-enabled false           | warn-check-kvm-enabled           |
            | crc config set warn-check-libvirt-driver true        | crc config set warn-check-libvirt-driver false        | warn-check-libvirt-driver        |
            | crc config set warn-check-libvirt-installed true     | crc config set warn-check-libvirt-installed false     | warn-check-libvirt-installed     |
            | crc config set warn-check-libvirt-enabled true       | crc config set warn-check-libvirt-enabled false       | warn-check-libvirt-enabled       |
            | crc config set warn-check-libvirt-running true       | crc config set warn-check-libvirt-running false       | warn-check-libvirt-running       |
            | crc config set warn-check-crc-network true           | crc config set warn-check-crc-network false           | warn-check-crc-network           |
            | crc config set warn-check-crc-network-active true    | crc config set warn-check-crc-network-active false    | warn-check-crc-network-active    |

# SKIP

    Scenario Outline: CRC config checks (skips)
        When executing "<check>" succeeds
        Then file "crc.json" exists in CRC home folder
        And "JSON" config file "crc.json" in CRC home folder contains key "<property>" with value matching "true"
        When executing "crc setup" succeeds
        Then stdout should contain "Skipping above check ..."
        When executing "<nocheck>" succeeds
        Then "JSON" config file "crc.json" in CRC home folder contains key "<property>" with value matching "false"
        When executing "crc setup" succeeds
        Then stdout should not contain "Skipping above check ..."
        
        Examples:
            | check                                                | nocheck                                               | property                         |
            | crc config set skip-check-virt-enabled true          | crc config set skip-check-virt-enabled false          | skip-check-virt-enabled          |
            | crc config set skip-check-kvm-enabled true           | crc config set skip-check-kvm-enabled false           | skip-check-kvm-enabled           |
            | crc config set skip-check-libvirt-enabled true       | crc config set skip-check-libvirt-enabled false       | skip-check-libvirt-enabled       |
            | crc config set skip-check-libvirt-running true       | crc config set skip-check-libvirt-running false       | skip-check-libvirt-running       |
            | crc config set skip-check-crc-network-active true    | crc config set skip-check-crc-network-active false    | skip-check-crc-network-active    |
            | crc config set skip-check-libvirt-installed true     | crc config set skip-check-libvirt-installed false     | skip-check-libvirt-installed     |
            | crc config set skip-check-libvirt-driver true        | crc config set skip-check-libvirt-driver false        | skip-check-libvirt-driver        |
            | crc config set skip-check-crc-network true           | crc config set skip-check-crc-network false           | skip-check-crc-network           |
            | crc config set skip-check-crc-dnsmasq-file true      | crc config set skip-check-crc-dnsmasq-file false      | skip-check-crc-dnsmasq-file      |
            | crc config set skip-check-user-in-libvirt-group true | crc config set skip-check-user-in-libvirt-group false | skip-check-user-in-libvirt-group |

# --------------------------------------
# Specific Scenarios

   Scenario: Check network setup and destroy it, then check again
       When removing file "crc.json" from CRC home folder succeeds
       And executing "crc setup" succeeds
       And executing "sudo virsh net-list --name" succeeds
       Then stdout contains "crc"
       When executing "sudo virsh net-undefine crc && sudo virsh net-destroy crc" succeeds
       And executing "sudo virsh net-list --name" succeeds
       Then stdout should not contain "crc"

   Scenario: Running `crc setup` with checks enabled restores destroyed network
       # When network destroyed, crc start should fail
       When executing "crc config set skip-check-crc-network false" succeeds
       And executing "crc config set skip-check-crc-network-active false" succeeds
       Then executing "crc setup" succeeds
       And executing "sudo virsh net-list --name" succeeds
       And stdout contains "crc"

   Scenario: Running `crc start` without `crc setup` and with checks disabled fails when network destroyed
       # Destroy network again
       When executing "sudo virsh net-undefine crc && sudo virsh net-destroy crc" succeeds
       And executing "sudo virsh net-list --name" succeeds
       Then stdout should not contain "crc"
       # Disable checks
       When executing "crc config set skip-check-crc-network true" succeeds
       And executing "crc config set skip-check-crc-network-active true" succeeds
       # Start CRC
       Then executing "crc start -b ~/Downloads/crc_libvirt_4.1.0-rc.5.tar.xz" fails
       And stdout contains "Network not found: no network with matching name 'crc'"

   Scenario: Clean-up
       # Remove the config file
       When removing file "crc.json" from CRC home folder succeeds
       And executing "crc setup" succeeds
       Then stdout should not contain "Skipping above check"
       
