*** Settings ***
Documentation       CSI Snapshotting Tests

Library             OperatingSystem
Library             ../resources/DataFormats.py
Resource            ../resources/microshift-process.resource
Resource            ../resources/common.resource
Resource            ../resources/kubeconfig.resource
Resource            ../resources/oc.resource
Resource            ../resources/microshift-config.resource

Suite Setup         Test Suite Setup
Suite Teardown      Teardown Suite With Namespace


*** Variables ***
${POD_NAME_STATIC}      base
${TEST_DATA}            FOOBAR
${SOURCE_KUSTOMIZE}     assets/kustomizations/patches/pvc-thin
${RESTORE_KUSTOMIZE}    assets/kustomizations/patches/pvc-from-snapshot
${SNAPSHOT}             assets/snapshot.yaml
${STORAGE_CLASS}        assets/storage-class-thin.yaml
${SNAPSHOT_CLASS}       assets/volume-snapshot-class.yaml
${LVMD_PATCH}           assets/lvmd.patch.yaml


*** Test Cases ***
Snapshotter Smoke Test
    [Documentation]    Write data to a volume, snapshot it, restore the snapshot and verify the data is present
    [Tags]    smoke    snapshot
    [Setup]    Test Case Setup

    Oc Apply    -f ${SNAPSHOT} -n ${NAMESPACE}
    Oc Wait For    volumesnapshot/my-snap    jsonpath\='{.status.readyToUse}=true'
    Oc Apply    -k ${RESTORE_KUSTOMIZE} -n ${NAMESPACE}
    Oc Wait For    pod/${POD_NAME_STATIC}    condition\=Ready
    ${data}=    Read From Volume    ${POD_NAME_STATIC}
    Should Be Equal As Strings    ${TESTDATA}    ${data}
    [Teardown]    Test Case Teardown


*** Keywords ***
Test Suite Setup
    [Documentation]    Setup test namespace, patch the lvmd for thin-volume support, and restart microshift
    Setup Suite With Namespace
    Extend LVMD Config
    Restart Microshift

Test Case Setup
    [Documentation]    Prepare the cluster-level APIs and a data-volume with some simple text
    Oc Apply    -f ${STORAGE_CLASS} -f ${SNAPSHOT_CLASS}
    Oc Apply    -k ${SOURCE_KUSTOMIZE} -n ${NAMESPACE}
    Oc Wait For    pod/${POD_NAME_STATIC}    condition\=Ready    timeout=60s
    Write To Volume    ${POD_NAME_STATIC}    ${TEST_DATA}
    Oc Delete    pod ${POD_NAME_STATIC} -n ${NAMESPACE}
    Oc Wait For    pod/${POD_NAME_STATIC}    delete

Test Case Teardown
    [Documentation]    Remove cluster-scoped test APIs
    Oc Delete    -f ${STORAGE_CLASS} -f ${SNAPSHOT_CLASS}

Write To Volume
    [Documentation]    Write some simple text to the data volume
    [Arguments]    ${to_pod}    ${data}
    Oc Exec    ${to_pod}    echo "${data}" > /vol/data

Read From Volume
    [Documentation]    Read textfile from the datavolume and return as a string.
    [Arguments]    ${from_pod}
    ${output}=    Oc Exec    ${from_pod}    cat /vol/data
    RETURN    ${output}

Extend LVMD Config
    [Documentation]    The cluster is not expected to have the necessary configuration to provision thin volumes.
    ...    Instead, we need to patch the lvmd configuration to include a deviceClass for the thin-pool.
    ...    This assumes a thin-pool named 'thin', which is provided by CI.    The base lvmd is pulled from
    ...    the lvmd configMap. This is preferable since no default lvmd.yaml is written to disk. And if one
    ...    is written to disk, it will have been populated into the configMap already.
    # Get the existing lvmd from the cluster.
    ${cm}=    Oc Get    cm    openshift-storage    lvmd
    ${default_cfg}=    Set Variable    ${cm['data']['lvmd.yaml']}
    # Load the lvmd config patch
    ${patch_cfg}=    OperatingSystem.Get File    ${LVMD_PATCH}
    ${merged_cfg}=    Lvmd Merge    ${default_cfg}    ${patch_cfg}
    Upload String To File    ${merged_cfg}    /etc/microshift/lvmd.yaml
