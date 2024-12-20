*** Settings ***
Documentation       Keywords for OSTree-based systems

Resource            systemd.resource
Resource            microshift-process.resource
Resource            ostree-data.resource
Resource            ostree-health.resource
Resource            microshift-host.resource
Library             libostree.py
Library             ../resources/TestAgent.py


*** Keywords ***
Current Deployment Should Be
    [Documentation]    Checks if currently booted deployment is as expected
    [Arguments]    ${expected_deploy}

    Make New SSH Connection

    ${current_deploy}=    libostree.Get Booted Deployment Id
    Should Be Equal As Strings    ${expected_deploy}    ${current_deploy}

Deploy Commit
    [Documentation]    Deploys the given reference on ostree or bootc systems.
    ...    When the '${registry_url}' argument is not empty, the function assumes
    ...    that a bootc deployment needs to be performed.
    [Arguments]    ${ref}    ${write_agent_cfg}    ${registry_url}

    # Retry on rebasing is an attempt to work around the issue described in
    # https://github.com/ostreedev/ostree-rs-ext/issues/657
    IF    "${registry_url}"!="${EMPTY}"
        # Allow insecure registry access when pulling the container image to be used
        Write Insecure Registry Url    ${registry_url}
        ${deploy_id}=    Wait Until Keyword Succeeds    3x    5s
        ...    Rebase Bootc System    ${registry_url}/${ref}
        Remove Insecure Registry Url
    ELSE
        ${deploy_id}=    Wait Until Keyword Succeeds    3x    5s
        ...    Rebase System    ${ref}
    END
    IF    ${write_agent_cfg}    TestAgent.Write

    RETURN    ${deploy_id}

Deploy Commit Expecting A Rollback
    [Documentation]    Deploys given ref and configures test agent for failing greenboot.
    ...    It expects the system to roll back. The '${registry_url}' optional
    ...    argument should be provided on bootc systems.
    [Arguments]    ${ref}    ${write_agent_cfg}=${TRUE}    ${registry_url}=${EMPTY}

    ${initial_deploy_id}=    Get Booted Deployment Id
    ${deploy_id}=    Deploy Commit    ${ref}    ${write_agent_cfg}    ${registry_url}

    Write Greenboot Microshift Wait Timeout    90
    Reboot MicroShift Host

    Log To Console    "System rebased to ${ref} and rebooted - waiting for system to roll back"
    Wait Until Keyword Succeeds    20m    15s
    ...    Current Deployment Should Be    ${initial_deploy_id}
    Log To Console    "System rolled back"

    [Teardown]    Remove Greenboot Microshift Wait Timeout

Deploy Commit Not Expecting A Rollback
    [Documentation]    Deploys given ref and configures test agent for failing greenboot.
    ...    It does not expect the system to roll back. The '${registry_url}' optional
    ...    argument should be provided on bootc systems.
    [Arguments]    ${ref}    ${write_agent_cfg}=${FALSE}    ${registry_url}=${EMPTY}

    ${initial_deploy_id}=    Get Booted Deployment Id
    ${deploy_id}=    Deploy Commit    ${ref}    ${write_agent_cfg}    ${registry_url}
    Reboot MicroShift Host

    Log To Console    "System rebased to ${ref} and rebooted - starting health checking"
    Wait Until Keyword Succeeds    10m    15s
    ...    System Is Running Right Ref And Healthy    ${deploy_id}    ${initial_deploy_id}

System Is Running Right Ref And Healthy
    [Documentation]    Checks if system is running right reference and is healthy
    [Arguments]    ${expected_deploy}    ${initial_deploy}

    Make New SSH Connection

    ${current_deploy}=    libostree.Get Booted Deployment Id
    IF    "${current_deploy}" == "${initial_deploy}"
        Fatal Error    "System rolled back to initial deployment"
    END

    Should Be Equal As Strings    ${expected_deploy}    ${current_deploy}
    Greenboot Health Check Exited

Wait For Transaction To End
    [Documentation]    Wait for any ostree transaction to end.
    ...    When grub boots previous deployment due to greenboot failure,
    ...    ostree status is updated by greenboot running `rpm-ostree rollback`,
    ...    so test must wait until that transaction is over before staging
    ...    new deployment.

    Wait Until Keyword Succeeds    2m    15s
    ...    No Transaction In Progress

Journal Should Have Information About Failed Version Comparison
    [Documentation]    Assert that unhealthy deployment's journal contains
    ...    information about failed version compatibility check

    FOR    ${boot}    IN RANGE    -3    0
        ${stdout}    ${rc}=    Execute Command
        ...    journalctl --unit=microshift --boot=${boot} | grep "FAIL version compatibility checks"
        ...    sudo=True
        ...    return_stdout=True
        ...    return_rc=True

        Log Many    ${stdout}    ${rc}
        Should Be Equal As Integers    0    ${rc}
    END

Expected Boot Count
    [Documentation]    Validate that the host rebooted only the specified number of times
    [Arguments]    ${reboot_count}

    ${stdout}=    Execute Command
    ...    journalctl --list-boots --quiet | wc -l
    ...    sudo=True    return_rc=False

    Should Be Equal As Integers    ${reboot_count}    ${stdout}

Create Usr Directory Overlay
    [Documentation]    Make /usr dir writable by creating an overlay, rebooting
    ...    will go back to being immutable.

    ${is_bootc}=    Is System Bootc
    IF    ${is_bootc}
        ${cmd}=    Set Variable    bootc usr-overlay
    ELSE
        ${cmd}=    Set Variable    rpm-ostree usroverlay
    END

    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    ${cmd}
    ...    sudo=True    return_rc=True
    ...    return_stdout=True    return_stderr=True
    Should Be Equal As Integers    0    ${rc}
