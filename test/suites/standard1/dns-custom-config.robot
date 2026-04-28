*** Settings ***
Documentation       Tests for custom DNS configuration (dns.configFile)

Resource            ../../resources/common.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/microshift-network.resource
Resource            ../../resources/hosts.resource

Suite Setup         Setup Suite With Namespace
Suite Teardown      Teardown Suite With Namespace

Test Tags           slow


*** Variables ***
${FAKE_LISTEN_IP}           99.99.99.98
${CUSTOM_COREFILE_PATH}     /etc/microshift/dns/Corefile
${COREFILE_TEMPLATE}        ./assets/dns/Corefile.template
${HOSTNAME}                 ${EMPTY}

${CUSTOM_DNS_CONFIG}        SEPARATOR=\n
...                         ---
...                         dns:
...                         \ \ configFile: /etc/microshift/dns/Corefile


*** Test Cases ***
Custom DNS Config File Is Used By CoreDNS
    [Documentation]    Provide a valid custom Corefile and verify CoreDNS uses it
    ...    instead of the default template-rendered configuration.
    [Setup]    Setup Custom Corefile With Hosts Entry
    Verify ConfigMap Uses Custom Corefile
    Resolve Host From Pod    ${HOSTNAME}
    [Teardown]    Teardown Custom Corefile

Runtime Reload Without MicroShift Restart
    [Documentation]    Modify the custom Corefile while MicroShift is running
    ...    and verify CoreDNS picks up the change without restart.
    [Setup]    Setup Custom Corefile With Hosts Entry
    Resolve Host From Pod    ${HOSTNAME}
    ${updated_hostname}=    Generate Random HostName
    Update Corefile With New Host    ${updated_hostname}
    Resolve Host From Pod    ${updated_hostname}
    [Teardown]    Teardown Custom Corefile

Cluster Local Resolution With Custom Corefile
    [Documentation]    Verify that a custom Corefile with the kubernetes plugin
    ...    still resolves cluster-local services correctly.
    [Setup]    Setup Custom Corefile With Hosts Entry
    Resolve Host From Pod    kubernetes.default.svc.cluster.local
    [Teardown]    Teardown Custom Corefile


*** Keywords ***
Build Custom Corefile
    [Documentation]    Build a Corefile from the template file, substituting
    ...    the hostname and IP placeholders.
    [Arguments]    ${hostname}    ${ip}
    ${template}=    OperatingSystem.Get File    ${COREFILE_TEMPLATE}
    VAR    ${COREFILE_HOSTNAME}=    ${hostname}
    VAR    ${COREFILE_HOST_IP}=    ${ip}
    ${corefile}=    Replace Variables    ${template}
    RETURN    ${corefile}

Setup Custom Corefile With Hosts Entry
    [Documentation]    Create a custom Corefile with a fake hostname entry,
    ...    upload it, configure MicroShift to use it, and restart.
    ${HOSTNAME}=    Generate Random HostName
    VAR    ${HOSTNAME}=    ${HOSTNAME}    scope=SUITE
    Add Fake IP To NIC    ${FAKE_LISTEN_IP}
    ${corefile}=    Build Custom Corefile    ${HOSTNAME}    ${FAKE_LISTEN_IP}
    Create Remote Dir For Path    ${CUSTOM_COREFILE_PATH}
    Upload String To File    ${corefile}    ${CUSTOM_COREFILE_PATH}
    Drop In MicroShift Config    ${CUSTOM_DNS_CONFIG}    20-dns-custom
    Restart MicroShift

Update Corefile With New Host
    [Documentation]    Add a new hosts entry to the custom Corefile without
    ...    restarting MicroShift. The file watcher and CoreDNS reload plugin
    ...    should pick up the change.
    [Arguments]    ${hostname}
    ${corefile}=    Build Custom Corefile    ${hostname}    ${FAKE_LISTEN_IP}
    Upload String To File    ${corefile}    ${CUSTOM_COREFILE_PATH}

Verify ConfigMap Uses Custom Corefile
    [Documentation]    Verify the dns-default ConfigMap contains custom content
    ...    rather than the default template-rendered Corefile.
    Wait Until Keyword Succeeds    20x    5s
    ...    ConfigMap Should Contain Hostname    ${HOSTNAME}

ConfigMap Should Contain Hostname
    [Documentation]    Check the dns-default ConfigMap Corefile data for the hostname
    [Arguments]    ${hostname}
    ${corefile_data}=    Oc Get JsonPath
    ...    configmap    openshift-dns    dns-default    .data.Corefile
    Should Contain    ${corefile_data}    ${hostname}

Resolve Host From Pod
    [Documentation]    Verify DNS resolution from a pod using nslookup
    [Arguments]    ${hostname}
    Wait Until Keyword Succeeds    40x    5s
    ...    Router Should Resolve Hostname    ${hostname}

Router Should Resolve Hostname
    [Documentation]    Check if the router pod resolves the given hostname
    [Arguments]    ${hostname}
    ${output}=    Oc Exec    router-default    nslookup ${hostname}    openshift-ingress    deployment
    Should Contain    ${output}    Name:
    Should Contain    ${output}    ${hostname}

Teardown Custom Corefile
    [Documentation]    Remove custom Corefile, drop-in config, fake IP, and restart
    Run Keywords
    ...    Remove Drop In MicroShift Config    20-dns-custom
    ...    AND    Remove Custom Corefile
    ...    AND    Remove Fake IP From NIC    ${FAKE_LISTEN_IP}
    ...    AND    Restart MicroShift

Remove Custom Corefile
    [Documentation]    Remove the custom Corefile from the host
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    rm -f ${CUSTOM_COREFILE_PATH}
    ...    sudo=True    return_rc=True    return_stdout=True    return_stderr=True
    Should Be Equal As Integers    0    ${rc}

Add Fake IP To NIC
    [Documentation]    Add the given IP to br-ex temporarily.
    [Arguments]    ${ip_address}=${FAKE_LISTEN_IP}    ${nic_name}=br-ex
    ${stdout}    ${stderr}    ${rc}=    SSHLibrary.Execute Command
    ...    ip address add ${ip_address}/32 dev ${nic_name}
    ...    sudo=True    return_rc=True    return_stderr=True    return_stdout=True
    Log Many    ${stdout}    ${stderr}
    Should Be Equal As Integers    0    ${rc}

Remove Fake IP From NIC
    [Documentation]    Remove the given IP from br-ex.
    [Arguments]    ${ip_address}=${FAKE_LISTEN_IP}    ${nic_name}=br-ex
    ${stdout}    ${stderr}    ${rc}=    SSHLibrary.Execute Command
    ...    ip address delete ${ip_address}/32 dev ${nic_name}
    ...    sudo=True    return_rc=True    return_stderr=True    return_stdout=True
    Log Many    ${stdout}    ${stderr}
    Should Be Equal As Integers    0    ${rc}
