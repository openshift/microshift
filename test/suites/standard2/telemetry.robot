*** Settings ***
Documentation       Tests for Telemetry

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-process.resource
Library             ../../resources/journalctl.py

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           restart    slow


*** Variables ***
# Temporary use of the telemetry production server until staging is ready.
${TELEMETRY_WRITE_ENDPOINT}     https://infogw.api.openshift.com
${ENABLE_TELEMETRY}             SEPARATOR=\n
...                             telemetry:
...                             \ \ status: Enabled
...                             \ \ endpoint: ${TELEMETRY_WRITE_ENDPOINT}
${DISABLE_TELEMETRY}            SEPARATOR=\n
...                             telemetry:
...                             \ \ status: Disabled
${JOURNAL_CURSOR}               ${EMPTY}
${PULL_SECRET}                  /etc/crio/openshift-pull-secret
${PULL_SECRET_METRICS}          /etc/crio/openshift-pull-secret-with-telemetry
${PULL_SECRET_NO_METRICS}       /etc/crio/openshift-pull-secret-without-telemetry


*** Test Cases ***
MicroShift Reports Metrics To Server
    [Documentation]    Check MicroShift is able to send metrics to the telemetry server without errors.
    [Tags]    robot:exclude
    [Setup]    Setup Telemetry Configuration    ${ENABLE_TELEMETRY}    ${PULL_SECRET_METRICS}

    Should Find Metrics Success    Metrics sent successfully
    Should Find Metrics Success    MicroShift telemetry starting, sending first metrics collection.

    [Teardown]    Remove Telemetry Configuration

MicroShift Fails to Report Metrics To Server: Telemetry Disabled
    [Documentation]    Check MicroShift is not able to send metrics to the telemetry server when it is disabled.
    [Tags]    robot:exclude
    [Setup]    Setup Telemetry Configuration    ${DISABLE_TELEMETRY}    ${PULL_SECRET_METRICS}

    Should Find Metrics Success    Telemetry is disabled
    Should Find Metrics Fails    Metrics sent successfully

    [Teardown]    Remove Telemetry Configuration

MicroShift Fails to Report Metrics To Server: Wrong Pull Secret
    [Documentation]    Check MicroShift is not able to send metrics to the telemetry server when the pull secret is wrong.
    [Tags]    robot:exclude    # comm
    [Setup]    Setup Telemetry Configuration    ${ENABLE_TELEMETRY}    ${PULL_SECRET_NO_METRICS}

    Should Find Metrics Success    MicroShift telemetry starting, sending first metrics collection.
    Should Find Metrics Success    Unable to get pull secret: cloud.openshift.com not found
    Should Find Metrics Fails    Metrics sent successfully

    [Teardown]    Remove Telemetry Configuration


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host
    Remove Kubeconfig

Setup Telemetry Configuration
    [Documentation]    Enables the telemetry feature in MicroShift configuration file
    ...    and restarts microshift.service
    [Arguments]    ${telemetry_config}    ${new_pull_secret}
    Configure Pull Secrets    ${new_pull_secret}
    Setup Kubeconfig
    Drop In MicroShift Config    ${telemetry_config}    10-telemetry
    Stop MicroShift
    ${cursor}=    Get Journal Cursor
    Set Test Variable    \${JOURNAL_CURSOR}    ${cursor}
    Start MicroShift

Remove Telemetry Configuration
    [Documentation]    Removes the telemetry feature from MicroShift configuration file
    ...    and restarts microshift.service
    Remove Drop In MicroShift Config    10-telemetry
    Restore Pull Secrets
    Restart MicroShift

Configure Pull Secrets
    [Documentation]    Sets up the pull secrets for the MicroShift cluster.
    [Arguments]    ${new_pull_secret}
    ${rc}=    SSHLibrary.Execute Command
    ...    grep -q cloud.openshift.com ${PULL_SECRET} || sudo ln -sf ${new_pull_secret} ${PULL_SECRET}
    ...    sudo=True
    ...    return_rc=True
    ...    return_stderr=False
    ...    return_stdout=False
    Should Be Equal As Integers    ${rc}    0

Restore Pull Secrets
    [Documentation]    Restores the original pull secrets for the MicroShift cluster if it was changed by the test.
    ${rc}=    SSHLibrary.Execute Command
    ...    test -f ${PULL_SECRET_NO_METRICS} && sudo ln -sf ${PULL_SECRET_NO_METRICS} ${PULL_SECRET} || true
    ...    sudo=True    return_rc=True    return_stderr=False    return_stdout=False
    Should Be Equal As Integers    ${rc}    0

Should Find Metrics Success
    [Documentation]    Logs should contain metrics message
    [Arguments]    ${pattern}
    Wait Until Keyword Succeeds    10x    10s
    ...    Pattern Should Appear In Log Output    ${JOURNAL_CURSOR}    ${pattern}

Should Find Metrics Fails
    [Documentation]    Logs should not contain metrics message
    [Arguments]    ${pattern}
    Pattern Should Not Appear In Log Output    ${JOURNAL_CURSOR}    ${pattern}
