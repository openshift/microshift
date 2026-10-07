*** Settings ***
Documentation       Compare an uncached signed CRI-O pull before and after a temporary configuration workaround

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-host.resource
Resource            ../../resources/systemd.resource
Library             assets/temporary_container_config_workaround.py    AS    Workaround

Suite Setup         Setup
Suite Teardown      Teardown


*** Variables ***
${POLICY_JSON_PATH}             /etc/containers/policy.json
${TEST_POLICY_PATH}             /tmp/signed-image-pull-test-policy.json
${SHIPPED_POLICY_PATH}          ${EMPTY}
${POLICY_BACKUP_DIR}            ${EMPTY}
${POLICY_BACKUP_READY}          ${FALSE}
${POLICY_ORIGINAL_EXISTED}      ${FALSE}
${SYSTEM_POLICY_PATH}           /usr/share/containers/policy.json
${PRESERVED_POLICY_PATH}        /etc/containers/policy.json.orig
${WORKAROUND_REMOTE_PATH}       /tmp/temporary_container_config_workaround.py
${WORKAROUND_STATE_DIR}         /var/tmp/microshift-signed-image-pull-workaround
${WORKAROUND_ATTEMPTED}         ${FALSE}
${WORKAROUND_APPLIED}           ${FALSE}
${MISMATCH_REPRODUCED}          ${FALSE}
${SIGNED_IMAGE}                 registry.access.redhat.com/ubi9/ubi-minimal:9.6
${SIGNED_IMAGE_REPOSITORY}      registry.access.redhat.com/ubi9/ubi-minimal
${SIGNATURE_FAILURE_CAUSE}      SignatureValidationFailed: Source image rejected: A signature was required, but no signature exists


*** Test Cases ***
Compare Signed Image Pull Before And After Temporary Workaround
    [Documentation]    Compare identical enforced-trust, uncached CRI-O pulls around the temporary configuration
    ...    workaround for https://redhat.atlassian.net/browse/OCPBUGS-129494. Remove the workaround when the
    ...    issue is resolved. This test only verifies signed-pull behavior; it does not claim storage behavior is fixed.
    Enforce Signed Policy For Test Image
    Pull Signed Image Before Workaround

    Apply Temporary Container Configuration Workaround
    Pull Signed Image After Workaround

    IF    ${MISMATCH_REPRODUCED}
        Log    Configuration mismatch reproduced before the workaround and corrected after it
    ELSE
        Log    Configuration mismatch was not reproduced; both signed-image pulls succeeded
    END


*** Keywords ***
Setup
    [Documentation]    Back up policy and log limited runtime configuration metadata
    Login MicroShift Host
    Back Up Container Policy
    Validate Workaround Runtime
    List Container Configuration Layout
    ${shipped_policy_path}=    Find Shipped Container Policy
    VAR    ${SHIPPED_POLICY_PATH}=    ${shipped_policy_path}    scope=SUITE

Validate Workaround Runtime
    [Documentation]    Verify the disposable COS 10 VM has the installed runtime and Python TOML parser
    Command Should Work    rpm -q containers-common cri-o python3
    Command Should Work    crio --version
    Command Should Work    python3 --version
    Command Should Work    python3 -c 'import tomllib; assert tomllib.__name__ == "tomllib"'
    Command Should Work    test ! -e ${WORKAROUND_STATE_DIR}

