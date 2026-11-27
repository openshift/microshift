*** Settings ***
Documentation       This suite performs basic log scans to determine if the opentelemetry-collector is exporting data
...                 and in a healthy state

Library             OperatingSystem
Library             SSHLibrary
Library             ../../resources/journalctl.py
Library             ../../resources/prometheus.py
Library             ../../resources/loki.py
Library             ../../resources/DataFormats.py
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
${CERTIFICATE_LOCK}         /var/lib/microshift-backups/certs.lock
${CERT_LOCK_UNIT}           microshift-cert-metrics-lock
# A one-year standard certificate enters its rotation window after about 153 days.
${FUTURE_DAYS}              160


*** Test Cases ***
Host Metrics Are Exported
    [Documentation]    The opentelemetry-collector should be able to export host metrics.

    VAR    ${METRIC}    system_cpu_time_seconds_total    scope=TEST
    Wait Until Keyword Succeeds    60s    5s
    ...    Check Prometheus Query    ${PROMETHEUS_HOST}    ${PROMETHEUS_PORT}    ${METRIC}
    Wait Until Keyword Succeeds    60s    5s
    ...    Check Prometheus Exporter    ${USHIFT_HOST}    ${PROM_EXPORTER_PORT}    ${METRIC}

Kube Metrics Are Exported
    [Documentation]    The opentelemetry-collector should be able to export kube metrics.

    VAR    ${METRIC}    container_cpu_time_seconds_total    scope=TEST
    Wait Until Keyword Succeeds    60s    5s
    ...    Check Prometheus Query    ${PROMETHEUS_HOST}    ${PROMETHEUS_PORT}    ${METRIC}
    Wait Until Keyword Succeeds    60s    5s
    ...    Check Prometheus Exporter    ${USHIFT_HOST}    ${PROM_EXPORTER_PORT}    ${METRIC}

    VAR    ${METRIC}    k8s_pod_cpu_time_seconds_total    scope=TEST
    Wait Until Keyword Succeeds    60s    5s
    ...    Check Prometheus Query    ${PROMETHEUS_HOST}    ${PROMETHEUS_PORT}    ${METRIC}
    Wait Until Keyword Succeeds    60s    5s
    ...    Check Prometheus Exporter    ${USHIFT_HOST}    ${PROM_EXPORTER_PORT}    ${METRIC}

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

    VAR    ${METRIC}    kube_node_info    scope=TEST
    Wait Until Keyword Succeeds    60s    5s
    ...    Check Prometheus Query    ${PROMETHEUS_HOST}    ${PROMETHEUS_PORT}    ${METRIC}
    Wait Until Keyword Succeeds    60s    5s
    ...    Check Prometheus Exporter    ${USHIFT_HOST}    ${PROM_EXPORTER_PORT}    ${METRIC}

Node Exporter Metrics Are Exported Via Scrape Drop-In
    [Documentation]    The prometheus receiver should scrape node-exporter via the
    ...    scrape.d drop-in and export node_cpu_seconds_total to the prometheus exporter.

    VAR    ${METRIC}    node_cpu_seconds_total    scope=TEST
    Wait Until Keyword Succeeds    60s    5s
    ...    Check Prometheus Query    ${PROMETHEUS_HOST}    ${PROMETHEUS_PORT}    ${METRIC}
    Wait Until Keyword Succeeds    60s    5s
    ...    Check Prometheus Exporter    ${USHIFT_HOST}    ${PROM_EXPORTER_PORT}    ${METRIC}

Metrics Server Metrics Are Exported Via Scrape Drop-In
    [Documentation]    The prometheus receiver should scrape metrics-server via the
    ...    scrape.d drop-in and export metrics to the prometheus exporter.

    VAR    ${METRIC}    metrics_server_kubelet_request_total    scope=TEST
    Wait Until Keyword Succeeds    60s    5s
    ...    Check Prometheus Query    ${PROMETHEUS_HOST}    ${PROMETHEUS_PORT}    ${METRIC}
    Wait Until Keyword Succeeds    60s    5s
    ...    Check Prometheus Exporter    ${USHIFT_HOST}    ${PROM_EXPORTER_PORT}    ${METRIC}

