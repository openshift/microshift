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

Systemctl    # robocop: disable=too-long-keyword
    [Documentation]    Run a systemctl command on the microshift host.
    ...    The intent is to start, stop, or restart a service. Other
    ...    commands should be implemented separately.    When the verb is
    ...    "start" or "restart", this keyword will wait for the unit
    ...    to report that it is "running". When the verb is "stop", this
    ...    keyword will wait for the unit to report that it is "dead".
    ...    When the verb is "enable", this keyword will only check success
    ...    of the operation and not wait for any change to the service..
    [Arguments]    ${verb}    ${unit_name}

    IF    "${verb}" in {"restart", "start"}
        ${state}=    Set Variable    running
    ELSE IF    "${verb}" in {"stop"}
        ${state}=    Set Variable    dead
    END

    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    systemctl ${verb} ${unit_name}
    ...    sudo=True
    ...    return_stdout=True
    ...    return_stderr=True
    ...    return_rc=True
    IF    ${rc} != 0
        ${status_text}=    Execute Command
        ...    systemctl status ${unit_name}
        ...    sudo=True
        ...    return_stdout=True
        ${log_text}=    Execute Command
        ...    journalctl -u ${unit_name} -o short | tail -n 100
        ...    sudo=True
        ...    return_stdout=True
    END
    Should Be Equal As Integers    0    ${rc}

    IF    "${verb}" == "enable"    RETURN

    # It takes a bit for systemd to respond, and if we check too soon
    # then it looks like microshift is up, even though it is about to
    # be restarted.
    Sleep    5s

    Wait Until Keyword Succeeds    10x    10s
    ...    Execute Command    [ $(systemctl show -p SubState --value ${unit_name}) = ${state} ]
    ...        timeout=10s
    ...        return_stdout=True
    ...        return_stderr=True

Systemctl With Retry
    [Documentation]    Run Systemctl keyword but retry 10 times
    [Arguments]    ${verb}    ${unit_name}
    Wait Until Keyword Succeeds    10x    10s
    ...    Systemctl    ${verb}    ${unit_name}
