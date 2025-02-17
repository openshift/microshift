*** Settings ***
Documentation       Tests related to upgrading MicroShift

Resource            ../../resources/common.resource
Resource            ../../resources/selinux.resource
Library             Collections

Suite Setup         Setup
Suite Teardown      Teardown


*** Variables ***
${USHIFT_HOST}      ${EMPTY}
${USHIFT_USER}      ${EMPTY}


*** Test Cases ***
Standard SELinux Validation
    [Documentation]    Performs a standard SELinux validation against the host
    Validate SELinux


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host
