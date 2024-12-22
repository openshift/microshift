*** Settings ***
Documentation       show-config command tests

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-host.resource
Resource            ../../resources/microshift-process.resource
Library             ../../resources/DataFormats.py

Suite Setup         Setup
Suite Teardown      Teardown


*** Variables ***
${MEMLIMIT256}      SEPARATOR=\n
...                 ---
...                 etcd:
...                 \ \ memoryLimitMB: 256


*** Test Cases ***
No Sudo Command
    [Documentation]    Test without priviledge elevation
    ${output}    ${rc}=    Execute Command
    ...    microshift show-config
    ...    sudo=False    return_rc=True
    Should Not Be Equal As Integers    0    ${rc}

No Mode Argument
    [Documentation]    Test without any explicit --mode
    ${output}    ${rc}=    Execute Command
    ...    microshift show-config
    ...    sudo=True    return_rc=True
    Should Be Equal As Integers    0    ${rc}
    ${config}=    Yaml Parse    ${output}
    Should Be Equal As Integers    256    ${config.etcd.memoryLimitMB}

Explicit Mode Default
    [Documentation]    Test with explicit '--mode default'
    ${config}=    Show Config    default
    Should Be Equal As Integers    0    ${config.etcd.memoryLimitMB}

Explicit Mode Effective
    [Documentation]    Test with explicit '--mode effective'
    ${config}=    Show Config    effective
    Should Be Equal As Integers    256    ${config.etcd.memoryLimitMB}

Mode Unknown
    [Documentation]    Test with explicit '--mode no-such-mode'
    ${output}    ${rc}=    Execute Command
    ...    microshift show-config --mode no-such-mode
    ...    sudo=True    return_rc=True
    Should Not Be Equal As Integers    0    ${rc}


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host
    Drop In MicroShift Config    ${MEMLIMIT256}    10-etcd

Teardown
    [Documentation]    Test suite teardown
    Remove Drop In MicroShift Config    10-etcd
    Logout MicroShift Host

Show Config
    [Documentation]    Run microshift show-config with ${mode}
    [Arguments]    ${mode}
    ${output}    ${rc}=    Execute Command
    ...    microshift show-config --mode ${mode}
    ...    sudo=True    return_rc=True
    Should Be Equal As Integers    0    ${rc}
    ${yaml_data}=    Yaml Parse    ${output}
    RETURN    ${yaml_data}
