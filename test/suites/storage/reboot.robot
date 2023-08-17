*** Settings ***
Documentation       Tests related to a functional PVC after reboot

Resource            ../../resources/common.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/ostree-health.resource

Suite Setup         Setup Suite With Namespace

Test Tags           restart    slow


*** Variables ***
${SOURCE_POD}           ./assets/reboot/pod-with-pvc.yaml
${POD_NAME_STATIC}      test-pod


*** Test Cases ***
Rebooting Healthy System Should Keep Functional PVC
    [Documentation]    Run a reboot test and make sure pod with PVC remains functional.
    [Setup]    Test Case Setup
    Reboot MicroShift Host
    Wait Until Greenboot Health Check Exited
    Oc Wait For    pod/${POD_NAME_STATIC}    condition\=Ready    timeout=60s
    [Teardown]    Test Case Teardown


*** Keywords ***
Test Case Setup
    [Documentation]    Prepare the cluster env and test pod workload.
    Oc Create    -f ${SOURCE_POD} -n ${NAMESPACE}
    Oc Wait For    pod/${POD_NAME_STATIC}    condition\=Ready    timeout=60s

Test Case Teardown
    [Documentation]    Clean up test suite resources
    Oc Delete    -f ${SOURCE_POD} -n ${NAMESPACE}
