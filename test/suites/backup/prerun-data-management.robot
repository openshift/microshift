*** Settings ***
Documentation       Tests verifying MicroShift prerun data management behavior
...                 when version file indicates an incompatible version or
...                 when data directory is missing but health info is present.
...
...                 Ported from openshift-tests-private:
...                 OCP-66820, OCP-66882

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-host.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/ostree-health.resource
Library             ../../resources/libostree.py

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           ostree    restart    slow


*** Variables ***
${DATA_DIR}             /var/lib/microshift
${BACKUP_STORAGE}       /var/lib/microshift-backups
${VERSION_FILE}         /var/lib/microshift/version
${PRERUN_FAILED_LOG}    /var/lib/microshift-backups/prerun_failed.log
${VERSION_BACKUP}       /var/tmp/microshift-version.backup


*** Test Cases ***
Prerun Failure Is Logged When Version Is Too Old
    [Documentation]    Verify that when the version file indicates a version
    ...    3 minor versions behind the executable, MicroShift fails to start
    ...    after reboot and the failure reason is logged to prerun_failed.log
    ...    and reported by greenboot healthcheck.
    ...    OCP-66820

    Wait For MicroShift Service
    Stop MicroShift

    Save Version File
    Set Version N Minors Back    3

    Reboot MicroShift Host And Wait For Greenboot

    Greenboot Journal Should Report Prerun Failure
    Prerun Failed Log Should Contain Version Mismatch

    [Teardown]    Run Keywords
    ...    Run Keyword And Ignore Error    Restore Version File
    ...    AND
    ...    Run Keyword And Ignore Error    Remove Prerun Failed Log
    ...    AND
    ...    Start MicroShift

Data Missing With Healthy Status Starts Fresh
    [Documentation]    Verify that when MicroShift data directory is removed
    ...    but health.json shows healthy status, MicroShift ignores
    ...    the health info and starts fresh as if it were the first run.
    ...    OCP-66882

    Wait For MicroShift Service
    Reboot MicroShift Host
    Wait For MicroShift Service
    Backup Should Be Present
    Health File Should Show Healthy

    Stop MicroShift
    Remove MicroShift Data

    Health File Should Show Healthy

    Reboot MicroShift Host And Wait For Greenboot
    Wait For MicroShift Service

    Current Boot Journal Should Show Fresh Start

    [Teardown]    Ensure MicroShift Is Running


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host

Save Version File
    [Documentation]    Save the current version file for later restoration
    Command Should Work    cp ${VERSION_FILE} ${VERSION_BACKUP}

Restore Version File
    [Documentation]    Restore the version file from backup
    Command Should Work    cp ${VERSION_BACKUP} ${VERSION_FILE}
    Command Should Work    rm -f ${VERSION_BACKUP}

Set Version N Minors Back
    [Documentation]    Modify the version string in the version file to be N minor versions back.
    ...    The version file stores version as a string "Major.Minor.Patch".
    [Arguments]    ${n}
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    jq -c '.version |= (split(".") | .[1] = (.[1] | tonumber - ${n} | tostring) | join("."))' ${VERSION_FILE} > /tmp/microshift-version.new
    ...    sudo=True
    ...    return_stdout=True
    ...    return_stderr=True
    ...    return_rc=True
    Should Be Equal As Integers    0    ${rc}
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    sudo mv /tmp/microshift-version.new ${VERSION_FILE}
    ...    sudo=True    return_stderr=True    return_rc=True
    Should Be Equal As Integers    0    ${rc}

Greenboot Health Check Should Be Finished
    [Documentation]    Check that greenboot-healthcheck has finished running
    ...    with either SubState=exited (success) or SubState=failed (failure)
    ...    and that no rollback reboot is pending. After greenboot-healthcheck
    ...    reaches SubState=failed, the rollback handler may trigger a reboot.
    ...    Comparing boot IDs before and after a delay detects this and forces
    ...    a retry via the outer Wait Until Keyword Succeeds loop.
    Make New SSH Connection
    ${stdout}=    Command Should Work
    ...    systemctl show -p SubState --value greenboot-healthcheck.service
    Should Match Regexp    ${stdout}    ^(exited|failed)$
    ${bootid_before}=    Get Current Boot Id
    Sleep    10s
    Make New SSH Connection
    ${bootid_after}=    Get Current Boot Id
    Should Be Equal    ${bootid_before}    ${bootid_after}

Greenboot Journal Should Report Prerun Failure
    [Documentation]    Verify the greenboot healthcheck journal reports
    ...    the prerun failure log from the microshift service
    ${stdout}=    Command Should Work
    ...    journalctl -o cat -u greenboot-healthcheck -b 0 --no-pager
    Should Contain    ${stdout}    Prerun failure log

Prerun Failed Log Should Contain Version Mismatch
    [Documentation]    Verify prerun_failed.log contains the version mismatch error
    ${stdout}=    Command Should Work    cat ${PRERUN_FAILED_LOG}
    Should Contain    ${stdout}    too recent compared to existing data
    Should Contain    ${stdout}    maximum allowed

Remove Prerun Failed Log
    [Documentation]    Remove the prerun_failed.log file
    Command Should Work    rm -f ${PRERUN_FAILED_LOG}

Backup Should Be Present
    [Documentation]    Verify that at least one backup subdirectory exists in the backup storage
    ${stdout}=    Command Should Work    find ${BACKUP_STORAGE} -mindepth 1 -maxdepth 1 -type d
    Should Not Be Empty    ${stdout}

Health File Should Show Healthy
    [Documentation]    Verify health.json exists and shows healthy status
    ${health}=    Get Persisted System Health
    Should Be Equal As Strings    ${health}    healthy

Remove MicroShift Data
    [Documentation]    Remove the MicroShift data directory while preserving backups
    Command Should Work    rm -rf ${DATA_DIR}

Current Boot Journal Should Show Fresh Start
    [Documentation]    Verify current boot journal contains messages indicating
    ...    MicroShift started fresh without existing data
    ${stdout}=    Command Should Work
    ...    journalctl -o cat -u microshift -b 0 --no-pager
    Should Contain    ${stdout}    MicroShift data does not exist - skipping backup, continuing startup
    Should Contain    ${stdout}    Version file does not exist yet - assuming first run

Reboot MicroShift Host And Wait For Greenboot
    [Documentation]    Reboot and wait for greenboot to finish without asserting
    ...    reboot count. Use instead of Reboot MicroShift Host when greenboot
    ...    rollback loops are expected (e.g. intentional prerun failures or
    ...    slow fresh starts on ARM that exceed greenboot timeout).
    SSHLibrary.Start Command    reboot    sudo=True
    Wait Until Keyword Succeeds    10m    15s
    ...    Greenboot Health Check Should Be Finished

Ensure MicroShift Is Running
    [Documentation]    Make sure MicroShift is running after test
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    systemctl is-active microshift
    ...    sudo=True    return_stdout=True    return_stderr=True    return_rc=True
    IF    ${rc} != 0    Start MicroShift
