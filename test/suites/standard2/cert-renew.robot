*** Settings ***
Documentation       Online certificate preparation, startup activation, and structured output.

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-host.resource
Resource            ../../resources/microshift-process.resource
Library             Collections
Library             ../../resources/DataFormats.py

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           restart


*** Variables ***
${CERTIFICATE_LOCK}     /var/lib/microshift-backups/certs.lock
${RUNTIME_LOCK}         /var/lib/microshift-backups/certs-runtime.lock


*** Test Cases ***
Running Service Allows Pending Renewal
    [Documentation]    Runtime and operation locks use existing SELinux policy and separate activation from preparation.
    ${metadata}=    Command Should Work    stat -c '%a %C' ${CERTIFICATE_LOCK}
    Should Match Regexp    ${metadata}    ^600 .*:container_var_lib_t:
    ${rc}=    Execute Command
    ...    flock -n -x ${CERTIFICATE_LOCK} true    sudo=True    return_stdout=False    return_rc=True
    Should Be Equal As Integers    ${rc}    0
    ${metadata}=    Command Should Work    stat -c '%a %C' ${RUNTIME_LOCK}
    Should Match Regexp    ${metadata}    ^600 .*:container_var_lib_t:
    ${rc}=    Execute Command
    ...    flock -n -x ${RUNTIME_LOCK} true    sudo=True    return_stdout=False    return_rc=True
    Should Be Equal As Integers    ${rc}    1
    Certificate Status Document

Dry Run Validates Both Modes Without Changing Material
    [Documentation]    JSON and YAML planning work while running and preserve certificates, keys, and kubeconfigs.
    ${before}=    Certificate Material Digest
    ${status}=    Certificate Status Document
    FOR    ${mode}    IN    serving    ca
        FOR    ${format}    IN    json    yaml
            ${result}=    Renewal Document    ${mode}    ${format}    --dry-run
            Validate Renewal Document    ${result}    ${mode}    ${True}    ${status}
        END
    END
    ${after}=    Certificate Material Digest
    Should Be Equal    ${before}    ${after}
    MicroShift Service Is Active

Startup Waits For An Existing Certificate Operation
    [Documentation]    A busy operation lock delays startup without causing a process restart.
    Stop MicroShift
    Command Should Work
    ...    systemd-run --quiet --collect --unit=microshift-cert-lock-test flock -x ${CERTIFICATE_LOCK} sleep 120
    Wait Until Keyword Succeeds    10x    500ms    Certificate Lock Is Held    ${CERTIFICATE_LOCK}
    Command Should Work    systemctl start --no-block microshift
    Wait Until Keyword Succeeds    10x    500ms    Certificate Lock Is Held    ${RUNTIME_LOCK}
    ${pid}=    Get Systemd Setting    microshift.service    MainPID
    Should Not Be Equal As Integers    ${pid}    0
    Command Should Work    systemctl stop microshift-cert-lock-test
    Wait For MicroShift
    ${current_pid}=    Get Systemd Setting    microshift.service    MainPID
    Should Be Equal    ${pid}    ${current_pid}
    Certificate Lock Is Stable
    Command Should Work    flock -n -x ${CERTIFICATE_LOCK} true
    [Teardown]    Run Keywords
    ...    Execute Command    systemctl stop microshift-cert-lock-test    sudo=True
    ...    AND    Restart MicroShift

Renewal Can Replace Pending Material While Running
    [Documentation]    Both modes prepare successfully without touching active files or restarting the service.
    ${before}=    Certificate Material Digest
    ${pid}=    MicroShift Process ID
    ${status}=    Certificate Status Document
    FOR    ${mode}    IN    serving    ca
        ${result}=    Renewal Document    ${mode}    json
        Validate Renewal Document    ${result}    ${mode}    ${False}    ${status}
        Pending Renewal Matches Result    ${result}
    END
    ${after}=    Certificate Material Digest
    Should Be Equal    ${before}    ${after}
    ${current_pid}=    MicroShift Process ID
    Should Be Equal    ${pid}    ${current_pid}
    MicroShift Service Is Active
    Wait For MicroShift
    [Teardown]    Restart MicroShift

Exactly One Renewal Mode Is Required
    [Documentation]    Missing and conflicting mode selections are rejected before staging.
    FOR    ${format}    IN    json    yaml
        Renewal Error Should Be Reported
        ...    microshift certs renew --dry-run -o ${format}    ${format}    InvalidArguments
        Renewal Error Should Be Reported
        ...    microshift certs renew --serving --ca --dry-run -o ${format}    ${format}    InvalidArguments
    END

