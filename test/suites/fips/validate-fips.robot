*** Settings ***
Documentation       Tests related to FIPS Validation

Resource            ../../resources/ostree-health.resource
Resource            ../../resources/common.resource
Resource            ../../resources/selinux.resource
Resource            ../../resources/microshift-process.resource
Library             Collections

Suite Setup         Setup
Suite Teardown      Teardown


*** Variables ***
${USHIFT_HOST}              ${EMPTY}
${USHIFT_USER}              ${EMPTY}
${USHIFT_LIBS_DUMP_FILE}    /tmp/microshift-libs
${FIPS_PATTERN}             ossl-modules/fips.so$


*** Test Cases ***
Verify Host Is FIPS Enabled
    [Documentation]    Performs a FIPS validation against the host
    Fips Should Be Enabled

Verify Binary Is FIPS Compliant
    [Documentation]    Performs a FIPS validation against the Microshift binary
    Microshift Binary Should Dynamically Link FIPS Ossl Module


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig
    Wait Until Greenboot Health Check Exited
    Stop MicroShift

Teardown
    [Documentation]    Test suite teardown
    # Download the binary Libs dump files to the artifacts
    Run Keyword And Ignore Error
    ...    SSHLibrary.Get File    ${USHIFT_LIBS_DUMP_FILE}*    ${OUTPUTDIR}/
    Start MicroShift
    Wait For MicroShift
    Logout MicroShift Host

Microshift Binary Should Dynamically Link FIPS Ossl Module
    [Documentation]    Check if Microshift binary is FIPS compliant.
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    LD_DEBUG_OUTPUT=${USHIFT_LIBS_DUMP_FILE} LD_DEBUG=libs microshift run
    ...    sudo=False    return_rc=True    return_stdout=True    return_stderr=True
    Should Be Equal As Integers    1    ${rc}
    ${stdout}    ${rc}=    Execute Command
    ...    grep ${FIPS_PATTERN} ${USHIFT_LIBS_DUMP_FILE}*
    ...    sudo=False    return_rc=True
    Should Be Equal As Integers    0    ${rc}

Fips Should Be Enabled
    [Documentation]    Check if FIPS is enabled on RHEL.
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    bash -x fips-mode-setup --check
    ...    sudo=True    return_rc=True    return_stdout=True    return_stderr=True
    Should Be Equal As Integers    0    ${rc}
    Should Match    ${stdout}    FIPS mode is enabled.
