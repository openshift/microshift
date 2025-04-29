*** Settings ***
Documentation       Test suite for MicroShift configuration error scenarios

Resource            ../../resources/fault-tests.resource
Resource            ../../resources/microshift-host.resource
Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-process.resource
Library             OperatingSystem
Library             Process
Library             Collections
Library             String
Library             yaml
Library             ../../resources/journalctl.py

Suite Setup         Setup
Suite Teardown      Teardown


*** Variables ***
${CONFIG_PATH}              /etc/microshift/config.yaml
${BACKUP_CONFIG}            /tmp/microshift_config.yaml.bak
${FAULT_TIMEOUT}            60s
${UNSUPPORTED_CIPHER}       SEPARATOR=\n
...                         apiServer:
...                         \ \ tls:
...                         \ \ \ \ cipherSuites:
...                         \ \ \ \ - TLS_INVALID_CIPHER
...                         \ \ \ \ minVersion: VersionTLS12
${INVALID_TLS_VERSION}      SEPARATOR=\n
...                         apiServer:
...                         \ \ tls:
...                         \ \ \ \ cipherSuites:
...                         \ \ \ \ - TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256
...                         \ \ \ \ minVersion: VersionTLSInvalid


*** Test Cases ***
TLS Config: Unsupported Cipher Suite
    [Documentation]    Test behavior when TLS version is set to an unsupported cipher suite
    [Setup]    Setup TLS Configuration    ${UNSUPPORTED_CIPHER}

    Check Systemctl Status Logs    config    invalid_cipher

    [Teardown]    Run Keywords
    ...    Remove TLS Drop In Config
    ...    Restart MicroShift

TLS Config: Invalid TLS Version
    [Documentation]    Test behavior when TLS version is set to an invalid value
    [Setup]    Setup TLS Configuration    ${INVALID_TLS_VERSION}

    Check Systemctl Status Logs    config    invalid_tls_version

    [Teardown]    Run Keywords
    ...    Remove TLS Drop In Config
    ...    Restart MicroShift


*** Keywords ***
Setup TLS Configuration
    [Documentation]    Apply the TLS configuration in the argument
    [Arguments]    ${config}
    Drop In MicroShift Config    ${config}    10-tls
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    systemctl restart microshift.service
    ...    sudo=True    return_stdout=True    return_stderr=True    return_rc=True
    Should Be Equal As Numbers    1    ${rc}

Remove TLS Drop In Config
    [Documentation]    Remove the previously created drop-in config for storage
    Remove Drop In MicroShift Config    10-tls

Setup
    [Documentation]    Set up the test suite
    Check Required Env Variables
    Login MicroShift Host
    Backup Config File

Teardown
    [Documentation]    Clean up after tests
    Restore Config File
    Logout MicroShift Host

Backup Config File
    [Documentation]    Create backup of original config file
    Command Should Work    cp ${CONFIG_PATH} ${BACKUP_CONFIG}    sudo_mode=True

Restore Config File
    [Documentation]    Restore original config file
    Command Should Work    cp ${BACKUP_CONFIG} ${CONFIG_PATH}    sudo_mode=True
    Command Should Work    rm ${BACKUP_CONFIG}    sudo_mode=True
    Command Execution    systemctl restart microshift
    Wait For MicroShift
    All Pods Should Be Running    timeout=300s

Drop In Invalid Config
    [Documentation]    Create an invalid configuration file
    [Arguments]    ${file_name}    ${config_yaml}
    ${config_dict}=    yaml.Safe Load    ${config_yaml}
    ${yaml_str}=    yaml.Dump    ${config_dict}
    Drop In MicroShift Config    ${yaml_str}    ${file_name}

Check Systemctl Status Logs
    [Documentation]    Verify system logs contain expected error messages for configuration errors
    [Arguments]    ${error_type}    ${error_case}
    ${log_text}=    SSHLibrary.Execute Command
    ...    journalctl -u microshift.service -o short | tail -c 32000
    ...    sudo=True    return_stdout=True
    @{expected_lines}=    Get Expected Fault Messages    ${error_type}    ${error_case}
    Compare Output Logs    ${log_text}    @{expected_lines}
