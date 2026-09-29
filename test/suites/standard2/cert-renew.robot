*** Settings ***
Documentation       Offline certificate renewal, read-only planning, and structured output.

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-host.resource
Resource            ../../resources/microshift-process.resource
Library             Collections
Library             ../../resources/DataFormats.py

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           restart


*** Variables ***
${CERTIFICATE_LOCK}     /var/lib/microshift-backups/certs.lock


*** Test Cases ***
Running Service Holds The Certificate Lock
    [Documentation]    The persistent lock uses existing SELinux policy and excludes writers while MicroShift runs.
    ${metadata}=    Command Should Work    stat -c '%a %C' ${CERTIFICATE_LOCK}
    Should Match Regexp    ${metadata}    ^600 .*:container_var_lib_t:
    ${rc}=    Execute Command
    ...    flock -n -x ${CERTIFICATE_LOCK} true    sudo=True    return_stdout=False    return_rc=True
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

Apply Is Refused While Running
    [Documentation]    Applying either mode must fail with empty stdout and a structured service-state error.
    ${before}=    Certificate Material Digest
    FOR    ${format}    IN    json    yaml
        FOR    ${mode}    IN    serving    ca
            Renewal Error Should Be Reported
            ...    microshift certs renew --${mode} -o ${format}    ${format}    MicroShiftRunning
        END
    END
    ${after}=    Certificate Material Digest
    Should Be Equal    ${before}    ${after}
    MicroShift Service Is Active

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
    END

Leaf Renewal Keeps CAs Unchanged
    [Documentation]    Renew leaves offline, verify CA bytes, then verify API readiness with regenerated kubeconfigs.
    Stop MicroShift
    ${before}=    Certificate Status Document
    ${ca_before}=    CA Material Digest
    ${result}=    Renewal Document    serving    json
    Validate Renewal Document    ${result}    serving    ${False}    ${before}
    ${ca_after}=    CA Material Digest
    Should Be Equal    ${ca_before}    ${ca_after}
    Renewal Matches Current Status    ${result}
    Renewal Has Finished Offline
    [Teardown]    Start And Wait For MicroShift

CA Renewal Updates The Whole Chain
    [Documentation]    Rotate all CAs and descendants, expose redistribution impact, and restart successfully.
    Stop MicroShift
    ${before}=    Certificate Status Document
    ${ca_before}=    CA Material Digest
    ${result}=    Renewal Document    ca    yaml
    Validate Renewal Document    ${result}    ca    ${False}    ${before}
    ${ca_after}=    CA Material Digest
    Should Not Be Equal    ${ca_before}    ${ca_after}
    Renewal Matches Current Status    ${result}
    Renewal Has Finished Offline
    [Teardown]    Start And Wait For MicroShift


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
    Should Be True    $result['status'] == ('validated' if $dry_run else 'completed')
    Validate Renewal Items    ${result}    ${mode}    ${dry_run}    ${before}
    Should Be Equal    ${result}[impact][serviceRestartRequired]    ${True}
    Should Be True    $result['impact']['kubeconfigRedistributionRequired'] == ($mode == 'ca')
    Should Be Equal    ${result}[impact][applicationReloadMayBeRequired]    ${True}
    Should Not Be Empty    ${result}[warnings]

Validate Renewal Items
    [Documentation]    Check exact inventory selection, deterministic ordering, and per-certificate state.
    [Arguments]    ${result}    ${mode}    ${dry_run}    ${before}
    ${identities}=    Evaluate    [(i['service'], i['name'], i['role']) for i in $result['items']]
    ${expected}=    Evaluate
    ...    sorted((i['service'], i['name'], i['role']) for i in $before['items'] if $mode == 'ca' or i['role'] != 'ca')
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

Renewal Has Finished Offline
    [Documentation]    Renewal must not restart the service or leave a transaction behind.
    ${state}=    Get Systemd Setting    microshift.service    ActiveState
    Should Be Equal    ${state}    inactive
    Command Should Work    test ! -e /var/lib/microshift/.cert-renewal
    Certificate Lock Is Stable
    Command Should Work    flock -n -x ${CERTIFICATE_LOCK} true

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
