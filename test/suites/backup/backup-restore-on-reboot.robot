*** Settings ***
Documentation       Tests related to MicroShift upgradeability on OSTree-based systems

Resource            ../../resources/common.resource
Resource            ../../resources/ostree.resource
Library             Collections

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           ostree


*** Variables ***
${USHIFT_HOST}      ${EMPTY}
${USHIFT_USER}      ${EMPTY}


*** Test Cases ***
Rebooting Healthy System Should Result In Data Backup
    [Documentation]    Check if rebooting healthy system will result in backing up of MicroShift data

    Wait Until Greenboot Health Check Exited
    System Should Be Healthy
    Remove Existing Backup For Current Deployment
    ${future_backup}=    Get Future Backup Name For Current Boot

    Reboot MicroShift Host
    Wait For MicroShift Service

    Backup Should Exist    ${future_backup}

Rebooting Unhealthy System Should Result In Restoring Data From A Backup
    [Documentation]    Check if rebooting unhealthy system will result
    ...    restoring MicroShift data from a backup

    Wait Until Greenboot Health Check Exited    # we don't want greenboot to overwrite health.json
    Remove Existing Backup For Current Deployment

    ${backup_name}=    Make Masquerading Backup

    Mark System As Unhealthy    # to trigger restore after reboot
    Reboot MicroShift Host
    Wait For MicroShift Service

    Marker Should Exist In Data Dir
    Remove Marker From Data Dir


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host

Make Masquerading Backup
    [Documentation]    Stops MicroShift and creates manual backup that
    ...    masquerades as automated one (by prefixing the backup name with deployment ID)
    Systemctl    stop    microshift.service

    ${deploy_id}=    Get Booted Deployment ID
    ${backup_name}=    Set Variable    ${deploy_id}_manual

    Create Backup    ${backup_name}    ${TRUE}

    RETURN    ${backup_name}
