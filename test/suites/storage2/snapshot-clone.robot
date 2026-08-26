*** Settings ***
Documentation       Snapshot restore, snapshot content lifecycle, and PVC clone tests.
...                 Migrated from openshift-tests-private OCP-64839, 64840, 64842, 64843, 64856, 64857, 64858.

Library             SSHLibrary
Library             String
Library             ../../resources/DataFormats.py
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/common.resource
Resource            ../../resources/kubeconfig.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/microshift-config.resource
Resource            ../../resources/ostree-health.resource
Resource            ../../resources/storage.resource

Suite Setup         Test Suite Setup
Suite Teardown      Test Suite Teardown


*** Variables ***
${THIN_SC}          topolvm-provisioner-thin
${STORAGE_CLASS}    assets/storage/storage-class-thin.yaml
${MOUNT_PATH}       /mnt/storage
${BLOCK_DEVICE}     /dev/dblock


*** Test Cases ***
OCP-64839 Snapshot Restore Filesystem
    [Documentation]    Provision storage with snapshot datasource and restore successfully using Filesystem mode.
    [Tags]    snapshot
    [Setup]    Test Case Setup
    Create PVC    pvc-ori    ${THIN_SC}    1Gi
    Create Pod    pod-ori    pvc-ori    mount_path=${MOUNT_PATH}
    Named Pod Should Be Ready    pod-ori
    Volume Should Be RW    pod-ori    ${MOUNT_PATH}
    Oc Exec    pod-ori    sync

    Create VolumeSnapshotClass    vsc-64839    deletion_policy=Delete
    Create VolumeSnapshot    snap-64839    pvc-ori    vsc-64839
    Named VolumeSnapshot Should Be Ready    snap-64839

    Create PVC From Snapshot    pvc-restore    ${THIN_SC}    1Gi    snap-64839
    Create Pod    pod-restore    pvc-restore    mount_path=${MOUNT_PATH}
    Named Pod Should Be Ready    pod-restore

    ${output}=    Oc Exec    pod-restore    cat ${MOUNT_PATH}/testfile
    Should Contain    ${output}    storage test
    Volume Should Be RW    pod-restore    ${MOUNT_PATH}
    [Teardown]    Snapshot Filesystem Teardown    vsc-64839    snap-64839

OCP-64840 Snapshot Restore Block
    [Documentation]    Provision storage with snapshot datasource and restore successfully using Block mode.
    [Tags]    snapshot
    [Setup]    Test Case Setup
    Create PVC    pvc-ori    ${THIN_SC}    1Gi    volume_mode=Block
    Create Pod With Block Volume    pod-ori    pvc-ori    device_path=${BLOCK_DEVICE}
    Named Pod Should Be Ready    pod-ori
    Write Data To Block Volume    pod-ori    ${BLOCK_DEVICE}
    Oc Exec    pod-ori    sync

    Create VolumeSnapshotClass    vsc-64840    deletion_policy=Delete
    Create VolumeSnapshot    snap-64840    pvc-ori    vsc-64840
    Named VolumeSnapshot Should Be Ready    snap-64840

    Create PVC From Snapshot    pvc-restore    ${THIN_SC}    1Gi    snap-64840    volume_mode=Block
    Create Pod With Block Volume    pod-restore    pvc-restore    device_path=${BLOCK_DEVICE}
    Named Pod Should Be Ready    pod-restore
    Verify Data In Block Volume    pod-restore    ${BLOCK_DEVICE}
    [Teardown]    Snapshot Filesystem Teardown    vsc-64840    snap-64840

OCP-64842 VolumeSnapshotContent Deleted With Delete Policy
    [Documentation]    VolumeSnapshotContent should be removed when snapshot is deleted with deletionPolicy Delete.
    [Tags]    snapshot
    [Setup]    Test Case Setup
    Create PVC    pvc-ori    ${THIN_SC}    1Gi
    Create Pod    pod-ori    pvc-ori    mount_path=${MOUNT_PATH}
    Named Pod Should Be Ready    pod-ori

    Create VolumeSnapshotClass    vsc-64842    deletion_policy=Delete
    Create VolumeSnapshot    snap-64842    pvc-ori    vsc-64842
    Named VolumeSnapshot Should Be Ready    snap-64842

    ${content_name}=    Get VolumeSnapshotContent Name    snap-64842
    Oc Delete    volumesnapshot snap-64842 -n ${NAMESPACE}
    Wait Until Keyword Succeeds    30s    5s
    ...    Resource Should Not Exist Cluster Scoped    volumesnapshotcontent    ${content_name}
    [Teardown]    Snapshot Content Delete Teardown    vsc-64842

