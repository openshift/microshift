*** Settings ***
Documentation       NetworkPolicy tests: ingress/egress rules, hairpin traffic,
...                 podSelector-based allow/deny policies.

Resource            ../../resources/common.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/microshift-network.resource
Resource            ../../resources/network-testing.resource

Suite Setup         Setup
Suite Teardown      Teardown Suite With Namespace

Test Tags           network    slow


*** Variables ***
${NETPOL_ASSETS}    ./assets/network-policy
${CURL_TIMEOUT}     5


*** Test Cases ***
Mixed Ingress And Egress NetworkPolicy
    [Documentation]    Verify the egress policy (ns1 hello pod may only reach namespaces labeled
    ...    team=openshift) and the ingress policy (ns1 hello pod accepts traffic only from
    ...    name=test-pods pods in team=operations namespaces), each with an allow and a deny path.
    ...    Covers QE test 60331.
    [Setup]    Setup Mixed Policy Test

    ${hello_ip_ns1}=    Get Pod IP    hello-microshift    ${NS_MIXED_1}
    ${hello_ip_ns2}=    Get Pod IP    hello-microshift    ${NS_MIXED_2}
    ${hello_ip_ns3}=    Get Pod IP    hello-microshift    ${NS_MIXED_3}

    # Egress ALLOWED: ns1 hello -> ns3 (team=openshift)
    Wait Until Keyword Succeeds    5x    2s
    ...    Curl From Pod Should Succeed    hello-microshift    ${NS_MIXED_1}
    ...    http://${hello_ip_ns3}:8080    expected=Hello MicroShift

    # Egress DENIED: ns1 hello -> ns2 (not team=openshift)
    Wait Until Keyword Succeeds    5x    5s
    ...    Curl From Pod Should Be Blocked    hello-microshift    ${NS_MIXED_1}
    ...    http://${hello_ip_ns2}:8080    timeout=${CURL_TIMEOUT}

    # Ingress ALLOWED: ns4 test-pods -> ns1 (team=operations + name=test-pods)
    Wait Until Keyword Succeeds    5x    2s
    ...    Curl From Pod Should Succeed    test-pods    ${NS_MIXED_4}
    ...    http://${hello_ip_ns1}:8080    expected=Hello MicroShift

    # Ingress DENIED: ns2 hello -> ns1 (not team=operations / not name=test-pods)
    Wait Until Keyword Succeeds    5x    5s
    ...    Curl From Pod Should Be Blocked    hello-microshift    ${NS_MIXED_2}
    ...    http://${hello_ip_ns1}:8080    timeout=${CURL_TIMEOUT}

    [Teardown]    Teardown Mixed Policy Test

Hairpin Traffic Through Service With NetworkPolicy
    [Documentation]    Verify NetworkPolicy allowing same-namespace traffic permits hairpin
    ...    traffic: hello-pod1 reaching a service whose only endpoint is hello-pod1 loops
    ...    back to itself. Also checks direct pod-to-pod and pod2-to-service connectivity.
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

PodSelector Allow To And Allow From    # robocop: off=too-long-test-case
    [Documentation]    Verify that podSelector-based NetworkPolicies (allow-from-red,
    ...    allow-to-blue, default-deny-ingress) work together. Tests 7 connectivity
    ...    scenarios across 2 namespaces with labeled pods:
    ...    - pod-red(type=red) -> pod-plain: ALLOWED (allow-from-red)
    ...    - pod-plain -> pod-blue(type=blue): ALLOWED (allow-to-blue)
    ...    - pod-plain-src -> pod-blue: ALLOWED (allow-to-blue; positive control)
    ...    - pod-plain-src -> pod-plain: DENIED (default-deny, no matching allow)
    ...    - ns2/pod-plain -> ns1/pod-plain: DENIED (cross-namespace, no allow)
    ...    - ns2/pod-plain -> ns1/pod-blue: ALLOWED (allow-to-blue allows all ingress)
    ...    - ns2/pod-plain -> ns1/pod-red: DENIED
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

    # pod-plain-src in ns1 -> pod-blue in ns1: ALLOWED (allow-to-blue; positive control for pod-plain-src)
    Wait Until Keyword Succeeds    5x    2s
    ...    Curl From Pod Should Succeed
    ...    pod-plain-src    ${NS_SEL_1}    http://${pod_blue_ip_ns1}:8080
    ...    expected=Hello MicroShift

    # pod-plain-src in ns1 -> pod-plain in ns1: DENIED (no matching allow rule)
    Wait Until Keyword Succeeds    5x    5s
    ...    Curl From Pod Should Be Blocked
    ...    pod-plain-src    ${NS_SEL_1}    http://${pod_plain_ip_ns1}:8080
    ...    timeout=${CURL_TIMEOUT}

    # ns2/pod-plain -> ns1/pod-plain: DENIED (cross-namespace, no allow rule)
    Wait Until Keyword Succeeds    5x    5s
    ...    Curl From Pod Should Be Blocked
    ...    pod-plain    ${NS_SEL_2}    http://${pod_plain_ip_ns1}:8080
    ...    timeout=${CURL_TIMEOUT}

    # ns2/pod-plain -> ns1/pod-blue: ALLOWED (allow-to-blue allows all ingress to blue)
    Wait Until Keyword Succeeds    5x    2s
    ...    Curl From Pod Should Succeed
    ...    pod-plain    ${NS_SEL_2}    http://${pod_blue_ip_ns1}:8080
    ...    expected=Hello MicroShift

    # ns2/pod-plain -> ns1/pod-red: DENIED (default-deny, pod-plain has no matching label)
    Wait Until Keyword Succeeds    5x    5s
    ...    Curl From Pod Should Be Blocked
    ...    pod-plain    ${NS_SEL_2}    http://${pod_red_ip_ns1}:8080
    ...    timeout=${CURL_TIMEOUT}

    [Teardown]    Teardown PodSelector Test


