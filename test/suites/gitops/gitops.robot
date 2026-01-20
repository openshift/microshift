*** Settings ***
Documentation       MicroShift GitOps tests

Resource            ../../resources/microshift-process.resource

Suite Setup         Setup Suite
Suite Teardown      Teardown Suite


*** Test Cases ***
Verify GitOps Pods Start Correctly
    [Documentation]    Waits for pods to enter a running state

    Wait Until Keyword Succeeds    2min    10s
    ...    All Pods Should Be Running    openshift-gitops

Verify Workload Deployed Correctly
    [Documentation]    Deploys an application and waits for it to be Healthy
    ...    using the example from official docs: https://docs.redhat.com/en/documentation/red_hat_build_of_microshift/4.20/html/running_applications/microshift-gitops#microshift-gitops-adding-apps_microshift-gitops

    VAR    ${manifest_path}=    ${CURDIR}/spring-petclinic-app.yaml
    Oc Apply    -f ${manifest_path}
    Wait Until Resource Exists    applications    openshift-gitops    spring-petclinic
    Oc Wait
    ...    -n openshift-gitops application spring-petclinic
    ...    --for=jsonpath='{.status.sync.status}'=Synced --timeout=300s
