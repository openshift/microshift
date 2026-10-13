*** Settings ***
Documentation       Keywords for testing the MicroShift network

Library             Process
Library             SSHLibrary
Library             libipv6.py
Resource            oc.resource
Resource            common.resource


*** Variables ***
${HELLO_USHIFT}     ./assets/hello-microshift.yaml
${LB_PORT}          5678
${NP_PORT}          32222
${HTTP_PORT}        80
${HTTPS_PORT}       443


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
    [Documentation]    Use curl to make a GET request to the specified scheme, host, port and path.
    ...    Use hello-microshift as authority and swap it with --connect-to option.
    ...    Returns all output and return code.
    ...    Times out after 15s.
    [Arguments]
    ...    ${ushift_port}
    ...    ${ushift_ip}=${USHIFT_HOST}
    ...    ${path}=${EMPTY}
    ...    ${scheme}=http
    ...    ${hostname}=hello-microshift.cluster.local

    ${ip}=    Add Brackets If Ipv6    ${ushift_ip}

    ${connect_to}=    Set Variable    "${hostname}::${ip}:"
    ${url_path}=    Set Variable    "${scheme}://${hostname}:${ushift_port}${path}"

    ${result}=    Run Process
    ...    curl -k -i ${url_path} --connect-to ${connect_to}
    ...    shell=True
    ...    timeout=15s
    RETURN    ${result.rc}    ${result.stdout}    ${result.stderr}

Access Hello MicroShift Success
    [Documentation]    Expect 200 OK when accessing "hello microshift" through the router.
    [Arguments]
    ...    ${ushift_port}
    ...    ${ushift_ip}=${USHIFT_HOST}
    ...    ${path}=${EMPTY}
    ...    ${scheme}=http
    ...    ${hostname}=hello-microshift.cluster.local

    ${rc}    ${stdout}    ${stderr}=    Access Hello MicroShift
    ...    ushift_port=${ushift_port}
    ...    ushift_ip=${ushift_ip}
    ...    path=${path}
    ...    scheme=${scheme}
    ...    hostname=${hostname}
    Should Be Equal As Integers    ${rc}    0
    Should Match Regexp    ${stdout}    HTTP.*200
    Should Match    ${stdout}    *Hello MicroShift*

Access Hello MicroShift No Route
    [Documentation]    Expect 503 Service Unavailable when accessing "hello microshift" through the router.
    [Arguments]
    ...    ${ushift_port}
    ...    ${ushift_ip}=${USHIFT_HOST}
    ...    ${path}=${EMPTY}
    ...    ${scheme}=http
    ...    ${hostname}=hello-microshift.cluster.local

    ${rc}    ${stdout}    ${stderr}=    Access Hello MicroShift
    ...    ushift_port=${ushift_port}
    ...    ushift_ip=${ushift_ip}
    ...    path=${path}
    ...    scheme=${scheme}
    ...    hostname=${hostname}
    Should Be Equal As Integers    ${rc}    0
    Should Match Regexp    ${stdout}    HTTP.*503

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

Skip If Service Network Has No IPv4
    [Documentation]    Skip the test when MicroShift's service network has no IPv4 family.
    ...    MicroShift only computes and configures the IPv4 KAS advertise address (e.g.
    ...    10.44.0.0) when the service network includes an IPv4 CIDR, so IPv4-literal checks
    ...    cannot pass on an IPv6 single stack deployment. The service network family, not the
    ...    host NIC addresses, is the correct signal: an IPv6 single stack VM still carries
    ...    IPv4 management addresses, so a host-level IPv4 check would fail to skip.
    ${families}=    Run With Kubeconfig
    ...    oc get service kubernetes -n default -o jsonpath={.spec.ipFamilies[*]}
    # Resolve the membership test with a keyword so the raw command output is never
    # substituted into the Skip If Python expression; ${has_ipv4} is a boolean.
    ${has_ipv4}=    Run Keyword And Return Status
    ...    Should Contain    ${families}    IPv4
    Skip If    not ${has_ipv4}
    ...    MicroShift service network has no IPv4 family (IPv6 single stack); skipping IPv4-only test
