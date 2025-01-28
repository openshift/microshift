*** Settings ***
Documentation       Container policy verification

Resource            ../../resources/microshift-process.resource
Library             OperatingSystem
Library             Collections

Suite Setup         Setup
Suite Teardown      Teardown


*** Variables ***
${POLICY_JSON_PATH}             /etc/containers/policy.json
${IMAGE_SIGSTORE_ENABLED}       False


*** Test Cases ***
Verify Policy JSON Contents
    [Documentation]    Verify container policy contents
    ${policy_contents}=    Command Should Work    cat ${POLICY_JSON_PATH}
    ${policy}=    Json Parse    ${policy_contents}

    IF    ${IMAGE_SIGSTORE_ENABLED}
        Verify Sigstore Signing Enabled    ${policy}
    ELSE
        Verify Sigstore Signing Disabled    ${policy}
    END


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Login MicroShift Host

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host

Verify Sigstore Signing Enabled    # robocop: disable=too-many-calls-in-keyword
    [Documentation]    Verify the policy file contents when sigstore signing
    ...    verification is enabled
    [Arguments]    ${policy}

    # This verification should match the policy contents defined in
    # https://github.com/openshift/microshift/blob/main/test/kickstart-templates/includes/post-containers-sigstore.cfg

    # Verify default entry
    ${default_type}=    Evaluate    "${policy}[default][0][type]"
    Should Be Equal As Strings    ${default_type}    reject

    # Verify quay.io entry
    ${quay_type}=    Evaluate    "${policy}[transports][docker][quay.io/openshift-release-dev][0][type]"
    Should Be Equal    ${quay_type}    sigstoreSigned
    ${quay_key}=    Evaluate    "${policy}[transports][docker][quay.io/openshift-release-dev][0][keyPath]"
    Should Be Equal    ${quay_key}    /etc/containers/RedHat_ReleaseKey3.pub
    ${quay_ident}=    Evaluate
    ...    "${policy}[transports][docker][quay.io/openshift-release-dev][0][signedIdentity][type]"
    Should Be Equal    ${quay_ident}    matchRepoDigestOrExact

    # Verify registry.redhat.io entry
    ${redhat_type}=    Evaluate    "${policy}[transports][docker][registry.redhat.io][0][type]"
    Should Be Equal    ${redhat_type}    sigstoreSigned
    ${redhat_key}=    Evaluate    "${policy}[transports][docker][registry.redhat.io][0][keyPath]"
    Should Be Equal    ${redhat_key}    /etc/containers/RedHat_ReleaseKey3.pub
    ${redhat_ident}=    Evaluate    "${policy}[transports][docker][registry.redhat.io][0][signedIdentity][type]"
    Should Be Equal    ${redhat_ident}    matchRepoDigestOrExact

Verify Sigstore Signing Disabled
    [Documentation]    Verify the policy file contents when sigstore signing
    ...    verification is disabled
    [Arguments]    ${policy}
    # This verification should match the policy contents defined in
    # https://github.com/openshift/microshift/blob/main/test/kickstart-templates/includes/post-containers.cfg

    # Verify default entry
    ${default_type}=    Evaluate    "${policy}[default][0][type]"
    Should Be Equal As Strings    ${default_type}    insecureAcceptAnything

    # Verify transports entry
    ${quay_type}=    Evaluate    '${policy}[transports][docker-daemon][][0][type]'
    Should Be Equal    ${quay_type}    insecureAcceptAnything
