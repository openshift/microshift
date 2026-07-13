*** Settings ***
Documentation       Router tuning, admission policies, and log format tests (disruptive)
...                 Migrated from openshift-tests-private:
...                 OCP-77349, OCP-80518, OCP-80520, OCP-82014

Resource            ../../resources/common.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/ostree-health.resource
Resource            ../../resources/router.resource
Variables           configs.py

Suite Setup         Setup Suite With Namespace
Suite Teardown      Teardown Suite With Namespace

Test Tags           restart    slow


*** Variables ***
${BASE_DOMAIN}      apps.example.com


*** Test Cases ***
Tuning Options Customization
    [Documentation]    Verify default tuning env vars and haproxy config values, then apply
    ...    custom tuning and verify all values are updated.
    ...    OCP-77349

    VAR    ${http_host}=    service-unsecure-ocp77349.${BASE_DOMAIN}
    Deploy Web Server
    Create OC Route    ${NAMESPACE}    http    route-http    service-unsecure    --hostname=${http_host}
    Verify Default Router Tuning Env Vars
    Verify Default Router Tuning Haproxy
    Setup Router Config And Restart    ${CONFIG_TUNING_CUSTOM}
    Verify Custom Router Tuning Env Vars
    Verify Custom Router Tuning Haproxy
    [Teardown]    Run Keywords
    ...    Remove Router Config And Restart
    ...    AND    Cleanup Test Workloads

MTLS Subject Filter
    [Documentation]    Verify mTLS with allowedSubjectPatterns allows a cert matching the CN filter
    ...    and blocks a cert not matching it.
    ...    OCP-80518
    [Setup]    Prepare Two MTLS Certs For Test    80518    route-edge80518.${BASE_DOMAIN}    route2-edge80518.${BASE_DOMAIN}

    Run With Kubeconfig
    ...    oc create configmap ocp80518 --from-file=ca-bundle.pem=${MTLS2_TMPDIR}/ca.crt -n ${ROUTER_NS}
    Setup Router Config And Restart    ${CONFIG_MTLS18_SUBJECT_FILTER}

    ${env}=    Oc Get JsonPath
    ...    pod
    ...    ${ROUTER_NS}
    ...    ${EMPTY}
    ...    .items[*].spec.containers[*].env[?(@.name=="ROUTER_MUTUAL_TLS_AUTH_FILTER")].value
    Should Contain    ${env}    example-test.com

    Deploy Web Server
    Deploy Test Client Pod
    Copy Files To Pod    ${NAMESPACE}    ${CLIENT_POD_NAME}    ${MTLS2_TMPDIR}    /data/certs
    Create OC Route    ${NAMESPACE}    edge    route-edge    service-unsecure
    ...    --hostname=${MTLS2_HOST1}    --cert=${MTLS2_TMPDIR}/usr1.crt    --key=${MTLS2_TMPDIR}/usr1.key
    Create OC Route    ${NAMESPACE}    edge    route-edge2    service-unsecure
    ...    --hostname=${MTLS2_HOST2}    --cert=${MTLS2_TMPDIR}/usr2.crt    --key=${MTLS2_TMPDIR}/usr2.key
    Route Should Be Admitted    route-edge
    Route Should Be Admitted    route-edge2

    ${router_ip}=    Get Router Pod IP
    Wait Until Curl With Cert File Succeeds
    ...    ${CLIENT_POD_NAME}    ${NAMESPACE}    https://${MTLS2_HOST1}    ${MTLS2_HOST1}:443:${router_ip}    usr1
    Wait Until Curl With Cert File Returns 403
    ...    ${CLIENT_POD_NAME}    ${NAMESPACE}    https://${MTLS2_HOST2}    ${MTLS2_HOST2}:443:${router_ip}    usr2
    [Teardown]    Run Keywords
    ...    Remove Router Config And Restart
    ...    AND    Run With Kubeconfig    oc delete configmap ocp80518 -n ${ROUTER_NS} --ignore-not-found
    ...    AND    Cleanup Test Workloads