Teardown    # robocop: off=too-many-calls-in-keyword
    [Documentation]    Restore policy and every helper-managed file, retaining recovery data if restoration fails
    Run Keyword And Ignore Error    Command Should Work    crictl rmi ${SIGNED_IMAGE}
    ${workaround_restore_status}    ${workaround_restore_error}=    Restore Workaround For Teardown
    ${policy_restore_status}    ${policy_restore_error}=    Restore Policy For Teardown
    ${policy_cleanup_status}    ${policy_cleanup_error}=    Clean Up Policy Backup    ${policy_restore_status}
    ${helper_cleanup_status}    ${helper_cleanup_error}=    Clean Up Copied Helper    ${workaround_restore_status}

    TRY
        IF    ${WORKAROUND_ATTEMPTED}
            Should Be Equal    ${workaround_restore_status}    PASS
            ...    msg=Failed to restore helper-managed configuration: ${workaround_restore_error}
        END
        IF    ${POLICY_BACKUP_READY}
            Should Be Equal    ${policy_restore_status}    PASS
            ...    msg=Failed to restore ${POLICY_JSON_PATH}: ${policy_restore_error}
        END
        Should Be Equal    ${policy_cleanup_status}    PASS
        ...    msg=Failed to clean incomplete policy backup: ${policy_cleanup_error}
        Should Be Equal    ${helper_cleanup_status}    PASS
        ...    msg=Failed to remove copied workaround helper: ${helper_cleanup_error}
    FINALLY
        Logout MicroShift Host
    END

Restore Workaround For Teardown
    [Documentation]    Restore an applied workaround or retry rollback while preserving failed-apply recovery data
    VAR    ${status}=    PASS
    VAR    ${error}=    Workaround was not attempted
    IF    ${WORKAROUND_APPLIED}
        ${status}    ${error}=    Run Keyword And Ignore Error    Restore Temporary Container Configuration
    ELSE IF    ${WORKAROUND_ATTEMPTED}
        ${status}    ${error}=    Run Keyword And Ignore Error    Restore Files After Failed Workaround Apply
    END
    RETURN    ${status}    ${error}

Restore Policy For Teardown
    [Documentation]    Restore policy without preventing the remaining teardown steps after a failure
    VAR    ${status}=    NOT RUN
    VAR    ${error}=    Policy backup was not completed
    IF    ${POLICY_BACKUP_READY}
        ${status}    ${error}=    Run Keyword And Ignore Error    Restore Container Policy
    END
    RETURN    ${status}    ${error}

Clean Up Copied Helper
    [Documentation]    Keep the helper beside retained recovery state after any failed apply
    [Arguments]    ${workaround_restore_status}
    VAR    ${status}=    PASS
    VAR    ${error}=    Helper retained with failed-apply recovery data
    IF    not ${WORKAROUND_ATTEMPTED}
        ${status}    ${error}=    Run Keyword And Ignore Error
        ...    Command Should Work    rm -f ${WORKAROUND_REMOTE_PATH}
    ELSE IF    ${WORKAROUND_APPLIED} and '${workaround_restore_status}' == 'PASS'
        ${status}    ${error}=    Run Keyword And Ignore Error
        ...    Command Should Work    rm -f ${WORKAROUND_REMOTE_PATH}
    END
    RETURN    ${status}    ${error}

Back Up Container Policy
    [Documentation]    Record both the original policy contents and the originally-absent case
    ${backup_dir}=    Command Should Work    mktemp --directory /var/tmp/signed-image-pull-policy.XXXXXX
    VAR    ${POLICY_BACKUP_DIR}=    ${backup_dir}    scope=SUITE
    ${command}=    Workaround.Build Policy Backup Command    ${POLICY_JSON_PATH}    ${POLICY_BACKUP_DIR}
    ${original_state}=    Command Should Work    ${command}
    IF    '${original_state}' == 'present'
        Command Should Work
        ...    test -f ${POLICY_BACKUP_DIR}/policy.json && test ! -L ${POLICY_BACKUP_DIR}/policy.json
        VAR    ${POLICY_ORIGINAL_EXISTED}=    ${TRUE}    scope=SUITE
    ELSE
        Should Be Equal    ${original_state}    absent
    END
    VAR    ${POLICY_BACKUP_READY}=    ${TRUE}    scope=SUITE

