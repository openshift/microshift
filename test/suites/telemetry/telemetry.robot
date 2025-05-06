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
${PROXY_HOST}                       ${EMPTY}
${PROXY_PORT}                       ${EMPTY}
${PROMETHEUS_HOST}                  ${EMPTY}
${PROMETHEUS_PORT}                  ${EMPTY}
${TELEMETRY_WRITE_ENDPOINT}         https://infogw.api.openshift.com/metrics/v1/receive
${PROXY_ENDPOINT}                   http://${PROXY_HOST}:${PROXY_PORT}/api/v1/write
${ENABLE_TELEMETRY}                 SEPARATOR=\n
...                                 telemetry:
...                                 \ \ status: Enabled
...                                 \ \ endpoint: ${TELEMETRY_WRITE_ENDPOINT}
${DISABLE_TELEMETRY}                SEPARATOR=\n
...                                 telemetry:
...                                 \ \ status: Disabled
...                                 \ \ endpoint: ${TELEMETRY_WRITE_ENDPOINT}
${ENABLE_TELEMETRY_WITH_PROXY}      SEPARATOR=\n
...                                 telemetry:
...                                 \ \ status: Enabled
...                                 \ \ endpoint: ${TELEMETRY_WRITE_ENDPOINT}
...                                 \ \ proxy: ${PROXY_ENDPOINT}
${ENABLE_TELEMETRY_TO_PROMETHUS}    SEPARATOR=\n
...                                 telemetry:
...                                 \ \ status: Enabled
...                                 \ \ endpoint: http://${PROMETHEUS_HOST}:${PROMETHEUS_PORT}/api/v1/write
${JOURNAL_CURSOR}                   ${EMPTY}
${PULL_SECRET}                      /etc/crio/openshift-pull-secret
${PULL_SECRET_METRICS}              /etc/crio/openshift-pull-secret-with-telemetry
${PULL_SECRET_NO_METRICS}           /etc/crio/openshift-pull-secret-without-telemetry


*** Test Cases ***
MicroShift Reports Metrics To Server
    [Documentation]    Check MicroShift is able to send metrics to the telemetry server without errors.
    [Setup]    Setup Telemetry Configuration    ${ENABLE_TELEMETRY}    ${PULL_SECRET_METRICS}

    Should Find Metrics In Journal Log Success    MicroShift telemetry starting, sending first metrics collection.
    Should Find Metrics In Journal Log Success    Metrics sent successfully

    [Teardown]    Remove Telemetry Configuration

MicroShift Reports Metrics To Server Through Proxy
    [Documentation]    Check MicroShift is able to send metrics to the telemetry server through a proxy without errors.
    [Setup]    Setup Telemetry Configuration With Proxy    ${ENABLE_TELEMETRY_WITH_PROXY}    ${PULL_SECRET_METRICS}

    Should Find Metrics In Journal Log Success    MicroShift telemetry starting, sending first metrics collection.
    Should Find Metrics In Journal Log Success    Metrics sent successfully

    [Teardown]    Remove Telemetry Configuration With Proxy

MicroShift Fails to Report Metrics To Server Telemetry Disabled
    [Documentation]    Check MicroShift is not able to send metrics to the telemetry server when it is disabled.
    [Setup]    Setup Telemetry Configuration    ${DISABLE_TELEMETRY}    ${PULL_SECRET_METRICS}

    Should Find Metrics In Journal Log Success    Telemetry is disabled
    Should Find Metrics In Journal Log Fails    Metrics sent successfully

    [Teardown]    Remove Telemetry Configuration

MicroShift Fails to Report Metrics To Server Wrong Pull Secret
    [Documentation]    Check MicroShift is not able to send metrics to the telemetry server when the pull secret is wrong.
    [Setup]    Setup Telemetry Configuration    ${ENABLE_TELEMETRY}    ${PULL_SECRET_NO_METRICS}

    Should Find Metrics In Journal Log Success    MicroShift telemetry starting, sending first metrics collection.
    Should Find Metrics In Journal Log Success    Unable to get pull secret: cloud.openshift.com not found
    Should Find Metrics In Journal Log Fails    Metrics sent successfully

    [Teardown]    Remove Telemetry Configuration

