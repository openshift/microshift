*** Settings ***
Documentation       Tests for ClusterIP service creation and connectivity

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-network.resource
Resource            ../../resources/oc.resource

Suite Setup         Setup Suite With Namespace
Suite Teardown      Teardown Suite With Namespace


*** Test Cases ***
Create ClusterIP Service With Explicit IP
    [Documentation]    Create a ClusterIP service with an explicit cluster IP
    ...    and verify connectivity from within a pod.
    [Setup]    Create Hello MicroShift Pod
    ${k8s_ip}=    Oc Get JsonPath    service    default    kubernetes    .spec.clusterIP
    ${explicit_ip}=    Evaluate    '.'.join("${k8s_ip}".strip().split('.')[:-1]) + '.200'
    Oc Create    service clusterip hello-microshift --tcp=8080:8080 --clusterip=${explicit_ip} -n ${NAMESPACE}
    ${svc_ip}=    Oc Get JsonPath    service    ${NAMESPACE}    hello-microshift    .spec.clusterIP
    Should Be Equal    ${svc_ip}    ${explicit_ip}
    Wait Until Keyword Succeeds    10x    5s
    ...    Verify ClusterIP Connectivity    ${svc_ip}

    [Teardown]    Cleanup Hello MicroShift Pod And Service

Create ClusterIP Service Without Explicit IP
    [Documentation]    Create a ClusterIP service with an auto-assigned IP
    ...    and verify connectivity from within a pod.
    [Setup]    Create Hello MicroShift Pod
    Oc Create    service clusterip hello-microshift --tcp=8080:8080 -n ${NAMESPACE}
    ${svc_ip}=    Oc Get JsonPath    service    ${NAMESPACE}    hello-microshift    .spec.clusterIP
    Should Not Be Empty    ${svc_ip}
    Wait Until Keyword Succeeds    10x    5s
    ...    Verify ClusterIP Connectivity    ${svc_ip}

    [Teardown]    Cleanup Hello MicroShift Pod And Service

Create ClusterIP Service Without TCP Option Should Fail
    [Documentation]    Creating a ClusterIP service without the --tcp option should fail.
    [Setup]    Create Hello MicroShift Pod
    ${stdout}    ${rc}=    Run With Kubeconfig
    ...    oc create service clusterip hello-microshift -n ${NAMESPACE}
    ...    allow_fail=${TRUE}    return_rc=${TRUE}
    Should Not Be Equal As Integers    ${rc}    0

    [Teardown]    Oc Delete    -f ${HELLO_USHIFT} -n ${NAMESPACE} --ignore-not-found

Create ClusterIP Service Dry Run
    [Documentation]    Creating a service with --dry-run=client should not persist it.
    [Setup]    Create Hello MicroShift Pod
    Oc Create    service clusterip hello-microshift --tcp=8080:8080 --dry-run=client -n ${NAMESPACE}
    ${stdout}    ${rc}=    Run With Kubeconfig
    ...    oc get service hello-microshift -n ${NAMESPACE}
    ...    allow_fail=${TRUE}    return_rc=${TRUE}
    Should Not Be Equal As Integers    ${rc}    0

    [Teardown]    Oc Delete    -f ${HELLO_USHIFT} -n ${NAMESPACE} --ignore-not-found


*** Keywords ***
Cleanup Hello MicroShift Pod And Service
    [Documentation]    Delete the hello-microshift service and pod, ignoring errors
    Run With Kubeconfig    oc delete service/hello-microshift -n ${NAMESPACE} --ignore-not-found
    Oc Delete    -f ${HELLO_USHIFT} -n ${NAMESPACE} --ignore-not-found

Verify ClusterIP Connectivity
    [Documentation]    Verify connectivity to the ClusterIP service from within the pod
    [Arguments]    ${ip}
    ${output}=    Run With Kubeconfig
    ...    oc exec -n ${NAMESPACE} pod/hello-microshift -- wget -qO- http://${ip}:8080 --timeout=5
    Should Contain    ${output}    Hello MicroShift
