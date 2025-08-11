*** Settings ***
Documentation       MicroShift GitOps tests

Resource            ../../resources/microshift-process.resource

Suite Setup         Setup
Suite Teardown      Teardown


*** Variables ***
${TMPDIR}    ${EMPTY}


*** Test Cases ***
Verify GitOps Pods Start Correctly
    [Documentation]    Waits for pods to enter a running state.

    Wait Until Keyword Succeeds    2min    10s
    ...    All Pods Should Be Running    openshift-gitops

Verify Workload Deployed Correctly
    [Documentation]    Deploys workload and waits for ready status.
    Deploy Guestbook


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

Deploy Guestbook
    [Documentation]    Deploy sample workload
    ${tmp}=    Set Variable    /tmp
    Set Global Variable    ${TMPDIR}    ${tmp}

    # should we verify gitops pods here again?
    ${cmd1}=    Catenate
    ...    curl
    ...    https://raw.githubusercontent.com/argoproj/argocd-example-apps/refs/heads/master/guestbook/guestbook-ui-deployment.yaml
    ...    -o
    ...    ${TMPDIR}/guestbook-ui-deployment.yaml

    ${cmd2}=    Catenate
    ...    curl
    ...    https://raw.githubusercontent.com/argoproj/argocd-example-apps/refs/heads/master/guestbook/guestbook-ui-svc.yaml
    ...    -o
    ...    ${TMPDIR}/guestbook-ui-svc.yaml

    Local Command Should Work    ${cmd1}
    Local Command Should Work    ${cmd2}
    Oc Apply    -f ${TMPDIR}/guestbook-ui-deployment.yaml
    Oc Apply    -f ${TMPDIR}/guestbook-ui-svc.yaml
    Wait Until Keyword Succeeds    2min    10s
    ...    Named Deployment Should Be Available    guestbook-ui    default
