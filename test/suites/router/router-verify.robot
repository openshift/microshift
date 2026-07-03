*** Settings ***
Documentation       Router configuration verification tests

Resource            router.resource

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           restart    slow


*** Variables ***
${ROUTER_TUNING_CONFIG}                     SEPARATOR=\n
...                                         ---
...                                         ingress:
...                                         \ \ defaultHTTPVersion: 2
...                                         \ \ forwardedHeaderPolicy: Never
...                                         \ \ httpEmptyRequestsPolicy: Ignore
...                                         \ \ logEmptyRequests: Ignore
...                                         \ \ httpCompression:
...                                         \ \ \ \ mimeTypes:
...                                         \ \ \ \ - "text/html"
...                                         \ \ \ \ - "application/*"
...                                         \ \ tuningOptions:
...                                         \ \ \ \ headerBufferBytes: 5556
...                                         \ \ \ \ headerBufferMaxRewriteBytes: 8000
...                                         \ \ \ \ healthCheckInterval: 4s
...                                         \ \ \ \ clientTimeout: 20s
...                                         \ \ \ \ clientFinTimeout: 1.5s
...                                         \ \ \ \ serverTimeout: 40s
...                                         \ \ \ \ serverFinTimeout: 2s
...                                         \ \ \ \ tunnelTimeout: 1h30m0s
...                                         \ \ \ \ tlsInspectDelay: 6s
...                                         \ \ \ \ threadCount: 3
...                                         \ \ \ \ maxConnections: 60000
${ROUTER_SECURITY_CONFIG}                   SEPARATOR=\n
...                                         ---
...                                         ingress:
...                                         \ \ certificateSecret: router-certs-custom
...                                         \ \ routeAdmissionPolicy:
...                                         \ \ \ \ wildcardPolicy: WildcardsAllowed
...                                         \ \ clientTLS:
...                                         \ \ \ \ allowedSubjectPatterns: ["route-custom.apps.example.com"]
...                                         \ \ \ \ clientCertificatePolicy: Required
...                                         \ \ \ \ clientCA:
...                                         \ \ \ \ \ \ name: router-ca-certs-custom
...                                         \ \ tlsSecurityProfile:
...                                         \ \ \ \ type: Custom
...                                         \ \ \ \ custom:
...                                         \ \ \ \ \ \ Ciphers:
...                                         \ \ \ \ \ \ - ECDHE-RSA-AES256-GCM-SHA384
...                                         \ \ \ \ \ \ - DHE-RSA-AES256-GCM-SHA384
...                                         \ \ \ \ \ \ - TLS_CHACHA20_POLY1305_SHA256
...                                         \ \ \ \ \ \ MinTLSVersion: VersionTLS13
${ROUTER_ACCESS_LOGGING_CONFIG}             SEPARATOR=\n
...                                         ---
...                                         ingress:
...                                         \ \ accessLogging:
...                                         \ \ \ \ status: Enabled
...                                         \ \ \ \ destination:
...                                         \ \ \ \ \ \ type: Container
...                                         \ \ \ \ \ \ container:
...                                         \ \ \ \ \ \ \ \ maxLength: 2000
...                                         \ \ \ \ httpCaptureCookies:
...                                         \ \ \ \ - matchType: Exact
...                                         \ \ \ \ \ \ maxLength: 20
...                                         \ \ \ \ \ \ name: cookie
...                                         \ \ \ \ httpCaptureHeaders:
...                                         \ \ \ \ \ \ request:
...                                         \ \ \ \ \ \ - maxLength: 11
...                                         \ \ \ \ \ \ \ \ name: header1
...                                         \ \ \ \ \ \ response:
...                                         \ \ \ \ \ \ - maxLength: 12
...                                         \ \ \ \ \ \ \ \ name: header2
...                                         \ \ \ \ httpLogFormat: some-format
...                                         \ \ httpErrorCodePages:
...                                         \ \ \ \ name: router-error-pages
${ROUTER_ACCESS_LOGGING_CONFIG_SYSLOG}      SEPARATOR=\n
...                                         ---
...                                         ingress:
...                                         \ \ accessLogging:
...                                         \ \ \ \ status: Enabled
...                                         \ \ \ \ destination:
...                                         \ \ \ \ \ \ type: Syslog
...                                         \ \ \ \ \ \ syslog:
...                                         \ \ \ \ \ \ \ \ address: 1.2.3.4
...                                         \ \ \ \ \ \ \ \ port: 9000
...                                         \ \ \ \ \ \ \ \ facility: local7
...                                         \ \ \ \ \ \ \ \ maxLength: 1000

