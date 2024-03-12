*** Settings ***
Documentation       Keywords for testing the MicroShift network

Library             Process
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
    [Arguments]    ${ns}=${NAMESPACE}
    Oc Create    -f ${HELLO_USHIFT} -n ${ns}
    Labeled Pod Should Be Ready    app\=hello-microshift    timeout=120s    ns=${ns}

Expose Hello MicroShift
    [Documentation]    Expose the "hello microshift" application via cluster ip service
    [Arguments]    ${ns}=${NAMESPACE}
    Run With Kubeconfig    oc expose pod hello-microshift -n ${ns}

Expose Hello MicroShift Pod Via LB
    [Documentation]    Expose the "hello microshift" application through the load balancer
    Run With Kubeconfig    oc create service loadbalancer hello-microshift --tcp=${LB_PORT}:8080 -n ${NAMESPACE}

Expose Hello MicroShift Pod Via NodePort
    [Documentation]    Expose the "hello microshift" application through the NodePort
    Run With Kubeconfig
    ...    oc create service nodeport hello-microshift --node-port=${NP_PORT} --tcp=8080 -n ${NAMESPACE}

Access Hello Microshift
    [Documentation]    Try to retrieve data from the "hello microshift" service end point
    [Arguments]    ${ushift_port}    ${ushift_ip}=${USHIFT_HOST}    ${path}=${EMPTY}    ${ns}=${NAMESPACE}

    Run With Kubeconfig    oc get all -n ${ns}

    ${connect_to}=    Set Variable    "hello-microshift.cluster.local::${ushift_ip}:"
    ${url_path}=    Set Variable    "http://hello-microshift.cluster.local:${ushift_port}${path}"

    ${result}=    Run Process
    ...    curl -i ${url_path} --connect-to ${connect_to}
    ...    shell=True
    ...    timeout=15s
    Log Many    ${result.rc}    ${result.stdout}    ${result.stderr}
    Should Be Equal As Integers    ${result.rc}    0
    RETURN    ${result.stdout}

Access Hello MicroShift Success
    [Documentation]    Expect 200 OK when accessing "hello microshift" through the router.
    [Arguments]    ${ushift_port}    ${ushift_ip}=${USHIFT_HOST}    ${path}=${EMPTY}    ${ns}=${NAMESPACE}

    ${result}=    Access Hello MicroShift
    ...    ushift_port=${ushift_port}
    ...    ushift_ip=${ushift_ip}
    ...    path=${path}
    ...    ns=${ns}
    Should Match Regexp    ${result}    HTTP.*200
    Should Match    ${result}    *Hello MicroShift*

Access Hello MicroShift Unavailable
    [Documentation]    Expect 503 Service Unavailable when accessing "hello microshift" through the router.
    [Arguments]    ${ushift_port}    ${ushift_ip}=${USHIFT_HOST}    ${path}=${EMPTY}    ${ns}=${NAMESPACE}

    ${result}=    Access Hello MicroShift
    ...    ushift_port=${ushift_port}
    ...    ushift_ip=${ushift_ip}
    ...    path=${path}
    ...    ns=${ns}
    Should Match Regexp    ${result}    HTTP.*503

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
    [Arguments]    ${ns}=${NAMESPACE}
    Oc Delete    service/hello-microshift -n ${ns}
    Oc Delete    -f ${HELLO_USHIFT} -n ${ns}
