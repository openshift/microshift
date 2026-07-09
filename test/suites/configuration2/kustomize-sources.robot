*** Settings ***
Documentation       Tests for configurable kustomize manifest paths and glob pattern scanning

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
${CONFIGMAP_NAME}       test-configmap
${MANIFEST_DIR_1}       /etc/microshift/manifests.d/ksrc-test-1
${MANIFEST_DIR_2}       /etc/microshift/manifests.d/ksrc-test-2
${GLOB_BASE}            /etc/microshift/manifests.d/ksrc-glob
${GLOB_DIR_A}           /etc/microshift/manifests.d/ksrc-glob/app-a
${GLOB_DIR_B}           /etc/microshift/manifests.d/ksrc-glob/app-b

${EMPTY_PATHS}          SEPARATOR=\n
...                     manifests:
...                     \ \ kustomizePaths: []

${SINGLE_PATH}          SEPARATOR=\n
...                     manifests:
...                     \ \ kustomizePaths:
...                     \ \ \ \ - /etc/microshift/manifests.d/ksrc-test-1

${MULTI_PATHS}          SEPARATOR=\n
...                     manifests:
...                     \ \ kustomizePaths:
...                     \ \ \ \ - /etc/microshift/manifests.d/ksrc-test-1
...                     \ \ \ \ - /etc/microshift/manifests.d/ksrc-test-2

${NULL_PATHS}           SEPARATOR=\n
...                     manifests:
...                     \ \ kustomizePaths:

${GLOB_PATHS}           SEPARATOR=\n
...                     manifests:
...                     \ \ kustomizePaths:
...                     \ \ \ \ - /etc/microshift/manifests.d/ksrc-glob/*/


*** Test Cases ***
Empty Kustomize Paths Disables Manifests
    [Documentation]    Setting kustomizePaths to an empty list disables all manifest loading.
    Deploy Test Manifests    ${MANIFEST_DIR_1}    ksrc-empty-ns
    Drop In MicroShift Config    ${EMPTY_PATHS}    10-kustomize

    Restart MicroShift

    ConfigMap Should Not Exist    ksrc-empty-ns

    [Teardown]    Cleanup Kustomize Test    ksrc-empty-ns    ${MANIFEST_DIR_1}

Single Kustomize Path
    [Documentation]    Setting a single kustomizePath should load manifests only from that path.
    Deploy Test Manifests    ${MANIFEST_DIR_1}    ksrc-single-ns
    Drop In MicroShift Config    ${SINGLE_PATH}    10-kustomize

    Restart MicroShift

    Wait Until Keyword Succeeds    20x    10s
    ...    Oc Get    configmap    ksrc-single-ns    ${CONFIGMAP_NAME}

    [Teardown]    Cleanup Kustomize Test    ksrc-single-ns    ${MANIFEST_DIR_1}

Multiple Kustomize Paths
    [Documentation]    Multiple paths in kustomizePaths should all be loaded.
    Deploy Test Manifests    ${MANIFEST_DIR_1}    ksrc-multi-ns-1
    Deploy Test Manifests    ${MANIFEST_DIR_2}    ksrc-multi-ns-2
    Drop In MicroShift Config    ${MULTI_PATHS}    10-kustomize

    Restart MicroShift

    Wait Until Keyword Succeeds    20x    10s
    ...    Oc Get    configmap    ksrc-multi-ns-1    ${CONFIGMAP_NAME}
    Wait Until Keyword Succeeds    20x    10s
    ...    Oc Get    configmap    ksrc-multi-ns-2    ${CONFIGMAP_NAME}

    [Teardown]    Run Keywords
    ...    Remove Drop In MicroShift Config    10-kustomize
    ...    AND    Remove Manifest Directory    ${MANIFEST_DIR_1}
    ...    AND    Remove Manifest Directory    ${MANIFEST_DIR_2}
    ...    AND    Oc Delete    namespace ksrc-multi-ns-1 --ignore-not-found
    ...    AND    Oc Delete    namespace ksrc-multi-ns-2 --ignore-not-found

Path Without Kustomization File Is Ignored
    [Documentation]    A path that exists but has no kustomization.yaml should be silently ignored.
    Command Should Work    mkdir -p ${MANIFEST_DIR_1}
    Drop In MicroShift Config    ${SINGLE_PATH}    10-kustomize

    Restart MicroShift

    [Teardown]    Run Keywords
    ...    Remove Drop In MicroShift Config    10-kustomize
    ...    AND    Remove Manifest Directory    ${MANIFEST_DIR_1}

Non Existent Path Is Ignored
    [Documentation]    A non-existent path in kustomizePaths should be silently ignored.
    Drop In MicroShift Config    ${SINGLE_PATH}    10-kustomize

    Restart MicroShift

    [Teardown]    Remove Drop In MicroShift Config    10-kustomize

Unset Kustomize Paths Uses Defaults
    [Documentation]    Setting kustomizePaths to null should result in the default paths.
    Drop In MicroShift Config    ${NULL_PATHS}    10-kustomize

    ${config}=    Show Config    effective
    ${paths}=    Evaluate    str(${config.manifests.kustomizePaths})
    Should Contain    ${paths}    /etc/microshift/manifests
    Should Contain    ${paths}    /etc/microshift/manifests.d/*
    Should Contain    ${paths}    /usr/lib/microshift/manifests
    Should Contain    ${paths}    /usr/lib/microshift/manifests.d/*

    [Teardown]    Remove Drop In MicroShift Config    10-kustomize

Glob Patterns In Kustomize Paths
    [Documentation]    Glob patterns in kustomizePaths should match subdirectories.
    Deploy Test Manifests    ${GLOB_DIR_A}    ksrc-glob-ns-a
    Deploy Test Manifests    ${GLOB_DIR_B}    ksrc-glob-ns-b
    Drop In MicroShift Config    ${GLOB_PATHS}    10-kustomize

    Restart MicroShift

    Wait Until Keyword Succeeds    20x    10s
    ...    Oc Get    configmap    ksrc-glob-ns-a    ${CONFIGMAP_NAME}
    Wait Until Keyword Succeeds    20x    10s
    ...    Oc Get    configmap    ksrc-glob-ns-b    ${CONFIGMAP_NAME}

    [Teardown]    Run Keywords
    ...    Remove Drop In MicroShift Config    10-kustomize
    ...    AND    Remove Manifest Directory    ${GLOB_BASE}
    ...    AND    Oc Delete    namespace ksrc-glob-ns-a --ignore-not-found
    ...    AND    Oc Delete    namespace ksrc-glob-ns-b --ignore-not-found


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig

Teardown
    [Documentation]    Restart MicroShift to restore clean state after the last test
    ...    (per-test teardowns skip the restart), then clean up.
    Remove Drop In MicroShift Config    10-kustomize
    Restart MicroShift
    Remove Kubeconfig
    Logout MicroShift Host

Cleanup Kustomize Test
    [Documentation]    Standard cleanup for a single-path kustomize test without restarting.
    ...    The next test's setup will restart MicroShift.
    [Arguments]    ${namespace}    ${manifest_dir}
    Run Keywords
    ...    Remove Drop In MicroShift Config    10-kustomize
    ...    AND    Remove Manifest Directory    ${manifest_dir}
    ...    AND    Oc Delete    namespace ${namespace} --ignore-not-found
