*** Settings ***
Documentation       Shared keywords for MicroShift storage tests migrated from OTP.

Library             OperatingSystem
Library             String
Library             StorageTemplates.py
Resource            common.resource
Resource            oc.resource
Resource            kubeconfig.resource
Resource            microshift-host.resource


*** Variables ***
${TOPOLVM_PROVISIONER}      topolvm.io
${DEFAULT_SC}               topolvm-provisioner
${HELLO_IMAGE}              quay.io/openshifttest/hello-openshift@sha256:56c354e7885051b6bb4263f9faa58b2c292d44790599b7dde0e49e7c466cf339
${DEFAULT_MOUNT_PATH}       /mnt/storage
${PVC_TPL}                  assets/storage/pvc-template.yaml
${PVC_NO_SC_TPL}            assets/storage/pvc-no-sc-template.yaml
${PVC_SNAPSHOT_TPL}         assets/storage/pvc-snapshot-template.yaml
${PVC_CLONE_TPL}            assets/storage/pvc-clone-template.yaml
${DEPLOYMENT_TPL}           assets/storage/deployment-template.yaml
${DEPLOYMENT_NO_VOL_TPL}    assets/storage/deployment-no-volume-template.yaml
${POD_TPL}                  assets/storage/pod-template.yaml
${POD_BLOCK_TPL}            assets/storage/pod-block-template.yaml
${POD_SELINUX_TPL}          assets/storage/pod-selinux-template.yaml
${POD_EPHEMERAL_TPL}        assets/storage/pod-ephemeral-template.yaml
${SC_TPL}                   assets/storage/storageclass-template.yaml
${STS_TPL}                  assets/storage/statefulset-template.yaml
${VS_TPL}                   assets/storage/volumesnapshot-template.yaml
${VSC_TPL}                  assets/storage/volumesnapshotclass-template.yaml
${RQ_NS_TPL}                assets/storage/resourcequota-namespace-template.yaml
${RQ_SC_TPL}                assets/storage/resourcequota-storageclass-template.yaml


*** Keywords ***
Create PVC
    [Documentation]    Create a PersistentVolumeClaim with the given parameters.
    [Arguments]    ${name}    ${sc}    ${capacity}    ${ns}=${NAMESPACE}
    ...    ${volume_mode}=Filesystem    ${access_mode}=ReadWriteOnce
    ${file}=    Render Template    ${PVC_TPL}
    ...    PVCNAME=${name}    SCNAME=${sc}    PVCCAPACITY=${capacity}
    ...    VOLUMEMODE=${volume_mode}    ACCESSMODE=${access_mode}
    Apply Rendered Template    ${file}    -n ${ns}

Create PVC Without StorageClass
    [Documentation]    Create a PVC without specifying a storageClassName.
    [Arguments]    ${name}    ${capacity}    ${ns}=${NAMESPACE}
    ...    ${volume_mode}=Filesystem    ${access_mode}=ReadWriteOnce
    ${file}=    Render Template    ${PVC_NO_SC_TPL}
    ...    PVCNAME=${name}    PVCCAPACITY=${capacity}
    ...    VOLUMEMODE=${volume_mode}    ACCESSMODE=${access_mode}
    Apply Rendered Template    ${file}    -n ${ns}

Create PVC From Snapshot
    [Documentation]    Create a PVC that restores from a VolumeSnapshot dataSource.
    [Arguments]    ${name}    ${sc}    ${capacity}    ${snapshot_name}    ${ns}=${NAMESPACE}
    ...    ${volume_mode}=Filesystem
    ${file}=    Render Template    ${PVC_SNAPSHOT_TPL}
    ...    PVCNAME=${name}    SCNAME=${sc}    PVCCAPACITY=${capacity}
    ...    SNAPSHOTNAME=${snapshot_name}    VOLUMEMODE=${volume_mode}
    Apply Rendered Template    ${file}    -n ${ns}

Create PVC From Clone
    [Documentation]    Create a PVC cloned from another PVC.
    [Arguments]    ${name}    ${sc}    ${capacity}    ${source_pvc}    ${ns}=${NAMESPACE}
    ...    ${volume_mode}=Filesystem
    ${file}=    Render Template    ${PVC_CLONE_TPL}
    ...    PVCNAME=${name}    SCNAME=${sc}    PVCCAPACITY=${capacity}
    ...    SOURCEPVC=${source_pvc}    VOLUMEMODE=${volume_mode}
    Apply Rendered Template    ${file}    -n ${ns}

