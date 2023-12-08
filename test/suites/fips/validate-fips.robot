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
${CHECK_PAYLOAD_IMAGE}      registry.ci.openshift.org/ci/check-payload:latest


*** Test Cases ***
Verify Host Is FIPS Enabled
    [Documentation]    Performs a FIPS validation against the host
    Fips Should Be Enabled

Verify Binary Is FIPS Compliant
    [Documentation]    Performs a FIPS validation against the Microshift binary
    Microshift Binary Should Dynamically Link FIPS Ossl Module

Verify Node RPMs FIPS Compliant
    [Documentation]    Performs a FIPS validation against the Installed RPMs
    Check Payload Tool Must Pass


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
    Run Keyword And Ignore Error
    ...    SSHLibrary.Get File    ${CHECK_PAYLOAD_OUTPUT_FILE}    ${OUTPUTDIR}/check-payload.log
    Start MicroShift
    Wait For MicroShift
    Logout MicroShift Host

Check Payload Tool Must Pass
    [Documentation]    Run check-paylod Tool
    ${podman_args}=    Set Variable    --authfile /etc/crio/openshift-pull-secret --privileged -i -v /:/myroot
    ${scan_command}=    Set Variable    scan node --root /myroot
    ${rand}=    Generate Random String
    ${path}=    Join Path    /tmp    ${rand}
    Set Global Variable    ${CHECK_PAYLOAD_OUTPUT_FILE}    ${path}
    ${rc}=    Execute Command
    ...    podman run ${podman_args} ${CHECK_PAYLOAD_IMAGE} ${scan_command} >${CHECK_PAYLOAD_OUTPUT_FILE} 2>&1
    ...    sudo=True    return_rc=True    return_stdout=False    return_stderr=False
    Should Be Equal As Integers    0    ${rc}

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
