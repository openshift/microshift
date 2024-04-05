*** Settings ***
Documentation       Router tests

Resource            ../../resources/common.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/ostree-health.resource
Resource            ../../resources/microshift-network.resource
Resource            ../../resources/microshift-config.resource

Suite Setup         Setup Suite With Namespace
Suite Teardown      Teardown Suite With Namespace

Test Tags           restart    slow


*** Variables ***
${NS_OWNERSHIP_1}               ${EMPTY}
${NS_OWNERSHIP_2}               ${EMPTY}
${ALTERNATIVE_HTTP_PORT}        8000
${ALTERNATIVE_HTTPS_PORT}       8001
${HOSTNAME}                     hello-microshift.cluster.local
${ROUTER_MANAGED}               SEPARATOR=\n
...                             ---
...                             ingress:
...                             \ \ status: Managed
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
${ROUTER_CUSTOM_PORTS}          SEPARATOR=\n
...                             ---
...                             ingress:
...                             \ \ status: Managed
...                             \ \ ports:
...                             \ \ \ \ http: ${ALTERNATIVE_HTTP_PORT}
...                             \ \ \ \ https: ${ALTERNATIVE_HTTPS_PORT}
${ROUTER_LISTEN_INTERNAL}       SEPARATOR=\n
...                             ---
...                             ingress:
...                             \ \ status: Managed
...                             \ \ listenAddress:
...                             \ \ - 10.44.0.0


*** Test Cases ***
Router Namespace Ownership Allowed
    [Documentation]    Test InterNamespaceAllow configuration option in
    ...    Router admission policy.
    [Setup]    Run Keywords
    ...    Save Default MicroShift Config
    ...    Configure Namespace Ownership Allowed
    ...    Setup Namespaces
    ...    Setup Hello MicroShift Pods In Multiple Namespaces
    ...    Restart Router

    Wait Until Keyword Succeeds    10x    6s
    ...    Access Hello Microshift Success    ${HTTP_PORT}    path=/${NS_OWNERSHIP_1}

    Wait Until Keyword Succeeds    10x    6s
    ...    Access Hello Microshift Success    ${HTTP_PORT}    path=/${NS_OWNERSHIP_2}

    [Teardown]    Run Keywords
    ...    Delete Namespaces
    ...    Restore Default MicroShift Config
    ...    Restart MicroShift

Router Namespace Ownership Strict
    [Documentation]    Test Strict configuration option in Router
    ...    admission policy.
    [Setup]    Run Keywords
    ...    Save Default MicroShift Config
    ...    Configure Namespace Ownership Strict
    ...    Setup Namespaces
    ...    Setup Hello MicroShift Pods In Multiple Namespaces
    ...    Restart Router

    ${result_1}=    Run Keyword And Return Status
    ...    Access Hello Microshift Success    ${HTTP_PORT}    path=/${NS_OWNERSHIP_1}
    ${result_2}=    Run Keyword And Return Status
    ...    Access Hello Microshift Success    ${HTTP_PORT}    path=/${NS_OWNERSHIP_2}

    IF    (${result_1}==True and ${result_2}==True) or (${result_1}==False and ${result_2}==False)
        Fail
    END

    [Teardown]    Run Keywords
    ...    Delete Namespaces
    ...    Restore Default MicroShift Config
    ...    Restart MicroShift

Router Enabled
    [Documentation]    Check default configuration, router enabled and standard ports and expose.
    [Setup]    Run Keywords
    ...    Save Default MicroShift Config
    ...    Enable Router
    ...    Restart Router

    Wait Until Keyword Succeeds    10x    6s
    ...    Access Hello MicroShift No Route    ${HTTP_PORT}

    [Teardown]    Run Keywords
    ...    Restore Default MicroShift Config
    ...    Restart MicroShift

Router Disabled
    [Documentation]    Disable the router and check the namespace does not exist.
    [Setup]    Run Keywords
    ...    Save Default MicroShift Config
    ...    Disable Router

    Run With Kubeconfig    oc wait --for=delete namespace/openshift-ingress --timeout=60s

    [Teardown]    Run Keywords
    ...    Restore Default MicroShift Config
    ...    Restart MicroShift

Router Listen Custom Ports
    [Documentation]    Change default listening ports in the router and check the router is listening. This test
    ...    only checks connectivity, it does not go into router internals such as where traffic lands inside the
    ...    cluster.
    [Setup]    Run Keywords
    ...    Save Default MicroShift Config

    Port Should Be Closed    ${ALTERNATIVE_HTTP_PORT}
    Port Should Be Closed    ${ALTERNATIVE_HTTPS_PORT}
    Configure Listening Ports
    Restart Router
    Http Port Should Be Open    ${ALTERNATIVE_HTTP_PORT}
    Https Port Should Be Open    ${ALTERNATIVE_HTTPS_PORT}

    [Teardown]    Run Keywords
    ...    Restore Default MicroShift Config
    ...    Restart MicroShift

