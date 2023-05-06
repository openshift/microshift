*** Settings ***
Library    Process
Library    SSHLibrary
Library    String
Library    OperatingSystem
Resource    ../resources/systemd.resource
Resource    ../resources/microshift-host.resource
Library    ../resources/YAML.py


*** Keywords ***
MicroShift Version
    [Documentation]    Run the version command
    Login MicroShift host
    ${version_text}    ${rc}=    Execute Command
    ...    microshift version -o yaml
    ...    sudo=True    return_rc=True
    Should Be Equal As Integers    ${rc}    0
    Should Not Be Empty    ${version_text}
    Logout MicroShift Host
    ${version}=  Yaml Parse  ${version_text}
    RETURN  ${version}
