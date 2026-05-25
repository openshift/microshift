*** Settings ***
Documentation       Router ingress controller configuration tests (disruptive)
...                 Migrated from openshift-tests-private:
...                 OCP-73203, OCP-73209, OCP-77349, OCP-80508, OCP-80510, OCP-80514,
...                 OCP-80517, OCP-80518, OCP-80520, OCP-81996, OCP-81997, OCP-82000,
...                 OCP-82003, OCP-82004, OCP-82014, OCP-82015, OCP-84260

Resource            ../../resources/common.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/ostree-health.resource
Resource            ../../resources/router.resource

Suite Setup         Setup Suite With Namespace
Suite Teardown      Teardown Suite With Namespace

Test Tags           restart    slow


*** Variables ***
${BASE_DOMAIN}                          apps.example.com
${ALT_HTTP_PORT}                        10080
${ALT_HTTPS_PORT}                       10443

${CONFIG_OLD_TLS}                       SEPARATOR=\n
...                                     ---
...                                     ingress:
...                                     \ \ tlsSecurityProfile:
...                                     \ \ \ \ old: {}
...                                     \ \ \ \ type: Old

${CONFIG_INTERMEDIATE_TLS}              SEPARATOR=\n
...                                     ---
...                                     ingress:
...                                     \ \ tlsSecurityProfile:
...                                     \ \ \ \ intermediate: {}
...                                     \ \ \ \ type: Intermediate

${CONFIG_MODERN_TLS}                    SEPARATOR=\n
...                                     ---
...                                     ingress:
...                                     \ \ tlsSecurityProfile:
...                                     \ \ \ \ modern: {}
...                                     \ \ \ \ type: Modern

${CONFIG_CUSTOM_TLS}                    SEPARATOR=\n
...                                     ---
...                                     ingress:
...                                     \ \ tlsSecurityProfile:
...                                     \ \ \ \ custom:
...                                     \ \ \ \ \ \ ciphers:
...                                     \ \ \ \ \ \ - DHE-RSA-AES256-GCM-SHA384
...                                     \ \ \ \ \ \ - ECDHE-ECDSA-AES256-GCM-SHA384
...                                     \ \ \ \ \ \ minTLSVersion: VersionTLS12
...                                     \ \ \ \ type: Custom

${CONFIG_WILDCARD_ALLOWED}              SEPARATOR=\n
...                                     ---
...                                     ingress:
...                                     \ \ routeAdmissionPolicy:
...                                     \ \ \ \ wildcardPolicy: "WildcardsAllowed"

${CONFIG_WILDCARD_DISALLOWED}           SEPARATOR=\n
...                                     ---
...                                     ingress:
...                                     \ \ routeAdmissionPolicy:
...                                     \ \ \ \ wildcardPolicy: "WildcardsDisallowed"

${CONFIG_TUNING_CUSTOM}                 SEPARATOR=\n
...                                     ---
...                                     ingress:
...                                     \ \ forwardedHeaderPolicy: "Replace"
...                                     \ \ httpCompression:
...                                     \ \ \ \ mimeTypes:
...                                     \ \ \ \ - "image"
...                                     \ \ logEmptyRequests: "Ignore"
...                                     \ \ tuningOptions:
...                                     \ \ \ \ clientFinTimeout: "2s"
...                                     \ \ \ \ clientTimeout: "60s"
...                                     \ \ \ \ headerBufferBytes: 65536
...                                     \ \ \ \ headerBufferMaxRewriteBytes: 16384
...                                     \ \ \ \ healthCheckInterval: "10s"
...                                     \ \ \ \ maxConnections: 100000
...                                     \ \ \ \ serverFinTimeout: "2s"
...                                     \ \ \ \ serverTimeout: "60s"
...                                     \ \ \ \ threadCount: 8
...                                     \ \ \ \ tlsInspectDelay: "10s"
...                                     \ \ \ \ tunnelTimeout: "2h"

${CONFIG_MTLS17_REQUIRED}               SEPARATOR=\n
...                                     ---
...                                     ingress:
...                                     \ \ clientTLS:
...                                     \ \ \ \ clientCA:
...                                     \ \ \ \ \ \ name: "ocp80517"
...                                     \ \ \ \ clientCertificatePolicy: "Required"

${CONFIG_MTLS17_OPTIONAL}               SEPARATOR=\n
...                                     ---
...                                     ingress:
...                                     \ \ clientTLS:
...                                     \ \ \ \ clientCA:
...                                     \ \ \ \ \ \ name: "ocp80517"
...                                     \ \ \ \ clientCertificatePolicy: "Optional"

${CONFIG_MTLS18_SUBJECT_FILTER}         SEPARATOR=\n
...                                     ---
...                                     ingress:
...                                     \ \ clientTLS:
...                                     \ \ \ \ allowedSubjectPatterns: ["/CN=example-test.com"]
...                                     \ \ \ \ clientCA:
...                                     \ \ \ \ \ \ name: "ocp80518"
...                                     \ \ \ \ clientCertificatePolicy: "Required"

