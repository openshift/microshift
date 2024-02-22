*** Settings ***
Documentation       Tests for applying manifests automatically via the kustomize controller

Resource            ../../resources/common.resource
Resource            ../../resources/systemd.resource
Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-process.resource

Suite Setup         Setup Suite
Suite Teardown      Teardown Suite

Test Tags           restart    slow


*** Variables ***
${CONFIGMAP_NAME}       test-configmap
${NON_DEFAULT_DIR}      /home/${USHIFT_USER}/test-manifests
${UNCONFIGURED_DIR}     /home/${USHIFT_USER}/test-manifests.d/unconfigured
${YAML_PATH}            /etc/microshift/manifests.d/yaml-ext
${YML_PATH}             /etc/microshift/manifests.d/yml-ext
${NOEXT_PATH}           /etc/microshift/manifests.d/no-ext


*** Test Cases ***
Load From /etc/microshift/manifests
    [Documentation]    /etc/microshift/manifests
    ConfigMap Path Should Match    ${ETC_NAMESPACE}    /etc/microshift/manifests

Load From /etc/microshift/manifestsd
    # Keyword names cannot have '.' in them
    [Documentation]    Subdir of /etc/microshift/manifests.d
    ConfigMap Path Should Match    ${ETC_SUBDIR_NAMESPACE}    ${ETC_SUBDIR}

Load From /usr/lib/microshift/manifests
    [Documentation]    /usr/lib/microshift/manifests
    ConfigMap Path Should Match    ${USR_NAMESPACE}    /usr/lib/microshift/manifests

Load From /usr/lib/microshift/manifestsd
    # Keyword names cannot have '.' in them
    [Documentation]    Subdir of /usr/lib/microshift/manifests.d
    ConfigMap Path Should Match    ${USR_SUBDIR_NAMESPACE}    ${USR_SUBDIR}

Load From Configured Dir
    [Documentation]    Non-default directory
    ConfigMap Path Should Match    ${NON_DEFAULT_NAMESPACE}    ${NON_DEFAULT_DIR}

Do Not Load From Unconfigured Dir
    [Documentation]    Manifests from a directory not in the config should not be loaded
    ConfigMap Should Be Missing    ${UNCONFIGURED_NAMESPACE}

Yaml Extension
    [Documentation]    Root file kustomization.yaml
    ConfigMap Path Should Match    ${YAML_NAMESPACE}    ${YAML_PATH}

Yml Extension
    [Documentation]    Root file kustomization.yml
    ConfigMap Path Should Match    ${YML_NAMESPACE}    ${YML_PATH}

No Extension
    [Documentation]    Root file Kustomization
    ConfigMap Path Should Match    ${NOEXT_NAMESPACE}    ${NOEXT_PATH}


