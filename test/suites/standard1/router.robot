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

    Run With Kubeconfig    oc wait --for=delete namespace/openshift-ingress --timeout=60s

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


*** Keywords ***
Setup
    [Documentation]    Special setup for the suite. As every test case configures MicroShift in
    ...    different ways there is no need to restart before/after each one of them. Instead, store
    ...    the original configuration here to restore it at the end.
    Setup Suite With Namespace
    Save Default MicroShift Config

Teardown
    [Documentation]    Special teardown for the suite, will finish off by restoring the original
    ...    configuration and restarting MicroShift.
    Restore Default MicroShift Config
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
    Run With Kubeconfig    oc wait --for jsonpath='{.status.phase}=Active' --timeout=5m namespace/openshift-ingress
    Named Deployment Should Be Available    router-default    openshift-ingress    5m

Setup With Custom Config
    [Documentation]    Install a custom config and restart MicroShift.
    [Arguments]    ${config_content}
    ${merged}=    Extend MicroShift Config    ${config_content}
    Upload MicroShift Config    ${merged}
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