${CONFIG_COOKIE_EXACT_100}              SEPARATOR=\n
...                                     ---
...                                     ingress:
...                                     \ \ accessLogging:
...                                     \ \ \ \ httpCaptureCookies:
...                                     \ \ \ \ - matchType: Exact
...                                     \ \ \ \ \ \ maxLength: 100
...                                     \ \ \ \ \ \ name: foo
...                                     \ \ \ \ status: Enabled

${CONFIG_COOKIE_EXACT_10}               SEPARATOR=\n
...                                     ---
...                                     ingress:
...                                     \ \ accessLogging:
...                                     \ \ \ \ httpCaptureCookies:
...                                     \ \ \ \ - matchType: Exact
...                                     \ \ \ \ \ \ maxLength: 10
...                                     \ \ \ \ \ \ name: foo
...                                     \ \ \ \ status: Enabled

${CONFIG_LOG_FORMAT_1}                  SEPARATOR=\n
...                                     ---
...                                     ingress:
...                                     \ \ accessLogging:
...                                     \ \ \ \ httpLogFormat: "%{+Q}r"
...                                     \ \ \ \ status: Enabled

${CONFIG_LOG_FORMAT_2}                  SEPARATOR=\n
...                                     ---
...                                     ingress:
...                                     \ \ accessLogging:
...                                     \ \ \ \ httpLogFormat: "%ci:%cp %si:%sp %HU %ST"
...                                     \ \ \ \ status: Enabled

${LOGGING_INVALID_MAXLENGTH_NEG1}       SEPARATOR=\n
...                                     ---
...                                     ingress:
...                                     \ \ accessLogging:
...                                     \ \ \ \ httpCaptureCookies:
...                                     \ \ \ \ - matchType: Exact
...                                     \ \ \ \ \ \ maxLength: -1
...                                     \ \ \ \ \ \ name: foo
...                                     \ \ \ \ status: Enabled

${LOGGING_INVALID_MAXLENGTH_ZERO}       SEPARATOR=\n
...                                     ---
...                                     ingress:
...                                     \ \ accessLogging:
...                                     \ \ \ \ httpCaptureCookies:
...                                     \ \ \ \ - matchType: Exact
...                                     \ \ \ \ \ \ maxLength: 0
...                                     \ \ \ \ \ \ name: foo
...                                     \ \ \ \ status: Enabled

${LOGGING_INVALID_COOKIE_NAME}          SEPARATOR=\n
...                                     ---
...                                     ingress:
...                                     \ \ accessLogging:
...                                     \ \ \ \ httpCaptureCookies:
...                                     \ \ \ \ - matchType: Exact
...                                     \ \ \ \ \ \ maxLength: 100
...                                     \ \ \ \ \ \ name: "foo 33#?-"
...                                     \ \ \ \ status: Enabled

${LOGGING_INVALID_HEADER_MAXLENGTH}     SEPARATOR=\n
...                                     ---
...                                     ingress:
...                                     \ \ accessLogging:
...                                     \ \ \ \ httpCaptureHeaders:
...                                     \ \ \ \ \ \ request:
...                                     \ \ \ \ \ \ - maxLength: -1
...                                     \ \ \ \ \ \ \ \ name: Host
...                                     \ \ \ \ \ \ response:
...                                     \ \ \ \ \ \ - maxLength: 10
...                                     \ \ \ \ \ \ \ \ name: "Server"
...                                     \ \ \ \ status: Enabled

${LOGGING_INVALID_STATUS}               SEPARATOR=\n
...                                     ---
...                                     ingress:
...                                     \ \ accessLogging:
...                                     \ \ \ \ httpCaptureHeaders:
...                                     \ \ \ \ \ \ request:
...                                     \ \ \ \ \ \ - maxLength: 10
...                                     \ \ \ \ \ \ \ \ name: Host
...                                     \ \ \ \ status: Enable

${LOGGING_COOKIES_NO_STATUS}            SEPARATOR=\n
...                                     ---
...                                     ingress:
...                                     \ \ accessLogging:
...                                     \ \ \ \ httpCaptureCookies:
...                                     \ \ \ \ - matchType: Prefix
...                                     \ \ \ \ \ \ maxLength: 100
...                                     \ \ \ \ \ \ namePrefix: foo


*** Test Cases ***
Custom Listening IPs And Ports
    [Documentation]    Verify configuring a specific listen address and custom HTTP/HTTPS ports
    ...    causes the router LB to expose only that IP on the custom ports, and routes
    ...    are reachable through those ports.
    ...    OCP-73203

    ${iface}    ${host_ip}=    Get First Host Interface And IP Via SSH
    ${config}=    Catenate    SEPARATOR=\n
    ...    ---
    ...    ingress:
    ...    \ \ listenAddress:
    ...    \ \ - ${iface}
    ...    \ \ ports:
    ...    \ \ \ \ http: ${ALT_HTTP_PORT}
    ...    \ \ \ \ https: ${ALT_HTTPS_PORT}
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
    [Teardown]    Remove Router Config And Restart

