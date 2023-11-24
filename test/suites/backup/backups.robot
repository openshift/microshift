*** Settings ***
Documentation       Tests related to MicroShift backup functionality on OSTree-based systems

Resource            ../../resources/common.resource
Resource            ../../resources/ostree.resource
Resource            ../../resources/ostree-data.resource
Library             Collections
Library             ../../resources/libostree.py

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           ostree


*** Variables ***
${USHIFT_HOST}          ${EMPTY}
${USHIFT_USER}          ${EMPTY}

${FAKE_DIR_REG}         *_fake*
${UNKOWN_DIR_REG}       unknown*


*** Test Cases ***
Rebooting Healthy System Should Result In Pruning Old Backups
    [Documentation]    Check that MicroShift backup logic prunes old backups

    Validate Fake Backups Have Been Deleted

Rebooting Healthy System Should Not Delete Unknown Directories
    [Documentation]    Check that MicroShift backup logic does not delete unknown directories

    Validate Unknown Directories Exist

Restarting MicroShift Should Not Result In Backup Creation
    [Documentation]    Check that backup won't be created on MicroShift restart
    ...    (which would block creating right backup after reboot)

    ${deploy_id}=    Get Booted Deployment ID
    ${boot_id}=    Get Current Boot Id
    Restart MicroShift
    Backup Should Not Exist    ${deploy_id}    ${boot_id}


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

Create Unknown Directories
    [Documentation]    Create unknown fake folders that the pruning logic should not touch
    Create Fake Backups    2    True

Validate Unknown Directories Exist
    [Documentation]    Make sure the unknown fake backups exist
    List Backups With Expected RC    0    ${UNKOWN_DIR_REG}

Delete Unknown Directories
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

    Create Fake Backup Directories
    Validate Fake Backups Exist

    Create Unknown Directories
    Validate Unknown Directories Exist

    Reboot MicroShift Host
    Wait Until Greenboot Health Check Exited

Teardown
    [Documentation]    Test suite teardown
    Delete Unknown Directories
    Logout MicroShift Host
