*** Settings ***
Documentation       Tests related to upgrading MicroShift

Resource            ../resources/common.resource
Resource            ../resources/ostree.resource
Library             Collections

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           ostree


*** Variables ***
${USHIFT_HOST}      ${EMPTY}
${USHIFT_USER}      ${EMPTY}

${FAILING_REF}      ${EMPTY}
${TARGET_REF}       ${EMPTY}


*** Test Cases ***
FIDO Onboarding Device
    [Documentation]    No-MicroShift-system is staged with unhealthy deployment
    ...    which rolls back and leaves stale data and staged again with healthy deployment.
    ...
    ...    It is expected that final deployment will gracefully handle existence of:
    ...    MicroShift data, unhealthy stored in health file,
    ...    and a deployment gap (no-microshift rollback ).

    System Should Not Feature MicroShift
    ${initial_deploy_id}=    Get Booted Deployment Id
    Deploy Commit Expecting A Rollback    ${initial_deploy_id}
    Deploy Commit Not Expecting A Rollback    ${initial_deploy_id}


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    IF    "${TARGET_REF}"=="${EMPTY}"
        Fatal Error    TARGET_REF variable is required
    END
    IF    "${FAILING_REF}"=="${EMPTY}"
        Fatal Error    FAILING_REF variable is required
    END
    Login MicroShift Host

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host

System Should Not Feature MicroShift
    [Documentation]    Check if system contains MicroShift binary, service, or data
    SSHLibrary.Directory Should Not Exist    ${BACKUP_STORAGE}
    SSHLibrary.Directory Should Not Exist    ${DATA_DIR}
    SSHLibrary.File Should Not Exist    /usr/bin/microshift

Deploy Commit Expecting A Rollback
    [Documentation]    Deploys given ref and configures test agent for failing greenboot.
    ...    It expects the system to roll back.
    [Arguments]    ${initial_deploy_id}

    ${deploy_id}=    Rebase System    ${FAILING_REF}
    ${cfg}=    Evaluate    json.dumps({"${deploy_id}" : {"every": ["fail_greenboot"]}})
    Create Agent Config    ${cfg}
    Reboot MicroShift Host

    Log To Console    "Failing ref deployed - waiting for system to roll back"
    Wait Until Keyword Succeeds    20m    15s
    ...    Check If Current Deployment Is    ${initial_deploy_id}
    Log To Console    "System rolled back - deploying target ref"

Deploy Commit Not Expecting A Rollback
    [Documentation]    Deploys given ref and configures test agent for failing greenboot.
    ...    It expects the system to roll back.
    [Arguments]    ${initial_deploy_id}

    ${deploy_id}=    Rebase System    ${TARGET_REF}
    Reboot MicroShift Host

    Log To Console    "Target ref deployed - starting health checking"
    Wait Until Keyword Succeeds    10m    15s
    ...    System Is Running Right Ref And Healthy    ${deploy_id}    ${initial_deploy_id}

Make New SSH Connection
    [Documentation]    Closes all SSH connections and makes a new one.
    # Staging deployments potentially introduces multiple reboots
    # which could break existing SSH connection

    SSHLibrary.Close All Connections
    Login MicroShift Host

Check If Current Deployment Is
    [Documentation]    Checks if currently booted deployment is as expected
    [Arguments]    ${expected_deploy}

    Make New SSH Connection

    ${current_deploy}=    libostree.Get Booted Deployment Id
    Should Be Equal As Strings    ${expected_deploy}    ${current_deploy}

System Is Running Right Ref And Healthy
    [Documentation]    Checks if system is running right reference and is healthy
    [Arguments]    ${expected_deploy}    ${initial_deploy}

    Make New SSH Connection

    ${current_deploy}=    libostree.Get Booted Deployment Id
    IF    "${current_deploy}" == "${initial_deploy}"
        Fatal Error    "System rolled back to initial deployment"
    END

    Should Be Equal As Strings    ${expected_deploy}    ${current_deploy}
    Wait Until Greenboot Health Check Exited
    System Should Be Healthy
