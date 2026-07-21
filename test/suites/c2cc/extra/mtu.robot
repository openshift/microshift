*** Settings ***
Documentation       MTU boundary tests for C2CC at jumbo frame size (9000 MTU).
...                 Pod MTU 9000 is set via kickstart ovn.yaml (OVN-K bakes the
...                 MTU into its database at initial creation). Validates DF-bit
...                 acceptance and rejection at the boundary, large TCP transfers,
...                 and full cross-cluster connectivity.
...                 When IPsec is configured (IPSEC variable set to true), also
...                 verifies ESP encapsulation and tunnel reestablishment.
...                 Used by both c2cc-mtu and c2cc-ipsec-mtu scenarios.

Resource            ../../../resources/microshift-process.resource
Resource            ../../../resources/kubeconfig.resource
Resource            ../../../resources/oc.resource
Resource            ../../../resources/c2cc.resource
Resource            ../../../resources/ipsec.resource

Suite Setup         C2CC Suite Setup    deploy_workloads=${TRUE}
Suite Teardown      C2CC Suite Teardown    cleanup_workloads=${TRUE}

Test Tags           c2cc    mtu


*** Variables ***
${IPSEC}    false


*** Test Cases ***
C2CC At Jumbo MTU 9000
    [Documentation]    Validate MTU boundary behavior across all cluster pairs.
    ...    Pod MTU 9000 is set via kickstart ovn.yaml. When IPSEC=true,
    ...    also verifies ESP encapsulation.
    Verify Pod MTU On All Clusters    9000
    IF    "${IPSEC}" == "true"
        Verify ESP Encapsulation On All Clusters
    ELSE
        Verify Full C2CC Connectivity
    END
    Ping DF Bit Between All Clusters    64
    # 8952 + 48 (IPv6 headers) = 9000 = pod MTU → passes
    Ping DF Bit Between All Clusters    8952
    # 8973 + 28 (IPv4 headers) = 9001 > 9000 pod MTU → rejected
    Ping DF Bit Should Fail Between All Clusters    8973
    Large Payload Between All Clusters    65536
