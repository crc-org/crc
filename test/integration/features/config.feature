@config
Feature: Test configuration settings

    User checks whether CRC `config set` command works as expected 
    in conjunction with `crc setup` and `crc start` commands.

# SETTINGS

    Scenario Outline: CRC config checks (settings)
        When setting config property "<property>" to value "<value1>" succeeds
        Then file "crc.json" exists in CRC home folder
        And "JSON" config file "crc.json" in CRC home folder contains key "<property>" with value matching "<value1>"
        And setting config property "<property>" to value "<value2>" fails
        When unsetting config property "<property>" succeeds
        Then "JSON" config file "crc.json" in CRC home folder does not contain key "<property>"
        
        # always return to default values
        @darwin
        Examples: Config settings on Mac 
            | property         |    value1 |            value2 |
            | cpus             |         4 |                 3 |
            | memory           |      9216 |              4096 |
            | nameserver       | 120.0.0.1 |   999.999.999.999 |
            | pull-secret-file |      /etc | /nonexistent-file |

        @linux
        Examples: Config settings on Linux
            | property         |    value1 |            value2 |
            | cpus             |         4 |                 3 |
            | memory           |      9216 |              4096 |
            | nameserver       | 120.0.0.1 |   999.999.999.999 |
            | pull-secret-file |      /etc | /nonexistent-file |

        @windows
        Examples: Config settings on Windows
            | property         |    value1 |            value2 |
            | cpus             |         4 |                 3 |
            | memory           |      9216 |              4096 |
            | nameserver       | 120.0.0.1 |   999.999.999.999 |
            | pull-secret-file |    /Users | /nonexistent-file |

    @linux @darwin @windows
    Scenario: CRC config checks (bundle version)
        Given executing "crc setup" succeeds
        When setting config property "bundle" to value "current bundle" succeeds
        And "JSON" config file "crc.json" in CRC home folder contains key "bundle" with value matching "current bundle"
        And setting config property "bundle" to value "/path/to/nonexistent/bundle/crc_hypervisor_version.tar.xz" fails
        When unsetting config property "bundle" succeeds
        Then "JSON" config file "crc.json" in CRC home folder does not contain key "bundle"


    @linux @darwin @windows
    Scenario: CRC config checks (update check)
        When setting config property "disable-update-check" to value "true" succeeds
        Then  "JSON" config file "crc.json" in CRC home folder contains key "disable-update-check" with value matching "true"
        When unsetting config property "disable-update-check" succeeds
        Then "JSON" config file "crc.json" in CRC home folder does not contain key "disable-update-check"

# WARNINGS

    Scenario Outline: CRC config checks (warnings)
        When setting config property "<property>" to value "<value1>" succeeds
        Then "JSON" config file "crc.json" in CRC home folder contains key "<property>" with value matching "<value1>"
        When setting config property "<property>" to value "<value2>" succeeds
        Then "JSON" config file "crc.json" in CRC home folder contains key "<property>" with value matching "<value2>"
        When unsetting config property "<property>" succeeds
        Then "JSON" config file "crc.json" in CRC home folder does not contain key "<property>"

        @darwin
        Examples: Config warnings on Mac
            | property                             | value1 | value2 |
            | warn-check-bundle-extracted          | true   | false  |
            | warn-check-hosts-file-permissions    | true   | false  |
            | warn-check-hyperkit-driver           | true   | false  |
            | warn-check-hyperkit-installed        | true   | false  |
            | warn-check-resolver-file-permissions | true   | false  |
            | warn-check-root-user                 | true   | false  |

        @linux
        Examples: Config warnings on Linux
            | property                             | value1 | value2 |
            | warn-check-bundle-extracted          | true   | false  |
            | warn-check-crc-dnsmasq-file          | true   | false  |
            | warn-check-crc-network               | true   | false  |
            | warn-check-crc-network-active        | true   | false  |
            | warn-check-kvm-enabled               | true   | false  |
            | warn-check-libvirt-driver            | true   | false  |
            | warn-check-libvirt-installed         | true   | false  |
            | warn-check-libvirt-running           | true   | false  |
            | warn-check-libvirt-version           | true   | false  |
            | warn-check-network-manager-config    | true   | false  |
            | warn-check-network-manager-installed | true   | false  |
            | warn-check-network-manager-running   | true   | false  |
            | warn-check-root-user                 | true   | false  |
            | warn-check-user-in-libvirt-group     | true   | false  |
            | warn-check-virt-enabled              | true   | false  |

        @windows
        Examples: Config warnings on Windows
            | property                        | value1 | value2 |
            | warn-check-administrator-user   | true   | false  |
            | warn-check-bundle-extracted     | true   | false  |
            | warn-check-hyperv-installed     | true   | false  |
            | warn-check-hyperv-switch        | true   | false  |
            | warn-check-user-in-hyperv-group | true   | false  |
            | warn-check-windows-version      | true   | false  |

# SKIP

    Scenario Outline: CRC config checks (skips)
        When setting config property "<property>" to value "<value1>" succeeds
        And "JSON" config file "crc.json" in CRC home folder contains key "<property>" with value matching "<value1>"
        When executing "crc setup" succeeds
        Then stderr should contain "Skipping above check ..."
        When setting config property "<property>" to value "<value2>" succeeds
        Then "JSON" config file "crc.json" in CRC home folder contains key "<property>" with value matching "<value2>"
        When executing "crc setup" succeeds
        Then stderr should not contain "Skipping above check ..."
        When unsetting config property "<property>" succeeds
        Then "JSON" config file "crc.json" in CRC home folder does not contain key "<property>"
        
        @darwin
        Examples:
            | property                             | value1 | value2 |
            | skip-check-bundle-extracted          | true   | false  |
            | skip-check-hosts-file-permissions    | true   | false  |
            | skip-check-hyperkit-driver           | true   | false  |
            | skip-check-hyperkit-installed        | true   | false  |
            | skip-check-resolver-file-permissions | true   | false  |
            | skip-check-root-user                 | true   | false  |

        @linux
        Examples:
            | property                             | value1 | value2 |
            | skip-check-bundle-extracted          | true   | false  |
            | skip-check-crc-dnsmasq-file          | true   | false  |
            | skip-check-crc-network               | true   | false  |
            | skip-check-crc-network-active        | true   | false  |
            | skip-check-kvm-enabled               | true   | false  |
            | skip-check-libvirt-driver            | true   | false  |
            | skip-check-libvirt-installed         | true   | false  |
            | skip-check-libvirt-running           | true   | false  |
            | skip-check-libvirt-version           | true   | false  |
            | skip-check-network-manager-config    | true   | false  |
            | skip-check-network-manager-installed | true   | false  |
            | skip-check-network-manager-running   | true   | false  |
            | skip-check-root-user                 | true   | false  |
            | skip-check-user-in-libvirt-group     | true   | false  |
            | skip-check-virt-enabled              | true   | false  |

        @windows
        Examples:
            | property                        | value1 | value2 |
            | skip-check-administrator-user   | true   | false  |
            | skip-check-bundle-extracted     | true   | false  |
            | skip-check-hyperv-installed     | true   | false  |
            | skip-check-hyperv-switch        | true   | false  |
            | skip-check-user-in-hyperv-group | true   | false  |
            | skip-check-windows-version      | true   | false  |

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
       Then starting CRC with default bundle fails
       And stderr contains "Network not found: no network with matching name 'crc'"

   @linux
   Scenario: Clean-up
       # Remove the config file
       When removing file "crc.json" from CRC home folder succeeds
       And executing "crc setup" succeeds
       Then stderr should not contain "Skipping above check"
       
