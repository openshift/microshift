*** Settings ***
Documentation       Keywords for running the microshift command line.

Library             Process
Library             String
Library             OperatingSystem
Library             SSHLibrary
Library             DataFormats.py


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

Clear MicroShift Config
    [Documentation]    Remove any configuration file
    ${stdout}    ${rc}=    Execute Command
    ...    rm -f /etc/microshift/config.yaml
    ...    sudo=True    return_rc=True

Upload MicroShift Config
    [Documentation]    Upload a new configuration file to the MicroShift host
    [Arguments]    ${config_content}
    Upload String To File    ${config_content}    /etc/microshift/config.yaml
