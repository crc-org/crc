@story_9pfs @windows
Feature: Verify 9pfs mount works

    Verify 9pfs mount of user's home directory is mounted, accessible
    and that basic file operations work. Only relevant on Windows.

    Scenario: Test mounted home directory functionality
        Given home directory mount exists in VM
        And filesystem is mounted
        Then listing files in mounted home directory should succeed
        And basic file operations in mounted home directory should succeed
        And basic directory operations in mounted home directory should succeed
