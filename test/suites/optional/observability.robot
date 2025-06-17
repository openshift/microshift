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
Resource            ../../resources/systemd.resource
Resource            ../../resources/observability.resource

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

    ${pattern}    Catenate    SEPARATOR=
    ...    info\\s+MetricsExporter\\s+{"kind": "exporter", "data_type": "metrics", "name": "debug", "resource metrics":
    ...    \ [0-9]+, "metrics": [0-9]+, "data points": [0-9]+}
    Pattern Should Appear In Log Output    ${JOURNAL_CUR}    ${pattern}    unit="microshift-observability"

    Set Test Variable    ${METRIC}    system_cpu_time_seconds_total{cpu="cpu0",state="idle"}
    Check Prometheus Query    ${PROMETHEUS_HOST}    ${PROMETHEUS_PORT}    ${METRIC}
    Check Prometheus Exporter    ${USHIFT_HOST}    ${PROM_EXPORTER_PORT}    ${METRIC}

Kube Metrics Are Exported
    [Documentation]    The opentelemetry-collector should be able to export kube metrics.

    ${pattern}    Catenate    SEPARATOR=
    ...    info\\s+MetricsExporter\\s+{"kind": "exporter", "data_type": "metrics", "name": "debug", "resource metrics":
    ...    \ [0-9]+, "metrics": [0-9]+, "data points": [0-9]+}
    Pattern Should Appear In Log Output    ${JOURNAL_CUR}    ${pattern}    unit="microshift-observability"

    Set Test Variable    ${METRIC}    container_cpu_time_seconds_total
    Check Prometheus Query    ${PROMETHEUS_HOST}    ${PROMETHEUS_PORT}    ${METRIC}
    Check Prometheus Exporter    ${USHIFT_HOST}    ${PROM_EXPORTER_PORT}    ${METRIC}

    Set Test Variable    ${METRIC}    k8s_pod_cpu_time_seconds_total
    Check Prometheus Query    ${PROMETHEUS_HOST}    ${PROMETHEUS_PORT}    ${METRIC}
    Check Prometheus Exporter    ${USHIFT_HOST}    ${PROM_EXPORTER_PORT}    ${METRIC}

Journald Logs Are Exported
    [Documentation]    The opentelemetry-collector should be able to export logs to journald.

    ${pattern}    Catenate
    ...    SEPARATOR=
    ...    info\\s+LogsExporter\\s+\\{"kind": "exporter", "data_type": "logs", "name": "debug", "resource logs": [0-9]+,
    ...    \ "log records": [0-9]+\\}
    Pattern Should Appear In Log Output    ${JOURNAL_CUR}    ${pattern}    unit="microshift-observability"

    Check Loki Query    ${LOKI_HOST}    ${LOKI_PORT}    {job="journald",exporter="OTLP"}

Kube Events Logs Are Exported
    [Documentation]    The opentelemetry-collector should be able to export logs to journald.

    ${pattern}    Catenate
    ...    SEPARATOR=
    ...    info\\s+LogsExporter\\s+\\{"kind": "exporter", "data_type": "logs", "name": "debug", "resource logs": [0-9]+,
    ...    \ "log records": [0-9]+\\}
    Pattern Should Appear In Log Output    ${JOURNAL_CUR}    ${pattern}    unit="microshift-observability"

    Check Loki Query    ${LOKI_HOST}    ${LOKI_PORT}    {job="kube_events",exporter="OTLP"}

Logs Should Not Contain Receiver Errors
    [Documentation]    Internal receiver errors are not treated as fatal. Typically these are due to a misconfiguration
    ...    and thus indicate the provided default config should be reviewed.

    ${pattern}    Catenate    SEPARATOR=
    ...    \\s+\\{"error":.*\\}
    Pattern Should Not Appear In Log Output    ${JOURNAL_CUR}    ${pattern}    unit="microshift-observability"


*** Keywords ***
Setup Suite And Prepare Test Host
    [Documentation]    The service starts after MicroShift starts and thus will start generating pertinent log data
    ...    right away. When the suite is executed, immediately get the cursor for the microshift-observability unit.
    Start Prometheus Server    ${PROMETHEUS_PORT}
    Start Loki Server    ${LOKI_PORT}
    Setup Suite
    Check Required Observability Variables
    Set Test OTEL Configuration
    ${cur}    Get Journal Cursor    unit=microshift-observability
    Set Suite Variable    ${JOURNAL_CUR}    ${cur}
    Wait Until Keyword Succeeds    5 min    10 sec
    ...    Journal Contains Enough Lines To Test

Check Required Observability Variables
    [Documentation]    Check if the required proxy variables are set
    Should Not Be Empty    ${PROMETHEUS_HOST}    PROMETHEUS_HOST variable is required
    ${string_value}    Convert To String    ${PROMETHEUS_PORT}
    Should Not Be Empty    ${string_value}    PROMETHEUS_PORT variable is required
    Should Not Be Empty    ${LOKI_HOST}    LOKI_HOST variable is required
    ${string_value}    Convert To String    ${LOKI_PORT}
    Should Not Be Empty    ${string_value}    LOKI_PORT variable is required
    ${string_value}    Convert To String    ${PROM_EXPORTER_PORT}
    Should Not Be Empty    ${string_value}    PROM_EXPORTER_PORT variable is required

Journal Contains Enough Lines To Test
    [Documentation]    Execution should wait until there are at least 20 lines of journal data to process. This is
    ...    necessary because opentelemetry-collector will write debug output in batches. Thus, it can often happen
    ...    that there is not enough log data to gain an accurate read of the process's state, resulting in a false
    ...    negative signal (the opentelemetry-collector is healthy, but did not yet write data to journal.

    ${output}    ${rc}    Get Log Output With Pattern    ${JOURNAL_CUR}    .*    microshift-observability
    ${line_cnt}    Get Line Count    ${output}
    Should Be True    ${line_cnt} > 20

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
    Teardown Suite
    Stop Loki Server    ${LOKI_PORT}
    Stop Prometheus Server    ${PROMETHEUS_PORT}

Set Back Original OTEL Configuration
    [Documentation]    Set Back Original OTEL Configuration

    ${def_config_str}    Command Should Work    cp ${DEFAULT_CONFIG_PATH} ${OTEL_CONFIG_PATH}
    Systemctl    restart    microshift-observability

Render Otel Test Config
    [Documentation]    Set IPs and ports in OTEL config yaml

    Command Should Work    sed -i "s|{{NODE_IP}}|${USHIFT_HOST}|g" ${OTEL_CONFIG_PATH}
    Command Should Work    sed -i "s|{{PROMETHEUS_HOST}}|${PROMETHEUS_HOST}|g" ${OTEL_CONFIG_PATH}
    Command Should Work    sed -i "s|{{PROMETHEUS_PORT}}|${PROMETHEUS_PORT}|g" ${OTEL_CONFIG_PATH}
    Command Should Work    sed -i "s|{{LOKI_HOST}}|${LOKI_HOST}|g" ${OTEL_CONFIG_PATH}
    Command Should Work    sed -i "s|{{LOKI_PORT}}|${LOKI_PORT}|g" ${OTEL_CONFIG_PATH}
    Command Should Work    sed -i "s|{{PROM_EXPORTER_PORT}}|${PROM_EXPORTER_PORT}|g" ${OTEL_CONFIG_PATH}
