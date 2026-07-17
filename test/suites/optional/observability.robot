*** Settings ***
Documentation       This suite performs basic log scans to determine if the opentelemetry-collector is exporting data
...                 and in a healthy state

Library             OperatingSystem
Library             SSHLibrary
Library             ../../resources/journalctl.py
Library             ../../resources/prometheus.py
Library             ../../resources/loki.py
Resource            ../../resources/kubeconfig.resource
Resource            ../../resources/common.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/optional-config.resource
Resource            ../../resources/systemd.resource
Resource            ../../resources/observability.resource
Resource            ../../resources/microshift-network.resource

Suite Setup         Setup Suite And Prepare Test Host
Suite Teardown      Teardown Suite And Revert Test Host


*** Variables ***
${JOURNAL_CUR}              ${EMPTY}
${DEFAULT_CONFIG_PATH}      /etc/microshift/observability/opentelemetry-collector-large.yaml
${OTEL_CONFIG_PATH}         /etc/microshift/observability/opentelemetry-collector.yaml
${TEST_CONFIG_PATH}         assets/observability/otel_config.yaml


*** Test Cases ***
Host Metrics Are Exported
    [Documentation]    The opentelemetry-collector should be able to export host metrics.

    VAR    ${METRIC}    system_cpu_time_seconds_total    scope=TEST
    Check Prometheus Query    ${PROMETHEUS_HOST}    ${PROMETHEUS_PORT}    ${METRIC}
    Check Prometheus Exporter    ${USHIFT_HOST}    ${PROM_EXPORTER_PORT}    ${METRIC}

Kube Metrics Are Exported
    [Documentation]    The opentelemetry-collector should be able to export kube metrics.

    VAR    ${METRIC}    container_cpu_time_seconds_total    scope=TEST
    Check Prometheus Query    ${PROMETHEUS_HOST}    ${PROMETHEUS_PORT}    ${METRIC}
    Check Prometheus Exporter    ${USHIFT_HOST}    ${PROM_EXPORTER_PORT}    ${METRIC}

    VAR    ${METRIC}    k8s_pod_cpu_time_seconds_total    scope=TEST
    Check Prometheus Query    ${PROMETHEUS_HOST}    ${PROMETHEUS_PORT}    ${METRIC}
    Check Prometheus Exporter    ${USHIFT_HOST}    ${PROM_EXPORTER_PORT}    ${METRIC}

Journald Logs Are Exported
    [Documentation]    The opentelemetry-collector should be able to export journald logs.

    Wait Until Keyword Succeeds    10x    5s
    ...    Check Loki Query    ${LOKI_HOST}    ${LOKI_PORT}    {service_name="journald"}

Kube Events Logs Are Exported
    [Documentation]    The opentelemetry-collector should be able to export Kubernetes events.

    Wait Until Keyword Succeeds    10x    5s
    ...    Check Loki Query    ${LOKI_HOST}    ${LOKI_PORT}    {service_name="kube_events"}

KSM Metrics Are Exported Via Scrape Drop-In
    [Documentation]    The prometheus receiver should scrape kube-state-metrics via the
    ...    scrape.d drop-in and export kube_node_info to the prometheus exporter.

    Wait Until Keyword Succeeds    60s    5s
    ...    Check Prometheus Exporter    ${USHIFT_HOST}    ${PROM_EXPORTER_PORT}    kube_node_info

Node Exporter Metrics Are Exported Via Scrape Drop-In
    [Documentation]    The prometheus receiver should scrape node-exporter via the
    ...    scrape.d drop-in and export node_cpu_seconds_total to the prometheus exporter.

    Wait Until Keyword Succeeds    60s    5s
    ...    Check Prometheus Exporter    ${USHIFT_HOST}    ${PROM_EXPORTER_PORT}    node_cpu_seconds_total

Metrics Server Metrics Are Exported Via Scrape Drop-In
    [Documentation]    The prometheus receiver should scrape metrics-server via the
    ...    scrape.d drop-in and export metrics to the prometheus exporter.

    Wait Until Keyword Succeeds    60s    5s
    ...    Check Prometheus Exporter    ${USHIFT_HOST}    ${PROM_EXPORTER_PORT}    metrics_server_kubelet_request_total

Logs Should Not Contain Receiver Errors
    [Documentation]    Internal receiver errors are not treated as fatal. Typically these are due to a misconfiguration
    ...    and thus indicate the provided default config should be reviewed.

    ${pattern}    Catenate    SEPARATOR=${EMPTY}
    ...    \\s+\\{"error":.*\\}
    Pattern Should Not Appear In Log Output    ${JOURNAL_CUR}    ${pattern}    unit="microshift-observability"


