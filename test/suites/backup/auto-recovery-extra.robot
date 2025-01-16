*** Settings ***
Documentation       Extended tests related to auto-recovery functionality

Resource            ../../resources/microshift-host.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/ostree-health.resource
Library             Collections
Library             DateTime
Library             ../../resources/libostree.py

Suite Setup         Setup
Suite Teardown      Teardown


*** Variables ***
${WORKDIR}          /var/lib/microshift-auto-recovery/
${SAFETY_BACKUP}    /var/lib/microshift.safety


*** Test Cases ***
Using Systemd Dropin To React On Failure Of MicroShift
    [Documentation]    Verify systemd integration to start auto-recovery procedure.

    Stop MicroShift
    Create Safety Backup
    Create Backup With Marker
    Set Up MicroShift Auto Recovery Unit

    Corrupt Etcd Database
    ${dt}=    Command Should Work    date +"%s"
    Start MicroShift Expecting Failure
    Wait For Auto Recovery Unit To Finish    ${dt}
    Wait Until Keyword Succeeds    20x    5s
    ...    Systemctl Check Service SubState    microshift.service    running
    Command Should Work    ls /var/lib/microshift/marker

    [Teardown]    Run Keywords
    ...    Restore Safety Backup
    ...    AND
    ...    Clean Up MicroShift Auto Recovery Unit
    ...    AND
    ...    Command Should Work    rm -rf ${WORKDIR}

Auto Recovery On Red Boot
    [Documentation]    Verify greenboot integration to start auto-recovery procedure.

    Greenboot Workaround For Boot Counter
    Command Should Work    rm -rf ${SAFETY_BACKUP}

    Stop MicroShift
    Create Safety Backup
    Create Backup With Marker
    Set Up Greenboot Red Script

    Corrupt Etcd Database
    Start MicroShift Expecting Failure

    ${bootid}=    Get Current Boot Id
    Command Should Fail    systemctl restart greenboot-healthcheck
    Wait Until Keyword Succeeds    5m    15s
    ...    System Should Be Rebooted    ${bootid}
    Wait Until Greenboot Health Check Exited

    [Teardown]    Run Keywords
    ...    Restore Safety Backup
    ...    AND
    ...    Command Should Work    rm -f /etc/greenboot/red.d/100-auto-recovery.sh
    ...    AND
    ...    Command Should Work    rm -rf ${WORKDIR}


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Login MicroShift Host

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host

Create Backup With Marker
    [Documentation]    Create a backup with a marker file inside.
    ...    Marker is later used to verify if the backup was actually restored.
    ${backup_path}=    Command Should Work    microshift backup --auto-recovery ${WORKDIR}
    Command Should Work    touch ${backup_path}/marker

Set Up MicroShift Auto Recovery Unit    # robocop: disable=too-many-calls-in-keyword
    [Documentation]    Sets up drop-in and service to run on MicroShift's failure.

    Command Should Work    mkdir -p /etc/systemd/system/microshift.service.d
    ${drop_in}=    Operating System.Get File    ./assets/auto-recovery/10-auto-recovery.conf
    Upload String To File    ${drop_in}    /etc/systemd/system/microshift.service.d/10-auto-recovery.conf

    ${svc}=    Operating System.Get File    ./assets/auto-recovery/microshift-auto-recovery.service
    Upload String To File    ${svc}    /etc/systemd/system/microshift-auto-recovery.service

    ${bin}=    Operating System.Get File    ./assets/auto-recovery/microshift-auto-recovery
    Upload String To File    ${bin}    /var/opt/microshift-auto-recovery
    Command Should Work    chmod +x /var/opt/microshift-auto-recovery
    # Following is needed for a systemd unit to call binary on different filesystem
    Command Should Work    semanage fcontext --add --type bin_t "/var/opt(/.*)?"
    Command Should Work    restorecon -r -v /var/opt/

    Command Should Work    systemctl daemon-reload

    ${stdout}    ${stderr}    ${rc}=    SSHLibrary.Execute Command
    ...    systemctl status microshift-auto-recovery.service
    ...    sudo=True
    ...    return_rc=True
    ...    return_stderr=True
    ...    return_stdout=True
    Command Should Work    systemctl show microshift.service

