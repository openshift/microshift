*** Settings ***
Documentation       Storage provisioning and lifecycle tests migrated from OTP.

Library             String
Resource            ../../resources/common.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/kubeconfig.resource
Resource            ../../resources/storage.resource
Resource            ../../resources/microshift-host.resource

Suite Setup         Setup Suite
Suite Teardown      Teardown Suite


*** Variables ***
${SC_NAME}          topolvm-provisioner
${TOPOLVM_PROV}     topolvm.io
${MOUNT_PATH}       /mnt/storage


*** Test Cases ***
OCP-59668 Default SC XFS Dynamic Provision RW Exec Scale
    [Documentation]    Verify default StorageClass provisions xfs volumes that support
    ...    read/write, exec, and data persistence across scale-down/up.
    [Setup]    Test Case Setup
    ${pvc}=    Set Variable    pvc-59668
    ${dep}=    Set Variable    dep-59668

    # Verify default SC configuration
    ${is_default}=    Oc Get JsonPath    sc    ${EMPTY}    ${SC_NAME}
    ...    .metadata.annotations.storageclass\\.kubernetes\\.io/is-default-class
    Should Be Equal    ${is_default}    true
    ${reclaim}=    Oc Get JsonPath    sc    ${EMPTY}    ${SC_NAME}    .reclaimPolicy
    Should Be Equal    ${reclaim}    Delete
    ${binding}=    Oc Get JsonPath    sc    ${EMPTY}    ${SC_NAME}    .volumeBindingMode
    Should Be Equal    ${binding}    WaitForFirstConsumer
    ${fstype}=    Oc Get JsonPath    sc    ${EMPTY}    ${SC_NAME}
    ...    .parameters.csi\\.storage\\.k8s\\.io/fstype
    Should Be Equal    ${fstype}    xfs

    # Create PVC and deployment
    Create PVC    ${pvc}    ${SC_NAME}    1Gi
    Create Deployment    ${dep}    ${pvc}    mount_path=${MOUNT_PATH}
    Named Deployment Should Be Available    ${dep}

    # Verify xfs, RW, and exec
    ${pod}=    Get Deployment Pod Name    ${dep}
    Volume Should Contain Fstype    ${pod}    xfs    ${MOUNT_PATH}
    Volume Should Be RW    ${pod}    ${MOUNT_PATH}
    Volume Should Have Exec Right    ${pod}    ${MOUNT_PATH}

    # Verify volume mount on node
    ${pv}=    Get PV Name From PVC    ${pvc}
    ${node}=    Get Pod Node Name    ${pod}
    Check Volume Mount On Node    ${pv}    ${node}    xfs

    # Scale down and verify unmount
    Scale Deployment And Wait    ${dep}    0
    Check Volume Not Mounted On Node    ${pv}    ${node}

    # Scale up and verify data persists
    Scale Deployment And Wait    ${dep}    1
    Deployment Volume Should Contain Data    ${dep}    ${MOUNT_PATH}
    ${pod2}=    Get Deployment Pod Name    ${dep}
    Volume Should Have Exec Right    ${pod2}    ${MOUNT_PATH}

    [Teardown]    Teardown Test Case With Resources    deployment/${dep}    pvc/${pvc}

OCP-59655 WaitForFirstConsumer Binding Mode
    [Documentation]    Verify PVC stays Pending with WaitForFirstConsumer until a consumer pod is created.
    [Setup]    Test Case Setup
    ${sc}=    Set Variable    sc-59655
    ${pvc}=    Set Variable    pvc-59655
    ${dep}=    Set Variable    dep-59655

    Create StorageClass    ${sc}    binding_mode=WaitForFirstConsumer
    Create PVC    ${pvc}    ${sc}    1Gi

    # Verify PVC description mentions WFC
    Wait Until Keyword Succeeds    30s    5s
    ...    PVC Description Should Contain    ${pvc}    WaitForFirstConsumer

    # Verify PVC stays Pending
    PVC Should Stay Pending    ${pvc}

    # Create deployment — PVC should bind
    Create Deployment    ${dep}    ${pvc}    mount_path=${MOUNT_PATH}
    Named Deployment Should Be Available    ${dep}

    [Teardown]    Teardown Test Case With Resources    deployment/${dep}    pvc/${pvc}    sc:${sc}

