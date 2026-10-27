*** Settings ***
Documentation       End-to-end image pull tests for the kubelet image credential provider.
...                 Exercises real pulls from local authenticated registries using the same
...                 exec-contract mock provider as kubelet-credential-provider.robot, covering:
...                 F3 pull without pre-populated CRI-O auth, F4 pull with pre-populated CRI-O auth,
...                 and F5 mixed registries working independently.

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/microshift-host.resource
Resource            ../../resources/ostree.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/network-testing.resource
Library             ../../resources/journalctl.py

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           restart    slow


*** Variables ***
${CURSOR}                   ${EMPTY}
${CP_DROPIN}                10-credential-provider
${CP_CONFIGURED_LOG}        Kubelet image credential provider configured
${CP_BIN_DIR}               /usr/libexec/microshift/credential-providers
${CP_MOCK_PROVIDER}         ${CP_BIN_DIR}/mock-credential-provider
${CP_CONFIG_FILE}           /etc/microshift/credential-providers.yaml

# Local test registries, served over HTTP on the loopback interface of the MicroShift host.
${REG1_HOST}                127.0.0.1:5000
${REG1_NAME}                microshift-cp-reg1
${REG1_USER}                user1
${REG1_PASS}                pass1
${REG2_HOST}                127.0.0.1:5001
${REG2_NAME}                microshift-cp-reg2
${REG2_USER}                user2
${REG2_PASS}                pass2

# Images. Both are on quay.io, which CI mirrors; they are referenced in
# test/assets/credential-provider/mirror-images.yaml so the mirror pulls them in.
${CP_REGISTRY_IMAGE}        quay.io/libpod/registry:2.8.2
${CP_TEST_IMAGE}            quay.io/microshift/busybox:1.36
${TEST_IMAGE_REPO}          busybox:1.36

${CRIO_PULL_SECRET}         /etc/crio/openshift-pull-secret
${CRIO_REG_DROPIN}          /etc/containers/registries.conf.d/99-microshift-cp-test.conf

${CP_VALID}                 SEPARATOR=\n
...                         ---
...                         kubelet:
...                         \ \ imageCredentialProviderConfigPath: ${CP_CONFIG_FILE}
...                         \ \ imageCredentialProviderBinDir: ${CP_BIN_DIR}

# Provider config matching both test registries, served by the single mock provider.
${CP_PROVIDER_CONFIG}       SEPARATOR=\n
...                         apiVersion: kubelet.config.k8s.io/v1
...                         kind: CredentialProviderConfig
...                         providers:
...                         \ \ - name: mock-credential-provider
...                         \ \ \ \ matchImages:
...                         \ \ \ \ \ \ - "${REG1_HOST}"
...                         \ \ \ \ \ \ - "${REG2_HOST}"
...                         \ \ \ \ defaultCacheDuration: "1m"
...                         \ \ \ \ apiVersion: credentialprovider.kubelet.k8s.io/v1

# Mock provider returning valid credentials for both registries. The response cacheDuration is
# 0s (overrides defaultCacheDuration), so kubelet does not cache the credentials and re-execs the
# provider on every pull; this lets the negative test swap the binary without a MicroShift restart.
${CP_MOCK_SCRIPT}           SEPARATOR=\n
...                         \#!/bin/bash
...                         \# Mock kubelet image credential provider: returns static credentials.
...                         cat >/dev/null
...                         cat <<'JSON'
...                         {"kind":"CredentialProviderResponse",
...                         "apiVersion":"credentialprovider.kubelet.k8s.io/v1",
...                         "cacheKeyType":"Registry","cacheDuration":"0s",
...                         "auth":{"${REG1_HOST}":{"username":"${REG1_USER}","password":"${REG1_PASS}"},
...                         "${REG2_HOST}":{"username":"${REG2_USER}","password":"${REG2_PASS}"}}}
...                         JSON

# Mock provider returning wrong credentials for registry 1 (negative path). cacheDuration 0s as above.
${CP_MOCK_SCRIPT_BAD}       SEPARATOR=\n
...                         \#!/bin/bash
...                         \# Mock kubelet image credential provider: returns wrong credentials.
...                         cat >/dev/null
...                         cat <<'JSON'
...                         {"kind":"CredentialProviderResponse",
...                         "apiVersion":"credentialprovider.kubelet.k8s.io/v1",
...                         "cacheKeyType":"Registry","cacheDuration":"0s",
...                         "auth":{"${REG1_HOST}":{"username":"wrong","password":"wrong"}}}
...                         JSON


