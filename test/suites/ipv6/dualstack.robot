*** Settings ***
Documentation       Tests related to MicroShift running in a dual stack ipv4+6 host

Resource            ../../resources/common.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/ostree-health.resource
Resource            ../../resources/microshift-network.resource
Resource            ../../resources/microshift-config.resource
Library             ../../resources/libipv6.py

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           ipv6    network


*** Variables ***
${USHIFT_HOST_IP1}      ${EMPTY}
${USHIFT_HOST_IP2}      ${EMPTY}
${HOSTNAME}             hello-microshift.dualstack.cluster.local


*** Test Cases ***
Verify New Pod Works With IPv6
    [Documentation]    Verify IPv6 services are routable.
    [Setup]    Run Keywords
    ...    Migrate To Dual Stack
    ...    Create Hello MicroShift Pod
    ...    Expose Hello MicroShift Service Via Route IPv6
    ...    Restart Router

    ${ip_families}=    Oc Get JsonPath    svc    ${NAMESPACE}    hello-microshift    .spec.ipFamilies
    Should Contain    ${ip_families}    IPv6
    ${addr_type}=    Oc Get JsonPath    endpointslice    ${NAMESPACE}    -l kubernetes.io/service-name=hello-microshift    .items[0].addressType
    Should Be Equal    ${addr_type}    IPv6

    Wait Until Keyword Succeeds    20x    10s
    ...    Access Hello Microshift Success    ushift_ip=${USHIFT_HOST_IP1}
    ...    ushift_port=${HTTP_PORT}
    ...    hostname=${HOSTNAME}

    [Teardown]    Run Keywords
    ...    Delete Hello MicroShift Route
    ...    Delete Hello MicroShift Pod And Service
    ...    Wait For Service Deletion With Timeout
    ...    Remove Dual Stack Config Drop In
    ...    Restart MicroShift

Verify New Pod Works With IPv4
    [Documentation]    Verify IPv4 services are routable.
    [Setup]    Run Keywords
    ...    Migrate To Dual Stack
    ...    Create Hello MicroShift Pod
    ...    Expose Hello MicroShift Service Via Route IPv4
    ...    Restart Router

    ${ip_families}=    Oc Get JsonPath    svc    ${NAMESPACE}    hello-microshift    .spec.ipFamilies
    Should Contain    ${ip_families}    IPv4
    ${addr_type}=    Oc Get JsonPath    endpointslice    ${NAMESPACE}    -l kubernetes.io/service-name=hello-microshift    .items[0].addressType
    Should Be Equal    ${addr_type}    IPv4

    Wait Until Keyword Succeeds    20x    10s
    ...    Access Hello Microshift Success    ushift_ip=${USHIFT_HOST_IP1}
    ...    ushift_port=${HTTP_PORT}
    ...    hostname=${HOSTNAME}

    [Teardown]    Run Keywords
    ...    Delete Hello MicroShift Route
    ...    Delete Hello MicroShift Pod And Service
    ...    Wait For Service Deletion With Timeout
    ...    Remove Dual Stack Config Drop In
    ...    Restart MicroShift

Verify Host Network Pods Get Dual Stack IP Addresses
    [Documentation]    Verify host network pods get dual stack IP addresses
    [Setup]    Run Keywords
    ...    Migrate To Dual Stack

    # Wait a bit, as this is updated by kubelet and may not be immediate.
    Wait Until Keyword Succeeds    20x    6s
    ...    Host Network Pods Should Have Dual Stack IPs

    [Teardown]    Run Keywords
    ...    Remove Dual Stack Config Drop In
    ...    Restart MicroShift


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Initialize Global Variables
    Login MicroShift Host
    Setup Suite With Namespace
    Wait Until Greenboot Health Check Exited

Teardown
    [Documentation]    Test suite teardown
    Teardown Suite With Namespace
    Logout MicroShift Host

Remove Dual Stack Config Drop In
    [Documentation]    Remove dual stack config drop-in
    Remove Drop In MicroShift Config    10-dualstack

Initialize Global Variables
    [Documentation]    Initializes global variables.
    Log    IP1: ${USHIFT_HOST_IP1} IP2: ${USHIFT_HOST_IP2}
    Should Not Be Empty    ${USHIFT_HOST_IP1}    USHIFT_HOST_IP1 variable is required
    Should Not Be Empty    ${USHIFT_HOST_IP2}    USHIFT_HOST_IP2 variable is required

Migrate To Dual Stack
    [Documentation]    Configure MicroShift to enable dual stack network

    ${dual_stack}=    CATENATE    SEPARATOR=\n
    ...    ---
    ...    network:
    ...    \ \ clusterNetwork: [10.42.0.0/16, fd01::/48]
    ...    \ \ serviceNetwork: [10.43.0.0/16, fd02::/112]
    Drop In MicroShift Config    ${dual_stack}    10-dualstack
    Restart MicroShift

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

Expose Hello MicroShift Service Via Route IPv4
    [Documentation]    Expose the "hello microshift" application through the Route
    Run With Kubeconfig    oc apply -n ${NAMESPACE} -f assets/hello/hello-microshift-service.yaml
    Oc Expose    svc hello-microshift --hostname ${HOSTNAME} -n ${NAMESPACE}

Expose Hello MicroShift Service Via Route IPv6
    [Documentation]    Expose the "hello microshift" application through the Route
    Run With Kubeconfig    oc apply -n ${NAMESPACE} -f assets/hello/hello-microshift-service-ipv6.yaml
    Oc Expose    svc hello-microshift --hostname ${HOSTNAME} -n ${NAMESPACE}

Network APIs With Test Label Are Gone
    [Documentation]    Check for service and endpoint by "app=hello-microshift" label. Succeeds if response matches
    ...    "No resources found in <namespace> namespace." Fail if not.
    ${match_string}=    Catenate    No resources found in    ${NAMESPACE}    namespace.
    ${match_string}=    Remove String    ${match_string}    "
    ${response}=    Run With Kubeconfig    oc get svc,endpointslices -l app\=hello-microshift -n ${NAMESPACE}
    Should Be Equal As Strings    ${match_string}    ${response}    strip_spaces=True

Restart Router
    [Documentation]    Restart the router and wait for readiness again. The router is sensitive to apiserver
    ...    downtime and might need a restart (after the apiserver is ready) to resync all the routes.
    Run With Kubeconfig    oc rollout restart deployment router-default -n openshift-ingress
    Named Deployment Should Be Available    router-default    openshift-ingress    5m

Host Network Pods Should Have Dual Stack IPs
    [Documentation]    Check a system hostNetwork enabled pod for dual stack IP addresses
    ${pod_ips}=    Oc Get JsonPath
    ...    pod
    ...    openshift-dns
    ...    -l dns.operator.openshift.io/daemonset-node-resolver
    ...    .items[*].status.podIPs[*].ip
    Should Contain    ${pod_ips}    ${USHIFT_HOST_IP1}
    Should Contain    ${pod_ips}    ${USHIFT_HOST_IP2}
