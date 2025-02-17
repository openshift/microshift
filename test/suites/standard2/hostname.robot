*** Settings ***
Documentation       Tests verifying hostname resolution

Resource            ../../resources/microshift-process.resource
Resource            ../../resources/microshift-host.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/ostree-health.resource

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           restart    slow


*** Variables ***
${NEW_HOSTNAME}     microshift.local
${OLD_HOSTNAME}     ${EMPTY}


*** Test Cases ***
Verify local name resolution
    [Documentation]    Verify correct name resolution through mDNS
    [Setup]    Configure New Hostname

    Named Deployment Should Be Available    router-default    timeout=120s    ns=openshift-ingress
    Oc Logs    namespace="openshift-ingress"    opts="deployment/router-default"

    [Teardown]    Restore Old Hostname


*** Keywords ***
Setup
    [Documentation]    Set up all of the tests in this suite
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig

Teardown
    [Documentation]    Test suite teardown
    Remove Kubeconfig
    Logout MicroShift Host

Configure New Hostname
    [Documentation]    Configures ${NEW_HOSTNAME} in the MicroShift host.
    ${old}=    Setup Hostname    ${NEW_HOSTNAME}
    Set Suite Variable    \${OLD_HOSTNAME}    ${old}

Restore Old Hostname
    [Documentation]    Configure old hostname again in the MicroShift host.
    Setup Hostname    ${OLD_HOSTNAME}

Setup Hostname
    [Documentation]    Setup a new hostname and return the old one.
    [Arguments]    ${hostname}
    IF    "${hostname}"=="${EMPTY}"    RETURN
    ${old}=    Change Hostname    ${hostname}
    Cleanup MicroShift    --all    --keep-images
    Start MicroShift
    Restart Greenboot And Wait For Success
    RETURN    ${old}