Certificate Metrics Follow Renewal Preparation And Activation
    [Documentation]    A prepared CA renewal sets microshift_certificate_renewal_pending to 1
    ...    without changing the active certificates' metrics, and the value reaches Prometheus
    ...    through the OpenTelemetry pipeline. While the operation lock is held the collection
    ...    reports failure and drops its series instead of serving a stale observation. One
    ...    activating restart clears the flag, every renewed certificate reports the reissued
    ...    validity, and the metrics match the CLI. CA renewal reissues every certificate, so it
    ...    covers serving renewal too. OCPEDGE-3226.
    [Tags]    certificates    restart
    ${before}    Wait For Certificate Metrics    0
    ${stdout}    Command Should Work    microshift certs renew --ca -o json
    ${result}    Json Parse    ${stdout}
    ${pending}    Wait For Certificate Metrics    1
    Should Be Equal    ${before}[certs]    ${pending}[certs]
    Backend Should Report Renewal Pending
    Stale Collection Is Reported During Lock Contention
    # Stop the collector across the restart so it never logs transient scrape errors
    # against the briefly-unavailable endpoint; the teardown restarts it, which also
    # picks up its reissued client certificate.
    Systemctl    stop    microshift-observability
    Restart MicroShift
    ${active}    Wait For Certificate Metrics    0
    Renewed Metrics Match Renewal Result    ${result}    ${before}    ${active}
    Metrics Match Certificate Status    ${active}
    [Teardown]    Run Keywords
    ...    Execute Command    systemctl stop ${CERT_LOCK_UNIT}    sudo=True
    ...    AND    Run Keyword If Test Failed    Restart MicroShift
    ...    AND    Systemctl    restart    microshift-observability

Certificate Metrics Follow Automatic Rotation
    [Documentation]    When the clock moves past the standard certificates' rotation window,
    ...    MicroShift regenerates them on the automatic restart and every standard-policy
    ...    certificate reports a later not_after, with metrics matching the CLI. The collector
    ...    is stopped across the clock jump, and the future-dated Kubernetes events are deleted
    ...    before it starts again, so nothing future-dated reaches the shared Prometheus and Loki
    ...    backends. OCPEDGE-3226.
    [Tags]    certificates    restart
    [Setup]    Systemctl    stop    microshift-observability
    ${before}    Wait For Certificate Metrics    0
    Move Clock Past Standard Rotation
    ${after}    Wait For Certificate Metrics    0
    Standard Certificates Should Have Rotated    ${before}    ${after}
    Metrics Match Certificate Status    ${after}
    [Teardown]    Run Keywords    Restore System Date
    ...    AND    Run With Kubeconfig    oc delete events --all --all-namespaces
    ...    AND    Systemctl    start    microshift-observability

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

Read Local Certificate Metrics
    [Documentation]    Scrape MicroShift's loopback metrics endpoint on the host and parse it.
    ${output}    Command Should Work
    ...    curl --noproxy '*' --fail --silent --show-error --max-time 5 ${LOCAL_METRICS_URL}
    ${metrics}    Parse Certificate Metrics    ${output}
    RETURN    ${metrics}

Wait For Certificate Metrics
    [Documentation]    Wait for a successful collection reporting the given renewal_pending value
    ...    and return it. A changed pending value, or a new MicroShift process (which starts
    ...    with no previous collection), proves the snapshot is fresh.
    [Arguments]    ${pending}
    ${metrics}    Wait Until Keyword Succeeds    150s    5s
    ...    Certificate Metrics Should Report    ${pending}
    RETURN    ${metrics}

Certificate Metrics Should Report
    [Documentation]    Read the endpoint and require a successful collection with the given pending value.
    [Arguments]    ${pending}
    ${metrics}    Read Local Certificate Metrics
    Should Be Equal As Numbers    ${metrics}[success]    1
    Should Be Equal As Numbers    ${metrics}[pending]    ${pending}
    RETURN    ${metrics}

Renewed Metrics Match Renewal Result
    [Documentation]    Every renewed certificate reports the newNotAfter the renewal printed and a
    ...    later not_before than before. not_before is the activation proof: same-day renewals
    ...    keep the same midnight-aligned not_after, but they are always issued later.
    [Arguments]    ${result}    ${before}    ${active}
    FOR    ${item}    IN    @{result}[items]
        VAR    ${key}    ${item}[service]/${item}[name]
        ${expected}    Evaluate
        ...    int(datetime.datetime.fromisoformat($item['newNotAfter'].replace('Z', '+00:00')).timestamp())
        ...    modules=datetime
        Should Be Equal As Numbers    ${active}[certs][${key}][not_after]    ${expected}
        Should Be True    ${active}[certs][${key}][not_before] > ${before}[certs][${key}][not_before]
        ...    msg=${key} was not reissued
    END

