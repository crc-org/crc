@story_9pfs @windows
Feature: Verify 9pfs mount works

    Verify 9pfs mount of user's home directory is mounted, accessible
    and that basic file operations work. Only relevant on Windows.

    Scenario: Test mounted directory functionality
        Given directory "/mnt/c/user" exists
        And executing "mount" succeeds
        Then stdout should contain "fuse.9pfs"
        Then listing files in directory "/mnt/c/user" should succeed
        And basic file operations in directory "/mnt/c/user" should succeed
        And basic directory operations in directory "/mnt/c/user" should succeed
