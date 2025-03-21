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
${JOURNAL_CURSOR}               ${EMPTY}
${PULL_SECRET}                  /etc/crio/openshift-pull-secret
${PULL_SECRET_METRICS}          /etc/crio/openshift-pull-secret-with-telemetry
${PULL_SECRET_NO_METRICS}       /etc/crio/openshift-pull-secret-without-telemetry


*** Test Cases ***
MicroShift Reports Metrics To Server
    [Documentation]    Check MicroShift is able to send metrics to the telemetry server without errors.
    [Tags]    robot:exclude
    [Setup]    Setup Telemetry Configuration

    Wait Until Keyword Succeeds    10x    10s
    ...    Should Find Metrics Success

    [Teardown]    Remove Telemetry Configuration


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host
    Configure Pull Secrets
    Setup Kubeconfig

Teardown
    [Documentation]    Test suite teardown
    Restore Pull Secrets
    Restart MicroShift
    Logout MicroShift Host
    Remove Kubeconfig

Setup Telemetry Configuration
    [Documentation]    Enables the telemetry feature in MicroShift configuration file
    ...    and restarts microshift.service
    Drop In MicroShift Config    ${ENABLE_TELEMETRY}    10-telemetry
    Stop MicroShift
    ${cursor}=    Get Journal Cursor
    Set Suite Variable    \${CURSOR}    ${cursor}
    Restart MicroShift

Remove Telemetry Configuration
    [Documentation]    Removes the telemetry feature from MicroShift configuration file
    ...    and restarts microshift.service
    Remove Drop In MicroShift Config    10-telemetry
    Restart MicroShift

Configure Pull Secrets
    [Documentation]    Sets up the pull secrets for the MicroShift cluster.
    ${rc}=    SSHLibrary.Execute Command
    ...    grep -q cloud.openshift.com ${PULL_SECRET} || sudo ln -sf ${PULL_SECRET_METRICS} ${PULL_SECRET}
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
    [Documentation]    Logs should contain metrics success message
    Pattern Should Appear In Log Output    ${CURSOR}    Metrics sent successfully
