*** Settings ***
Documentation       Keywords for interacting with systemd

Library             Collections
Library             OperatingSystem
Library             Process
Library             String
Library             libipv6.py
Resource            ../resources/microshift-host.resource


*** Variables ***
${SYSTEMD_DIAGNOSTIC_REMOTE_TIMEOUT}        25s
${SYSTEMD_DIAGNOSTIC_COMMAND_TIMEOUT}       30s
${SYSTEMD_DIAGNOSTIC_TRANSFER_TIMEOUT}      60s
${SYSTEMD_DIAGNOSTIC_CLEANUP_TIMEOUT}       10s


*** Keywords ***
Get Systemd Setting
    [Documentation]    Fetch one setting from systemd for the named unit.
    ...    Take care to get the unit_name value _exactly_ right, or
    ...    systemd will report a default value without reporting any error
    ...    or warning.
    [Arguments]    ${unit_name}    ${property}

    ${stdout}    ${rc}=    Execute Command
    ...    systemctl show --property=${property} --value ${unit_name}
    ...    sudo=True    return_rc=True
    Should Be Equal As Integers    ${rc}    0
    Should Not Be Empty    ${stdout}

    ${result}=    Strip String    ${stdout}
    RETURN    ${result}

Systemctl Print Service Status And Logs
    [Documentation]    Dumps the status and latest logs for a specified service.
    [Arguments]    ${unit_name}

    ${status_result}    ${status_text}=    Run Keyword And Ignore Error
    ...    Systemctl Get Service Status    ${unit_name}
    IF    '${status_result}' == 'PASS'
        Log    ${status_text}
    ELSE
        Log    Unable to collect status for ${unit_name}: ${status_text}    WARN
    END

    ${journal_result}    ${journal_text}=    Run Keyword And Ignore Error
    ...    Systemctl Get Service Journal Tail    ${unit_name}
    IF    '${journal_result}' == 'PASS'
        Log    ${journal_text}
    ELSE
        Log    Unable to collect journal for ${unit_name}: ${journal_text}    WARN
    END

Systemctl Get Service Status
    [Documentation]    Return service status without asserting that the service is healthy.
    [Arguments]    ${unit_name}

    Validate Systemd Unit Name    ${unit_name}
    ${command}=    Catenate    SEPARATOR=${SPACE}
    ...    timeout --signal=TERM --kill-after=5s ${SYSTEMD_DIAGNOSTIC_REMOTE_TIMEOUT}
    ...    systemctl status --no-pager --full ${unit_name}
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    ${command}
    ...    sudo=True    return_stdout=True    return_stderr=True    return_rc=True
    ...    timeout=${SYSTEMD_DIAGNOSTIC_COMMAND_TIMEOUT}
    ${status_text}=    Catenate    SEPARATOR=${\n}
    ...    systemctl status exit code: ${rc}
    ...    ${stdout}
    ...    ${stderr}
    RETURN    ${status_text}

Systemctl Get Service Journal Tail
    [Documentation]    Return at most the last 32 KB of journal for a service.
    [Arguments]    ${unit_name}

    ${journal_text}=    Systemctl Get Service Journal Tail Using Keyword
    ...    ${unit_name}    Execute Command
    RETURN    ${journal_text}

Systemctl Get Service Journal Tail Using Keyword
    [Documentation]    Return bounded service journal output using the supplied command keyword.
    [Arguments]    ${unit_name}    ${execute_keyword}

    Validate Systemd Unit Name    ${unit_name}
    ${command}=    Catenate    SEPARATOR=${SPACE}
    ...    timeout --signal=TERM --kill-after=5s ${SYSTEMD_DIAGNOSTIC_REMOTE_TIMEOUT}
    ...    bash -o pipefail -c 'journalctl --no-pager -u ${unit_name} -o short 2>&1 | tail -c 32000'
    ${stdout}    ${stderr}    ${rc}=    Run Keyword
    ...    ${execute_keyword}    ${command}
    ...    sudo=True    return_stdout=True    return_stderr=True    return_rc=True
    ...    timeout=${SYSTEMD_DIAGNOSTIC_COMMAND_TIMEOUT}
    ${journal_text}=    Evaluate
    ...    ($stdout + chr(10) + $stderr + chr(10) + "journal pipeline exit code: " + str($rc)).encode("utf-8")[-32000:].decode("utf-8", errors="ignore")
    RETURN    ${journal_text}

