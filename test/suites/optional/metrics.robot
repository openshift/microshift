*** Settings ***
Documentation       Verify that each optional metrics exporter is running
...                 and producing metrics.

Resource            ../../resources/common.resource
Resource            ../../resources/kubeconfig.resource
Resource            ../../resources/microshift-host.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/optional-config.resource
Resource            ../../resources/microshift-process.resource

Suite Setup         Setup
Suite Teardown      Teardown


*** Variables ***
${METRICS_NS}           openshift-monitoring
${CLIENT_CERT}          /tmp/metrics-test-client.crt
${CLIENT_KEY}           /tmp/metrics-test-client.key
${SERVICE_CA}           /tmp/metrics-test-service-ca.crt
${RESTART_ATTEMPTS}     3


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

Metrics Server Recovers After Restart Storm
    [Documentation]    Service-ca must recreate the deleted metrics-server serving Secret
    ...    after MicroShift is interrupted repeatedly during startup. The retry ceiling
    ...    makes the pre-fix failure deterministic without creating a certificate here.
    ${old_uid}=    Simulate Metrics Server Serving Certificate Retry Ceiling
    FOR    ${attempt}    IN RANGE    ${RESTART_ATTEMPTS}
        Stop MicroShift
        Start MicroShift Without Waiting For Systemd Readiness
        Sleep    1s
        Restart MicroShift
    END
    Wait Until Keyword Succeeds    5m    5s
    ...    Metrics Server Serving Certificate Secret Should Be Recreated    ${old_uid}
    Named Deployment Should Be Available    metrics-server    ns=${METRICS_NS}
    Metrics Server API Should Be Available
    [Teardown]    Clear Metrics Server Serving Certificate Failure State


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
    Command Should Work
    ...    oc get configmap signing-cabundle -n openshift-service-ca --kubeconfig\=/var/lib/microshift/resources/kubeadmin/kubeconfig -o jsonpath\='{.data.ca-bundle\\.crt}' > ${SERVICE_CA}

Cleanup Metrics Client Certs
    [Documentation]    Remove temporary client cert files from the remote host.
    Command Should Work    rm -f ${CLIENT_CERT} ${CLIENT_KEY} ${SERVICE_CA}

Simulate Metrics Server Serving Certificate Retry Ceiling
    [Documentation]    Delete the service-ca-managed Secret after setting the controller's
    ...    documented retry ceiling. The test never writes certificate data itself.
    ${old_uid}=    Run With Kubeconfig
    ...    oc get secret metrics-server-tls -n ${METRICS_NS} -o jsonpath\\='{.metadata.uid}'
    Should Not Be Empty    ${old_uid}
    Run With Kubeconfig
    ...    oc annotate service metrics-server -n ${METRICS_NS} --overwrite service.beta.openshift.io/serving-cert-generation-error-num=10
    Run With Kubeconfig
    ...    oc annotate service metrics-server -n ${METRICS_NS} --overwrite service.alpha.openshift.io/serving-cert-generation-error-num=10
    ${retry_count}=    Run With Kubeconfig
    ...    oc get service metrics-server -n ${METRICS_NS} -o jsonpath\\='{.metadata.annotations.service\\.beta\\.openshift\\.io/serving-cert-generation-error-num}'
    Should Be Equal As Integers    ${retry_count}    10
    Run With Kubeconfig    oc delete secret metrics-server-tls -n ${METRICS_NS}
    Metrics Server Serving Certificate Secret Should Be Absent
    RETURN    ${old_uid}

Metrics Server Serving Certificate Secret Should Be Absent
    [Documentation]    Prove the service-ca serving Secret was deleted before recovery.
    ${output}    ${rc}=    Run With Kubeconfig
    ...    oc get secret metrics-server-tls -n ${METRICS_NS}
    ...    allow_fail=${TRUE}
    ...    return_rc=${TRUE}
    Should Not Be Equal As Integers    ${rc}    0

