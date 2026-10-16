*** Settings ***
Documentation       Tests related to FIPS Validation

Resource            ../../resources/ostree-health.resource
Resource            ../../resources/common.resource
Resource            ../../resources/selinux.resource
Resource            ../../resources/microshift-host.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/router.resource
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

Verify Ingress Router TLS Curves Are FIPS Restricted
    [Documentation]    Verify ROUTER_CURVES env var on the router pod contains
    ...    only NIST-approved curves when running in FIPS mode.
    Router Pod Env Should Have Value    ROUTER_CURVES    P-256:P-384:P-521


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
    VAR    ${CHECK_PAYLOAD_OUTPUT_FILE}=    ${path}    scope=GLOBAL
    ${rc}=    Execute Command    rpm -qi microshift >${CHECK_PAYLOAD_OUTPUT_FILE} 2>&1
    ...    sudo=True    return_rc=True    return_stdout=False    return_stderr=False
    Should Be Equal As Integers    0    ${rc}
    ${rc}=    Execute Command
    ...    podman run ${podman_args} ${CHECK_PAYLOAD_IMAGE} ${scan_command} >>${CHECK_PAYLOAD_OUTPUT_FILE} 2>&1
    ...    sudo=True    return_rc=True    return_stdout=False    return_stderr=False
    Should Be Equal As Integers    0    ${rc}

Check Container Images In Release Must Pass
    [Documentation]    Run check-payload Tool for Release containers
    VAR    ${podman_pull_secret}=    /root/.config/containers/auth.json
    VAR    ${podman_mounts}=    -v ${PULL_SECRET_PATH}:${podman_pull_secret}
    VAR    ${podman_args}=    --rm --authfile ${PULL_SECRET_PATH} --privileged ${podman_mounts}
    ${path}=    Create Random Temp File
    VAR    ${CHECK_PAYLOAD_REL_OUTPUT_FILE}=    ${path}    scope=GLOBAL
    ${pod_image}=    Get Pod Image From Release File
    @{images}=    Get Images From Release File
    FOR    ${image}    IN    @{images}
        Scan Release Image For FIPS    ${image}    ${pod_image}    ${podman_args}
    END

Scan Release Image For FIPS
    [Documentation]    Scan a single release image with check-payload and fail on
    ...    any FIPS finding, except the pause (pod) container's known-benign
    ...    findings (see Count Pod Scan Failures). The pause container ships a
    ...    statically-linked /pause binary on a "FROM scratch" image
    ...    (OCPBUGS-123651); check-payload flags the static binary and the missing
    ...    /etc/redhat-release, but neither is a cryptographic concern and OpenShift
    ...    waives them (ART-23962). The ci/check-payload:latest image is currently
    ...    stale and lacks that waiver, so tolerate ONLY those findings and ONLY for
    ...    the pod image. Remove this once the check-payload image is refreshed.
    [Arguments]    ${image}    ${pod_image}    ${podman_args}
    ${image_log}=    Create Random Temp File
    VAR    ${scan_command}=    scan operator --spec ${image}
    ${rc}=    Execute Command
    ...    podman run ${podman_args} ${CHECK_PAYLOAD_IMAGE} ${scan_command} >${image_log} 2>&1
    ...    sudo=True    return_rc=True    return_stdout=False    return_stderr=False
    # Keep the aggregated report for the artifacts (downloaded in Teardown).
    Execute Command    cat ${image_log} >>${CHECK_PAYLOAD_REL_OUTPUT_FILE}    sudo=True
    IF    ${rc} == 0    RETURN
    IF    "${image}" != "${pod_image}"
        Fail    check-payload scan reported FIPS failures for ${image}
    END
    # Accept the pod image's non-zero scan ONLY when the known waived findings are
    # actually present and nothing else failed. A tool crash or pull error produces
    # no Failure Report (waived == 0), so it is still treated as a failure.
    ${waived}    ${others}=    Count Pod Scan Failures    ${image_log}
    Should Be True
    ...    ${waived} > 0 and ${others} == 0
    ...    msg=check-payload scan for pod image ${image} failed with non-waived findings (waived=${waived}, others=${others})

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

Get Images From Release File
    [Documentation]    Obtains list of Images from Release.
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    jq -r '.images | .[]' /usr/share/microshift/release/release-$(uname -m).json
    ...    return_stdout=True    return_stderr=True    return_rc=True
    Should Be Equal As Integers    0    ${rc}
    Log Many    ${stdout}    ${stderr}    ${rc}

    @{images}=    Split String    ${stdout}    \n
    RETURN    @{images}

Get Pod Image From Release File
    [Documentation]    Return the 'pod' (pause) container image pullspec from the
    ...    MicroShift release metadata.
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    jq -r '.images.pod' /usr/share/microshift/release/release-$(uname -m).json
    ...    return_stdout=True    return_stderr=True    return_rc=True
    Should Be Equal As Integers    0    ${rc}
    RETURN    ${stdout.strip()}

Count Pod Scan Failures
    [Documentation]    Return two counts from the check-payload "Failure Report"
    ...    in ${scan_output}: waived findings and all other (non-waived) findings.
    ...    OpenShift waives two findings for the pause (pod) container (ART-23962):
    ...    the statically linked /pause or /usr/bin/pod binary, and the missing
    ...    /etc/redhat-release on its "FROM scratch" image (OCPBUGS-123651). Each
    ...    binary path is bracketed by the table's column pipes, so a similarly
    ...    named binary (e.g. /pause-helper) is not matched.
    [Arguments]    ${scan_output}
    VAR    ${not_dyn}=
    ...    openshift-enterprise-pod-container *[|] *(/pause|/usr/bin/pod) *[|] *executable is not dynamically linked
    VAR    ${no_distro}=
    ...    openshift-enterprise-pod-container *[|] */etc/redhat-release *[|] *could not find distribution file
    # Failure Report table rows: the section runs from the "---- Failure Report"
    # header to the next "---- " header; its rows start with "|" (border rows
    # start with "+"). Drop the column header row.
    VAR    ${rows}=
    ...    awk '/---- Failure Report/{f=1;next} /^---- /{f=0} f' ${scan_output} | grep '^|' | grep -v 'OPERATOR NAME'
    ${waived}=    Execute Command    ${rows} | grep -cE '${not_dyn}|${no_distro}'
    ...    sudo=True    return_stdout=True
    ${others}=    Execute Command    ${rows} | grep -vE '${not_dyn}|${no_distro}' | grep -c '^|'
    ...    sudo=True    return_stdout=True
    RETURN    ${waived.strip()}    ${others.strip()}
