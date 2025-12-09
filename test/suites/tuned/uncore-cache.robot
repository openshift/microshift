*** Settings ***
Documentation       Tests for Workload partitioning

Library             ../../resources/journalctl.py
Resource            ../../resources/microshift-config.resource
Resource            ../../resources/common.resource
Resource            ../../resources/microshift-process.resource

Suite Setup         Run Keywords
...                     Setup Suite
...                     AND    Check FeatureGates Is Enabled
Suite Teardown      Teardown Suite


*** Variables ***
${MANAGEMENT_CPU}               0
${KUBELET_CPU_STATE_FILE}       /var/lib/kubelet/cpu_manager_state


*** Test Cases ***
Workload Partitioning Should Work With Uncore-cache
    [Documentation]    Verify that all the Control Plane pods are properly annotated.
    [Setup]    Configure Kubelet For Uncore-Cache    ${MANAGEMENT_CPU}
    Verify Uncore-cache Feature Is Enabled
    [Teardown]    Teardown For Workload Partitioning With Uncore-cache


*** Keywords ***
Teardown For Workload Partitioning With Uncore-cache
    [Documentation]    Teardown for Workload Partitioning with Uncore-cache
    Remove Drop In MicroShift Config    11-kubelet-uncore-cache
    Cleanup CPU State
    Restart MicroShift

Configure Kubelet For Uncore-Cache
    [Documentation]    configure microshift with kubelet CPU configuration
    [Arguments]    ${cpus}

    ${kubelet_config}=    CATENATE    SEPARATOR=\n
    ...    ---
    ...    apiServer:
    ...    \ \ featureGates:
    ...    \ \ \ \ featureSet: "CustomNoUpgrade"
    ...    \ \ \ \ customNoUpgrade:
    ...    \ \ \ \ \ \ enabled: ["CPUManagerPolicyBetaOptions"]
    ...    kubelet:
    ...    \ \ reservedSystemCPUs: "${cpus}"
    ...    \ \ cpuManagerPolicy: static
    ...    \ \ cpuManagerPolicyOptions:
    ...    \ \ \ \ prefer-align-cpus-by-uncorecache: "true"
    Drop In MicroShift Config    ${kubelet_config}    11-kubelet-uncore-cache
    Cleanup CPU State

Verify Uncore-cache Feature Is Enabled
    [Documentation]    Verify that the kubelet uncore-cache feature is enabled
    ${cursor}=    Get Journal Cursor
    Restart MicroShift
    Pattern Should Appear In Log Output
    ...    ${cursor}
    ...    kube-apiserver I.*CPUManagerPolicyBetaOptions=true
    ...    unit=microshift
    ...    wait=5
    Pattern Should Appear In Log Output
    ...    ${cursor}
    ...    kubelet I.*CPUManagerPolicyBetaOptions:true
    ...    unit=microshift
    ...    wait=5
    Pattern Should Appear In Log Output
    ...    ${cursor}
    ...    kubelet I.*prefer-align-cpus-by-uncorecache":"true"
    ...    unit=microshift
    ...    wait=5

Cleanup CPU State
    [Documentation]    cleanup microshift and recreate the namespace for workloads
    Cleanup MicroShift    --all    --keep-images
    Remove Files    ${KUBELET_CPU_STATE_FILE}

Check FeatureGates Is Enabled
    [Documentation]    Skip suite if FeatureGates feature is not available
    ${config}=    Show Config    default
    TRY
        VAR    ${featuregates}=    ${config}[apiServer][featureGates]
    EXCEPT
        Skip    FeatureGates feature not available in this MicroShift version
    END
