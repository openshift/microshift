*** Settings ***
Documentation       Cross-cluster DNS tests for C2CC.
...                 Verifies CoreDNS server blocks are injected for remote domains,
...                 DNS resolution works across clusters, and service access via
...                 DNS names works end-to-end.

Resource            ../../resources/microshift-process.resource
Resource            ../../resources/kubeconfig.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/c2cc.resource

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           c2cc


*** Variables ***
&{NAMESPACES}       cluster-a=${EMPTY}    cluster-b=${EMPTY}


*** Test Cases ***
Corefile Contains C2CC Server Block On Cluster A
    [Documentation]    Verify Cluster A's Corefile has a server block for Cluster B's domain.
    Verify Corefile Contains C2CC Server Block    cluster-a    ${CLUSTER_B_DOMAIN}

Corefile Contains C2CC Server Block On Cluster B
    [Documentation]    Verify Cluster B's Corefile has a server block for Cluster A's domain.
    Verify Corefile Contains C2CC Server Block    cluster-b    ${CLUSTER_A_DOMAIN}

Resolve Remote Service DNS From Cluster A
    [Documentation]    Verify pod on Cluster A can resolve a service on Cluster B via DNS.
    DNS Resolve From Cluster    cluster-a
    ...    hello-microshift.${NAMESPACES}[cluster-b].svc.${CLUSTER_B_DOMAIN}

Resolve Remote Service DNS From Cluster B
    [Documentation]    Verify pod on Cluster B can resolve a service on Cluster A via DNS.
    DNS Resolve From Cluster    cluster-b
    ...    hello-microshift.${NAMESPACES}[cluster-a].svc.${CLUSTER_A_DOMAIN}

Curl Remote Service Via DNS From Cluster A
    [Documentation]    Verify pod on Cluster A can reach a service on Cluster B using the remote DNS name.
    ${stdout}=    Curl DNS From Cluster    cluster-a
    ...    hello-microshift.${NAMESPACES}[cluster-b].svc.${CLUSTER_B_DOMAIN}    8080
    Should Contain    ${stdout}    Hello from

Curl Remote Service Via DNS From Cluster B
    [Documentation]    Verify pod on Cluster B can reach a service on Cluster A using the remote DNS name.
    ${stdout}=    Curl DNS From Cluster    cluster-b
    ...    hello-microshift.${NAMESPACES}[cluster-a].svc.${CLUSTER_A_DOMAIN}    8080
    Should Contain    ${stdout}    Hello from


*** Keywords ***
Setup
    [Documentation]    Set up clusters and deploy test workloads on both.
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig
    Register Local Cluster    cluster-a
    Register Remote Cluster    cluster-b    ${HOST2_IP}    ${HOST2_SSH_PORT}    ${KUBECONFIG_B}
    Deploy DNS Test Workloads

Teardown
    [Documentation]    Remove test workloads and close connections.
    Cleanup DNS Test Workloads
    Teardown All Remote Clusters
    Remove Kubeconfig
    Logout MicroShift Host

Deploy DNS Test Workloads
    [Documentation]    Create namespace and deploy hello-microshift + curl-pod on both clusters.
    VAR    ${assets}=    ${EXECDIR}/assets/c2cc
    FOR    ${alias}    IN    cluster-a    cluster-b
        ${ns}=    Create Unique Namespace On Cluster    ${alias}
        Set To Dictionary    ${NAMESPACES}    ${alias}    ${ns}
        Oc On Cluster    ${alias}    oc apply -n ${ns} -f ${assets}/hello-microshift.yaml
        Oc On Cluster    ${alias}    oc apply -n ${ns} -f ${assets}/curl-pod.yaml
    END
    Wait For DNS Test Pods

Wait For DNS Test Pods
    [Documentation]    Wait for all test pods to be Ready on both clusters.
    FOR    ${alias}    IN    cluster-a    cluster-b
        Oc On Cluster
        ...    ${alias}
        ...    oc wait pod/hello-microshift pod/curl-pod -n ${NAMESPACES}[${alias}] --for=condition=Ready --timeout=120s
    END

Cleanup DNS Test Workloads
    [Documentation]    Delete test namespace on both clusters. Ignores errors.
    FOR    ${alias}    IN    cluster-a    cluster-b
        Run Keyword And Ignore Error
        ...    Oc On Cluster    ${alias}    oc delete namespace ${NAMESPACES}[${alias}] --timeout=60s
    END

DNS Resolve From Cluster
    [Documentation]    Resolve a DNS name from curl-pod on the given cluster. Retries for up to 60s.
    [Arguments]    ${alias}    ${fqdn}
    Wait Until Keyword Succeeds    12x    5s
    ...    DNS Lookup Should Succeed    ${alias}    ${fqdn}

DNS Lookup Should Succeed
    [Documentation]    Resolve a DNS name from curl-pod using getent hosts.
    [Arguments]    ${alias}    ${fqdn}
    ${stdout}=    Oc On Cluster    ${alias}
    ...    oc exec curl-pod -n ${NAMESPACES}[${alias}] -- getent hosts ${fqdn}
    Should Not Be Empty    ${stdout}

Curl DNS From Cluster
    [Documentation]    Curl a service by DNS name from curl-pod on the given cluster.
    [Arguments]    ${alias}    ${fqdn}    ${port}
    ${stdout}=    Wait Until Keyword Succeeds    12x    5s
    ...    Curl DNS Should Succeed    ${alias}    ${fqdn}    ${port}
    RETURN    ${stdout}

Curl DNS Should Succeed
    [Documentation]    Single attempt to curl a DNS name from curl-pod.
    [Arguments]    ${alias}    ${fqdn}    ${port}
    ${stdout}=    Oc On Cluster    ${alias}
    ...    oc exec curl-pod -n ${NAMESPACES}[${alias}] -- curl -sS --max-time 10 http://${fqdn}:${port}/cgi-bin/hello
    Should Contain    ${stdout}    Hello from
    RETURN    ${stdout}