Enable Disable Router
    [Documentation]    Verify setting ingress status to Removed deletes the openshift-ingress namespace,
    ...    and setting it back to Managed restores the router LB with correct IPs and ports.
    ...    OCP-73209

    ${config_removed}=    Catenate    SEPARATOR=\n
    ...    ---
    ...    ingress:
    ...    \ \ status: Removed
    Setup Router Config And Restart    ${config_removed}
    Oc Wait    namespace/openshift-ingress    --for=delete --timeout=300s

    ${config_managed}=    Catenate    SEPARATOR=\n
    ...    ---
    ...    ingress:
    ...    \ \ status: Managed
    Setup Router Config And Restart    ${config_managed}

    ${svc_type}=    Oc Get JsonPath    service    ${ROUTER_NS}    router-default    .spec.type
    Should Be Equal As Strings    ${svc_type}    LoadBalancer
    ${http_port}=    Get LB Port    http
    Should Be Equal As Strings    ${http_port}    80
    ${https_port}=    Get LB Port    https
    Should Be Equal As Strings    ${https_port}    443
    ${lb_ips}=    Get LB IPs
    Should Not Be Empty    ${lb_ips}
    [Teardown]    Remove Router Config And Restart

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
    [Teardown]    Remove Router Config And Restart

Custom Default Certificate
    [Documentation]    Verify a custom TLS certificate secret can be configured as the default
    ...    ingress certificate, and that routes use the custom cert for TLS.
    ...    OCP-80508
    [Setup]    Prepare Custom Cert For Test    80508    route-edge80508.${BASE_DOMAIN}

    ${config}=    Catenate    SEPARATOR=\n
    ...    ---
    ...    ingress:
    ...    \ \ certificateSecret: "router-test-cert"
    Setup Router Config And Restart    ${config}

    Create OC Route    ${NAMESPACE}    edge    route-edge    service-unsecure    --hostname=${CERT_EDGE_HOST}
    Route Should Be Admitted    route-edge
    Verify Custom Cert Is Active
    Deploy Test Client Pod
    Copy Files To Pod    ${NAMESPACE}    ${CLIENT_POD_NAME}    ${CERT_TMPDIR}    /data/certs
    ${router_ip}=    Get Router Pod IP
    VAR    ${resolve}=    ${CERT_EDGE_HOST}:443:${router_ip}
    Wait Until Curl With Client Cert Succeeds
    ...    ${CLIENT_POD_NAME}    ${NAMESPACE}    https://${CERT_EDGE_HOST}    ${resolve}
    Wait Until HTTPS Curl Succeeds From Pod
    ...    ${CLIENT_POD_NAME}    ${NAMESPACE}    https://${CERT_EDGE_HOST}    ${resolve}
    [Teardown]    Run Keywords
    ...    Remove Router Config And Restart
    ...    AND    Run With Kubeconfig    oc delete secret router-test-cert -n ${ROUTER_NS} --ignore-not-found

Old And Intermediate TLS Profiles
    [Documentation]    Verify the default Intermediate TLS profile cipher settings, then apply Old
    ...    profile and verify updated cipher settings, then restore Intermediate.
    ...    OCP-80510

    Router Pod Env Should Have Value    SSL_MIN_VERSION    TLSv1.2
    ${ciphers}=    Oc Get JsonPath
    ...    pod    ${ROUTER_NS}    ${EMPTY}    .items[*].spec.containers[*].env[?(@.name=="ROUTER_CIPHERSUITES")].value
    Should Contain    ${ciphers}    TLS_AES_128_GCM_SHA256

    Setup Router Config And Restart    ${CONFIG_OLD_TLS}
    Router Pod Env Should Have Value    SSL_MIN_VERSION    TLSv1.1
    ${ciphers}=    Oc Get JsonPath
    ...    pod    ${ROUTER_NS}    ${EMPTY}    .items[*].spec.containers[*].env[?(@.name=="ROUTER_CIPHERS")].value
    Should Contain    ${ciphers}    DES-CBC3-SHA

    Setup Router Config And Restart    ${CONFIG_INTERMEDIATE_TLS}
    Router Pod Env Should Have Value    SSL_MIN_VERSION    TLSv1.2
    ${ciphers}=    Oc Get JsonPath
    ...    pod    ${ROUTER_NS}    ${EMPTY}    .items[*].spec.containers[*].env[?(@.name=="ROUTER_CIPHERSUITES")].value
    Should Contain    ${ciphers}    TLS_AES_128_GCM_SHA256
    [Teardown]    Remove Router Config And Restart