${ROUTER_ERROR_CODE_CONFIGMAP}              assets/network/router-error-configmap.yaml


*** Test Cases ***
Router Verify Tuning Configuration
    [Documentation]    Test ingress tuning configuration.
    [Setup]    Run Keywords
    ...    Setup With Custom Config    ${ROUTER_TUNING_CONFIG}
    ...    AND
    ...    Wait For Router Ready

    Pod Environment Should Match Value    openshift-ingress    ROUTER_BUF_SIZE    5556
    Pod Environment Should Match Value    openshift-ingress    ROUTER_MAX_REWRITE_SIZE    8000
    Pod Environment Should Match Value    openshift-ingress    ROUTER_BACKEND_CHECK_INTERVAL    4s
    Pod Environment Should Match Value    openshift-ingress    ROUTER_DEFAULT_CLIENT_TIMEOUT    20s
    Pod Environment Should Match Value    openshift-ingress    ROUTER_CLIENT_FIN_TIMEOUT    1500ms
    Pod Environment Should Match Value    openshift-ingress    ROUTER_DEFAULT_SERVER_TIMEOUT    40s
    Pod Environment Should Match Value    openshift-ingress    ROUTER_DEFAULT_SERVER_FIN_TIMEOUT    2s
    Pod Environment Should Match Value    openshift-ingress    ROUTER_DEFAULT_TUNNEL_TIMEOUT    90m
    Pod Environment Should Match Value    openshift-ingress    ROUTER_INSPECT_DELAY    6s
    Pod Environment Should Match Value    openshift-ingress    ROUTER_THREADS    3
    Pod Environment Should Match Value    openshift-ingress    ROUTER_MAX_CONNECTIONS    60000
    Pod Environment Should Match Value    openshift-ingress    ROUTER_SET_FORWARDED_HEADERS    Never
    Pod Environment Should Match Value    openshift-ingress    ROUTER_HTTP_IGNORE_PROBES    true
    Pod Environment Should Match Value    openshift-ingress    ROUTER_DONT_LOG_NULL    true
    Pod Environment Should Match Value    openshift-ingress    ROUTER_ENABLE_COMPRESSION    true
    Pod Environment Should Match Value    openshift-ingress    ROUTER_COMPRESSION_MIME    text/html application/*
    Pod Environment Should Match Value    openshift-ingress    ROUTER_DISABLE_HTTP2    false

Router Verify Security Configuration
    [Documentation]    Test ingress security configuration.
    [Setup]    Run Keywords
    ...    Create Custom Resources
    ...    AND
    ...    Setup With Custom Config    ${ROUTER_SECURITY_CONFIG}
    ...    AND
    ...    Wait For Router Ready

    Pod Environment Should Match Value    openshift-ingress    ROUTER_ALLOW_WILDCARD_ROUTES    true
    Pod Environment Should Match Value    openshift-ingress    ROUTER_MUTUAL_TLS_AUTH    required
    Pod Environment Should Match Value
    ...    openshift-ingress
    ...    ROUTER_MUTUAL_TLS_AUTH_CA
    ...    /etc/pki/tls/client-ca/ca-bundle.pem
    Pod Environment Should Match Value
    ...    openshift-ingress
    ...    ROUTER_CIPHERS
    ...    ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384
    Pod Environment Should Match Value    openshift-ingress    ROUTER_CIPHERSUITES    TLS_CHACHA20_POLY1305_SHA256
    Pod Environment Should Match Value    openshift-ingress    SSL_MIN_VERSION    TLSv1.3
    Pod Environment Should Match Value
    ...    openshift-ingress
    ...    ROUTER_MUTUAL_TLS_AUTH_FILTER
    ...    (?:route-custom.apps.example.com)
    Pod Volume Should Contain Secret    openshift-ingress    default-certificate    router-certs-custom

    [Teardown]    Delete Custom CA Secret

Router Verify Access Logging Configuration Container
    [Documentation]    Test ingress access logging configuration.
    [Setup]    Run Keywords
    ...    Remove Custom Config
    ...    AND
    ...    Restart MicroShift
    ...    AND
    ...    Oc Apply    -f ${ROUTER_ERROR_CODE_CONFIGMAP}
    ...    AND
    ...    Setup With Custom Config    ${ROUTER_ACCESS_LOGGING_CONFIG}
    ...    AND
    ...    Wait For Router Ready

    Pod Environment Should Match Value    openshift-ingress    ROUTER_SYSLOG_ADDRESS    /var/lib/rsyslog/rsyslog.sock
    Pod Environment Should Match Value    openshift-ingress    ROUTER_LOG_LEVEL    info
    Pod Environment Should Match Value    openshift-ingress    ROUTER_LOG_MAX_LENGTH    2000
    Pod Environment Should Match Value    openshift-ingress    ROUTER_SYSLOG_FORMAT    "some-format"
    Pod Environment Should Match Value    openshift-ingress    ROUTER_CAPTURE_HTTP_REQUEST_HEADERS    header1:11
    Pod Environment Should Match Value    openshift-ingress    ROUTER_CAPTURE_HTTP_RESPONSE_HEADERS    header2:12
    Pod Environment Should Match Value    openshift-ingress    ROUTER_CAPTURE_HTTP_COOKIE    cookie=:20
    Pod Environment Should Match Value
    ...    openshift-ingress
    ...    ROUTER_ERRORFILE_503
    ...    /var/lib/haproxy/errorfiles/error-page-503.http
    Pod Environment Should Match Value
    ...    openshift-ingress
    ...    ROUTER_ERRORFILE_404
    ...    /var/lib/haproxy/errorfiles/error-page-404.http
    Check Access Logs    some-format

    [Teardown]    Oc Delete    -f ${ROUTER_ERROR_CODE_CONFIGMAP}

Router Verify Access Logging Configuration Syslog
    [Documentation]    Test ingress access logging configuration.
    [Setup]    Run Keywords
    ...    Setup With Custom Config    ${ROUTER_ACCESS_LOGGING_CONFIG_SYSLOG}
    ...    AND
    ...    Wait For Router Ready

    Pod Environment Should Match Value    openshift-ingress    ROUTER_SYSLOG_ADDRESS    1.2.3.4:9000
    Pod Environment Should Match Value    openshift-ingress    ROUTER_LOG_LEVEL    info
    Pod Environment Should Match Value    openshift-ingress    ROUTER_LOG_MAX_LENGTH    1000
    Pod Environment Should Match Value    openshift-ingress    ROUTER_LOG_FACILITY    local7


*** Keywords ***
Create Custom Resources
    [Documentation]    Copy Default certs to custom
    Run With Kubeconfig
    ...    oc get secret router-certs-default -n openshift-ingress -oyaml | sed 's/name: .*/name: router-certs-custom/' | oc create -f - -oyaml | true
    Run With Kubeconfig    oc extract configmap/openshift-service-ca.crt --to=/tmp/ --confirm
    Run With Kubeconfig
    ...    oc create configmap router-ca-certs-custom -n openshift-ingress --from-file=ca-bundle.pem=/tmp/service-ca.crt --dry-run -o yaml | oc apply -f -

Delete Custom CA Secret
    [Documentation]    Copy Default certs to custom
    Oc Delete    secret/router-certs-custom -n openshift-ingress
    Oc Delete    configmap/router-ca-certs-custom -n openshift-ingress

Check Access Logs
    [Documentation]    Retrieve and check if a pattern appears in the router's access logs.
    [Arguments]    ${pattern}
    ${logs}=    Oc Logs    deployment/router-default -c logs    openshift-ingress
    Should Contain    ${logs}    ${pattern}