OCP-59657 Immediate Binding Mode
    [Documentation]    Verify PVC binds immediately with Immediate volumeBindingMode.
    [Setup]    Test Case Setup
    ${sc}=    Set Variable    sc-59657
    ${pvc}=    Set Variable    pvc-59657
    ${dep}=    Set Variable    dep-59657

    Create StorageClass    ${sc}    binding_mode=Immediate
    Create PVC    ${pvc}    ${sc}    1Gi
    Wait For PVC Bound    ${pvc}

    Create Deployment    ${dep}    ${pvc}    mount_path=${MOUNT_PATH}
    Named Deployment Should Be Available    ${dep}

    ${pod}=    Get Deployment Pod Name    ${dep}
    Volume Should Be RW    ${pod}    ${MOUNT_PATH}
    Volume Should Have Exec Right    ${pod}    ${MOUNT_PATH}

    [Teardown]    Teardown Test Case With Resources    deployment/${dep}    pvc/${pvc}    sc:${sc}

OCP-59658 Block VolumeMode PVC
    [Documentation]    Verify a PVC with Block volumeMode creates a matching PV.
    [Setup]    Test Case Setup
    ${sc}=    Set Variable    sc-59658
    ${pvc}=    Set Variable    pvc-59658

    Create StorageClass    ${sc}    binding_mode=Immediate
    Create PVC    ${pvc}    ${sc}    1Gi    volume_mode=Block
    Wait For PVC Bound    ${pvc}

    ${pvc_mode}=    Oc Get JsonPath    pvc    ${NAMESPACE}    ${pvc}    .spec.volumeMode
    Should Be Equal    ${pvc_mode}    Block
    ${pv}=    Get PV Name From PVC    ${pvc}
    ${pv_mode}=    Oc Get JsonPath    pv    ${EMPTY}    ${pv}    .spec.volumeMode
    Should Be Equal    ${pv_mode}    Block

    [Teardown]    Teardown Test Case With Resources    pvc/${pvc}    sc:${sc}

OCP-59659 Filesystem VolumeMode PVC
    [Documentation]    Verify a PVC with Filesystem volumeMode creates a matching PV.
    [Setup]    Test Case Setup
    ${sc}=    Set Variable    sc-59659
    ${pvc}=    Set Variable    pvc-59659

    Create StorageClass    ${sc}    binding_mode=Immediate
    Create PVC    ${pvc}    ${sc}    1Gi    volume_mode=Filesystem
    Wait For PVC Bound    ${pvc}

    ${pvc_mode}=    Oc Get JsonPath    pvc    ${NAMESPACE}    ${pvc}    .spec.volumeMode
    Should Be Equal    ${pvc_mode}    Filesystem
    ${pv}=    Get PV Name From PVC    ${pvc}
    ${pv_mode}=    Oc Get JsonPath    pv    ${EMPTY}    ${pv}    .spec.volumeMode
    Should Be Equal    ${pv_mode}    Filesystem

    [Teardown]    Teardown Test Case With Resources    pvc/${pvc}    sc:${sc}

OCP-59660 Online Volume Resize
    [Documentation]    Verify PVC can be resized online while a pod is consuming it.
    [Tags]    serial
    [Setup]    Test Case Setup
    ${pvc}=    Set Variable    pvc-59660
    ${dep}=    Set Variable    dep-59660

    Create PVC    ${pvc}    ${SC_NAME}    1Gi
    Create Deployment    ${dep}    ${pvc}    mount_path=${MOUNT_PATH}
    Named Deployment Should Be Available    ${dep}

    # Write data before resize
    ${pod}=    Get Deployment Pod Name    ${dep}
    Volume Should Be RW    ${pod}    ${MOUNT_PATH}

    # Resize PVC
    Oc Patch    pvc/${pvc}    '{"spec":{"resources":{"requests":{"storage":"5Gi"}}}}'
    Named PVC Should Be Resized    ${pvc}    5Gi    timeout=120s

    # Verify original data survives resize
    Deployment Volume Should Contain Data    ${dep}    ${MOUNT_PATH}

    [Teardown]    Teardown Test Case With Resources    deployment/${dep}    pvc/${pvc}

