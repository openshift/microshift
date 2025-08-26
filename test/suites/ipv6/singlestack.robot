*** Settings ***
Documentation       Tests related to MicroShift running in an IPv6-only host

Resource            ../../resources/common.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/ostree-health.resource
Resource            ../../resources/microshift-network.resource
Library             ../../resources/libipv6.py

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           ipv6    network


*** Variables ***
${USHIFT_HOST}      ${EMPTY}
${HOSTNAME}         hello-microshift.ipv6.cluster.local


*** Test Cases ***
Verify Router Serves IPv6
    [Documentation]    Verify router is capable of serving ipv6 traffic.
    [Setup]    Run Keywords
    ...    Create Hello MicroShift Pod
    ...    Expose Hello MicroShift Service Via Route
    ...    Restart Router

    Must Be Ipv6    ${USHIFT_HOST}
    Wait Until Keyword Succeeds    10x    6s
    ...    Access Hello Microshift Success    ushift_port=${HTTP_PORT}    hostname=${HOSTNAME}

    DNS Entry For Route Should Resolve

    [Teardown]    Run Keywords
    ...    Delete Hello MicroShift Route
    ...    Delete Hello MicroShift Pod And Service
    ...    Wait For Service Deletion With Timeout

Verify All Services Are Ipv6
    [Documentation]    Check all services are running IPv6 only

    All Services Are Ipv6


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Login MicroShift Host
    Setup Suite With Namespace
    Wait Until Greenboot Health Check Exited

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host

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

Expose Hello MicroShift Service Via Route
    [Documentation]    Expose the "hello microshift" application through the Route
    Oc Expose    pod hello-microshift -n ${NAMESPACE}
    Oc Expose    svc hello-microshift --hostname ${HOSTNAME} -n ${NAMESPACE}

Network APIs With Test Label Are Gone
    [Documentation]    Check for service and endpoint by "app=hello-microshift" label. Succeeds if response matches
    ...    "No resources found in <namespace> namespace." Fail if not.
    ${match_string}=    Catenate    No resources found in    ${NAMESPACE}    namespace.
    ${match_string}=    Remove String    ${match_string}    "
    ${response}=    Run With Kubeconfig    oc get svc,endpointslices -l app\=hello-microshift -n ${NAMESPACE}
    Should Be Equal As Strings    ${match_string}    ${response}    strip_spaces=True

DNS Entry For Route Should Resolve
    [Documentation]    Resolve hello-microshift route via mDNS from the hypervisor/RF runner.
    ...    Expects RF runner host has opened port 5353 for libvirt zone.

    ${result}=    Run Process
    ...    avahi-resolve-host-name ${HOSTNAME}
    ...    shell=True
    ...    timeout=15s
    Should Be Equal As Integers    0    ${result.rc}
    Log Many    ${result.stdout}    ${result.stderr}
    @{words}=    Split String    ${result.stdout}
    Must Be Ipv6    ${words}[1]

Restart Router
    [Documentation]    Restart the router and wait for readiness again. The router is sensitive to apiserver
    ...    downtime and might need a restart (after the apiserver is ready) to resync all the routes.
    Run With Kubeconfig    oc rollout restart deployment router-default -n openshift-ingress
    Named Deployment Should Be Available    router-default    openshift-ingress    5m

All Services Are Ipv6
    [Documentation]    Retrieve all services and check none of them have an IPv4 family
    ${response}=    Run With Kubeconfig    oc get svc -A -o jsonpath='{.items[*].spec.ipFamilies[*]}'
    Should Not Contain    ${response}    IPv4