Modern And Custom TLS Profiles
    [Documentation]    Verify Modern TLS profile enforces TLSv1.3 in env vars and haproxy config,
    ...    then apply a Custom profile with specific ciphers.
    ...    OCP-80514

    Setup Router Config And Restart    ${CONFIG_MODERN_TLS}
    Router Pod Env Should Have Value    SSL_MIN_VERSION    TLSv1.3
    ${ciphers}=    Oc Get JsonPath
    ...    pod    ${ROUTER_NS}    ${EMPTY}    .items[*].spec.containers[*].env[?(@.name=="ROUTER_CIPHERSUITES")].value
    Should Contain    ${ciphers}    TLS_AES_128_GCM_SHA256
    ${haproxy}=    Read Haproxy Config
    Should Contain    ${haproxy}    ssl-default-bind-options ssl-min-ver TLSv1.3

    Setup Router Config And Restart    ${CONFIG_CUSTOM_TLS}
    Router Pod Env Should Have Value    SSL_MIN_VERSION    TLSv1.2
    ${ciphers}=    Oc Get JsonPath
    ...    pod    ${ROUTER_NS}    ${EMPTY}    .items[*].spec.containers[*].env[?(@.name=="ROUTER_CIPHERS")].value
    Should Contain    ${ciphers}    DHE-RSA-AES256-GCM-SHA384
    [Teardown]    Remove Router Config And Restart

MTLS Optional And Required Policy
    [Documentation]    Verify mTLS with clientCertificatePolicy Required rejects connections without
    ...    client cert, and Optional policy allows them.
    ...    OCP-80517
    [Setup]    Prepare MTLS Cert For Test    80517    route-edge80517.${BASE_DOMAIN}

    Run With Kubeconfig    oc create configmap ocp80517 --from-file=ca-bundle.pem=${MTLS_TMPDIR}/ca.crt -n ${ROUTER_NS}

    Setup Router Config And Restart    ${CONFIG_MTLS17_REQUIRED}
    Router Pod Env Should Have Value    ROUTER_MUTUAL_TLS_AUTH    required
    Deploy MTLS Test Workloads
    ${router_ip}=    Get Router Pod IP
    VAR    ${resolve}=    ${MTLS_EDGE_HOST}:443:${router_ip}
    Wait Until Curl With Client Cert Succeeds
    ...    ${CLIENT_POD_NAME}
    ...    ${NAMESPACE}
    ...    https://${MTLS_EDGE_HOST}
    ...    ${resolve}
    Curl Without Cert Should Return SSL Error
    ...    ${CLIENT_POD_NAME}
    ...    ${NAMESPACE}
    ...    https://${MTLS_EDGE_HOST}
    ...    ${resolve}

    Setup Router Config And Restart    ${CONFIG_MTLS17_OPTIONAL}
    Router Pod Env Should Have Value    ROUTER_MUTUAL_TLS_AUTH    optional
    ${router_ip}=    Get Router Pod IP
    VAR    ${resolve}=    ${MTLS_EDGE_HOST}:443:${router_ip}
    Wait Until Curl With Client Cert Succeeds
    ...    ${CLIENT_POD_NAME}
    ...    ${NAMESPACE}
    ...    https://${MTLS_EDGE_HOST}
    ...    ${resolve}
    Wait Until HTTPS Curl Succeeds From Pod
    ...    ${CLIENT_POD_NAME}
    ...    ${NAMESPACE}
    ...    https://${MTLS_EDGE_HOST}
    ...    ${resolve}
    [Teardown]    Run Keywords
    ...    Remove Router Config And Restart
    ...    AND    Run With Kubeconfig    oc delete configmap ocp80517 -n ${ROUTER_NS} --ignore-not-found

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
    [Teardown]    Remove Router Config And Restart

HTTP Capture Cookies Prefix Match
    [Documentation]    Verify httpCaptureCookies with Prefix match captures cookies matching the
    ...    prefix in router logs across HTTP, edge, and reencrypt routes.
    ...    OCP-81996

    ${config}=    Catenate    SEPARATOR=\n
    ...    ---
    ...    ingress:
    ...    \ \ accessLogging:
    ...    \ \ \ \ httpCaptureCookies:
    ...    \ \ \ \ - matchType: Prefix
    ...    \ \ \ \ \ \ maxLength: 100
    ...    \ \ \ \ \ \ namePrefix: foo
    ...    \ \ \ \ status: Enabled
    Setup Router Config And Restart    ${config}

    Deploy Web Server Signed
    Deploy Test Client Pod
    VAR    ${routehost}=    route-unsec81996.${BASE_DOMAIN}
    VAR    ${edge_host}=    route-edge81996.${BASE_DOMAIN}
    VAR    ${reen_host}=    route-reen81996.${BASE_DOMAIN}
    Create OC Route    ${NAMESPACE}    http    route-http    service-unsecure    --hostname=${routehost}
    Route Should Be Admitted    route-http
    Create OC Route    ${NAMESPACE}    edge    route-edge    service-unsecure    --hostname=${edge_host}
    Route Should Be Admitted    route-edge
    Create OC Route    ${NAMESPACE}    reencrypt    route-reen    service-secure    --hostname=${reen_host}
    Route Should Be Admitted    route-reen
    ${router_ip}=    Get Router Pod IP
    Curl All Cookie Routes And Verify Logs    ${routehost}    ${edge_host}    ${reen_host}    ${router_ip}
    [Teardown]    Remove Router Config And Restart

