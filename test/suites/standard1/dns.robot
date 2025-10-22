*** Settings ***
Documentation       Networking smoke tests

Resource            ../../resources/common.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/microshift-network.resource
Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/hosts.resource

Suite Setup         Setup Suite With Namespace
Suite Teardown      Teardown Suite With Namespace

Test Tags           slow


*** Variables ***
${FAKE_LISTEN_IP}           99.99.99.99
${CUSTOM_HOSTS_FILE}        /tmp/hosts23
${PODS_HOSTS_FILE}          /tmp/hosts/hosts
${HOSTNAME}                 ${EMPTY}
${HOSTSFILE_ENABLED}        SEPARATOR=\n
...                         ---
...                         dns:
...                         \ \ hosts:
...                         \ \ \ status: Enabled
${HOSTS_CONFIG_CUSTOM}      SEPARATOR=\n
...                         ---
...                         dns:
...                         \ \ hosts:
...                         \ \ \ status: Enabled
...                         \ \ \ file: ${CUSTOM_HOSTS_FILE}


*** Test Cases ***
Resolve Host from Default Hosts File
    [Documentation]    Resolve host from default hosts file
    [Setup]    Setup With Custom Config    ${HOSTSFILE_ENABLED}    /etc/hosts
    Resolve Host From Pod    ${HOSTNAME}
    [Teardown]    Teardown Hosts File    ${HOSTNAME}

Resolve Host from Non-Default Hosts File
    [Documentation]    Resolve host from default hosts file
    [Setup]    Setup With Custom Config    ${HOSTS_CONFIG_CUSTOM}    ${CUSTOM_HOSTS_FILE}
    Resolve Host From Pod    ${HOSTNAME}
    [Teardown]    Teardown Hosts File    ${HOSTNAME}


*** Keywords ***
Wait For DNS Updated With Hostname
    [Documentation]    Wait for the file to be updated
    [Arguments]    ${hostname}
    Wait Until Keyword Succeeds    2m    2s
    ...    DNS Pod Contain Hosts    ${hostname}

DNS Pod Contain Hosts
    [Documentation]    Check if the hosts file contains the given hostname
    [Arguments]    ${hostname}

    # Find the dns-default-.* pod name in openshift-dns namespace, e.g., dns-default-dqp9n
    ${dns_pod}=    Run With Kubeconfig
    ...    oc get pod -n openshift-dns -o jsonpath='{.items[*].metadata.name}' | grep '^dns-default-'
    ${fuse_device}=    Oc Exec    ${dns_pod}    cat ${PODS_HOSTS_FILE}    openshift-dns
    Should Contain    ${fuse_device}    ${hostname}

Resolve Host From Pod
    [Documentation]    Resolve host from pod
    [Arguments]    ${hostname}
    Create Hello MicroShift Pod
    Expose Hello MicroShift
    Wait For DNS Updated With Hostname    ${hostname}
    Sleep    5 seconds
    ${fuse_device}=    Oc Exec    hello-microshift    nslookup ${hostname}    ${NAMESPACE}    /bin/sh
    Should Contain    ${fuse_device}    Name:    ${hostname}

Setup With Custom Config
    [Documentation]    Install a custom config and restart MicroShift
    [Arguments]    ${config_content}    ${hostsFile}
    ${HOSTNAME}=    Generate Random HostName
    VAR    ${HOSTNAME}=    ${HOSTNAME}    scope=SUITE
    Add Fake IP To NIC    ${FAKE_LISTEN_IP}
    Add Entry To Hosts    ${FAKE_LISTEN_IP}    ${HOSTNAME}    ${hostsFile}
    Drop In MicroShift Config    ${config_content}    20-dns
    Restart MicroShift

Teardown Hosts File
    [Documentation]    Teardown the hosts file
    [Arguments]    ${hostname}
    Run Keywords
    ...    Delete Hello MicroShift Pod And Service
    ...    AND
    ...    Remove Entry From Hosts    ${hostname}
    ...    AND
    ...    Remove Fake IP From NIC    ${FAKE_LISTEN_IP}
    ...    AND
    ...    Remove Drop In MicroShift Config    20-dns

Add Fake IP To NIC
    [Documentation]    Add the given IP to the given NIC temporarily.
    [Arguments]    ${ip_address}=${FAKE_LISTEN_IP}    ${nic_name}=br-ex
    ${stdout}    ${stderr}    ${rc}=    SSHLibrary.Execute Command
    ...    ip address add ${ip_address}/32 dev ${nic_name}
    ...    sudo=True    return_rc=True    return_stderr=True    return_stdout=True
    Log Many    ${stdout}    ${stderr}
    Should Be Equal As Integers    0    ${rc}

Remove Fake IP From NIC
    [Documentation]    Remove the given IP from the given NIC.
    [Arguments]    ${ip_address}=${FAKE_LISTEN_IP}    ${nic_name}=br-ex
    ${stdout}    ${stderr}    ${rc}=    SSHLibrary.Execute Command
    ...    ip address delete ${ip_address}/32 dev ${nic_name}
    ...    sudo=True    return_rc=True    return_stderr=True    return_stdout=True
    Log Many    ${stdout}    ${stderr}
    Should Be Equal As Integers    0    ${rc}