OCP-59661 StatefulSet Volumes RW And Exec
    [Documentation]    Verify StatefulSet with 3 replicas provisions PVCs that support RW and exec.
    [Setup]    Test Case Setup
    ${sts}=    Set Variable    sts-59661

    Create StatefulSet    ${sts}    ${SC_NAME}    1Gi    replicas=3    mount_path=${MOUNT_PATH}
    Wait For StatefulSet Ready    ${sts}    3    timeout=120s

    StatefulSet PVC Count Should Match    ${sts}    3
    StatefulSet Volume Should Be RW    ${sts}    3    ${MOUNT_PATH}
    StatefulSet Volume Should Have Exec Right    ${sts}    3    ${MOUNT_PATH}

    [Teardown]    Teardown Test Case With StatefulSet    ${sts}    3

OCP-59662 Pod Mounts Multiple PVCs
    [Documentation]    Verify a deployment can mount multiple PVCs at different paths.
    [Setup]    Test Case Setup
    ${sc}=    Set Variable    sc-59662
    ${pvc1}=    Set Variable    pvc1-59662
    ${pvc2}=    Set Variable    pvc2-59662
    ${dep}=    Set Variable    dep-59662
    ${xfs_params}=    Catenate    SEPARATOR=\n
    ...    parameters:
    ...    ${SPACE}${SPACE}csi.storage.k8s.io/fstype: xfs

    Create StorageClass    ${sc}    binding_mode=Immediate    parameters=${xfs_params}
    Create PVC    ${pvc1}    ${sc}    1Gi
    Create PVC    ${pvc2}    ${sc}    1Gi

    Create Deployment Without Volume    ${dep}
    Named Deployment Should Be Available    ${dep}

    # Add PVCs to deployment via oc set volumes
    Run With Kubeconfig
    ...    oc set volumes deployment ${dep} --add --name\=vol1 -t pvc --claim-name\=${pvc1} -m /mnt/storage1 --overwrite -n ${NAMESPACE}
    Named Deployment Should Be Available    ${dep}    timeout=120s
    Run With Kubeconfig
    ...    oc set volumes deployment ${dep} --add --name\=vol2 -t pvc --claim-name\=${pvc2} -m /mnt/storage2 --overwrite -n ${NAMESPACE}
    Named Deployment Should Be Available    ${dep}    timeout=120s

    # Verify both mounts show xfs
    ${pod}=    Get Deployment Pod Name    ${dep}
    Volume Should Contain Fstype    ${pod}    xfs    /mnt/storage1
    Volume Should Contain Fstype    ${pod}    xfs    /mnt/storage2

    [Teardown]    Teardown Test Case With Resources    deployment/${dep}    pvc/${pvc1}    pvc/${pvc2}    sc:${sc}

