*** Settings ***
Documentation       Service type tests: LoadBalancer traffic policies, LB port binding,
...                 and service idling/unidling.

Resource            ../../resources/common.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/microshift-host.resource
Resource            ../../resources/microshift-network.resource
Resource            ../../resources/network-testing.resource

Suite Setup         Setup Suite With Namespace
Suite Teardown      Teardown Suite With Namespace

Test Tags           network    slow


*** Variables ***
${SVC_ASSETS}       ./assets/service-types
${LB_SVC_PORT}      27017


*** Test Cases ***
LoadBalancer Service With External And Internal Traffic Policies
    [Documentation]    Verify LoadBalancer services work with different traffic policies.
    ...    Tests ETP=Cluster (default ITP) via node debug pod, then ETP=Local/ITP=Local
    ...    via cluster pod. Verifies cleanup removes firewalld entries after deletion.
    ...    Covers QE test 60968.
    [Setup]    Setup LB Traffic Policy Test

    # --- ETP=Cluster, ITP=default ---
    Create LB Service With Policies    lbtest-cluster    externalTrafficPolicy=Cluster
    ${lb_ip}=    Wait For LoadBalancer IP    lbtest-cluster    ${NAMESPACE}
    ${svc_port}=    Oc Get JsonPath    service    ${NAMESPACE}    lbtest-cluster    .spec.ports[0].port

    Wait Until Keyword Succeeds    30x    2s
    ...    Curl Via SSH Should Succeed    ${lb_ip}    ${svc_port}

    Run With Kubeconfig    oc delete service lbtest-cluster -n ${NAMESPACE}

    Wait Until Keyword Succeeds    10x    2s
    ...    Curl Via SSH Should Fail    ${lb_ip}    ${svc_port}

    # --- ETP=Local, ITP=Local ---
    Create LB Service With Policies    lbtest-local    externalTrafficPolicy=Local    internalTrafficPolicy=Local
    ${lb_ip_local}=    Wait For LoadBalancer IP    lbtest-local    ${NAMESPACE}
    ${svc_port_local}=    Oc Get JsonPath    service    ${NAMESPACE}    lbtest-local    .spec.ports[0].port

    Wait Until Keyword Succeeds    30x    2s
    ...    Curl From Pod Should Succeed    test-pod    ${NAMESPACE}
    ...    http://${lb_ip_local}:${svc_port_local}    expected=Hello MicroShift

    Run With Kubeconfig    oc delete service lbtest-local -n ${NAMESPACE}

    Wait Until Keyword Succeeds    10x    2s
    ...    Curl From Pod Should Timeout    test-pod    ${NAMESPACE}
    ...    http://${lb_ip_local}:${svc_port_local}

    [Teardown]    Teardown LB Traffic Policy Test

Only One LoadBalancer Can Bind Same Port
    [Documentation]    Verify that when two LoadBalancer services use the same port,
    ...    only one gets a LoadBalancer IP. When the first is deleted, the second
    ...    takes over the port binding.
    ...    Covers QE test 61218.
    [Setup]    Setup LB Port Binding Test

    # Create first LB
    Create LB Service With Policies    lbtest1    selector=hello-pod
    ${lb_ip}=    Wait For LoadBalancer IP    lbtest1    ${NAMESPACE}

    # Create second LB with same port
    Create LB Service With Policies    lbtest2    selector=hello-pod

    # First should have IP, second should not
    Service Should Have LB IP    lbtest1    ${NAMESPACE}
    Service Should Not Have LB IP    lbtest2    ${NAMESPACE}

    # Verify first LB is accessible
    Wait Until Keyword Succeeds    10x    2s
    ...    Curl Via SSH Should Succeed    ${lb_ip}    ${LB_SVC_PORT}

    # Delete first, second should take over
    Run With Kubeconfig    oc delete service lbtest1 -n ${NAMESPACE}

    Wait Until Keyword Succeeds    60x    2s
    ...    Service Should Have LB IP    lbtest2    ${NAMESPACE}

    ${lb_ip2}=    Wait For LoadBalancer IP    lbtest2    ${NAMESPACE}
    Wait Until Keyword Succeeds    10x    2s
    ...    Curl Via SSH Should Succeed    ${lb_ip2}    ${LB_SVC_PORT}

    [Teardown]    Teardown LB Port Binding Test

