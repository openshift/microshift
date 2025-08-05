*** Settings ***
Documentation       Tests related to FIPS Validation

Resource            ../../resources/ostree-health.resource
Resource            ../../resources/common.resource
Resource            ../../resources/selinux.resource
Resource            ../../resources/microshift-host.resource
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
    ${is_bootc}=    Is System Bootc
    IF    ${is_bootc}
        Fips Should Be Enabled Bootc
    ELSE
        Fips Should Be Enabled Non-Bootc
    END

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
    VAR    ${podman_args}=    --authfile /etc/crio/openshift-pull-secret --privileged -i -v /:/myroot
    VAR    ${scan_command}=    scan node --root /myroot
    ${path}=    Create Random Temp File
    VAR    ${CHECK_PAYLOAD_OUTPUT_FILE}=    ${path}    scope=GLOBAL    # robocop: off=no-global-variable
    ${rc}=    Execute Command    rpm -qi microshift >${CHECK_PAYLOAD_OUTPUT_FILE} 2>&1
    ...    sudo=True    return_rc=True    return_stdout=False    return_stderr=False
    Should Be Equal As Integers    0    ${rc}
    ${rc}=    Execute Command
    ...    podman run ${podman_args} ${CHECK_PAYLOAD_IMAGE} ${scan_command} >>${CHECK_PAYLOAD_OUTPUT_FILE} 2>&1
    ...    sudo=True    return_rc=True    return_stdout=False    return_stderr=False
    Should Be Equal As Integers    0    ${rc}

Check Container Images In Release Must Pass
    [Documentation]    Run check-paylod Tool for Release containers
    VAR    ${podman_pull_secret}=    /root/.config/containers/auth.json
    VAR    ${podman_mounts}=    -v ${PULL_SECRET_PATH}:${podman_pull_secret}
    VAR    ${podman_args}=    --rm --authfile ${PULL_SECRET_PATH} --privileged ${podman_mounts}
    ${path}=    Create Random Temp File
    VAR    ${CHECK_PAYLOAD_REL_OUTPUT_FILE}=    ${path}    scope=GLOBAL    # robocop: off=no-global-variable
    @{images}=    Get Images From Release File
    FOR    ${image}    IN    @{images}
        VAR    ${scan_command}=    scan operator --spec ${image}
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

Fips Should Be Enabled Non-Bootc
    [Documentation]    Check if FIPS is enabled on a non-bootc RHEL
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    bash -x fips-mode-setup --check
    ...    sudo=True    return_rc=True    return_stdout=True    return_stderr=True
    Should Be Equal As Integers    0    ${rc}
    Should Match    ${stdout}    FIPS mode is enabled.

Fips Should Be Enabled Bootc
    [Documentation]    Check if FIPS is enabled on a bootc RHEL

    # Verify FIPS crypto flag is enabled in the system
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    cat /proc/sys/crypto/fips_enabled
    ...    sudo=False    return_rc=True    return_stdout=True    return_stderr=True
    Should Be Equal As Integers    0    ${rc}
    Should Be Equal As Strings    ${stdout.strip()}    1

    # Verify crypto policies are set to FIPS
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    update-crypto-policies --show
    ...    sudo=False    return_rc=True    return_stdout=True    return_stderr=True
    Should Be Equal As Integers    0    ${rc}
    Should Be Equal As Strings    ${stdout.strip()}    FIPS

    # Verify initramfs FIPS module presence
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    bash -c 'lsinitrd -m 2>/dev/null | grep -Fxq fips'
    ...    sudo=False    return_rc=True    return_stdout=True    return_stderr=True
    Should Be Equal As Integers    0    ${rc}

Get Images From Release File
    [Documentation]    Obtains list of Images from Release.
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    jq -r '.images | .[]' /usr/share/microshift/release/release-$(uname -m).json
    ...    return_stdout=True    return_stderr=True    return_rc=True
    Should Be Equal As Integers    0    ${rc}
    Log Many    ${stdout}    ${stderr}    ${rc}

    @{images}=    Split String    ${stdout}    \n
    RETURN    @{images}
