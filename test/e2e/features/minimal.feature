@minimal @darwin @linux @windows
Feature: Minimal user story

    User starts CRC cluster and checks status.

    @cleanup
    Scenario Outline: Start OpenShift cluster:
        Given setting config property "preset" to value "<preset-value>" succeeds
        And executing single crc setup command succeeds
        And starting CRC with default bundle succeeds
        Then checking that CRC is running

            @podman-preset
            Scenarios:
            | preset-value |
            | podman       |

            @microshift-preset
            Scenarios:
            | preset-value |
            | microshift   |

            @openshift-preset
            Scenarios:
            | preset-value |
            | openshift    |
