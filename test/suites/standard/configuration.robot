*** Settings ***
Documentation       Tests for configuration changes

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-process.resource
Library             ../../resources/journalctl.py

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           restart    slow


*** Variables ***
${CURSOR}               ${EMPTY}    # The journal cursor before restarting MicroShift
${BAD_LOG_LEVEL}        SEPARATOR=\n
...                     ---
...                     debugging:
...                     \ \ logLevel: unknown-value
${DEBUG_LOG_LEVEL}      SEPARATOR=\n
...                     ---
...                     debugging:
...                     \ \ logLevel: debug


*** Test Cases ***
Unknown Log Level Produces Warning
    [Documentation]    Logs should warn that the log level setting is unknown
    Setup With Bad Log Level
    Pattern Should Appear In Log Output    ${CURSOR}    Unrecognized log level "unknown-value", defaulting to "Normal"

Debug Log Level Produces No Warning
    [Documentation]    Logs should not warn that the log level setting is unknown
    Setup With Debug Log Level
    Pattern Should Not Appear In Log Output    ${CURSOR}    Unrecognized log level "debug", defaulting to "Normal"


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig    # for readiness checks
    Save Default MicroShift Config
    Save Journal Cursor

Teardown
    [Documentation]    Test suite teardown
    Restore Default MicroShift Config
    Restart MicroShift
    Logout MicroShift Host
    Remove Kubeconfig

Save Journal Cursor
    [Documentation]
    ...    Save the journal cursor then restart MicroShift so we capture the
    ...    shutdown messages and startup messages.
    ${cursor}=    Get Journal Cursor
    Set Suite Variable    \${CURSOR}    ${cursor}

Setup With Bad Log Level
    [Documentation]    Set log level to an unknown value and restart
    ${merged}=    Extend MicroShift Config    ${BAD_LOG_LEVEL}
    Upload MicroShift Config    ${merged}
    Restart MicroShift

Setup With Debug Log Level
    [Documentation]    Set log level to debug and restart
    ${merged}=    Extend MicroShift Config    ${BAD_LOG_LEVEL}
    Upload MicroShift Config    ${merged}
    Restart MicroShift
