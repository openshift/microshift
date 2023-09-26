*** Settings ***
Documentation       Tests related to upgrading MicroShift

Resource            ../../resources/common.resource
Resource            ../../resources/ostree.resource
Resource            ../../resources/marker-file.resource
Library             Collections

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           ostree


*** Variables ***
${USHIFT_HOST}              ${EMPTY}
${USHIFT_USER}              ${EMPTY}

${UPGRADE_REF}              ${EMPTY}

${UNHEALTHY_UPGRADE_LOG}
...                         Unhealthy deployment stored in health.json matches
...                         rollback deployment - assuming upgrade from unhealthy
...                         system - continuing startup with existing data


*** Test Cases ***
Upgrading Unhealthy System Should Be Successful
    [Documentation]    Verifies that staging new deployment
    ...    over unhealthy system should not result in deleting MicroShift data
    ...    and become healthy on first boot, i.e. MicroShift should not
    ...    refuse to run and cause a reboot.

    Wait For Healthy System
    ${backup}=    Get Future Backup Name For Current Boot

    Mark System As Unhealthy
    Oc Create    configmap -n default unhealthy-upgrade-test
    Create Marker File In Data Dir
    Deploy Commit Not Expecting A Rollback    ${UPGRADE_REF}

    # Because system was unhealthy, backup will not be created
    Backup Should Not Exist    ${backup}
    Expected Boot Count    2
    MicroShift Should Log Upgrade From Unhealthy Happened
    Marker Should Exist In Data Dir
    Oc Get    configmap    default    unhealthy-upgrade-test


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Should Not Be Empty    ${UPGRADE_REF}    UPGRADE_REF variable is required
    Login MicroShift Host
    Setup Kubeconfig

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host
    Remove Kubeconfig

MicroShift Should Log Upgrade From Unhealthy Happened
    [Documentation]    Checks microshift's journal for presence of log message
    ...    indicating microshift recognized upgrade from unhealthy system and
    ...    decided to continue startup.

    ${stdout}    ${rc}=    Execute Command
    ...    journalctl --unit=microshift --boot=0 | grep "${UNHEALTHY_UPGRADE_LOG}"
    ...    sudo=True
    ...    return_rc=True
    Should Be Equal As Integers    0    ${rc}
