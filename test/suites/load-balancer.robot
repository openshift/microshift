*** Settings ***
Documentation       Load balancer test suite.

Library             Process
Resource            ../resources/common.resource
Resource            ../resources/kubeconfig.resource

Suite Setup         Setup Suite With Namespace
Suite Teardown      Teardown Suite With Namespace


*** Variables ***
${HELLO_USHIFT}     assets/hello-microshift.yaml
${LB_PORT}          ${EMPTY}


*** Test Cases ***
Load Balancer Smoke Test
    [Documentation]    Verify that Load Balancer correctly exposes HTTP service
    [Tags]    smoke
    [Setup]    Run Keywords
    ...    Create Hello MicroShift Pod
    ...    AND
    ...    Expose Hello MicroShift Pod Via LB

    Wait Until Keyword Succeeds    10x    6s
    ...    Access Hello Microshift Via LB

    [Teardown]    Run Keywords
    ...    Delete Hello MicroShift Pod Route And Service


*** Keywords ***
Create Hello MicroShift Pod
    [Documentation]    Create a pod running the "hello microshift" application
    Run With Kubeconfig    oc create -f ${HELLO_USHIFT} -n ${NAMESPACE}
    Run With Kubeconfig    oc wait pods -l app\=hello-microshift --for condition\=Ready --timeout\=60s -n ${NAMESPACE}

Expose Hello MicroShift Pod Via LB
    [Documentation]    Expose the "hello microshift" application through the load balancer
    Run With Kubeconfig    oc create service loadbalancer hello-microshift --tcp=5678:8080 -n ${NAMESPACE}

Access Hello Microshift Via LB
    [Documentation]    Try to retrieve data from the "hello microshift" service end point
    IF    "${LB_PORT}"=="${EMPTY}"
        ${connect_to}=    Set Variable    "hello-microshift.cluster.local:80:${USHIFT_HOST}:5678"
    ELSE
        ${connect_to}=    Set Variable    "hello-microshift.cluster.local:80:${USHIFT_HOST}:${LB_PORT}"
    END
    ${result}=    Run Process
    ...    curl -i http://hello-microshift.cluster.local --connect-to ${connect_to}
    ...    shell=True
    ...    timeout=15s
    Log Many    ${result.rc}    ${result.stdout}    ${result.stderr}
    Should Be Equal As Integers    ${result.rc}    0
    Should Match Regexp    ${result.stdout}    HTTP.*200
    Should Match    ${result.stdout}    *Hello MicroShift*

Delete Hello MicroShift Pod Route And Service
    [Documentation]    Remove the "hello microshift" resources
    Run With Kubeconfig    oc delete route hello-microshift -n ${NAMESPACE}    True
    Run With Kubeconfig    oc delete service hello-microshift -n ${NAMESPACE}    True
    Run With Kubeconfig    oc delete -f ${HELLO_USHIFT} -n ${NAMESPACE}    True