OCP-64843 VolumeSnapshotContent Retained With Retain Policy
    [Documentation]    VolumeSnapshotContent should NOT be removed when snapshot is deleted with deletionPolicy Retain.
    [Tags]    snapshot
    [Setup]    Test Case Setup
    Create PVC    pvc-ori    ${THIN_SC}    1Gi
    Create Pod    pod-ori    pvc-ori    mount_path=${MOUNT_PATH}
    Named Pod Should Be Ready    pod-ori

    Create VolumeSnapshotClass    vsc-64843    deletion_policy=Retain
    Create VolumeSnapshot    snap-64843    pvc-ori    vsc-64843
    Named VolumeSnapshot Should Be Ready    snap-64843

    ${content_name}=    Get VolumeSnapshotContent Name    snap-64843
    VAR    ${VSCONTENT_NAME}=    ${content_name}    scope=TEST

    Oc Delete    volumesnapshot snap-64843 -n ${NAMESPACE}
    VolumeSnapshotContent Should Still Exist    ${content_name}
    [Teardown]    Snapshot Retain Teardown    vsc-64843    ${VSCONTENT_NAME}

OCP-64856 Snapshot Restore With Different SC Same Device Class
    [Documentation]    Snapshot restore should work with a different StorageClass using the same device class.
    [Tags]    snapshot
    [Setup]    Test Case Setup
    Create PVC    pvc-ori    ${THIN_SC}    1Gi
    Create Pod    pod-ori    pvc-ori    mount_path=${MOUNT_PATH}
    Named Pod Should Be Ready    pod-ori
    Volume Should Be RW    pod-ori    ${MOUNT_PATH}
    Oc Exec    pod-ori    sync

    Create VolumeSnapshotClass    vsc-64856    deletion_policy=Delete
    Create VolumeSnapshot    snap-64856    pvc-ori    vsc-64856
    Named VolumeSnapshot Should Be Ready    snap-64856

    ${parameters}=    Catenate    SEPARATOR=\n
    ...    parameters:
    ...    ${SPACE}${SPACE}csi.storage.k8s.io/fstype: xfs
    ...    ${SPACE}${SPACE}topolvm.io/device-class: "thin"
    Create StorageClass    sc-64856-alt    binding_mode=WaitForFirstConsumer    parameters=${parameters}

    Create PVC From Snapshot    pvc-restore    sc-64856-alt    1Gi    snap-64856
    Create Pod    pod-restore    pvc-restore    mount_path=${MOUNT_PATH}
    Named Pod Should Be Ready    pod-restore

    ${output}=    Oc Exec    pod-restore    cat ${MOUNT_PATH}/testfile
    Should Contain    ${output}    storage test
    Volume Should Be RW    pod-restore    ${MOUNT_PATH}
    [Teardown]    Snapshot Different SC Teardown    vsc-64856    snap-64856    sc-64856-alt

OCP-64857 Clone PVC Filesystem
    [Documentation]    Clone a PVC with Filesystem VolumeMode and verify data in the clone.
    [Tags]    clone
    [Setup]    Test Case Setup
    Create PVC    pvc-ori    ${THIN_SC}    1Gi
    Create Pod    pod-ori    pvc-ori    mount_path=${MOUNT_PATH}
    Named Pod Should Be Ready    pod-ori
    Volume Should Be RW    pod-ori    ${MOUNT_PATH}
    Oc Exec    pod-ori    sync

    Create PVC From Clone    pvc-clone    ${THIN_SC}    1Gi    pvc-ori
    Create Pod    pod-clone    pvc-clone    mount_path=${MOUNT_PATH}
    Named Pod Should Be Ready    pod-clone

    Oc Delete    pod pod-ori -n ${NAMESPACE}
    Named Pod Should Be Deleted    pod-ori
    Oc Delete    pvc pvc-ori -n ${NAMESPACE}

    ${output}=    Oc Exec    pod-clone    cat ${MOUNT_PATH}/testfile
    Should Contain    ${output}    storage test
    Volume Should Be RW    pod-clone    ${MOUNT_PATH}
    [Teardown]    Test Case Teardown

OCP-64858 Clone PVC Block
    [Documentation]    Clone a PVC with Block VolumeMode and verify data in the clone.
    [Tags]    clone
    [Setup]    Test Case Setup
    Create PVC    pvc-ori    ${THIN_SC}    1Gi    volume_mode=Block
    Create Pod With Block Volume    pod-ori    pvc-ori    device_path=${BLOCK_DEVICE}
    Named Pod Should Be Ready    pod-ori
    Write Data To Block Volume    pod-ori    ${BLOCK_DEVICE}
    Oc Exec    pod-ori    sync

    Create PVC From Clone    pvc-clone    ${THIN_SC}    1Gi    pvc-ori    volume_mode=Block
    Create Pod With Block Volume    pod-clone    pvc-clone    device_path=${BLOCK_DEVICE}
    Named Pod Should Be Ready    pod-clone

    Oc Delete    pod pod-ori -n ${NAMESPACE}
    Named Pod Should Be Deleted    pod-ori
    Oc Delete    pvc pvc-ori -n ${NAMESPACE}

    Verify Data In Block Volume    pod-clone    ${BLOCK_DEVICE}
    [Teardown]    Test Case Teardown


