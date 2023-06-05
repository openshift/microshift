*** Settings ***
Documentation     Tests for applying manifests automatically via the kustomize controller

Resource    ../resources/common.resource
Resource    ../resources/systemd.resource
Resource    ../resources/microshift-config.resource
Resource    ../resources/microshift-process.resource

Test Tags         restart  slow


*** Variables ***
${CONFIGMAP_NAME}  test-configmap


*** Test Cases ***
Load From /etc/microshift/manifests
    [Documentation]  /etc/microshift/manifests
    [Setup]  Setup  /etc/microshift/manifests
    ConfigMap Path Should Match
    [Teardown]  Teardown

Load From /etc/microshift/manifestsd
    # Keyword names cannot have '.' in them
    [Documentation]  Subdir of /etc/microshift/manifests.d
    [Setup]  Setup With Subdir  /etc/microshift/manifests.d
    ConfigMap Path Should Match
    [Teardown]  Teardown With Subdir

Load From /usr/lib/microshift/manifests
    [Documentation]  /usr/lib/microshift/manifests
    [Setup]  Setup  /usr/lib/microshift/manifests
    ConfigMap Path Should Match
    [Teardown]  Teardown

Load From /usr/lib/microshift/manifestsd
    # Keyword names cannot have '.' in them
    [Documentation]  Subdir of /usr/lib/microshift/manifests.d
    [Setup]  Setup With Subdir  /usr/lib/microshift/manifests.d
    ConfigMap Path Should Match
    [Teardown]  Teardown With Subdir

Load From Configured Dir
    [Documentation]  Non-default directory
    [Setup]  Setup With Config  /usr/lib/microshift/test-manifests
    ConfigMap Path Should Match
    [Teardown]  Teardown With Config

Do Not Load From Unonfigured Dir
    [Documentation]  Remove default directory from config and ensure manifests are not loaded
    [Setup]  Setup With Limited Config
    ConfigMap Should Be Missing
    [Teardown]  Teardown With Config


*** Keywords ***
Setup
    [Documentation]  Set up for test
    [Arguments]  ${path}
    Set Suite Variable  \${MANIFEST_DIR}  ${path}
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig  # for readiness checks
    ${ns}=  Create Random Namespace
    Set Suite Variable  \${NAMESPACE}  ${ns}
    Clear Manifest Directory
    Write Manifests
    Restart MicroShift

Teardown
    [Documentation]  Clean up after test
    Clear Manifest Directory
    Run With Kubeconfig  oc delete namespace ${NAMESPACE}  allow_fail=True
    Logout MicroShift Host
    Remove Kubeconfig

Setup With Subdir
    [Documentation]  Set up for test using dynamically created subdir
    [Arguments]  ${path}
    ${rand}=    Generate Random String
    Setup  ${path}/${rand}

Teardown With Subdir
    [Documentation]  Remove MANIFEST_DIR and clean up
    ${stdout}  ${rc}=  Execute Command
    ...    rm -rf ${MANIFEST_DIR}
    ...    sudo=True  return_rc=True
    Should Be Equal As Integers  0  ${rc}
    Teardown

Setup With Config  # robocop: disable=too-many-calls-in-keyword
    [Documentation]  Set up for test using non-default directory
    [Arguments]  ${path}
    Set Suite Variable  \${MANIFEST_DIR}  ${path}
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig  # for readiness checks

    # Extend the configuration setting to add the path
    Save Default MicroShift Config
    ${config_content}=  Catenate  SEPARATOR=\n
    ...  manifests:
    ...  \ \ kustomizePaths:
    ...  \ \ \ \ - ${path}
    ${merged}=  Extend MicroShift Config  ${config_content}
    Upload MicroShift Config  ${merged}

    ${ns}=  Create Random Namespace
    Set Suite Variable  \${NAMESPACE}  ${ns}
    Clear Manifest Directory
    Write Manifests
    Restart MicroShift

Teardown With Config
    [Documentation]  Restore default config and clean up
    Restore Default Config
    Teardown With Subdir

Setup With Limited Config  # robocop: disable=too-many-calls-in-keyword
    [Documentation]  Ensure the configuration does *not* include the manifest directory
    Set Suite Variable  \${MANIFEST_DIR}    /usr/lib/microshift/manifests
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig  # for readiness checks

    # Write a config that does not include our path
    Save Default MicroShift Config
    ${config_content}=  Catenate  SEPARATOR=\n
    ...  manifests:
    ...  \ \ kustomizePaths:
    ...  \ \ \ \ - /etc/microshift/manifests
    ${merged}=  Extend MicroShift Config  ${config_content}
    Upload MicroShift Config  ${merged}

    ${ns}=  Create Random Namespace
    Set Suite Variable  \${NAMESPACE}  ${ns}
    Clear Manifest Directory
    Write Manifests
    Restart MicroShift

Restore Default Config
    [Documentation]  Remove any custom config and restart MicroShift
    Restore Default MicroShift Config
    Restart MicroShift
    Sleep  10 seconds  # Wait for systemd to catch up

ConfigMap Path Should Match
    [Documentation]  Ensure the config map path value matches the manifest dir
    ${configmap}=  Oc Get  configmap  ${NAMESPACE}  ${CONFIGMAP_NAME}
    Should Be Equal  ${MANIFEST_DIR}  ${configmap.data.path}

ConfigMap Should Be Missing
    [Documentation]  Ensure the config map was not created
    ${result}=    Run Process    oc get configmap -n ${NAMESPACE} ${CONFIGMAP_NAME}
    ...  env:KUBECONFIG=${KUBECONFIG}
    ...  stderr=STDOUT
    ...  shell=True
    Should Be Equal As Integers  ${result.rc}  1

Clear Manifest Directory
    [Documentation]  Remove the contents of the manifest directory
    ${stdout}  ${rc}=  Execute Command
    ...    rm -rf ${MANIFEST_DIR}/*
    ...    sudo=True  return_rc=True
    Should Be Equal As Integers  0  ${rc}

Write Manifests
    [Documentation]  Install manifests
    # Make sure the manifest directory exists
    ${stdout}  ${rc}=  Execute Command
    ...    mkdir -p ${MANIFEST_DIR}
    ...    sudo=True  return_rc=True
    Should Be Equal As Integers  0  ${rc}
    # Configure kustomization to use the namespace created in Setup
    ${kustomization}=  Catenate  SEPARATOR=\n
    ...  resources:
    ...  - configmap.yaml
    ...  namespace: ${NAMESPACE}
    ...
    Upload String To File  ${kustomization}  ${MANIFEST_DIR}/kustomization.yaml
    # Build a configmap with unique data for this scenario
    ${configmap}=  Catenate  SEPARATOR=\n
    ...  apiVersion: v1
    ...  kind: ConfigMap
    ...  metadata:
    ...  \ \ name: ${CONFIGMAP_NAME}
    ...  data:
    ...  \ \ path: ${MANIFEST_DIR}
    ...
    Upload String To File  ${configmap}  ${MANIFEST_DIR}/configmap.yaml
