Feature: Test configuration settings
Checks whether CRC `config set` command works as expected in conjunction with `crc setup` and `crc start`.

# SETTINGS

    Scenario Outline: CRC config checks (settings)
        When setting config property "<property>" to value "<value1>" succeeds
        Then file "crc.json" exists in CRC home folder
        And "JSON" config file "crc.json" in CRC home folder contains key "<property>" with value matching "<value1>"
        And setting config property "<property>" to value "<value2>" fails
        
        # always return to default values
        @darwin
        Examples: Config settings on Mac 
            | property  | value1                | value2                                     |
            | cpus      | 4                     | 3                                          |
            | memory    | 8192                  | 4096                                       |
            | vm-driver | virtualbox            | libvirt                                    |

        @linux
        Examples: Config settings on Linux
            | property  | value1                   | value2                                               |
            | cpus      | 4                        | 3                                                    |
            | memory    | 8192                     | 4096                                                 |
            | vm-driver | libvirt                  | hyperkit                                             |

    @darwin @linux
    Scenario: CRC config checks (bundle version)
        When setting config property "bundle" to value "current bundle" succeeds
        And "JSON" config file "crc.json" in CRC home folder contains key "bundle" with value matching "current bundle"
        And setting config property "bundle" to value "/path/to/nonexistent/bundle/crc_hypervisor_version.tar.xz" fails

# WARNINGS

    Scenario Outline: CRC config checks (warnings)
        When setting config property "<property>" to value "<value1>" succeeds
        Then "JSON" config file "crc.json" in CRC home folder contains key "<property>" with value matching "<value1>"
        When setting config property "<property>" to value "<value2>" succeeds
        Then "JSON" config file "crc.json" in CRC home folder contains key "<property>" with value matching "<value2>"

        @darwin
        Examples: Config warnings on Mac
            | property                             | value1 | value2 |
            | warn-check-virtualbox-installed      | true   | false  |
            | warn-check-resolver-file-permissions | true   | false  |

        @linux
        Examples: Config warnings on Linux
            | property                         | value1 | value2 |
            | warn-check-user-in-libvirt-group | true   | false  |
            | warn-check-crc-dnsmasq-file      | true   | false  |
            | warn-check-virt-enabled          | true   | false  |
            | warn-check-kvm-enabled           | true   | false  |
            | warn-check-libvirt-driver        | true   | false  |
            | warn-check-libvirt-installed     | true   | false  |
            | warn-check-libvirt-enabled       | true   | false  |
            | warn-check-libvirt-running       | true   | false  |
            | warn-check-crc-network           | true   | false  |
            | warn-check-crc-network-active    | true   | false  |

# SKIP

    Scenario Outline: CRC config checks (skips)
        When setting config property "<property>" to value "<value1>" succeeds
        And "JSON" config file "crc.json" in CRC home folder contains key "<property>" with value matching "<value1>"
        When executing "crc setup" succeeds
        Then stdout should contain "Skipping above check ..."
        When setting config property "<property>" to value "<value2>" succeeds
        Then "JSON" config file "crc.json" in CRC home folder contains key "<property>" with value matching "<value2>"
        When executing "crc setup" succeeds
        Then stdout should not contain "Skipping above check ..."
        
        @darwin
        Examples:
            | property                             | value1 | value2 |
            | skip-check-virtualbox-installed      | true   | false  |
            | skip-check-resolver-file-permissions | true   | false  |

        @linux
        Examples:
            | property                         | value1 | value2 |
            | skip-check-virt-enabled          | true   | false  |
            | skip-check-kvm-enabled           | true   | false  |
            | skip-check-libvirt-enabled       | true   | false  |
            | skip-check-libvirt-running       | true   | false  |
            | skip-check-crc-network-active    | true   | false  |
            | skip-check-libvirt-installed     | true   | false  |
            | skip-check-libvirt-driver        | true   | false  |
            | skip-check-crc-network           | true   | false  |
            | skip-check-crc-dnsmasq-file      | true   | false  |
            | skip-check-user-in-libvirt-group | true   | false  |

# --------------------------------------
# Linux-specific Scenarios

   @linux
   Scenario: Check network setup and destroy it, then check again
       When removing file "crc.json" from CRC home folder succeeds
       And executing "crc setup" succeeds
       And executing "sudo virsh net-list --name" succeeds
       Then stdout contains "crc"
       When executing "sudo virsh net-undefine crc && sudo virsh net-destroy crc" succeeds
       And executing "sudo virsh net-list --name" succeeds
       Then stdout should not contain "crc"

   @linux
   Scenario: Running `crc setup` with checks enabled restores destroyed network
       When executing "crc config set skip-check-crc-network false" succeeds
       And executing "crc config set skip-check-crc-network-active false" succeeds
       Then executing "crc setup" succeeds
       And executing "sudo virsh net-list --name" succeeds
       And stdout contains "crc"

   @linux
   Scenario: Running `crc start` without `crc setup` and with checks disabled fails when network destroyed
       # Destroy network again
       When executing "sudo virsh net-undefine crc && sudo virsh net-destroy crc" succeeds
       And executing "sudo virsh net-list --name" succeeds
       Then stdout should not contain "crc"
       # Disable checks
       When executing "crc config set skip-check-crc-network true" succeeds
       And executing "crc config set skip-check-crc-network-active true" succeeds
       # Start CRC
       Then starting CRC with default bundle and default hypervisor fails
       And stdout contains "Network not found: no network with matching name 'crc'"

   @linux
   Scenario: Clean-up
       # Remove the config file
       When removing file "crc.json" from CRC home folder succeeds
       And executing "crc setup" succeeds
       Then stdout should not contain "Skipping above check"
       
