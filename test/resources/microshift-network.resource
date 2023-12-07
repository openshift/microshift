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

    ${connect_to}=    Set Variable    "hello-microshift.cluster.local:${HTTP_PORT}:${ushift_ip}:${ushift_port}"
    ${url_path}=    Set Variable    "http://hello-microshift.cluster.local${path}"

    ${result}=    Run Process
    ...    curl -i ${url_path} --connect-to ${connect_to}
    ...    shell=True
    ...    timeout=15s
    Log Many    ${result.rc}    ${result.stdout}    ${result.stderr}
    Should Be Equal As Integers    ${result.rc}    0
    Should Match Regexp    ${result.stdout}    HTTP.*200
    Should Match    ${result.stdout}    *Hello MicroShift*

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