Systemctl Save Service Status And Logs
    [Documentation]    Save full service status and journal output to a local diagnostic file.
    [Arguments]    ${unit_name}    ${local_path}    ${diagnostic_id}

    Systemctl Save Service Status And Logs Using Keyword
    ...    ${unit_name}    ${local_path}    ${diagnostic_id}
    ...    Download Systemd Diagnostic File

Systemctl Save Service Status And Logs Using Keyword
    [Documentation]    Save full service diagnostics using the supplied download keyword.
    [Arguments]    ${unit_name}    ${local_path}    ${diagnostic_id}    ${download_keyword}

    Validate Systemd Unit Name    ${unit_name}
    Should Match Regexp    ${diagnostic_id}    ^[0-9a-f]{32}$
    VAR    ${remote_path}=    /tmp/systemd-diagnostics-${diagnostic_id}.log
    ${collect_command}=    Catenate    SEPARATOR=${SPACE}
    ...    timeout --signal=TERM --kill-after=5s ${SYSTEMD_DIAGNOSTIC_REMOTE_TIMEOUT}
    ...    bash -c 'umask 077; owner="$SUDO_USER"; group=$(id -gn -- "$owner") || exit;
    ...    install -m 600 -o "$owner" -g "$group" /dev/null ${remote_path} || exit;
    ...    {
    ...    systemctl status --no-pager --full ${unit_name}; status_rc=$?;
    ...    printf "\nsystemctl status exit code: %s\n" "$status_rc";
    ...    journalctl --no-pager -u ${unit_name} -o short; journal_rc=$?;
    ...    printf "\njournalctl exit code: %s\n" "$journal_rc";
    ...    } >> ${remote_path} 2>&1;
    ...    chmod 600 ${remote_path}'
    TRY
        ${stdout}    ${stderr}    ${rc}=    Execute Command
        ...    ${collect_command}
        ...    sudo=True    return_stdout=True    return_stderr=True    return_rc=True
        ...    timeout=${SYSTEMD_DIAGNOSTIC_COMMAND_TIMEOUT}
        IF    ${rc} != 0
            Fail    Unable to prepare full diagnostics for ${unit_name} (rc=${rc}): ${stderr}
        END
        Run Keyword    ${download_keyword}    ${remote_path}    ${local_path}
    FINALLY
        Run Keyword And Ignore Error
        ...    Execute Command    rm -f -- ${remote_path}
        ...    return_rc=True    timeout=${SYSTEMD_DIAGNOSTIC_CLEANUP_TIMEOUT}
    END

Download Systemd Diagnostic File
    [Documentation]    Download a systemd diagnostic file from the MicroShift host.
    [Arguments]    ${remote_path}    ${local_path}

    Download Systemd Diagnostic File Using Keyword
    ...    ${remote_path}    ${local_path}    Process.Run Process

Download Systemd Diagnostic File Using Keyword
    [Documentation]    Download diagnostics with a hard process deadline using the supplied runner.
    [Arguments]    ${remote_path}    ${local_path}    ${process_keyword}

    VAR    @{scp_args}=
    ...    -B    -q
    ...    -o    ConnectTimeout=10
    ...    -o    ConnectionAttempts=1
    ...    -o    ServerAliveInterval=5
    ...    -o    ServerAliveCountMax=1
    ...    -o    StrictHostKeyChecking=accept-new
    IF    '${SSH_PORT}'    Append To List    ${scp_args}    -P    ${SSH_PORT}
    IF    '${SSH_PRIV_KEY}'
        Append To List    ${scp_args}    -i    ${SSH_PRIV_KEY}
    END
    VAR    ${partial_path}=    ${local_path}.partial
    ${scp_host}=    Add Brackets If Ipv6    ${USHIFT_HOST}
    Append To List
    ...    ${scp_args}
    ...    --    ${USHIFT_USER}@${scp_host}:${remote_path}    ${partial_path}
    TRY
        ${result}=    Run Keyword
        ...    ${process_keyword}    scp    @{scp_args}
        ...    timeout=${SYSTEMD_DIAGNOSTIC_TRANSFER_TIMEOUT}
        ...    on_timeout=kill    stdout=DEVNULL    stderr=DEVNULL
        Should Be Equal As Integers
        ...    ${result.rc}    0
        ...    msg=Diagnostic transfer failed or timed out (rc=${result.rc})
        OperatingSystem.Move File    ${partial_path}    ${local_path}
    FINALLY
        Run Keyword And Ignore Error    OperatingSystem.Remove File    ${partial_path}
    END

