*** Settings ***
Documentation       Networking smoke tests

Resource            ../../resources/common.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/ostree-health.resource
Resource            ../../resources/microshift-network.resource

Suite Setup         Setup Suite With Namespace
Suite Teardown      Teardown Suite With Namespace

Test Tags           slow


*** Variables ***
${HELLO_USHIFT_INGRESS}     ./assets/hello-microshift-ingress.yaml
${HOSTNAME}                 hello-microshift.cluster.local


*** Test Cases ***
Router Smoke Test
    [Documentation]    Run a router smoke test
    [Setup]    Run Keywords
    ...    Create Hello MicroShift Pod
    ...    Expose Hello MicroShift Service Via Route
    ...    Restart Router

    Wait Until Keyword Succeeds    10x    6s
    ...    Access Hello Microshift    ${HTTP_PORT}

    DNS Entry For Route Should Resolve

    [Teardown]    Run Keywords
    ...    Delete Hello MicroShift Route
    ...    Delete Hello MicroShift Pod And Service
    ...    Wait For Service Deletion With Timeout

Load Balancer Smoke Test
    [Documentation]    Verify that Load Balancer correctly exposes HTTP service
    Verify Hello MicroShift LB

Ingress Smoke Test
    [Documentation]    Verify a simple ingress rule correctly exposes HTTP service
    [Setup]    Run Keywords
    ...    Create Hello MicroShift Pod
    ...    Expose Hello MicroShift
    ...    Create Hello MicroShift Ingress
    ...    Restart Router

    Wait Until Keyword Succeeds    10x    6s
    ...    Access Hello Microshift    ${HTTP_PORT}    path=/principal

    [Teardown]    Run Keywords
    ...    Delete Hello MicroShift Ingress
    ...    Delete Hello MicroShift Pod And Service
    ...    Wait For Service Deletion With Timeout


*** Keywords ***
Expose Hello MicroShift Service Via Route
    [Documentation]    Expose the "hello microshift" application through the Route
    Oc Expose    pod hello-microshift -n ${NAMESPACE}
    Oc Expose    svc hello-microshift --hostname hello-microshift.cluster.local -n ${NAMESPACE}

Delete Hello MicroShift Route
    [Documentation]    Delete route for cleanup.
    Oc Delete    route/hello-microshift -n ${NAMESPACE}

Wait For Service Deletion With Timeout
    [Documentation]    Polls for service and endpoint by "app=hello-microshift" label. Fails if timeout
    ...    expires. This check is unique to this test suite because each test here reuses the same namespace. Since
    ...    the tests reuse the service name, a small race window exists between the teardown of one test and the setup
    ...    of the next. This produces flakey failures when the service or endpoint names collide.
    Wait Until Keyword Succeeds    30s    1s
    ...    Network APIs With Test Label Are Gone

Restart Router
    [Documentation]    Restart the router and wait for readiness again. The router is sensitive to apiserver
    ...    downtime and might need a restart (after the apiserver is ready) to resync all the routes.
    Run With Kubeconfig    oc rollout restart deployment router-default -n openshift-ingress
    Named Deployment Should Be Available    router-default    openshift-ingress    5m

Network APIs With Test Label Are Gone
    [Documentation]    Check for service and endpoint by "app=hello-microshift" label. Succeeds if response matches
    ...    "No resources found in <namespace> namespace." Fail if not.
    ${match_string}=    Catenate    No resources found in    ${NAMESPACE}    namespace.
    ${match_string}=    Remove String    ${match_string}    "
    ${response}=    Run With Kubeconfig    oc get svc,ep -l app\=hello-microshift -n ${NAMESPACE}
    Should Be Equal As Strings    ${match_string}    ${response}    strip_spaces=True

Create Hello MicroShift Ingress
    [Documentation]    Create ingress rule.
    Oc Create    -f ${HELLO_USHIFT_INGRESS} -n ${NAMESPACE}

Delete Hello MicroShift Ingress
    [Documentation]    Delete ingress for cleanup.
    Oc Delete    -f ${HELLO_USHIFT_INGRESS} -n ${NAMESPACE}

DNS Entry For Route Should Resolve
    [Documentation]    Resolve hello-microshift route via mDNS from the hypervisor/RF runner.
    ...    Expects RF runner host has opened port 5353 for libvirt zone.

    ${result}=    Run Process
    ...    avahi-resolve-host-name ${HOSTNAME}
    ...    shell=True
    ...    timeout=15s
    Log Many    ${result.stdout}    ${result.stderr}

    # avahi-resolve-host-name always returns rc=0, even if it failed to resolve.
    # In case of failure, stdout will be empty and stderr will contain error.
    # Expected success stdout:
    # > hello-microshift.cluster.local 192.168.124.5
    # Possible stderr:
    # > Failed to resolve host name 'hello-microshift.cluster.local': Timeout reached
    Should Not Be Empty    ${result.stdout}
    Should Be Empty    ${result.stderr}