HTTP Capture Cookies Exact Match And MaxLength
    [Documentation]    Verify httpCaptureCookies with Exact match captures only exact-named cookies,
    ...    and maxLength truncates cookie values in logs.
    ...    OCP-81997

    Setup Router Config And Restart    ${CONFIG_COOKIE_EXACT_100}
    Deploy Web Server
    Deploy Test Client Pod
    VAR    ${routehost}=    route-unsec81997.${BASE_DOMAIN}
    Create OC Route    ${NAMESPACE}    http    route-http    service-unsecure    --hostname=${routehost}
    Route Should Be Admitted    route-http
    ${router_ip}=    Get Router Pod IP
    Wait Until Curl With Cookie Succeeds From Pod
    ...    ${CLIENT_POD_NAME}
    ...    ${NAMESPACE}
    ...    http://${routehost}/index.html
    ...    ${routehost}:80:${router_ip}
    ...    fooor=nobar
    Wait Until Curl With Cookie Succeeds From Pod
    ...    ${CLIENT_POD_NAME}
    ...    ${NAMESPACE}
    ...    http://${routehost}/index.html
    ...    ${routehost}:80:${router_ip}
    ...    foo=bar
    Wait For Router Logs To Contain    foo=bar
    Router Logs Should Not Contain    fooor=nobar

    Setup Router Config And Restart    ${CONFIG_COOKIE_EXACT_10}
    ${router_ip}=    Get Router Pod IP
    Wait Until Curl With Cookie Succeeds From Pod
    ...    ${CLIENT_POD_NAME}
    ...    ${NAMESPACE}
    ...    http://${routehost}/index.html
    ...    ${routehost}:80:${router_ip}
    ...    foo=bar89abdef
    Wait For Router Logs To Contain    foo=bar89a
    Router Logs Should Not Contain    foo=bar89ab
    [Teardown]    Remove Router Config And Restart

HTTP Capture Headers Request And Response
    [Documentation]    Verify httpCaptureHeaders captures request Host and response Server headers
    ...    in router logs, including for edge and reencrypt routes.
    ...    OCP-82000

    ${config}=    Catenate    SEPARATOR=\n
    ...    ---
    ...    ingress:
    ...    \ \ accessLogging:
    ...    \ \ \ \ httpCaptureHeaders:
    ...    \ \ \ \ \ \ request:
    ...    \ \ \ \ \ \ - maxLength: 120
    ...    \ \ \ \ \ \ \ \ name: Host
    ...    \ \ \ \ \ \ response:
    ...    \ \ \ \ \ \ - maxLength: 120
    ...    \ \ \ \ \ \ \ \ name: "Server"
    ...    \ \ \ \ status: Enabled
    Setup Router Config And Restart    ${config}

    Deploy Web Server Signed
    Deploy Test Client Pod
    VAR    ${routehost}=    route-unsec82000.${BASE_DOMAIN}
    VAR    ${edge_host}=    route-edge82000.${BASE_DOMAIN}
    VAR    ${reen_host}=    route-reen82000.${BASE_DOMAIN}
    Create OC Route    ${NAMESPACE}    http    route-http    service-unsecure    --hostname=${routehost}
    Route Should Be Admitted    route-http
    Create OC Route    ${NAMESPACE}    edge    route-edge    service-unsecure    --hostname=${edge_host}
    Route Should Be Admitted    route-edge
    Create OC Route    ${NAMESPACE}    reencrypt    route-reen    service-secure    --hostname=${reen_host}
    Route Should Be Admitted    route-reen
    Verify Header Capture Config And Logs    ${routehost}    ${edge_host}    ${reen_host}
    [Teardown]    Remove Router Config And Restart

HTTP Capture Headers MaxLength Adherence
    [Documentation]    Verify httpCaptureHeaders maxLength truncates captured header values in logs.
    ...    OCP-82003

    VAR    ${routehost}=    route-unsec82003.${BASE_DOMAIN}
    ${config}=    Catenate    SEPARATOR=\n
    ...    ---
    ...    ingress:
    ...    \ \ accessLogging:
    ...    \ \ \ \ httpCaptureHeaders:
    ...    \ \ \ \ \ \ request:
    ...    \ \ \ \ \ \ - maxLength: 16
    ...    \ \ \ \ \ \ \ \ name: Host
    ...    \ \ \ \ \ \ response:
    ...    \ \ \ \ \ \ - maxLength: 5
    ...    \ \ \ \ \ \ \ \ name: "Server"
    ...    \ \ \ \ status: Enabled
    Setup Router Config And Restart    ${config}

    Deploy Web Server
    Deploy Test Client Pod
    Create OC Route    ${NAMESPACE}    http    route-http    service-unsecure    --hostname=${routehost}
    Route Should Be Admitted    route-http
    ${router_ip}=    Get Router Pod IP
    Wait Until Curl Succeeds From Pod
    ...    ${CLIENT_POD_NAME}    ${NAMESPACE}    http://${routehost}/index.html    ${routehost}:80:${router_ip}
    # maxLength is 16, so hostname is truncated to exactly 16 chars: route-unsec82003
    Wait For Router Logs To Contain    route-unsec82003
    ${logs}=    Get Router Access Logs
    Should Not Contain    ${logs}    ${routehost}
    # nginx server version is 5+ chars, so only "nginx" should appear without the version
    Should Contain    ${logs}    nginx
    Should Not Contain    ${logs}    nginx/
    [Teardown]    Remove Router Config And Restart

