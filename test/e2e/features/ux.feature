@ux 
Feature: UX Test

    Tets CRC usage with ux components. 
        * Handle CRC install/ uninstall operations based on installer distribution
        * Manage CRC through the tray
    
    @install @darwin
    Scenario: Install CRC 
        Given an environment where CRC is not installed
        When install CRC from installer
        Then CRC is installed
		
	@install @windows
    Scenario: Install CRC 
        Given an environment where CRC is not installed
        When install CRC from installer
        Then reboot is required
		
	@darwin @windows
    Scenario: Install tray  
        When install CRC tray
        Then tray should be installed
        And tray icon should be accessible

	@darwin @windows
    Scenario: Start the cluster
        Given fresh tray installation   
        When start the cluster from the tray
        And set the pull secret
        Then cluster should be running
        And tray should show cluster as running
        And user should get notified with cluster state as running

	@darwin @windows
    Scenario Outline: Connect to the cluster
        Given a running cluster   
        When using copied oc login command for <ocp-user>  
        Then user is connected to the cluster as <ocp-user> 
        # TODO notifications inconsistent https://github.com/code-ready/tray-macos/issues/84
        # And user should get notified with command copied

    Examples:
            | ocp-user   |
            | kubeadmin |
            | developer |

	@darwin @windows
    Scenario: Stop the cluster
        Given a running cluster   
        When stop the cluster from the tray 
        Then cluster should be stopped
        And tray should show cluster as stopped
        And user should get notified with cluster state as stopped

	@darwin @windows
    Scenario: Restart the cluster
        Given a stopped cluster   
        When start the cluster from the tray 
        Then cluster should be running
        And tray should show cluster as running
        And user should get notified with cluster state as running
