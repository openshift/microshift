*** Settings ***
Documentation       Networking smoke tests

Resource            ../../resources/common.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/ostree-health.resource
Resource            ../../resources/microshift-network.resource

Suite Setup         Setup Suite With Namespace
Suite Teardown      Teardown Suite With Namespace

Test Tags           smoke

*** Variables ***
${HELLO_USHIFT_INGRESS}     ./assets/hello-microshift-ingress.yaml

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
    Verify Hello MicroShift LB

Ingress Smoke Test
    [Documentation]    Verify a simple ingress rule correctly exposes HTTP service
    [Setup]    Run Keywords
    ...    Create Hello MicroShift Pod
    ...    Expose Hello MicroShift
    ...    Create Hello MicroShift Ingress

    Wait Until Keyword Succeeds    10x    6s
    ...    Access Hello Microshift    ${HTTP_PORT}    path="/principal"

    [Teardown]    Run Keywords
    ...    Delete Hello MicroShift Ingress
    ...    Delete Hello MicroShift Pod And Service

*** Keywords ***
Expose Hello MicroShift Service Via Route
    [Documentation]    Expose the "hello microshift" application through the Route
    Oc Expose    pod hello-microshift -n ${NAMESPACE}
    Oc Expose    svc hello-microshift --hostname hello-microshift.cluster.local -n ${NAMESPACE}

Delete Hello MicroShift Route
    [Documentation]    Delete route for cleanup.
    Oc Delete    route/hello-microshift -n ${NAMESPACE}

Create Hello MicroShift Ingress
    [Documentation]    Create ingress rule.
    Oc Create    -f ${HELLO_USHIFT_INGRESS} -n ${NAMESPACE}

Delete Hello MicroShift Ingress
    [Documentation]    Delete ingress for cleanup.
    Oc Delete    -f ${HELLO_USHIFT_INGRESS} -n ${NAMESPACE}
