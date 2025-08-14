*** Settings ***
Documentation       MicroShift GitOps tests

Resource            ../../resources/microshift-process.resource
Library             RequestsLibrary

Suite Setup         Setup
Suite Teardown      Teardown


*** Variables ***
${URL}      https://raw.githubusercontent.com/argoproj/argocd-example-apps/refs/heads/master/guestbook/guestbook-ui-deployment.yaml


*** Test Cases ***
Verify GitOps Pods Start Correctly
    [Documentation]    Waits for pods to enter a running state

    Wait Until Keyword Succeeds    2min    10s
    ...    All Pods Should Be Running    openshift-gitops

Verify Workload Deployed Correctly
    [Documentation]    Deploys workload and waits for ready status
    Deploy Guestbook And Verify


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Login MicroShift Host
    Setup Kubeconfig
    Restart MicroShift

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host
    Remove Kubeconfig

Deploy Guestbook And Verify
    [Documentation]    Deploys Guestbook app as test workload
    Oc Apply    -f ${URL}
    Wait Until Keyword Succeeds    2min    10s
    ...    Named Deployment Should Be Available    guestbook-ui    default
