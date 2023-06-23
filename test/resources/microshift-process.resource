*** Settings ***
Documentation       Keywords for running the microshift command line.

Library             Process
Library             String
Library             OperatingSystem
Library             SSHLibrary
Resource            oc.resource
Resource            systemd.resource
Resource            microshift-host.resource
Library             DataFormats.py


*** Keywords ***
MicroShift Version
    [Documentation]    Run the version command
    ${version_text}    ${rc}=    Execute Command
    ...    microshift version -o yaml
    ...    sudo=True    return_rc=True
    Should Be Equal As Integers    ${rc}    0
    Should Not Be Empty    ${version_text}
    ${version}=    Yaml Parse    ${version_text}
    RETURN    ${version}

MicroShift Is Ready
    [Documentation]    Check the /readyz endpoint
    ${stdout}=    Run With Kubeconfig    oc get --raw='/readyz'
    Should Be Equal As Strings    ${stdout}    ok    strip_spaces=True

MicroShift Is Live
    [Documentation]    Check the /livez endpoint
    ${stdout}=    Run With Kubeconfig    oc get --raw='/livez'
    Should Be Equal As Strings    ${stdout}    ok    strip_spaces=True

Wait For MicroShift
    [Documentation]    Wait for various checks to ensure MicroShift is online.
    Wait Until Keyword Succeeds    30x    10s
    ...    MicroShift Is Ready
    Wait Until Keyword Succeeds    30x    10s
    ...    MicroShift Is Live
    # We could also wait for relevant pods. Can we restructure the
    # greenboot check script to let us use it even when not on a
    # greenboot host?

MicroShift Service Is Active
    [Documentation]    Check if microshift.service is active
    ...    (process announced readiness)
    ${state}=    Get Systemd Setting    microshift.service    ActiveState
    Should Be Equal As Strings    ${state}    active

Wait For MicroShift Service
    [Documentation]    Wait for MicroShift service to be active
    Wait Until Keyword Succeeds    50x    10s
    ...    MicroShift Service Is Active

Restart MicroShift
    [Documentation]    Restart the MicroShift service
    # Use separate stop and start steps to avoid a race condition with
    # restart where we can't tell when it is down and so we can't tell
    # whether we should count it as back up yet. Forcing the service
    # down, then back up, ensures that when we do things like check
    # systemd settings for the running process we get the values we
    # expect without worrying about the race. Any test that modifies
    # the MicroShift configuration file will be more reliable as a
    # result.
    Systemctl With Retry    stop    microshift.service
    Sleep    10 seconds
    Systemctl With Retry    start    microshift.service
    Wait For MicroShift

Stop MicroShift
    [Documentation]    Stop the MicroShift service
    Systemctl With Retry    stop    microshift.service

Cleanup MicroShift
    [Documentation]    Cleanup
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    echo 1 | sudo microshift-cleanup-data --all
    ...    sudo=True    return_stderr=True    return_rc=True
    Log Many    ${stdout}    ${stderr}    ${rc}
    Should Be Equal As Integers    ${rc}    0
