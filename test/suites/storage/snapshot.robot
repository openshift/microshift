*** Settings ***
Documentation       CSI Snapshotting Tests

Library             SSHLibrary
Library             ../../resources/DataFormats.py
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/common.resource
Resource            ../../resources/kubeconfig.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/microshift-config.resource
Resource            ../../resources/ostree-health.resource

Suite Setup         Test Suite Setup
Suite Teardown      Test Suite Teardown


*** Variables ***
${POD_NAME_STATIC}      base
${TEST_DATA}            FOOBAR
${SOURCE_KUSTOMIZE}     assets/kustomizations/patches/pvc-thin
${RESTORE_KUSTOMIZE}    assets/kustomizations/patches/pvc-from-snapshot
${SNAPSHOT}             assets/snapshot.yaml
${STORAGE_CLASS}        assets/storage-class-thin.yaml
${SNAPSHOT_CLASS}       assets/volume-snapshot-class.yaml


*** Test Cases ***
Snapshotter Smoke Test
    [Documentation]    Write data to a volume, snapshot it, restore the snapshot and verify the data is present
    [Tags]    smoke    snapshot
    [Setup]    Test Case Setup
    Oc Apply    -f ${SNAPSHOT} -n ${NAMESPACE}
    Named VolumeSnapshot Should Be Ready    my-snap
    Oc Apply    -k ${RESTORE_KUSTOMIZE} -n ${NAMESPACE}
    Named Pod Should Be Ready    ${POD_NAME_STATIC}
    ${data}=    Read From Volume    ${POD_NAME_STATIC}
    Should Be Equal As Strings    ${TESTDATA}    ${data}
    [Teardown]    Test Case Teardown


*** Keywords ***
Test Suite Setup
    [Documentation]    Setup test namespace, patch the lvmd for thin-volume support, and restart microshift for
    ...    it to take effect
    Setup Suite With Namespace
    Create Thin Storage Pool
    Save Lvmd Config
    ${config}=    Extend Lvmd Config
    Upload Lvmd Config    ${config}
    Oc Apply    -f ${STORAGE_CLASS} -f ${SNAPSHOT_CLASS}
    Restart Microshift
    Restart Greenboot And Wait For Success

Test Suite Teardown
    [Documentation]    Clean up test suite resources
    Oc Delete    -f ${STORAGE_CLASS} -f ${SNAPSHOT_CLASS}
    Restore Lvmd Config
    Delete Thin Storage Pool
    Restart Microshift
    Restart Greenboot And Wait For Success
    Teardown Suite With Namespace

Test Case Setup
    [Documentation]    Prepare the cluster-level APIs and a data-volume with some simple text
    Oc Apply    -k ${SOURCE_KUSTOMIZE} -n ${NAMESPACE}
    Named Pod Should Be Ready    ${POD_NAME_STATIC}
    Write To Volume    ${POD_NAME_STATIC}    ${TEST_DATA}
    Oc Delete    pod ${POD_NAME_STATIC} -n ${NAMESPACE}
    Named Pod Should Be Deleted    ${POD_NAME_STATIC}

Test Case Teardown
    [Documentation]    Remove cluster-scoped test APIs
    Oc Delete    pod ${POD_NAME_STATIC} -n ${NAMESPACE}
    Named Pod Should Be Deleted    ${POD_NAME_STATIC}
    Oc Delete    volumesnapshot my-snap -n ${NAMESPACE}
    Named VolumeSnapshot Should Be Deleted    my-snap
    Oc Delete    pvc test-claim-thin -n ${NAMESPACE}
    Oc Delete    pvc snapshot-restore -n ${NAMESPACE}
    Named PVC Should Be Deleted    test-claim-thin
    Named PVC Should Be Deleted    snapshot-restore

Write To Volume
    [Documentation]    Write some simple text to the data volume
    [Arguments]    ${to_pod}    ${data}
    Oc Exec    ${to_pod}    echo "${data}" > /vol/data

Read From Volume
    [Documentation]    Read textfile from the datavolume and return as a string.
    [Arguments]    ${from_pod}
    ${output}=    Oc Exec    ${from_pod}    cat /vol/data
    RETURN    ${output}
