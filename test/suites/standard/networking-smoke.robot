*** Settings ***
Documentation       Networking small tests

Resource            ../../resources/common.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/ostree-health.resource

Suite Setup         Setup Suite With Namespace
Suite Teardown      Teardown Suite With Namespace

Test Tags           smoke


*** Variables ***
${HELLO_USHIFT}     ./assets/hello-microshift.yaml
${LB_PORT}          5678
${HTTP_PORT}        80


*** Test Cases ***
Router Smoke Test
    [Documentation]    Run a router smoke test
    [Setup]    Run Keywords
    ...    Create Hello MicroShift Pod
    ...    AND
    ...    Expose Hello MicroShift Service Via Route

    Wait Until Keyword Succeeds    10x    6s
    ...    Access Hello Microshift    ${HTTP_PORT}

    [Teardown]    Run Keywords
    ...    Delete Hello MicroShift Route
    ...    Delete Hello MicroShift Pod And Service

Load Balancer Smoke Test
    [Documentation]    Verify that Load Balancer correctly exposes HTTP service
    [Setup]    Run Keywords
    ...    Create Hello MicroShift Pod
    ...    AND
    ...    Expose Hello MicroShift Pod Via LB

    Wait Until Keyword Succeeds    10x    6s
    ...    Access Hello Microshift    ${LB_PORT}

    [Teardown]    Run Keywords
    ...    Delete Hello MicroShift Pod And Service


*** Keywords ***
Create Hello MicroShift Pod
    [Documentation]    Create a pod running the "hello microshift" application
    Oc Create    -f ${HELLO_USHIFT} -n ${NAMESPACE}
    Oc Wait For    pods -l app=hello-microshift    condition\=Ready    timeout=60s

Expose Hello MicroShift Pod Via LB
    [Documentation]    Expose the "hello microshift" application through the load balancer
    Run With Kubeconfig    oc create service loadbalancer hello-microshift --tcp=${LB_PORT}:8080 -n ${NAMESPACE}

Expose Hello MicroShift Service Via Route
    [Documentation]    Expose the "hello microshift" application through the Route
    Oc Expose    pod hello-microshift -n ${NAMESPACE}
    Oc Expose    svc hello-microshift --hostname hello-microshift.cluster.local -n ${NAMESPACE}

Access Hello Microshift
    [Documentation]    Try to retrieve data from the "hello microshift" service end point
    [Arguments]    ${ushift_port}

    ${connect_to}=    Set Variable    "hello-microshift.cluster.local:${HTTP_PORT}:${USHIFT_HOST}:${ushift_port}"

    ${result}=    Run Process
    ...    curl -i http://hello-microshift.cluster.local --connect-to ${connect_to}
    ...    shell=True
    ...    timeout=15s
    Log Many    ${result.rc}    ${result.stdout}    ${result.stderr}
    Should Be Equal As Integers    ${result.rc}    0
    Should Match Regexp    ${result.stdout}    HTTP.*200
    Should Match    ${result.stdout}    *Hello MicroShift*

Delete Hello MicroShift Pod And Service
    [Documentation]    Delete service and pod for cleanup.
    Oc Delete    service/hello-microshift -n ${NAMESPACE}
    Oc Delete    -f ${HELLO_USHIFT} -n ${NAMESPACE}

Delete Hello MicroShift Route
    [Documentation]    Delete route for cleanup.
    Oc Delete    route/hello-microshift -n ${NAMESPACE}
