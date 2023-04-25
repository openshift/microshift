*** Settings ***
Documentation   Load balancer test suite.

Library         Process

Resource        ../resources/common.resource
Resource        ../resources/kubeconfig.resource

Suite Setup     Setup Suite With Namespace
Suite Teardown  Teardown Suite With Namespace

*** Variables ***
${USHIFT_HOST}    ${EMPTY}
${USHIFT_USER}    ${EMPTY}

*** Test Cases ***
Load Balancer Smoke Test
    [Documentation]    Verify that Load Balancer correctly exposes HTTP service
    [Tags]    smoke
    [Setup]    Run Keywords
    ...    Create Hello MicroShift Pod    AND
    ...    Expose Hello MicroShift Pod Via LB

    Wait Until Keyword Succeeds    3x    3s    Access Hello Microshift via LB

    [Teardown]    Run Keywords
    ...    Delete Hello MicroShift Pod Route And Service

*** Keywords ***
Create Hello MicroShift Pod
    Run With Kubeconfig    oc create -f ../e2e/tests/assets/hello-microshift.yaml -n ${NAMESPACE}
    Run With Kubeconfig    oc wait pods -l app\=hello-microshift --for condition\=Ready --timeout\=60s -n ${NAMESPACE}

Expose Hello MicroShift Pod Via LB
    Run With Kubeconfig    oc create service loadbalancer hello-microshift --tcp=5678:8080 -n ${NAMESPACE}

Access Hello Microshift via LB
    ${result}=    Run Process
    ...    curl -i http://hello-microshift.cluster.local --connect-to "hello-microshift.cluster.local:80:${USHIFT_HOST}:5678"
    ...    shell=True    timeout=15s
    Log    ${result}
    Should Be Equal As Integers    ${result.rc}    0
    Should Match Regexp    ${result.stdout}    HTTP.*200
    Should Match    ${result.stdout}    *Hello MicroShift*

Delete Hello MicroShift Pod Route And Service
    Run With Kubeconfig    oc delete route hello-microshift -n ${NAMESPACE}    True
    Run With Kubeconfig    oc delete service hello-microshift -n ${NAMESPACE}   True
    Run With Kubeconfig    oc delete -f ../e2e/tests/assets/hello-microshift.yaml -n ${NAMESPACE}    True
