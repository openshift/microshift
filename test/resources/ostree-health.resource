*** Settings ***
Documentation       Keywords for OSTree-based systems

Resource            systemd.resource
Resource            microshift-process.resource
Resource            ostree-data.resource
Resource            microshift-host.resource


*** Keywords ***
Wait Until Greenboot Health Check Exited
    [Documentation]    Wait until greenboot healthchecks are done

    # The host may reboot (for example during a rollback) while we are waiting,
    # which invalidates the current SSH session. Reconnect on every retry so a
    # stale session does not turn a transient reboot into a hard 10 minute
    # failure (SSHException: SSH session not active).
    Wait Until Keyword Succeeds    10m    15s
    ...    Reconnect And Check Greenboot Health Check Exited

Reconnect And Check Greenboot Health Check Exited
    [Documentation]    Re-establish the SSH connection and verify that greenboot finished.

    Make New SSH Connection
    Greenboot Health Check Exited

Greenboot Health Check Exited
    [Documentation]    Checks if greenboot-healthcheck finished running successfully (exited)

    Systemctl Check Service SubState    greenboot-healthcheck.service    exited

Wait For MicroShift Healthcheck Success
    [Documentation]    Verify the health of the deployments by running the
    ...    MicroShift healthcheck command and waiting for it to succeed
    [Arguments]    ${wait_timeout}=600s

    # Note that the newer implementation of greenboot-healthcheck service no
    # longer allows service restart. Using the MicroShift healthcheck command
    # to verify the health of the deployments.
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    microshift healthcheck -v=2 --timeout="${wait_timeout}"
    ...    sudo=True    return_stdout=True    return_stderr=True    return_rc=True
    IF    ${rc} != 0    Log Many    ${stdout}    ${stderr}
    Should Be Equal As Integers    0    ${rc}
