*** Settings ***
Documentation       Tests related to MicroShift running in an isolated network

Resource            ../../resources/ostree-health.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/microshift-host.resource
Library             Collections

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           network


*** Variables ***
${USHIFT_HOST}      ${EMPTY}
${LB_CONFIG}        assets/isolated-lb-service.yaml
${LB_NSPACE}        openshift-ingress
${LB_SRVNAME}       isolated-lb-service
${LB_PORTNUM}       31111


*** Test Cases ***
Verify Load Balancer Services Are Running Correctly In Isolated Network
    [Documentation]    Verifies that isolated network does not negatively
    ...    impact Load Balancer Services.

    Setup Kubeconfig
    Create Load Balancer
    Wait Until Keyword Succeeds    1m    5s
    ...    Check Load Balancer Service Access

    [Teardown]    Run Keywords
    ...    Delete Load Balancer

Verify MicroShift Is Healthy In Isolated Network After Clean And Reboot
    [Documentation]    Makes sure that MicroShift running in an isolated network
    ...    remains healthy after clean and reboot.

    Cleanup MicroShift
    Enable MicroShift
    Reboot MicroShift Host

    Verify No Internet Access
    Wait Until Greenboot Health Check Exited


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host

    Verify No Internet Access
    Wait Until Greenboot Health Check Exited

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host

Verify No Internet Access
    [Documentation]    Verifies that Internet is not accessible
    ${rc}=    Execute Command
    ...    curl -I redhat.com quay.io registry.redhat.io
    ...    return_stdout=False    return_stderr=False    return_rc=True
    Should Not Be Equal As Integers    ${rc}    0

Create Load Balancer
    [Documentation]    Creates a load balancer service backed by router pods.
    Run With Kubeconfig    oc create -f ${LB_CONFIG} -n ${LB_NSPACE}

Check Load Balancer Service Access
    [Documentation]    Checks if a load balancer service can be accessed.
    ${rc}=    Run And Return RC    curl -I ${USHIFT_HOST}:${LB_PORTNUM}
    Should Be Equal As Integers    ${rc}    0

Delete Load Balancer
    [Documentation]    Deletes a load balancer service backed by router pods.
    Run With Kubeconfig    oc delete svc/${LB_SRVNAME} -n ${LB_NSPACE}
