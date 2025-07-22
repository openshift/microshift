*** Settings ***
Documentation       MicroShift GitOps tests

Resource            ../../resources/microshift-process.resource

Suite Setup          Setup
Suite Teardown       Teardown


*** Test Cases ***
Verify GitOps Pods Start Correctly After Restart
    [Documentation]    Restarts MicroShift and waits for pods to enter a running state.

    # Define the pods we need to check
    @{expected_pods}=    Create List
    ...    argocd-application-controller
    ...    argocd-redis
    ...    argocd-repo-server

    # Restart the service
    Restart MicroShift

    # Wait up to 2 minutes for our verification keyword to pass, retrying every 10s
    Wait Until Keyword Succeeds    2min    10s
    ...    Verify Pods Are Running    openshift-gitops    @{expected_pods}


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Login MicroShift Host
    Setup Kubeconfig

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host
    Remove Kubeconfig

Verify Pods Are Running
    [Arguments]    ${namespace}    @{pod_names}
    [Documentation]    Checks if a list of pods are all in a 'Running' state.

    # Get the current pod status
    ${stdout}    ${rc}=    Run With Kubeconfig    oc get pods -n ${namespace}    return_rc=${True}
    Should Be Equal As Integers    ${rc}    0    Failed to run "oc get pods".

    # Loop through each pod name and check its specific line for "Running"
    FOR    ${pod_name}    IN    @{pod_names}
        ${pod_line}=    Get Lines Containing String    ${stdout}    ${pod_name}
        ${line_count}=    Get Line Count    ${pod_line}
        Should Be Equal As Integers    ${line_count}    1    Could not find a unique entry for pod: ${pod_name}
        Should Contain    ${pod_line}    Running    Pod "${pod_name}" is not 'Running'. Got: ${pod_line}
    END