Metrics Match Certificate Status
    [Documentation]    The certificate set and each certificate's status match `microshift certs status`.
    [Arguments]    ${metrics}
    ${stdout}    Command Should Work    microshift certs status -o json
    ${status}    Json Parse    ${stdout}
    ${expected}    Evaluate    {i['service'] + '/' + i['name']: i['status'] for i in $status['items']}
    ${actual}    Evaluate    {k: v['status'] for k, v in $metrics['certs'].items()}
    Should Be Equal    ${actual}    ${expected}

Standard Certificates Should Have Rotated
    [Documentation]    Every standard-policy certificate (the one-year leaves that cross their
    ...    rotation window after the clock jump) reports a later not_after than before.
    [Arguments]    ${before}    ${after}
    ${keys}    Evaluate    [k for k, v in $after['certs'].items() if v['rotation_policy'] == 'standard']
    Should Not Be Empty    ${keys}
    FOR    ${key}    IN    @{keys}
        Should Be True    ${after}[certs][${key}][not_after] > ${before}[certs][${key}][not_after]
        ...    msg=${key} did not rotate
    END

Backend Should Report Renewal Pending
    [Documentation]    renewal_pending=1 reaches Prometheus for this host through the OpenTelemetry
    ...    pipeline. The wait covers the collector refresh, the scrape and the remote write.
    ${hostname}    Command Should Work    hostname
    VAR    ${query}    microshift_certificate_renewal_pending{host_name="${hostname.strip()}"} == 1
    Wait Until Keyword Succeeds    180s    10s
    ...    Check Prometheus Query    ${PROMETHEUS_HOST}    ${PROMETHEUS_PORT}    ${query}    ${False}

Stale Collection Is Reported During Lock Contention
    [Documentation]    While the certificate operation lock is held exclusively, the collection
    ...    reports failure and drops the certificate and pending series without advancing
    ...    last-success. Releasing the lock restores a successful collection with pending intact.
    # Hold the lock until stopped explicitly, so it cannot expire mid-check.
    Command Should Work
    ...    systemd-run --quiet --collect --unit=${CERT_LOCK_UNIT} flock -x ${CERTIFICATE_LOCK} sleep infinity
    Wait Until Keyword Succeeds    10x    1s    Command Should Fail    flock -n ${CERTIFICATE_LOCK} true
    # Read the epoch only once the lock is held: any later collection must fail,
    # so a successful one can only predate this moment.
    ${lock_epoch}    Command Should Work    date +%s
    Wait Until Keyword Succeeds    90s    5s    Collection Reports Failure    ${lock_epoch}
    Command Should Work    systemctl stop ${CERT_LOCK_UNIT}
    Wait For Certificate Metrics    1

Collection Reports Failure
    [Documentation]    A failed observation sets collection_success to 0, drops the certificate and
    ...    renewal_pending series, and keeps the previous last-success timestamp.
    [Arguments]    ${lock_epoch}
    ${metrics}    Read Local Certificate Metrics
    Should Be Equal As Numbers    ${metrics}[success]    0
    Should Be Equal    ${metrics}[pending]    ${None}
    Should Be Empty    ${metrics}[certs]
    Should Be True    0 < ${metrics}[last_success] <= ${lock_epoch.strip()}

Move Clock Past Standard Rotation
    [Documentation]    Stop time sync and move the clock ${FUTURE_DAYS} days ahead, past the
    ...    rotation window of the one-year standard certificates, then wait for MicroShift to
    ...    restart itself and regenerate them.
    ${ushift_pid}    MicroShift Process ID
    Systemctl    stop    chronyd
    ${future_date}    Command Should Work    TZ=UTC date "+%Y-%m-%d %H:%M:%S" -d "+${FUTURE_DAYS} day"
    Command Should Work    TZ=UTC timedatectl set-time "${future_date}"
    Wait Until MicroShift Process ID Changes    ${ushift_pid}
    Setup Kubeconfig
    Wait For MicroShift Healthcheck Success

Restore System Date
    [Documentation]    Restart time sync so the clock returns to real time and wait for
    ...    MicroShift to restart onto it. Skips if chronyd is already active.
    ${stdout}    ${rc}    Execute Command
    ...    systemctl is-active chronyd    sudo=True    return_rc=True
    IF    '${stdout.strip()}' == 'active'    RETURN
    ${ushift_pid}    MicroShift Process ID
    Systemctl    start    chronyd
    Wait Until MicroShift Process ID Changes    ${ushift_pid}
    Setup Kubeconfig
    Wait For MicroShift Healthcheck Success
