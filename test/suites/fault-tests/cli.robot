*** Settings ***
Documentation       Test suite for MicroShift CLI error scenarios based on pkg/cmd code

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-host.resource
Library             OperatingSystem
Library             Process
Library             Collections
Library             String
Library             yaml

Suite Setup         Setup
Suite Teardown      Teardown
Test Template       ${COMMAND} Fault Test ${RETURN_RC} ${NEEDS_SUDO}


*** Variables ***
${INVALID_FLAG}     --invalid-flag


*** Test Cases ***    COMMAND    RETURN_RC    NEEDS_SUDO
Command "microshift" Fault Test    microshift    ${1}    ${FALSE}
Command "microshift backup" Fault Test    microshift backup    ${1}    ${FALSE}
Command "microshift completion" Fault Test    microshift completion    ${0}    ${FALSE}
Command "microshift help" Fault Test    microshift help    ${0}    ${FALSE}
Command "microshift restore" Fault Test    microshift restore    ${1}    ${FALSE}


*** Keywords ***
Setup
    [Documentation]    Set up all of the tests in this suite
    Check Required Env Variables
    Login MicroShift Host

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host

${command} Fault Test ${return_rc} ${needs_sudo}
    [Documentation]    Test CLI behavior for ${command} command in various fault scenarios
    IF    ${needs_sudo} == ${TRUE}
        Command Should Work    ${command}    sudo_mode=${TRUE}
        Command Should Fail    ${command}    sudo_mode=${FALSE}
    ELSE
        Command Should    ${command}    sudo_mode=${TRUE}    expected_rc=${return_rc}
        Command Should    ${command}    sudo_mode=${FALSE}    expected_rc=${return_rc}
    END
    Command Contains    ${command}    ${needs_sudo}    default
    Command Contains    ${command}    ${needs_sudo}    --help
    Command Fails With Invalid Flag    ${command}    ${needs_sudo}

Command Should
    [Documentation]    Run a command and expect it to fail with specified message
    [Arguments]    ${command}    ${error_message}=${EMPTY}    ${sudo_mode}=True    ${expected_rc}=0
    ${stdout}    ${stderr}    ${rc}=    Command Execution    ${command}    sudo_mode=${sudo_mode}
    Should Be Equal As Numbers    ${rc}    ${expected_rc}
    IF    '${error_message}' != '${EMPTY}'
        Should Contain    ${stderr}    ${error_message}
    END

Command Fails With Invalid Flag
    [Documentation]    Test CLI behavior when a command is executed with an invalid flag
    [Arguments]    ${command}    ${sudo_mode}
    Command Should
    ...    ${command} ${INVALID_FLAG}
    ...    error_message=Error: unknown flag: ${INVALID_FLAG}
    ...    sudo_mode=${sudo_mode}
    ...    expected_rc=1

Command Contains
    [Documentation]    Test CLI command contains a specific message
    [Arguments]    ${command}    ${sudo_mode}    ${message_path}
    IF    '${message_path}' != 'default'
        ${stdout}=    Command Execution    ${command} ${message_path}    sudo_mode=${sudo_mode}
    ELSE
        ${stdout}=    Command Execution    ${command}    sudo_mode=${sudo_mode}
    END
    ${expected_message}=    Get CLI Message    ${command}    ${message_path}
    Should Contain    ${stdout}    ${expected_message.rstrip()}

Get CLI Message
    [Documentation]    Get the message for a specific command from a YAML file
    [Arguments]    ${command_path}    ${message_path}
    ${yaml_file}=    OperatingSystem.Get File    suites/fault-tests/cli-messages.yaml
    ${messages_dict}=    yaml.Safe Load    ${yaml_file}
    ${message}=    Set Variable    ${messages_dict}[${command_path}][${message_path}]
    RETURN    ${message}
