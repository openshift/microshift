*** Settings ***
Documentation       NetworkPolicy tests: ingress/egress rules, hairpin traffic,
...                 podSelector-based allow/deny policies.

Resource            ../../resources/common.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/microshift-network.resource
Resource            ../../resources/network-testing.resource

Suite Setup         Setup Suite With Namespace
Suite Teardown      Teardown Suite With Namespace

Test Tags           network    slow


*** Variables ***
${NETPOL_ASSETS}    ./assets/network-policy
${CURL_TIMEOUT}     5


*** Test Cases ***
Mixed Ingress And Egress NetworkPolicy
    [Documentation]    Verify that mixed ingress and egress NetworkPolicies block
    ...    cross-namespace traffic correctly. Egress policy restricts pods to
    ...    only reach namespaces labeled team=openshift, while ingress policy
    ...    restricts to traffic from pods labeled name=test-pods.
    ...    Covers QE test 60331.
    [Setup]    Setup Mixed Policy Test

    ${hello_pod_ns1}=    Set Variable    hello-microshift
    ${hello_pod_ip_ns2}=    Get Pod IP    hello-microshift    ${NS_MIXED_2}

    Wait Until Keyword Succeeds    5x    5s
    ...    Curl From Pod Should Timeout
    ...    ${hello_pod_ns1}    ${NS_MIXED_1}
    ...    http://${hello_pod_ip_ns2}:8080    timeout=${CURL_TIMEOUT}

    [Teardown]    Teardown Mixed Policy Test

Hairpin Traffic Through Service With NetworkPolicy
    [Documentation]    Verify that NetworkPolicy allowing same-namespace traffic works
    ...    correctly when traffic hairpins through a service back to the same
    ...    source pod. Tests both direct pod-to-pod and pod-to-service connectivity.
    ...    Covers QE test 60332.
    [Setup]    Setup Hairpin Test

    ${pod1_ip}=    Get Pod IP    hello-pod1    ${NAMESPACE}
    ${pod2_ip}=    Get Pod IP    hello-pod2    ${NAMESPACE}
    ${svc_ip}=    Get Service ClusterIP    test-service    ${NAMESPACE}

    Wait Until Keyword Succeeds    5x    2s
    ...    Curl From Pod Should Succeed
    ...    hello-pod1    ${NAMESPACE}    http://${pod2_ip}:8080
    ...    expected=Hello MicroShift

    Wait Until Keyword Succeeds    5x    2s
    ...    Curl From Pod Should Succeed
    ...    hello-pod2    ${NAMESPACE}    http://${pod1_ip}:8080
    ...    expected=Hello MicroShift

    FOR    ${i}    IN RANGE    5
        Curl From Pod Should Succeed
        ...    hello-pod1    ${NAMESPACE}    http://${svc_ip}:27017
        ...    expected=Hello MicroShift
        Curl From Pod Should Succeed
        ...    hello-pod2    ${NAMESPACE}    http://${svc_ip}:27017
        ...    expected=Hello MicroShift
    END

    [Teardown]    Teardown Hairpin Test

PodSelector Allow To And Allow From
    [Documentation]    Verify that podSelector-based NetworkPolicies (allow-from-red,
    ...    allow-to-blue, default-deny-ingress) work together. Tests 6 connectivity
    ...    scenarios across 2 namespaces with labeled pods:
    ...    - pod-red(type=red) -> pod-plain: ALLOWED (allow-from-red)
    ...    - pod-plain -> pod-blue(type=blue): ALLOWED (allow-to-blue)
    ...    - pod-plain -> pod-plain: DENIED (default-deny, no matching allow)
    ...    - ns2/pod-plain -> ns1/pod-plain: DENIED (cross-namespace, no allow)
    ...    - ns2/pod-plain -> ns1/pod-blue: ALLOWED (allow-to-blue allows all ingress)
    ...    - ns2/pod-plain -> ns1/pod-plain: DENIED
    ...    Covers QE test 60426.
    [Setup]    Setup PodSelector Test

    ${pod_red_ip_ns1}=    Get Pod IP    pod-red    ${NS_SEL_1}
    ${pod_blue_ip_ns1}=    Get Pod IP    pod-blue    ${NS_SEL_1}
    ${pod_plain_ip_ns1}=    Get Pod IP    pod-plain    ${NS_SEL_1}

    # pod-red(type=red) in ns1 -> pod-plain in ns1: ALLOWED (allow-from-red matches source label)
    Wait Until Keyword Succeeds    5x    2s
    ...    Curl From Pod Should Succeed
    ...    pod-red    ${NS_SEL_1}    http://${pod_plain_ip_ns1}:8080
    ...    expected=Hello MicroShift

    # pod-plain in ns1 -> pod-blue(type=blue) in ns1: ALLOWED (allow-to-blue allows all ingress)
    Wait Until Keyword Succeeds    5x    2s
    ...    Curl From Pod Should Succeed
    ...    pod-plain    ${NS_SEL_1}    http://${pod_blue_ip_ns1}:8080
    ...    expected=Hello MicroShift

    # pod-plain in ns1 -> pod-plain in ns1: DENIED (no matching allow rule)
    Wait Until Keyword Succeeds    5x    5s
    ...    Curl From Pod Should Timeout
    ...    pod-plain-src    ${NS_SEL_1}    http://${pod_plain_ip_ns1}:8080
    ...    timeout=${CURL_TIMEOUT}

    # ns2/pod-plain -> ns1/pod-plain: DENIED (cross-namespace, no allow rule)
    Wait Until Keyword Succeeds    5x    5s
    ...    Curl From Pod Should Timeout
    ...    pod-plain    ${NS_SEL_2}    http://${pod_plain_ip_ns1}:8080
    ...    timeout=${CURL_TIMEOUT}

    # ns2/pod-plain -> ns1/pod-blue: ALLOWED (allow-to-blue allows all ingress to blue)
    Wait Until Keyword Succeeds    5x    2s
    ...    Curl From Pod Should Succeed
    ...    pod-plain    ${NS_SEL_2}    http://${pod_blue_ip_ns1}:8080
    ...    expected=Hello MicroShift

    # ns2/pod-plain -> ns1/pod-red: DENIED (default-deny, pod-plain has no matching label)
    Wait Until Keyword Succeeds    5x    5s
    ...    Curl From Pod Should Timeout
    ...    pod-plain    ${NS_SEL_2}    http://${pod_red_ip_ns1}:8080
    ...    timeout=${CURL_TIMEOUT}

    [Teardown]    Teardown PodSelector Test


