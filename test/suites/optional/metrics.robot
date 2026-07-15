*** Settings ***
Documentation       Verify that each optional metrics exporter is running
...                 and producing metrics.

Resource            ../../resources/common.resource
Resource            ../../resources/kubeconfig.resource
Resource            ../../resources/microshift-host.resource

Suite Setup         Setup Metrics Suite
Suite Teardown      Teardown Metrics Suite


*** Variables ***
${METRICS_NS}       openshift-monitoring
${CLIENT_CERT}      /tmp/metrics-test-client.crt
${CLIENT_KEY}       /tmp/metrics-test-client.key


*** Test Cases ***
Kube State Metrics Deployment Is Running
    [Documentation]    The kube-state-metrics deployment should be available.
    ${out}=    Run With Kubeconfig
    ...    oc rollout status deploy/kube-state-metrics -n ${METRICS_NS} --timeout\=120s
    Should Contain    ${out}    successfully rolled out

Kube State Metrics Main Metrics Are Exported
    [Documentation]    The main metrics port (8443) should return kube_node_info.
    ${metrics}=    Scrape Kube State Metrics    8443
    Should Contain    ${metrics}    kube_node_info

Kube State Metrics Self Metrics Are Exported
    [Documentation]    The self/telemetry port (9443) should return process_* metrics.
    ${metrics}=    Scrape Kube State Metrics    9443
    Should Contain    ${metrics}    process_cpu_seconds_total

Node Exporter DaemonSet Is Running
    [Documentation]    The node-exporter daemonset should be available.
    ${out}=    Run With Kubeconfig
    ...    oc rollout status ds/node-exporter -n ${METRICS_NS} --timeout\=120s
    Should Contain    ${out}    successfully rolled out

Node Exporter Metrics Are Exported
    [Documentation]    Port 9100 on localhost should return node_cpu_seconds_total.
    ${metrics}=    Scrape Node Exporter
    Should Contain    ${metrics}    node_cpu_seconds_total

Metrics Server Deployment Is Running
    [Documentation]    The metrics-server deployment should be available.
    ${out}=    Run With Kubeconfig
    ...    oc rollout status deploy/metrics-server -n ${METRICS_NS} --timeout\=120s
    Should Contain    ${out}    successfully rolled out

Metrics Server API Is Available
    [Documentation]    The v1beta1.metrics.k8s.io APIService should report Available.
    ${out}=    Run With Kubeconfig
    ...    oc get apiservice v1beta1.metrics.k8s.io -o jsonpath\='{.status.conditions[?(@.type\=\="Available")].status}'
    Should Be Equal    ${out}    True

Metrics Server Reports Node Metrics
    [Documentation]    oc adm top nodes should return resource usage data.
    ${out}=    Run With Kubeconfig    oc adm top nodes --no-headers
    Should Match Regexp    ${out}    \\d+m


*** Keywords ***
Setup Metrics Suite
    [Documentation]    Login, setup kubeconfig, and extract client certs for mTLS.
    Setup Suite
    Extract Metrics Client Certs

Teardown Metrics Suite
    [Documentation]    Remove temporary client certs and tear down suite.
    Cleanup Metrics Client Certs
    Teardown Suite

Extract Metrics Client Certs
    [Documentation]    Extract the admin kubeconfig client cert and key to temp files
    ...    on the remote host. These are signed by the admin-kubeconfig-signer CA,
    ...    which the kube-rbac-proxy sidecars trust via the metrics-client-ca ConfigMap.
    Command Should Work
    ...    oc config view --kubeconfig=/var/lib/microshift/resources/kubeadmin/$(hostname)/kubeconfig --raw -o jsonpath\='{.users[0].user.client-certificate-data}' | base64 -d > ${CLIENT_CERT}
    Command Should Work
    ...    oc config view --kubeconfig=/var/lib/microshift/resources/kubeadmin/$(hostname)/kubeconfig --raw -o jsonpath\='{.users[0].user.client-key-data}' | base64 -d > ${CLIENT_KEY}

Cleanup Metrics Client Certs
    [Documentation]    Remove temporary client cert files from the remote host.
    Command Should Work    rm -f ${CLIENT_CERT} ${CLIENT_KEY}

Scrape Kube State Metrics
    [Documentation]    Curl the kube-state-metrics pod on the given port with mTLS client certs.
    [Arguments]    ${port}
    ${pod_ip}=    Run With Kubeconfig
    ...    oc get pod -n ${METRICS_NS} -l app.kubernetes.io/name\=kube-state-metrics -o jsonpath\='{.items[0].status.podIP}'
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