Custom HTTP Error Pages
    [Documentation]    Verify custom 503 and 404 error pages are served when configured via
    ...    httpErrorCodePages configmap.
    ...    OCP-82004

    Create Configmap From Files    ${ROUTER_NS}    custom-82004-error-code-pages
    ...    --from-file=./assets/router/error-page-503.http    --from-file=./assets/router/error-page-404.http
    ${config}=    Catenate    SEPARATOR=\n
    ...    ---
    ...    ingress:
    ...    \ \ httpErrorCodePages:
    ...    \ \ \ \ name: custom-82004-error-code-pages
    Setup Router Config And Restart    ${config}

    Deploy Web Server
    Deploy Test Client Pod
    VAR    ${routehost}=    route-unsec82004.${BASE_DOMAIN}
    VAR    ${noexist_host}=    not-exist82004.${BASE_DOMAIN}
    Create OC Route    ${NAMESPACE}    http    route-http    service-unsecure    --hostname=${routehost}
    Route Should Be Admitted    route-http
    Verify Custom Error Pages    ${routehost}    ${noexist_host}
    [Teardown]    Run Keywords
    ...    Remove Router Config And Restart
    ...    AND    Run With Kubeconfig    oc delete configmap custom-82004-error-code-pages -n ${ROUTER_NS} --ignore-not-found

HTTP Log Format
    [Documentation]    Verify httpLogFormat with HAProxy format directives produces structured log
    ...    output with the correct format.
    ...    OCP-82014

    Setup Router Config And Restart    ${CONFIG_LOG_FORMAT_1}
    Deploy Web Server
    Deploy Test Client Pod
    VAR    ${routehost}=    route-unsec82014.${BASE_DOMAIN}
    Create OC Route    ${NAMESPACE}    http    route-http    service-unsecure    --hostname=${routehost}
    Route Should Be Admitted    route-http
    ${router_ip}=    Get Router Pod IP
    Wait Until Curl Succeeds From Pod
    ...    ${CLIENT_POD_NAME}
    ...    ${NAMESPACE}
    ...    http://${routehost}/path/second/index.html
    ...    ${routehost}:80:${router_ip}
    Wait For Router Logs To Contain    /path/second/index.html
    ${logs}=    Get Router Access Logs
    Should Match Regexp    ${logs}    haproxy\\[[0-9]+\\]: "HEAD /path/second/index.html HTTP

    Setup Router Config And Restart    ${CONFIG_LOG_FORMAT_2}
    ${router_ip}=    Get Router Pod IP
    Wait Until Curl Succeeds From Pod
    ...    ${CLIENT_POD_NAME}
    ...    ${NAMESPACE}
    ...    http://${routehost}/path/second/index.html
    ...    ${routehost}:80:${router_ip}
    Wait For Router Logs To Contain    /path/second/index.html
    ${logs}=    Get Router Access Logs
    Should Match Regexp    ${logs}
    ...    haproxy\\[[0-9]+\\]: [0-9\\.a-fA-F:]+:[0-9]+ [0-9\\.a-fA-F:]+:8080 /path/second/index.html 200
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
    [Teardown]    Remove Router Config And Restart

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

Prepare Custom Cert For Test
    [Documentation]    Generate CA and user cert for custom certificate test, create TLS secret.
    [Arguments]    ${case_id}    ${edge_host}
    VAR    ${tmpdir}=    /tmp/ocp-${case_id}
    Create Directory    ${tmpdir}
    VAR    ${CERT_TMPDIR}=    ${tmpdir}    scope=TEST
    VAR    ${CERT_EDGE_HOST}=    ${edge_host}    scope=TEST
    Generate CA Certificate    ${tmpdir}/ca.key    ${tmpdir}/ca.crt    /CN=MS-default-CA
    ${san}=    Catenate    SEPARATOR=\n
    ...    [ v3_req ]
    ...    subjectAltName = @alt_names
    ...    [ alt_names ]
    ...    DNS.1 = *.${BASE_DOMAIN}
    Generate CSR And Key    ${tmpdir}/usr.key    ${tmpdir}/usr.csr    /CN=example-ne.com
    Sign CSR With CA    ${tmpdir}/usr.csr    ${tmpdir}/ca.crt    ${tmpdir}/ca.key    ${tmpdir}/usr.crt    ${san}
    Run With Kubeconfig
    ...    oc create secret tls router-test-cert --cert=${tmpdir}/ca.crt --key=${tmpdir}/ca.key -n ${ROUTER_NS}
    Deploy Web Server