Unprivileged Dry Run Is Refused
    [Documentation]    Read-only planning still requires root privileges.
    FOR    ${format}    IN    json    yaml
        Renewal Error Should Be Reported
        ...    runuser -u nobody -- microshift certs renew --serving --dry-run -o ${format}
        ...    ${format}    InsufficientPrivileges
        Renewal Error Should Be Reported
        ...    runuser -u nobody -- microshift certs renew --ca -o ${format}
        ...    ${format}    InsufficientPrivileges
    END

Leaf Renewal Keeps CAs Unchanged
    [Documentation]    Prepare leaves online, preserve active TLS and CA bytes, and activate only on restart.
    ${before}=    Certificate Status Document
    ${files_before}=    Certificate Material Digest
    ${served_before}=    Served Certificate Fingerprints
    ${ca_before}=    CA Material Digest
    ${result}=    Renewal Document    serving    json
    Validate Renewal Document    ${result}    serving    ${False}    ${before}
    ${ca_after}=    CA Material Digest
    Should Be Equal    ${ca_before}    ${ca_after}
    Renewal Has Not Changed Active Material    ${files_before}    ${served_before}
    Pending Renewal Matches Result    ${result}
    Restart MicroShift
    Renewal Matches Current Status    ${result}
    ${ca_after}=    CA Material Digest
    Should Be Equal    ${ca_before}    ${ca_after}
    ${served_after}=    Served Certificate Fingerprints
    Should Not Be Equal    ${served_before}    ${served_after}
    Renewal Activation Has Finished

CA Renewal Updates The Whole Chain
    [Documentation]    Prepare all CAs online with old credentials still usable; activate the new chain on restart.
    ${before}=    Certificate Status Document
    ${files_before}=    Certificate Material Digest
    ${served_before}=    Served Certificate Fingerprints
    ${ca_before}=    CA Material Digest
    ${result}=    Renewal Document    ca    yaml
    Validate Renewal Document    ${result}    ca    ${False}    ${before}
    Renewal Has Not Changed Active Material    ${files_before}    ${served_before}
    Pending Renewal Matches Result    ${result}
    Restart MicroShift
    ${ca_after}=    CA Material Digest
    Should Not Be Equal    ${ca_before}    ${ca_after}
    Renewal Matches Current Status    ${result}
    Renewal Activation Has Finished

Stopped Renewal Waits For Startup Too
    [Documentation]    Preparation while stopped has the same pending contract and does not start the service.
    Stop MicroShift
    ${before}=    Certificate Status Document
    ${files_before}=    Certificate Material Digest
    ${result}=    Renewal Document    serving    yaml
    Validate Renewal Document    ${result}    serving    ${False}    ${before}
    ${files_after}=    Certificate Material Digest
    Should Be Equal    ${files_before}    ${files_after}
    Pending Renewal Matches Result    ${result}
    ${state}=    Get Systemd Setting    microshift.service    ActiveState
    Should Be Equal    ${state}    inactive
    Start And Wait For MicroShift
    Renewal Matches Current Status    ${result}
    Renewal Activation Has Finished
    [Teardown]    Start And Wait For MicroShift

Unrelated Configuration Change Still Activates Renewal
    [Documentation]    Logging configuration does not invalidate prepared certificates.
    ${result}=    Renewal Document    serving    json
    Drop In MicroShift Config    debugging:\n\ \ logLevel: Debug\n    99-cert-renew-test
    Restart MicroShift
    Renewal Matches Current Status    ${result}
    Renewal Activation Has Finished
    [Teardown]    Run Keywords
    ...    Remove Drop In MicroShift Config    99-cert-renew-test
    ...    AND    Restart MicroShift

