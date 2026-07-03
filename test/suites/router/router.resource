*** Settings ***
Documentation       Shared keywords for router test suites

Resource            ../../resources/common.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/ostree-health.resource
Resource            ../../resources/microshift-network.resource
Resource            ../../resources/microshift-config.resource


*** Keywords ***
Setup
    [Documentation]    Special setup for the suite. As every test case configures MicroShift in
    ...    different ways there is no need to restart before/after each one of them. Instead, store
    ...    the original configuration here to restore it at the end.
    Setup Suite With Namespace

Teardown
    [Documentation]    Special teardown for the suite, will finish off by restoring the original
    ...    configuration and restarting MicroShift.
    Remove Custom Config
    Restart MicroShift
    Teardown Suite With Namespace

Setup With Custom Config
    [Documentation]    Install a custom config and restart MicroShift.
    [Arguments]    ${config_content}
    Drop In MicroShift Config    ${config_content}    10-ingress
    Restart MicroShift

Remove Custom Config
    [Documentation]    Remove custom config.
    Remove Drop In MicroShift Config    10-ingress

Wait For Router Ready
    [Documentation]    Wait for the default router to be ready.
    Oc Wait    namespace/openshift-ingress    --for jsonpath='{.status.phase}=Active' --timeout=5m
    Named Deployment Should Be Available    router-default    openshift-ingress    5m

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
