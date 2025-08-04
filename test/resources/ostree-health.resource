*** Settings ***
Documentation       Keywords for OSTree-based systems

Resource            systemd.resource
Resource            microshift-process.resource
Resource            ostree-data.resource


*** Keywords ***
Wait Until Greenboot Health Check Exited
    [Documentation]    Wait until greenboot healthchecks are done

    Wait Until Keyword Succeeds    10m    15s
    ...    Greenboot Health Check Exited

Greenboot Health Check Exited
    [Documentation]    Checks if greenboot-healthcheck finished running successfully (exited)

    Systemctl Check Service SubState    greenboot-healthcheck.service    exited

Restart Greenboot And Wait For Success
    [Documentation]    Restart the greenboot-healthcheck service and check its status

    VAR    ${unit_name}    greenboot-healthcheck.service

    # Note that the Systemctl keyword from systemd.resource cannot be used to
    # restart the greenboot-healthcheck service due to the keyword expecting
    # the 'running' state after the restart. This condition does not apply on
    # services like greenboot that exit after their startup finishes.
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    systemctl restart ${unit_name}
    ...    sudo=True    return_stdout=True    return_stderr=True    return_rc=True
    IF    ${rc} != 0    Systemctl Print Service Status And Logs    ${unit_name}
    Should Be Equal As Integers    0    ${rc}

    Wait Until Greenboot Health Check Exited
