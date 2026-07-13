*** Settings ***
Documentation       Router infrastructure configuration tests (disruptive)
...                 Migrated from openshift-tests-private:
...                 OCP-73203, OCP-73209, OCP-82015, OCP-84260

Resource            ../../resources/common.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/ostree-health.resource
Resource            ../../resources/router.resource
Variables           configs.py
Library             configs.py

Suite Setup         Setup Suite With Namespace
Suite Teardown      Teardown Suite With Namespace

Test Tags           restart    slow


*** Variables ***
${BASE_DOMAIN}          apps.example.com
${ALT_HTTP_PORT}        10080
${ALT_HTTPS_PORT}       10443


*** Test Cases ***
Custom Listening IPs And Ports
    [Documentation]    Verify configuring a specific listen address and custom HTTP/HTTPS ports
    ...    causes the router LB to expose only that IP on the custom ports, and routes
    ...    are reachable through those ports.
    ...    OCP-73203

    ${iface}    ${host_ip}=    Get First Host Interface And IP Via SSH
    ${config}=    Config Custom Listen    ${iface}    ${ALT_HTTP_PORT}    ${ALT_HTTPS_PORT}
    Setup Router Config And Restart    ${config}

    Verify Custom LB Ports And IP    ${host_ip}

    Deploy Web Server Signed
    Deploy Test Client Pod
    VAR    ${HTTP_HOST}=    service-unsecure-ocp73203.${BASE_DOMAIN}    scope=TEST
    VAR    ${EDGE_HOST}=    route-edge-ocp73203.${BASE_DOMAIN}    scope=TEST
    VAR    ${PASS_HOST}=    route-passth-ocp73203.${BASE_DOMAIN}    scope=TEST
    VAR    ${REEN_HOST}=    route-reen-ocp73203.${BASE_DOMAIN}    scope=TEST
    Create And Admit Four Route Types    ${HTTP_HOST}    ${EDGE_HOST}    ${PASS_HOST}    ${REEN_HOST}
    Curl Four Routes Via Custom Ports    ${host_ip}
    [Teardown]    Run Keywords
    ...    Remove Router Config And Restart
    ...    AND    Cleanup Test Workloads

Enable Disable Router
    [Documentation]    Verify setting ingress status to Removed deletes the openshift-ingress namespace,
    ...    and setting it back to Managed restores the router LB with correct IPs and ports.
    ...    OCP-73209

    Drop In MicroShift Config    ${CONFIG_ROUTER_REMOVED}    10-router
    Restart MicroShift
    Oc Wait    namespace/openshift-ingress    --for=delete --timeout=300s

    Setup Router Config And Restart    ${CONFIG_ROUTER_MANAGED}

    ${svc_type}=    Oc Get JsonPath    service    ${ROUTER_NS}    router-default    .spec.type
    Should Be Equal As Strings    ${svc_type}    LoadBalancer
    ${http_port}=    Get LB Port    http
    Should Be Equal As Strings    ${http_port}    80
    ${https_port}=    Get LB Port    https
    Should Be Equal As Strings    ${https_port}    443
    ${lb_ips}=    Get LB IPs
    Should Not Be Empty    ${lb_ips}
    [Teardown]    Remove Router Config And Restart

Syslog Logging Destination
    [Documentation]    Verify logging to a syslog server delivers router access logs to the syslog pod,
    ...    and changing the facility is reflected in the haproxy global config.
    ...    OCP-82015

    Privileged Namespace
    Deploy Rsyslogd Pod
    Deploy Web Server
    Deploy Test Client Pod
    ${syslog_ip}=    Oc Get JsonPath    pod    ${NAMESPACE}    rsyslogd-pod    .status.podIP
    Verify Syslog Logging    ${syslog_ip}
    [Teardown]    Run Keywords
    ...    Remove Router Config And Restart
    ...    AND    Cleanup Test Workloads

Negative Logging Config Validation
    [Documentation]    Verify invalid logging configurations are rejected by microshift show-config,
    ...    and that setting httpCaptureCookies without status: Enabled does not activate logging.
    ...    OCP-84260

    Show Invalid Drop In Config Should Fail With    ${LOGGING_INVALID_MAXLENGTH_NEG1}    Must be between 1 and 1024
    Show Invalid Drop In Config Should Fail With    ${LOGGING_INVALID_MAXLENGTH_ZERO}    Must be between 1 and 1024
    Show Invalid Drop In Config Should Fail With    ${LOGGING_INVALID_COOKIE_NAME}    contains invalid characters
    Show Invalid Drop In Config Should Fail With    ${LOGGING_INVALID_HEADER_MAXLENGTH}    maxLength must be at least 1
    Show Invalid Drop In Config Should Fail With    ${LOGGING_INVALID_STATUS}    invalid access logging status: Enable

    ${gen_before}=    Get Router Deployment Generation
    Drop In MicroShift Config    ${LOGGING_COOKIES_NO_STATUS}    10-router
    Restart MicroShift
    Wait For Router Ready
    ${gen_after}=    Get Router Deployment Generation
    Should Be Equal As Strings    ${gen_before}    ${gen_after}
    ${haproxy}=    Read Haproxy Config
    Should Not Contain    ${haproxy}    capture cookie foo len 100
    [Teardown]    Remove Router Config And Restart