Wildcard Route Admission Policy
    [Documentation]    Verify WildcardsDisallowed rejects wildcard routes, WildcardsAllowed admits them,
    ...    and reverting to WildcardsDisallowed rejects them again.
    ...    OCP-80520
    [Setup]    Run Keywords
    ...    Deploy Web Server
    ...    AND    Deploy Test Client Pod

    VAR    ${wildcard_host}=    wildcard.${BASE_DOMAIN}
    VAR    ${any_host}=    any.${BASE_DOMAIN}

    ${env}=    Oc Get JsonPath
    ...    pod
    ...    ${ROUTER_NS}
    ...    ${EMPTY}
    ...    .items[*].spec.containers[*].env[?(@.name=="ROUTER_ALLOW_WILDCARD_ROUTES")].value
    Should Be Equal As Strings    ${env}    false
    Create OC Route    ${NAMESPACE}    http    unsecure80520    service-unsecure
    ...    --hostname=${wildcard_host}    --wildcard-policy=Subdomain
    Route Should Not Be Admitted    unsecure80520

    Setup Router Config And Restart    ${CONFIG_WILDCARD_ALLOWED}
    Router Pod Env Should Have Value    ROUTER_ALLOW_WILDCARD_ROUTES    true
    Route Should Be Admitted    unsecure80520
    ${router_ip}=    Get Router Pod IP
    Wait Until Curl Succeeds From Pod
    ...    ${CLIENT_POD_NAME}    ${NAMESPACE}    http://${wildcard_host}    ${wildcard_host}:80:${router_ip}
    Wait Until Curl Succeeds From Pod
    ...    ${CLIENT_POD_NAME}    ${NAMESPACE}    http://${any_host}    ${any_host}:80:${router_ip}

    Setup Router Config And Restart    ${CONFIG_WILDCARD_DISALLOWED}
    Router Pod Env Should Have Value    ROUTER_ALLOW_WILDCARD_ROUTES    false
    Route Should Not Be Admitted    unsecure80520
    [Teardown]    Run Keywords
    ...    Remove Router Config And Restart
    ...    AND    Cleanup Test Workloads

HTTP Log Format
    [Documentation]    Verify httpLogFormat with HAProxy format directives produces structured log
    ...    output with the correct format.
    ...    OCP-82014

    Setup Router Config And Restart    ${CONFIG_LOG_FORMAT_1}
    Deploy Web Server
    Deploy Test Client Pod
    VAR    ${routehost}=    route-unsec82014.${BASE_DOMAIN}
    Create OC Route And Admit    ${NAMESPACE}    http    route-http    service-unsecure    --hostname=${routehost}
    ${router_ip}=    Get Router Pod IP
    Wait Until Curl Succeeds From Pod
    ...    ${CLIENT_POD_NAME}    ${NAMESPACE}
    ...    http://${routehost}/path/second/index.html    ${routehost}:80:${router_ip}
    Wait For Router Logs To Contain    /path/second/index.html
    ${logs}=    Get Router Access Logs
    Should Match Regexp    ${logs}    haproxy\\[[0-9]+\\]: "HEAD /path/second/index.html HTTP

    Setup Router Config And Restart    ${CONFIG_LOG_FORMAT_2}
    ${router_ip}=    Get Router Pod IP
    Wait Until Curl Succeeds From Pod
    ...    ${CLIENT_POD_NAME}    ${NAMESPACE}
    ...    http://${routehost}/path/second/index.html    ${routehost}:80:${router_ip}
    Wait For Router Logs To Contain    /path/second/index.html
    ${logs}=    Get Router Access Logs
    Should Match Regexp    ${logs}
    ...    haproxy\\[[0-9]+\\]: [0-9\\.a-fA-F:]+:[0-9]+ [0-9\\.a-fA-F:]+:8080 /path/second/index.html 200
    [Teardown]    Run Keywords
    ...    Remove Router Config And Restart
    ...    AND    Cleanup Test Workloads


