*** Settings ***
Documentation       Tests related to a functional PVC after reboot

Resource            ../../resources/common.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/ostree-health.resource

Suite Setup         Setup Suite
Suite Teardown      Teardown Suite


*** Variables ***
${SOURCE_POD}           ./assets/reboot/pod-with-pvc.yaml
${POD_NAME_STATIC}      test-pod
${RESIZE_TO}            2Gi
${PVC_CLAIM_NAME}       test-claim


*** Test Cases ***
Increase Running Pod PV Size
    [Documentation]    Increase Running Pod PV size by changing the PVC spec.
    [Setup]    Test Case Setup
    Oc Patch    pvc/${PVC_CLAIM_NAME}    '{"spec":{"resources":{"requests":{"storage":"${RESIZE_TO}"}}}}'
    Named PVC Should Be Resized    ${PVC_CLAIM_NAME}    ${RESIZE_TO}
    [Teardown]    Test Case Teardown


*** Keywords ***
Test Case Setup
    [Documentation]    Prepare the cluster env and test pod workload.
    ${ns}=    Create Unique Namespace
    VAR    ${NAMESPACE}=    ${ns}    scope=SUITE
    Oc Create    -f ${SOURCE_POD} -n ${NAMESPACE}
    Named Pod Should Be Ready    ${POD_NAME_STATIC}

Test Case Teardown
    [Documentation]    Clean up test suite resources
    Oc Delete    -f ${SOURCE_POD} -n ${NAMESPACE}
    Remove Namespace    ${NAMESPACE}