*** Test Cases ***
Pull Succeeds Via Provider Without Crio Auth
    [Documentation]    F3: with a valid provider binary and config, a private-registry image pulls
    ...    successfully even though CRI-O has no stored auth for the registry. The provider is installed
    ...    in Suite Setup, so no restart is needed here.
    Pattern Should Appear In Log Output    ${CURSOR}    ${CP_CONFIGURED_LOG}
    Crio Pull Secret Should Not Contain    ${REG1_HOST}
    Deploy Pull Test Pod    f3-pull    ${REG1_HOST}/${TEST_IMAGE_REPO}
    Named Pod Should Be Ready    f3-pull
    [Teardown]    Delete Pull Test Pod    f3-pull

Pull Fails When Provider Returns Wrong Credentials
    [Documentation]    F3 negative: the provider is the credential source, so wrong credentials from
    ...    the provider leave the pod in ImagePullBackOff. Swapping the provider binary takes effect
    ...    without a MicroShift restart because caching is disabled (response cacheDuration 0s), so
    ...    kubelet re-execs the provider on the next pull.
    Swap Mock Provider To Bad
    Deploy Pull Test Pod    f3-badcreds    ${REG1_HOST}/${TEST_IMAGE_REPO}
    Pod Image Pull Should Fail    f3-badcreds
    [Teardown]    Run Keywords    Delete Pull Test Pod    f3-badcreds
    ...    AND    Swap Mock Provider To Good

Pull Succeeds With Pre Populated Crio Auth
    [Documentation]    F4: with a valid provider and a pre-populated CRI-O auth entry for the same
    ...    registry, the pull still succeeds; the two credential sources do not interfere. The provider
    ...    is installed in Suite Setup, so no MicroShift restart is needed here.
    Add Crio Pull Secret Entry    ${REG1_HOST}    ${REG1_USER}    ${REG1_PASS}
    Deploy Pull Test Pod    f4-pull    ${REG1_HOST}/${TEST_IMAGE_REPO}
    Named Pod Should Be Ready    f4-pull
    [Teardown]    Run Keywords    Delete Pull Test Pod    f4-pull
    ...    AND    Restore Crio Pull Secret

Mixed Registries Work Independently
    [Documentation]    F5: two registries, each with its own credentials, pull independently through
    ...    the provider without interfering with each other. The provider is installed in Suite Setup,
    ...    so no restart is needed here.
    Deploy Pull Test Pod    f5-reg1    ${REG1_HOST}/${TEST_IMAGE_REPO}
    Deploy Pull Test Pod    f5-reg2    ${REG2_HOST}/${TEST_IMAGE_REPO}
    Named Pod Should Be Ready    f5-reg1
    Named Pod Should Be Ready    f5-reg2
    [Teardown]    Run Keywords    Delete Pull Test Pod    f5-reg1
    ...    AND    Delete Pull Test Pod    f5-reg2


*** Keywords ***
Setup
    [Documentation]    Stand up the credential provider fixtures and two local authenticated
    ...    registries holding the test image, and mark them insecure for CRI-O.
    Setup Suite With Namespace
    Set Namespace Privileged    ${NAMESPACE}
    Make Usr Writable If Needed
    Command Should Work    dnf install -y httpd-tools
    Command Should Work    install -d -o root -g root -m 0755 ${CP_BIN_DIR}
    Upload String To File    ${CP_PROVIDER_CONFIG}    ${CP_CONFIG_FILE}
    Start Test Registries
    # Install the good provider and restart once here. F3/F4/F5 all use it, so they
    # avoid their own restart; only the negative test swaps to bad creds and back.
    Install Good Mock Provider And Restart

Teardown
    [Documentation]    Remove fixtures, stop the registries, restore CRI-O config, restart to a
    ...    clean state. Each step is idempotent so a mid-suite failure still cleans up.
    Remove Drop In MicroShift Config    ${CP_DROPIN}
    Remove Test Registries And Crio Config
    Command Should Work    rm -rf ${CP_BIN_DIR} ${CP_CONFIG_FILE}
    Restart MicroShift
    Teardown Suite With Namespace