Metrics Server Serving Certificate Secret Should Be Recreated
    [Documentation]    Verify service-ca recreated a new TLS Secret with its ownership annotation.
    [Arguments]    ${old_uid}
    ${new_uid}=    Run With Kubeconfig
    ...    oc get secret metrics-server-tls -n ${METRICS_NS} -o jsonpath\\='{.metadata.uid}'
    Should Not Be Empty    ${new_uid}
    Should Not Be Equal    ${new_uid}    ${old_uid}
    Run With Kubeconfig
    ...    test "$(oc get secret metrics-server-tls -n ${METRICS_NS} -o jsonpath\\='{.data.tls\\.crt}' | base64 -d | wc -c)" -gt 0
    Run With Kubeconfig
    ...    test "$(oc get secret metrics-server-tls -n ${METRICS_NS} -o jsonpath\\='{.data.tls\\.key}' | base64 -d | wc -c)" -gt 0
    ${owner}=    Run With Kubeconfig
    ...    oc get secret metrics-server-tls -n ${METRICS_NS} -o jsonpath\\='{.metadata.annotations.service\\.beta\\.openshift\\.io/service-name}'
    Should Be Equal    ${owner}    metrics-server

Clear Metrics Server Serving Certificate Failure State
    [Documentation]    Leave the service-ca retry bookkeeping clear if the test fails.
    Run With Kubeconfig
    ...    oc annotate service metrics-server -n ${METRICS_NS} service.beta.openshift.io/serving-cert-generation-error- service.beta.openshift.io/serving-cert-generation-error-num-
    ...    allow_fail=${TRUE}
    Run With Kubeconfig
    ...    oc annotate service metrics-server -n ${METRICS_NS} service.alpha.openshift.io/serving-cert-generation-error- service.alpha.openshift.io/serving-cert-generation-error-num-
    ...    allow_fail=${TRUE}

Metrics Server API Should Be Available
    [Documentation]    Wait until the aggregated metrics API can route to metrics-server.
    Oc Wait    apiservice v1beta1.metrics.k8s.io    --for=condition=Available --timeout\\=120s

Metrics Endpoint Should Contain
    [Documentation]    Scrape kube-state-metrics on the given port and assert the
    ...    response contains the expected metric name.
    [Arguments]    ${port}    ${expected}
    ${metrics}=    Scrape Metrics From    kube-state-metrics    ${port}
    Should Contain    ${metrics}    ${expected}

Node Exporter Should Contain
    [Documentation]    Scrape node-exporter and assert the response contains the
    ...    expected metric name.
    [Arguments]    ${expected}
    ${metrics}=    Scrape Metrics From    node-exporter    9100
    Should Contain    ${metrics}    ${expected}

Scrape Metrics From
    [Documentation]    Curl the named metrics Service endpoint on the given port
    ...    with mTLS client certs. Resolves the IP from the Endpoints object
    ...    to validate that the headless Service selector matches the pod.
    [Arguments]    ${metrics_service}    ${port}
    # Oc Get JsonPath uses Run With Kubeconfig which merges stderr into stdout;
    # the v1 Endpoints deprecation warning from K8s 1.33+ pollutes the IP value.
    ${ep_ip}=    Command Should Work
    ...    oc get endpoints -n ${METRICS_NS} ${metrics_service} -o jsonpath\='{.subsets[0].addresses[0].ip}' --kubeconfig\=/var/lib/microshift/resources/kubeadmin/kubeconfig
    VAR    ${svc_host}=    ${metrics_service}.${METRICS_NS}.svc
    ${stdout}=    Command Should Work
    ...    curl -s --resolve ${svc_host}:${port}:${ep_ip} --cacert ${SERVICE_CA} --cert ${CLIENT_CERT} --key ${CLIENT_KEY} https://${svc_host}:${port}/metrics
    Should Not Be Empty    ${stdout}
    RETURN    ${stdout}
