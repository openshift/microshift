*** Settings ***
Documentation       MicroShift SR-IOV tests

Resource            ../../resources/microshift-process.resource

Suite Setup         Setup Suite
Suite Teardown      Teardown Suite


*** Variables ***
${TEMPLATE_PATH}    ${CURDIR}/../../assets/sriov/sriov-network-policy-template.yaml


*** Test Cases ***
Create VFs And Verify
    [Documentation]    Deploys sriovnetworknodepolicy and verifies the VF configuration

    VAR    ${cmd}=
    ...    oc get sriovnetworknodestate -o yaml -n sriov-network-operator -o json | jq -r '.items[].status.interfaces[0].pciAddress'
    ${pci_address}=    Run With Kubeconfig    ${cmd}

    ${template_content}=    OperatingSystem.Get File    ${TEMPLATE_PATH}
    ${final_yaml}=    Replace String    ${template_content}    PLACEHOLDER_PCI_ADDRESS    ${pci_address}

    Create File    ${OUTPUT DIR}/final-sriov-policy.yaml    ${final_yaml}

    Oc Apply    -f ${OUTPUT_DIR}/final-sriov-policy.yaml -n sriov-network-operator

    Wait Until Resource Exists    sriovnetworknodepolicy    policy-1    sriov-network-operator

    Wait Until Keyword Succeeds    2min    10s
    ...    Verify VF Count    2

    Wait Until Keyword Succeeds    2min    10s
    ...    All Pods Should Be Running    sriov-network-operator

    [Teardown]    Cleanup SR-IOV Policy


*** Keywords ***
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
