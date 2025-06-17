*** Settings ***
Documentation       Tests for Workload partitioning

Resource            ../../resources/microshift-config.resource
Resource            ../../resources/common.resource
Resource            ../../resources/systemd.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/microshift-network.resource
Resource            ../../resources/ostree-health.resource

Suite Setup         Setup Suite And Wait For Greenboot
Suite Teardown      Teardown Suite


*** Variables ***
${MANAGEMENT_CPU}                       0
${SYSTEMD_CRIO_DROPIN}                  /etc/systemd/system/crio.service.d/microshift-cpuaffinity.conf
${SYSTEMD_MICROSHIFT_DROPIN}            /etc/systemd/system/microshift.service.d/microshift-cpuaffinity.conf
${SYSTEMD_OVS_DROPIN}                   /etc/systemd/system/ovs-vswitchd.service.d/microshift-cpuaffinity.conf
${SYSTEMD_OVSDB_DROPIN}                 /etc/systemd/system/ovsdb-server.service.d/microshift-cpuaffinity.conf
${CRIO_CONFIG_DROPIN}                   /etc/crio/crio.conf.d/20-microshift-wp.conf
${KUBELET_CPU_STATE_FILE}               /var/lib/kubelet/cpu_manager_state
${KUBELET_WORKLOAD_PINNING_CONFIG}      /etc/kubernetes/openshift-workload-pinning


*** Test Cases ***
Control Plane Pods Must Be Annotated
    [Documentation]    Verify that all the Control Plane pods are properly annotated.
    All Pods Should Be Annotated As Management

Workload Partitioning Should Work
    [Documentation]    Verify that all the Control Plane pods are properly annotated.
    [Setup]    Setup For Workload Partitioning
    Microshift Services Should Be Running On Reserved CPU    ${MANAGEMENT_CPU}
    All Pods Should Run On Reserved CPU    ${MANAGEMENT_CPU}
    Create Hello MicroShift Pod
    All Pods Should Run On Reserved CPU    ${MANAGEMENT_CPU}    true
    [Teardown]    Teardown For Workload Partitioning


*** Keywords ***
Setup For Workload Partitioning
    [Documentation]    Setup for Workload Partitioning
    Configure Kubelet For Workload Partitioning    ${MANAGEMENT_CPU}
    Configure CRIO For Workload Partitioning    ${MANAGEMENT_CPU}
    Configure CPUAffinity In Systemd    ${MANAGEMENT_CPU}    ${SYSTEMD_CRIO_DROPIN}
    Configure CPUAffinity In Systemd    ${MANAGEMENT_CPU}    ${SYSTEMD_MICROSHIFT_DROPIN}
    Configure CPUAffinity In Systemd    ${MANAGEMENT_CPU}    ${SYSTEMD_OVS_DROPIN}
    Configure CPUAffinity In Systemd    ${MANAGEMENT_CPU}    ${SYSTEMD_OVSDB_DROPIN}
    Systemctl Daemon Reload
    Systemctl    restart    crio.service
    Cleanup And Create NS

Teardown For Workload Partitioning
    [Documentation]    Setup for Workload Partitioning
    Run Keyword And Ignore Error
    ...    SSHLibrary.Get File    ${DEBUG_OUTPUT_FILE}    ${OUTPUTDIR}/pod-crio-inspect-output.json
    Cleanup MicroShift    --all    --keep-images
    Remove Files    ${KUBELET_CPU_STATE_FILE}
    ...    ${SYSTEMD_CRIO_DROPIN}
    ...    ${SYSTEMD_MICROSHIFT_DROPIN}
    ...    ${CRIO_CONFIG_DROPIN}
    Systemctl Daemon Reload
    Remove Drop In MicroShift Config    10-kubelet
    Systemctl    restart    crio.service
    Restart MicroShift

Configure Kubelet For Workload Partitioning
    [Documentation]    configure microshift with kubelet CPU configuration
    [Arguments]    ${cpus}

    ${kubelet_config}=    CATENATE    SEPARATOR=\n
    ...    ---
    ...    kubelet:
    ...    \ \ reservedSystemCPUs: "${cpus}"
    ...    \ \ cpuManagerPolicy: static
    ...    \ \ cpuManagerPolicyOptions:
    ...    \ \ \ \ full-pcpus-only: "true"
    ...    \ \ cpuManagerReconcilePeriod: 5s

    Drop In MicroShift Config    ${kubelet_config}    10-kubelet

    ${kubelet_pinning_config}=    CATENATE    SEPARATOR=\n
    ...    {
    ...    \ \ "management": {
    ...    \ \ \ \ "cpuset": "${cpus}"
    ...    \ \ }
    ...    }

    Upload String To File    ${kubelet_pinning_config}    ${KUBELET_WORKLOAD_PINNING_CONFIG}