*** Keywords ***
Setup
    [Documentation]    Suite setup with a namespace that allows directly-created test pods.
    Setup Suite With Namespace
    Set Namespace Privileged

Setup Mixed Policy Test    # robocop: off=too-many-calls-in-keyword
    [Documentation]    Create namespaces and pods to exercise both the egress and ingress
    ...    NetworkPolicies applied to the hello-microshift pod in ns1:
    ...    ns2 (unlabeled), ns3 (team=openshift egress target), ns4 (team=operations ingress source).
    ${ns1}=    Create Random Namespace
    ${ns2}=    Create Random Namespace
    ${ns3}=    Create Random Namespace
    ${ns4}=    Create Random Namespace
    VAR    ${NS_MIXED_1}=    ${ns1}    scope=SUITE
    VAR    ${NS_MIXED_2}=    ${ns2}    scope=SUITE
    VAR    ${NS_MIXED_3}=    ${ns3}    scope=SUITE
    VAR    ${NS_MIXED_4}=    ${ns4}    scope=SUITE
    Set Namespace Privileged    ${ns1}
    Set Namespace Privileged    ${ns2}
    Set Namespace Privileged    ${ns3}
    Set Namespace Privileged    ${ns4}
    Label Namespace    ${ns3}    team=openshift
    Label Namespace    ${ns4}    team=operations

    Create Hello MicroShift Pod    ns=${ns1}
    Create Hello MicroShift Pod    ns=${ns2}
    Create Hello MicroShift Pod    ns=${ns3}
    Create Labeled Pod    test-pods    ${ns4}    labels=name=test-pods

    Oc Apply    -f ${NETPOL_ASSETS}/netpol-egress-ns-label.yaml -n ${ns1}
    Oc Apply    -f ${NETPOL_ASSETS}/netpol-ingress-pod-ns-label.yaml -n ${ns1}

Teardown Mixed Policy Test
    [Documentation]    Delete the extra namespaces.
    Run With Kubeconfig    oc delete namespace ${NS_MIXED_1}    allow_fail=True
    Run With Kubeconfig    oc delete namespace ${NS_MIXED_2}    allow_fail=True
    Run With Kubeconfig    oc delete namespace ${NS_MIXED_3}    allow_fail=True
    Run With Kubeconfig    oc delete namespace ${NS_MIXED_4}    allow_fail=True

Setup Hairpin Test
    [Documentation]    Create 2 pods and a ClusterIP service whose only endpoint is hello-pod1,
    ...    plus the allow-from-same-namespace NetworkPolicy. A single endpoint forces the
    ...    hello-pod1 -> service request to hairpin back to hello-pod1 itself.
    Create Labeled Pod    hello-pod1    ${NAMESPACE}    labels=name=hello-pod,role=hairpin
    Create Labeled Pod    hello-pod2    ${NAMESPACE}    labels=name=hello-pod
    Run With Kubeconfig
    ...    oc create service clusterip test-service --tcp=27017:8080 -n ${NAMESPACE}
    Run With Kubeconfig
    ...    oc set selector service test-service role=hairpin -n ${NAMESPACE}
    Oc Apply    -f ${NETPOL_ASSETS}/netpol-allow-same-namespace.yaml -n ${NAMESPACE}

Teardown Hairpin Test
    [Documentation]    Remove hairpin test resources.
    Run With Kubeconfig    oc delete networkpolicy allow-from-same-namespace -n ${NAMESPACE}    allow_fail=True
    Run With Kubeconfig    oc delete service test-service -n ${NAMESPACE}    allow_fail=True
    Run With Kubeconfig    oc delete pod hello-pod1 hello-pod2 -n ${NAMESPACE} --grace-period=0    allow_fail=True

Setup PodSelector Test    # robocop: off=too-many-calls-in-keyword
    [Documentation]    Create 2 namespaces with labeled pods and apply NetworkPolicies.
    ${ns1}=    Create Random Namespace
    ${ns2}=    Create Random Namespace
    VAR    ${NS_SEL_1}=    ${ns1}    scope=SUITE
    VAR    ${NS_SEL_2}=    ${ns2}    scope=SUITE
    Set Namespace Privileged    ${ns1}
    Set Namespace Privileged    ${ns2}

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
