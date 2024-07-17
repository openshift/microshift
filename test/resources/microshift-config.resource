*** Settings ***
Documentation       Keywords for running the microshift command line.

Library             Process
Library             String
Library             OperatingSystem
Library             SSHLibrary
Library             DataFormats.py
Resource            common.resource


*** Keywords ***
Clear MicroShift Config
    [Documentation]    Remove any configuration file
    ${stdout}    ${rc}=    Execute Command
    ...    rm -f /etc/microshift/config.yaml
    ...    sudo=True    return_rc=True

Drop In MicroShift Config
    [Documentation]    Upload a new configuration file to the MicroShift host
    [Arguments]    ${config_content}    ${name}
    Upload String To File    ${config_content}    /etc/microshift/config.d/${name}.yaml

Remove Drop In MicroShift Config
    [Documentation]    Upload a new configuration file to the MicroShift host
    [Arguments]    ${name}
    ${stdout}    ${rc}=    Execute Command
    ...    rm -f /etc/microshift/config.d/${name}.yaml
    ...    sudo=True    return_rc=True

Save Lvmd Config
    [Documentation]    If an lvmd.yaml file already exists, preserver it
    ${stdout}    ${rc}=    Execute Command
    ...    cat /etc/microshift/lvmd.yaml
    ...    sudo=True    return_rc=True
    IF    ${rc} == 0
        Set Suite Variable    \${DEFAULT_LVMD_CONFIG}    ${stdout}
    ELSE
        Set Suite Variable    \${DEFAULT_LVMD_CONFIG}    ${EMPTY}
    END

Extend Lvmd Config
    [Documentation]    Combines the test's device-classes with the remote's lvmd.yaml.
    ${patch_cfg}=    OperatingSystem.Get File    ./assets/lvmd.yaml
    ${merged_cfg}=    Lvmd Merge    ${DEFAULT_LVMD_CONFIG}    ${patch_cfg}
    RETURN    ${merged_cfg}

Restore Lvmd Config
    [Documentation]    Replace the lvmd config file with the original defaults.
    ${len}=    Get Length    ${DEFAULT_LVMD_CONFIG}
    IF    ${len} == 0
        Clear Lvmd Config
    ELSE
        Upload Lvmd Config    ${DEFAULT_LVMD_CONFIG}
    END

Upload Lvmd Config
    [Documentation]    Upload a test's lvmd.yaml file to the MicroShift host
    [Arguments]    ${config_content}
    Upload String To File    ${config_content}    /etc/microshift/lvmd.yaml

Clear Lvmd Config
    [Documentation]    Removes the LVMD configuration as part of restoring test environment
    ${stderr}    ${rc}=    sshLibrary.Execute Command    rm -f /etc/microshift/lvmd.yaml
    ...    sudo=True    return_rc=True    return_stderr=True    return_stdout=False
    Log    ${stderr}
    Should Be Equal As Integers    0    ${rc}