Configure CRIO For Workload Partitioning
    [Documentation]    add crio Dropin configuration
    [Arguments]    ${cpus}
    ${crio_configuration}=    CATENATE    SEPARATOR=\n
    ...    [crio.runtime]
    ...    infra_ctr_cpuset = "${cpus}"
    ...    [crio.runtime.workloads.management]
    ...    activation_annotation = "target.workload.openshift.io/management"
    ...    annotation_prefix = "resources.workload.openshift.io"
    ...    resources = { "cpushares" = 0, "cpuset" = "${cpus}" }

    Upload String To File    ${crio_configuration}    ${CRIO_CONFIG_DROPIN}

Configure CPUAffinity In Systemd
    [Documentation]    add CPUAffinity in systemd unit dropin
    [Arguments]    ${cpus}    ${target_file}
    Create Remote Dir For Path    ${target_file}
    ${systemd_cpu_affinity}=    CATENATE    SEPARATOR=\n
    ...    [Service]
    ...    CPUAffinity=${cpus}
    Upload String To File    ${systemd_cpu_affinity}    ${target_file}

All Pods Should Run On Reserved CPU
    [Documentation]    Verify all the PODs runs on explicit reserved CPU
    [Arguments]    ${configured_cpus}    ${is_workloads}=${EMPTY}
    ${crio_json}=    Get Json From Crio Output    ${is_workloads}

    ${json_status}=    Json Parse    ${crio_json}
    FOR    ${pod}    IN    @{json_status}
        ${pod_id}    ${pod_pid}    ${pod_cpuset}    ${pod_info}=    Construct Pod Info    ${pod}

        IF    "${is_workloads}"=="${EMPTY}"
            IF    ${pod_cpuset} != ${configured_cpus}
                Crio Save Pod Manifest    ${pod_id}
                Fail
                ...    Management ${pod_info} running on CPU ${pod_cpuset} instead of ${configured_cpus}
            END
            Proccess Should Be Running On Host CPU    ${pod_pid}    ${configured_cpus}
        ELSE IF    "${pod_cpuset}" == "${configured_cpus}"
            Crio Save Pod Manifest    ${pod_id}
            Fail
            ...    Workload ${pod_info} running on CPU ${pod_cpuset} instead of ${configured_cpus}
        END
    END

Construct Pod Info
    [Documentation]    json
    [Arguments]    ${pod_json}
    ${pod_cpuset}=    Evaluate    "${pod_json}[info][runtimeSpec][linux][resources][cpu][cpus]"
    ${pod_name}=    Evaluate    "${pod_json}[status][metadata][name]"
    ${container_name}=    Evaluate    "${pod_json}[info][runtimeSpec][annotations][io.kubernetes.pod.name]"
    ${pod_pid}=    Evaluate    "${pod_json}[info][pid]"
    ${pod_id}=    Evaluate    "${pod_json}[status][id]"
    ${namespace}=    Evaluate    "${pod_json}[info][runtimeSpec][annotations][io.kubernetes.pod.namespace]"
    ${pod_info}=    Catenate    SEPARATOR=
    ...    container: ${container_name}
    ...    ${EMPTY} pod: ${pod_name}
    ...    ${EMPTY} pid: ${pod_pid}
    ...    ${EMPTY} namespace: ${namespace}
    RETURN    ${pod_id}    ${pod_pid}    ${pod_cpuset}    ${pod_info}

Get Json From Crio Output
    [Documentation]    get json from the crio command
    [Arguments]    ${is_workloads}
    Set Global Variable    ${NOT_WORKLOADS}    ${EMPTY}
    IF    "${is_workloads}"!="${EMPTY}"
        Set Global Variable    ${NOT_WORKLOADS}    | not
    END

    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    crictl ps -q | xargs sudo crictl inspect | jq -rs '[.[][] | select(.info.runtimeSpec.annotations["target.workload.openshift.io/management"] ${NOT_WORKLOADS})]'
    ...    sudo=True
    ...    return_stdout=True
    ...    return_stderr=True
    ...    return_rc=True
    Should Be Equal As Integers    ${rc}    0
    RETURN    ${stdout}

