*** Settings ***
Documentation       Tests related to auto-recovery functionality

Resource            ../../resources/microshift-host.resource
Resource            ../../resources/microshift-process.resource
Library             Collections
Library             DateTime
Library             ../../resources/libostree.py

Suite Setup         Setup
Suite Teardown      Teardown


*** Test Cases ***
Backups Are Created Successfully With Expected Schema
    [Documentation]    Verify that backups intended for auto-recovery are created successfully
    ...    and they are named with expected schema of DATETIME_DEPLOY-ID for ostree/bootc and DATETIME_MICROSHIFT-VERSION.
    [Setup]    Stop MicroShift

    Command Should Work    microshift backup --auto-recovery /var/lib/microshift-auto-recovery
    ${backup_name}=    Command Should Work    ls /var/lib/microshift-auto-recovery
    Verify Backup Name    ${backup_name}

    [Teardown]    Run Keywords
    ...    Start MicroShift
    ...    AND
    ...    Command Should Work    rm -rf /var/lib/microshift-auto-recovery


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Login MicroShift Host

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host

Verify Backup Name
    [Documentation]    Verify backup's naming schema
    [Arguments]    ${backup_name}
    @{backup}=    Split String    ${backup_name}    _
    Length Should Be    ${backup}    2

    Verify Backup Datetime Against Host Time    ${backup[0]}
    Verify Backup Version Against The System    ${backup[1]}

Verify Backup Datetime Against Host Time
    [Documentation]    Compare given datetime against current time on the VM
    [Arguments]    ${backup_datetime}
    ${backup_dt}=    Convert Date    ${backup_datetime}    datetime
    ${vm_datetime}=    Command Should Work    date +'%Y-%m-%d %H:%M:%S'
    ${vm_dt}=    Convert Date    ${vm_datetime}    datetime
    ${time_diff}=    Subtract Date From Date    ${vm_dt}    ${backup_dt}
    Should Be True    ${time_diff}<60

Verify Backup Version Against The System
    [Documentation]    Verify version part of the backup's name against the system:
    ...    Deployment ID for ostree and bootc; MicroShift X.Y.Z version for RPM systems.
    [Arguments]    ${ver}

    ${is_ostree}=    Is System OSTree
    IF    ${is_ostree}
        ${current_deploy_id}=    Get Booted Deployment Id
        Should Be Equal As Strings    ${current_deploy_id}    ${ver}
    ELSE
        ${microshift_ver}=    MicroShift Version
        Should Be Equal As Strings    ${microshift_ver.major}.${microshift_ver.minor}.${microshift_ver.patch}    ${ver}
    END
