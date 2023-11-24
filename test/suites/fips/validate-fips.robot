*** Settings ***
Documentation       Tests related to FIPS Validation

Resource            ../../resources/ostree-health.resource
Resource            ../../resources/common.resource
Resource            ../../resources/selinux.resource
Library             Collections

Suite Setup         Setup
Suite Teardown      Teardown


*** Variables ***
${USHIFT_HOST}      ${EMPTY}
${USHIFT_USER}      ${EMPTY}


*** Test Cases ***
Verify Host Is FIPS Enabled
    [Documentation]    Performs a FIPS validation against the host
    Wait Until Greenboot Health Check Exited
    Fips Should Be Enabled

Verify Binary Is FIPS Compliant
    [Documentation]    Performs a FIPS validation against the Microshift binary
    Microshift Binary Should Dynamically Link FIPS Ossl Module


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host

Microshift Binary Should Dynamically Link FIPS Ossl Module
    [Documentation]    Check if Microshift binary is FIPS compliant.
    ${stdout}    ${rc}=    Execute Command
    ...    LD_DEBUG=symbols microshift run 2>&1 | grep ossl-modules/fips.so$
    ...    sudo=False    return_rc=True
    Should Be Equal As Integers    0    ${rc}

Fips Should Be Enabled
    [Documentation]    Check if FIPS is enabled on RHEL.
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    bash -x fips-mode-setup --check
    ...    sudo=True    return_rc=True    return_stderr=True
    Should Be Equal As Integers    0    ${rc}
    Should Match    ${stdout}    FIPS mode is enabled.