Create Deployment
    [Documentation]    Create a Deployment with a PVC-backed volume mount.
    [Arguments]    ${name}    ${pvc_name}    ${ns}=${NAMESPACE}
    ...    ${mount_path}=${DEFAULT_MOUNT_PATH}    ${replicas}=1
    ${file}=    Render Template    ${DEPLOYMENT_TPL}
    ...    DNAME=${name}    PVCNAME=${pvc_name}    MOUNTPATH=${mount_path}
    ...    REPLICAS=${replicas}    IMAGE=${HELLO_IMAGE}
    Apply Rendered Template    ${file}    -n ${ns}

Create Deployment Without Volume
    [Documentation]    Create a Deployment with no volume mounts.
    [Arguments]    ${name}    ${ns}=${NAMESPACE}    ${replicas}=1
    ${file}=    Render Template    ${DEPLOYMENT_NO_VOL_TPL}
    ...    DNAME=${name}    REPLICAS=${replicas}    IMAGE=${HELLO_IMAGE}
    Apply Rendered Template    ${file}    -n ${ns}

Create Pod
    [Documentation]    Create a Pod with a PVC-backed filesystem volume mount.
    [Arguments]    ${name}    ${pvc_name}    ${ns}=${NAMESPACE}
    ...    ${mount_path}=${DEFAULT_MOUNT_PATH}
    ${file}=    Render Template    ${POD_TPL}
    ...    PODNAME=${name}    PVCNAME=${pvc_name}    MOUNTPATH=${mount_path}
    ...    IMAGE=${HELLO_IMAGE}
    Apply Rendered Template    ${file}    -n ${ns}

Create Pod With Block Volume
    [Documentation]    Create a Pod with a PVC-backed raw block volume device.
    [Arguments]    ${name}    ${pvc_name}    ${ns}=${NAMESPACE}
    ...    ${device_path}=/dev/dblock
    ${file}=    Render Template    ${POD_BLOCK_TPL}
    ...    PODNAME=${name}    PVCNAME=${pvc_name}    DEVICEPATH=${device_path}
    ...    IMAGE=${HELLO_IMAGE}
    Apply Rendered Template    ${file}    -n ${ns}

Create Pod With SELinux Context
    [Documentation]    Create a Pod with SELinux securityContext and a PVC-backed volume.
    [Arguments]    ${name}    ${pvc_name}    ${ns}=${NAMESPACE}
    ...    ${mount_path}=${DEFAULT_MOUNT_PATH}    ${selinux_level}=s0:c345,c789
    ${file}=    Render Template    ${POD_SELINUX_TPL}
    ...    PODNAME=${name}    PVCNAME=${pvc_name}    MOUNTPATH=${mount_path}
    ...    SELINUXLEVEL=${selinux_level}    IMAGE=${HELLO_IMAGE}
    Apply Rendered Template    ${file}    -n ${ns}

Create Pod With Ephemeral Volume
    [Documentation]    Create a Pod with a generic ephemeral volume.
    [Arguments]    ${name}    ${sc}    ${capacity}    ${ns}=${NAMESPACE}
    ...    ${mount_path}=${DEFAULT_MOUNT_PATH}
    ${file}=    Render Template    ${POD_EPHEMERAL_TPL}
    ...    PODNAME=${name}    SCNAME=${sc}    PVCCAPACITY=${capacity}
    ...    MOUNTPATH=${mount_path}    IMAGE=${HELLO_IMAGE}
    Apply Rendered Template    ${file}    -n ${ns}

Create StorageClass
    [Documentation]    Create a StorageClass with the given parameters.
    [Arguments]    ${name}    ${provisioner}=${TOPOLVM_PROVISIONER}
    ...    ${binding_mode}=WaitForFirstConsumer    ${reclaim_policy}=Delete
    ...    ${parameters}=${EMPTY}    ${allow_expansion}=true
    ${file}=    Render Template    ${SC_TPL}
    ...    SCNAME=${name}    PROVISIONER=${provisioner}    BINDINGMODE=${binding_mode}
    ...    RECLAIMPOLICY=${reclaim_policy}    ALLOWEXPANSION=${allow_expansion}
    IF    "${parameters}" != "${EMPTY}"
        ${content}=    OperatingSystem.Get File    ${file}
        ${content}=    Catenate    SEPARATOR=\n    ${content}    ${parameters}
        Create File    ${file}    ${content}
    END
    TRY
        Run With Kubeconfig    oc apply -f ${file}
    FINALLY
        Cleanup Rendered File    ${file}
    END