OCP-59663 StorageClass Scoped Resource Quota
    [Documentation]    Verify StorageClass-scoped ResourceQuota limits PVC count and storage.
    [Setup]    Test Case Setup
    ${sc}=    Set Variable    sc-59663
    ${pvc1}=    Set Variable    pvc1-59663
    ${pvc2}=    Set Variable    pvc2-59663
    ${pvc3}=    Set Variable    pvc3-59663
    ${quota}=    Set Variable    quota-59663

    Create StorageClass    ${sc}    binding_mode=WaitForFirstConsumer
    Create StorageClass ResourceQuota    ${quota}    ${sc}    6Gi    2

    # PVC 1 and 2 should succeed
    Create PVC    ${pvc1}    ${sc}    2Gi
    Create PVC    ${pvc2}    ${sc}    2Gi

    # PVC 3 should exceed PVC count quota
    ${err}=    PVC Create Should Fail    ${pvc3}    ${sc}    2Gi
    Should Contain    ${err}    exceeded quota

    # Delete PVC 2, recreate with 5Gi should exceed storage quota
    Oc Delete    pvc/${pvc2} -n ${NAMESPACE}
    ${err2}=    PVC Create Should Fail    ${pvc2}    ${sc}    5Gi
    Should Contain    ${err2}    exceeded quota

    # Recreate PVC 2 at 4Gi should succeed (2+4 <= 6)
    Wait Until Keyword Succeeds    60s    5s
    ...    Create PVC    ${pvc2}    ${sc}    4Gi

    # PVC 3 on a different SC should succeed — quota not scoped to it
    Create PVC    ${pvc3}    ${SC_NAME}    8Gi

    [Teardown]    Teardown Test Case With Resources    pvc/${pvc1}    pvc/${pvc2}    pvc/${pvc3}    sc:${sc}

OCP-59664 Namespace Storage PVC Quota
    [Documentation]    Verify namespace-level ResourceQuota limits PVC count and storage.
    [Setup]    Test Case Setup
    ${pvc1}=    Set Variable    pvc1-59664
    ${pvc2}=    Set Variable    pvc2-59664
    ${pvc3}=    Set Variable    pvc3-59664
    ${quota}=    Set Variable    quota-59664

    Create Namespace ResourceQuota    ${quota}    6Gi    2

    # PVC 1 and 2 should succeed
    Create PVC    ${pvc1}    ${SC_NAME}    2Gi
    Create PVC    ${pvc2}    ${SC_NAME}    2Gi

    # PVC 3 should exceed PVC count quota
    ${err}=    PVC Create Should Fail    ${pvc3}    ${SC_NAME}    2Gi
    Should Contain    ${err}    exceeded quota

    # Delete PVC 2, recreate with 5Gi should exceed storage quota
    Oc Delete    pvc/${pvc2} -n ${NAMESPACE}
    ${err2}=    PVC Create Should Fail    ${pvc2}    ${SC_NAME}    5Gi
    Should Contain    ${err2}    exceeded quota

    # Recreate PVC 2 at 4Gi should succeed (2+4 <= 6)
    Wait Until Keyword Succeeds    60s    5s
    ...    Create PVC    ${pvc2}    ${SC_NAME}    4Gi

    [Teardown]    Teardown Test Case With Resources    pvc/${pvc1}    pvc/${pvc2}

OCP-59665 Delete Unconsumed PVC
    [Documentation]    Verify a PVC not consumed by a pod can be deleted successfully.
    [Setup]    Test Case Setup
    ${pvc}=    Set Variable    pvc-59665

    Create PVC    ${pvc}    ${SC_NAME}    1Gi

    # Verify description contains WFC and pvc-protection finalizer
    Wait Until Keyword Succeeds    30s    5s
    ...    PVC Description Should Contain    ${pvc}    WaitForFirstConsumer
    PVC Description Should Contain    ${pvc}    pvc-protection

    # Delete PVC and verify it is gone
    Oc Delete    pvc/${pvc} -n ${NAMESPACE}
    Named PVC Should Be Deleted    ${pvc}

    [Teardown]    Teardown Test Case Namespace Only

