@story_manpages
Feature: Check generation and cleanup of manpages

  @linux @darwin @cleanup
  Scenario: verify man pages are generated after crc setup and deleted on cleanup
    When executing single crc setup command succeeds
    Then accessing crc man pages succeeds

    When executing crc cleanup command succeeds
    Then accessing crc man pages fails