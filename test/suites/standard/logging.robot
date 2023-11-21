*** Settings ***
Documentation       Tests related to different logging levels

Resource            ../../resources/common.resource
Resource            ../../resources/systemd.resource
Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-process.resource
Library             Collections
Library             ../../resources/journalctl.py

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           configuration    logging    restart


*** Variables ***
${CURSOR}           ${EMPTY}    # The journal cursor before restarting MicroShift

${LEVELNORMAL}      SEPARATOR=\n
...                 ---
...                 debugging:
...                 \ \ LogLevel: NORMAL

${LEVELDEBUG}       SEPARATOR=\n
...                 ---
...                 debugging:
...                 \ \ LogLevel: debug

${LEVELUNKNOWN}     SEPARATOR=\n
...                 ---
...                 debugging:
...                 \ \ LogLevel: Normal


*** Test Cases ***
Test LogLevel Normal
    [Documentation]    Set LogLevel to NORMAL
    ...
    ...    Test various spellings of the LogLevel keyword
    ...    uppercase, lowercase, Capitol, and camelcase are supported
    [Setup]    Setup With Custom Config    ${LEVELNORMAL}
    Pattern Should Appear In Log Output    ${CURSOR}    NORMAL
    Expect LogLevel    NORMAL

Test LogLevel Debug
    [Documentation]    Set LogLevel to debug
    ...
    ...    Test various spellings of the logLevelkeyword
    ...    uppercase, lowercase, Capitol, and camelcase are supported
    [Setup]    Setup With Custom Config    ${LEVELDEBUG}
    Pattern Should Appear In Log Output    ${CURSOR}    debug
    Expect LogLevel    debug

Test LogLevel UNKNOWN
    [Documentation]    Set logLevel to unknown
    ...
    ...    Test various spellings of the logLevelkeyword
    ...    uppercase, lowercase, Capitol, and camelcase are supported
    [Setup]    Setup With Custom Config    ${LEVELUNKNOWN}
    Pattern Should Appear In Log Output    ${CURSOR}    Normal
    Expect LogLevel    Normal


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig    # for readiness checks
    Save Default MicroShift Config

Teardown
    [Documentation]    Test suite teardown
    Restore Default Config
    Logout MicroShift Host
    Remove Kubeconfig

Restore Default Config
    [Documentation]    Remove any custom config and restart MicroShift
    Restore Default MicroShift Config
    Restart MicroShift

Setup With Custom Config
    [Documentation]    Install a custom config and restart MicroShift
    [Arguments]    ${config_content}
    ${merged}=    Extend MicroShift Config    ${config_content}
    # Save the journal cursor then restart MicroShift so we capture
    # the shutdown messages and startup messages.
    ${cursor}=    Get Journal Cursor
    Set Suite Variable    \${CURSOR}    ${cursor}
    Upload MicroShift Config    ${merged}
    Restart MicroShift

Expect LogLevel
    [Documentation]    Verify that the logLevel matches the expected value
    [Arguments]    ${expected}

    ${output}    ${stdout}=    Execute Command
    ...    microshift show-config
    ...    sudo=True    return_rc=True
    ${config}=    Yaml Parse    ${output}
    Should Be Equal As Strings    ${expected}    ${config.debugging.logLevel}