Start Test Registries
    [Documentation]    Mark both registries insecure for CRI-O, start them, and push the test image.
    Configure Insecure Registries
    Start Auth Registry    ${REG1_NAME}    5000    ${REG1_USER}    ${REG1_PASS}
    Start Auth Registry    ${REG2_NAME}    5001    ${REG2_USER}    ${REG2_PASS}
    Push Test Image    ${REG1_HOST}    ${REG1_USER}    ${REG1_PASS}
    Push Test Image    ${REG2_HOST}    ${REG2_USER}    ${REG2_PASS}

Remove Test Registries And Crio Config
    [Documentation]    Stop the registries, drop the insecure config and any CRI-O auth, restart CRI-O.
    Command Should Work    podman rm -f ${REG1_NAME} ${REG2_NAME} 2>/dev/null || true
    Command Should Work    rm -rf /tmp/${REG1_NAME}-auth /tmp/${REG2_NAME}-auth
    Command Should Work    rm -f ${CRIO_REG_DROPIN}
    Restore Crio Pull Secret
    Command Should Work    systemctl restart crio

Configure Insecure Registries
    [Documentation]    Mark both local registries insecure so CRI-O pulls over HTTP, then restart CRI-O.
    ${content}=    Catenate    SEPARATOR=\n
    ...    [[registry]]
    ...    location = "${REG1_HOST}"
    ...    insecure = true
    ...    ${EMPTY}
    ...    [[registry]]
    ...    location = "${REG2_HOST}"
    ...    insecure = true
    Upload String To File    ${content}    ${CRIO_REG_DROPIN}
    Command Should Work    systemctl restart crio

Start Auth Registry
    [Documentation]    Run a registry:2 container with htpasswd basic auth on the host loopback.
    [Arguments]    ${name}    ${port}    ${user}    ${pass}
    ${htpasswd}=    Command Should Work    htpasswd -Bbn ${user} ${pass}
    Command Should Work    install -d -m 0755 /tmp/${name}-auth
    Upload String To File    ${htpasswd}    /tmp/${name}-auth/htpasswd
    Command Should Work    podman rm -f ${name} 2>/dev/null || true
    VAR    ${run_cmd}=    SEPARATOR=${SPACE}
    ...    podman run -d --name ${name}
    ...    -p 127.0.0.1:${port}:5000
    ...    -v /tmp/${name}-auth:/auth:Z
    ...    -e REGISTRY_AUTH=htpasswd
    ...    -e REGISTRY_AUTH_HTPASSWD_REALM=Registry
    ...    -e REGISTRY_AUTH_HTPASSWD_PATH=/auth/htpasswd
    ...    ${CP_REGISTRY_IMAGE}
    Command Should Work    ${run_cmd}
    Wait Until Keyword Succeeds    30s    1s
    ...    Command Should Work    curl -sf -u ${user}:${pass} http://127.0.0.1:${port}/v2/

Push Test Image
    [Documentation]    Copy the test image into a local registry using its credentials.
    [Arguments]    ${host}    ${user}    ${pass}
    Command Should Work
    ...    skopeo copy --dest-tls-verify=false --dest-creds ${user}:${pass} docker://${CP_TEST_IMAGE} docker://${host}/${TEST_IMAGE_REPO}

Install Good Mock Provider And Restart
    [Documentation]    Install the valid-credentials mock provider and apply the config, restarting
    ...    MicroShift and recording the journal cursor.
    Install Mock Provider    ${CP_MOCK_SCRIPT}
    Apply Credential Provider Config

Swap Mock Provider To Bad
    [Documentation]    Replace the provider binary in place with the wrong-credentials version. No
    ...    MicroShift restart: the provider config is unchanged and caching is disabled, so kubelet
    ...    execs the new binary on the next pull.
    Install Mock Provider    ${CP_MOCK_SCRIPT_BAD}

Swap Mock Provider To Good
    [Documentation]    Restore the valid-credentials provider binary in place (no restart).
    Install Mock Provider    ${CP_MOCK_SCRIPT}

Install Mock Provider
    [Documentation]    Write the given mock provider script and make it executable.
    [Arguments]    ${script}
    Upload String To File    ${script}    ${CP_MOCK_PROVIDER}
    Command Should Work    chmod 0755 ${CP_MOCK_PROVIDER}

Apply Credential Provider Config
    [Documentation]    Apply the credential provider drop-in and restart MicroShift, recording the
    ...    pre-restart journal cursor. The cursor is suite-scoped because the provider is installed in
    ...    Suite Setup and the F3 test reads it without restarting itself; suite scope keeps it valid
    ...    regardless of test execution order (--randomize all).
    Drop In MicroShift Config    ${CP_VALID}    ${CP_DROPIN}
    ${cursor}=    Get Journal Cursor
    VAR    ${CURSOR}=    ${cursor}    scope=SUITE
    Restart MicroShift

