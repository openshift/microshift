*** Settings ***
Documentation       Tests related to upgrading MicroShift

Resource            ../../resources/common.resource
Resource            ../../resources/ostree.resource
Library             Collections

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           ostree


*** Variables ***
${FAILING_REF}          ${EMPTY}
${REASON}               ${EMPTY}
${BOOTC_REGISTRY}       ${EMPTY}


*** Test Cases ***
New Deployment Is Consistently Unhealthy And Rolls Back
    [Documentation]    Verifies that whether the new deployment
    ...    cannot create a backup or is simply unhealthy,
    ...    system will rollback and become healthy.

    Wait Until Greenboot Health Check Exited
    ${backup}=    Get Future Backup Name For Current Boot

    Oc Create    configmap -n default unhealthy-upgrade-test

    TestAgent.Add Action For Next Deployment    every    ${REASON}
    Deploy Commit Expecting A Rollback
    ...    ${FAILING_REF}
    ...    True
    ...    ${BOOTC_REGISTRY}

    Wait For MicroShift Service
    Backup Should Exist    ${backup}
    Expected Boot Count    5
    Oc Get    configmap    default    unhealthy-upgrade-test


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Should Not Be Empty    ${FAILING_REF}    FAILING_REF variable is required
    Should Not Be Empty    ${REASON}    REASON variable is required
    Login MicroShift Host
    Setup Kubeconfig

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host
    Remove Kubeconfig
