*** Settings ***
Documentation       Tests related to running MicroShift on a disconnected host.

Resource            ../../resources/offline.resource

Suite Setup         offline.Setup Suite

Test Tags           offline


*** Variables ***
${SERVER_POD}           assets/hello/hello-microshift.yaml
${SERVER_INGRESS}       assets/hello/hello-microshift-ingress.yaml
${SERVER_SVC}           assets/hello/hello-microshift-service.yaml
${TEST_NS}              ${EMPTY}


*** Test Cases ***
MicroShift Should Boot Into Healthy State When Host Is Offline
    [Documentation]    Test that MicroShift starts when the host is offline.
    offline.Reboot MicroShift Host
    Wait For Greenboot Health Check To Exit

Load Balancer Services Should Be Running Correctly When Host Is Offline
    [Documentation]    Test that the load balancer services are running correctly when the host is offline.
    [Setup]    Setup Test

    Wait Until Keyword Succeeds    2m    10s
    ...    Pod Should Be Reachable Via Ingress

    [Teardown]    Teardown Test


*** Keywords ***
Setup Test
    [Documentation]    Deploy cluster resources to provide an endpoint for testing. Wait for greenboot
    Wait For Greenboot Health Check To Exit
    ${ns}=    Create Test Namespace
    Send File    ${SERVER_SVC}    /tmp/hello-microshift-service.yaml
    offline.Run With Kubeconfig    oc    create    -n\=${ns}    -f\=/tmp/hello-microshift-service.yaml
    Send File    ${SERVER_INGRESS}    /tmp/hello-microshift-ingress.yaml
    offline.Run With Kubeconfig    oc    create    -n\=${ns}    -f\=/tmp/hello-microshift-ingress.yaml
    Send File    ${SERVER_POD}    /tmp/hello-microshift.yaml
    offline.Run With Kubeconfig    oc    create    -n\=${ns}    -f\=/tmp/hello-microshift.yaml
    # Set this last, it's not available within this scope anyway
    offline.Run With Kubeconfig    oc    wait    -n\=${ns}    --for\=condition=Ready    pod/hello-microshift
    Set Test Variable    ${TEST_NS}    ${ns}

Teardown Test
    [Documentation]    Test teardown
    offline.Run With Kubeconfig    oc    delete    -n\=${TEST_NS}    -f\=/tmp/hello-microshift-service.yaml
    offline.Run With Kubeconfig    oc    delete    -n\=${TEST_NS}    -f\=/tmp/hello-microshift-ingress.yaml
    offline.Run With Kubeconfig    oc    delete    -n\=${TEST_NS}    -f\=/tmp/hello-microshift.yaml
    offline.Run With Kubeconfig    oc    delete    namespace    ${TEST_NS}

Pod Should Be Reachable Via Ingress
    [Documentation]    Checks if a load balancer service can be accessed. The curl command must be executed in a
    ...    subshell.
    ${result}    ${ignore}=    Wait Until Keyword Succeeds    5x    1s
    ...    Run Guest Process    ${GUEST_NAME}
    ...        bash
    ...        -c
    ...        curl --fail -I --max-time 15 -H \"Host: hello-microshift.cluster.local\" ${NODE_IP}:80/principal
    Log Many    ${result["stdout"]}    ${result["stderr"]}
    Should Be Equal As Integers    ${result["rc"]}    0

Create Test Namespace
    [Documentation]    Create a test namespace with a random name

    ${rand}=    Generate Random String
    ${rand}=    Convert To Lower Case    ${rand}
    ${ns}=    Set Variable    test-${rand}
    Wait Until Keyword Succeeds    5m    10s
    ...    offline.Run With Kubeconfig    oc    create    namespace    ${ns}
    RETURN    ${ns}
