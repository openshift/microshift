*** Settings ***
Documentation       Tests related to auto-recovery functionality

Resource            ../../resources/microshift-host.resource
Resource            ../../resources/microshift-process.resource
Library             Collections
Library             DateTime
Library             ../../resources/libostree.py

Suite Setup         Setup
Suite Teardown      Teardown


*** Variables ***
${WORKDIR}      /var/lib/microshift-auto-recovery/


*** Test Cases ***
Backups Are Created Successfully With Expected Schema
    [Documentation]    Verify that backups intended for auto-recovery are created successfully
    ...    and they are named with expected schema of DATETIME_DEPLOY-ID for ostree/bootc and DATETIME_MICROSHIFT-VERSION.

    ${backup_path}=    Command Should Work    microshift backup --auto-recovery ${WORKDIR}
    ${backup_name}=    Remove String    ${backup_path}    ${WORKDIR}
    Verify Backup Name    ${backup_name}

    [Teardown]    Command Should Work    rm -rf ${WORKDIR}

Restore Fails When There Are No Backups
    [Documentation]    Restoring fail if the storage contains no backups at all.

    Command Should Fail    microshift restore --auto-recovery ${WORKDIR}

Restore Fails When There Are No Suitable Backups
    [Documentation]    Create 3 backups, but change their versions to other
    ...    MicroShift version/Deployment ID, so even there are backups present,
    ...    none is suitable for restoring.

    FOR    ${counter}    IN RANGE    1    4
        ${backup_path}=    Command Should Work    microshift backup --auto-recovery ${WORKDIR}
        Command Should Work    bash -c "echo ${counter} > ${backup_path}/marker"
        ${new_path}=    Set Variable    ${{ "${backup_path}[:-1]" + str(int("${backup_path}[-1]")+1) }}
        Command Should Work    sudo mv ${backup_path} ${new_path}
        Sleep    2s
    END
    Command Should Fail    microshift restore --auto-recovery ${WORKDIR}

    [Teardown]    Command Should Work    rm -rf ${WORKDIR}

Restore Selects Right Backup
    [Documentation]    Test creates 3 backups and changes the last one's version.
    ...    Restore command should use the second created backup as it's the most recent
    ...    among backups that match the version.

    ${last_backup}=    Create Backups    3    ${TRUE}

    # Rename the last backup to different deployment ID. When restoring it should be skipped.
    # Incrementing the last character is enough and works for both ostree/bootc and rpm systems.
    ${new_path}=    Set Variable    ${{ "${last_backup}[:-1]" + str(int("${last_backup}[-1]")+1) }}
    Command Should Work    sudo mv ${last_backup} ${new_path}

    Command Should Work    microshift restore --auto-recovery ${WORKDIR}
    ${backup_marker}=    Command Should Work    cat /var/lib/microshift/marker
    Should Be Equal As Numbers    ${backup_marker}    2
    Command Should Work    microshift restore --auto-recovery ${WORKDIR}

    [Teardown]    Run Keywords
    ...    Command Should Work    rm -rf ${WORKDIR}
    ...    AND
    ...    Command Should Work    rm -f /var/lib/microshift/marker

Previously Restored Backup Is Moved To Special Subdirectory
    [Documentation]    Executing Restore command results in
    ...    moving previously restored backup to a "restored" subdir.

    ${last_backup}=    Create Backups    3
    ${expected_path}=    Set Variable
    ...    ${{ "/".join( "${last_backup}".split("/")[:-1] + ["restored"] + [ "${last_backup}".split("/")[-1] ] ) }}
    Log    ${expected_path}
    Command Should Work    microshift restore --auto-recovery ${WORKDIR}
    # On second restore, previously restored backup ^ is moved into restored/ subdir.
    Command Should Work    microshift restore --auto-recovery ${WORKDIR}
    Command Should Work    ls ${expected_path}

    [Teardown]    Command Should Work    rm -rf ${WORKDIR}

MicroShift Data Is Backed Up For Later Analysis
    [Documentation]    When --save-failed option is given, MicroShift data
    ...    should be copied to a "failed" subdirectory for later analysis.

    ${last_backup}=    Create Backups    1
    Command Should Work    microshift restore --save-failed --auto-recovery ${WORKDIR}
    ${data_backup}=    Command Should Work    ls ${WORKDIR}/failed
    Verify Backup Name    ${data_backup}

    [Teardown]    Command Should Work    rm -rf ${WORKDIR}

State File Is Created
    [Documentation]    Restoring causes a state file creation.

    ${last_backup}=    Create Backups    1
    ${backup_name}=    Remove String    ${last_backup}    ${WORKDIR}

    Command Should Work    microshift restore --auto-recovery ${WORKDIR}
    ${state_json}=    Command Should Work    cat ${WORKDIR}/state.json
    ${state}=    Json Parse    ${state_json}
    Should Be Equal As Strings    ${backup_name}    ${state}[LastBackup]

    [Teardown]    Command Should Work    rm -rf ${WORKDIR}


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Login MicroShift Host
    # Scenario needs to start with MicroShift running, so there's something to back up.
    Stop MicroShift

Teardown
    [Documentation]    Test suite teardown
    Start MicroShift
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

Create Backups
    [Documentation]    Create specified amount of backups and return path to the last one.
    [Arguments]    ${amount}    ${markers}=${FALSE}
    FOR    ${counter}    IN RANGE    1    ${{ ${amount}+1 }}
        ${backup_path}=    Command Should Work    microshift backup --auto-recovery ${WORKDIR}
        IF    ${markers}
            Command Should Work    bash -c "echo ${counter} > ${backup_path}/marker"
        END
        Sleep    2s
    END
    RETURN    ${backup_path}
