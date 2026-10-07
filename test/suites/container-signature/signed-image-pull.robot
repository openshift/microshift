*** Settings ***
Documentation       Verify CRI-O can pull a signed image using the shipped registry discovery configuration

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-host.resource

Suite Setup         Setup
Suite Teardown      Teardown


*** Variables ***
${POLICY_JSON_PATH}             /etc/containers/policy.json
${ORIGINAL_POLICY_PATH}         /tmp/signed-image-pull-policy.json
${TEST_POLICY_PATH}             /tmp/signed-image-pull-test-policy.json
${SHIPPED_POLICY_PATH}          ${EMPTY}
${POLICY_BACKUP_READY}          ${FALSE}
${SYSTEM_POLICY_PATH}           /usr/share/containers/policy.json
${PRESERVED_POLICY_PATH}        /etc/containers/policy.json.orig
${SIGNED_IMAGE}                 registry.access.redhat.com/ubi9/ubi-minimal:9.6
${SIGNED_IMAGE_REPOSITORY}      registry.access.redhat.com/ubi9/ubi-minimal


*** Test Cases ***
Pull Signed Image Through CRI-O
    [Documentation]    Verify CRI-O can pull a signed Red Hat image using the registry discovery configuration
    ...    shipped by containers-common. Regression test for https://github.com/microshift-io/microshift/issues/250.
    Enforce Signed Policy For Test Image
    Remove Test Image From CRI-O Storage

    ${image_ref}=    Command Should Work    crictl pull ${SIGNED_IMAGE}
    Should Not Be Empty    ${image_ref}


*** Keywords ***
Setup
    [Documentation]    Save the default policy and log the runtime signature configuration
    Login MicroShift Host
    Command Should Work    cp --preserve ${POLICY_JSON_PATH} ${ORIGINAL_POLICY_PATH}
    VAR    ${POLICY_BACKUP_READY}=    ${TRUE}    scope=SUITE
    Command Should Work    rpm -q containers-common cri-o
    Command Should Work    crio --version
    List Signature Discovery Configuration
    ${shipped_policy_path}=    Find Shipped Container Policy
    VAR    ${SHIPPED_POLICY_PATH}=    ${shipped_policy_path}    scope=SUITE

Teardown
    [Documentation]    Restore the default policy and remove the test image
    VAR    ${restore_status}=    NOT RUN
    VAR    ${restore_error}=    Policy backup was not completed
    IF    ${POLICY_BACKUP_READY}
        ${restore_status}    ${restore_error}=    Run Keyword And Ignore Error
        ...    Command Should Work    cp --preserve ${ORIGINAL_POLICY_PATH} ${POLICY_JSON_PATH}
    END
    Run Keyword And Ignore Error    Command Should Work    crictl rmi ${SIGNED_IMAGE}
    Run Keyword And Ignore Error
    ...    Command Should Work
    ...    rm -f ${TEST_POLICY_PATH}
    IF    ${POLICY_BACKUP_READY} and '${restore_status}' == 'PASS'
        Run Keyword And Ignore Error    Command Should Work    rm -f ${ORIGINAL_POLICY_PATH}
    END
    TRY
        IF    ${POLICY_BACKUP_READY}
            Should Be Equal    ${restore_status}    PASS
            ...    msg=Failed to restore ${POLICY_JSON_PATH}: ${restore_error}
        END
    FINALLY
        Logout MicroShift Host
    END

List Signature Discovery Configuration
    [Documentation]    Log registry discovery paths and filenames without exposing file contents
    ${command}=    Catenate
    ...    bash -c 'for path in /etc/containers/registries.d /usr/share/containers/registries.d;
    ...    do echo "### $path";
    ...    if test -d "$path";
    ...    then find "$path" -maxdepth 1 -type f -name "*.yaml" -printf "%f\\n" | sort || true;
    ...    else echo "(missing)";
    ...    fi;
    ...    done'
    Command Should Work    ${command}

Find Shipped Container Policy
    [Documentation]    Select the containers-common policy path used by the installed package layout
    ${command}=    Catenate
    ...    bash -c 'if test -f ${SYSTEM_POLICY_PATH};
    ...    then echo ${SYSTEM_POLICY_PATH};
    ...    elif test -f ${PRESERVED_POLICY_PATH};
    ...    then echo ${PRESERVED_POLICY_PATH};
    ...    else echo "No shipped containers policy found" >&2; exit 1;
    ...    fi'
    ${policy_path}=    Command Should Work    ${command}
    RETURN    ${policy_path}

Enforce Signed Policy For Test Image
    [Documentation]    Keep the default policy permissive while requiring a valid signature for only the test image
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
    [Documentation]    Ensure the following CRI-O pull must fetch and verify the image
    Command Should Work
    ...    bash -o pipefail -c 'crictl images --quiet --no-trunc ${SIGNED_IMAGE} | xargs --no-run-if-empty crictl rmi'
    ${cached_image}=    Command Should Work    crictl images --quiet ${SIGNED_IMAGE}
    Should Be Empty    ${cached_image}
