*** Settings ***
Documentation       Tests for PVC label propagation in StatefulSets.
...
...                 Ported from openshift-tests-private:
...                 Medium-28018

Resource            ../../resources/common.resource
Resource            ../../resources/oc.resource

Suite Setup         Setup Suite
Suite Teardown      Teardown Suite


*** Variables ***
${STATEFULSET_YAML}     ./assets/otp-workloads/statefulset-pvc.yaml
${PVC_NAME}             www-hello-statefulset-0
${POD_NAME}             hello-statefulset-0


*** Test Cases ***
Custom Label For PVC In StatefulSets
    [Documentation]    Verify that custom labels defined in a StatefulSet
    ...    are propagated to the PVCs created by volumeClaimTemplates.
    ...    OCP-28018
    [Setup]    Create StatefulSet Resources

    Named Pod Should Be Ready    ${POD_NAME}    ns=${NAMESPACE}    timeout=5m
    Wait Until Keyword Succeeds    60s    5s
    ...    PVC Should Have Label    ${PVC_NAME}    ${NAMESPACE}    app    hello-pod

    [Teardown]    Remove StatefulSet Resources


*** Keywords ***
Setup Suite
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig
    Verify Default StorageClass Exists

Teardown Suite
    [Documentation]    Test suite teardown
    Remove Kubeconfig
    Logout MicroShift Host

Verify Default StorageClass Exists
    [Documentation]    Skip test if no default StorageClass is available.
    ${output}=    Oc Get JsonPath    sc    ${EMPTY}    ${EMPTY}
    ...    .items[?(@.metadata.annotations.storageclass\\.kubernetes\\.io/is-default-class=="true")].metadata.name
    Skip If    "${output}" == "${EMPTY}"
    ...    No default StorageClass found

Create StatefulSet Resources
    [Documentation]    Create namespace and deploy the StatefulSet.
    ${ns}=    Create Unique Namespace
    VAR    ${NAMESPACE}=    ${ns}    scope=SUITE
    Oc Create    -f ${STATEFULSET_YAML} -n ${NAMESPACE}

Remove StatefulSet Resources
    [Documentation]    Delete the StatefulSet and namespace.
    Run With Kubeconfig
    ...    oc delete -f ${STATEFULSET_YAML} -n ${NAMESPACE}    allow_fail=True
    Run With Kubeconfig
    ...    oc delete pvc --all -n ${NAMESPACE}    allow_fail=True
    Remove Namespace    ${NAMESPACE}

PVC Should Have Label
    [Documentation]    Verify that a PVC has the expected label.
    [Arguments]    ${pvc_name}    ${namespace}    ${label_key}    ${label_value}
    ${actual_value}=    Oc Get JsonPath    pvc    ${namespace}    ${pvc_name}
    ...    .metadata.labels.${label_key}
    Should Be Equal As Strings    ${actual_value}    ${label_value}
