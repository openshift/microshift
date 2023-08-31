*** Settings ***
Documentation       Tests related to upgrading MicroShift on a non-ostree system

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-rpm.resource
Resource            ../../resources/microshift-host.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/ostree-data.resource
Resource            ../../resources/ostree-health.resource
Library             Collections

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           rpm-based-system


*** Variables ***
${TARGET_REPO_URL}      ${EMPTY}


*** Test Cases ***
Upgrade
    [Documentation]    Performs an upgrade from a given RPM repository URL
    ...    and checks if it was successful

    # Health of the system is implicitly checked by greenboot successful exit
    MicroShift 413 Should Not Have Upgrade Artifacts
    Wait Until Greenboot Health Check Exited

    # Upgrade MicroShift and reboot
    Install MicroShift RPM Packages    ${TARGET_REPO_URL}
    Reboot MicroShift Host

    # Health of the system is implicitly checked by greenboot successful exit
    Wait Until Greenboot Health Check Exited


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Should Not Be Empty    ${TARGET_REPO_URL}    TARGET_REPO_URL variable is required
    Login MicroShift Host

    # Make sure we run on a non-ostree system
    ${is_ostree}=    Is System OSTree
    Should Not Be True    ${is_ostree}

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host
