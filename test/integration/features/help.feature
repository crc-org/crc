Feature: Help command

  Scenario: Alex checks help command for CRC
     When executing "crc help" succeeds
     Then stdout should contain
     """
     Available Commands:
       help        Help about any command
       start       start cluster
       stop        stop cluster
     """