*** Keywords ***
Test Suite Setup
    [Documentation]    Create thin storage pool, configure LVMD, restart MicroShift.
    Setup Suite
    Create Thin Storage Pool
    Save Lvmd Config
    ${config}=    Extend Lvmd Config
    Upload Lvmd Config    ${config}
    Oc Apply    -f ${STORAGE_CLASS}
    Restart Microshift
    Restart Greenboot And Wait For Success

Test Suite Teardown
    [Documentation]    Restore LVMD config, remove thin pool, restart MicroShift.
    Oc Delete    -f ${STORAGE_CLASS}
    Restore Lvmd Config
    Delete Thin Storage Pool
    Restart Microshift
    Restart Greenboot And Wait For Success
    Teardown Suite

Test Case Setup
    [Documentation]    Create a unique test namespace.
    ${ns}=    Create Unique Namespace
    VAR    ${NAMESPACE}=    ${ns}    scope=TEST

Test Case Teardown
    [Documentation]    Remove the test namespace and verify no leftover logical volumes.
    Run Keyword And Ignore Error
    ...    Remove Namespace    ${NAMESPACE}
    Run Keyword And Ignore Error
    ...    No Topolvm LogicalVolumes Should Exist

Snapshot Filesystem Teardown
    [Documentation]    Clean up snapshot, VolumeSnapshotClass, and namespace.
    [Arguments]    ${vsc_name}    ${snap_name}
    Run Keyword And Ignore Error
    ...    Oc Delete    volumesnapshot ${snap_name} -n ${NAMESPACE}
    Run Keyword And Ignore Error
    ...    Delete VolumeSnapshotClass    ${vsc_name}
    Run Keyword And Ignore Error
    ...    Remove Namespace    ${NAMESPACE}
    Run Keyword And Ignore Error
    ...    No Topolvm LogicalVolumes Should Exist

Snapshot Content Delete Teardown
    [Documentation]    Clean up VolumeSnapshotClass and namespace after content-deletion test.
    [Arguments]    ${vsc_name}
    Run Keyword And Ignore Error
    ...    Delete VolumeSnapshotClass    ${vsc_name}
    Run Keyword And Ignore Error
    ...    Remove Namespace    ${NAMESPACE}
    Run Keyword And Ignore Error
    ...    No Topolvm LogicalVolumes Should Exist

Snapshot Retain Teardown
    [Documentation]    Clean up VolumeSnapshotContent, VolumeSnapshotClass, LogicalVolume, and namespace.
    [Arguments]    ${vsc_name}    ${content_name}
    ${lv_name}=    Replace String    ${content_name}    content    shot
    Run Keyword And Ignore Error
    ...    Delete VolumeSnapshotContent    ${content_name}
    Run Keyword And Ignore Error
    ...    Delete LogicalVolume    ${lv_name}
    Run Keyword And Ignore Error
    ...    Delete VolumeSnapshotClass    ${vsc_name}
    Run Keyword And Ignore Error
    ...    Remove Namespace    ${NAMESPACE}
    Run Keyword And Ignore Error
    ...    No Topolvm LogicalVolumes Should Exist

Snapshot Different SC Teardown
    [Documentation]    Clean up snapshot, alternative SC, VolumeSnapshotClass, and namespace.
    [Arguments]    ${vsc_name}    ${snap_name}    ${alt_sc_name}
    Run Keyword And Ignore Error
    ...    Oc Delete    volumesnapshot ${snap_name} -n ${NAMESPACE}
    Run Keyword And Ignore Error
    ...    Delete StorageClass    ${alt_sc_name}
    Run Keyword And Ignore Error
    ...    Delete VolumeSnapshotClass    ${vsc_name}
    Run Keyword And Ignore Error
    ...    Remove Namespace    ${NAMESPACE}
    Run Keyword And Ignore Error
    ...    No Topolvm LogicalVolumes Should Exist

VolumeSnapshotContent Should Still Exist
    [Documentation]    Verify VolumeSnapshotContent still exists consistently over 30 seconds.
    [Arguments]    ${content_name}
    FOR    ${i}    IN RANGE    6
        Sleep    5s
        ${output}=    Run With Kubeconfig    oc get volumesnapshotcontent ${content_name}
        Should Contain    ${output}    ${content_name}
    END