*** Keywords ***
Verify Custom LB Ports And IP
    [Documentation]    Check the router-default LB has the expected IP and custom port numbers.
    [Arguments]    ${host_ip}
    ${lb_ips}=    Get LB IPs
    Should Be Equal As Strings    ${lb_ips}    ${host_ip}
    ${http_port}=    Get LB Port    http
    Should Be Equal As Strings    ${http_port}    ${ALT_HTTP_PORT}
    ${https_port}=    Get LB Port    https
    Should Be Equal As Strings    ${https_port}    ${ALT_HTTPS_PORT}

Curl Four Routes Via Custom Ports
    [Documentation]    Curl all four route types through the custom ports.
    [Arguments]    ${host_ip}
    Wait Until Curl Succeeds From Pod
    ...    ${CLIENT_POD_NAME}    ${NAMESPACE}
    ...    http://${HTTP_HOST}:${ALT_HTTP_PORT}    ${HTTP_HOST}:${ALT_HTTP_PORT}:${host_ip}
    Wait Until HTTPS Curl Succeeds From Pod
    ...    ${CLIENT_POD_NAME}    ${NAMESPACE}
    ...    https://${EDGE_HOST}:${ALT_HTTPS_PORT}    ${EDGE_HOST}:${ALT_HTTPS_PORT}:${host_ip}
    Wait Until HTTPS Curl Succeeds From Pod
    ...    ${CLIENT_POD_NAME}    ${NAMESPACE}
    ...    https://${PASS_HOST}:${ALT_HTTPS_PORT}    ${PASS_HOST}:${ALT_HTTPS_PORT}:${host_ip}
    Wait Until HTTPS Curl Succeeds From Pod
    ...    ${CLIENT_POD_NAME}    ${NAMESPACE}
    ...    https://${REEN_HOST}:${ALT_HTTPS_PORT}    ${REEN_HOST}:${ALT_HTTPS_PORT}:${host_ip}

Verify Syslog Logging
    [Documentation]    Apply syslog config and verify log delivery, then verify facility change.
    [Arguments]    ${syslog_ip}
    ${config1}=    Config Syslog    ${syslog_ip}
    Setup Router Config And Restart    ${config1}
    VAR    ${routehost}=    route-unsec82015.${BASE_DOMAIN}
    Create OC Route And Admit    ${NAMESPACE}    http    route-http    service-unsecure    --hostname=${routehost}
    Verify Syslog Haproxy Config    ${syslog_ip}    local1
    Verify Syslog Log Delivery    ${routehost}
    ${config2}=    Config Syslog    ${syslog_ip}    facility=local2
    Setup Router Config And Restart    ${config2}
    Verify Syslog Haproxy Config    ${syslog_ip}    local2

Verify Syslog Haproxy Config
    [Documentation]    Verify the syslog server address and facility in haproxy global config.
    [Arguments]    ${syslog_ip}    ${facility}
    ${haproxy}=    Read Haproxy Config
    Should Contain    ${haproxy}    log ${syslog_ip}:514 len 1024 ${facility} info

Verify Syslog Log Delivery
    [Documentation]    Curl a route and verify the log entry appears in the syslog pod.
    [Arguments]    ${routehost}
    ${router_ip}=    Get Router Pod IP
    Wait Until Curl Succeeds From Pod
    ...    ${CLIENT_POD_NAME}    ${NAMESPACE}
    ...    http://${routehost}/path/second/index.html    ${routehost}:80:${router_ip}
    Wait Until Keyword Succeeds    60s    3s    Syslog Pod Should Contain    /path/second/index.html

Syslog Pod Should Contain
    [Documentation]    Check that the rsyslogd pod logs contain a pattern.
    [Arguments]    ${pattern}
    ${logs}=    Oc Logs    rsyslogd-pod --tail=20    ${NAMESPACE}
    Should Contain    ${logs}    ${pattern}