*** Keywords ***
Setup Mixed Policy Test
    [Documentation]    Create 2 namespaces, deploy pods with labels, apply egress and ingress policies.
    ${ns1}=    Create Random Namespace
    ${ns2}=    Create Random Namespace
    Set Suite Variable    ${NS_MIXED_1}    ${ns1}
    Set Suite Variable    ${NS_MIXED_2}    ${ns2}

    Create Hello MicroShift Pod    ns=${ns1}
    Create Hello MicroShift Pod    ns=${ns2}

    Oc Apply    -f ${NETPOL_ASSETS}/netpol-egress-ns-label.yaml -n ${ns1}
    Oc Apply    -f ${NETPOL_ASSETS}/netpol-ingress-pod-ns-label.yaml -n ${ns1}

Teardown Mixed Policy Test
    [Documentation]    Delete the extra namespaces.
    Run With Kubeconfig    oc delete namespace ${NS_MIXED_1}    allow_fail=True
    Run With Kubeconfig    oc delete namespace ${NS_MIXED_2}    allow_fail=True

Setup Hairpin Test
    [Documentation]    Create 2 pods with shared label, a ClusterIP service, and the
    ...    allow-from-same-namespace NetworkPolicy.
    Create Labeled Pod    hello-pod1    ${NAMESPACE}    labels=name=hello-pod
    Create Labeled Pod    hello-pod2    ${NAMESPACE}    labels=name=hello-pod
    Run With Kubeconfig
    ...    oc create service clusterip test-service --tcp=27017:8080 -n ${NAMESPACE}
    Run With Kubeconfig
    ...    oc set selector service test-service name=hello-pod -n ${NAMESPACE}
    Oc Apply    -f ${NETPOL_ASSETS}/netpol-allow-same-namespace.yaml -n ${NAMESPACE}

Teardown Hairpin Test
    [Documentation]    Remove hairpin test resources.
    Run With Kubeconfig    oc delete networkpolicy allow-from-same-namespace -n ${NAMESPACE}    allow_fail=True
    Run With Kubeconfig    oc delete service test-service -n ${NAMESPACE}    allow_fail=True
    Run With Kubeconfig    oc delete pod hello-pod1 hello-pod2 -n ${NAMESPACE} --grace-period=0    allow_fail=True

Setup PodSelector Test
    [Documentation]    Create 2 namespaces with labeled pods and apply NetworkPolicies.
    ${ns1}=    Create Random Namespace
    ${ns2}=    Create Random Namespace
    Set Suite Variable    ${NS_SEL_1}    ${ns1}
    Set Suite Variable    ${NS_SEL_2}    ${ns2}

    # ns1: pod-red (type=red), pod-blue (type=blue), pod-plain (no type label), pod-plain-src (curl source)
    Create Labeled Pod    pod-red    ${ns1}    labels=name=pod-red,type=red
    Create Labeled Pod    pod-blue    ${ns1}    labels=name=pod-blue,type=blue
    Create Labeled Pod    pod-plain    ${ns1}    labels=name=pod-plain
    Create Labeled Pod    pod-plain-src    ${ns1}    labels=name=pod-plain-src

    # ns2: pod-red (type=red), pod-plain (no type label)
    Create Labeled Pod    pod-red    ${ns2}    labels=name=pod-red,type=red
    Create Labeled Pod    pod-plain    ${ns2}    labels=name=pod-plain

    # Apply policies in ns1
    Oc Apply    -f ${NETPOL_ASSETS}/netpol-default-deny-ingress.yaml -n ${ns1}
    Oc Apply    -f ${NETPOL_ASSETS}/netpol-allow-from-red.yaml -n ${ns1}
    Oc Apply    -f ${NETPOL_ASSETS}/netpol-allow-to-blue.yaml -n ${ns1}

Teardown PodSelector Test
    [Documentation]    Delete the extra namespaces.
    Run With Kubeconfig    oc delete namespace ${NS_SEL_1}    allow_fail=True
    Run With Kubeconfig    oc delete namespace ${NS_SEL_2}    allow_fail=True
