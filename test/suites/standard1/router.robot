*** Settings ***
Documentation       Router configuration tests
Resource            ../../resources/common.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/ostree-health.resource
Resource            ../../resources/microshift-network.resource
Resource            ../../resources/microshift-config.resource

Suite Setup         Setup Suite With Namespace
Suite Teardown      Teardown Suite With Namespace

Test Tags           restart    slow


*** Variables ***
${NS_OWNERSHIP_1}
${NS_OWNERSHIP_2}
${HOSTNAME}    hello-microshift.cluster.local
${OWNERSHIP_ALLOW}      SEPARATOR=\n
...                     ---
...                     ingress:
...                     \ \ routeAdmissionPolicy:
...                     \ \ \ \ namespaceOwnership: InterNamespaceAllowed
${OWNERSHIP_STRICT}     SEPARATOR=\n
...                     ---
...                     ingress:
...                     \ \ routeAdmissionPolicy:
...                     \ \ \ \ namespaceOwnership: Strict

*** Test Cases ***
Router Namespace Ownership Allowed
    [Documentation]    Test InterNamespaceAllow configuration option in
    ...    Router admission policy.
    [Setup]    Run Keywords
    ...    Save Default MicroShift Config
    ...    Configure Namespace Ownership Allowed
    ...    Restart MicroShift
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
    ...    Restart MicroShift
    ...    Setup Namespaces
    ...    Setup Hello MicroShift Pods In Multiple Namespaces
    ...    Restart Router

    ${result_1}=    Run Keyword And Return Status    Access Hello Microshift Success    ${HTTP_PORT}    path=/${NS_OWNERSHIP_1}    ns=${NS_OWNERSHIP_1}
    ${result_2}=    Run Keyword And Return Status    Access Hello Microshift Success    ${HTTP_PORT}    path=/${NS_OWNERSHIP_2}    ns=${NS_OWNERSHIP_2}

    Run Keyword If    (${result_1}==True and ${result_2}==True) or (${result_1}==False and ${result_2}==False)    Fail

    [Teardown]    Run Keywords
    ...    Delete Namespaces
    ...    Restore Default MicroShift Config
    ...    Restart MicroShift

*** Keywords ***
Configure Namespace Ownership Allowed
    [Documentation]    Configure MicroShift to use InterNamespaceAllowed namespace ownership
    Setup With Custom Config    ${OWNERSHIP_ALLOW}

Configure Namespace Ownership Strict
    [Documentation]    Configure MicroShift to use Strict namespace ownership
    Setup With Custom Config    ${OWNERSHIP_STRICT}

Setup With Custom Config
    [Documentation]    Install a custom config and restart MicroShift
    [Arguments]    ${config_content}
    ${merged}=    Extend MicroShift Config    ${config_content}
    Upload MicroShift Config    ${merged}

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

Restart Router
    [Documentation]    Restart the router and wait for readiness again. The router is sensitive to apiserver
    ...    downtime and might need a restart (after the apiserver is ready) to resync all the routes.
    Run With Kubeconfig    oc rollout restart deployment router-default -n openshift-ingress
    Named Deployment Should Be Available    router-default    openshift-ingress    5m
