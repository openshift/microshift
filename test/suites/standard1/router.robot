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
${NS_OWNERSHIP_1}       ${EMPTY}
${NS_OWNERSHIP_2}       ${EMPTY}
${HOSTNAME}             hello-microshift.cluster.local
${ROUTER_MANAGED}       SEPARATOR=\n
...                     ---
...                     ingress:
...                     \ \ status: Managed
${ROUTER_REMOVED}       SEPARATOR=\n
...                     ---
...                     ingress:
...                     \ \ status: Removed
${OWNERSHIP_ALLOW}      SEPARATOR=\n
...                     ---
...                     ingress:
...                     \ \ status: Managed
...                     \ \ routeAdmissionPolicy:
...                     \ \ \ \ namespaceOwnership: InterNamespaceAllowed
${OWNERSHIP_STRICT}     SEPARATOR=\n
...                     ---
...                     ingress:
...                     \ \ status: Managed
...                     \ \ routeAdmissionPolicy:
...                     \ \ \ \ namespaceOwnership: Strict


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
    ...    Access Hello Microshift Success    ${HTTP_PORT}    path=/${NS_OWNERSHIP_1}    ns=${NS_OWNERSHIP_1}

    Wait Until Keyword Succeeds    10x    6s
    ...    Access Hello Microshift Success    ${HTTP_PORT}    path=/${NS_OWNERSHIP_2}    ns=${NS_OWNERSHIP_2}

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
    ...    Access Hello Microshift Success    ${HTTP_PORT}    path=/${NS_OWNERSHIP_1}    ns=${NS_OWNERSHIP_1}
    ${result_2}=    Run Keyword And Return Status
    ...    Access Hello Microshift Success    ${HTTP_PORT}    path=/${NS_OWNERSHIP_2}    ns=${NS_OWNERSHIP_2}

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
    ...    Create Hello MicroShift Pod
    ...    Expose Hello MicroShift Service Via Route
    ...    Restart Router

    Wait Until Keyword Succeeds    10x    6s
    ...    Access Hello Microshift    ${HTTP_PORT}

    [Teardown]    Run Keywords
    ...    Delete Hello MicroShift Route
    ...    Delete Hello MicroShift Pod And Service
    ...    Wait For Service Deletion With Timeout
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

Expose Hello MicroShift Service Via Route
    [Documentation]    Expose the "hello microshift" application through the Route
    Oc Expose    pod hello-microshift -n ${NAMESPACE}
    Oc Expose    svc hello-microshift --hostname hello-microshift.cluster.local -n ${NAMESPACE}

Delete Hello MicroShift Route
    [Documentation]    Delete route for cleanup.
    Oc Delete    route/hello-microshift -n ${NAMESPACE}

Wait For Service Deletion With Timeout
    [Documentation]    Polls for service and endpoint by "app=hello-microshift" label. Fails if timeout
    ...    expires. This check is unique to this test suite because each test here reuses the same namespace. Since
    ...    the tests reuse the service name, a small race window exists between the teardown of one test and the setup
    ...    of the next. This produces flakey failures when the service or endpoint names collide.
    Wait Until Keyword Succeeds    30s    1s
    ...    Network APIs With Test Label Are Gone

Network APIs With Test Label Are Gone
    [Documentation]    Check for service and endpoint by "app=hello-microshift" label. Succeeds if response matches
    ...    "No resources found in <namespace> namespace." Fail if not.
    ${match_string}=    Catenate    No resources found in    ${NAMESPACE}    namespace.
    ${match_string}=    Remove String    ${match_string}    "
    ${response}=    Run With Kubeconfig    oc get svc,ep -l app\=hello-microshift -n ${NAMESPACE}
    Should Be Equal As Strings    ${match_string}    ${response}    strip_spaces=True

Disable Router
    [Documentation]    Disable router
    Setup With Custom Config    ${ROUTER_REMOVED}

Enable Router
    [Documentation]    Disable router
    Setup With Custom Config    ${ROUTER_MANAGED}

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
