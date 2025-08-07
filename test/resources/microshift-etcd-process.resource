*** Settings ***
Documentation       Keywords for running the microshift-etcd command line.

Library             OperatingSystem
Library             Process
Library             String
Library             SSHLibrary
Resource            oc.resource
Resource            systemd.resource
Library             DataFormats.py


*** Keywords ***
MicroShift Etcd Process ID
    [Documentation]    Return the current MicroShift Etcd process ID
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    pidof microshift-etcd
    ...    return_stdout=True    return_stderr=True    return_rc=True
    Log    ${stderr}
    RETURN    ${stdout}

MicroShift Etcd Process ID Changed
    [Documentation]    Compare the specified and current MicroShift Etcd process ID
    [Arguments]    ${old_pid}
    ${cur_pid}=    MicroShift Etcd Process ID
    Should Not Be Equal    ${cur_pid}    ${old_pid}

Wait Until MicroShift Etcd Process ID Changes
    [Documentation]    Wait until the current MicroShift Etcd process ID changes
    [Arguments]    ${old_pid}

    Wait Until Keyword Succeeds    30x    10s
    ...    MicroShift Etcd Process ID Changed    ${old_pid}
