*** Settings ***
Documentation       DNS resource configuration tests

Resource            ../../resources/common.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-process.resource
Library             ../../resources/journalctl.py

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           slow    restart


*** Variables ***
${CURSOR}                           ${EMPTY}
${DNS_DROPIN}                       10-dns-resources
${DNS_RESOURCE_PATH}                .spec.template.spec.containers[0].resources
${DNS_CUSTOM_RESOURCES}             SEPARATOR=\n
...                                 ---
...                                 dns:
...                                 \ \ resources:
...                                 \ \ \ requests:
...                                 \ \ \ \ cpu: "100m"
...                                 \ \ \ \ memory: "150Mi"
...                                 \ \ \ limits:
...                                 \ \ \ \ cpu: "200m"
...                                 \ \ \ \ memory: "256Mi"
${DNS_REQUESTS_ONLY}                SEPARATOR=\n
...                                 ---
...                                 dns:
...                                 \ \ resources:
...                                 \ \ \ requests:
...                                 \ \ \ \ cpu: "100m"
...                                 \ \ \ \ memory: "150Mi"
${DNS_PARTIAL_REQUESTS}             SEPARATOR=\n
...                                 ---
...                                 dns:
...                                 \ \ resources:
...                                 \ \ \ requests:
...                                 \ \ \ \ cpu: "100m"
${DNS_INVALID_QUANTITY}             SEPARATOR=\n
...                                 ---
...                                 dns:
...                                 \ \ resources:
...                                 \ \ \ requests:
...                                 \ \ \ \ cpu: "abc"
${DNS_LIMITS_ONLY}                  SEPARATOR=\n
...                                 ---
...                                 dns:
...                                 \ \ resources:
...                                 \ \ \ limits:
...                                 \ \ \ \ cpu: "200m"
...                                 \ \ \ \ memory: "256Mi"
${DNS_LIMIT_LESS_THAN_REQUEST}      SEPARATOR=\n
...                                 ---
...                                 dns:
...                                 \ \ resources:
...                                 \ \ \ requests:
...                                 \ \ \ \ cpu: "200m"
...                                 \ \ \ limits:
...                                 \ \ \ \ cpu: "50m"


*** Test Cases ***
Default DNS Resources
    [Documentation]    Verify default DNS resources when no custom config is applied
    [Setup]    Run Keywords    Remove DNS Resource Config    AND    Restart MicroShift
    DNS Resource Value Should Be    ${DNS_RESOURCE_PATH}.requests.cpu    50m
    DNS Resource Value Should Be    ${DNS_RESOURCE_PATH}.requests.memory    70Mi

Custom DNS Resources With Requests And Limits
    [Documentation]    Configure custom CPU and memory requests and limits via drop-in config
    [Setup]    Apply DNS Resource Config    ${DNS_CUSTOM_RESOURCES}
    DNS Resource Value Should Be    ${DNS_RESOURCE_PATH}.requests.cpu    100m
    DNS Resource Value Should Be    ${DNS_RESOURCE_PATH}.requests.memory    150Mi
    DNS Resource Value Should Be    ${DNS_RESOURCE_PATH}.limits.cpu    200m
    DNS Resource Value Should Be    ${DNS_RESOURCE_PATH}.limits.memory    256Mi
    [Teardown]    Remove DNS Resource Config

Requests Only Without Limits
    [Documentation]    Configure only requests without limits and verify no limits are injected
    [Setup]    Apply DNS Resource Config    ${DNS_REQUESTS_ONLY}
    DNS Resource Value Should Be    ${DNS_RESOURCE_PATH}.requests.cpu    100m
    DNS Resource Value Should Be    ${DNS_RESOURCE_PATH}.requests.memory    150Mi
    DNS Resource Value Should Be Empty    ${DNS_RESOURCE_PATH}.limits
    [Teardown]    Remove DNS Resource Config

Partial Requests Preserves Defaults
    [Documentation]    Configure only CPU request and verify memory default is preserved
    [Setup]    Apply DNS Resource Config    ${DNS_PARTIAL_REQUESTS}
    DNS Resource Value Should Be    ${DNS_RESOURCE_PATH}.requests.cpu    100m
    DNS Resource Value Should Be    ${DNS_RESOURCE_PATH}.requests.memory    70Mi
    [Teardown]    Remove DNS Resource Config

