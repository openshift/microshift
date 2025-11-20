*** Settings ***
Documentation       Core DNS smoke tests

Resource            ../../resources/common.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/microshift-network.resource
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
    [Setup]    Setup With Custom Config    ${HOSTS_CONFIG_CUSTOM}    ${CUSTOM_HOSTS_FILE}
    Resolve Host From Pod    ${HOSTNAME}
    [Teardown]    Teardown Hosts File    ${HOSTNAME}

Dynamic Hosts File Update Without Restart
    [Documentation]    Verify hosts file changes are reflected without MicroShift or pod restarts
    [Setup]    Setup With Custom Config    ${HOSTS_CONFIG_CUSTOM}    ${CUSTOM_HOSTS_FILE}
    Resolve Host From Pod    ${HOSTNAME}
    ${updated_hostname}=    Generate Random HostName
    Add Entry To Hosts    ${FAKE_LISTEN_IP}    ${updated_hostname}    ${CUSTOM_HOSTS_FILE}
    Resolve Host From Pod    ${updated_hostname}
    [Teardown]    Run Keywords
    ...    Remove Entry From Hosts    ${updated_hostname}    ${CUSTOM_HOSTS_FILE}
    ...    AND    Teardown Hosts File    ${HOSTNAME}

Service DNS Takes Precedence Over Hosts File
    [Documentation]    Verify that Kubernetes service DNS resolution takes precedence over hosts file entries
    [Setup]    Setup Service DNS Precedence
    VAR    ${service_fqdn}=    myservice.${NAMESPACE}.svc.cluster.local
    Wait Until Keyword Succeeds    40x    2s
    ...    Router Should Resolve Hostname    ${service_fqdn}    ${SERVICE_CLUSTER_IP}
    [Teardown]    Teardown Service DNS Precedence Test

Disable CoreDNS Hosts And Verify ConfigMap Removed
    [Documentation]    Enable CoreDNS hosts, then disable it and verify hosts-file configmap is removed
    [Setup]    Setup With Custom Config    ${HOSTSFILE_ENABLED}    /etc/hosts
    Disable CoreDNS Hosts
    Run Keyword And Expect Error    *NotFound*
    ...    Oc Get    configmap/hosts-file -n openshift-dns
    [Teardown]    Teardown Hosts File    ${HOSTNAME}


*** Keywords ***
Resolve Host From Pod
    [Documentation]    Resolve host from pod
    [Arguments]    ${hostname}
    Wait Until Keyword Succeeds    40x    2s
    ...    Router Should Resolve Hostname    ${hostname}

Router Should Resolve Hostname
    [Documentation]    Check if the router pod resolves the given hostname
    [Arguments]    ${hostname}    ${expected_ip}=${EMPTY}
    ${fuse_device}=    Oc Exec    router-default    nslookup ${hostname}    openshift-ingress    deployment
    Should Contain    ${fuse_device}    Name:    ${hostname}
    IF    "${expected_ip}" != ""
        Should Contain    ${fuse_device}    ${expected_ip}
        ...    msg=Hostname should resolve to ${expected_ip}
    END

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

Check CoreDNS Hosts Feature
    [Documentation]    Skip suite if CoreDNS hosts feature is not available
    ${config}=    Show Config    default
    TRY
        VAR    ${hosts}=    ${config}[dns][hosts]
    EXCEPT
        Skip    CoreDNS hosts feature not available in this MicroShift version
    END

Setup Service DNS Precedence
    [Documentation]    Setup to verify service DNS takes precedence over hosts file
    # Create a service named "myservice"
    Create Hello MicroShift Pod    ${NAMESPACE}
    Expose Hello MicroShift    ${NAMESPACE}
    Labeled Pod Should Be Ready    app\=hello-microshift    timeout=120s    ns=${NAMESPACE}
    # Get the actual service cluster IP
    ${service_ip}=    Oc Get JsonPath    service    ${NAMESPACE}    hello-microshift    .spec.clusterIP
    VAR    ${SERVICE_CLUSTER_IP}=    ${service_ip}    scope=TEST
    # Add conflicting entry to hosts file with a different IP
    VAR    ${service_fqdn}=    myservice.${NAMESPACE}.svc.cluster.local
    Add Entry To Hosts    ${FAKE_LISTEN_IP}    ${service_fqdn}    /etc/hosts
    Drop In MicroShift Config    ${HOSTSFILE_ENABLED}    20-dns
    Restart MicroShift

Teardown Service DNS Precedence Test
    [Documentation]    Cleanup service DNS precedence test resources
    Run Keywords
    ...    Remove Entry From Hosts    myservice.${NAMESPACE}.svc.cluster.local
    ...    AND
    ...    Remove Drop In MicroShift Config    20-dns
    ...    AND
    ...    Restart MicroShift
    ...    AND
    ...    Delete Hello MicroShift Pod And Service    ${NAMESPACE}

Disable CoreDNS Hosts
    [Documentation]    Disable CoreDNS hosts feature
    Drop In MicroShift Config    ${HOSTSFILE_DISABLED}    20-dns
    Restart MicroShift