*** Keywords ***
Setup Suite    # robocop: disable=too-long-keyword
    [Documentation]    Set up all of the tests in this suite
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig    # for readiness checks
    Make Usr Writable If OSTree System

    # Used by "Load From /etc/microshift/manifests"
    ${ns}=    Generate Manifests    /etc/microshift/manifests
    Set Suite Variable    \${ETC_NAMESPACE}    ${ns}

    # Used by "Load From /etc/microshift/manifestsd"
    ${rand}=    Generate Random String
    Set Suite Variable    \${ETC_SUBDIR}    /etc/microshift/manifests.d/${rand}
    ${ns}=    Generate Manifests    ${ETC_SUBDIR}
    Set Suite Variable    \${ETC_SUBDIR_NAMESPACE}    ${ns}

    # Used by "Load From /usr/lib/microshift/manifests"
    ${ns}=    Generate Manifests    /usr/lib/microshift/manifests
    Set Suite Variable    \${USR_NAMESPACE}    ${ns}

    # Used by "Load From /usr/lib/microshift/manifestsd"
    ${rand}=    Generate Random String
    Set Suite Variable    \${USR_SUBDIR}    /usr/lib/microshift/manifests.d/${rand}
    ${ns}=    Generate Manifests    ${USR_SUBDIR}
    Set Suite Variable    \${USR_SUBDIR_NAMESPACE}    ${ns}

    # Used by "Load From Configured Dir"
    ${ns}=    Generate Manifests    ${NON_DEFAULT_DIR}
    Set Suite Variable    \${NON_DEFAULT_NAMESPACE}    ${ns}

    # Used by "Do Not Load From Unconfigured Dir"
    ${ns}=    Generate Manifests    ${UNCONFIGURED_DIR}
    Set Suite Variable    \${UNCONFIGURED_NAMESPACE}    ${ns}

    # Used by "Yaml Extension"
    ${ns}=    Generate Manifests    ${YAML_PATH}    kustomization.yaml
    Set Suite Variable    \${YAML_NAMESPACE}    ${ns}

    # Used by "Yml Extension"
    ${ns}=    Generate Manifests    ${YML_PATH}    kustomization.yml
    Set Suite Variable    \${YML_NAMESPACE}    ${ns}

    # Used by "No Extension"
    ${ns}=    Generate Manifests    ${NOEXT_PATH}    Kustomization
    Set Suite Variable    \${NOEXT_NAMESPACE}    ${ns}

    # Extend the configuration setting to add the unique path to the defaults
    Save Default MicroShift Config
    ${config_content}=    Catenate    SEPARATOR=\n
    ...    manifests:
    ...    \ \ kustomizePaths:
    ...    \ \ \ \ - /etc/microshift/manifests
    ...    \ \ \ \ - /etc/microshift/manifests.d/*
    ...    \ \ \ \ - /usr/lib/microshift/manifests
    ...    \ \ \ \ - /usr/lib/microshift/manifests.d/*
    ...    \ \ \ \ - ${NON_DEFAULT_DIR}
    ...    \ \ \ \ # Add a directory _without_ the glob for unconfigured test
    ...    \ \ \ \ - /home/${USHIFT_USER}/test-manifests.d
    ${merged}=    Extend MicroShift Config    ${config_content}
    Upload MicroShift Config    ${merged}

    Restart MicroShift

Teardown Suite    # robocop: disable=too-many-calls-in-keyword
    [Documentation]    Clean up all of the tests in this suite

    Clear Manifest Directory    /etc/microshift/manifests
    Remove Manifest Directory    ${ETC_SUBDIR}
    Clear Manifest Directory    /usr/lib/microshift/manifests
    Remove Manifest Directory    ${USR_SUBDIR}
    Remove Manifest Directory    ${NON_DEFAULT_DIR}
    Remove Manifest Directory    ${UNCONFIGURED_DIR}
    Remove Manifest Directory    ${YAML_PATH}
    Remove Manifest Directory    ${YML_PATH}
    Remove Manifest Directory    ${NOEXT_PATH}

    Run With Kubeconfig    oc delete namespace ${ETC_NAMESPACE}    allow_fail=True
    Run With Kubeconfig    oc delete namespace ${ETC_SUBDIR_NAMESPACE}    allow_fail=True
    Run With Kubeconfig    oc delete namespace ${USR_NAMESPACE}    allow_fail=True
    Run With Kubeconfig    oc delete namespace ${USR_SUBDIR_NAMESPACE}    allow_fail=True
    Run With Kubeconfig    oc delete namespace ${NON_DEFAULT_NAMESPACE}    allow_fail=True
    Run With Kubeconfig    oc delete namespace ${UNCONFIGURED_NAMESPACE}    allow_fail=True
    Run With Kubeconfig    oc delete namespace ${YAML_NAMESPACE}    allow_fail=True
    Run With Kubeconfig    oc delete namespace ${YML_NAMESPACE}    allow_fail=True
    Run With Kubeconfig    oc delete namespace ${NOEXT_NAMESPACE}    allow_fail=True

    Restore Default Config
    Logout MicroShift Host
    Remove Kubeconfig

Make Usr Writable If OSTree System
    [Documentation]    Makes /usr directory writable if host is an OSTree system.
    ${is_ostree}=    Is System OSTree
    IF    ${is_ostree}    Create Usr Directory Overlay

Generate Manifests
    [Documentation]    Create a namespace and the manifests in the given path.
    ...    Return the namespace.
    [Arguments]    ${manifest_dir}    ${kfile}=kustomization.yaml
    ${ns}=    Create Random Namespace
    Clear Manifest Directory    ${manifest_dir}
    Write Manifests    ${manifest_dir}    ${ns}    ${kfile}
    RETURN    ${ns}

Write Manifests
    [Documentation]    Install manifests
    [Arguments]    ${manifest_dir}    ${namespace}    ${kfile}
    # Make sure the manifest directory exists
    ${stdout}    ${rc}=    Execute Command
    ...    mkdir -p ${manifest_dir}
    ...    sudo=True    return_rc=True
    Should Be Equal As Integers    0    ${rc}
    # Configure kustomization to use the namespace created in Setup
    ${kustomization}=    Catenate    SEPARATOR=\n
    ...    resources:
    ...    - configmap.yaml
    ...    namespace: ${namespace}
    ...
    Upload String To File    ${kustomization}    ${manifest_dir}/${kfile}
    # Build a configmap with unique data for this scenario
    ${configmap}=    Catenate    SEPARATOR=\n
    ...    apiVersion: v1
    ...    kind: ConfigMap
    ...    metadata:
    ...    \ \ name: ${CONFIGMAP_NAME}
    ...    data:
    ...    \ \ path: ${manifest_dir}
    ...
    Upload String To File    ${configmap}    ${manifest_dir}/configmap.yaml

ConfigMap Path Should Match
    [Documentation]    Ensure the config map path value matches the manifest dir
    [Arguments]    ${namespace}    ${manifest_dir}
    ${configmap}=    Oc Get    configmap    ${namespace}    ${CONFIGMAP_NAME}
    Should Be Equal    ${manifest_dir}    ${configmap.data.path}

ConfigMap Should Be Missing
    [Documentation]    Ensure the config map was not created
    [Arguments]    ${namespace}
    ${result}=    Run Process    oc get configmap -n ${namespace} ${CONFIGMAP_NAME}
    ...    env:KUBECONFIG=${KUBECONFIG}
    ...    stderr=STDOUT
    ...    shell=True
    Should Be Equal As Integers    ${result.rc}    1

Create Usr Directory Overlay
    [Documentation]    Make /usr dir writable by creating an overlay, rebooting will go back to being immutable.
    ${stdout}    ${rc}=    Execute Command
    ...    rpm-ostree usroverlay
    ...    sudo=True    return_rc=True
    Should Be Equal As Integers    0    ${rc}

Clear Manifest Directory
    [Documentation]    Remove the contents of the manifest directory
    [Arguments]    ${manifest_dir}
    ${stdout}    ${rc}=    Execute Command
    ...    rm -rf ${manifest_dir}/*
    ...    sudo=True    return_rc=True
    Should Be Equal As Integers    0    ${rc}

Remove Manifest Directory
    [Documentation]    Completely remove the directory
    [Arguments]    ${manifest_dir}
    ${stdout}    ${rc}=    Execute Command
    ...    rm -rf ${manifest_dir}
    ...    sudo=True    return_rc=True
    Should Be Equal As Integers    0    ${rc}

Restore Default Config
    [Documentation]    Remove any custom config and restart MicroShift
    Restore Default MicroShift Config

    # When restoring, we check if ostree is active, if so we reboot
    # to convert everything back to normal, MicroShift restart should not
    # be needed in that instance
    ${is_ostree}=    Is System OSTree
    IF    ${is_ostree}
        Reboot MicroShift Host
        Wait For MicroShift
    ELSE
        Restart MicroShift
        Sleep    10 seconds    # Wait for systemd to catch up
    END