Clean Up Policy Backup
    [Documentation]    Remove incomplete backup data, but retain a valid backup after failed restoration
    [Arguments]    ${policy_restore_status}
    VAR    ${status}=    PASS
    VAR    ${error}=    No policy backup directory was created
    IF    '${POLICY_BACKUP_DIR}' != '' and (not ${POLICY_BACKUP_READY} or '${policy_restore_status}' == 'PASS')
        ${status}    ${error}=    Run Keyword And Ignore Error
        ...    Command Should Work    rm -rf -- ${POLICY_BACKUP_DIR}
    END
    RETURN    ${status}    ${error}

Restore Container Policy
    [Documentation]    Atomically restore the original policy or its original absence
    IF    ${POLICY_ORIGINAL_EXISTED}
        ${command}=    Catenate
        ...    bash -c 'set -e;
        ...    test ! -L ${POLICY_JSON_PATH};
        ...    restore_path=${POLICY_JSON_PATH}.signed-image-pull-restore;
        ...    test ! -e "$restore_path";
        ...    cp --archive --no-dereference ${POLICY_BACKUP_DIR}/policy.json "$restore_path";
        ...    mv -fT "$restore_path" ${POLICY_JSON_PATH}'
    ELSE
        ${command}=    Catenate
        ...    bash -c 'test ! -L ${POLICY_JSON_PATH} && rm -f ${POLICY_JSON_PATH}'
    END
    Command Should Work    ${command}
    Command Should Work    rm -f ${TEST_POLICY_PATH}
    Command Should Work    rm -rf -- ${POLICY_BACKUP_DIR}

List Container Configuration Layout
    [Documentation]    Log package-owned configuration filenames without exposing file contents
    ${command}=    Catenate
    ...    bash -c 'for path in /etc/containers/registries.d /usr/share/containers/registries.d
    ...    /usr/share/containers/storage.conf.d /usr/share/containers/storage.rootful.conf.d
    ...    /etc/containers/storage.conf.d /etc/containers/storage.rootful.conf.d;
    ...    do echo "### $path";
    ...    if test -d "$path";
    ...    then find "$path" -maxdepth 1 -type f \( -name "*.yaml" -o -name "*.conf" \)
    ...    -printf "%f\\n" | sort || true;
    ...    else echo "(missing)"; fi; done;
    ...    for path in /etc/containers/storage.conf /usr/share/containers/storage.conf;
    ...    do if test -f "$path"; then echo "$path"; else echo "$path (missing)"; fi; done'
    Command Should Work    ${command}

Find Shipped Container Policy
    [Documentation]    Select the containers-common policy path used by the installed package layout
    ${command}=    Catenate
    ...    bash -c 'if test -f ${SYSTEM_POLICY_PATH};
    ...    then echo ${SYSTEM_POLICY_PATH};
    ...    elif test -f ${PRESERVED_POLICY_PATH};
    ...    then echo ${PRESERVED_POLICY_PATH};
    ...    else echo "No shipped containers policy found" >&2; exit 1; fi'
    ${policy_path}=    Command Should Work    ${command}
    RETURN    ${policy_path}

Enforce Signed Policy For Test Image
    [Documentation]    Keep the default permissive while requiring a valid signature only for the test repository
    ${command}=    Catenate
    ...    jq --arg repository '${SIGNED_IMAGE_REPOSITORY}'
    ...    '{"default":[{"type":"insecureAcceptAnything"}],
    ...    "transports":{"docker":{($repository):.transports.docker["registry.access.redhat.com"]}}}'
    ...    ${SHIPPED_POLICY_PATH} >${TEST_POLICY_PATH}
    Command Should Work    ${command}
    Command Should Work    install -o root -g root -m 0644 ${TEST_POLICY_PATH} ${POLICY_JSON_PATH}
    ${policy_type}=    Command Should Work
    ...    jq -er '.transports.docker["${SIGNED_IMAGE_REPOSITORY}"][0].type' ${POLICY_JSON_PATH}
    Should Match Regexp    ${policy_type}    ^(signedBy|sigstoreSigned)$

