*** Settings ***
Documentation       Certificate status JSON and YAML output against a running MicroShift.

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-host.resource
Resource            ../../resources/microshift-process.resource
Library             Collections
Library             ../../resources/DataFormats.py

Suite Setup         Setup
Suite Teardown      Teardown


*** Test Cases ***
JSON And YAML Report The Same Certificates
    [Documentation]    Validate both formats, their field types, and deterministic certificate ordering.
    ${json_status}=    Read Certificate Status    json
    ${yaml_status}=    Read Certificate Status    yaml
    Validate Status Document    ${json_status}
    Validate Status Document    ${yaml_status}
    Should Be Equal    ${json_status}[config]    ${yaml_status}[config]
    Should Be Equal    ${json_status}[warnings]    ${yaml_status}[warnings]
    # Invocation times and remaining seconds may differ; certificate identity and validity must agree.
    ${json_items}=    Stable Certificate Fields    ${json_status}
    ${yaml_items}=    Stable Certificate Fields    ${yaml_status}
    Should Be Equal    ${json_items}    ${yaml_items}

Table Reports Descriptive Certificate States
    [Documentation]    The table uses the same descriptive state names as structured output, not colors.
    ${stdout}=    Command Should Work    microshift certs status
    ${rows}=    Evaluate    $stdout.splitlines()
    Should Be True    $rows[0].split() == ['SERVICE', 'CERTIFICATE', 'STATUS', 'EXPIRY', 'MESSAGE']
    Should Be True    len($rows) > 3
    Should Match Regexp
    ...    ${rows}[-1]
    ...    ^Force restart on expiration imminent: (true|false) \\(configured; applied at service start\\)$
    FOR    ${row}    IN    @{rows}[1:-2]
        Validate Certificate Table Message    ${row}
    END

Unknown Output Format Is Rejected
    [Documentation]    Unsupported formats fail without producing status on stdout.
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    microshift certs status -o xml
    ...    sudo=True    return_stdout=True    return_stderr=True    return_rc=True
    Should Not Be Equal As Integers    ${rc}    0
    Should Be Empty    ${stdout}
    Should Contain    ${stderr}    unsupported output format

Invalid Arguments Produce Structured Errors
    [Documentation]    Positional and flag errors emit only the requested Error document on stderr.
    FOR    ${format}    IN    json    yaml
        Certificate Error Should Be Reported
        ...    microshift certs status unexpected --output=${format}    ${format}    InvalidArguments
        Certificate Error Should Be Reported
        ...    microshift certs status --invalid-status-flag -o ${format}    ${format}    InvalidArguments
        Certificate Error Should Be Reported
        ...    microshift certs status --help --invalid-status-flag -o ${format}    ${format}    InvalidArguments
        Certificate Error Should Be Reported
        ...    microshift certs status -h --invalid-status-flag -o ${format}    ${format}    InvalidArguments
    END

Unprivileged Status Produces Structured Errors
    [Documentation]    Root privilege failures must follow the same machine-readable error contract.
    FOR    ${format}    IN    json    yaml
        Certificate Error Should Be Reported
        ...    runuser -u nobody -- microshift certs status -o ${format}    ${format}    InsufficientPrivileges
    END

Invalid Configuration Does Not Disclose Credentials
    [Documentation]    The real config loader must not expose malformed YAML or proxy credentials in any format.
    VAR    ${invalid_config}=
    ...    proxy: http://test-user:certificate-test-secret@proxy.invalid:8080\ninvalid: [
    Drop In MicroShift Config    ${invalid_config}    99-cert-status-invalid
    FOR    ${format}    IN    ${EMPTY}    json    yaml
        Configuration Error Should Be Sanitized    ${format}
    END
    [Teardown]    Remove Drop In MicroShift Config    99-cert-status-invalid


*** Keywords ***
Setup
    [Documentation]    Wait for MicroShift and its PKI, without requiring unrelated workloads to be healthy.
    Setup Suite
    Wait For MicroShift

Teardown
    [Documentation]    Remove the local kubeconfig and close the host connection.
    Remove Kubeconfig
    Logout MicroShift Host

Read Certificate Status
    [Documentation]    Read one status document and ensure no diagnostics accompany successful output.
    [Arguments]    ${format}
    ${stdout}    ${stderr}    ${rc}=    Command Execution    microshift certs status --output=${format}
    Should Be Equal As Integers    ${rc}    0
    Should Be Empty    ${stderr}
    ${status}=    Parse Certificate Document    ${stdout}    ${format}
    RETURN    ${status}

Validate Certificate Table Message
    [Documentation]    Messages distinguish upcoming validity from expiry and include thresholds when expiring.
    [Arguments]    ${row}
    ${fields}=    Evaluate    $row.split(maxsplit=4)
    Length Should Be    ${fields}    5
    VAR    ${state}=    ${fields}[2]
    VAR    ${message}=    ${fields}[4]
    Should Be True    $state in ('Healthy', 'ExpiresSoon', 'ExpirationImminent', 'Expired', 'NotYetValid')
    IF    $state == 'Healthy'
        Should Match Regexp
        ...    ${message}    ^Valid for [0-9]+ days$
    ELSE IF    $state == 'ExpiresSoon'
        Should Match Regexp
        ...    ${message}    ^Expires in [0-9]+ days, which is at or below the warning threshold of [0-9]+ days$
    ELSE IF    $state == 'ExpirationImminent'
        Should Match Regexp
        ...    ${message}
        ...    ^Expires in [0-9]+ days, which is at or below the critical threshold of [0-9]+ days$
    ELSE IF    $state == 'NotYetValid'
        Should Match Regexp    ${message}    ^Valid in [0-9]+ days$
    ELSE
        Should Match Regexp    ${message}    ^Expired [0-9]+ days ago$
    END

