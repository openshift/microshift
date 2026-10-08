*** Settings ***
Documentation       Compare strict uncached signed CRI-O pulls before and after a temporary COS 10 workaround

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-host.resource
Resource            ../../resources/systemd.resource

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
${BASELINE_PULL_FAILED}         ${FALSE}
${SIGNED_IMAGE}                 registry.access.redhat.com/ubi9/ubi-minimal:9.6
${SIGNED_IMAGE_REPOSITORY}      registry.access.redhat.com/ubi9/ubi-minimal


*** Test Cases ***
Compare Signed Image Pull Before And After Temporary Workaround
    [Documentation]    Compare the same enforced-policy, uncached pull around the temporary workaround for
    ...    https://redhat.atlassian.net/browse/OCPBUGS-129494. Remove the workaround when that issue is resolved.
    ...    A nonzero baseline is recorded without inferring its cause; the post-workaround pull must succeed.
    Enforce Signed Policy For Test Image
    Pull Signed Image Before Workaround

    Apply Temporary Container Configuration Workaround
    Pull Signed Image After Workaround

    IF    ${BASELINE_PULL_FAILED}
        Log    Baseline pull failed without attributing a cause; the post-workaround pull passed
    ELSE
        Log    Baseline pull passed, so the reported failure was not reproduced; the post-workaround pull also passed
    END


*** Keywords ***
Setup
    [Documentation]    Log in, validate the COS 10 VM runtime, and protect the original policy
    Login MicroShift Host
    Validate Workaround Runtime
    Back Up Container Policy
    List Container Configuration Layout
    ${shipped_policy_path}=    Find Shipped Container Policy
    VAR    ${SHIPPED_POLICY_PATH}=    ${shipped_policy_path}    scope=SUITE

Validate Workaround Runtime
    [Documentation]    Require the existing VM packages and Python 3.12 native TOML reader
    Command Should Work    rpm -q containers-common cri-o python3
    Command Should Work    crio --version
    Command Should Work
    ...    python3 -c 'import sys, tomllib; assert sys.version_info[:2] == (3, 12); assert tomllib.__name__ == "tomllib"'

Teardown    # robocop: off=too-many-calls-in-keyword
    [Documentation]    Restore policy and log out; the one-shot workaround VM is destroyed by the scenario
    TRY
        Run Keyword And Ignore Error    Command Should Work    crictl rmi ${SIGNED_IMAGE}
        ${policy_status}    ${policy_error}=    Restore Policy For Teardown
        ${helper_status}    ${helper_error}=    Run Keyword And Ignore Error
        ...    Command Should Work    rm -f ${WORKAROUND_REMOTE_PATH} ${TEST_POLICY_PATH}
        Clean Incomplete Policy Backup

        IF    ${POLICY_BACKUP_READY}
            Should Be Equal    ${policy_status}    PASS
            ...    msg=Failed to restore ${POLICY_JSON_PATH}: ${policy_error}
        END
        Should Be Equal    ${helper_status}    PASS
        ...    msg=Failed to clean up the copied helper: ${helper_error}
    FINALLY
        Logout MicroShift Host
    END

Restore Policy For Teardown
    [Documentation]    Restore policy without blocking the remaining teardown work
    VAR    ${status}=    NOT RUN
    VAR    ${error}=    Policy backup was not completed
    IF    ${POLICY_BACKUP_READY}
        ${status}    ${error}=    Run Keyword And Ignore Error    Restore Container Policy
    END
    RETURN    ${status}    ${error}

