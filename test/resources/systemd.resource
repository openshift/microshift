*** Settings ***
Documentation       Keywords for interacting with systemd

Library             Process
Library             String
Resource            ../resources/microshift-host.resource


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

    ${status_text}=    Execute Command
    ...    systemctl status ${unit_name}
    ...    sudo=True    return_stdout=True
    ${log_text}=    Execute Command
    ...    journalctl -u ${unit_name} -o short | tail -c 32000
    ...    sudo=True    return_stdout=True

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
        VAR    ${state}    running
    ELSE IF    "${verb}" in {"stop"}
        VAR    ${state}    dead
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
