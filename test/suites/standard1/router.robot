*** Settings ***
Documentation       Router tests

Resource            ../../resources/common.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/ostree-health.resource
Resource            ../../resources/microshift-network.resource
Resource            ../../resources/microshift-config.resource

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           restart    slow


*** Variables ***
${NS_OWNERSHIP_1}               ${EMPTY}
${NS_OWNERSHIP_2}               ${EMPTY}
${ALTERNATIVE_HTTP_PORT}        8000
${ALTERNATIVE_HTTPS_PORT}       8001
${HOSTNAME}                     hello-microshift.cluster.local
${FAKE_LISTEN_IP}               99.99.99.99
${ROUTER_REMOVED}               SEPARATOR=\n
...                             ---
...                             ingress:
...                             \ \ status: Removed
${OWNERSHIP_ALLOW}              SEPARATOR=\n
...                             ---
...                             ingress:
...                             \ \ status: Managed
...                             \ \ routeAdmissionPolicy:
...                             \ \ \ \ namespaceOwnership: InterNamespaceAllowed
${OWNERSHIP_STRICT}             SEPARATOR=\n
...                             ---
...                             ingress:
...                             \ \ status: Managed
...                             \ \ routeAdmissionPolicy:
...                             \ \ \ \ namespaceOwnership: Strict
${ROUTER_EXPOSE_FULL}           SEPARATOR=\n
...                             ---
...                             ingress:
...                             \ \ status: Managed
...                             \ \ ports:
...                             \ \ \ \ http: ${ALTERNATIVE_HTTP_PORT}
...                             \ \ \ \ https: ${ALTERNATIVE_HTTPS_PORT}
...                             \ \ listenAddress:
...                             \ \ - br-ex
${ROUTER_TUNING_CONFIG}         SEPARATOR=\n
...                             ---
...                             ingress:
...                             \ \ defaultHTTPVersion: 2
...                             \ \ forwardedHeaderPolicy: Never
...                             \ \ httpEmptyRequestsPolicy: Ignore
...                             \ \ logEmptyRequests: Ignore
...                             \ \ httpCompression:
...                             \ \ \ \ mimeTypes:
...                             \ \ \ \ - "text/html"
...                             \ \ \ \ - "application/*"
...                             \ \ tuningOptions:
...                             \ \ \ \ headerBufferBytes: 5556
...                             \ \ \ \ headerBufferMaxRewriteBytes: 8000
...                             \ \ \ \ healthCheckInterval: 4s
...                             \ \ \ \ clientTimeout: 20s
...                             \ \ \ \ clientFinTimeout: 1.5s
...                             \ \ \ \ serverTimeout: 40s
...                             \ \ \ \ serverFinTimeout: 2s
...                             \ \ \ \ tunnelTimeout: 1h30m0s
...                             \ \ \ \ tlsInspectDelay: 6s
...                             \ \ \ \ threadCount: 3
...                             \ \ \ \ maxConnections: 60000
${ROUTER_SECURITY_CONFIG}       SEPARATOR=\n
...                             ---
...                             ingress:
...                             \ \ certificateSecret: router-certs-custom
...                             \ \ routeAdmissionPolicy:
...                             \ \ \ \ wildcardPolicy: WildcardsAllowed
...                             \ \ clientTLS:
...                             \ \ \ \ allowedSubjectPatterns: ["route-custom.apps.example.com"]
...                             \ \ \ \ clientCertificatePolicy: Required
...                             \ \ \ \ clientCA:
...                             \ \ \ \ \ \ name: router-ca-certs-custom
...                             \ \ tlsSecurityProfile:
...                             \ \ \ \ type: Custom
...                             \ \ \ \ custom:
...                             \ \ \ \ \ \ Ciphers:
...                             \ \ \ \ \ \ - ECDHE-RSA-AES256-GCM-SHA384
...                             \ \ \ \ \ \ - DHE-RSA-AES256-GCM-SHA384
...                             \ \ \ \ \ \ - TLS_CHACHA20_POLY1305_SHA256
...                             \ \ \ \ \ \ MinTLSVersion: VersionTLS13