Verify Custom Cert Is Active
    [Documentation]    Verify the custom cert secret is mounted and the cert issuer is correct.
    ${vol}=    Oc Get JsonPath
    ...    deployment
    ...    ${ROUTER_NS}
    ...    router-default
    ...    ..volumes[?(@.name=="default-certificate")].secret.secretName
    Should Contain    ${vol}    router-test-cert
    ${router_pod}=    Get Router Pod Name
    ${cert_info}=    Run With Kubeconfig
    ...    oc exec -n ${ROUTER_NS} ${router_pod} -- openssl x509 -noout -in /etc/pki/tls/private/tls.crt -text
    Should Contain    ${cert_info}    CN = MS-default-CA

Prepare MTLS Cert For Test
    [Documentation]    Generate CA and client cert for mTLS tests.
    [Arguments]    ${case_id}    ${edge_host}
    VAR    ${tmpdir}=    /tmp/ocp-${case_id}-ca
    Create Directory    ${tmpdir}
    VAR    ${MTLS_TMPDIR}=    ${tmpdir}    scope=TEST
    VAR    ${MTLS_EDGE_HOST}=    ${edge_host}    scope=TEST
    Generate MTLS Client Cert    ${tmpdir}    ${edge_host}

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

Deploy MTLS Test Workloads
    [Documentation]    Deploy workloads and create the mTLS edge route.
    Deploy Web Server
    Deploy Test Client Pod
    Copy Files To Pod    ${NAMESPACE}    ${CLIENT_POD_NAME}    ${MTLS_TMPDIR}    /data/certs
    Create OC Route    ${NAMESPACE}    edge    route-edge    service-unsecure
    ...    --hostname=${MTLS_EDGE_HOST}    --cert=${MTLS_TMPDIR}/usr.crt    --key=${MTLS_TMPDIR}/usr.key
    Route Should Be Admitted    route-edge

Curl All Cookie Routes And Verify Logs
    [Documentation]    Curl all routes with cookies and verify correct log entries.
    [Arguments]    ${routehost}    ${edge_host}    ${reen_host}    ${router_ip}
    Wait Until Curl With Cookie Succeeds From Pod
    ...    ${CLIENT_POD_NAME}
    ...    ${NAMESPACE}
    ...    http://${routehost}/index.html
    ...    ${routehost}:80:${router_ip}
    ...    fo=nobar
    Wait Until Curl With Cookie Succeeds From Pod
    ...    ${CLIENT_POD_NAME}
    ...    ${NAMESPACE}
    ...    http://${routehost}/index.html
    ...    ${routehost}:80:${router_ip}
    ...    foo=bar
    Wait Until Curl With Cookie Succeeds From Pod
    ...    ${CLIENT_POD_NAME}
    ...    ${NAMESPACE}
    ...    http://${routehost}/index.html
    ...    ${routehost}:80:${router_ip}
    ...    foo22=bar22
    Wait Until HTTPS Curl With Cookie Succeeds From Pod
    ...    ${CLIENT_POD_NAME}
    ...    ${NAMESPACE}
    ...    https://${edge_host}/index.html
    ...    ${edge_host}:443:${router_ip}
    ...    foo=barforedge
    Wait Until HTTPS Curl With Cookie Succeeds From Pod
    ...    ${CLIENT_POD_NAME}
    ...    ${NAMESPACE}
    ...    https://${reen_host}/index.html
    ...    ${reen_host}:443:${router_ip}
    ...    foo=barforreen
    Wait For Router Logs To Contain    foo=bar
    Wait For Router Logs To Contain    foo22=bar22
    Wait For Router Logs To Contain    foo=barforedge
    Wait For Router Logs To Contain    foo=barforreen
    Router Logs Should Not Contain    fo=nobar

Verify Header Capture Config And Logs
    [Documentation]    Verify haproxy header capture config and check captured headers in logs.
    [Arguments]    ${routehost}    ${edge_host}    ${reen_host}
    Verify Header Capture In Haproxy
    ${router_ip}=    Get Router Pod IP
    Wait Until Curl Succeeds From Pod
    ...    ${CLIENT_POD_NAME}    ${NAMESPACE}    http://${routehost}/index.html    ${routehost}:80:${router_ip}
    Wait Until HTTPS Curl Succeeds From Pod
    ...    ${CLIENT_POD_NAME}    ${NAMESPACE}    https://${edge_host}/index.html    ${edge_host}:443:${router_ip}
    Wait Until HTTPS Curl Succeeds From Pod
    ...    ${CLIENT_POD_NAME}    ${NAMESPACE}    https://${reen_host}/index.html    ${reen_host}:443:${router_ip}
    Wait For Router Logs To Contain    ${routehost}
    Wait For Router Logs To Contain    ${edge_host}
    Wait For Router Logs To Contain    ${reen_host}