Router Listen In Internal Addresses Only
    [Documentation]    Configure the default router to respond only to an internal IP. The node IP
    ...    must reject connections (which means external router connection attempts), while the
    ...    internal IP will accept them and serve content.
    [Setup]    Run Keywords
    ...    Save Default MicroShift Config
    ...    Expose Router Internally
    ...    Restart Router

    # Following two keywords try to connect using the node IP.
    Port Should Be Closed    80
    Port Should Be Closed    443
    Internal Router Access Success    10.44.0.0

    [Teardown]    Run Keywords
    ...    Restore Default MicroShift Config
    ...    Restart MicroShift


*** Keywords ***
Configure Namespace Ownership Allowed
    [Documentation]    Configure MicroShift to use InterNamespaceAllowed namespace ownership
    Setup With Custom Config    ${OWNERSHIP_ALLOW}

Configure Namespace Ownership Strict
    [Documentation]    Configure MicroShift to use Strict namespace ownership
    Setup With Custom Config    ${OWNERSHIP_STRICT}

Restart Router
    [Documentation]    Restart the router and wait for readiness again. The router is sensitive to apiserver
    ...    downtime and might need a restart (after the apiserver is ready) to resync all the routes.
    Run With Kubeconfig    oc rollout restart deployment router-default -n openshift-ingress
    Named Deployment Should Be Available    router-default    openshift-ingress    5m

Disable Router
    [Documentation]    Disable router
    Setup With Custom Config    ${ROUTER_REMOVED}

Enable Router
    [Documentation]    Disable router
    Setup With Custom Config    ${ROUTER_MANAGED}

Configure Listening Ports
    [Documentation]    Enable router and change the default listening ports
    Setup With Custom Config    ${ROUTER_CUSTOM_PORTS}

Expose Router Internally
    [Documentation]    Configure the router to get exposed internally and block
    ...    all external connections.
    Setup With Custom Config    ${ROUTER_LISTEN_INTERNAL}

Setup With Custom Config
    [Documentation]    Install a custom config and restart MicroShift
    [Arguments]    ${config_content}
    ${merged}=    Extend MicroShift Config    ${config_content}
    Upload MicroShift Config    ${merged}
    Restart MicroShift

Setup Namespaces
    [Documentation]    Configure the required namespaces for namespace ownership tests
    Set Suite Variable    \${NS_OWNERSHIP_1}    ${NAMESPACE}-ownership-1
    Set Suite Variable    \${NS_OWNERSHIP_2}    ${NAMESPACE}-ownership-2
    Create Namespace    ${NS_OWNERSHIP_1}
    Create Namespace    ${NS_OWNERSHIP_2}

Delete Namespaces
    [Documentation]    Remove namespace ownership namespaces
    Remove Namespace    ${NS_OWNERSHIP_1}
    Remove Namespace    ${NS_OWNERSHIP_2}

Setup Hello MicroShift Pods In Multiple Namespaces
    [Documentation]    Create and expose hello microshift pods in two namespaces
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

Https Port Should Be Open
    [Documentation]    Connect to the router and expect a response using https. A 503 response means the router
    ...    is up but no routes are configured for the requested path.
    [Arguments]    ${port}
    Access Hello MicroShift No Route    ${port}    scheme=https

Port Should Be Closed
    [Documentation]    Try to connect to the router and expect a failure when connecting.
    [Arguments]    ${port}
    ${rc}    ${ignore_out}    ${ignore_err}=    Access Hello MicroShift    ${port}
    # 7 is the error code for connection refused when using curl.
    Should Be Equal As Integers    ${rc}    7

Internal Router Access Success
    [Documentation]    Connect and send a request to the given IP (where router should
    ...    be listening) from within the MicroShift host. Expects a 503 assuming there
    ...    are no routes configured.
    [Arguments]    ${router_ip}
    ${stdout}    ${stderr}    ${rc}=    SSHLibrary.Execute Command
    ...    curl -I http://${router_ip}
    ...    sudo=False    return_rc=True    return_stderr=True    return_stdout=True
    Should Match Regexp    ${stdout}    HTTP.*503
    Log Many    ${stdout}    ${stderr}
    Should Be Equal As Integers    0    ${rc}