Deploy Pull Test Pod
    [Documentation]    Create a pod that force-pulls the given image, so the credential provider is
    ...    exercised on every run regardless of the node image cache.
    [Arguments]    ${name}    ${image}
    Run With Kubeconfig
    ...    oc run ${name} --image\=${image} --image-pull-policy\=Always --restart\=Never -n ${NAMESPACE} --command -- sleep 3600

Delete Pull Test Pod
    [Documentation]    Delete the named pod and wait for it to be gone.
    [Arguments]    ${name}
    Run With Kubeconfig    oc delete pod ${name} -n ${NAMESPACE} --ignore-not-found --wait\=true

Pod Image Pull Should Fail
    [Documentation]    Wait until the pod's container reports an image pull failure caused by an
    ...    authentication rejection. Asserting the auth-specific message (not just any pull error)
    ...    proves the provider's wrong credentials caused the failure, rather than an unrelated
    ...    unreachable-registry or missing-image error.
    [Arguments]    ${name}    ${timeout}=120s
    Wait Until Keyword Succeeds    ${timeout}    5s    Pod Waiting Reason Is Auth Failure    ${name}

Pod Waiting Reason Is Auth Failure
    [Documentation]    Assert the pod's container is waiting with an image pull error whose message
    ...    indicates the registry rejected the supplied credentials.
    [Arguments]    ${name}
    ${reason}=    Run With Kubeconfig
    ...    oc get pod ${name} -n ${NAMESPACE} -o jsonpath\='{.status.containerStatuses[0].state.waiting.reason}'
    Should Match Regexp    ${reason}    ImagePullBackOff|ErrImagePull
    ${message}=    Run With Kubeconfig
    ...    oc get pod ${name} -n ${NAMESPACE} -o jsonpath\='{.status.containerStatuses[0].state.waiting.message}'
    Should Match Regexp    ${message}    (?i)unauthor|authentication required|invalid username/password|401

Add Crio Pull Secret Entry
    [Documentation]    Back up the CRI-O pull secret, add an auth entry for the registry, restart CRI-O.
    ...    ${CRIO_PULL_SECRET} is a symlink, so the backup dereferences it (cp -L) to capture the real
    ...    contents in an independent regular file. The edit reads that backup and writes the result via
    ...    a temp file, so it never reads and truncates the same file at once (which would corrupt it),
    ...    a failing jq cannot clobber the live secret, and the contents never land in Robot variables
    ...    (TRACE exposure). The final cp writes back through the symlink, preserving its target's 0600.
    [Arguments]    ${host}    ${user}    ${pass}
    Command Should Work    cp -L -p ${CRIO_PULL_SECRET} ${CRIO_PULL_SECRET}.cpbak
    ${auth}=    Command Should Work    printf '%s' '${user}:${pass}' | base64 -w0
    Command Should Work
    ...    jq '.auths["${host}"]={"auth":"${auth}"}' ${CRIO_PULL_SECRET}.cpbak > /tmp/cp-crio-secret.json
    Command Should Work    cp /tmp/cp-crio-secret.json ${CRIO_PULL_SECRET}
    Command Should Work    rm -f /tmp/cp-crio-secret.json
    Command Should Work    systemctl restart crio

Restore Crio Pull Secret
    [Documentation]    Restore the original CRI-O pull secret contents if a backup exists, then restart
    ...    CRI-O. Writes the backup back through the symlink (cp, not mv) so the symlink and its target's
    ...    permissions are preserved, then removes the backup.
    ${rc}=    Execute Command    test -f ${CRIO_PULL_SECRET}.cpbak
    ...    sudo=True    return_rc=True    return_stdout=False
    IF    ${rc} == 0
        Command Should Work    cp ${CRIO_PULL_SECRET}.cpbak ${CRIO_PULL_SECRET}
        Command Should Work    rm -f ${CRIO_PULL_SECRET}.cpbak
        Command Should Work    systemctl restart crio
    END

Crio Pull Secret Should Not Contain
    [Documentation]    Assert the CRI-O pull secret has no auth entry for the given registry host.
    [Arguments]    ${host}
    ${present}=    Command Should Work    jq -r '.auths | has("${host}")' ${CRIO_PULL_SECRET}
    Should Be Equal As Strings    ${present}    false
