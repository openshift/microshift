*** Settings ***
Documentation       Keywords for running the microshift command line.

Library             Process
Library             String
Library             OperatingSystem
Library             SSHLibrary
Library             DataFormats.py
Resource            common.resource


*** Variables ***
${LVMD_VG_OVERRIDE}=    ${EMPTY}


*** Keywords ***
Save Default MicroShift Config
    [Documentation]    Fetch the current config settings and preserve them as the default
    ...
    ...    Sets the suite variable DEFAULT_MICROSHIFT_CONFIG to the text value
    ...    based on the contents of /etc/microshift/config.yaml, if it exists, or
    ...    an empty string if the file does not exist.
    ...
    ...    This keyword is meant to be used from a Setup step.
    ${stdout}    ${rc}=    Execute Command
    ...    cat /etc/microshift/config.yaml
    ...    sudo=True    return_rc=True
    IF    ${rc} == 0
        Set Suite Variable    \${DEFAULT_MICROSHIFT_CONFIG}    ${stdout}
    ELSE
        Set Suite Variable    \${DEFAULT_MICROSHIFT_CONFIG}    ${EMPTY}
    END

Restore Default MicroShift Config
    [Documentation]    Replace the microshift config file with the original defaults.
    ...
    ...    If there was no configuration file originally, delete any file that is there now.
    ${len}=    Get Length    ${DEFAULT_MICROSHIFT_CONFIG}
    IF    ${len} == 0
        # If there was no configuration file to start with, we do not want to create
        # a new one, even if it is empty.
        Clear MicroShift Config
    ELSE
        Upload MicroShift Config    ${DEFAULT_MICROSHIFT_CONFIG}
    END

Extend MicroShift Config
    [Documentation]    Return combination of default config and input argument as a string.
    ...
    ...    The values are parsed as YAML and merged additively (no keys are deleted
    ...    and list values are extended but not replaced) by 'Yaml Merge'.
    [Arguments]    ${config}
    ${merged}=    Yaml Merge    ${DEFAULT_MICROSHIFT_CONFIG}    ${config}
    RETURN    ${merged}

Replace MicroShift Config
    [Documentation]    Return a replacement of the input argument in the default config
    ...    as a string.
    [Arguments]    ${config}
    ${replaced}=    Yaml Replace    ${DEFAULT_MICROSHIFT_CONFIG}    ${config}
    RETURN    ${replaced}

Clear MicroShift Config
    [Documentation]    Remove any configuration file
    ${stdout}    ${rc}=    Execute Command
    ...    rm -f /etc/microshift/config.yaml
    ...    sudo=True    return_rc=True

Upload MicroShift Config
    [Documentation]    Upload a new configuration file to the MicroShift host
    [Arguments]    ${config_content}
    Upload String To File    ${config_content}    /etc/microshift/config.yaml

Drop In MicroShift Config
    [Documentation]    Upload a drop-in configuration file to the MicroShift host
    [Arguments]    ${config_content}    ${name}
    Upload String To File    ${config_content}    /etc/microshift/config.d/${name}.yaml

Remove Drop In MicroShift Config
    [Documentation]    Remove a drop-in configuration file from MicroShift host
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
    ${cfg}=    OperatingSystem.Get File    ./assets/lvmd.yaml
    ${lvmd_vg}=    Set Variable If    '${LVMD_VG_OVERRIDE}' != ''    ${LVMD_VG_OVERRIDE}    rhel
    ${templated}=    Replace Variables    ${cfg}
    ${merged_cfg}=    Lvmd Merge    ${DEFAULT_LVMD_CONFIG}    ${templated}
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
