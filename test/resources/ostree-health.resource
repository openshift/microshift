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

    ${value}=    Get Systemd Setting    greenboot-healthcheck.service    SubState
    Should Be Equal As Strings    ${value}    exited

Restart Greenboot And Wait For Success
    [Documentation]    Restart the greenboot-healthcheck service and check its status
    Systemctl    restart    greenboot-healthcheck.service
    Wait Until Greenboot Health Check Exited
