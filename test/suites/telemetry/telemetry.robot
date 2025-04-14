*** Settings ***
Documentation       Tests for Telemetry

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-host.resource
Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-process.resource
Library             ../../resources/journalctl.py
Library             ../../resources/prometheus.py
Library             ../../resources/ProxyLibrary.py

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           restart    slow


*** Variables ***
${PROXY_HOST}                   ${EMPTY}
${PROXY_PORT}                   ${EMPTY}
${PROMETHEUS_HOST}              ${EMPTY}
${PROMETHEUS_PORT}              ${EMPTY}
${TELEMETRY_WRITE_ENDPOINT}     https://infogw.api.openshift.com/metrics/v1/receive
${ENABLE_TELEMETRY}             SEPARATOR=\n
...                             telemetry:
...                             \ \ status: Enabled
${JOURNAL_CURSOR}               ${EMPTY}
${PULL_SECRET}                  /etc/crio/openshift-pull-secret
${PULL_SECRET_METRICS}          /etc/crio/openshift-pull-secret-with-telemetry
${PULL_SECRET_NO_METRICS}       /etc/crio/openshift-pull-secret-without-telemetry


*** Test Cases ***
MicroShift Reports Metrics To Server
    [Documentation]    Check MicroShift is able to send metrics to the telemetry server without errors.
    [Setup]    Setup Telemetry Configuration

    Wait Until Keyword Succeeds    10x    10s
    ...    Should Find Metrics Success

    [Teardown]    Remove Telemetry Configuration

MicroShift Reports Metrics To Server Through Proxy
    [Documentation]    Check MicroShift is able to send metrics to the telemetry server through a proxy without errors.
    [Setup]    Setup Telemetry Configuration With Proxy

    Wait Until Keyword Succeeds    10x    10s
    ...    Should Find Metrics Success

    [Teardown]    Remove Telemetry Configuration

Check MicroShift Metrics In Local Server    # robocop: disable=too-long-test-case
    [Documentation]    Check the expected metrics are sent to the local server.
    [Setup]    Setup Local Telemetry Configuration

    ${system_arch}=    Get System Architecture
    ${deployment_type}=    Get Deployment Type
    ${arch}=    Set Variable If    "${system_arch}" == "x86_64"    amd64    arm64
    ${cluster_id}=    Get MicroShift Cluster ID From File
    ${os_version}=    Get Host OS Version
    ${microshift_ver}=    MicroShift Version
    ${microshift_version}=    Set Variable    ${microshift_ver.major}.${microshift_ver.minor}.${microshift_ver.patch}

    # Ensure metrics have been sent before checking in Prometheus.
    Wait Until Keyword Succeeds    10x    10s
    ...    Should Find Metrics Success

    Check Prometheus Query
    ...    ${PROMETHEUS_HOST}
    ...    ${PROMETHEUS_PORT}
    ...    cluster:capacity_cpu_cores:sum{_id="${cluster_id}",label_beta_kubernetes_io_instance_type="rhde",label_node_openshift_io_os_id="rhel",label_kubernetes_io_arch="${arch}"}
    Check Prometheus Query
    ...    ${PROMETHEUS_HOST}
    ...    ${PROMETHEUS_PORT}
    ...    cluster:capacity_memory_bytes:sum{_id="${cluster_id}",label_beta_kubernetes_io_instance_type="rhde",label_node_openshift_io_os_id="rhel",label_kubernetes_io_arch="${arch}"}
    Check Prometheus Query
    ...    ${PROMETHEUS_HOST}
    ...    ${PROMETHEUS_PORT}
    ...    cluster:cpu_usage_cores:sum{_id="${cluster_id}"}
    Check Prometheus Query
    ...    ${PROMETHEUS_HOST}
    ...    ${PROMETHEUS_PORT}
    ...    cluster:memory_usage_bytes:sum{_id="${cluster_id}"}
    Check Prometheus Query
    ...    ${PROMETHEUS_HOST}
    ...    ${PROMETHEUS_PORT}
    ...    cluster:usage:resources:sum{_id="${cluster_id}",resource="pods"}
    Check Prometheus Query
    ...    ${PROMETHEUS_HOST}
    ...    ${PROMETHEUS_PORT}
    ...    cluster:usage:resources:sum{_id="${cluster_id}",resource="namespaces"}
    Check Prometheus Query
    ...    ${PROMETHEUS_HOST}
    ...    ${PROMETHEUS_PORT}
    ...    cluster:usage:resources:sum{_id="${cluster_id}",resource="services"}
    Check Prometheus Query
    ...    ${PROMETHEUS_HOST}
    ...    ${PROMETHEUS_PORT}
    ...    cluster:usage:resources:sum{_id="${cluster_id}",resource="ingresses.networking.k8s.io"}
    Check Prometheus Query
    ...    ${PROMETHEUS_HOST}
    ...    ${PROMETHEUS_PORT}
    ...    cluster:usage:resources:sum{_id="${cluster_id}",resource="routes.route.openshift.io"}
    Check Prometheus Query
    ...    ${PROMETHEUS_HOST}
    ...    ${PROMETHEUS_PORT}
    ...    cluster:usage:resources:sum{_id="${cluster_id}",resource="customresourcedefinitions.apiextensions.k8s.io"}
    Check Prometheus Query
    ...    ${PROMETHEUS_HOST}
    ...    ${PROMETHEUS_PORT}
    ...    cluster:usage:containers:sum{_id="${cluster_id}"}
    Check Prometheus Query
    ...    ${PROMETHEUS_HOST}
    ...    ${PROMETHEUS_PORT}
    ...    microshift_version{_id="${cluster_id}",deployment_type="${deployment_type}",os_version_id="${os_version}",version="${microshift_version}"}

    [Teardown]    Remove Telemetry Configuration


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Check Required Telemetry Variables
    Login MicroShift Host
    Configure Pull Secrets
    Setup Kubeconfig

