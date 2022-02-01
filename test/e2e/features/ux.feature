@ux 
Feature: UX Test

    Tets CRC usage with ux components. 
        * Handle app install/ uninstall operations based on installer distribution
        * Manage CRC through the app
    
    @install @darwin
    Scenario: Install app 
        Given an environment where CRC app is not installed
        When install CRC app from installer
        Then CRC app is installed
		
	@install @windows
    Scenario: Install app 
        Given an environment where CRC app is not installed
        When install CRC app from installer
        Then reboot is required

	@darwin @windows
    Scenario: App onboarding  
        Given a fresh CRC app installation
        When onboarding CRC app setting <preset-id> preset
        Then CRC app should be running
        And CRC app should be accessible
        And CRC app should be ready to start a environment for <preset-id> preset

    Examples:
            | preset-id |
            | openshift |
            | podman    |

	@darwin @windows
    Scenario: Start instance of <preset-id> preset
        Given crc app configured to run then <preset-id> preset   
        When click start button from the app
        Then get notification about the starting process 
        And the <preset-id> instance should be running
        And app shows running as the state for the instance

    Examples:
            | preset-id |
            | openshift |
            | podman    |

	@darwin @windows
    Scenario Outline: Connect to the cluster
        Given a running instance for openshift preset    
        When using copied oc login command for <ocp-user>  
        Then user is connected to the cluster as <ocp-user>

    Examples:
            | ocp-user  |
            | kubeadmin |
            | developer |

    @darwin @windows
    Scenario: Stop instance of <preset-id> preset
        Given running instance for <preset-id> preset
        When click stop button from the app
        Then get notification about the stopping process 
        And the <preset-id> instance should be stopped
        And app should show instance as stopped

    Examples:
            | preset-id |
            | openshift |
            | podman    |

	@darwin @windows
    Scenario: Restart instance of <preset-id> preset
        Given stopped instance for <preset-id> preset
        When click start button from the app
        Then get notification about the starting process 
        And the <preset-id> instance should be running
        And app shows running as the state for the instance

    Examples:
            | preset-id |
            | openshift |
            | podman    |