Back Up Container Policy    # robocop: off=too-many-calls-in-keyword
    [Documentation]    Back up the original policy or record its original absence before replacement
    ${backup_dir}=    Command Should Work    mktemp --directory /var/tmp/signed-image-pull-policy.XXXXXX
    VAR    ${POLICY_BACKUP_DIR}=    ${backup_dir}    scope=SUITE
    ${exists_stdout}    ${exists_stderr}    ${exists_rc}=    Command Execution
    ...    test -e ${POLICY_JSON_PATH} || test -L ${POLICY_JSON_PATH}
    IF    ${exists_rc} == 0
        Command Should Work    test -f ${POLICY_JSON_PATH} && test ! -L ${POLICY_JSON_PATH}
        Command Should Work
        ...    cp --archive --no-dereference ${POLICY_JSON_PATH} ${POLICY_BACKUP_DIR}/policy.json
        Command Should Work
        ...    test -f ${POLICY_BACKUP_DIR}/policy.json && test ! -L ${POLICY_BACKUP_DIR}/policy.json
        Command Should Work    printf 'present\\n' >${POLICY_BACKUP_DIR}/original-state
        VAR    ${POLICY_ORIGINAL_EXISTED}=    ${TRUE}    scope=SUITE
    ELSE
        Command Should Work    test ! -e ${POLICY_JSON_PATH} && test ! -L ${POLICY_JSON_PATH}
        Command Should Work    printf 'absent\\n' >${POLICY_BACKUP_DIR}/original-state
    END
    Command Should Work    test -f ${POLICY_BACKUP_DIR}/original-state && test ! -L ${POLICY_BACKUP_DIR}/original-state
    VAR    ${POLICY_BACKUP_READY}=    ${TRUE}    scope=SUITE

Restore Container Policy
    [Documentation]    Atomically restore the policy and remove backup data only after success
    IF    ${POLICY_ORIGINAL_EXISTED}
        ${command}=    Catenate
        ...    bash -c 'set -eu;
        ...    test "$(cat ${POLICY_BACKUP_DIR}/original-state)" = present;
        ...    test -f ${POLICY_BACKUP_DIR}/policy.json && test ! -L ${POLICY_BACKUP_DIR}/policy.json;
        ...    test ! -L ${POLICY_JSON_PATH};
        ...    restore=${POLICY_JSON_PATH}.signed-image-pull-restore;
        ...    test ! -e "$restore" && test ! -L "$restore";
        ...    cp --archive --no-dereference ${POLICY_BACKUP_DIR}/policy.json "$restore";
        ...    test -f "$restore" && test ! -L "$restore";
        ...    mv -T "$restore" ${POLICY_JSON_PATH}'
    ELSE
        ${command}=    Catenate
        ...    bash -c 'set -eu;
        ...    test "$(cat ${POLICY_BACKUP_DIR}/original-state)" = absent;
        ...    test ! -L ${POLICY_JSON_PATH};
        ...    rm -f ${POLICY_JSON_PATH}'
    END
    Command Should Work    ${command}
    Command Should Work    rm -rf -- ${POLICY_BACKUP_DIR}

Clean Incomplete Policy Backup
    [Documentation]    Remove preparation debris, but retain a completed backup after restore failure
    IF    '${POLICY_BACKUP_DIR}' != '' and not ${POLICY_BACKUP_READY}
        Command Should Work    rm -rf -- ${POLICY_BACKUP_DIR}
    END

List Container Configuration Layout
    [Documentation]    Log configuration filenames without exposing policy or key contents
    ${command}=    Catenate
    ...    bash -c 'for path in /etc/containers /usr/share/containers
    ...    /etc/containers/registries.d /usr/share/containers/registries.d
    ...    /etc/containers/storage.conf.d /usr/share/containers/storage.conf.d
    ...    /etc/containers/storage.rootful.conf.d /usr/share/containers/storage.rootful.conf.d;
    ...    do echo "### $path";
    ...    if test -d "$path";
    ...    then find "$path" -maxdepth 1 -type f -printf "%f\\n" | sort || true;
    ...    else echo "(missing)"; fi; done'
    Command Should Work    ${command}

Find Shipped Container Policy
    [Documentation]    Select the containers-common policy used to build the strict test policy
    ${command}=    Catenate
    ...    bash -c 'if test -f ${SYSTEM_POLICY_PATH};
    ...    then echo ${SYSTEM_POLICY_PATH};
    ...    elif test -f ${PRESERVED_POLICY_PATH};
    ...    then echo ${PRESERVED_POLICY_PATH};
    ...    else echo "No shipped containers policy found" >&2; exit 1; fi'
    ${policy_path}=    Command Should Work    ${command}
    RETURN    ${policy_path}

