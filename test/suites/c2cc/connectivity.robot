*** Settings ***
Documentation       Cross-cluster connectivity tests for C2CC.
...                 Deploys test workloads on both clusters and verifies pod-to-pod
...                 and pod-to-service communication in both directions.

Resource            ../../resources/microshift-process.resource
Resource            ../../resources/kubeconfig.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/c2cc.resource

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           c2cc


*** Variables ***
&{NAMESPACES}       cluster-a=${EMPTY}    cluster-b=${EMPTY}    cluster-c=${EMPTY}


*** Test Cases ***
Test Cross Cluster Connectivity
    [Documentation]    Verify pods on all clusters can reach pods/services on all other clusters.
    [Template]    Test Connectivity Between Clusters
    cluster-a    cluster-b    pod
    cluster-a    cluster-b    service
    cluster-a    cluster-c    pod
    cluster-a    cluster-c    service
    cluster-b    cluster-a    pod
    cluster-b    cluster-a    service
    cluster-b    cluster-c    pod
    cluster-b    cluster-c    service
    cluster-c    cluster-a    pod
    cluster-c    cluster-a    service
    cluster-c    cluster-b    pod
    cluster-c    cluster-b    service

Test Cross Cluster Source IP Preservation
    [Documentation]    Verify cross cluster traffic preserves source pod IP (no SNAT).
    [Template]    Test Source IP Preserved Between Clusters
    cluster-a    cluster-b    pod
    cluster-a    cluster-b    service
    cluster-a    cluster-c    pod
    cluster-a    cluster-c    service
    cluster-b    cluster-a    pod
    cluster-b    cluster-a    service
    cluster-b    cluster-c    pod
    cluster-b    cluster-c    service
    cluster-c    cluster-a    pod
    cluster-c    cluster-a    service
    cluster-c    cluster-b    pod
    cluster-c    cluster-b    service


*** Keywords ***
Setup
    [Documentation]    Set up clusters and deploy test workloads on all.
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig
    Register Local Cluster    cluster-a
    Register Remote Cluster    cluster-b    ${HOST2_IP}    ${HOST2_SSH_PORT}    ${KUBECONFIG_B}
    Register Remote Cluster    cluster-c    ${HOST3_IP}    ${HOST3_SSH_PORT}    ${KUBECONFIG_C}
    Deploy Test Workloads

Teardown
    [Documentation]    Remove test workloads and close connections.
    Cleanup Test Workloads
    Teardown All Remote Clusters
    Remove Kubeconfig
    Logout MicroShift Host

Test Connectivity Between Clusters
    [Documentation]    Verify pod on ${source} can reach ${endpoint_type} IP on ${destination}
    [Arguments]    ${source}    ${destination}    ${endpoint_type}
    IF    '${endpoint_type}' == 'pod'
        ${ip_dest}=    Get Hello Pod IP    ${destination}
    ELSE IF    '${endpoint_type}' == 'service'
        ${ip_dest}=    Get Hello Service IP    ${destination}
    ELSE
        Fail    Invalid endpoint_type: ${endpoint_type}. Must be 'pod' or 'service'.
    END

    ${stdout}=    Curl From Cluster    ${source}    ${ip_dest}    8080
    Should Contain    ${stdout}    Hello from

Test Source IP Preserved Between Clusters
    [Documentation]    Verify ${source} to ${destination} pod-to-${endpoint_type} traffic preserves the source pod IP (no SNAT).
    [Arguments]    ${source}    ${destination}    ${endpoint_type}
    ${curl_pod_ip}=    Get Curl Pod IP    ${source}
    IF    '${endpoint_type}' == 'pod'
        ${ip_dest}=    Get Hello Pod IP    ${destination}
    ELSE IF    '${endpoint_type}' == 'service'
        ${ip_dest}=    Get Hello Service IP    ${destination}
    ELSE
        Fail    Invalid endpoint_type: ${endpoint_type}. Must be 'pod' or 'service'.
    END

    ${stdout}=    Curl From Cluster    ${source}    ${ip_dest}    8080
    Should Contain    ${stdout}    source: ${curl_pod_ip}

Deploy Test Workloads
    [Documentation]    Create namespace and deploy hello-microshift + curl-pod on all clusters.
    VAR    ${assets}=    ${EXECDIR}/assets/c2cc
    FOR    ${alias}    IN    cluster-a    cluster-b    cluster-c
        Oc On Cluster    ${alias}    oc create namespace ${NAMESPACE}
        Oc On Cluster    ${alias}    oc apply -n ${NAMESPACE} -f ${assets}/hello-microshift.yaml
        Oc On Cluster    ${alias}    oc apply -n ${NAMESPACE} -f ${assets}/curl-pod.yaml
    END
    Wait For Test Pods

Wait For Test Pods
    [Documentation]    Wait for all test pods to be Ready on all clusters.
    FOR    ${alias}    IN    cluster-a    cluster-b    cluster-c
        Oc On Cluster    ${alias}
        ...    oc wait pod/hello-microshift pod/curl-pod -n ${NAMESPACE} --for=condition=Ready --timeout=120s
    END

Cleanup Test Workloads
    [Documentation]    Delete test namespace on all clusters. Ignores errors.
    FOR    ${alias}    IN    cluster-a    cluster-b    cluster-c
        Run Keyword And Ignore Error
        ...    Oc On Cluster    ${alias}    oc delete namespace ${NAMESPACE} --timeout=60s
    END

Get Hello Pod IP
    [Documentation]    Get the pod IP of hello-microshift on the given cluster.
    [Arguments]    ${alias}
    ${ip}=    Oc On Cluster    ${alias}
    ...    oc get pod hello-microshift -n ${NAMESPACES}[${alias}] -o jsonpath='{.status.podIP}'
    RETURN    ${ip}

Get Hello Service IP
    [Documentation]    Get the ClusterIP of the hello-microshift service on the given cluster.
    [Arguments]    ${alias}
    ${ip}=    Oc On Cluster    ${alias}
    ...    oc get svc hello-microshift -n ${NAMESPACES}[${alias}] -o jsonpath='{.spec.clusterIP}'
    RETURN    ${ip}

Get Curl Pod IP
    [Documentation]    Get the pod IP of curl-pod on the given cluster.
    [Arguments]    ${alias}
    ${ip}=    Oc On Cluster    ${alias}
    ...    oc get pod curl-pod -n ${NAMESPACES}[${alias}] -o jsonpath='{.status.podIP}'
    RETURN    ${ip}

Curl From Cluster
    [Documentation]    Exec curl from curl-pod on the given cluster to the target IP and port.
    [Arguments]    ${alias}    ${ip}    ${port}
    ${stdout}=    Oc On Cluster    ${alias}
    ...    oc exec curl-pod -n ${NAMESPACES}[${alias}] -- curl -sS --max-time 10 http://${ip}:${port}/cgi-bin/hello
    RETURN    ${stdout}
