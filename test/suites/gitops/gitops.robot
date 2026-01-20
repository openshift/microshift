*** Settings ***
Documentation       MicroShift GitOps tests

Resource            ../../resources/microshift-process.resource

Suite Setup         Setup Suite
Suite Teardown      Teardown Suite


*** Variables ***
${APPLICATION_MANIFEST_PATH}    ${CURDIR}/spring-petclinic-app.yaml
${APPLICATION_NAMESPACE}        spring-petclinic
${APPLICATION_NAME}             spring-petclinic
${GITOPS_NAMESPACE}             openshift-gitops


*** Test Cases ***
Verify GitOps Pods Start Correctly
    [Documentation]    Waits for pods to enter a running state

    Wait Until Keyword Succeeds    2min    10s
    ...    All Pods Should Be Running    openshift-gitops

Verify Application Deployed Correctly
    [Documentation]    Deploys an application and waits for it to be Healthy
    ...    using the example from official docs: https://docs.redhat.com/en/documentation/red_hat_build_of_microshift/4.20/html/running_applications/microshift-gitops#microshift-gitops-adding-apps_microshift-gitops
    [Setup]    Setup Application Deployment

    Wait Until Resource Exists    applications    ${APPLICATION_NAME}    ${GITOPS_NAMESPACE}    timeout=120s
    Oc Wait
    ...    -n ${GITOPS_NAMESPACE} application ${APPLICATION_NAME}
    ...    --for=jsonpath='{.status.sync.status}'=Synced --timeout=300s
    Oc Wait
    ...    -n ${APPLICATION_NAMESPACE} pod --selector=app=${APPLICATION_NAME}
    ...    --for=condition=Ready --timeout=300s

    [Teardown]    Teardown Application Deployment


*** Keywords ***
Setup Application Deployment
    [Documentation]    Setup the application deployment
    Oc Apply    -f ${APPLICATION_MANIFEST_PATH}

Teardown Application Deployment
    [Documentation]    Teardown the application deployment
    Oc Delete    -f ${APPLICATION_MANIFEST_PATH}
    Oc Delete    ns ${APPLICATION_NAMESPACE}