Create StatefulSet
    [Documentation]    Create a StatefulSet with volumeClaimTemplates.
    [Arguments]    ${name}    ${sc}    ${capacity}    ${ns}=${NAMESPACE}
    ...    ${replicas}=3    ${mount_path}=${DEFAULT_MOUNT_PATH}
    ${file}=    Render Template    ${STS_TPL}
    ...    STSNAME=${name}    SCNAME=${sc}    PVCCAPACITY=${capacity}
    ...    REPLICAS=${replicas}    MOUNTPATH=${mount_path}    IMAGE=${HELLO_IMAGE}
    Apply Rendered Template    ${file}    -n ${ns}

Create VolumeSnapshot
    [Documentation]    Create a VolumeSnapshot from a PVC.
    [Arguments]    ${name}    ${pvc_name}    ${vsc_name}    ${ns}=${NAMESPACE}
    ${file}=    Render Template    ${VS_TPL}
    ...    VSNAME=${name}    SOURCEPVCNAME=${pvc_name}    VSCNAME=${vsc_name}
    Apply Rendered Template    ${file}    -n ${ns}

Create VolumeSnapshotClass
    [Documentation]    Create a VolumeSnapshotClass.
    [Arguments]    ${name}    ${driver}=${TOPOLVM_PROVISIONER}    ${deletion_policy}=Delete
    ${file}=    Render Template    ${VSC_TPL}
    ...    VSCNAME=${name}    DRIVER=${driver}    DELETIONPOLICY=${deletion_policy}
    Apply Rendered Template    ${file}

Create Namespace ResourceQuota
    [Documentation]    Create a namespace-scoped ResourceQuota for storage.
    [Arguments]    ${name}    ${hard_storage}    ${pvc_limit}    ${ns}=${NAMESPACE}
    ${file}=    Render Template    ${RQ_NS_TPL}
    ...    QUOTANAME=${name}    HARDSTORAGE=${hard_storage}    PVCLIMIT=${pvc_limit}
    Apply Rendered Template    ${file}    -n ${ns}

Create StorageClass ResourceQuota
    [Documentation]    Create a StorageClass-scoped ResourceQuota for storage.
    [Arguments]    ${name}    ${sc_name}    ${hard_storage}    ${pvc_limit}    ${ns}=${NAMESPACE}
    ${file}=    Render Template    ${RQ_SC_TPL}
    ...    QUOTANAME=${name}    SCNAME=${sc_name}
    ...    HARDSTORAGE=${hard_storage}    PVCLIMIT=${pvc_limit}
    Apply Rendered Template    ${file}    -n ${ns}

Apply Rendered Template
    [Documentation]    Apply a rendered template file and clean it up, even on failure.
    [Arguments]    ${file}    ${extra_opts}=${EMPTY}
    TRY
        Run With Kubeconfig    oc apply -f ${file} ${extra_opts}
    FINALLY
        Cleanup Rendered File    ${file}
    END

PVC Create Should Fail
    [Documentation]    Attempt to create a PVC and expect it to fail. Returns the error output.
    [Arguments]    ${name}    ${sc}    ${capacity}    ${ns}=${NAMESPACE}
    ...    ${volume_mode}=Filesystem
    ${file}=    Render Template    ${PVC_TPL}
    ...    PVCNAME=${name}    SCNAME=${sc}    PVCCAPACITY=${capacity}
    ...    VOLUMEMODE=${volume_mode}    ACCESSMODE=ReadWriteOnce
    TRY
        ${output}    ${rc}=    Run With Kubeconfig    oc apply -f ${file} -n ${ns}
        ...    allow_fail=True    return_rc=True
    FINALLY
        Cleanup Rendered File    ${file}
    END
    Should Not Be Equal As Integers    ${rc}    0
    ...    PVC creation was expected to fail but succeeded with: ${output}
    RETURN    ${output}

Volume Should Be RW
    [Documentation]    Write a test string to a file on the mounted volume and read it back.
    [Arguments]    ${pod}    ${mount_path}=${DEFAULT_MOUNT_PATH}    ${ns}=${NAMESPACE}
    ${rand}=    Generate Random String
    Oc Exec    ${pod}    echo "storage test ${rand}" > ${mount_path}/testfile    ns=${ns}
    ${output}=    Oc Exec    ${pod}    cat ${mount_path}/testfile    ns=${ns}
    Should Contain    ${output}    storage test ${rand}

