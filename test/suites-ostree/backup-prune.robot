*** Settings ***
Documentation       Tests related to MicroShift backup functionality on OSTree-based systems

Resource            ../resources/common.resource
Resource            ../resources/ostree-data.resource
Library             Collections

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           ostree


*** Variables ***
${USHIFT_HOST}          ${EMPTY}
${USHIFT_USER}          ${EMPTY}

${FAKE_DIR_REG}         *_fake*
${UNKOWN_DIR_REG}       unknown_fake*


*** Test Cases ***
Rebooting Healthy System Should Result In Pruning Old Backups And Keeping Unknown Directories
    [Documentation]    Check that MicroShift backup logic prunes old backups and keeps unknown directories

    Create Fake Backup Directories
    Validate Fake Backups Exist

    Create Unknown Fake Backup Directories
    Validate Unknown Fake Backups Exist

    Reboot MicroShift Host
    Wait For Healthy System

    Validate Fake Backups Have Been Deleted
    Validate Unknown Fake Backups Exist


*** Keywords ***
Create Fake Backup Directories
    [Documentation]    Create fake backup folders to be pruned
    Create Fake Backups    5

Validate Fake Backups Exist
    [Documentation]    Make sure the fake backups exist
    List Backups With Expected RC    0    ${FAKE_DIR_REG}

Validate Fake Backups Have Been Deleted
    [Documentation]    Listing the fake backups should return an error code of 2 if no longer exists
    List Backups With Expected RC    2    ${FAKE_DIR_REG}

Create Unknown Fake Backup Directories
    [Documentation]    Create unknown fake backup folders
    Create Fake Backups    2    True

Validate Unknown Fake Backups Exist
    [Documentation]    Make sure the unknown fake backups exist
    List Backups With Expected RC    0    ${UNKOWN_DIR_REG}

Delete Unkown Fake Backups
    [Documentation]    Remove unknown directories

    ${stdout}    ${rc}=    Execute Command
    ...    rm -rf ${BACKUP_STORAGE}/${UNKOWN_DIR_REG}
    ...    sudo=True    return_rc=True
    Should Be Equal As Integers    0    ${rc}

List Backups With Expected RC
    [Documentation]    Lists backup directory content and checks against a desirec RC
    [Arguments]    ${expected_exit_code}    ${ls_regex}
    ${stdout}    ${rc}=    Execute Command
    ...    ls -lah ${BACKUP_STORAGE}/${ls_regex}
    ...    sudo=True    return_rc=True
    Should Be Equal As Integers    ${expected_exit_code}    ${rc}

Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host

Teardown
    [Documentation]    Test suite teardown
    Delete Unkown Fake Backups
    Logout MicroShift Host
