*** Settings ***
Documentation       Verify a strict uncached signed image pull through CRI-O

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-host.resource

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
${SIGNED_IMAGE}                 registry.access.redhat.com/ubi9/ubi-minimal:9.6
${SIGNED_IMAGE_REPOSITORY}      registry.access.redhat.com/ubi9/ubi-minimal


*** Test Cases ***
Pull Signed Image With Strict Policy
    [Documentation]    Require one uncached pull under the shipped signature policy to succeed
    Enforce Signed Policy For Test Image
    Pull Signed Image


*** Keywords ***
Setup
    [Documentation]    Log in, validate the VM runtime, and protect the original policy
    Login MicroShift Host
    Validate Signed Pull Runtime
    Back Up Container Policy
    ${shipped_policy_path}=    Find Shipped Container Policy
    VAR    ${SHIPPED_POLICY_PATH}=    ${shipped_policy_path}    scope=SUITE

Validate Signed Pull Runtime
    [Documentation]    Require the existing VM container runtime packages
    Command Should Work    rpm -q containers-common cri-o
    Command Should Work    crio --version

Teardown    # robocop: off=too-many-calls-in-keyword
    [Documentation]    Restore policy, clean temporary data, report errors, and log out
    TRY
        Run Keyword And Ignore Error    Command Should Work    crictl rmi ${SIGNED_IMAGE}
        ${policy_status}    ${policy_error}=    Restore Policy For Teardown
        ${cleanup_status}    ${cleanup_error}=    Run Keyword And Ignore Error
        ...    Command Should Work    rm -f ${TEST_POLICY_PATH}
        Clean Incomplete Policy Backup

        IF    ${POLICY_BACKUP_READY}
            Should Be Equal    ${policy_status}    PASS
            ...    msg=Failed to restore ${POLICY_JSON_PATH}: ${policy_error}
        END
        Should Be Equal    ${cleanup_status}    PASS
        ...    msg=Failed to clean temporary policy data: ${cleanup_error}
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
        Command Should Work    bash -c "printf 'present\\n' >${POLICY_BACKUP_DIR}/original-state"
        VAR    ${POLICY_ORIGINAL_EXISTED}=    ${TRUE}    scope=SUITE
    ELSE
        Command Should Work    test ! -e ${POLICY_JSON_PATH} && test ! -L ${POLICY_JSON_PATH}
        Command Should Work    bash -c "printf 'absent\\n' >${POLICY_BACKUP_DIR}/original-state"
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
    [Documentation]    Ensure the enforced test policy was not overwritten or weakened
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

Pull Signed Image
    [Documentation]    Require the enforced-policy, uncached CRI-O pull to return zero
    Remove Test Image From CRI-O Storage
    Verify Strict Test Policy Is Unchanged
    ${stdout}    ${stderr}    ${rc}=    Command Execution    crictl pull ${SIGNED_IMAGE}
    Log    ${stdout}
    Log    ${stderr}
    Should Be Equal As Integers    ${rc}    0
