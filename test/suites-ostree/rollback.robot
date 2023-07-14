*** Settings ***
Documentation       Tests related to upgrading MicroShift

Resource            ../resources/common.resource
Resource            ../resources/ostree.resource
Library             Collections

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           ostree


*** Variables ***
${USHIFT_HOST}              ${EMPTY}
${USHIFT_USER}              ${EMPTY}

${FAKE_NEXT_MINOR_REF}      ${EMPTY}


*** Test Cases ***
Rollback
    [Documentation]    Verifies that rolling back the system can be used to
    ...    go back to older MicroShift version by restoring existing backup.

    ${initial_deploy_backup}=    Get Future Backup Name For Current Boot
    ${initial_deploy_id}=    Get Booted Deployment Id

    Deploy Commit Not Expecting A Rollback    ${FAKE_NEXT_MINOR_REF}
    Backup Should Exist    ${initial_deploy_backup}

    ${upgrade_deploy_backup}=    Get Future Backup Name For Current Boot
    ${upgrade_deploy_id}=    Get Booted Deployment Id

    # Creating a configmap while in "upgrade" deployment.
    # It should not exist after first rollback (back to initial deployment),
    # but should reappear after second rollback (back to "upgrade").
    Run With Kubeconfig    oc create configmap -n default rollback-test

    Rollback System And Verify    ${upgrade_deploy_backup}    ${initial_deploy_id}
    ${rolled_back_deploy_backup}=    Get Future Backup Name For Current Boot

    # Check that `oc get`'s error code is 1.
    Run Keyword And Expect Error    1 != 0
    ...    Oc Get    configmap    default    rollback-test

    Rollback System And Verify    ${rolled_back_deploy_backup}    ${upgrade_deploy_id}
    Oc Get    configmap    default    rollback-test


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Should Not Be Empty    ${FAKE_NEXT_MINOR_REF}    FAKE_NEXT_MINOR_REF variable is required
    Login MicroShift Host
    Setup Kubeconfig
    Wait For Healthy System

Teardown
    [Documentation]    Test suite teardown
    Oc Delete    configmap -n default rollback-test
    Remove Kubeconfig
    Logout MicroShift Host

Rollback System And Verify
    [Documentation]    Rolls back and waits for OSTree system to be healthy
    [Arguments]    ${backup_that_should_exist}    ${expected_deployment_id}

    libostree.Rpm Ostree Rollback
    Reboot MicroShift Host
    Wait For Healthy System

    Backup Should Exist    ${backup_that_should_exist}
    Current Deployment Should Be    ${expected_deployment_id}