Teardown
    [Documentation]    Test suite teardown
    Restore Pull Secrets
    Restart MicroShift
    Logout MicroShift Host
    Remove Kubeconfig

Check Required Telemetry Variables
    [Documentation]    Check if the required proxy variables are set
    Should Not Be Empty    ${PROXY_HOST}    PROXY_HOST variable is required
    ${string_value}=    Convert To String    ${PROXY_PORT}
    Should Not Be Empty    ${string_value}    PROXY_PORT variable is required
    Should Not Be Empty    ${PROMETHEUS_HOST}    PROMETHEUS_HOST variable is required
    ${string_value}=    Convert To String    ${PROMETHEUS_PORT}
    Should Not Be Empty    ${string_value}    PROMETHEUS_PORT variable is required

Setup Telemetry Configuration
    [Documentation]    Enables the telemetry feature in MicroShift configuration file
    ...    and restarts microshift.service
    ${config}=    Catenate
    ...    SEPARATOR=\n
    ...    ${ENABLE_TELEMETRY}
    ...    \ \ endpoint: ${TELEMETRY_WRITE_ENDPOINT}
    Drop In MicroShift Config    ${config}    10-telemetry
    Stop MicroShift
    ${cursor}=    Get Journal Cursor
    Set Suite Variable    \${CURSOR}    ${cursor}
    Restart MicroShift

Setup Telemetry Configuration With Proxy
    [Documentation]    Enables the telemetry feature in MicroShift configuration file
    ...    and restarts microshift.service
    Start Proxy Server    host=${PROXY_HOST}    port=${PROXY_PORT}
    ${proxy_config}=    Catenate
    ...    SEPARATOR=\n
    ...    ${ENABLE_TELEMETRY}
    ...    \ \ endpoint: ${TELEMETRY_WRITE_ENDPOINT}
    ...    \ \ proxy: http://${PROXY_HOST}:${PROXY_PORT}
    Drop In MicroShift Config    ${proxy_config}    10-telemetry
    Stop MicroShift
    ${cursor}=    Get Journal Cursor
    Set Suite Variable    \${CURSOR}    ${cursor}
    Restart MicroShift

Setup Local Telemetry Configuration
    [Documentation]    Enables the telemetry feature in MicroShift configuration file
    ...    and restarts microshift.service
    ${config}=    Catenate
    ...    SEPARATOR=\n
    ...    ${ENABLE_TELEMETRY}
    ...    \ \ endpoint: http://${PROMETHEUS_HOST}:${PROMETHEUS_PORT}/api/v1/write
    Drop In MicroShift Config    ${config}    10-telemetry
    Stop MicroShift
    ${cursor}=    Get Journal Cursor
    Set Suite Variable    \${CURSOR}    ${cursor}
    Restart MicroShift

Remove Telemetry Configuration
    [Documentation]    Removes the telemetry feature from MicroShift configuration file
    ...    and restarts microshift.service
    Remove Drop In MicroShift Config    10-telemetry
    Restart MicroShift

Configure Pull Secrets
    [Documentation]    Sets up the pull secrets for the MicroShift cluster.
    ${rc}=    SSHLibrary.Execute Command
    ...    grep -q cloud.openshift.com ${PULL_SECRET} || sudo ln -sf ${PULL_SECRET_METRICS} ${PULL_SECRET}
    ...    sudo=True
    ...    return_rc=True
    ...    return_stderr=False
    ...    return_stdout=False
    Should Be Equal As Integers    ${rc}    0

Restore Pull Secrets
    [Documentation]    Restores the original pull secrets for the MicroShift cluster if it was changed by the test.
    ${rc}=    SSHLibrary.Execute Command
    ...    test -f ${PULL_SECRET_NO_METRICS} && sudo ln -sf ${PULL_SECRET_NO_METRICS} ${PULL_SECRET} || true
    ...    sudo=True    return_rc=True    return_stderr=False    return_stdout=False
    Should Be Equal As Integers    ${rc}    0

Should Find Metrics Success
    [Documentation]    Logs should contain metrics success message
    Pattern Should Appear In Log Output    ${CURSOR}    Metrics sent successfully