Volume Should Have Exec Right
    [Documentation]    Write a shell script to the volume, make it executable, and run it.
    [Arguments]    ${pod}    ${mount_path}=${DEFAULT_MOUNT_PATH}    ${ns}=${NAMESPACE}
    Oc Exec    ${pod}    printf "#!/bin/sh\\necho hello\\n" > ${mount_path}/hello    ns=${ns}
    Oc Exec    ${pod}    chmod +x ${mount_path}/hello    ns=${ns}
    ${output}=    Oc Exec    ${pod}    ${mount_path}/hello    ns=${ns}
    Should Contain    ${output}    hello

Volume Should Contain Fstype
    [Documentation]    Check that the mounted volume uses the expected filesystem type.
    [Arguments]    ${pod}    ${expected_fstype}    ${mount_path}=${DEFAULT_MOUNT_PATH}    ${ns}=${NAMESPACE}
    ${output}=    Oc Exec    ${pod}    mount | grep ${mount_path}    ns=${ns}
    Should Contain    ${output}    ${expected_fstype}

Write Data To Block Volume
    [Documentation]    Write data to a raw block device on a pod.
    [Arguments]    ${pod}    ${device_path}=/dev/dblock    ${ns}=${NAMESPACE}
    Oc Exec    ${pod}    dd if=/dev/null of=${device_path} bs=512 count=1    ns=${ns}
    Oc Exec    ${pod}    echo "storage test" > ${device_path}    ns=${ns}

Verify Data In Block Volume
    [Documentation]    Read and verify data from a raw block device on a pod.
    [Arguments]    ${pod}    ${device_path}=/dev/dblock    ${ns}=${NAMESPACE}
    Oc Exec    ${pod}    dd if=${device_path} of=/tmp/testfile bs=512 count=1    ns=${ns}
    ${output}=    Oc Exec    ${pod}    cat /tmp/testfile    ns=${ns}
    Should Contain    ${output}    storage test

Scale Deployment And Wait
    [Documentation]    Scale a deployment to the given replica count and wait for rollout.
    [Arguments]    ${name}    ${replicas}    ${ns}=${NAMESPACE}    ${timeout}=${DEFAULT_WAIT_TIMEOUT}
    Run With Kubeconfig    oc scale deployment ${name} --replicas\=${replicas} -n ${ns}
    IF    ${replicas} > 0
        Named Deployment Should Be Available    ${name}    ns=${ns}    timeout=${timeout}
    ELSE
        Wait Until Keyword Succeeds    ${timeout}    5s
        ...    Deployment Has Zero Replicas    ${name}    ${ns}
    END

Deployment Has Zero Replicas
    [Documentation]    Check that a deployment has 0 available replicas.
    [Arguments]    ${name}    ${ns}=${NAMESPACE}
    ${count}=    Oc Get JsonPath    deployment    ${ns}    ${name}    .status.availableReplicas
    Should Be Equal    ${count}    ${EMPTY}

Set StorageClass As Default
    [Documentation]    Patch a StorageClass to be the default.
    [Arguments]    ${sc_name}
    Run With Kubeconfig
    ...    oc patch sc ${sc_name} -p '{"metadata":{"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'

Set StorageClass As Non Default
    [Documentation]    Patch a StorageClass to be non-default.
    [Arguments]    ${sc_name}
    Run With Kubeconfig
    ...    oc patch sc ${sc_name} -p '{"metadata":{"annotations":{"storageclass.kubernetes.io/is-default-class":"false"}}}'

Patch PV Reclaim Policy
    [Documentation]    Patch a PV's persistentVolumeReclaimPolicy.
    [Arguments]    ${pv_name}    ${policy}
    Run With Kubeconfig
    ...    oc patch pv ${pv_name} -p '{"spec":{"persistentVolumeReclaimPolicy":"${policy}"}}'

Get PV Name From PVC
    [Documentation]    Return the PV name bound to a PVC.
    [Arguments]    ${pvc_name}    ${ns}=${NAMESPACE}
    ${pv}=    Oc Get JsonPath    pvc    ${ns}    ${pvc_name}    .spec.volumeName
    Should Not Be Empty    ${pv}
    RETURN    ${pv}