Limits Only Preserves Default Requests
    [Documentation]    Configure only limits and verify default requests are preserved
    [Setup]    Apply DNS Resource Config    ${DNS_LIMITS_ONLY}
    DNS Resource Value Should Be    ${DNS_RESOURCE_PATH}.requests.cpu    50m
    DNS Resource Value Should Be    ${DNS_RESOURCE_PATH}.requests.memory    70Mi
    DNS Resource Value Should Be    ${DNS_RESOURCE_PATH}.limits.cpu    200m
    DNS Resource Value Should Be    ${DNS_RESOURCE_PATH}.limits.memory    256Mi
    [Teardown]    Remove DNS Resource Config

Invalid Resource Quantity Prevents Start
    [Documentation]    Verify MicroShift fails to start with invalid resource quantity
    [Setup]    Apply Invalid DNS Resource Config    ${DNS_INVALID_QUANTITY}
    Pattern Should Appear In Log Output    ${CURSOR}    invalid dns resource request
    [Teardown]    Run Keywords    Remove DNS Resource Config    AND    Restart MicroShift

Limit Less Than Request Prevents Start
    [Documentation]    Verify MicroShift fails to start when limit is less than request
    [Setup]    Apply Invalid DNS Resource Config    ${DNS_LIMIT_LESS_THAN_REQUEST}
    Pattern Should Appear In Log Output    ${CURSOR}    must be greater than or equal to request
    [Teardown]    Run Keywords    Remove DNS Resource Config    AND    Restart MicroShift

DNS Resolution After Resource Change
    [Documentation]    Verify CoreDNS resolves cluster-local services after resource change
    [Setup]    Apply DNS Resource Config    ${DNS_CUSTOM_RESOURCES}
    ${output}=    Oc Exec    router-default    nslookup kubernetes.default.svc.cluster.local
    ...    openshift-ingress    deployment
    Should Contain    ${output}    kubernetes.default.svc.cluster.local
    [Teardown]    Remove DNS Resource Config


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig

Teardown
    [Documentation]    Restart MicroShift to restore clean state after the last test
    ...    (per-test teardowns skip the restart), then clean up.
    Remove Drop In MicroShift Config    ${DNS_DROPIN}
    Restart MicroShift
    Remove Kubeconfig
    Logout MicroShift Host

Get DNS Resource Value
    [Documentation]    Get a resource value from the dns-default DaemonSet
    [Arguments]    ${jsonpath}
    ${value}=    Oc Get JsonPath    daemonset    openshift-dns    dns-default    ${jsonpath}
    RETURN    ${value}

DNS Resource Value Should Be
    [Documentation]    Wait for the dns-default DaemonSet resource value to match expected
    [Arguments]    ${jsonpath}    ${expected}
    Wait Until Keyword Succeeds    10x    5s
    ...    DNS Resource Value Should Match    ${jsonpath}    ${expected}

DNS Resource Value Should Be Empty
    [Documentation]    Wait for the dns-default DaemonSet resource value to be empty
    [Arguments]    ${jsonpath}
    Wait Until Keyword Succeeds    10x    5s
    ...    DNS Resource Value Should Match Empty    ${jsonpath}

DNS Resource Value Should Match
    [Documentation]    Assert a dns-default DaemonSet resource value matches expected
    [Arguments]    ${jsonpath}    ${expected}
    ${value}=    Get DNS Resource Value    ${jsonpath}
    Should Be Equal As Strings    ${value}    ${expected}

DNS Resource Value Should Match Empty
    [Documentation]    Assert a dns-default DaemonSet resource value is empty
    [Arguments]    ${jsonpath}
    ${value}=    Get DNS Resource Value    ${jsonpath}
    Should Be Empty    ${value}

Apply DNS Resource Config
    [Documentation]    Remove any existing drop-in, apply a new DNS resource config and restart MicroShift
    [Arguments]    ${config}
    Remove Drop In MicroShift Config    ${DNS_DROPIN}
    Drop In MicroShift Config    ${config}    ${DNS_DROPIN}
    Restart MicroShift

Apply Invalid DNS Resource Config
    [Documentation]    Apply an invalid DNS resource config that should prevent MicroShift from starting
    [Arguments]    ${config}
    Remove Drop In MicroShift Config    ${DNS_DROPIN}
    Restart MicroShift
    Drop In MicroShift Config    ${config}    ${DNS_DROPIN}
    ${cursor}=    Get Journal Cursor
    VAR    ${CURSOR}=    ${cursor}    scope=TEST
    Run Keyword And Expect Error    0 != 1    Restart MicroShift

Remove DNS Resource Config
    [Documentation]    Remove the DNS resource drop-in config without restarting.
    ...    The next test's setup will restart MicroShift.
    Remove Drop In MicroShift Config    ${DNS_DROPIN}