OCP-59666 Delete Active PVC Protection
    [Documentation]    Verify deleting a PVC in active use by a pod postpones deletion
    ...    and new pods consuming such PVC get FailedScheduling.
    [Setup]    Test Case Setup
    ${pvc}=    Set Variable    pvc-59666
    ${dep}=    Set Variable    dep-59666
    ${dep2}=    Set Variable    dep2-59666

    Create PVC    ${pvc}    ${SC_NAME}    1Gi
    Create Deployment    ${dep}    ${pvc}    mount_path=${MOUNT_PATH}
    Named Deployment Should Be Available    ${dep}

    # Attempt to delete PVC — should timeout (still in use)
    ${output}    ${rc}=    Run With Kubeconfig
    ...    oc delete pvc/${pvc} -n ${NAMESPACE} --timeout\=3s
    ...    allow_fail=True    return_rc=True
    Should Not Be Equal As Integers    ${rc}    0

    # Verify PVC stays Terminating
    Wait Until Keyword Succeeds    30s    5s
    ...    PVC Description Should Contain    ${pvc}    Terminating

    # New deployment using same PVC should get FailedScheduling
    Create Deployment    ${dep2}    ${pvc}    mount_path=${MOUNT_PATH}
    ${dep2_pod}=    Wait Until Keyword Succeeds    60s    5s
    ...    Get Deployment Pod Name    ${dep2}
    Wait Until Keyword Succeeds    30s    5s
    ...    Pod Should Have Event    ${dep2_pod}    FailedScheduling

    # Delete both deployments — PVC should then be released and deleted
    Run Keyword And Ignore Error
    ...    Oc Delete    deployment/${dep2} -n ${NAMESPACE}
    Run Keyword And Ignore Error
    ...    Oc Delete    deployment/${dep} -n ${NAMESPACE}
    Named PVC Should Be Deleted    ${pvc}    timeout=120s

    [Teardown]    Teardown Test Case Namespace Only

OCP-59667 Single Default SC And PVC Without SC Name
    [Documentation]    Verify cluster has exactly one default SC and PVCs without
    ...    specifying storageclass use the default SC.
    [Setup]    Test Case Setup
    ${pvc}=    Set Variable    pvc-59667
    ${pvc2}=    Set Variable    pvc2-59667
    ${dep}=    Set Variable    dep-59667
    ${dep2}=    Set Variable    dep2-59667
    ${sc}=    Set Variable    sc-59667

    # Verify exactly one default SC
    ${default_sc}=    Run With Kubeconfig
    ...    oc get sc -o jsonpath='{.items[?(@.metadata.annotations.storageclass\\.kubernetes\\.io/is-default-class=="true")].metadata.name}'
    @{sc_list}=    Split String    ${default_sc}
    ${sc_count}=    Get Length    ${sc_list}
    Should Be Equal As Integers    ${sc_count}    1

    # Create PVC without specifying SC
    Create PVC Without StorageClass    ${pvc}    1Gi
    Create Deployment    ${dep}    ${pvc}    mount_path=${MOUNT_PATH}
    Named Deployment Should Be Available    ${dep}

    ${pod}=    Get Deployment Pod Name    ${dep}
    Volume Should Be RW    ${pod}    ${MOUNT_PATH}
    Volume Should Have Exec Right    ${pod}    ${MOUNT_PATH}

    # Verify the PV got the default SC
    ${pv}=    Get PV Name From PVC    ${pvc}
    ${pv_sc}=    Oc Get JsonPath    pv    ${EMPTY}    ${pv}    .spec.storageClassName
    Should Be Equal    ${pv_sc}    ${default_sc}

    # Delete first dep and PVC
    Oc Delete    deployment/${dep} -n ${NAMESPACE}
    Oc Delete    pvc/${pvc} -n ${NAMESPACE}

    # Create new SC and make it default
    Create StorageClass    ${sc}    binding_mode=WaitForFirstConsumer
    Set StorageClass As Default    ${sc}

    # Create second PVC without SC — should use new default
    Create PVC Without StorageClass    ${pvc2}    1Gi
    Create Deployment    ${dep2}    ${pvc2}    mount_path=${MOUNT_PATH}
    Named Deployment Should Be Available    ${dep2}

    ${pv2}=    Get PV Name From PVC    ${pvc2}
    ${pv2_sc}=    Oc Get JsonPath    pv    ${EMPTY}    ${pv2}    .spec.storageClassName
    Should Be Equal    ${pv2_sc}    ${sc}

    [Teardown]    Teardown OCP 59667    ${dep2}    ${pvc2}    ${sc}

