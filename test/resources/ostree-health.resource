*** Settings ***
Documentation       Keywords for OSTree-based systems

Resource            systemd.resource
Resource            microshift-process.resource
Resource            ostree-data.resource
Resource            microshift-host.resource


*** Variables ***
${GREENBOOT_DIAGNOSTIC_RECONNECT_TIMEOUT}       10s


*** Keywords ***
Wait Until Greenboot Health Check Exited
    [Documentation]    Wait until greenboot healthchecks are done

    # The host may reboot (for example during a rollback) while we are waiting,
    # which invalidates the current SSH session. Reconnect on every retry so a
    # stale session does not turn a transient reboot into a hard 10 minute
    # failure (SSHException: SSH session not active).
    Wait Until Greenboot Health Check Exited Using Keywords
    ...    10m
    ...    15s
    ...    Reconnect And Check Greenboot Health Check Exited
    ...    Collect Greenboot Health Check Failure Diagnostics

Wait Until Greenboot Health Check Exited Using Keywords
    [Documentation]    Wait for greenboot and run the supplied diagnostics after the final failure.
    [Arguments]    ${wait_timeout}    ${retry_interval}    ${check_keyword}    ${diagnostics_keyword}

    ${wait_status}    ${wait_error}=    Run Keyword And Ignore Error
    ...    Wait Until Keyword Succeeds    ${wait_timeout}    ${retry_interval}
    ...    ${check_keyword}
    IF    '${wait_status}' == 'PASS'    RETURN

    # Diagnostics are deliberately best effort so failures while collecting
    # them cannot replace the original wait failure.
    Run Keyword And Ignore Error    ${diagnostics_keyword}
    Fail    ${wait_error}

Reconnect And Check Greenboot Health Check Exited
    [Documentation]    Re-establish the SSH connection and verify that greenboot finished.

    Make New SSH Connection
    Greenboot Health Check Exited

Greenboot Health Check Exited
    [Documentation]    Checks if greenboot-healthcheck finished running successfully (exited)

    Systemctl Check Service SubState    greenboot-healthcheck.service    exited

Collect Greenboot Health Check Failure Diagnostics
    [Documentation]    Capture greenboot status and logs after the final wait failure.

    Collect Greenboot Health Check Failure Diagnostics Using Keywords
    ...    Make New SSH Connection For Greenboot Diagnostics
    ...    Systemctl Print Service Status And Logs
    ...    Systemctl Save Service Status And Logs

Collect Greenboot Health Check Failure Diagnostics Using Keywords
    [Documentation]    Capture greenboot diagnostics using the supplied collection keywords.
    [Arguments]    ${reconnect_keyword}    ${print_keyword}    ${save_keyword}

    ${reconnect_status}    ${reconnect_error}=    Run Keyword And Ignore Error
    ...    ${reconnect_keyword}
    IF    '${reconnect_status}' == 'FAIL'
        Log    Unable to reconnect for greenboot diagnostics: ${reconnect_error}    WARN
    END

    Run Keyword And Ignore Error
    ...    ${print_keyword}    greenboot-healthcheck.service

    ${diagnostic_id}=    Evaluate    uuid.uuid4().hex    modules=uuid
    VAR    ${diagnostic_path}=
    ...    ${OUTPUTDIR}${/}greenboot-healthcheck-diagnostics-${diagnostic_id}.log
    ${save_status}    ${save_error}=    Run Keyword And Ignore Error
    ...    ${save_keyword}
    ...    greenboot-healthcheck.service    ${diagnostic_path}    ${diagnostic_id}
    IF    '${save_status}' == 'FAIL'
        Log    Unable to save full greenboot diagnostics: ${save_error}    WARN
    ELSE
        Log    Full greenboot diagnostics saved to ${diagnostic_path}
    END

Make New SSH Connection For Greenboot Diagnostics
    [Documentation]    Reconnect for failure diagnostics with a finite connection timeout.

    Make New SSH Connection    ${GREENBOOT_DIAGNOSTIC_RECONNECT_TIMEOUT}

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