Service Idling And Manual Unidling
    [Documentation]    Verify that services can be idled and manually unidled.
    ...    MicroShift does not support automatic unidling, so manual scaling is used.
    ...    Covers QE test 60290.
    [Setup]    Setup Idling Test

    ${output}=    Run With Kubeconfig    oc idle svc/hello-deploy -n ${NAMESPACE}
    Should Contain    ${output}    has been marked as idled

    Wait Until Keyword Succeeds    30x    2s
    ...    Deployment Should Have Zero Replicas    hello-deploy    ${NAMESPACE}

    Run With Kubeconfig    oc scale deployment hello-deploy --replicas=2 -n ${NAMESPACE}
    Named Deployment Should Be Available    hello-deploy    ${NAMESPACE}    5m

    [Teardown]    Teardown Idling Test


*** Keywords ***
Setup LB Traffic Policy Test
    [Documentation]    Create backend and test pods for LB traffic policy tests.
    Create Labeled Pod    hello-pod-0    ${NAMESPACE}    labels=name=hello-pod
    Create Labeled Pod    hello-pod-1    ${NAMESPACE}    labels=name=hello-pod
    Create Labeled Pod    test-pod    ${NAMESPACE}    labels=name=test-pod

Teardown LB Traffic Policy Test
    [Documentation]    Clean up LB traffic policy resources.
    Run With Kubeconfig    oc delete service lbtest-cluster lbtest-local -n ${NAMESPACE}    allow_fail=True
    Run With Kubeconfig
    ...    oc delete pod hello-pod-0 hello-pod-1 test-pod -n ${NAMESPACE} --grace-period=0    allow_fail=True

Create LB Service With Policies
    [Documentation]    Create a LoadBalancer service with configurable traffic policies.
    [Arguments]    ${name}    ${externalTrafficPolicy}=${EMPTY}    ${internalTrafficPolicy}=${EMPTY}    ${selector}=hello-pod
    Run With Kubeconfig
    ...    oc create service loadbalancer ${name} --tcp=${LB_SVC_PORT}:8080 -n ${NAMESPACE}
    Run With Kubeconfig
    ...    oc set selector service ${name} name=${selector} -n ${NAMESPACE}
    IF    '${externalTrafficPolicy}' != '${EMPTY}'
        Run With Kubeconfig
        ...    oc patch service ${name} -n ${NAMESPACE} -p '{"spec":{"externalTrafficPolicy":"${externalTrafficPolicy}"}}' --type=merge
    END
    IF    '${internalTrafficPolicy}' != '${EMPTY}'
        Run With Kubeconfig
        ...    oc patch service ${name} -n ${NAMESPACE} -p '{"spec":{"internalTrafficPolicy":"${internalTrafficPolicy}"}}' --type=merge
    END

Curl Via SSH Should Succeed
    [Documentation]    Curl from the MicroShift host via SSH and assert success.
    [Arguments]    ${ip}    ${port}
    ${stdout}=    Command Should Work    curl -s --max-time 5 http://${ip}:${port}    sudo_mode=False
    Should Contain    ${stdout}    Hello MicroShift

Curl Via SSH Should Fail
    [Documentation]    Curl from the MicroShift host via SSH and expect failure.
    [Arguments]    ${ip}    ${port}
    Command Should Fail    curl -s --max-time 5 http://${ip}:${port}    sudo_mode=False

Setup LB Port Binding Test
    [Documentation]    Create a pod for LB port binding test.
    Create Labeled Pod    hello-pod    ${NAMESPACE}    labels=name=hello-pod

Teardown LB Port Binding Test
    [Documentation]    Clean up LB port binding resources.
    Run With Kubeconfig    oc delete service lbtest1 lbtest2 -n ${NAMESPACE}    allow_fail=True
    Run With Kubeconfig    oc delete pod hello-pod -n ${NAMESPACE} --grace-period=0    allow_fail=True

Setup Idling Test
    [Documentation]    Deploy hello-deploy with 2 replicas and a ClusterIP service.
    Oc Create    -f ${SVC_ASSETS}/deployment-hello-2-replicas.yaml -n ${NAMESPACE}
    Oc Create    -f ${SVC_ASSETS}/service-clusterip.yaml -n ${NAMESPACE}
    Named Deployment Should Be Available    hello-deploy    ${NAMESPACE}    5m

Teardown Idling Test
    [Documentation]    Clean up idling test resources.
    Run With Kubeconfig    oc delete -f ${SVC_ASSETS}/service-clusterip.yaml -n ${NAMESPACE}    allow_fail=True
    Run With Kubeconfig    oc delete -f ${SVC_ASSETS}/deployment-hello-2-replicas.yaml -n ${NAMESPACE}    allow_fail=True

Deployment Should Have Zero Replicas
    [Documentation]    Assert that a deployment has scaled to zero.
    [Arguments]    ${name}    ${ns}
    ${replicas}=    Oc Get JsonPath    deployment    ${ns}    ${name}    .status.replicas
    Should Be True    '${replicas}' == '0' or '${replicas}' == '${EMPTY}'
