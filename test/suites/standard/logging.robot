*** Settings ***
Documentation       Tests related to different logging levels

Resource            ../../resources/common.resource
Resource            ../../resources/systemd.resource
Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-process.resource
Library             Collections

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           configuration    etcd    restart    slow


*** Variables ***
${LEVELNORMAL}          SEPARATOR=\n
...                     ---
...                     debugging:
...                     \ \ logLevel: NORMAL

${LEVELdebug}          SEPARATOR=\n
...                     ---
...                     debugging:
...                     \ \ logLevel: debug

${LEVELUNKNOWN}          SEPARATOR=\n
...                     ---
...                     debugging:
...                     \ \ logLevel: Normal
*** Test Cases ***
Test logLevel Normal
    [Documentation]    Set logLevel to NORMAL
    ...
    ...    Test various spellings of the logLevelkeyword
    ...    uppercase, lowercase, Capitol, and camelcase are supported
    [Setup]    Setup With Custom Config    ${LEVELNORMAL}
    Expect logLevel   NORMAL

Test logLevel Debug
    [Documentation]    Set logLevel to debug
    ...
    ...    Test various spellings of the logLevelkeyword
    ...    uppercase, lowercase, Capitol, and camelcase are supported
    [Setup]    Setup With Custom Config    ${LEVELdebug}
    Expect logLevel   debug

Test logLevel UNKNOWN
    [Documentation]    Set logLevel to unknown
    ...
    ...    Test various spellings of the logLevelkeyword
    ...    uppercase, lowercase, Capitol, and camelcase are supported
    [Setup]    Setup With Custom Config    ${LEVELUNKNOWN}
    Expect logLevel   Normal
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
    Upload MicroShift Config    ${merged}
    Restart MicroShift

Expect logLevel
    [Documentation]    Verify that the logLevel matches the expected value
    [Arguments]    ${expected}

    ${stdout}=    Execute Command
    ...    microshift show-config | awk '/logLevel/{print $2}'
    ...    sudo=True    return_rc=False

    Should Be Equal As Strings   ${expected}    ${stdout} 
