*** Settings ***
Documentation       Host networking tests: hostPort, br-ex NM state, conntrack cleanup.

Resource            ../../resources/common.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/microshift-host.resource
Resource            ../../resources/microshift-network.resource
Resource            ../../resources/network-testing.resource

Suite Setup         Setup Suite With Namespace
Suite Teardown      Teardown Suite With Namespace

Test Tags           network


*** Variables ***
${HOSTPORT_POD}         ./assets/host-networking/hostport-pod.yaml
${UDP_LISTENER_POD}     ./assets/host-networking/udp-listener-pod.yaml
${HOSTPORT}             9500


*** Test Cases ***
Pod Should Be Accessible Via Node IP And Host Port
    [Documentation]    Verify a pod with hostPort is reachable from the node IP.
    ...    Covers QE test 60550.
    [Setup]    Create HostPort Pod

    Wait Until Keyword Succeeds    30x    2s
    ...    HostPort Should Be Reachable

    [Teardown]    Oc Delete    -f ${HOSTPORT_POD} -n ${NAMESPACE}

Br-ex Should Be Unmanaged By NetworkManager
    [Documentation]    Verify that br-ex interface is not managed by NetworkManager.
    ...    Covers QE test 65838.
    ${stdout}=    Command Should Work    nmcli conn show
    Should Not Contain    ${stdout}    br-ex

Conntrack Entry Should Be Cleaned Up When UDP NodePort Endpoint Is Deleted
    [Documentation]    Verify that OVN-Kubernetes removes conntrack entries for UDP
    ...    traffic when the service endpoint pod is deleted.
    ...    Covers QE test 64752.
    [Setup]    Setup UDP Conntrack Test

    ${udp_pod_ip}=    Get Pod IP    udp-pod    ${NAMESPACE}
    ${node_port}=    Get Service NodePort    udp-pod    ${NAMESPACE}

    Send UDP Traffic    ${node_port}
    Wait Until Keyword Succeeds    10x    2s
    ...    Conntrack Should Contain    udp    8080    ${udp_pod_ip}

    Oc Delete    pod/udp-pod -n ${NAMESPACE}
    Oc Wait    pod/udp-pod -n ${NAMESPACE}    --for=delete --timeout=120s

    Wait Until Keyword Succeeds    30x    2s
    ...    Conntrack Should Not Contain    udp    8080    ${udp_pod_ip}

    [Teardown]    Run Keywords
    ...    Run With Kubeconfig    oc delete service udp-pod -n ${NAMESPACE}    allow_fail=True
    ...    AND
    ...    Run With Kubeconfig    oc delete -f ${UDP_LISTENER_POD} -n ${NAMESPACE}    allow_fail=True


*** Keywords ***
Create HostPort Pod
    [Documentation]    Create a pod with hostPort and wait for it to be ready.
    Oc Create    -f ${HOSTPORT_POD} -n ${NAMESPACE}
    Labeled Pod Should Be Ready    app\=hostport-pod    ns=${NAMESPACE}

HostPort Should Be Reachable
    [Documentation]    Curl the hostPort from the MicroShift host and assert success.
    ${stdout}=    Command Should Work    curl -s --max-time 5 http://127.0.0.1:${HOSTPORT}
    Should Contain    ${stdout}    Hello MicroShift

Setup UDP Conntrack Test
    [Documentation]    Deploy the UDP listener pod and expose it as a NodePort service.
    Oc Create    -f ${UDP_LISTENER_POD} -n ${NAMESPACE}
    Labeled Pod Should Be Ready    app\=udp-pod    ns=${NAMESPACE}
    Run With Kubeconfig
    ...    oc expose pod udp-pod --type=NodePort --port=8080 --protocol=UDP -n ${NAMESPACE}
    ${node_port}=    Get Service NodePort    udp-pod    ${NAMESPACE}
    Log    UDP NodePort: ${node_port}

Send UDP Traffic
    [Documentation]    Send UDP packets to the given NodePort from the host.
    [Arguments]    ${port}
    Command Should Work
    ...    bash -c 'for n in 1 2 3; do echo $n > /dev/udp/127.0.0.1/${port}; sleep 0.5; done'
    ...    sudo_mode=False
