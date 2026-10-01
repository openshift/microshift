*** Settings ***
Documentation       Shared keywords for DNS test suites

Library             SSHLibrary
Resource            oc.resource


*** Keywords ***
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

Add Fake IP To NIC
    [Documentation]    Add the given IP to br-ex temporarily.
    [Arguments]    ${ip_address}    ${nic_name}=br-ex
    ${stdout}    ${stderr}    ${rc}=    SSHLibrary.Execute Command
    ...    ip address add ${ip_address}/32 dev ${nic_name}
    ...    sudo=True    return_rc=True    return_stderr=True    return_stdout=True
    Log Many    ${stdout}    ${stderr}
    Should Be Equal As Integers    0    ${rc}

Remove Fake IP From NIC
    [Documentation]    Remove the given IP from br-ex.
    [Arguments]    ${ip_address}    ${nic_name}=br-ex
    ${stdout}    ${stderr}    ${rc}=    SSHLibrary.Execute Command
    ...    ip address delete ${ip_address}/32 dev ${nic_name}
    ...    sudo=True    return_rc=True    return_stderr=True    return_stdout=True
    Log Many    ${stdout}    ${stderr}
    Should Be Equal As Integers    0    ${rc}