Validate Systemd Unit Name
    [Documentation]    Reject values that are unsafe to interpolate into diagnostic shell commands.
    [Arguments]    ${unit_name}

    Should Match Regexp    ${unit_name}    ^[A-Za-z0-9_.@:][A-Za-z0-9_.@:-]*$

Systemctl Check Service SubState
    [Documentation]    Check if the current sub-state of a service is the same
    ...    as the expected sub-state.
    [Arguments]    ${unit_name}    ${expected_state}

    ${sub_state}=    Get Systemd Setting    ${unit_name}    SubState
    Should Be Equal As Strings    ${sub_state}    ${expected_state}

Systemctl
    [Documentation]    Run a systemctl command on the microshift host.
    ...    The intent is to enable, start, stop, or restart a service.
    ...    Other commands should be implemented separately. When the verb
    ...    is "start" or "restart", this keyword will wait for the unit
    ...    to report that it is "running". When the verb is "stop", this
    ...    keyword will wait for the unit to report that it is "dead".
    ...    When the verb is "enable", this keyword will only check success
    ...    of the operation and not wait for any change to the service.
    [Arguments]    ${verb}    ${unit_name}

    # Verify the input
    Should Be True    "${verb}" in {"restart", "start", "stop", "enable", "disable"}

    IF    "${verb}" in {"restart", "start"}
        VAR    ${state}=    running
    ELSE IF    "${verb}" in {"stop"}
        VAR    ${state}=    dead
    END

    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    systemctl ${verb} ${unit_name}
    ...    sudo=True    return_stdout=True    return_stderr=True    return_rc=True
    IF    ${rc} != 0    Systemctl Print Service Status And Logs    ${unit_name}
    Should Be Equal As Integers    0    ${rc}

    IF    "${verb}" in {"enable", "disable"}    RETURN

    # It takes a bit for systemd to respond, and if we check too soon
    # then it looks like microshift is up, even though it is about to
    # be restarted.
    Sleep    5s

    Wait Until Keyword Succeeds    10x    10s
    ...    Systemctl Check Service SubState    ${unit_name}    ${state}

Systemctl Daemon Reload
    [Documentation]    Reload the systemd daemon.
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    systemctl daemon-reload
    ...    sudo=True    return_stdout=True    return_stderr=True    return_rc=True
    Log Many    ${stdout}    ${stderr}
    Should Be Equal As Integers    0    ${rc}

Disable Journal Rate Limiting
    [Documentation]    Disable journald rate limiting by writing a drop-in that
    ...    sets RateLimitBurst=0, then restarting the journal service.
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    bash -c "mkdir -p /etc/systemd/journald.conf.d && printf '[Journal]\nRateLimitBurst=0\n' > /etc/systemd/journald.conf.d/disable-ratelimit.conf && systemctl restart systemd-journald"
    ...    sudo=True
    ...    return_stdout=True
    ...    return_stderr=True
    ...    return_rc=True
    Log Many    ${stdout}    ${stderr}
    Should Be Equal As Integers    0    ${rc}

Enable Journal Rate Limiting
    [Documentation]    Re-enable default journald rate limiting by removing the
    ...    drop-in created by Disable Journal Rate Limiting.
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    bash -c "rm -f /etc/systemd/journald.conf.d/disable-ratelimit.conf && systemctl restart systemd-journald"
    ...    sudo=True
    ...    return_stdout=True
    ...    return_stderr=True
    ...    return_rc=True
    Log Many    ${stdout}    ${stderr}
    Should Be Equal As Integers    0    ${rc}