*** Test Cases ***
Router Namespace Ownership
    [Documentation]    Test InterNamespaceAllow configuration options in
    ...    Router admission policy.
    [Setup]    Run Keywords
    ...    Setup Namespaces
    ...    Setup Hello MicroShift Pods In Multiple Namespaces

    Configure Namespace Ownership Strict
    Wait For Router Ready
    Wait Until Keyword Succeeds    60x    1s
    ...    Http Port Should Be Open    ${HTTP_PORT}
    ${result_1}=    Run Keyword And Return Status
    ...    Access Hello Microshift Success    ${HTTP_PORT}    path=/${NS_OWNERSHIP_1}
    ${result_2}=    Run Keyword And Return Status
    ...    Access Hello Microshift Success    ${HTTP_PORT}    path=/${NS_OWNERSHIP_2}
    IF    (${result_1}==True and ${result_2}==True) or (${result_1}==False and ${result_2}==False)
        Fail
    END

    Configure Namespace Ownership Allowed
    Wait For Router Ready
    Wait Until Keyword Succeeds    60x    1s
    ...    Http Port Should Be Open    ${HTTP_PORT}
    Wait Until Keyword Succeeds    60x    2s
    ...    Access Hello Microshift Success    ${HTTP_PORT}    path=/${NS_OWNERSHIP_1}
    Wait Until Keyword Succeeds    60x    2s
    ...    Access Hello Microshift Success    ${HTTP_PORT}    path=/${NS_OWNERSHIP_2}

    [Teardown]    Run Keywords
    ...    Delete Namespaces

Router Disabled
    [Documentation]    Disable the router and check the namespace does not exist.
    [Setup]    Run Keywords
    ...    Disable Router

    Oc Wait    namespace/openshift-ingress    --for=delete --timeout=300s

Router Exposure Configuration
    [Documentation]    Test custom ports and custom listening addresses.
    [Setup]    Run Keywords
    ...    Configure Router Exposure
    ...    Add Fake IP To NIC

    Wait Until Keyword Succeeds    60x    2s
    ...    Internal Router Port Should Be Open    10.44.0.0    ${ALTERNATIVE_HTTP_PORT}    http
    Wait Until Keyword Succeeds    60x    2s
    ...    Internal Router Port Should Be Open    10.44.0.0    ${ALTERNATIVE_HTTPS_PORT}    https
    # The link in which this IP was added was configured in MicroShift. Note the IP was
    # added after MicroShift started, therefore it must pick it up dynamically.
    Wait Until Keyword Succeeds    60x    2s
    ...    Internal Router Port Should Be Open    ${FAKE_LISTEN_IP}    ${ALTERNATIVE_HTTP_PORT}    http

    [Teardown]    Run Keywords
    ...    Remove Fake IP From NIC

Router Verify Tuning Configuration
    [Documentation]    Test ingress tuning configuration.
    [Setup]    Setup With Custom Config    ${ROUTER_TUNING_CONFIG}
    Wait For Router Ready
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
    ...    Setup With Custom Config    ${ROUTER_SECURITY_CONFIG}
    ...    AND
    ...    Create Custom Resources
    Wait For Router Ready
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


*** Keywords ***
Setup
    [Documentation]    Special setup for the suite. As every test case configures MicroShift in
    ...    different ways there is no need to restart before/after each one of them. Instead, store
    ...    the original configuration here to restore it at the end.
    Setup Suite With Namespace

Teardown
    [Documentation]    Special teardown for the suite, will finish off by restoring the original
    ...    configuration and restarting MicroShift.
    Remove Drop In MicroShift Config    10-ingress
    Restart MicroShift
    Teardown Suite With Namespace

Configure Namespace Ownership Allowed
    [Documentation]    Configure MicroShift to use InterNamespaceAllowed namespace ownership.
    Setup With Custom Config    ${OWNERSHIP_ALLOW}

Configure Namespace Ownership Strict
    [Documentation]    Configure MicroShift to use Strict namespace ownership.
    Setup With Custom Config    ${OWNERSHIP_STRICT}

Configure Router Exposure
    [Documentation]    Configure MicroShift to use Strict namespace ownership.
    Setup With Custom Config    ${ROUTER_EXPOSE_FULL}

Disable Router
    [Documentation]    Disable router.
    Setup With Custom Config    ${ROUTER_REMOVED}

Wait For Router Ready
    [Documentation]    Wait for the default router to be ready.
    # Wait for the namespace to be ready, as sometimes apiserver may signal readiness before all
    # the manifests have been applied.

    ${fail}=    Set Variable    ${FALSE}
    TRY
        Oc Wait    namespace/openshift-ingress    --for jsonpath='{.status.phase}=Active' --timeout=5m
        Named Deployment Should Be Available    router-default    openshift-ingress    5m
    EXCEPT
        ${fail}=    Set Variable    ${TRUE}
    END

    Oc Logs    deployment/router-default    openshift-ingress
    IF    ${fail}    Fail    router did not become ready

Setup With Custom Config
    [Documentation]    Install a custom config and restart MicroShift.
    [Arguments]    ${config_content}
    Drop In MicroShift Config    ${config_content}    10-ingress
    Restart MicroShift

