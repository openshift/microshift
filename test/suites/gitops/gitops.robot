*** Settings ***
Documentation       MicroShift GitOps tests

Resource            ../../resources/microshift-process.resource
Library             RequestsLibrary

Suite Setup         Setup
Suite Teardown      Teardown


*** Variables ***
${TMPDIR}       ${EMPTY}
${URL}          https://raw.githubusercontent.com/argoproj/argocd-example-apps/refs/heads/master/guestbook/guestbook-ui-deployment.yaml


*** Test Cases ***
Verify GitOps Pods Start Correctly
    [Documentation]    Waits for pods to enter a running state

    Wait Until Keyword Succeeds    2min    10s
    ...    All Pods Should Be Running    openshift-gitops

Verify Workload Deployed Correctly
    [Documentation]    Deploys workload and waits for ready status
    Deploy Guestbook


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Login MicroShift Host
    Setup Kubeconfig
    Restart MicroShift
    ${tmp}=    Create Random Temp Directory
    VAR    ${TMPDIR}=    ${tmp}    scope=GLOBAL

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host
    Remove Kubeconfig
    Remove Directory    ${TMPDIR}    recursive=${True}

Deploy Guestbook
    [Documentation]    Deploys Guestbook app as test workload
    VAR    ${file_path}=    ${TMPDIR}/guestbook-ui-deployment.yaml
    Download File    ${URL}    ${file_path}
    Oc Apply    -f ${file_path}
    Wait Until Keyword Succeeds    2min    10s
    ...    Named Deployment Should Be Available    guestbook-ui    default

Download File
    [Documentation]    Downloads and saves a file
    [Arguments]    ${url}    ${save_path}
    ${response}=    GET    ${url}
    Status Should Be    200    ${response}
    Create Binary File    ${save_path}    ${response.content}
    OperatingSystem.File Should Exist    ${save_path}