Remove Test Image From CRI-O Storage
    [Documentation]    Ensure the next pull fetches and verifies the image instead of using cached CRI-O storage
    Command Should Work
    ...    bash -o pipefail -c 'crictl images --quiet --no-trunc ${SIGNED_IMAGE} | xargs --no-run-if-empty crictl rmi'
    ${cached_image}=    Command Should Work    crictl images --quiet ${SIGNED_IMAGE}
    Should Be Empty    ${cached_image}

Pull Signed Image Before Workaround
    [Documentation]    Accept success, or only the exact known missing-signature rejection; preserve the output artifact
    Remove Test Image From CRI-O Storage
    ${stdout}    ${stderr}    ${rc}=    Command Execution    crictl pull ${SIGNED_IMAGE}
    Record Pull Artifact    before    ${stdout}    ${stderr}
    ${baseline_result}=    Workaround.Classify Baseline Result
    ...    ${rc}    ${stdout}    ${stderr}    ${SIGNATURE_FAILURE_CAUSE}
    IF    '${baseline_result}' == 'pass'
        VAR    ${MISMATCH_REPRODUCED}=    ${FALSE}    scope=SUITE
        Log    Baseline signed-image pull passed; the configuration mismatch was not reproduced
        RETURN
    END
    Should Be Equal    ${baseline_result}    signature_failure
    VAR    ${MISMATCH_REPRODUCED}=    ${TRUE}    scope=SUITE
    Log    Baseline reproduced the exact missing-signature rejection

Apply Temporary Container Configuration Workaround
    [Documentation]    TEMPORARY until OCPBUGS-129494 is resolved; materialize config only between A/B pulls
    VAR    ${WORKAROUND_ATTEMPTED}=    ${TRUE}    scope=SUITE
    ${stdout}    ${stderr}    ${rc}=    Command Execution
    ...    python3 ${WORKAROUND_REMOTE_PATH} apply --state-dir ${WORKAROUND_STATE_DIR}
    Record Pull Artifact    workaround    ${stdout}    ${stderr}
    Should Be Equal As Integers    ${rc}    0
    VAR    ${WORKAROUND_APPLIED}=    ${TRUE}    scope=SUITE
    Systemctl    restart    crio.service

Pull Signed Image After Workaround
    [Documentation]    Require the identical enforced-trust, uncached CRI-O pull to succeed; preserve its output
    Remove Test Image From CRI-O Storage
    ${stdout}    ${stderr}    ${rc}=    Command Execution    crictl pull ${SIGNED_IMAGE}
    Record Pull Artifact    after    ${stdout}    ${stderr}
    Should Be Equal As Integers    ${rc}    0
    Should Not Be Empty    ${stdout}

Restore Temporary Container Configuration
    [Documentation]    Restore helper-managed files and reload the disposable VM runtime configuration
    Command Should Work    python3 ${WORKAROUND_REMOTE_PATH} restore --state-dir ${WORKAROUND_STATE_DIR}
    Systemctl    restart    crio.service

Restore Files After Failed Workaround Apply
    [Documentation]    Retry rollback if complete state exists; reject an incomplete recovery directory
    ${stdout}    ${stderr}    ${state_dir_rc}=    Command Execution    test -e ${WORKAROUND_STATE_DIR}
    IF    ${state_dir_rc} != 0    RETURN
    Command Should Work
    ...    test -f ${WORKAROUND_STATE_DIR}/state.json && test ! -L ${WORKAROUND_STATE_DIR}/state.json
    Command Should Work
    ...    python3 ${WORKAROUND_REMOTE_PATH} restore --keep-state --state-dir ${WORKAROUND_STATE_DIR}

Record Pull Artifact
    [Documentation]    Preserve baseline, workaround, and after output even when the final pull succeeds
    [Arguments]    ${stage}    ${stdout}    ${stderr}
    ${output}=    Catenate    SEPARATOR=\n    ${stdout}    ${stderr}
    OperatingSystem.Create File    ${OUTPUTDIR}/signed-image-pull-${stage}.log    ${output}\n
    RETURN    ${output}
