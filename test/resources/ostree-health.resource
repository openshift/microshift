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

Wait For MicroShift Healthcheck Success
    [Documentation]    Verify the health of the deployments by running the
    ...    MicroShift healthcheck command and waiting for it to succeed
    [Arguments]    ${wait_timeout}=600s

    # Note that the newer implementation of greenboot-healthcheck service no
    # longer allows service restart. Using the MicroShift healthcheck command
    # to verify the health of the deployments.
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    microshift healthcheck -v=2 --timeout="${wait_timeout}"
    ...    sudo=True    return_stdout=True    return_stderr=True    return_rc=True
    IF    ${rc} != 0    Log Many    ${stdout}    ${stderr}
    Should Be Equal As Integers    0    ${rc}