OCP-59669 Pod With SELinux SecurityContext
    [Documentation]    Verify pod with SELinux securityContext labels the mounted volume correctly.
    [Setup]    Test Case Setup
    ${pvc}=    Set Variable    pvc-59669
    ${pod}=    Set Variable    pod-59669

    Create PVC    ${pvc}    ${SC_NAME}    1Gi
    Create Pod With SELinux Context    ${pod}    ${pvc}    mount_path=${MOUNT_PATH}
    Named Pod Should Be Ready    ${pod}

    Volume Should Have Exec Right    ${pod}    ${MOUNT_PATH}
    Oc Exec    ${pod}    sync

    # Verify SELinux labels
    ${dir_ctx}=    Oc Exec    ${pod}    ls -lZd ${MOUNT_PATH}
    Should Contain    ${dir_ctx}    s0:c345,c789
    ${file_ctx}=    Oc Exec    ${pod}    ls -lZ ${MOUNT_PATH}/hello
    Should Contain    ${file_ctx}    s0:c345,c789

    [Teardown]    Teardown Test Case With Resources    pod/${pod}    pvc/${pvc}

OCP-59670 Change Default SC To Non Default
    [Documentation]    Verify an admin can change the default StorageClass to non-default.
    [Tags]    serial
    [Setup]    Test Case Setup

    # Verify exactly one default SC
    ${default_sc}=    Run With Kubeconfig
    ...    oc get sc -o jsonpath='{.items[?(@.metadata.annotations.storageclass\\.kubernetes\\.io/is-default-class=="true")].metadata.name}'
    @{sc_list}=    Split String    ${default_sc}
    ${sc_count}=    Get Length    ${sc_list}
    Should Be Equal As Integers    ${sc_count}    1

    # Set default SC to non-default
    Set StorageClass As Non Default    ${default_sc}

    [Teardown]    Teardown OCP 59670    ${default_sc}

OCP-59671 PV Reclaim Policy Retain
    [Documentation]    Verify changing PV reclaim policy to Retain keeps PV after PVC deletion.
    [Setup]    Test Case Setup
    ${pvc}=    Set Variable    pvc-59671
    ${dep}=    Set Variable    dep-59671

    Create PVC    ${pvc}    ${SC_NAME}    1Gi
    Create Deployment    ${dep}    ${pvc}    mount_path=${MOUNT_PATH}
    Named Deployment Should Be Available    ${dep}

    # Patch PV reclaim policy to Retain
    ${pv}=    Get PV Name From PVC    ${pvc}
    Patch PV Reclaim Policy    ${pv}    Retain

    # Delete deployment and PVC
    Oc Delete    deployment/${dep} -n ${NAMESPACE}
    Oc Delete    pvc/${pvc} -n ${NAMESPACE}
    Named PVC Should Be Deleted    ${pvc}

    # PV should stay in Released status
    Wait Until Keyword Succeeds    30s    5s
    ...    PV Should Be In Status    ${pv}    Released

    # Delete PV and logicalvolume to free VG space
    Run With Kubeconfig    oc delete pv ${pv}
    Wait Until Keyword Succeeds    30s    5s
    ...    Resource Should Not Exist Cluster Scoped    pv    ${pv}
    Delete LogicalVolume    ${pv}

    [Teardown]    Teardown Test Case Namespace Only

