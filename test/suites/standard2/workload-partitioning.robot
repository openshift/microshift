*** Settings ***
Documentation       Tests for Workload partitioning

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/microshift-network.resource

Suite Setup         Setup Suite With Namespace
Suite Teardown      Teardown Suite With Namespace


*** Test Cases ***
Control Plane Pods Must Be Annotated
    [Documentation]    Verify that all the Control Plane pods are properly annotated.
    All Pods Should Be Annotated As Management


*** Keywords ***
All Pods Should Be Annotated As Management
    [Documentation]    Obtains list of Deployments created by CSV.
    ${pods_raw}=    Oc Get All Pods
    @{pods}=    Split String    ${pods_raw}
    FOR    ${pod}    IN    @{pods}
        ${ns}    ${pod}=    Split String    ${pod}    \@
        Pod Must Be Annotated    ${ns}    ${pod}
    END

Pod Must Be Annotated
    [Documentation]    Check management annotation for specified pod and namespace.
    [Arguments]    ${ns}    ${pod}
    ${management_annotation}=    Oc Get JsonPath
    ...    pod
    ...    ${ns}
    ...    ${pod}
    ...    .metadata.annotations.target\\.workload\\.openshift\\.io/management
    Should Not Be Empty    ${management_annotation}

Oc Get All Pods
    [Documentation]    Returns the running pods across all namespaces,
    ...    Returns the command output as formatted string <name-space>@<pod-name>

    ${data}=    Oc Get JsonPath
    ...    pods
    ...    ${EMPTY}
    ...    ${EMPTY}
    ...    range .items[*]}{\.metadata\.namespace}{"@"}{\.metadata\.name}{"\\n"}{end
    RETURN    ${data}
