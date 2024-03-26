*** Settings ***
Documentation       Tests for Multus and Bridge plugin on MicroShift

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-process.resource

Suite Setup         Setup
Suite Teardown      Teardown


*** Variables ***
${BRIDGE_INTERFACE}                 br-test
${BRIDGE_NAD_YAML}                  ./assets/multus/bridge-nad.yaml
${BRIDGE_POD_YAML}                  ./assets/multus/bridge-pod.yaml
${BRIDGE_POD_NAME}                  test-bridge
${BRIDGE_IP}                        10.10.0.10/24

${PREEXISTING_BRIDGE_INTERFACE}     br-preexisting
${PREEXISTING_BRIDGE_NAD_YAML}      ./assets/multus/bridge-preexisting-nad.yaml
${PREEXISTING_BRIDGE_POD_YAML}      ./assets/multus/bridge-preexisting-pod.yaml
${PREEXISTING_BRIDGE_POD_NAME}      test-bridge-preexisting
${PREEXISTING_BRIDGE_IP}            10.10.1.10/24

${NS}                               default


*** Test Cases ***
Pre-Existing Bridge Interface
    [Documentation]    Test verifies if Bridge CNI plugin will work correctly with pre-existing interface.
    [Setup]    Run Keywords
    ...    Interface Should Not Exist    ${PREEXISTING_BRIDGE_INTERFACE}
    ...    AND
    ...    Create Interface    ${PREEXISTING_BRIDGE_INTERFACE}
    ...    AND
    ...    Create NAD And Pod    ${PREEXISTING_BRIDGE_NAD_YAML}    ${PREEXISTING_BRIDGE_POD_YAML}
    ...    AND
    ...    Named Pod Should Be Ready    ${PREEXISTING_BRIDGE_POD_NAME}    ${NS}
    ...    AND
    ...    Interface Should Exist    ${PREEXISTING_BRIDGE_INTERFACE}
    ...    AND
    ...    Set IP For Host Interface    ${PREEXISTING_BRIDGE_INTERFACE}    ${PREEXISTING_BRIDGE_IP}

    Connect To Pod Over Local Interface    ${PREEXISTING_BRIDGE_POD_NAME}    ${NS}    ${PREEXISTING_BRIDGE_INTERFACE}

    [Teardown]    Cleanup Bridge Test
    ...    ${PREEXISTING_BRIDGE_NAD_YAML}
    ...    ${PREEXISTING_BRIDGE_POD_YAML}
    ...    ${PREEXISTING_BRIDGE_INTERFACE}

No Pre-Existing Bridge Interface
    [Documentation]    Test verifies if Bridge CNI plugin will
    ...    work correctly if there is no pre-existing bridge
    ...    interface - it needs to be created.
    [Setup]    Run Keywords
    ...    Interface Should Not Exist    ${BRIDGE_INTERFACE}
    ...    AND
    ...    Create NAD And Pod    ${BRIDGE_NAD_YAML}    ${BRIDGE_POD_YAML}
    ...    AND
    ...    Named Pod Should Be Ready    ${BRIDGE_POD_NAME}    ${NS}
    ...    AND
    ...    Interface Should Exist    ${BRIDGE_INTERFACE}
    ...    AND
    ...    Set IP For Host Interface    ${BRIDGE_INTERFACE}    ${BRIDGE_IP}

    Connect To Pod Over Local Interface    ${BRIDGE_POD_NAME}    ${NS}    ${BRIDGE_INTERFACE}

    [Teardown]    Cleanup Bridge Test
    ...    ${BRIDGE_NAD_YAML}
    ...    ${BRIDGE_POD_YAML}
    ...    ${BRIDGE_INTERFACE}


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host
    Remove Kubeconfig

Create NAD And Pod
    [Documentation]    Creates provided NetworkAttachmentDefinition and Pod.
    [Arguments]    ${nad}    ${pod}
    Oc Create    -f ${nad}
    Oc Create    -f ${pod}

Cleanup Bridge Test
    [Documentation]    Removes provided NetworkAttachmentDefinition, Pod and network interface to allow for test rerun.
    [Arguments]    ${nad}    ${pod}    ${if}
    Run Keyword And Continue On Failure
    ...    Oc Delete    -f ${pod}
    Run Keyword And Continue On Failure
    ...    Oc Delete    -f ${nad}
    Command Should Work    ip link delete ${if}

Connect To Pod Over Local Interface
    [Documentation]    Makes a HTTP request to 8080 for a given Pod over given interface.
    [Arguments]    ${pod}    ${ns}    ${if}

    ${networks}=    Get And Verify Pod Networks    ${pod}    ${ns}
    ${extra_ip}=    Set Variable    ${networks}[1][ips][0]

    ${stdout}=    Command Should Work    curl -v --interface ${if} ${extra_ip}:8080
    Should Contain    ${stdout}    Hello MicroShift

Interface Should Not Exist
    [Documentation]    Verifies that network interface does not exist.
    [Arguments]    ${if}
    Command Should Fail    ip link show ${if}

Create Interface
    [Documentation]    Creates network interface.
    [Arguments]    ${if}    ${type}=bridge
    Command Should Work    ip link add dev ${if} type ${type}

Interface Should Exist
    [Documentation]    Verifies that interface exists on the host.
    [Arguments]    ${if}
    Command Should Work    ip link show ${if}

Set IP For Host Interface
    [Documentation]    Sets IP address for the interface.
    [Arguments]    ${if}    ${cidr}
    Command Should Work    ip addr add ${cidr} dev ${if}

Get And Verify Pod Networks
    [Documentation]    Obtains interfaces of the Pod from its annotation.
    ...    The annotation is managed by Multus.
    [Arguments]    ${pod}    ${ns}

    ${networks_str}=    Oc Get JsonPath
    ...    pod
    ...    ${ns}
    ...    ${pod}
    ...    .metadata.annotations.k8s\\.v1\\.cni\\.cncf\\.io/network-status
    Should Not Be Empty    ${networks_str}

    ${networks}=    Json Parse    ${networks_str}
    ${n}=    Get Length    ${networks}
    Should Be Equal As Integers    ${n}    2
    Should Be Equal As Strings    ${networks}[0][name]    ovn-kubernetes
    Should Match    ${networks}[1][name]    default/bridge*-conf

    RETURN    ${networks}