Clean Up MicroShift Auto Recovery Unit
    [Documentation]    Remove the microshift-auto-recovery.service and microshift.service's drop-in.
    ${stdout}    ${stderr}    ${rc}=    SSHLibrary.Execute Command
    ...    systemctl status microshift-auto-recovery.service
    ...    sudo=True
    ...    return_rc=True
    ...    return_stderr=True
    ...    return_stdout=True
    Command Should Work    systemctl show microshift.service
    Command Should Work    rm -rf /etc/systemd/system/microshift.service.d
    Command Should Work    rm -f /etc/systemd/system/microshift-auto-recovery.service
    Command Should Work    rm -f /var/opt/microshift-auto-recovery
    Command Should Work    semanage fcontext --delete --type bin_t "/var/opt(/.*)?"
    Command Should Work    restorecon -r -v /var/opt/
    Command Should Work    systemctl daemon-reload

Start MicroShift Expecting Failure
    [Documentation]    Start MicroShift expecting it fails.
    Command Should Fail    systemctl start microshift

Create Safety Backup
    [Documentation]    Create an additional backup to restore during teardown
    ...    in case something goes wrong with auto-recovery to not disrupt next test.
    Command Should Work    microshift backup ${SAFETY_BACKUP}

Restore Safety Backup
    [Documentation]    Restores additional backup during teardown and starts MicroShift.
    Run Keyword And Ignore Error
    ...    Stop MicroShift
    Command Should Work    microshift restore ${SAFETY_BACKUP}
    Command Should Work    rm -rf ${SAFETY_BACKUP}
    Start MicroShift

Corrupt Etcd Database
    [Documentation]    Overwrites first megabyte of the etcd's database to simulate
    ...    corruption and cause MicroShift startup failure.
    Command Should Work
    ...    dd conv=notrunc if=/dev/zero of=/var/lib/microshift/etcd/member/wal/0000000000000000-0000000000000000.wal bs=1M count=1

Wait For Auto Recovery Unit To Finish
    [Documentation]    Wait until microshift-auto-recovery.service completes
    ...    running (unit enters state 'dead').
    [Arguments]    ${dt}
    Wait Until Keyword Succeeds    20x    5s
    ...    Auto Recovery Finished Recently    ${dt}
    Command Should Work
    ...    journalctl -u microshift-auto-recovery.service _SYSTEMD_INVOCATION_ID="$(systemctl show -p InvocationID --value microshift-auto-recovery.service)"

Auto Recovery Finished Recently
    [Documentation]    Check if microshift-auto-recovery.service is dead and finished later than provided timestamp.
    [Arguments]    ${dt}
    Systemctl Check Service SubState    microshift-auto-recovery.service    dead
    ${finish_dt}=    Command Should Work
    ...    date -d "$(systemctl show --property=ExecMainExitTimestamp --value microshift-auto-recovery)" +"%s"
    Should Be True    ${dt} < ${finish_dt}

Set Up Greenboot Red Script
    [Documentation]    Set up red script that triggers auto-recovery procedure.
    Command Should Work    mkdir -p /etc/greenboot/red.d
    ${drop_in}=    Operating System.Get File    ./assets/auto-recovery/red-script.sh
    Upload String To File    ${drop_in}    /etc/greenboot/red.d/99-auto-recovery.sh

Greenboot Workaround For Boot Counter
    [Documentation]    If the grub's boot_success is 1, clears boot_counter.
    # Because of greenboot's bug, we need to it here, so the system doesn't reboot after red boot.
    Command Should Work
    ...    bash -c "grub2-editenv list | grep -q boot_success=1 && /usr/bin/grub2-editenv /boot/grub2/grubenv unset boot_counter || true"