Setup Namespaces
    [Documentation]    Configure the required namespaces for namespace ownership tests.
    Set Suite Variable    \${NS_OWNERSHIP_1}    ${NAMESPACE}-ownership-1
    Set Suite Variable    \${NS_OWNERSHIP_2}    ${NAMESPACE}-ownership-2
    Create Namespace    ${NS_OWNERSHIP_1}
    Create Namespace    ${NS_OWNERSHIP_2}

Delete Namespaces
    [Documentation]    Remove namespace ownership namespaces.
    Remove Namespace    ${NS_OWNERSHIP_1}
    Remove Namespace    ${NS_OWNERSHIP_2}

Setup Hello MicroShift Pods In Multiple Namespaces
    [Documentation]    Create and expose hello microshift pods in two namespaces.
    Create Hello MicroShift Pod    ns=${NS_OWNERSHIP_1}
    Create Hello MicroShift Pod    ns=${NS_OWNERSHIP_2}
    Expose Hello MicroShift    ${NS_OWNERSHIP_1}
    Expose Hello MicroShift    ${NS_OWNERSHIP_2}
    Oc Expose    svc hello-microshift --hostname ${HOSTNAME} --path /${NS_OWNERSHIP_1} -n ${NS_OWNERSHIP_1}
    Oc Expose    svc hello-microshift --hostname ${HOSTNAME} --path /${NS_OWNERSHIP_2} -n ${NS_OWNERSHIP_2}

Http Port Should Be Open
    [Documentation]    Connect to the router and expect a response using http. A 503 response means the router
    ...    is up but no routes are configured for the requested path.
    [Arguments]    ${port}
    Access Hello MicroShift No Route    ${port}

Port Should Be Closed
    [Documentation]    Try to connect to the router and expect a failure when connecting.
    [Arguments]    ${port}
    ${rc}    ${ignore_out}    ${ignore_err}=    Access Hello MicroShift    ${port}
    # 7 is the error code for connection refused when using curl.
    Should Be Equal As Integers    ${rc}    7

Internal Router Port Should Be Open
    [Documentation]    Test whether the given router port is open from within MicroShift's host
    ...    using the given IP address.
    [Arguments]    ${router_ip}    ${port}    ${scheme}=http
    ${stdout}    ${stderr}    ${rc}=    SSHLibrary.Execute Command
    ...    curl -I -k ${scheme}://${router_ip}:${port}
    ...    sudo=False    return_rc=True    return_stderr=True    return_stdout=True
    Log Many    ${stdout}    ${stderr}
    Should Be Equal As Integers    0    ${rc}
    Should Match Regexp    ${stdout}    HTTP.*503

Add Fake IP To NIC
    [Documentation]    Add the given IP to the given NIC temporarily.
    [Arguments]    ${ip_address}=${FAKE_LISTEN_IP}    ${nic_name}=br-ex
    ${stdout}    ${stderr}    ${rc}=    SSHLibrary.Execute Command
    ...    ip address add ${ip_address}/32 dev ${nic_name}
    ...    sudo=True    return_rc=True    return_stderr=True    return_stdout=True
    Log Many    ${stdout}    ${stderr}
    Should Be Equal As Integers    0    ${rc}

Remove Fake IP From NIC
    [Documentation]    Remove the given IP from the given NIC.
    [Arguments]    ${ip_address}=${FAKE_LISTEN_IP}    ${nic_name}=br-ex
    ${stdout}    ${stderr}    ${rc}=    SSHLibrary.Execute Command
    ...    ip address delete ${ip_address}/32 dev ${nic_name}
    ...    sudo=True    return_rc=True    return_stderr=True    return_stdout=True
    Log Many    ${stdout}    ${stderr}
    Should Be Equal As Integers    0    ${rc}

Pod Environment Should Match Value
    [Documentation]    Check if config is Matching
    [Arguments]    ${name_space}    ${env_name}    ${expected_value}
    ${is}=    Oc Get JsonPath
    ...    pod
    ...    ${name_space}
    ...    ${EMPTY}
    ...    .items[*].spec.containers[*].env[?(@.name=="${env_name}")].value
    Should Be Equal As Strings    ${is}    ${expected_value}

Pod Volume Should Contain Secret
    [Documentation]    Check if pod volume exists by Name
    [Arguments]    ${name_space}    ${volume_name}    ${expected_value}
    ${is}=    Oc Get JsonPath
    ...    pod
    ...    ${name_space}
    ...    ${EMPTY}
    ...    .items[*].spec.volumes[?(@.name=="${volume_name}")].secret.secretName
    Should Be Equal As Strings    ${is}    ${expected_value}

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