Verify Header Capture In Haproxy
    [Documentation]    Verify the haproxy frontend has the expected header capture configuration.
    ${router_pod}=    Get Router Pod Name
    ${haproxy}=    Run With Kubeconfig
    ...    oc exec -n ${ROUTER_NS} ${router_pod} -- grep -A 20 "frontend fe_sni" haproxy.config
    Should Contain    ${haproxy}    capture request header Host len 120
    Should Contain    ${haproxy}    capture response header Server len 120

Verify Custom Error Pages
    [Documentation]    Verify custom 503 and 404 error pages are served by the router.
    [Arguments]    ${routehost}    ${noexist_host}
    ${router_pod}=    Get Router Pod Name
    ${error503}=    Run With Kubeconfig
    ...    oc exec -n ${ROUTER_NS} ${router_pod} -- cat /var/lib/haproxy/errorfiles/error-page-503.http
    Should Contain    ${error503}    Custom:Application Unavailable
    ${error404}=    Run With Kubeconfig
    ...    oc exec -n ${ROUTER_NS} ${router_pod} -- cat /var/lib/haproxy/errorfiles/error-page-404.http
    Should Contain    ${error404}    Custom:Not Found
    ${router_ip}=    Get Router Pod IP
    ${output}=    Run With Kubeconfig
    ...    oc exec -n ${NAMESPACE} ${CLIENT_POD_NAME} -- curl http://${noexist_host} -s --resolve ${noexist_host}:80:${router_ip} --connect-timeout 10
    Should Contain    ${output}    Custom:Not Found
    Scale Deployment    ${NAMESPACE}    web-server-deploy    0
    Wait Until Keyword Succeeds    60s    2s
    ...    Custom 503 Body Should Contain    ${routehost}    ${router_ip}

Custom 503 Body Should Contain
    [Documentation]    Curl a route (GET, full response body) and assert the custom 503 error page body is returned.
    [Arguments]    ${routehost}    ${router_ip}
    ${output}=    Run With Kubeconfig
    ...    oc exec -n ${NAMESPACE} ${CLIENT_POD_NAME} -- curl http://${routehost} -s --resolve ${routehost}:80:${router_ip} --connect-timeout 10
    Should Contain    ${output}    Custom:Application Unavailable

Verify Syslog Logging
    [Documentation]    Apply syslog config and verify log delivery, then verify facility change.
    [Arguments]    ${syslog_ip}
    VAR    ${config1}=    SEPARATOR=\n
    ...    ---
    ...    ingress:
    ...    \ \ accessLogging:
    ...    \ \ \ \ destination:
    ...    \ \ \ \ \ \ syslog:
    ...    \ \ \ \ \ \ \ \ address: ${syslog_ip}
    ...    \ \ \ \ \ \ \ \ port: 514
    ...    \ \ \ \ \ \ type: Syslog
    ...    \ \ \ \ status: Enabled
    Setup Router Config And Restart    ${config1}
    VAR    ${routehost}=    route-unsec82015.${BASE_DOMAIN}
    Create OC Route    ${NAMESPACE}    http    route-http    service-unsecure    --hostname=${routehost}
    Route Should Be Admitted    route-http
    Verify Syslog Haproxy Config    ${syslog_ip}    local1
    Verify Syslog Log Delivery    ${routehost}    ${syslog_ip}
    VAR    ${config2}=    SEPARATOR=\n
    ...    ---
    ...    ingress:
    ...    \ \ accessLogging:
    ...    \ \ \ \ destination:
    ...    \ \ \ \ \ \ syslog:
    ...    \ \ \ \ \ \ \ \ address: ${syslog_ip}
    ...    \ \ \ \ \ \ \ \ port: 514
    ...    \ \ \ \ \ \ \ \ facility: local2
    ...    \ \ \ \ \ \ type: Syslog
    ...    \ \ \ \ status: Enabled
    Setup Router Config And Restart    ${config2}
    Verify Syslog Haproxy Config    ${syslog_ip}    local2

Verify Syslog Haproxy Config
    [Documentation]    Verify the syslog server address and facility in haproxy global config.
    [Arguments]    ${syslog_ip}    ${facility}
    ${haproxy}=    Read Haproxy Config
    Should Contain    ${haproxy}    log ${syslog_ip}:514 len 1024 ${facility} info

Verify Syslog Log Delivery
    [Documentation]    Curl a route and verify the log entry appears in the syslog pod.
    [Arguments]    ${routehost}    ${syslog_ip}
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

Router Logs Should Not Contain
    [Documentation]    Verify the router access logs do NOT contain a given pattern.
    [Arguments]    ${pattern}
    ${logs}=    Get Router Access Logs
    Should Not Contain    ${logs}    ${pattern}