Crio Save Pod Manifest
    [Documentation]    Saves running pod manifest using crio inspect
    [Arguments]    ${pod_id}
    ${path}=    Create Random Temp File
    Set Global Variable    ${DEBUG_OUTPUT_FILE}    ${path}
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    crictl ps -q | xargs sudo crictl inspect | jq -rs '[.[][] | select(.status.id=="${pod_id}")]' >${DEBUG_OUTPUT_FILE} 2>&1
    ...    sudo=True
    ...    return_stdout=True
    ...    return_stderr=True
    ...    return_rc=True

Microshift Services Should Be Running On Reserved CPU
    [Documentation]    Verify all the Microshift Services runs on explicit reserved CPU
    [Arguments]    ${cpus}
    ${pid}=    MicroShift Process ID
    Proccess Should Be Running On Host CPU    ${pid}    ${cpus}

    ${pid}=    Crio Process ID
    Proccess Should Be Running On Host CPU    ${pid}    ${cpus}

    ${pid}=    MicroShift Etcd Process ID
    Proccess Should Be Running On Host CPU    ${pid}    ${cpus}

Proccess Should Be Running On Host CPU
    [Documentation]    Verify all the PODs runs on explicit reserved CPU
    [Arguments]    ${pid}    ${cpus}
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    taskset -cp ${pid} | awk '{print $6}'
    ...    sudo=True    return_stdout=True    return_stderr=True    return_rc=True
    Should Be Equal As Integers    ${rc}    0
    Should Be Equal As Strings    ${stdout}    ${cpus}

All Pods Should Be Annotated As Management
    [Documentation]    Obtains list of Deployments created by CSV.
    ${pods_raw}=    Oc Get All Pods
    @{pods}=    Split String    ${pods_raw}
    Set Test Variable    @{NS_TO_SKIP_LIST}    openshift-gateway-api    redhat-ods-applications
    FOR    ${pod}    IN    @{pods}
        ${ns}    ${pod}=    Split String    ${pod}    \@
        IF    "${ns}" not in "@{NS_TO_SKIP_LIST}"
            Pod Must Be Annotated    ${ns}    ${pod}
        ELSE
            Log    ${ns}@${pod} pod annotation check skipped because is not a MicroShift core pod
        END
    END

Pod Must Be Annotated
    [Documentation]    Check management annotation for specified pod and namespace.
    [Arguments]    ${ns}    ${pod}
    ${management_annotation}=    Oc Get JsonPath
    ...    pod
    ...    ${ns}
    ...    ${pod}
    ...    .metadata.annotations.target\\.workload\\.openshift\\.io/management
    Should Not Be Empty    ${management_annotation}

Oc Get All Pods
    [Documentation]    Returns the running pods across all namespaces,
    ...    Returns the command output as formatted string <name-space>@<pod-name>

    ${data}=    Oc Get JsonPath
    ...    pods
    ...    ${EMPTY}
    ...    ${EMPTY}
    ...    range .items[*]}{\.metadata\.namespace}{"@"}{\.metadata\.name}{"\\n"}{end
    RETURN    ${data}

Crio Process ID
    [Documentation]    Return the current crio process ID
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    pidof crio
    ...    return_stdout=True    return_stderr=True    return_rc=True
    Log    ${stderr}
    RETURN    ${stdout}

Remove Files
    [Documentation]    removes files from the microshit host
    [Arguments]    @{files}
    Log    ${files}
    ${files_path}=    Catenate    SEPARATOR=${SPACE}    @{files}
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    rm -f ${files_path}
    ...    sudo=True    return_stdout=True    return_stderr=True    return_rc=True
    Should Be Equal As Integers    ${rc}    0

Cleanup And Create NS
    [Documentation]    cleanup microshift and recreate the namespace for workloads
    Cleanup MicroShift    --all    --keep-images
    Remove Files    ${KUBELET_CPU_STATE_FILE}
    Restart MicroShift
    ${ns}=    Create Unique Namespace
    Set Suite Variable    \${NAMESPACE}    ${ns}

Setup Suite And Wait For Greenboot
    [Documentation]    Run setup suit and wait for greenboot to become ready
    Setup Suite
    Restart Greenboot And Wait For Success