Enforce Signed Policy For Test Image
    [Documentation]    Require a valid signature for the test repository and verify the installed policy
    Command Should Work
    ...    test ! -L ${POLICY_JSON_PATH} && test ! -e ${TEST_POLICY_PATH} && test ! -L ${TEST_POLICY_PATH}
    ${command}=    Catenate
    ...    jq --arg repository '${SIGNED_IMAGE_REPOSITORY}'
    ...    '{"default":[{"type":"insecureAcceptAnything"}],
    ...    "transports":{"docker":{($repository):.transports.docker["registry.access.redhat.com"]}}}'
    ...    ${SHIPPED_POLICY_PATH} >${TEST_POLICY_PATH}
    Command Should Work    ${command}
    Command Should Work    install -o root -g root -m 0644 ${TEST_POLICY_PATH} ${POLICY_JSON_PATH}
    Verify Strict Test Policy Is Unchanged

Verify Strict Test Policy Is Unchanged
    [Documentation]    Ensure the helper never overwrote or weakened the enforced test policy
    Command Should Work    cmp --silent ${TEST_POLICY_PATH} ${POLICY_JSON_PATH}
    ${policy_type}=    Command Should Work
    ...    jq -er '.transports.docker["${SIGNED_IMAGE_REPOSITORY}"][0].type' ${POLICY_JSON_PATH}
    Should Match Regexp    ${policy_type}    ^(signedBy|sigstoreSigned)$

Remove Test Image From CRI-O Storage
    [Documentation]    Ensure the next pull fetches and verifies the image instead of using cache
    Command Should Work
    ...    bash -o pipefail -c 'crictl images --quiet --no-trunc ${SIGNED_IMAGE} | xargs --no-run-if-empty crictl rmi'
    ${cached_image}=    Command Should Work    crictl images --quiet ${SIGNED_IMAGE}
    Should Be Empty    ${cached_image}

Pull Signed Image Before Workaround
    [Documentation]    Record the raw baseline result; any nonzero return code is an accepted baseline failure
    Remove Test Image From CRI-O Storage
    Verify Strict Test Policy Is Unchanged
    ${stdout}    ${stderr}    ${rc}=    Command Execution    crictl pull ${SIGNED_IMAGE}
    Record Pull Artifact    before    ${stdout}    ${stderr}    ${rc}
    IF    ${rc} == 0
        Log    Baseline pull returned zero; the reported failure was not reproduced
    ELSE
        VAR    ${BASELINE_PULL_FAILED}=    ${TRUE}    scope=SUITE
        Log    Baseline pull returned ${rc}; accepting the failure without attributing a cause
    END

Apply Temporary Container Configuration Workaround
    [Documentation]    TEMPORARY until OCPBUGS-129494 is resolved; remove after https://redhat.atlassian.net/browse/OCPBUGS-129494
    ${stdout}    ${stderr}    ${rc}=    Command Execution    python3 ${WORKAROUND_REMOTE_PATH}
    Log    ${stdout}
    Log    ${stderr}
    Should Be Equal As Integers    ${rc}    0
    Verify Strict Test Policy Is Unchanged
    Systemctl    restart    crio.service

Pull Signed Image After Workaround
    [Documentation]    Require the identical enforced-policy, uncached pull to return zero
    Remove Test Image From CRI-O Storage
    Verify Strict Test Policy Is Unchanged
    ${stdout}    ${stderr}    ${rc}=    Command Execution    crictl pull ${SIGNED_IMAGE}
    Record Pull Artifact    after    ${stdout}    ${stderr}    ${rc}
    Should Be Equal As Integers    ${rc}    0

Record Pull Artifact
    [Documentation]    Preserve plain return code, stdout, and stderr for each A/B pull
    [Arguments]    ${stage}    ${stdout}    ${stderr}    ${rc}
    ${artifact}=    Catenate    SEPARATOR=\n
    ...    return_code=${rc}
    ...    --- stdout ---
    ...    ${stdout}
    ...    --- stderr ---
    ...    ${stderr}
    OperatingSystem.Create File    ${OUTPUTDIR}/signed-image-pull-${stage}.log    ${artifact}\n
