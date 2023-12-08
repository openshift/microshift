*** Settings ***
Documentation       Keywords for testing the MicroShift network

Library             Process
Library             RequestsLibrary
Library             SSHLibrary
Resource            oc.resource
Resource            common.resource


*** Variables ***
${HELLO_USHIFT}     ./assets/hello-microshift.yaml
${LB_PORT}          5678
${NP_PORT}          32222
${HTTP_PORT}        80


*** Keywords ***
Create Hello MicroShift Pod
    [Documentation]    Create a pod running the "hello microshift" application
    Oc Create    -f ${HELLO_USHIFT} -n ${NAMESPACE}
    Labeled Pod Should Be Ready    app\=hello-microshift    timeout=120s

Expose Hello MicroShift
    [Documentation]    Expose the "hello microshift" application via cluster ip service
    Run With Kubeconfig    oc expose pod hello-microshift -n ${NAMESPACE}

Expose Hello MicroShift Pod Via LB
    [Documentation]    Expose the "hello microshift" application through the load balancer
    Run With Kubeconfig    oc create service loadbalancer hello-microshift --tcp=${LB_PORT}:8080 -n ${NAMESPACE}

Expose Hello MicroShift Pod Via NodePort
    [Documentation]    Expose the "hello microshift" application through the NodePort
    Run With Kubeconfig
    ...    oc create service nodeport hello-microshift --node-port=${NP_PORT} --tcp=8080 -n ${NAMESPACE}

Access Hello Microshift
    [Documentation]    Try to retrieve data from the "hello microshift" service end point
    [Arguments]    ${ushift_port}    ${ushift_ip}=${USHIFT_HOST}    ${path}=${EMPTY}
    # We do not set up DNS on the hypervisor in a way that allows
    # connections with only the URL, so we add the 'proxies' argument
    # with a per-host proxy for the VM under test using the URL that
    # will be recognized by ingress mapped to the host IP and port
    # created by the test harness. Note the use of '&' instead of '$'
    # in the variable name here and where the value is passed to GET.
    &{proxies}=    Create Dictionary    http://hello-microshift.cluster.local=http://${ushift_ip}:${ushift_port}
    ${response}=    GET
    ...    http://hello-microshift.cluster.local${path}
    ...    expected_status=200
    ...    verify=False
    ...    proxies=&{proxies}
    ...    timeout=15
    Log Many    ${response}
    Should Match    ${response.text}    *Hello MicroShift*

Verify Hello MicroShift LB
    [Documentation]    Run Hello MicroShift Load Balancer verification
    Create Hello MicroShift Pod
    Expose Hello MicroShift Pod Via LB
    Wait Until Keyword Succeeds    30x    10s
    ...    Access Hello Microshift    ${LB_PORT}

    [Teardown]    Run Keywords
    ...    Delete Hello MicroShift Pod And Service

Delete Hello MicroShift Pod And Service
    [Documentation]    Delete service and pod for cleanup.
    Oc Delete    service/hello-microshift -n ${NAMESPACE}
    Oc Delete    -f ${HELLO_USHIFT} -n ${NAMESPACE}