OCP-64231 Generic Ephemeral Volume
    [Documentation]    Verify pod with generic ephemeral volume auto-creates and auto-deletes PVC/PV.
    [Setup]    Test Case Setup
    ${pod}=    Set Variable    pod-64231

    Create Pod With Ephemeral Volume    ${pod}    ${SC_NAME}    1Gi    mount_path=${MOUNT_PATH}
    Named Pod Should Be Ready    ${pod}

    # Verify PVC auto-created with naming convention: {pod}-inline-volume
    ${pvc_name}=    Set Variable    ${pod}-inline-volume
    ${actual_pvc}=    Run With Kubeconfig
    ...    oc get pvc -n ${NAMESPACE} -l workloadName\=${pod} -o jsonpath='{.items[0].metadata.name}'
    Should Be Equal    ${actual_pvc}    ${pvc_name}

    # Verify ownerReference points to the pod
    ${owner}=    Oc Get JsonPath    pvc    ${NAMESPACE}    ${pvc_name}
    ...    .metadata.ownerReferences[?(@.kind=="Pod")].name
    Should Be Equal    ${owner}    ${pod}

    # Verify RW and exec
    Volume Should Be RW    ${pod}    ${MOUNT_PATH}
    Volume Should Have Exec Right    ${pod}    ${MOUNT_PATH}

    # Get PV name before deleting the pod
    ${pv}=    Get PV Name From PVC    ${pvc_name}

    # Delete pod and verify PVC and PV are auto-deleted
    Oc Delete    pod/${pod} -n ${NAMESPACE}
    Named Pod Should Be Deleted    ${pod}
    Wait Until Keyword Succeeds    60s    5s
    ...    Resource Should Not Exist    pvc    ${pvc_name}
    Wait Until Keyword Succeeds    60s    5s
    ...    Resource Should Not Exist Cluster Scoped    pv    ${pv}

    [Teardown]    Teardown Test Case Namespace Only

OCP-68580 Oc Set Volume Operations
    [Documentation]    Verify oc set volume commands for add, overwrite, and remove operations.
    [Setup]    Test Case Setup
    ${pvc1}=    Set Variable    pvc1-68580
    ${pvc2}=    Set Variable    pvc2-68580
    ${dep}=    Set Variable    dep-68580
    VAR    ${NEW_PVC}=    ${EMPTY}    scope=TEST

    # Create PVC1 and deployment with PVC1
    Create PVC    ${pvc1}    ${SC_NAME}    1Gi
    Create Deployment    ${dep}    ${pvc1}    mount_path=${MOUNT_PATH}
    Named Deployment Should Be Available    ${dep}

    # Create PVC2
    Create PVC    ${pvc2}    ${SC_NAME}    1Gi

    # Verify oc set volume --all lists PVC1
    ${result}=    Run With Kubeconfig
    ...    oc set volume deployment --all -n ${NAMESPACE}
    Should Contain    ${result}    ${dep}
    Should Contain    ${result}    ${pvc1}

    # Overwrite with PVC2
    ${result}=    Run With Kubeconfig
    ...    oc set volumes deployment ${dep} --add --name\=local -t pvc --claim-name\=${pvc2} --overwrite -n ${NAMESPACE}
    Should Contain    ${result}    volume updated
    Named Deployment Should Be Available    ${dep}    timeout=120s
    ${pod}=    Get Deployment Pod Name    ${dep}
    Volume Should Be RW    ${pod}    ${MOUNT_PATH}
    Oc Delete    pvc/${pvc1} -n ${NAMESPACE}    # PVC1 no longer used

    # Overwrite by creating a new volume with --claim-size
    ${result}=    Run With Kubeconfig
    ...    oc set volumes deployment ${dep} --add --name\=local -t pvc --claim-size\=2Gi --overwrite -n ${NAMESPACE}
    Should Contain    ${result}    volume updated
    Named Deployment Should Be Available    ${dep}    timeout=120s

    # Verify the new PVC is not PVC2
    ${new_pvc}=    Run With Kubeconfig
    ...    oc get deployment ${dep} -n ${NAMESPACE} -o jsonpath='{.spec.template.spec.volumes[0].persistentVolumeClaim.claimName}'
    Should Not Be Equal    ${new_pvc}    ${pvc2}
    ${pod2}=    Get Deployment Pod Name    ${dep}
    Volume Should Be RW    ${pod2}    ${MOUNT_PATH}
    Oc Delete    pvc/${pvc2} -n ${NAMESPACE}    # PVC2 no longer used

    # Change mount path
    ${result}=    Run With Kubeconfig
    ...    oc set volumes deployment ${dep} --add --name\=local -m /data/storage --overwrite -n ${NAMESPACE}
    Should Contain    ${result}    volume updated
    Named Deployment Should Be Available    ${dep}    timeout=120s
    Deployment Volume Should Contain Data    ${dep}    /data/storage

    # Remove volume
    ${result}=    Run With Kubeconfig
    ...    oc set volumes deployment ${dep} --remove --name\=local -n ${NAMESPACE}
    Should Contain    ${result}    volume updated
    Named Deployment Should Be Available    ${dep}    timeout=120s
    Deployment Volume Should Not Contain Data    ${dep}    /data/storage

    [Teardown]    Teardown OCP 68580    ${dep}    ${new_pvc}


