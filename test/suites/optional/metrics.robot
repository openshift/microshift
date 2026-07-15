*** Settings ***
Documentation       Verify that each optional metrics exporter is running
...                 and producing metrics.

Resource            ../../resources/common.resource
Resource            ../../resources/kubeconfig.resource
Resource            ../../resources/microshift-host.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/optional-config.resource

Suite Setup         Setup
Suite Teardown      Teardown


*** Variables ***
${METRICS_NS}       openshift-monitoring
${CLIENT_CERT}      /tmp/metrics-test-client.crt
${CLIENT_KEY}       /tmp/metrics-test-client.key


*** Test Cases ***
Kube State Metrics Deployment Is Running
    [Documentation]    The kube-state-metrics deployment should be available.
    Named Deployment Should Be Available    kube-state-metrics    ns=${METRICS_NS}

Kube State Metrics Main Metrics Are Exported
    [Documentation]    The main metrics port (8443) should return kube_node_info.
    Wait Until Keyword Succeeds    60s    5s
    ...    Metrics Endpoint Should Contain    8443    kube_node_info

Kube State Metrics Self Metrics Are Exported
    [Documentation]    The self/telemetry port (9443) should return process_* metrics.
    Wait Until Keyword Succeeds    60s    5s
    ...    Metrics Endpoint Should Contain    9443    process_cpu_seconds_total

Node Exporter DaemonSet Is Running
    [Documentation]    The node-exporter daemonset should be available.
    Named Daemonset Should Be Available    node-exporter    ns=${METRICS_NS}

Node Exporter Metrics Are Exported
    [Documentation]    Port 9100 on localhost should return node_cpu_seconds_total.
    Wait Until Keyword Succeeds    60s    5s
    ...    Node Exporter Should Contain    node_cpu_seconds_total

Metrics Server Deployment Is Running
    [Documentation]    The metrics-server deployment should be available.
    Named Deployment Should Be Available    metrics-server    ns=${METRICS_NS}

Metrics Server API Is Available
    [Documentation]    The v1beta1.metrics.k8s.io APIService should report Available.
    Oc Wait    apiservice v1beta1.metrics.k8s.io    --for=condition=Available --timeout\=120s

Metrics Server Reports Node Metrics
    [Documentation]    oc adm top nodes should return resource usage data.
    ${out}=    Run With Kubeconfig    oc adm top nodes --no-headers
    Should Match Regexp    ${out}    \\d+m


*** Keywords ***
Setup
    [Documentation]    Login, setup kubeconfig, enable metrics optionals, and extract client certs for mTLS.
    Setup Suite
    Setup MicroShift With Optionals
    ...    080-microshift-metrics-server
    ...    081-microshift-kube-state-metrics
    ...    082-microshift-node-exporter
    Extract Metrics Client Certs

Teardown
    [Documentation]    Remove temporary client certs, restore config, and tear down suite.
    Cleanup Metrics Client Certs
    Teardown MicroShift With Optionals
    Teardown Suite

Extract Metrics Client Certs
    [Documentation]    Extract the admin kubeconfig client cert and key to temp files
    ...    on the remote host. These are signed by the admin-kubeconfig-signer CA,
    ...    which the kube-rbac-proxy sidecars trust via the metrics-client-ca ConfigMap.
    ...    Uses SSH (Command Should Work) because the cert files must exist on the
    ...    remote host where curl runs, not locally.
    Command Should Work
    ...    oc config view --kubeconfig=/var/lib/microshift/resources/kubeadmin/kubeconfig --raw -o jsonpath\='{.users[0].user.client-certificate-data}' | base64 -d > ${CLIENT_CERT}
    Command Should Work
    ...    oc config view --kubeconfig=/var/lib/microshift/resources/kubeadmin/kubeconfig --raw -o jsonpath\='{.users[0].user.client-key-data}' | base64 -d > ${CLIENT_KEY}

Cleanup Metrics Client Certs
    [Documentation]    Remove temporary client cert files from the remote host.
    Command Should Work    rm -f ${CLIENT_CERT} ${CLIENT_KEY}

Metrics Endpoint Should Contain
    [Documentation]    Scrape kube-state-metrics on the given port and assert the
    ...    response contains the expected metric name.
    [Arguments]    ${port}    ${expected}
    ${metrics}=    Scrape Kube State Metrics    ${port}
    Should Contain    ${metrics}    ${expected}

Node Exporter Should Contain
    [Documentation]    Scrape node-exporter and assert the response contains the
    ...    expected metric name.
    [Arguments]    ${expected}
    ${metrics}=    Scrape Node Exporter
    Should Contain    ${metrics}    ${expected}

Scrape Kube State Metrics
    [Documentation]    Curl the kube-state-metrics pod on the given port with mTLS client certs.
    [Arguments]    ${port}
    ${pod_ip}=    Oc Get JsonPath
    ...    pod
    ...    ${METRICS_NS}
    ...    ${EMPTY}
    ...    .items[0].status.podIP
    ...    label=app.kubernetes.io/name\=kube-state-metrics
    Should Not Be Empty    ${pod_ip}
    ${stdout}=    Command Should Work
    ...    curl -sk --cert ${CLIENT_CERT} --key ${CLIENT_KEY} https://${pod_ip}:${port}/metrics
    Should Not Be Empty    ${stdout}
    RETURN    ${stdout}

Scrape Node Exporter
    [Documentation]    Curl the node-exporter on localhost:9100 with mTLS client certs.
    ...    node-exporter uses hostNetwork and hostPort so it is reachable on the host.
    ${stdout}=    Command Should Work
    ...    curl -sk --cert ${CLIENT_CERT} --key ${CLIENT_KEY} https://localhost:9100/metrics
    Should Not Be Empty    ${stdout}
    RETURN    ${stdout}
