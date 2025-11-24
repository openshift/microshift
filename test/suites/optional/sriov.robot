*** Settings ***
Documentation       MicroShift GitOps tests

Resource            ../../resources/microshift-process.resource

Suite Setup         Setup Suite With Namespace
Suite Teardown      Teardown Suite With Namespace


*** Test Cases ***
Verify GitOps Pods Start Correctly
    [Documentation]    Waits for pods to enter a running state

    Wait Until Keyword Succeeds    2min    10s
    ...    All Pods Should Be Running    sriov-network-operator
