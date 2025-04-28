*** Settings ***
Documentation       Keywords common to fault test suites

Library             OperatingSystem
Library             String
Library             yaml


*** Keywords ***
Compare Output Logs
    [Documentation]    Compares actual log output against expected patterns
    [Arguments]    ${actual_str}    @{expected_lines}
    FOR    ${pattern}    IN    @{expected_lines}
        ${pattern}=    Strip String    ${pattern}
        Should Match Regexp    ${actual_str}    ${pattern}
    END

Get Expected Messages
    [Documentation]    Get the usage message for a specific command from the YAML file
    [Arguments]    ${path}    ${action}    ${resource}
    ${yaml_file}=    OperatingSystem.Get File    ${path}
    ${messages_dict}=    yaml.Safe Load    ${yaml_file}
    ${messages}=    Set Variable    ${messages_dict}[${action}][${resource}]
    @{message_lines}=    Split To Lines    ${messages}
    RETURN    @{message_lines}
