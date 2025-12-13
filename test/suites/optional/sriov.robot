*** Settings ***
Documentation       MicroShift SR-IOV tests

Resource            ../../resources/microshift-process.resource

Suite Setup         Setup Suite
Suite Teardown      Teardown Suite

Test Tags           optional    sriov


*** Variables ***
${TEMPLATE_PATH}    ${CURDIR}/../../assets/sriov/sriov-network-policy-template.yaml


*** Test Cases ***
Create VFs And Verify
    [Documentation]    Deploys sriovnetworknodepolicy and verifies the VF configuration

    VAR    ${cmd_pci_address}=
    ...    oc get sriovnetworknodestate -n sriov-network-operator -o json | jq -r '.items[].status.interfaces[0].pciAddress'
    ${pci_address}=    Run With Kubeconfig    ${cmd_pci_address}

    VAR    ${cmd_device_id}=
    ...    oc get sriovnetworknodestate -n sriov-network-operator -o json | jq -r '.items[].status.interfaces[0].deviceID'
    ${device_id}=    Run With Kubeconfig    ${cmd_device_id}

    ${template_content}=    OperatingSystem.Get File    ${TEMPLATE_PATH}
    ${partial_yaml}=    Replace String    ${template_content}    PLACEHOLDER_PCI_ADDRESS    ${pci_address}
    ${final_yaml}=    Replace String    ${partial_yaml}    PLACEHOLDER_DEVICE_ID    ${device_id}

    Create File    ${OUTPUT DIR}/final-sriov-policy.yaml    ${final_yaml}
    Oc Apply    -f ${OUTPUT_DIR}/final-sriov-policy.yaml -n sriov-network-operator

    Wait Until Resource Exists    sriovnetworknodepolicy    policy-1    sriov-network-operator

    ${stdout}=    Execute Command    sudo ls -l /run/cni/bin
    Should Contain    ${stdout}    sriov

    Wait Until Keyword Succeeds    1min    5s
    ...    Verify VF Count    2

    Wait Until Keyword Succeeds    1min    5s
    ...    Device Plugin Should Be Running

    [Teardown]    Cleanup SR-IOV Policy


*** Keywords ***
Cleanup SR-IOV Policy
    [Documentation]    Deletes the policy
    Run With Kubeconfig    oc delete -f ${OUTPUT_DIR}/final-sriov-policy.yaml -n sriov-network-operator

Verify VF Count
    [Documentation]    Checks if the number of VFs matches the expected count
    [Arguments]    ${expected_vfs}

    ${matching_vfs}=    Run With Kubeconfig
    ...    oc get sriovnetworknodestate -n sriov-network-operator -o json | jq -r '.items[].status.interfaces[].Vfs | length'
    Should Be Equal As Integers    ${matching_vfs}    ${expected_vfs}

    ${allocatable}=    Run With Kubeconfig
    ...    oc get node -o=jsonpath='{ .items[].status.allocatable.openshift\\.io/intelnics }'
    Log    Allocatable openshift.io/intelnics: ${allocatable}
    Should Be Equal As Integers    ${expected_vfs}    ${allocatable}

Device Plugin Should Be Running
    [Documentation]    Checks if the device plugin is running

    ${device_plugin}=    Run With Kubeconfig
    ...    oc get daemonset -n sriov-network-operator sriov-device-plugin -o jsonpath='{ .status.numberReady }'
    Should Be Equal As Integers    ${device_plugin}    1