Certificate Configuration Change Discards Pending Renewal
    [Documentation]    Stale renewal must not block startup or skip normal generation for new SANs.
    ${ca_before}=    CA Material Digest
    Renewal Document    ca    json
    # Preserve the host SAN so Restart MicroShift can retrieve its kubeconfig.
    Drop In MicroShift Config
    ...    apiServer:\n\ \ subjectAltNames:\n\ \ \ \ - ${USHIFT_HOST}\n\ \ \ \ - renewal-config.example.test\n
    ...    99-cert-renew-test
    Restart MicroShift
    Renewal Activation Has Finished
    ${ca_after}=    CA Material Digest
    Should Be Equal    ${ca_before}    ${ca_after}
    ${certificate}=    Command Should Work
    ...    bash -c 'timeout 10 openssl s_client -connect 127.0.0.1:6443 -servername renewal-config.example.test </dev/null 2>/dev/null | openssl x509 -noout -checkhost renewal-config.example.test'
    Should Contain    ${certificate}    does match certificate
    [Teardown]    Run Keywords
    ...    Remove Drop In MicroShift Config    99-cert-renew-test
    ...    AND    Restart MicroShift


*** Keywords ***
Setup
    [Documentation]    Connect and wait only for MicroShift and its PKI.
    Setup Suite
    Wait For MicroShift
    ${lock_id}=    Command Should Work    stat -c '%d:%i' ${CERTIFICATE_LOCK}
    VAR    ${CERTIFICATE_LOCK_ID}=    ${lock_id}    scope=SUITE

Teardown
    [Documentation]    Remove the local kubeconfig and close the connection.
    Remove Kubeconfig
    Logout MicroShift Host

Start And Wait For MicroShift
    [Documentation]    Recover an interrupted renewal if needed, refresh kubeconfig, and check readiness.
    Command Should Work    microshift certs status -o json
    Start MicroShift
    Wait For MicroShift
    Certificate Lock Is Stable

Certificate Lock Is Stable
    [Documentation]    Renewal and service restarts must reuse the same lock inode outside the certificate tree.
    ${lock_id}=    Command Should Work    stat -c '%d:%i' ${CERTIFICATE_LOCK}
    Should Be Equal    ${lock_id}    ${CERTIFICATE_LOCK_ID}

Certificate Lock Is Held
    [Documentation]    Check contention without waiting for or changing the existing lock holder.
    [Arguments]    ${path}
    ${rc}=    Execute Command
    ...    flock -n -x ${path} true    sudo=True    return_stdout=False    return_rc=True
    Should Be Equal As Integers    ${rc}    1

Certificate Status Document
    [Documentation]    Read the current inventory for comparison with renewal selection and results.
    ${stdout}=    Command Should Work    microshift certs status -o json
    ${document}=    Json Parse    ${stdout}
    RETURN    ${document}

Certificate Material Digest
    [Documentation]    Hash the active PKI and resources without logging any certificate or key material.
    ${digest}=    Command Should Work
    ...    bash -o pipefail -c 'find /var/lib/microshift/certs /var/lib/microshift/resources -type f -print0 | sort -z | xargs -0 sha256sum | sha256sum'
    RETURN    ${digest}

CA Material Digest
    [Documentation]    Hash CA certificates and private keys without exposing their contents.
    ${digest}=    Command Should Work
    ...    bash -o pipefail -c 'find /var/lib/microshift/certs -type f "(" -name ca.crt -o -name ca.key ")" -print0 | sort -z | xargs -0 sha256sum | sha256sum'
    RETURN    ${digest}

Renewal Document
    [Documentation]    Require a single machine-readable success document with no stderr diagnostics.
    [Arguments]    ${mode}    ${format}    ${extra}=${EMPTY}
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    microshift certs renew --${mode} -o ${format} ${extra}
    ...    sudo=True    return_stdout=True    return_stderr=True    return_rc=True
    Should Be Equal As Integers    ${rc}    0    msg=${stderr}
    Should Be Empty    ${stderr}
    ${document}=    Yaml Parse    ${stdout}
    RETURN    ${document}

Validate Renewal Document
    [Documentation]    Check the versioned contract and inventory-derived selection, without assuming certificate counts.
    [Arguments]    ${result}    ${mode}    ${dry_run}    ${before}
    Should Be Equal    ${result}[apiVersion]    microshift.openshift.io/v1alpha1
    Should Be Equal    ${result}[kind]    CertificateRenewalResult
    Should Be Equal    ${result}[mode]    ${mode}
    Should Be Equal    ${result}[dryRun]    ${dry_run}
    Should Be True    $result['status'] == ('validated' if $dry_run else 'pending')
    Validate Renewal Items    ${result}    ${mode}    ${dry_run}    ${before}
    Should Be Equal    ${result}[impact][serviceRestartRequired]    ${True}
    Should Be True    $result['impact']['kubeconfigRedistributionRequired'] == ($mode == 'ca')
    Should Be Equal    ${result}[impact][applicationReloadMayBeRequired]    ${True}
    Should Not Be Empty    ${result}[warnings]

