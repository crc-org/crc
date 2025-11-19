@story_9pfs @windows
Feature: Verify 9pfs mount works

    Verify 9pfs mount of user's home directory is mounted, accessible
    and that basic file operations work. Only relevant on Windows.

    Scenario: Test mounted directory functionality
        Given directory "/mnt/c/Users/core" exists in VM
        And filesystem is mounted
        Then listing files in directory "/mnt/c/Users/core" should succeed
        And basic file operations in directory "/mnt/c/Users/core" should succeed
        And basic directory operations in directory "/mnt/c/Users/core" should succeed
