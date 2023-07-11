*** Settings ***
Documentation       Keywords for OSTree-based systems

Resource            systemd.resource
Resource            microshift-process.resource
Resource            ostree-data.resource
Resource            ostree-health.resource
Library             libostree.py
Library             ../resources/TestAgent.py


*** Keywords ***
Make New SSH Connection
    [Documentation]    Closes all SSH connections and makes a new one.
    # Staging deployments potentially introduces multiple reboots
    # which could break existing SSH connection

    SSHLibrary.Close All Connections
    Login MicroShift Host

Current Deployment Should Be
    [Documentation]    Checks if currently booted deployment is as expected
    [Arguments]    ${expected_deploy}

    Make New SSH Connection

    ${current_deploy}=    libostree.Get Booted Deployment Id
    Should Be Equal As Strings    ${expected_deploy}    ${current_deploy}

Deploy Commit Expecting A Rollback
    [Documentation]    Deploys given ref and configures test agent for failing greenboot.
    ...    It expects the system to roll back.
    [Arguments]    ${ref}

    ${initial_deploy_id}=    Get Booted Deployment Id
    ${deploy_id}=    Rebase System    ${ref}
    TestAgent.Write
    Write Greenboot Microshift Wait Timeout    60
    Reboot MicroShift Host

    Log To Console    "Failing ref deployed - waiting for system to roll back"
    Wait Until Keyword Succeeds    20m    15s
    ...    Current Deployment Should Be    ${initial_deploy_id}

    Log To Console    "System rolled back"
    Remove Greenboot Microshift Wait Timeout

Deploy Commit Not Expecting A Rollback
    [Documentation]    Deploys given ref and configures test agent for failing greenboot.
    ...    It expects the system to roll back.
    [Arguments]    ${target_ref}

    ${initial_deploy_id}=    Get Booted Deployment Id
    ${deploy_id}=    Rebase System    ${target_ref}
    Reboot MicroShift Host

    Log To Console    "Target ref deployed - starting health checking"
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
    System Should Be Healthy

Mask Grub Boot Success Timer
    [Documentation]    Effectively disable a timer that sets boot_success
    ...    to 1 after two minutes from user login.
    ...    It impacts tests because grub's script to decrement boot_counter
    ...    works only if boot_success is 0 (see /etc/grub.d/08_fallback_counting)
    ...    resulting in more than requested amount of redboot induced reboots
    ...    and system needing much more time to roll back.
    ${out}    ${err}    ${rc}=    Execute Command
    ...    systemctl --user mask grub-boot-success.timer
    ...    return_stdout=True
    ...    return_stderr=True
    ...    return_rc=True

    Should Be Equal As Integers    ${rc}    0