*** Keywords ***
Setup Suite And Prepare Test Host
    [Documentation]    The service starts after MicroShift starts and thus will start generating pertinent log data
    ...    right away. When the suite is executed, immediately get the cursor for the microshift-observability unit.
    Setup Suite
    Setup MicroShift With Optionals
    ...    003-microshift-observability
    ...    080-microshift-metrics-server
    ...    081-microshift-kube-state-metrics
    ...    082-microshift-node-exporter
    ${ns}    Create Unique Namespace
    VAR    ${NAMESPACE}    ${ns}    scope=SUITE
    Configure Firewall And Observability
    Create Hello MicroShift Pod
    Expose Hello MicroShift
    ${cur}    Get Journal Cursor    unit=microshift-observability
    VAR    ${JOURNAL_CUR}    ${cur}    scope=SUITE

Configure Firewall And Observability
    [Documentation]    Configure firewall for Prometheus exporter and set up observability
    Command Should Work    sudo firewall-cmd --permanent --zone=public --add-port=8889/tcp
    Command Should Work    sudo firewall-cmd --reload
    Check Required Observability Variables
    Ensure Loki Is Ready
    Set Test OTEL Configuration

Check Required Observability Variables
    [Documentation]    Check if the required proxy variables are set
    Should Not Be Empty    ${PROMETHEUS_HOST}    PROMETHEUS_HOST variable is required
    ${string_value}    Convert To String    ${PROMETHEUS_PORT}
    Should Not Be Empty    ${string_value}    PROMETHEUS_PORT variable is required
    Should Not Be Empty    ${LOKI_HOST}    LOKI_HOST variable is required
    ${string_value}    Convert To String    ${LOKI_HOST}
    Should Not Be Empty    ${string_value}    LOKI_HOST variable is required
    ${string_value}    Convert To String    ${LOKI_PORT}
    Should Not Be Empty    ${string_value}    LOKI_PORT variable is required

Ensure Loki Is Ready
    [Documentation]    Check if Loki's ingester is healthy, restart the container if not.
    ...    Loki's ingester can enter a shutdown state over time, causing it to
    ...    reject all writes with HTTP 503 while still responding to queries.
    ...    After a restart, the ingester needs ~15s before it reports ready.
    ${status}    ${error}    Run Keyword And Ignore Error
    ...    Check Loki Ready    ${LOKI_HOST}    ${LOKI_PORT}
    IF    "${status}" == "PASS"    RETURN
    Log    Loki is not ready: ${error}    console=True
    FOR    ${attempt}    IN RANGE    1    4
        Log    Restarting Loki (attempt ${attempt}/3)    console=True
        Local Command Should Work    ./bin/manage_loki.sh restart ${LOKI_PORT}
        ${poll_status}    ${poll_error}    Run Keyword And Ignore Error
        ...    Wait Until Keyword Succeeds    30s    5s    Check Loki Ready    ${LOKI_HOST}    ${LOKI_PORT}
        IF    "${poll_status}" == "PASS"    RETURN
    END
    Fail    Loki did not become ready after 3 restart attempts. Last error: ${poll_error}

Set Test OTEL Configuration
    [Documentation]    Set Test OTEL Configuration

    ${cfg}    Create Random Temp File
    SSHLibrary.Get File    ${OTEL_CONFIG_PATH}    ${cfg}
    ${config_str}    Local Command Should Work    cat ${TEST_CONFIG_PATH}
    Upload String To File    ${config_str}    ${OTEL_CONFIG_PATH}
    Render Otel Test Config
    Systemctl    restart    microshift-observability

Teardown Suite And Revert Test Host
    [Documentation]    Set back original OTEL config and teardown Suite
    Set Back Original OTEL Configuration
    Teardown MicroShift With Optionals
    Teardown Suite With Namespace

Set Back Original OTEL Configuration
    [Documentation]    Set Back Original OTEL Configuration

    ${def_config_str}    Command Should Work    cp ${DEFAULT_CONFIG_PATH} ${OTEL_CONFIG_PATH}
    Systemctl    restart    microshift-observability

Render Otel Test Config
    [Documentation]    Set IPs and ports in OTEL config yaml

    ${ushift_host_formatted}    Add Brackets If Ipv6    ${USHIFT_HOST}
    ${prometheus_host_formatted}    Add Brackets If Ipv6    ${PROMETHEUS_HOST}
    ${loki_host_formatted}    Add Brackets If Ipv6    ${LOKI_HOST}

    Command Should Work    sed -i "s|{{NODE_IP}}|${ushift_host_formatted}|g" ${OTEL_CONFIG_PATH}
    Command Should Work    sed -i "s|{{PROMETHEUS_HOST}}|${prometheus_host_formatted}|g" ${OTEL_CONFIG_PATH}
    Command Should Work    sed -i "s|{{PROMETHEUS_PORT}}|${PROMETHEUS_PORT}|g" ${OTEL_CONFIG_PATH}
    Command Should Work    sed -i "s|{{LOKI_HOST}}|${loki_host_formatted}|g" ${OTEL_CONFIG_PATH}
    Command Should Work    sed -i "s|{{LOKI_PORT}}|${LOKI_PORT}|g" ${OTEL_CONFIG_PATH}
    Command Should Work    sed -i "s|{{PROM_EXPORTER_PORT}}|${PROM_EXPORTER_PORT}|g" ${OTEL_CONFIG_PATH}
