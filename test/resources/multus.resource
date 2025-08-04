*** Settings ***
Documentation       Multus-related keywords used in different test suites.

Resource            common.resource
Library             DataFormats.py


*** Keywords ***
Create NAD And Pod
    [Documentation]    Creates provided NetworkAttachmentDefinition and Pod.
    [Arguments]    ${nad}    ${pod}
    Oc Create    -n ${NAMESPACE} -f ${nad}
    Oc Create    -n ${NAMESPACE} -f ${pod}

Remove NAD And Pod
    [Documentation]    Removes provided NetworkAttachmentDefinition and Pod.
    [Arguments]    ${nad}    ${pod}
    Run Keyword And Continue On Failure
    ...    Oc Delete    -n ${NAMESPACE} -f ${pod}
    Run Keyword And Continue On Failure
    ...    Oc Delete    -n ${NAMESPACE} -f ${nad}

Connect To Pod Over Local Interface
    [Documentation]    Makes a HTTP request to 8080 for a given Pod over given interface.
    [Arguments]    ${pod}    ${ns}    ${if}

    ${networks}=    Get And Verify Pod Networks    ${pod}    ${ns}    ${NAMESPACE}/bridge*-conf
    VAR    ${extra_ip}    ${networks}[1][ips][0]

    ${stdout}=    Command Should Work    curl -v --interface ${if} ${extra_ip}:8080
    Should Contain    ${stdout}    Hello MicroShift

Set IP For Host Interface
    [Documentation]    Sets IP address for the interface.
    [Arguments]    ${if}    ${cidr}
    Command Should Work    ip addr add ${cidr} dev ${if}

Get And Verify Pod Networks
    [Documentation]    Obtains interfaces of the Pod from its annotation.
    ...    The annotation is managed by Multus.
    [Arguments]    ${pod}    ${ns}    ${extra_cni_name}

    ${networks_str}=    Oc Get JsonPath
    ...    pod
    ...    ${ns}
    ...    ${pod}
    ...    .metadata.annotations.k8s\\.v1\\.cni\\.cncf\\.io/network-status
    Should Not Be Empty    ${networks_str}

    ${networks}=    Json Parse    ${networks_str}
    ${n}=    Get Length    ${networks}
    Should Be Equal As Integers    ${n}    2
    Should Be Equal As Strings    ${networks}[0][name]    ovn-kubernetes
    Should Match    ${networks}[1][name]    ${extra_cni_name}

    RETURN    ${networks}
