@story_daemon @darwin @linux @windows
Feature: CRC daemon

    User starts the CRC daemon and performs basic operations: start,
    stop, delete, and status.

    Scenario: Prepare
        When setting config property "pull-secret-file" to value "default pull secret" succeeds
        Then "JSON" config file "crc.json" in CRC home folder contains key "pull-secret-file"
        When executing "crc setup --enable-environmental-features" succeeds
        Then stdout should contain "Setup is complete"

    Scenario: 'Version' command
        Then sending "version" command to daemon succeeds

    Scenario Outline: Answer to 'version' command
        Then "JSON" config file "answer.json" in CRC home folder contains key "<key>" with value matching "<value>"

        Examples:

            | key              | value           |
            | Success          | true            |
            | CrcVersion       | \d+\.\d+\.\d+.* |
            | OpenshiftVersion | \d+\.\d+\.\d+.* |
            | CommitSha        | [A-Za-z0-9]{7}  |


    Scenario: 'Start' command via daemon
        Given sending "start" command to daemon succeeds

    Scenario Outline: Answer to 'start' command
        Then "JSON" config file "answer.json" in CRC home folder contains key "<key>" with value matching "<value>"

        Examples:

            | key            | value   |
            | Name           | crc     |
            | KubeletStarted | true    |

    Scenario: 'Status' command via daemon with CRC running
        Given sending "status" command to daemon succeeds

    Scenario Outline: Answer to 'status' command with CRC running
        Then "JSON" config file "answer.json" in CRC home folder contains key "<key>" with value matching "<value>"

        Examples:  

            | key             | value                       |
            | Name            | crc                         |
            | CrcStatus       | Running                     |
            | OpenshiftStatus | Running\ \(v\d+\.\d+\.\d+\) |
            | DiskUse         | \d+                         |
            | DiskSize        | \d+                         |
            | Success         | true                        |

    Scenario: 'Stop' via daemon
        Given sending "stop" command to daemon succeeds

    Scenario Outline: Answer to 'stop' command
       Then "JSON" config file "answer.json" in CRC home folder contains key "<key>" with value matching "<value>"

        Examples:  

            | key     | value |
            | Name    | crc   |
            | Success | true  |
            | State   | 1     |

    Scenario: 'Status' via daemon with CRC stopped
        Given sending "status" command to daemon succeeds

    Scenario Outline: Answer to 'status' command with CRC stopped
       Then "JSON" config file "answer.json" in CRC home folder contains key "<key>" with value matching "<value>"

        Examples:  

            | key             | value   |
            | Name            | crc     |
            | CrcStatus       | Stopped |
            | OpenshiftStatus | Stopped |
            | DiskSize        | 0       |
            | DiskUse         | 0       |
            | Success         | true    |

    Scenario: 'Delete' via daemon
        Given sending "delete" command to daemon succeeds

    Scenario Outline: Answer to 'delete' command
       Then "JSON" config file "answer.json" in CRC home folder contains key "<key>" with value matching "<value>"

        Examples:  

            | key             | value   |
            | Name            | crc     |
            | Success         | true    |


    Scenario: Kill daemon
        When executing "killall crc" succeeds

# NOTE: answer.json remains in ~/.crc after this feature, delete? no need? 
