*** Settings ***
Documentation       Tests related to upgrading MicroShift

Resource            ../resources/common.resource
Resource            ../resources/ostree.resource
Library             Collections

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           ostree


*** Variables ***
${USHIFT_HOST}      ${EMPTY}
${USHIFT_USER}      ${EMPTY}

${TARGET_REF}       ${EMPTY}


*** Test Cases ***
First Failed Staged Upgrade Should Restore Backup
    [Documentation]    Performs an upgrade, when first unhealthy a restore is performend

    Wait For Healthy System

    ${initial_backup}=    Get Future Backup Name For Current Boot

    TestAgent.Add Action For Next Deployment    1    fail_greenboot
    Deploy Commit Not Expecting A Rollback    ${TARGET_REF}

    Wait For Healthy System

    Validate Backup    ${initial_backup}


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Should Not Be Empty    ${TARGET_REF}    TARGET_REF variable is required
    Login MicroShift Host
    Setup Kubeconfig

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host

Validate Backup
    [Documentation]    Validate that the desired backup exists and is the only one that exists
    [Arguments]    ${initial_backup}

    Backup Should Exist    ${initial_backup}

    ${stdout}    ${rc}=    Execute Command
    ...    cd ${BACKUP_STORAGE} && ls -d1 rhel-*
    ...    sudo=False    return_rc=True

    Should Be Equal As Strings    ${stdout}    ${initial_backup}
