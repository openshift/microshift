*** Settings ***
Documentation       Kubelet image credential provider configuration tests

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-process.resource
Library             ../../resources/journalctl.py

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           slow    restart


*** Variables ***
${CURSOR}                       ${EMPTY}
${CP_DROPIN}                    10-credential-provider
${CP_CONFIGURED_LOG}            Kubelet image credential provider configured
${CP_BIN_DIR}                   /usr/libexec/microshift/credential-providers
${CP_MOCK_PROVIDER}             ${CP_BIN_DIR}/mock-credential-provider
${CP_CONFIG_FILE}               /etc/microshift/credential-providers.yaml
${KUBELET_GENERATED_CONFIG}     /var/lib/microshift/resources/kubelet/config/config.yaml
${CP_VALID}                     SEPARATOR=\n
...                             ---
...                             kubelet:
...                             \ \ imageCredentialProviderConfigPath: ${CP_CONFIG_FILE}
...                             \ \ imageCredentialProviderBinDir: ${CP_BIN_DIR}
${CP_MISSING_BIN_DIR}           SEPARATOR=\n
...                             ---
...                             kubelet:
...                             \ \ imageCredentialProviderConfigPath: ${CP_CONFIG_FILE}
...                             \ \ imageCredentialProviderBinDir: /usr/libexec/microshift/no-such-dir
${CP_ONLY_CONFIG_PATH}          SEPARATOR=\n
...                             ---
...                             kubelet:
...                             \ \ imageCredentialProviderConfigPath: ${CP_CONFIG_FILE}
${CP_PROVIDER_CONFIG}           SEPARATOR=\n
...                             apiVersion: kubelet.config.k8s.io/v1
...                             kind: CredentialProviderConfig
...                             providers:
...                             \ \ - name: mock-credential-provider
...                             \ \ \ \ matchImages:
...                             \ \ \ \ \ \ - "registry.example.invalid"
...                             \ \ \ \ defaultCacheDuration: "1m"
...                             \ \ \ \ apiVersion: credentialprovider.kubelet.k8s.io/v1
${CP_MOCK_SCRIPT}               SEPARATOR=\n
...                             \#!/bin/bash
...                             \# Mock kubelet image credential provider: returns static credentials.
...                             cat >/dev/null
...                             cat <<'JSON'
...                             {"kind":"CredentialProviderResponse",
...                             "apiVersion":"credentialprovider.kubelet.k8s.io/v1",
...                             "cacheKeyType":"Registry","cacheDuration":"1m",
...                             "auth":{"registry.example.invalid":
...                             {"username":"user","password":"pass"}}}
...                             JSON


*** Test Cases ***
Keys Absent Leaves Kubelet Unchanged
    [Documentation]    Without the keys, MicroShift starts and no credential provider is configured
    [Setup]    Run Keywords    Remove Credential Provider Config    AND    Restart MicroShift With Cursor
    Pattern Should Not Appear In Log Output    ${CURSOR}    ${CP_CONFIGURED_LOG}

Valid Configuration Applies Kubelet Flags
    [Documentation]    With both keys set, MicroShift starts, logs the configured paths, reports them in
    ...    show-config, and keeps them out of the generated KubeletConfiguration
    [Setup]    Apply Credential Provider Config    ${CP_VALID}
    Pattern Should Appear In Log Output    ${CURSOR}    ${CP_CONFIGURED_LOG}
    ${config}=    Show Config    effective
    Should Be Equal As Strings    ${config.kubelet.imageCredentialProviderConfigPath}    ${CP_CONFIG_FILE}
    Should Be Equal As Strings    ${config.kubelet.imageCredentialProviderBinDir}    ${CP_BIN_DIR}
    Command Should Fail    grep -q imageCredentialProvider ${KUBELET_GENERATED_CONFIG}
    [Teardown]    Remove Credential Provider Config

Missing Bin Directory Prevents Start
    [Documentation]    MicroShift fails to start when the bin directory does not exist
    [Setup]    Apply Invalid Credential Provider Config    ${CP_MISSING_BIN_DIR}
    Pattern Should Appear In Log Output    ${CURSOR}    imageCredentialProviderBinDir
    Pattern Should Appear In Log Output    ${CURSOR}    does not exist
    [Teardown]    Run Keywords    Remove Credential Provider Config    AND    Restart MicroShift

Only One Key Prevents Start
    [Documentation]    MicroShift fails to start when only one of the two keys is set
    [Setup]    Apply Invalid Credential Provider Config    ${CP_ONLY_CONFIG_PATH}
    Pattern Should Appear In Log Output    ${CURSOR}    must be set together
    [Teardown]    Run Keywords    Remove Credential Provider Config    AND    Restart MicroShift

World Writable Bin Directory Prevents Start
    [Documentation]    MicroShift refuses to start when the bin directory is writable by others
    [Setup]    Run Keywords    Command Should Work    chmod o+w ${CP_BIN_DIR}
    ...    AND    Apply Invalid Credential Provider Config    ${CP_VALID}
    Pattern Should Appear In Log Output    ${CURSOR}    must be owned by root and not writable by group or others
    [Teardown]    Run Keywords    Command Should Work    chmod o-w ${CP_BIN_DIR}
    ...    AND    Remove Credential Provider Config
    ...    AND    Restart MicroShift


*** Keywords ***
Setup
    [Documentation]    Test suite setup: install a mock provider binary and a provider config
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig
    Command Should Work    install -d -o root -g root -m 0755 ${CP_BIN_DIR}
    Upload String To File    ${CP_MOCK_SCRIPT}    ${CP_MOCK_PROVIDER}
    Command Should Work    chmod 0755 ${CP_MOCK_PROVIDER}
    Upload String To File    ${CP_PROVIDER_CONFIG}    ${CP_CONFIG_FILE}

Teardown
    [Documentation]    Remove the drop-in and fixtures, restart MicroShift to restore clean state
    Remove Credential Provider Config
    Command Should Work    rm -rf ${CP_BIN_DIR} ${CP_CONFIG_FILE}
    Restart MicroShift
    Remove Kubeconfig
    Logout MicroShift Host

Restart MicroShift With Cursor
    [Documentation]    Record the journal cursor, then restart MicroShift
    ${cursor}=    Get Journal Cursor
    VAR    ${CURSOR}=    ${cursor}    scope=TEST
    Restart MicroShift

Apply Credential Provider Config
    [Documentation]    Apply a drop-in config and restart MicroShift, recording the journal cursor
    [Arguments]    ${config}
    Remove Drop In MicroShift Config    ${CP_DROPIN}
    Drop In MicroShift Config    ${config}    ${CP_DROPIN}
    Restart MicroShift With Cursor

Apply Invalid Credential Provider Config
    [Documentation]    Apply a drop-in config that should prevent MicroShift from starting
    [Arguments]    ${config}
    Remove Drop In MicroShift Config    ${CP_DROPIN}
    Restart MicroShift
    Drop In MicroShift Config    ${config}    ${CP_DROPIN}
    ${cursor}=    Get Journal Cursor
    VAR    ${CURSOR}=    ${cursor}    scope=TEST
    Run Keyword And Expect Error    0 != 1    Restart MicroShift

Remove Credential Provider Config
    [Documentation]    Remove the credential provider drop-in without restarting.
    ...    The next test's setup restarts MicroShift.
    Remove Drop In MicroShift Config    ${CP_DROPIN}
