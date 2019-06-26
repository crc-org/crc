Feature:
    Health check of the cluster after CRC start.

    Scenario Outline: Set-up
        Given executing "crc setup" succeeds
        When starting CRC with default bundle and hypervisor "<vm-driver>" succeeds
        Then stdout should contain "CodeReady Containers instance is running"
        And executing "eval $(crc oc-env)" succeeds
        And executing "oc login --insecure-skip-tls-verify -u kubeadmin -p ehbg7-zu5i6-JKt7V-PvJsm https://api.crc.testing:6443" succeeds

        @darwin
        Examples:
        | vm-driver  |
        | virtualbox |

        @linux
        Examples:
        | vm-driver |
        | libvirt   |

    # Nodes
    @darwin @linux @windows
    Scenario: Checking master-worker node
        When executing "oc get nodes" succeeds
        Then stdout should contain "Ready"
        And stdout should not contain "Not Ready"
       

    # Cluster operators
    @darwin @linux @windows
    Scenario Outline: Checking cluster operators available
        Given check at most "2" times with delay of "60s" that cluster operator "<name>" is available

        Examples:
            | name                               |
            | authentication                     |
            | cloud-credential                   |
            | cluster-autoscaler                 |
            | console                            |
            | dns                                |
            | image-registry                     |
            | ingress                            |
            | kube-apiserver                     |
            | kube-controller-manager            |
            | kube-scheduler                     |
            | machine-api                        |
            | marketplace                        |
            | monitoring                         |
            | node-tuning                        |
            | openshift-apiserver                |
            | openshift-controller-manager       |
            | openshift-samples                  |
            | operator-lifecycle-manager         |
            | operator-lifecycle-manager-catalog |
            | service-ca                         |
            | service-catalog-apiserver          |
            | service-catalog-controller-manager |
            | storage                            |

    # machine-config is special: never available
    @darwin @linux @windows
    Scenario Outline: Checking cluster operators not available
        Given check at most "2" times with delay of "60s" that cluster operator "<name>" is not available

        Examples:
            | name           |
            | machine-config |

    # machine-config is special: forever progressing
    @darwin @linux @windows
    Scenario Outline: Checking cluster operators not progressing
        Given check at most "5" times with delay of "60s" that cluster operator "<name>" is not progressing

        Examples:
            | name                               |
            | authentication                     |
            | cloud-credential                   |
            | cluster-autoscaler                 |
            | console                            |
            | dns                                |
            | image-registry                     |
            | ingress                            |
            | kube-apiserver                     |
            | kube-controller-manager            |
            | kube-scheduler                     |
            | machine-api                        |
            | marketplace                        |
            | monitoring                         |
            | node-tuning                        |
            | openshift-apiserver                |
            | openshift-controller-manager       |
            | openshift-samples                  |
            | operator-lifecycle-manager         |
            | operator-lifecycle-manager-catalog |
            | service-ca                         |
            | service-catalog-apiserver          |
            | service-catalog-controller-manager |
            | storage                            |

    @darwin @linux @windows
    Scenario: CRC stop
        When executing "crc stop -f" succeeds
        Then stdout should contain "CodeReady Containers instance stopped"
       
    @darwin @linux @windows
    Scenario: CRC delete
        When executing "crc delete" succeeds
        Then stdout should contain "CodeReady Containers instance deleted"
