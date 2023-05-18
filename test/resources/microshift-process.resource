*** Settings ***
Documentation  Keywords for running the microshift command line.

Library    Process
Library    String
Library    OperatingSystem

Library    SSHLibrary
Resource    ../resources/microshift-host.resource
Library    ../resources/YAML.py


*** Keywords ***
MicroShift Version
    [Documentation]    Run the version command
    ${version_text}    ${rc}=    Execute Command
    ...    microshift version -o yaml
    ...    sudo=True    return_rc=True
    Should Be Equal As Integers    ${rc}    0
    Should Not Be Empty    ${version_text}
    ${version}=  Yaml Parse  ${version_text}
    RETURN  ${version}
