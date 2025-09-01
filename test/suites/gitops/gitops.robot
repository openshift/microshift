*** Settings ***
Documentation       MicroShift GitOps tests

Resource            ../../resources/microshift-process.resource

Suite Setup         Setup Suite With Namespace
Suite Teardown      Teardown Suite With Namespace


*** Test Cases ***
Verify GitOps Pods Start Correctly
    [Documentation]    Waits for pods to enter a running state

    Wait Until Keyword Succeeds    2min    10s
    ...    All Pods Should Be Running    openshift-gitops

Verify Workload Deployed Correctly
    [Documentation]    Deploys workload and waits for ready status

    VAR    ${manifest_path}=    ${CURDIR}/test-deployment.yaml
    Oc Apply    -f ${manifest_path} -n ${NAMESPACE}
    Wait Until Keyword Succeeds    5min    10s
    ...    Named Deployment Should Be Available    test-app
