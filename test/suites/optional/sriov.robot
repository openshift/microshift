*** Settings ***
Documentation       MicroShift SR-IOV tests

Resource            ../../resources/microshift-process.resource

Suite Setup         Setup Suite With Namespace
Suite Teardown      Teardown Suite With Namespace


*** Variables ***
${TEMPLATE_PATH}    ${CURDIR}/../../assets/sriov/sriov-network-policy-template.yaml
${VENDOR_ID}        8086
${DEVICE_ID}        10c9


*** Test Cases ***
Verify SR-IOV Pods Start Correctly
    [Documentation]    Waits for pods to enter a running state

    Wait Until Keyword Succeeds    2min    10s
    ...    All Pods Should Be Running    sriov-network-operator

Create VFs And Verify
    [Documentation]    Deploys sriovnetworknodepolicy and verifies the VF configuration

    ${pci_address}=    Get PCI Address For Device    ${VENDOR_ID}    ${DEVICE_ID}

    ${template_content}=    OperatingSystem.Get File    ${TEMPLATE_PATH}
    ${final_yaml}=    Replace String    ${template_content}    PLACEHOLDER_PCI_ADDRESS    ${pci_address}

    Create File    ${OUTPUT DIR}/final-sriov-policy.yaml    ${final_yaml}

    Oc Apply    -f ${OUTPUT_DIR}/final-sriov-policy.yaml -n sriov-network-operator

    Oc Get    sriovnetworknodepolicy    sriov-network-operator    policy-1

    Wait Until Keyword Succeeds    2min    10s
    ...    Verify VF Count    2

    Wait Until Keyword Succeeds    2min    10s
    ...    All Pods Should Be Running    sriov-network-operator

    [Teardown]    Cleanup SR-IOV Policy


*** Keywords ***
Get PCI Address For Device
    [Documentation]    Uses lspci to find the first device matching vendor:device
    [Arguments]    ${vendor}    ${device}

    VAR    ${cmd}=    lspci -D -nn | grep "${vendor}:${device}" | head -n 1 | awk '{print $1}'
    ${address}=    Execute Command    ${cmd}    sudo=True

    Should Not Be Empty    ${address}    Could not find any PCI device matching ${vendor}:${device}
    RETURN    ${address}

Cleanup SR-IOV Policy
    [Documentation]    Deletes the policy

    Oc Delete    sriovnetworknodepolicy policy-1 -n sriov-network-operator

Verify VF Count
    [Documentation]    Checks if the number of VFs matches the expected count
    [Arguments]    ${expected_vfs}

    ${node_name}=    Oc Get JsonPath    nodes    ${EMPTY}    ${EMPTY}    .items[0].metadata.name

    ${current_vfs}=    Oc Get JsonPath
    ...    node
    ...    ${EMPTY}
    ...    ${node_name}
    ...    .status.allocatable.openshift\\.io/intelnics

    Should Be Equal As Integers    ${expected_vfs}    ${current_vfs}
