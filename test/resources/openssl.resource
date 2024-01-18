*** Settings ***
Documentation       Keywords for using the openssl command line.
...

Resource            common.resource


*** Keywords ***
Create Random Temp Directory
    [Documentation]    Create a random temp file
    [Arguments]    ${root_path}=/tmp/certs
    ${rand}=    Generate Random String
    ${path}=    Join Path    ${root_path}    ${rand}
    Create Directory    ${path}
    RETURN    ${path}

Openssl
    [Arguments]    ${cmd}    ${cmd2}=
    ${result}=    Run Process    openssl ${cmd} ${cmd2}    shell=True    # robotcode: ignore
    Should Be Equal As Integers    ${result.rc}    0

Generate CSR Config
    [Documentation]    Generate csr based on config file
    [Arguments]    ${tempate_file}    ${output_file}
    ${template}=    OperatingSystem.Get File    ${tempate_file}
    ${message}=    Replace Variables    ${template}
    OperatingSystem.Append To File     ${output_file}   ${message}
