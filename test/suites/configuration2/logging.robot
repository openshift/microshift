*** Settings ***
Documentation       Tests for case-insensitive log level parsing

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-process.resource
Library             ../../resources/journalctl.py

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           restart    slow


*** Test Cases ***
Case Insensitive Log Levels
    [Documentation]    MicroShift should accept log levels in any case variation.
    [Template]    MicroShift Should Accept Log Level

    normal
    Debug
    TRACE
    TraceAll


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig

Teardown
    [Documentation]    Test suite teardown
    Remove Drop In MicroShift Config    10-loglevel
    Restart MicroShift
    Logout MicroShift Host
    Remove Kubeconfig

MicroShift Should Accept Log Level
    [Documentation]    Set log level via drop-in config, restart, and verify
    ...    that MicroShift logs the configured level.
    [Arguments]    ${level}

    ${config}=    Catenate    SEPARATOR=\n
    ...    debugging:
    ...    \ \ logLevel: ${level}
    Drop In MicroShift Config    ${config}    10-loglevel

    ${cursor}=    Get Journal Cursor
    Restart MicroShift

    Pattern Should Appear In Log Output    ${cursor}    logLevel.*${level}
