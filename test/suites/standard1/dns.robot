*** Settings ***
Documentation       Core DNS smoke tests

Resource            ../../resources/common.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/microshift-network.resource
Resource            ../../resources/dns.resource
Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/hosts.resource

Suite Setup         Run Keywords
...                     Setup Suite With Namespace
...                     AND    Check CoreDNS Hosts Feature
Suite Teardown      Teardown Suite With Namespace

Test Tags           slow


*** Variables ***
${FAKE_LISTEN_IP}           99.99.99.99
${CUSTOM_HOSTS_FILE}        /tmp/hosts23
${PODS_HOSTS_FILE}          /tmp/hosts/hosts
${HOSTNAME}                 ${EMPTY}
${SYNC_FREQUENCY}           ${EMPTY}
${HOSTSFILE_ENABLED}        SEPARATOR=\n
...                         ---
...                         dns:
...                         \ \ hosts:
...                         \ \ \ status: Enabled
${HOSTSFILE_DISABLED}       SEPARATOR=\n
...                         ---
...                         dns:
...                         \ \ hosts:
...                         \ \ \ status: Disabled


*** Test Cases ***
Resolve Host from Default Hosts File
    [Documentation]    Resolve host from default hosts file
    [Setup]    Setup With Custom Config    ${HOSTSFILE_ENABLED}    /etc/hosts
    Resolve Host From Pod    ${HOSTNAME}
    [Teardown]    Teardown Hosts File    ${HOSTNAME}

Resolve Host from Non-Default Hosts File
    [Documentation]    Resolve host from default hosts file
    [Setup]    Setup With Custom Hosts File
    Resolve Host From Pod    ${HOSTNAME}
    [Teardown]    Teardown Hosts File    ${HOSTNAME}

Dynamic Hosts File Update Without Restart
    [Documentation]    Verify hosts file changes are reflected without MicroShift or pod restarts
    [Setup]    Setup With Custom Hosts File
    Resolve Host From Pod    ${HOSTNAME}
    ${updated_hostname}=    Generate Random HostName
    Add Entry To Hosts    ${FAKE_LISTEN_IP}    ${updated_hostname}    ${CUSTOM_HOSTS_FILE}
    Resolve Host From Pod    ${updated_hostname}
    [Teardown]    Run Keywords
    ...    Remove Entry From Hosts    ${updated_hostname}    ${CUSTOM_HOSTS_FILE}
    ...    AND    Teardown Hosts File    ${HOSTNAME}

Disable CoreDNS Hosts And Verify ConfigMap Removed
    [Documentation]    Enable CoreDNS hosts, then disable it and verify hosts-file configmap is removed
    [Setup]    Setup With Custom Config    ${HOSTSFILE_ENABLED}    /etc/hosts
    Disable CoreDNS Hosts
    Run Keyword And Expect Error    1 != 0
    ...    Oc Get    configmap    openshift-dns    hosts-file
    [Teardown]    Teardown Hosts File    ${HOSTNAME}


*** Keywords ***
Get Hosts Config Custom
    [Documentation]    Build hosts config with optional syncFrequency
    ...    syncFrequency is configurable in order to speed up the ConfigMap synchronization time
    ...    for the pods that mount it.
    IF    "${SYNC_FREQUENCY}" != "${EMPTY}"
        ${config}=    Catenate    SEPARATOR=\n
        ...    ---
        ...    kubelet:
        ...    \ \ syncFrequency: ${SYNC_FREQUENCY}
        ...    dns:
        ...    \ \ hosts:
        ...    \ \ \ status: Enabled
        ...    \ \ \ file: ${CUSTOM_HOSTS_FILE}
    ELSE
        ${config}=    Catenate    SEPARATOR=\n
        ...    ---
        ...    dns:
        ...    \ \ hosts:
        ...    \ \ \ status: Enabled
        ...    \ \ \ file: ${CUSTOM_HOSTS_FILE}
    END
    RETURN    ${config}

Setup With Custom Config
    [Documentation]    Install a custom config and restart MicroShift
    [Arguments]    ${config_content}    ${hostsFile}
    ${HOSTNAME}=    Generate Random HostName
    VAR    ${HOSTNAME}=    ${HOSTNAME}    scope=SUITE
    Add Fake IP To NIC    ${FAKE_LISTEN_IP}
    Add Entry To Hosts    ${FAKE_LISTEN_IP}    ${HOSTNAME}    ${hostsFile}
    Drop In MicroShift Config    ${config_content}    20-dns
    Restart MicroShift

Setup With Custom Hosts File
    [Documentation]    Get custom hosts config and setup with it
    ${config}=    Get Hosts Config Custom
    Setup With Custom Config    ${config}    ${CUSTOM_HOSTS_FILE}

Teardown Hosts File
    [Documentation]    Teardown the hosts file
    [Arguments]    ${hostname}
    Run Keywords
    ...    Remove Entry From Hosts    ${hostname}
    ...    AND
    ...    Remove Fake IP From NIC    ${FAKE_LISTEN_IP}
    ...    AND
    ...    Remove Drop In MicroShift Config    20-dns

Check CoreDNS Hosts Feature
    [Documentation]    Skip suite if CoreDNS hosts feature is not available
    ${config}=    Show Config    default
    TRY
        VAR    ${hosts}=    ${config}[dns][hosts]
    EXCEPT
        Skip    CoreDNS hosts feature not available in this MicroShift version
    END

Disable CoreDNS Hosts
    [Documentation]    Disable CoreDNS hosts feature
    Drop In MicroShift Config    ${HOSTSFILE_DISABLED}    20-dns
    Restart MicroShift