Check MicroShift Metrics Sent To Prometheus Server    # robocop: disable=too-many-calls-in-test-case
    [Documentation]    Check the expected metrics are sent to the local server.
    Setup Telemetry Configuration    ${ENABLE_TELEMETRY_TO_PROMETHUS}    ${PULL_SECRET_METRICS}

    ${system_arch}=    Get System Architecture
    ${deployment_type}=    Get Deployment Type
    ${arch}=    Set Variable If    "${system_arch}" == "x86_64"    amd64    arm64
    ${cluster_id}=    Get MicroShift Cluster ID From File
    ${os_id}=    Get Host OS Id
    ${os_version}=    Get Host OS Version
    ${microshift_ver}=    MicroShift Version
    ${microshift_version}=    Set Variable    ${microshift_ver.major}.${microshift_ver.minor}.${microshift_ver.patch}

    Should Find Metrics In Prometheus Server
    ...    cluster:capacity_cpu_cores:sum{_id="${cluster_id}",label_beta_kubernetes_io_instance_type="rhde",label_node_openshift_io_os_id="${os_id}",label_kubernetes_io_arch="${arch}"}
    Should Find Metrics In Prometheus Server
    ...    cluster:capacity_memory_bytes:sum{_id="${cluster_id}",label_beta_kubernetes_io_instance_type="rhde",label_node_openshift_io_os_id="${os_id}",label_kubernetes_io_arch="${arch}"}
    Should Find Metrics In Prometheus Server
    ...    cluster:cpu_usage_cores:sum{_id="${cluster_id}"}
    Should Find Metrics In Prometheus Server
    ...    cluster:memory_usage_bytes:sum{_id="${cluster_id}"}
    Should Find Metrics In Prometheus Server
    ...    cluster:usage:resources:sum{_id="${cluster_id}",resource="pods"}
    Should Find Metrics In Prometheus Server
    ...    cluster:usage:resources:sum{_id="${cluster_id}",resource="namespaces"}
    Should Find Metrics In Prometheus Server
    ...    cluster:usage:resources:sum{_id="${cluster_id}",resource="services"}
    Should Find Metrics In Prometheus Server
    ...    cluster:usage:resources:sum{_id="${cluster_id}",resource="ingresses.networking.k8s.io"}
    Should Find Metrics In Prometheus Server
    ...    cluster:usage:resources:sum{_id="${cluster_id}",resource="routes.route.openshift.io"}
    Should Find Metrics In Prometheus Server
    ...    cluster:usage:resources:sum{_id="${cluster_id}",resource="customresourcedefinitions.apiextensions.k8s.io"}
    Should Find Metrics In Prometheus Server
    ...    cluster:usage:containers:sum{_id="${cluster_id}"}
    Should Find Metrics In Prometheus Server
    ...    microshift_version{_id="${cluster_id}",deployment_type="${deployment_type}",os_version_id="${os_version}",version="${microshift_version}"}

    [Teardown]    Remove Telemetry Configuration


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Check Required Telemetry Variables
    Login MicroShift Host
    Setup Kubeconfig

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host
    Remove Kubeconfig

Check Required Telemetry Variables
    [Documentation]    Check if the required telemetry variables are set
    Should Not Be Empty    ${PROXY_HOST}    PROXY_HOST variable is required
    ${string_value}=    Convert To String    ${PROXY_PORT}
    Should Not Be Empty    ${string_value}    PROXY_PORT variable is required
    Should Not Be Empty    ${PROMETHEUS_HOST}    PROMETHEUS_HOST variable is required
    ${string_value}=    Convert To String    ${PROMETHEUS_PORT}
    Should Not Be Empty    ${string_value}    PROMETHEUS_PORT variable is required
    Should Not Be Empty    ${PROXY_HOST}    PROXY_HOST variable is required
    ${string_value}=    Convert To String    ${PROXY_PORT}
    Should Not Be Empty    ${string_value}    PROXY_PORT variable is required

Setup Telemetry Configuration
    [Documentation]    Enables the telemetry feature in MicroShift configuration file
    ...    and restarts microshift.service
    [Arguments]    ${config}    ${new_pull_secret}
    Configure Pull Secrets    ${new_pull_secret}
    Drop In MicroShift Config    ${config}    10-telemetry
    Stop MicroShift
    ${cursor}=    Get Journal Cursor
    Set Test Variable    \${CURSOR}    ${cursor}
    Restart MicroShift

Setup Telemetry Configuration With Proxy
    [Documentation]    Start proxy server and setup Telemetry config
    [Arguments]    ${config}    ${new_pull_secret}
    Start Proxy Server    host=${PROXY_HOST}    port=${PROXY_PORT}
    Setup Telemetry Configuration    ${config}    ${new_pull_secret}

Remove Telemetry Configuration
    [Documentation]    Removes the telemetry feature from MicroShift configuration file
    ...    and restarts microshift.service
    Remove Drop In MicroShift Config    10-telemetry
    Restore Pull Secrets
    Restart MicroShift

Remove Telemetry Configuration With Proxy
    [Documentation]    Stop proxy server and remove Telemetry config
    Stop Proxy Server
    Remove Telemetry Configuration

Configure Pull Secrets
    [Documentation]    Sets up the pull secrets for the MicroShift cluster.
    [Arguments]    ${new_pull_secret}

    ${rc}=    SSHLibrary.Execute Command
    ...    grep -q cloud.openshift.com ${PULL_SECRET} || sudo ln -sf ${new_pull_secret} ${PULL_SECRET}
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

Should Find Metrics In Journal Log Success
    [Documentation]    Logs should contain metrics message
    [Arguments]    ${pattern}
    Wait Until Keyword Succeeds    10x    10s
    ...    Pattern Should Appear In Log Output    ${CURSOR}    ${pattern}

Should Find Metrics In Journal Log Fails
    [Documentation]    Logs should not contain metrics message
    [Arguments]    ${pattern}
    Pattern Should Not Appear In Log Output    ${CURSOR}    ${pattern}

Should Find Metrics In Prometheus Server
    [Documentation]    Check if the metric is on the Prometheus server
    [Arguments]    ${metric}
    Wait Until Keyword Succeeds    10x    10s
    ...    Check Prometheus Query    ${PROMETHEUS_HOST}    ${PROMETHEUS_PORT}    ${metric}
