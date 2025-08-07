*** Settings ***
Documentation       Keywords for using the openssl command line.
...

Resource            common.resource


*** Keywords ***
Openssl
    [Documentation]    openssl command wrapper
    [Arguments]    ${cmd}    ${cmd2}=${EMPTY}
    ${result}=    Run Process    openssl ${cmd} ${cmd2}    shell=True    # robotcode: ignore
    Should Be Equal As Integers    ${result.rc}    0

Generate CSR Config
    [Documentation]    Generate csr based on config template file
    [Arguments]    ${config_template_file}    ${out_file}
    ${template}=    OperatingSystem.Get File    ${config_template_file}
    ${message}=    Replace Variables    ${template}
    OperatingSystem.Append To File    ${out_file}    ${message}
