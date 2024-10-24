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
${ROUTER_TUNING_CONFIG}           SEPARATOR=\n
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


*** Test Cases ***

Router Verify Tuning Configuration
    [Documentation]    Test ingress tuning configuration.
    Configure Router Tuning
    Wait For Router Ready
    Pod Environment Should Match Value    openshift-ingress    ROUTER_BUF_SIZE    5556
    Pod Environment Should Match Value    openshift-ingress    ROUTER_MAX_REWRITE_SIZE    8000    
    Pod Environment Should Match Value    openshift-ingress    ROUTER_BACKEND_CHECK_INTERVAL    4s        
    Pod Environment Should Match Value    openshift-ingress    ROUTER_DEFAULT_CLIENT_TIMEOUT    20s    
    Pod Environment Should Match Value    openshift-ingress    ROUTER_CLIENT_FIN_TIMEOUT    1.5s
    Pod Environment Should Match Value    openshift-ingress    ROUTER_DEFAULT_SERVER_TIMEOUT    40s
    Pod Environment Should Match Value    openshift-ingress    ROUTER_DEFAULT_SERVER_FIN_TIMEOUT    2s
    Pod Environment Should Match Value    openshift-ingress    ROUTER_DEFAULT_TUNNEL_TIMEOUT    1h30m0s
    Pod Environment Should Match Value    openshift-ingress    ROUTER_INSPECT_DELAY    6s
    Pod Environment Should Match Value    openshift-ingress    ROUTER_THREADS    3
    Pod Environment Should Match Value    openshift-ingress    ROUTER_MAX_CONNECTIONS    60000    
    Pod Environment Should Match Value    openshift-ingress    ROUTER_SET_FORWARDED_HEADERS    Never
    Pod Environment Should Match Value    openshift-ingress    ROUTER_HTTP_IGNORE_PROBES    Ignore
    Pod Environment Should Match Value    openshift-ingress    ROUTER_DONT_LOG_NULL    Ignore
    Pod Environment Should Match Value    openshift-ingress    ROUTER_ENABLE_COMPRESSION    true
    Pod Environment Should Match Value    openshift-ingress    ROUTER_COMPRESSION_MIME    text/html application/*
    Pod Environment Should Match Value    openshift-ingress    ROUTER_DISABLE_HTTP2    false


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

Configure Router Tuning
    [Documentation]    Configure MicroShift to use Strict namespace ownership.
    Setup With Custom Config    ${ROUTER_TUNING_CONFIG}

Wait For Router Ready
    [Documentation]    Wait for the default router to be ready.
    # Wait for the namespace to be ready, as sometimes apiserver may signal readiness before all
    # the manifests have been applied.
    Run With Kubeconfig    oc wait --for jsonpath='{.status.phase}=Active' --timeout=5m namespace/openshift-ingress
    Named Deployment Should Be Available    router-default    openshift-ingress    5m

Setup With Custom Config
    [Documentation]    Install a custom config and restart MicroShift.
    [Arguments]    ${config_content}
    Drop In MicroShift Config    ${config_content}    10-ingress
    Restart MicroShift

Pod Environment Should Match Value
    [Documentation]    Check if config is Matching
    [Arguments]    ${name_space}    ${env_name}    ${expected_value}
    ${is}=    Oc Get JsonPath    pod    ${name_space}    ${EMPTY}    .items[*].spec.containers[*].env[?(@.name=="${env_name}")].value
    Should Be Equal As Strings    ${is}    ${expected_value}