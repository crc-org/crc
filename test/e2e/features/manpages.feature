@story_manpages
Feature: Check generation and cleanup of manpages

  Scenario: Setup CRC Cluster
    Given executing single crc setup command succeeds

  @linux @darwin
  Scenario Outline: verify man pages are accessible after setup
    Then executing "man -P cat 1 <crc-subcommand>" succeeds

    @linux @darwin
    Examples: Man pages to check
      | crc-subcommand       |
      | crc                  |
      | crc-bundle-generate  |
      | crc-config           |
      | crc-start            |
      | crc-bundle           |
      | crc-console          |
      | crc-status           |
      | crc-cleanup          |
      | crc-delete           |
      | crc-stop             |
      | crc-config-get       |
      | crc-ip               |
      | crc-version          |
      | crc-config-set       |
      | crc-oc-env           |
      | crc-config-unset     |
      | crc-podman-env       |
      | crc-config-view      |
      | crc-setup            |

    Scenario: Cleanup CRC Cluster
      When executing crc cleanup command succeeds

    Scenario Outline: verify man pages are NOT accessible after cleanup
      Then executing "man -P cat 1 <crc-subcommand>" fails

      @linux @darwin
      Examples: Man pages to check
        | crc-subcommand       |
        | crc                  |
        | crc-bundle-generate  |
        | crc-config           |
        | crc-start            |
        | crc-bundle           |
        | crc-console          |
        | crc-status           |
        | crc-cleanup          |
        | crc-delete           |
        | crc-stop             |
        | crc-config-get       |
        | crc-ip               |
        | crc-version          |
        | crc-config-set       |
        | crc-oc-env           |
        | crc-config-unset     |
        | crc-podman-env       |
        | crc-config-view      |
        | crc-setup            |