Validate Renewal Items
    [Documentation]    Check exact inventory selection, deterministic ordering, and per-certificate state.
    [Arguments]    ${result}    ${mode}    ${dry_run}    ${before}
    ${identities}=    Evaluate    [(i['service'], i['name'], i['role']) for i in $result['items']]
    VAR    &{namespace}=    mode=${mode}
    ${expected}=    Evaluate
    ...    sorted((i['service'], i['name'], i['role']) for i in $before['items'] if mode == 'ca' or i['role'] != 'ca')
    ...    namespace=${namespace}
    Should Be Equal    ${identities}    ${expected}
    Should Not Be Empty    ${result}[items]
    FOR    ${item}    IN    @{result}[items]
        Should Be True    $item['changed'] is not $dry_run
        Should Match Regexp    ${item}[newNotAfter]    ^[0-9]{4}-[0-9]{2}-[0-9]{2}T
        Dictionary Should Contain Key    ${item}    parentCA
    END

Renewal Matches Current Status
    [Documentation]    Committed dates must match the certificates subsequently read by status.
    [Arguments]    ${result}
    ${status}=    Certificate Status Document
    ${expiries}=    Evaluate    {(i['service'], i['name']): i['notAfter'] for i in $status['items']}
    FOR    ${item}    IN    @{result}[items]
        Should Be True    $expiries[($item['service'], $item['name'])] == $item['newNotAfter']
    END

Renewal Activation Has Finished
    [Documentation]    Activation removes the pending generation and releases the operation lock.
    MicroShift Service Is Active
    ${status}=    Certificate Status Document
    Dictionary Should Not Contain Key    ${status}    pendingRenewal
    Command Should Work    test ! -e /var/lib/microshift/.cert-renewal
    Certificate Lock Is Stable
    Command Should Work    flock -n -x ${CERTIFICATE_LOCK} true

Pending Renewal Matches Result
    [Documentation]    Status distinguishes active expiry dates from the prepared replacement dates.
    [Arguments]    ${result}
    ${status}=    Certificate Status Document
    Should Be Equal    ${status}[pendingRenewal][items]    ${result}[items]
    ${active}=    Evaluate    {(i['service'], i['name']): i['notAfter'] for i in $status['items']}
    FOR    ${item}    IN    @{result}[items]
        Should Be True    $active[($item['service'], $item['name'])] == $item['currentNotAfter']
    END

Served Certificate Fingerprints
    [Documentation]    Observe fresh API server, etcd and kubelet TLS handshakes without exposing private keys.
    # s_client may fail without client authentication; x509 must still parse a valid server certificate.
    ${command}=    Catenate    SEPARATOR=${SPACE}
    ...    bash -c 'for port in 6443 2379 10250; do
    ...    timeout 10 openssl s_client -connect 127.0.0.1:$port -servername localhost </dev/null 2>/dev/null
    ...    | openssl x509 -noout -fingerprint -sha256 || exit 1; done'
    ${fingerprints}=    Command Should Work    ${command}
    RETURN    ${fingerprints}

Renewal Has Not Changed Active Material
    [Documentation]    Active files, served TLS identities and API health remain unchanged before restart.
    [Arguments]    ${files_before}    ${served_before}
    ${files_after}=    Certificate Material Digest
    Should Be Equal    ${files_before}    ${files_after}
    FOR    ${attempt}    IN RANGE    3
        ${served_after}=    Served Certificate Fingerprints
        Should Be Equal    ${served_before}    ${served_after}
        Wait For MicroShift
    END

Renewal Error Should Be Reported
    [Documentation]    Failed invocations emit only an Error document on stderr.
    [Arguments]    ${command}    ${format}    ${code}
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    ${command}    sudo=True    return_stdout=True    return_stderr=True    return_rc=True
    Should Not Be Equal As Integers    ${rc}    0
    Should Be Empty    ${stdout}
    IF    $format == 'json'
        ${error}=    Json Parse    ${stderr}
    ELSE
        ${error}=    Yaml Parse    ${stderr}
    END
    Should Be Equal    ${error}[apiVersion]    microshift.openshift.io/v1alpha1
    Should Be Equal    ${error}[kind]    Error
    Should Be Equal    ${error}[code]    ${code}
    Should Not Be Empty    ${error}[message]