Get Pod Node Name
    [Documentation]    Return the node name where a pod is scheduled.
    [Arguments]    ${pod}    ${ns}=${NAMESPACE}
    ${node}=    Oc Get JsonPath    pod    ${ns}    ${pod}    .spec.nodeName
    Should Not Be Empty    ${node}
    RETURN    ${node}

Get Deployment Pod Name
    [Documentation]    Return the first pod name from a deployment.
    [Arguments]    ${dep_name}    ${ns}=${NAMESPACE}
    ${pod}=    Run With Kubeconfig
    ...    oc get pods -n ${ns} -l app\=${dep_name} -o jsonpath='{.items[0].metadata.name}'
    Should Not Be Empty    ${pod}
    RETURN    ${pod}

Wait For PVC Bound
    [Documentation]    Wait for a PVC to reach Bound status.
    [Arguments]    ${pvc_name}    ${ns}=${NAMESPACE}    ${timeout}=${DEFAULT_WAIT_TIMEOUT}
    Oc Wait    -n ${ns} pvc/${pvc_name}
    ...    --for=jsonpath='{.status.phase}'=Bound --timeout=${timeout}

PVC Should Stay Pending
    [Documentation]    Assert that a PVC stays in Pending status for approximately 90 seconds.
    [Arguments]    ${pvc_name}    ${ns}=${NAMESPACE}
    Wait Until Keyword Succeeds    30s    5s
    ...    PVC Status Should Be    ${pvc_name}    Pending    ${ns}
    FOR    ${i}    IN RANGE    6
        Sleep    10s
        PVC Status Should Be    ${pvc_name}    Pending    ${ns}
    END

PVC Status Should Be
    [Documentation]    Assert the PVC status phase matches expected.
    [Arguments]    ${pvc_name}    ${expected}    ${ns}=${NAMESPACE}
    ${status}=    Oc Get JsonPath    pvc    ${ns}    ${pvc_name}    .status.phase
    Should Be Equal    ${status}    ${expected}

PVC Description Should Contain
    [Documentation]    Run oc describe on a PVC and check for expected substring.
    [Arguments]    ${pvc_name}    ${expected}    ${ns}=${NAMESPACE}
    ${output}=    Run With Kubeconfig    oc describe pvc ${pvc_name} -n ${ns}
    Should Contain    ${output}    ${expected}

PV Should Be In Status
    [Documentation]    Assert the PV status phase matches expected.
    [Arguments]    ${pv_name}    ${expected}
    ${status}=    Oc Get JsonPath    pv    ${EMPTY}    ${pv_name}    .status.phase
    Should Be Equal    ${status}    ${expected}

Resource Should Not Exist
    [Documentation]    Verify that a resource returns NotFound.
    [Arguments]    ${type}    ${name}    ${ns}=${NAMESPACE}
    ${output}=    Run With Kubeconfig    oc get ${type} ${name} -n ${ns}    allow_fail=True
    Should Contain    ${output}    NotFound

Resource Should Not Exist Cluster Scoped
    [Documentation]    Verify that a cluster-scoped resource returns NotFound.
    [Arguments]    ${type}    ${name}
    ${output}=    Run With Kubeconfig    oc get ${type} ${name}    allow_fail=True
    Should Contain    ${output}    NotFound

Check Volume Mount On Node
    [Documentation]    Verify a volume is mounted on the MicroShift host with expected content.
    [Arguments]    ${vol_name}    ${node_name}    ${expected}
    ${output}=    Command Should Work    mount | grep ${vol_name}
    Should Contain    ${output}    ${expected}

Check Volume Not Mounted On Node
    [Documentation]    Verify a volume is NOT mounted on the MicroShift host.
    [Arguments]    ${vol_name}    ${node_name}
    ${stdout}    ${stderr}    ${rc}=    SSHLibrary.Execute Command
    ...    mount | grep ${vol_name}
    ...    sudo=True    return_rc=True    return_stdout=True    return_stderr=True
    Should Not Be Equal As Integers    ${rc}    0

Wait For StatefulSet Ready
    [Documentation]    Wait for a StatefulSet to have all replicas ready.
    [Arguments]    ${name}    ${replicas}    ${ns}=${NAMESPACE}    ${timeout}=120s
    Oc Wait    -n ${ns} statefulset/${name}
    ...    --for=jsonpath='{.status.readyReplicas}'=${replicas} --timeout=${timeout}

