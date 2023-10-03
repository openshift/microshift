*** Settings ***
Documentation       Tests related to upgrading MicroShift

Resource            ../../resources/common.resource
Resource            ../../resources/ostree.resource
Library             Collections

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           ostree


*** Variables ***
${USHIFT_HOST}          ${EMPTY}
${USHIFT_USER}          ${EMPTY}

${UPGRADE_REF}          ${EMPTY}

${TEST_NAMESPACE}       default


*** Test Cases ***
Manual Rollback Does Not Restore Data
    [Documentation]    Verifies that using rpm-ostree rollback will not
    ...    restore data automatically: making sure admin needs to perform
    ...    it manually.
    ...
    ...    Test expects that provided upgrade reference is within the same Y-stream,
    ...    e.g. 4.14.0 -> 4.14.1, not 4.14.0 -> 4.15.0.
    ...
    ...    Test will upgrade the system with provided reference, create ConfigMap,
    ...    roll system back to initial deployment, verify that ConfigMap exists,
    ...    create additional ConfigMap, roll system back to upgrade deployment
    ...    and verify that both ConfigMaps exist.

    ${initial_deploy_backup}=    Get Future Backup Name For Current Boot
    ${initial_deploy_id}=    Get Booted Deployment Id

    Deploy Commit Not Expecting A Rollback    ${UPGRADE_REF}
    Backup Should Exist    ${initial_deploy_backup}

    ${upgrade_deploy_backup}=    Get Future Backup Name For Current Boot
    ${upgrade_deploy_id}=    Get Booted Deployment Id

    ConfigMap Create    rollback-test

    Rollback System And Verify    ${upgrade_deploy_backup}    ${initial_deploy_id}
    ${rolled_back_deploy_backup}=    Get Future Backup Name For Current Boot

    ConfigMap Should Exist    rollback-test
    ConfigMap Create    rollback-test-2

    Rollback System And Verify    ${rolled_back_deploy_backup}    ${upgrade_deploy_id}
    ConfigMap Should Exist    rollback-test
    ConfigMap Should Exist    rollback-test-2


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Should Not Be Empty    ${UPGRADE_REF}    UPGRADE_REF variable is required
    Login MicroShift Host
    Setup Kubeconfig
    Wait Until Greenboot Health Check Exited

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
    Wait Until Greenboot Health Check Exited

    Backup Should Exist    ${backup_that_should_exist}
    Current Deployment Should Be    ${expected_deployment_id}

ConfigMap Create
    [Documentation]    Create config map in namespace
    [Arguments]    ${config_map_name}
    Oc Create    configmap -n ${TEST_NAMESPACE} ${config_map_name}

ConfigMap Should Exist
    [Documentation]    Verify config map exists in namespace
    [Arguments]    ${config_map_name}
    Oc Get    configmap    ${TEST_NAMESPACE}    ${config_map_name}
