*** Settings ***
Documentation       Tests for drop-in configuration directory merge and override semantics

Resource            ../../resources/common.resource
Resource            ../../resources/kustomize-test.resource
Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-host.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/oc.resource
Library             ../../resources/DataFormats.py

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           restart    slow


*** Variables ***
${MANIFEST_DIR_A}       /etc/microshift/manifests.d/dropin-test-a
${MANIFEST_DIR_B}       /etc/microshift/manifests.d/dropin-test-b

${KUSTOMIZE_A}          SEPARATOR=\n
...                     manifests:
...                     \ \ kustomizePaths:
...                     \ \ \ \ - /etc/microshift/manifests.d/dropin-test-a

${KUSTOMIZE_B}          SEPARATOR=\n
...                     manifests:
...                     \ \ kustomizePaths:
...                     \ \ \ \ - /etc/microshift/manifests.d/dropin-test-b

${SAN_10}               SEPARATOR=\n
...                     apiServer:
...                     \ \ subjectAltNames:
...                     \ \ \ \ - test1.example.com

${SAN_20}               SEPARATOR=\n
...                     apiServer:
...                     \ \ subjectAltNames:
...                     \ \ \ \ - test2.example.com

${ETCD_MEM}             SEPARATOR=\n
...                     etcd:
...                     \ \ memoryLimitMB: 180

${DEBUG_LEVEL}          SEPARATOR=\n
...                     debugging:
...                     \ \ logLevel: debug


*** Test Cases ***
Drop In Sets Kustomize Paths
    [Documentation]    A drop-in config file can set kustomizePaths and manifests are loaded from it.
    Drop In MicroShift Config    ${KUSTOMIZE_A}    10-kustomize
    Deploy Test Manifests    ${MANIFEST_DIR_A}    dropin-ns-a
    Restart MicroShift
    Wait Until Keyword Succeeds    10x    10s
    ...    Oc Get    configmap    dropin-ns-a    test-configmap

    [Teardown]    Run Keywords
    ...    Remove Drop In MicroShift Config    10-kustomize
    ...    AND    Remove Manifest Directory    ${MANIFEST_DIR_A}
    ...    AND    Oc Delete    namespace dropin-ns-a --ignore-not-found
    ...    AND    Restart MicroShift

Higher Numbered Drop In Overrides Array
    [Documentation]    When two drop-ins set the same array field, the higher-numbered
    ...    file wins (arrays are replaced, not merged).
    Drop In MicroShift Config    ${KUSTOMIZE_A}    10-kustomize
    Drop In MicroShift Config    ${KUSTOMIZE_B}    20-kustomize
    Deploy Test Manifests    ${MANIFEST_DIR_A}    dropin-ns-a2
    Deploy Test Manifests    ${MANIFEST_DIR_B}    dropin-ns-b

    Restart MicroShift

    Wait Until Keyword Succeeds    10x    10s
    ...    Oc Get    configmap    dropin-ns-b    test-configmap
    ConfigMap Should Not Exist    dropin-ns-a2

    [Teardown]    Run Keywords
    ...    Remove Drop In MicroShift Config    10-kustomize
    ...    AND    Remove Drop In MicroShift Config    20-kustomize
    ...    AND    Remove Manifest Directory    ${MANIFEST_DIR_A}
    ...    AND    Remove Manifest Directory    ${MANIFEST_DIR_B}
    ...    AND    Oc Delete    namespace dropin-ns-a2 --ignore-not-found
    ...    AND    Oc Delete    namespace dropin-ns-b --ignore-not-found
    ...    AND    Restart MicroShift

SAN Arrays Are Replaced Not Merged
    [Documentation]    SubjectAltNames is an array field. A higher-numbered drop-in
    ...    should replace the array, not merge with the lower-numbered one.
    ...    Verified via show-config without restarting to avoid breaking API access.
    Drop In MicroShift Config    ${SAN_10}    10-san
    Drop In MicroShift Config    ${SAN_20}    20-san

    ${config}=    Show Config    effective
    Should Contain    ${config.apiServer.subjectAltNames.__repr__()}    test2.example.com
    Should Not Contain    ${config.apiServer.subjectAltNames.__repr__()}    test1.example.com

    [Teardown]    Run Keywords
    ...    Remove Drop In MicroShift Config    10-san
    ...    AND    Remove Drop In MicroShift Config    20-san

Map Fields Merge Across Drop Ins
    [Documentation]    Map-type fields should merge across config.yaml and drop-ins.
    ...    Settings from different map keys should all be present.
    ...    Verified via show-config without restarting.
    Drop In MicroShift Config    ${ETCD_MEM}    10-etcd
    Drop In MicroShift Config    ${DEBUG_LEVEL}    20-debug

    ${config}=    Show Config    effective
    Should Be Equal As Integers    180    ${config.etcd.memoryLimitMB}
    Should Be Equal As Strings    debug    ${config.debugging.logLevel}

    [Teardown]    Run Keywords
    ...    Remove Drop In MicroShift Config    10-etcd
    ...    AND    Remove Drop In MicroShift Config    20-debug


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host
    Remove Kubeconfig