*** Keywords ***
Test Case Setup
    [Documentation]    Create a unique namespace for the test case.
    ${ns}=    Create Unique Namespace
    VAR    ${NAMESPACE}=    ${ns}    scope=TEST

Teardown Test Case With Resources
    [Documentation]    Delete the listed namespaced resources and remove the namespace.
    ...    Resources prefixed with "sc:" are cluster-scoped StorageClasses.
    [Arguments]    @{resources}
    FOR    ${res}    IN    @{resources}
        IF    "${res}".startswith("sc:")
            ${sc_name}=    Replace String    ${res}    sc:    ${EMPTY}
            Run Keyword And Ignore Error
            ...    Delete StorageClass    ${sc_name}
        ELSE
            Run Keyword And Ignore Error
            ...    Oc Delete    ${res} -n ${NAMESPACE}
        END
    END
    Run Keyword And Ignore Error
    ...    Remove Namespace    ${NAMESPACE}

Teardown Test Case Namespace Only
    [Documentation]    Remove the test namespace only.
    Run Keyword And Ignore Error
    ...    Remove Namespace    ${NAMESPACE}

Teardown Test Case With StatefulSet
    [Documentation]    Delete a StatefulSet and its PVCs, then remove the namespace.
    [Arguments]    ${sts_name}    ${replicas}
    Run Keyword And Ignore Error
    ...    Oc Delete    statefulset/${sts_name} -n ${NAMESPACE}
    FOR    ${i}    IN RANGE    ${replicas}
        Run Keyword And Ignore Error
        ...    Oc Delete    pvc/data-${sts_name}-${i} -n ${NAMESPACE}
    END
    Run Keyword And Ignore Error
    ...    Remove Namespace    ${NAMESPACE}

Teardown OCP 59667
    [Documentation]    Teardown for OCP-59667: clean up second SC default annotation.
    [Arguments]    ${dep2}    ${pvc2}    ${sc}
    Run Keyword And Ignore Error
    ...    Set StorageClass As Non Default    ${sc}
    Run Keyword And Ignore Error
    ...    Oc Delete    deployment/${dep2} -n ${NAMESPACE}
    Run Keyword And Ignore Error
    ...    Oc Delete    pvc/${pvc2} -n ${NAMESPACE}
    Run Keyword And Ignore Error
    ...    Delete StorageClass    ${sc}
    Run Keyword And Ignore Error
    ...    Remove Namespace    ${NAMESPACE}

Teardown OCP 59670
    [Documentation]    Teardown for OCP-59670: restore default SC.
    [Arguments]    ${sc_name}
    Run Keyword And Ignore Error
    ...    Set StorageClass As Default    ${sc_name}
    Run Keyword And Ignore Error
    ...    Remove Namespace    ${NAMESPACE}

Teardown OCP 68580
    [Documentation]    Teardown for OCP-68580: clean up deployment and dynamically created PVC.
    [Arguments]    ${dep}    ${last_pvc}
    Run Keyword And Ignore Error
    ...    Oc Delete    deployment/${dep} -n ${NAMESPACE}
    Run Keyword And Ignore Error
    ...    Oc Delete    pvc/${last_pvc} -n ${NAMESPACE}
    Run Keyword And Ignore Error
    ...    Remove Namespace    ${NAMESPACE}

Pod Should Have Event
    [Documentation]    Verify a pod has an event with the given reason.
    [Arguments]    ${pod_name}    ${reason}
    ${output}=    Run With Kubeconfig    oc describe pod ${pod_name} -n ${NAMESPACE}
    Should Contain    ${output}    ${reason}
