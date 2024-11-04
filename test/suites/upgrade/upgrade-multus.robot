*** Settings ***
Documentation       Tests related to upgrading MicroShift

Resource            ../../resources/common.resource
Resource            ../../resources/multus.resource
Resource            ../../resources/ostree.resource

Suite Setup         Setup
Suite Teardown      Teardown


*** Variables ***
${TARGET_REF}           ${EMPTY}
${BOOTC_REGISTRY}       ${EMPTY}

${BRIDGE_INTERFACE}     br-test
${BRIDGE_NAD_YAML}      ./assets/multus/bridge-nad.yaml
${BRIDGE_POD_YAML}      ./assets/multus/bridge-pod.yaml
${BRIDGE_POD_NAME}      test-bridge
${BRIDGE_IP}            10.10.0.10/24


*** Test Cases ***
Upgrade With Multus Workload
    [Documentation]    Performs an upgrade and verifies if Pod with extra network interface continues to work.

    Wait Until Greenboot Health Check Exited

    Create NAD And Pod    ${BRIDGE_NAD_YAML}    ${BRIDGE_POD_YAML}
    Named Pod Should Be Ready    ${BRIDGE_POD_NAME}    ${NAMESPACE}
    Set IP For Host Interface    ${BRIDGE_INTERFACE}    ${BRIDGE_IP}
    Connect To Pod Over Local Interface    ${BRIDGE_POD_NAME}    ${NAMESPACE}    ${BRIDGE_INTERFACE}

    Deploy Commit Not Expecting A Rollback
    ...    ${TARGET_REF}
    ...    ${TRUE}
    ...    ${BOOTC_REGISTRY}

    Named Pod Should Be Ready    ${BRIDGE_POD_NAME}    ${NAMESPACE}
    Set IP For Host Interface    ${BRIDGE_INTERFACE}    ${BRIDGE_IP}
    Connect To Pod Over Local Interface    ${BRIDGE_POD_NAME}    ${NAMESPACE}    ${BRIDGE_INTERFACE}

    [Teardown]    Remove NAD And Pod    ${BRIDGE_NAD_YAML}    ${BRIDGE_POD_YAML}


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Should Not Be Empty    ${TARGET_REF}    TARGET_REF variable is required
    Setup Suite With Namespace

Teardown
    [Documentation]    Test suite teardown
    Teardown Suite With Namespace