Certificate Error Should Be Reported
    [Documentation]    Failed commands return non-zero with empty stdout and exactly one versioned error.
    [Arguments]    ${command}    ${format}    ${code}
    ${stdout}    ${stderr}    ${rc}=    Command Execution    ${command}
    Should Not Be Equal As Integers    ${rc}    0
    Should Be Empty    ${stdout}
    ${error}=    Parse Certificate Document    ${stderr}    ${format}
    Validate Certificate Error    ${error}    ${code}

Validate Certificate Error
    [Documentation]    Check the version, failure code, and required fields of an Error document.
    [Arguments]    ${error}    ${code}
    Should Be Equal    ${error}[apiVersion]    microshift.openshift.io/v1alpha1
    Should Be Equal    ${error}[kind]    Error
    Should Be Equal    ${error}[code]    ${code}
    Should Not Be Empty    ${error}[message]
    Should Not Be Empty    ${error}[generatedAt]
    Should Be Equal    ${error}[details]    ${None}

Configuration Error Should Be Sanitized
    [Documentation]    Check empty stdout and an exact, configuration-independent error message.
    [Arguments]    ${format}
    ${output_flag}=    Set Variable If    $format == ''    ${EMPTY}    --output=${format}
    ${stdout}    ${stderr}    ${rc}=    Command Execution    microshift certs status ${output_flag}
    Should Not Be Equal As Integers    ${rc}    0
    Should Be Empty    ${stdout}
    Should Not Contain    ${stderr}    certificate-test-secret
    VAR    ${message}=
    ...    failed to load MicroShift configuration; check /etc/microshift/config.yaml and /etc/microshift/config.d
    IF    $format == ''
        Should Be Equal    ${stderr}    Error: ${message}
    ELSE
        ${error}=    Parse Certificate Document    ${stderr}    ${format}
        Validate Certificate Error    ${error}    InvalidConfiguration
        Should Be Equal    ${error}[message]    ${message}
    END

Parse Certificate Document
    [Documentation]    Parse a single document using the requested format.
    [Arguments]    ${document}    ${format}
    IF    $format == 'json'
        ${result}=    Json Parse    ${document}
    ELSE
        ${result}=    Yaml Parse    ${document}
    END
    RETURN    ${result}

Validate Status Document
    [Documentation]    Check the versioned document structure and enum values.
    [Arguments]    ${status}
    Should Be Equal    ${status}[apiVersion]    microshift.openshift.io/v1alpha1
    Should Be Equal    ${status}[kind]    CertificateStatusList
    Should Not Be Empty    ${status}[generatedAt]
    Dictionary Should Contain Key    ${status}    warnings
    Should Be True    isinstance($status['warnings'], list)
    Should Be True    isinstance($status['config']['forceRestartOnExpirationImminent'], bool)
    Dictionary Should Not Contain Key    ${status}[config]    forceRestartOnRedZone
    Should Not Be Empty    ${status}[items]
    FOR    ${item}    IN    @{status}[items]
        Validate Status Item    ${item}    ${status}[generatedAt]
    END
    Certificates Should Be Sorted    ${status}

Validate Status Item
    [Documentation]    Check the public fields for one managed certificate.
    [Arguments]    ${item}    ${generated_at}
    Should Not Be Empty    ${item}[service]
    Should Not Be Empty    ${item}[name]
    Should Be True    $item['role'] in ('ca', 'serving', 'client', 'peer')
    Should Be True    $item['rotationPolicy'] in ('standard', 'extended')
    Dictionary Should Not Contain Key    ${item}    zone
    Validate Certificate Validity    ${item}    ${generated_at}

Validate Certificate Validity
    [Documentation]    Invalid validity windows take precedence over policy-derived expiry thresholds.
    [Arguments]    ${item}    ${generated_at}
    Should Be True    $item['status'] in ('Healthy', 'ExpiresSoon', 'ExpirationImminent', 'Expired', 'NotYetValid')
    Should Be True    isinstance($item['remainingSeconds'], int)
    Should Be True    ($item['status'] == 'Expired') == ($item['remainingSeconds'] <= 0)
    Should Not Be Empty    ${item}[notBefore]
    Should Not Be Empty    ${item}[notAfter]
    ${not_yet_valid}=    Evaluate
    ...    datetime.datetime.fromisoformat($generated_at.replace('Z', '+00:00')) < datetime.datetime.fromisoformat($item['notBefore'].replace('Z', '+00:00'))
    ...    modules=datetime
    Should Be True    ($item['status'] == 'NotYetValid') == ($not_yet_valid and $item['remainingSeconds'] > 0)

Certificates Should Be Sorted
    [Documentation]    Check deterministic ordering by service and certificate name.
    [Arguments]    ${status}
    ${identities}=    Evaluate    [(item['service'], item['name']) for item in $status['items']]
    ${sorted_identities}=    Evaluate    sorted($identities)
    Should Be Equal    ${identities}    ${sorted_identities}

Stable Certificate Fields
    [Documentation]    Select fields that are independent of invocation time.
    [Arguments]    ${status}
    ${items}=    Evaluate
    ...    [{k: v for k, v in item.items() if k not in ('remainingSeconds', 'status')} for item in $status['items']]
    RETURN    ${items}
