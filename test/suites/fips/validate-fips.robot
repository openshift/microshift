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
${PULL_SECRET_PATH}         /etc/crio/openshift-pull-secret


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

Verify Container Images FIPS Compliant
    [Documentation]    Performs a FIPS validation against the Released payload
    Check Container Images In Release Must Pass


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig
    Start MicroShift
    Wait Until Greenboot Health Check Exited

Teardown
    [Documentation]    Test suite teardown
    # Download the binary Libs dump files to the artifacts
    Run Keyword And Ignore Error
    ...    SSHLibrary.Get File    ${USHIFT_LIBS_DUMP_FILE}*    ${OUTPUTDIR}/
    Run Keyword And Ignore Error
    ...    SSHLibrary.Get File    ${CHECK_PAYLOAD_OUTPUT_FILE}    ${OUTPUTDIR}/check-payload.log
    Run Keyword And Ignore Error
    ...    SSHLibrary.Get File    ${CHECK_PAYLOAD_REL_OUTPUT_FILE}    ${OUTPUTDIR}/check-release-payload.log
    Start MicroShift
    Wait For MicroShift
    Logout MicroShift Host

Check Payload Tool Must Pass
    [Documentation]    Run check-paylod Tool
    ${podman_args}=    Set Variable    --authfile /etc/crio/openshift-pull-secret --privileged -i -v /:/myroot
    ${scan_command}=    Set Variable    scan node --root /myroot
    ${path}=    Create Random Temp File
    Set Global Variable    ${CHECK_PAYLOAD_OUTPUT_FILE}    ${path}
    ${rc}=    Execute Command    rpm -qi microshift >${CHECK_PAYLOAD_OUTPUT_FILE} 2>&1
    ...    sudo=True    return_rc=True    return_stdout=False    return_stderr=False
    Should Be Equal As Integers    0    ${rc}
    ${rc}=    Execute Command
    ...    podman run ${podman_args} ${CHECK_PAYLOAD_IMAGE} ${scan_command} >>${CHECK_PAYLOAD_OUTPUT_FILE} 2>&1
    ...    sudo=True    return_rc=True    return_stdout=False    return_stderr=False
    Should Be Equal As Integers    0    ${rc}

Check Container Images In Release Must Pass
    [Documentation]    Run check-paylod Tool for Release containers
    ${podman_pull_secret}=    Set Variable    /root/.config/containers/auth.json
    ${podman_mounts}=    Set Variable    -v ${PULL_SECRET_PATH}:${podman_pull_secret}
    ${podman_args}=    Set Variable    --rm --authfile ${PULL_SECRET_PATH} --privileged ${podman_mounts}
    ${path}=    Create Random Temp File
    Set Global Variable    ${CHECK_PAYLOAD_REL_OUTPUT_FILE}    ${path}
    @{images}=    Get Images From Release File
    FOR    ${image}    IN    @{images}
        ${scan_command}=    Set Variable    scan operator --spec ${image}
        ${rc}=    Execute Command
        ...    podman run ${podman_args} ${CHECK_PAYLOAD_IMAGE} ${scan_command} >>${CHECK_PAYLOAD_REL_OUTPUT_FILE} 2>&1
        ...    sudo=True    return_rc=True    return_stdout=False    return_stderr=False
        Should Be Equal As Integers    0    ${rc}
    END

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

Get Images From Release File
    [Documentation]    Obtains list of Images from Release.
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    jq -r '.images | .[]' /usr/share/microshift/release/release-$(uname -m).json
    ...    return_stdout=True    return_stderr=True    return_rc=True
    Should Be Equal As Integers    0    ${rc}
    Log Many    ${stdout}    ${stderr}    ${rc}

    @{images}=    Split String    ${stdout}    \n
    RETURN    @{images}
