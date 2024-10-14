*** Settings ***
Documentation       Keywords for running the microshift command line.

Library             Process
Library             String
Library             OperatingSystem
Library             SSHLibrary
Resource            oc.resource
Resource            systemd.resource
Resource            microshift-host.resource
Resource            microshift-etcd-process.resource
Resource            ostree-health.resource
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
    [Arguments]    ${extra_args}=${EMPTY}
    ${stdout}=    Run With Kubeconfig    oc get --raw='/livez' ${extra_args}
    Should Be Equal As Strings    ${stdout}    ok    strip_spaces=True

Wait For MicroShift
    [Documentation]    Wait for various checks to ensure MicroShift is online.
    Wait For MicroShift Service
    Wait Until Keyword Succeeds    30x    10s
    ...    MicroShift Is Ready
    Wait Until Keyword Succeeds    30x    10s
    ...    MicroShift Is Live
    # We could also wait for relevant pods. Can we restructure the
    # greenboot check script to let us use it even when not on a
    # greenboot host?

Wait For Greenboot Health Check To Exit
    [Documentation]    Wait for the Greenboot Health Check systemd service to exit
    Wait Until Keyword Succeeds    10m    15s
    ...    Greenboot Health Check Exited

MicroShift Process ID
    [Documentation]    Return the current MicroShift process ID
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    pidof microshift
    ...    return_stdout=True    return_stderr=True    return_rc=True
    Log    ${stderr}
    RETURN    ${stdout}

MicroShift Process ID Changed
    [Documentation]    Compare the specified and current MicroShift process ID
    [Arguments]    ${old_pid}
    ${cur_pid}=    MicroShift Process ID
    Should Not Be Equal    ${cur_pid}    ${old_pid}

Wait Until MicroShift Process ID Changes
    [Documentation]    Wait until the current MicroShift process ID changes
    [Arguments]    ${old_pid}

    Wait Until Keyword Succeeds    30x    10s
    ...    MicroShift Process ID Changed    ${old_pid}

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
    ${ushift_pid}=    MicroShift Process ID
    ${ushift_etcd_pid}=    MicroShift Etcd Process ID
    Systemctl    restart    microshift.service
    Wait Until MicroShift Process ID Changes    ${ushift_pid}
    Wait Until MicroShift Etcd Process ID Changes    ${ushift_etcd_pid}
    # Kubeconfig may change between restarts, ensure we are taking the latest one.
    Setup Kubeconfig
    Wait For MicroShift

Stop MicroShift
    [Documentation]    Stop the MicroShift service
    Systemctl    stop    microshift.service

Start MicroShift
    [Documentation]    Start the MicroShift service
    Systemctl    start    microshift.service
    # Kubeconfig may change between restarts, ensure we are taking the latest one.
    Setup Kubeconfig

Enable MicroShift
    [Documentation]    Enable the MicroShift service
    Systemctl    enable    microshift.service

Disable MicroShift
    [Documentation]    Disable the MicroShift service
    Systemctl    disable    microshift.service

Cleanup MicroShift
    [Documentation]    Cleanup MicroShift data
    [Arguments]    ${cmd}="--all"    ${opt}=${EMPTY}

    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    echo 1 | sudo microshift-cleanup-data ${cmd} ${opt}
    ...    return_stdout=True    return_stderr=True    return_rc=True
    Log Many    ${stdout}    ${stderr}    ${rc}
    Should Be Equal As Integers    ${rc}    0

    Enable MicroShift