StatefulSet PVC Count Should Match
    [Documentation]    Verify that the number of PVCs matches the StatefulSet replica count.
    [Arguments]    ${sts_name}    ${expected_count}    ${ns}=${NAMESPACE}
    ${output}=    Run With Kubeconfig
    ...    oc get pvc -n ${ns} -l app\=${sts_name} -o jsonpath='{.items[*].metadata.name}'
    @{pvcs}=    Split String    ${output}
    ${count}=    Get Length    ${pvcs}
    Should Be Equal As Integers    ${count}    ${expected_count}

StatefulSet Volume Should Be RW
    [Documentation]    Check RW on each pod in a StatefulSet.
    [Arguments]    ${sts_name}    ${replicas}    ${mount_path}=${DEFAULT_MOUNT_PATH}    ${ns}=${NAMESPACE}
    FOR    ${i}    IN RANGE    ${replicas}
        ${pod}=    Set Variable    ${sts_name}-${i}
        Volume Should Be RW    ${pod}    ${mount_path}    ${ns}
    END

StatefulSet Volume Should Have Exec Right
    [Documentation]    Check exec right on each pod in a StatefulSet.
    [Arguments]    ${sts_name}    ${replicas}    ${mount_path}=${DEFAULT_MOUNT_PATH}    ${ns}=${NAMESPACE}
    FOR    ${i}    IN RANGE    ${replicas}
        ${pod}=    Set Variable    ${sts_name}-${i}
        Volume Should Have Exec Right    ${pod}    ${mount_path}    ${ns}
    END

Get VolumeSnapshotContent Name
    [Documentation]    Return the VolumeSnapshotContent name for a given VolumeSnapshot.
    [Arguments]    ${vs_name}    ${ns}=${NAMESPACE}
    ${content}=    Oc Get JsonPath    volumesnapshot    ${ns}    ${vs_name}
    ...    .status.boundVolumeSnapshotContentName
    Should Not Be Empty    ${content}
    RETURN    ${content}

Deployment Volume Should Contain Data
    [Documentation]    Check that previously written testfile data exists on the deployment pod volume.
    [Arguments]    ${dep_name}    ${mount_path}=${DEFAULT_MOUNT_PATH}    ${ns}=${NAMESPACE}
    ${pod}=    Get Deployment Pod Name    ${dep_name}    ${ns}
    ${output}=    Oc Exec    ${pod}    cat ${mount_path}/testfile    ns=${ns}
    Should Contain    ${output}    storage test

Deployment Volume Should Not Contain Data
    [Documentation]    Check that testfile does NOT exist on the deployment pod volume.
    [Arguments]    ${dep_name}    ${mount_path}=${DEFAULT_MOUNT_PATH}    ${ns}=${NAMESPACE}
    ${pod}=    Get Deployment Pod Name    ${dep_name}    ${ns}
    ${output}=    Run With Kubeconfig
    ...    oc exec -n ${ns} pod/${pod} -- /bin/sh -c 'cat ${mount_path}/testfile 2>&1'
    ...    allow_fail=True
    Should Contain    ${output}    No such file

Delete StorageClass
    [Documentation]    Delete a StorageClass by name.
    [Arguments]    ${name}
    Run With Kubeconfig    oc delete sc ${name}    allow_fail=True

Delete VolumeSnapshotClass
    [Documentation]    Delete a VolumeSnapshotClass by name.
    [Arguments]    ${name}
    Run With Kubeconfig    oc delete volumesnapshotclass ${name}    allow_fail=True

Delete VolumeSnapshotContent
    [Documentation]    Delete a VolumeSnapshotContent by name.
    [Arguments]    ${name}
    Run With Kubeconfig    oc delete volumesnapshotcontent ${name}    allow_fail=True

Delete LogicalVolume
    [Documentation]    Delete a topolvm LogicalVolume by name.
    [Arguments]    ${name}
    Run With Kubeconfig    oc delete logicalvolume ${name}    allow_fail=True

No Topolvm LogicalVolumes Should Exist
    [Documentation]    Verify no topolvm LogicalVolume CRs remain in the cluster.
    ${output}=    Run With Kubeconfig
    ...    oc get logicalvolumes -o jsonpath='{.items[*].metadata.name}'
    ...    allow_fail=True
    Should Be Empty    ${output}
