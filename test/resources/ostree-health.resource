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

Wait For Healthy System
    [Documentation]    Waits for greenboot healthchecks and checks health.json
    Wait Until Greenboot Health Check Exited
    System Should Be Healthy

System Should Be Healthy
    [Documentation]    Verifies if system is healthy by looking at health.json.
    ...    If this runs after greenboot-healthcheck finishes, it verifies
    ...    health of current boot.

    # Health of system with MicroShift 4.13 is implicitly checked by "Greenboot Health Check Exited"
    ${version}=    MicroShift Version
    IF    ${version.minor} == 13    RETURN

    ${health}=    Get Persisted System Health
    Should Be Equal As Strings    ${health}    healthy

Restart Greenboot And Wait For Success
    [Documentation]    Restart the greenboot-healthcheck service and check its status
    Systemctl    restart    greenboot-healthcheck.service
    Wait Until Greenboot Health Check Exited
