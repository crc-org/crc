@story_manpages
Feature: Check generation and cleanup of manpages

  @linux @darwin
  Scenario Outline: verify man pages are accessible after setup
    Given executing single crc setup command succeeds
    And executing "export MANPATH=$HOME/.local/share/man:$MANPATH" succeeds
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

  Scenario Outline: verify man pages are NOT accessible after cleanup
    Given executing crc cleanup command succeeds
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