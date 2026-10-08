*** Settings ***
Documentation       Keywords for OSTree-based systems

Library             greenboot_diagnostics.py    AS    GreenbootDiagnostics
Resource            systemd.resource
Resource            microshift-process.resource
Resource            ostree-data.resource
Resource            microshift-host.resource


*** Variables ***
${GREENBOOT_DIAGNOSTIC_RECONNECT_TIMEOUT}       10s
${GREENBOOT_DIAGNOSTIC_REMOTE_TIMEOUT}          25s
${GREENBOOT_DIAGNOSTIC_COMMAND_TIMEOUT}         30s


*** Keywords ***
Wait Until Greenboot Health Check Exited
    [Documentation]    Wait until greenboot healthchecks are done
    [Arguments]    ${check_keyword}=Reconnect And Check Greenboot Health Check Exited
    ...    ${diagnostics_keyword}=GreenbootDiagnostics.Collect Public Greenboot Diagnostic Summary

    # The host may reboot (for example during a rollback) while we are waiting,
    # which invalidates the current SSH session. Reconnect on every retry so a
    # stale session does not turn a transient reboot into a hard 10 minute
    # failure (SSHException: SSH session not active).
    Wait Until Greenboot Health Check Exited Using Keywords
    ...    10m
    ...    15s
    ...    ${check_keyword}
    ...    ${diagnostics_keyword}

Wait Until Greenboot Health Check Exited Using Keywords
    [Documentation]    Wait for greenboot and run the supplied diagnostics after the final failure.
    [Arguments]    ${wait_timeout}    ${retry_interval}    ${check_keyword}    ${diagnostics_keyword}
    Wait Until Keyword Succeeds    ${wait_timeout}    ${retry_interval}
    ...    ${check_keyword}
    [Teardown]
    ...    Collect Greenboot Diagnostics After Failed Wait
    ...    ${diagnostics_keyword}    ${KEYWORD STATUS}

Collect Greenboot Diagnostics After Failed Wait
    [Documentation]    Collect after failure without copying the primary exception text.
    [Arguments]    ${diagnostics_keyword}    ${wait_status}

    IF    '${wait_status}' == 'FAIL'
        Collect Greenboot Health Check Failure Diagnostics Using Keywords
        ...    greenboot-final-wait    ${None}    ${diagnostics_keyword}
    END

Reconnect And Check Greenboot Health Check Exited
    [Documentation]    Re-establish the SSH connection and verify that greenboot finished.

    Make New SSH Connection
    Greenboot Health Check Exited

Greenboot Health Check Exited
    [Documentation]    Checks if greenboot-healthcheck finished running successfully (exited)

    Systemctl Check Service SubState    greenboot-healthcheck.service    exited

Collect Greenboot Health Check Failure Diagnostics
    [Documentation]    Save an allowlisted public summary after a health failure.
    [Arguments]    ${trigger}    ${primary_exit_code}=${None}

    Collect Greenboot Health Check Failure Diagnostics Using Keywords
    ...    ${trigger}    ${primary_exit_code}
    ...    GreenbootDiagnostics.Collect Public Greenboot Diagnostic Summary

Collect Greenboot Health Check Failure Diagnostics Using Keywords
    [Documentation]    Save a public summary using the supplied collector keyword.
    [Arguments]    ${trigger}    ${primary_exit_code}    ${collector_keyword}

    GreenbootDiagnostics.Run Public Greenboot Diagnostics Safely
    ...    ${collector_keyword}
    ...    ${trigger}    ${primary_exit_code}
    ...    reconnect_timeout=${GREENBOOT_DIAGNOSTIC_RECONNECT_TIMEOUT}
    ...    remote_timeout=${GREENBOOT_DIAGNOSTIC_REMOTE_TIMEOUT}
    ...    command_timeout=${GREENBOOT_DIAGNOSTIC_COMMAND_TIMEOUT}

Wait For MicroShift Healthcheck Success
    [Documentation]    Verify the health of the deployments by running the
    ...    MicroShift healthcheck command and waiting for it to succeed
    [Arguments]    ${wait_timeout}=600s
    ...    ${execute_keyword}=Execute Command
    ...    ${diagnostics_keyword}=GreenbootDiagnostics.Collect Public Greenboot Diagnostic Summary

    Wait For MicroShift Healthcheck Success Using Keywords
    ...    ${wait_timeout}
    ...    ${execute_keyword}
    ...    ${diagnostics_keyword}

Wait For MicroShift Healthcheck Success Using Keywords
    [Documentation]    Run healthcheck and collect a public summary before asserting its result.
    [Arguments]    ${wait_timeout}    ${execute_keyword}    ${diagnostics_keyword}

    # Raw command output is intentionally neither returned nor logged. There
    # is no approved private artifact destination for it.
    VAR    ${command}=    microshift healthcheck -v=2 --timeout="${wait_timeout}"
    ${rc}=    Run Keyword
    ...    ${execute_keyword}
    ...    ${command}
    ...    sudo=True    return_stdout=False    return_stderr=False    return_rc=True
    IF    ${rc} != 0
        # ${rc} is the cause of this failure and is recorded as primary_exit_code.
        # The diagnostics query greenboot-healthcheck.service, which runs the same
        # check at boot, so its service_* fields reflect that boot-time run rather
        # than this ad-hoc invocation (useful only to tell a regression-since-boot
        # from a never-healthy boot). See greenboot_diagnostics.py _TRIGGERS.
        Collect Greenboot Health Check Failure Diagnostics Using Keywords
        ...    direct-healthcheck    ${rc}    ${diagnostics_keyword}
    END
    Should Be Equal As Integers    0    ${rc}