*** Keywords ***
Verify Default Router Tuning Env Vars
    [Documentation]    Verify the default router tuning environment variable values.
    Router Pod Env Should Have Value    ROUTER_BUF_SIZE    32768
    Router Pod Env Should Have Value    ROUTER_MAX_REWRITE_SIZE    8192
    Router Pod Env Should Have Value    ROUTER_DEFAULT_CLIENT_TIMEOUT    30s
    Router Pod Env Should Have Value    ROUTER_CLIENT_FIN_TIMEOUT    1s
    Router Pod Env Should Have Value    ROUTER_DEFAULT_SERVER_TIMEOUT    30s
    Router Pod Env Should Have Value    ROUTER_DEFAULT_SERVER_FIN_TIMEOUT    1s
    Router Pod Env Should Have Value    ROUTER_DEFAULT_TUNNEL_TIMEOUT    1h
    Router Pod Env Should Have Value    ROUTER_INSPECT_DELAY    5s
    Router Pod Env Should Have Value    ROUTER_THREADS    4
    Router Pod Env Should Have Value    ROUTER_MAX_CONNECTIONS    50000

Verify Default Router Tuning Haproxy
    [Documentation]    Verify the default tuning values are present in haproxy.config.
    ${haproxy}=    Read Haproxy Config
    Should Contain    ${haproxy}    tune.bufsize 32768
    Should Contain    ${haproxy}    tune.maxrewrite 8192
    Should Contain    ${haproxy}    timeout client 30s
    Should Contain    ${haproxy}    timeout server 30s
    Should Contain    ${haproxy}    timeout tunnel 1h
    Should Contain    ${haproxy}    nbthread 4
    Should Contain    ${haproxy}    maxconn 50000

Verify Custom Router Tuning Env Vars
    [Documentation]    Verify the custom tuning environment variable values after config change.
    Router Pod Env Should Have Value    ROUTER_BUF_SIZE    65536
    Router Pod Env Should Have Value    ROUTER_MAX_REWRITE_SIZE    16384
    Router Pod Env Should Have Value    ROUTER_DEFAULT_CLIENT_TIMEOUT    1m
    Router Pod Env Should Have Value    ROUTER_CLIENT_FIN_TIMEOUT    2s
    Router Pod Env Should Have Value    ROUTER_DEFAULT_SERVER_TIMEOUT    1m
    Router Pod Env Should Have Value    ROUTER_DEFAULT_SERVER_FIN_TIMEOUT    2s
    Router Pod Env Should Have Value    ROUTER_DEFAULT_TUNNEL_TIMEOUT    2h
    Router Pod Env Should Have Value    ROUTER_INSPECT_DELAY    10s
    Router Pod Env Should Have Value    ROUTER_THREADS    8
    Router Pod Env Should Have Value    ROUTER_MAX_CONNECTIONS    100000

Verify Custom Router Tuning Haproxy
    [Documentation]    Verify the custom tuning values are present in haproxy.config after config change.
    ${haproxy}=    Read Haproxy Config
    Should Contain    ${haproxy}    tune.bufsize 65536
    Should Contain    ${haproxy}    tune.maxrewrite 16384
    Should Contain    ${haproxy}    timeout client 1m
    Should Contain    ${haproxy}    timeout server 1m
    Should Contain    ${haproxy}    timeout tunnel 2h
    Should Contain    ${haproxy}    nbthread 8
    Should Contain    ${haproxy}    maxconn 100000

Prepare Two MTLS Certs For Test
    [Documentation]    Generate CA and two client certs with different subjects for mTLS subject filter test.
    [Arguments]    ${case_id}    ${host1}    ${host2}
    VAR    ${tmpdir}=    /tmp/ocp-${case_id}-ca
    Create Directory    ${tmpdir}
    VAR    ${MTLS2_TMPDIR}=    ${tmpdir}    scope=TEST
    VAR    ${MTLS2_HOST1}=    ${host1}    scope=TEST
    VAR    ${MTLS2_HOST2}=    ${host2}    scope=TEST
    Generate CA Certificate    ${tmpdir}/ca.key    ${tmpdir}/ca.crt    /CN=MS-Test-Root-CA
    Generate Client Cert File In Dir    ${tmpdir}    ${host1}    example-test.com    usr1
    Generate Client Cert File In Dir    ${tmpdir}    ${host2}    example-test2.com    usr2
