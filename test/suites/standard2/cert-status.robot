*** Settings ***
Documentation       Certificate status JSON and YAML output against a running MicroShift.

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-host.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/ostree-health.resource
Library             Collections
Library             ../../resources/DataFormats.py

Suite Setup         Setup
Suite Teardown      Logout MicroShift Host


*** Test Cases ***
JSON And YAML Report The Same Certificates
    [Documentation]    Validate both formats, their field types, and deterministic certificate ordering.
    ${json_output}=    Command Should Work    microshift certs status -o json
    ${yaml_output}=    Command Should Work    microshift certs status --output=yaml
    ${json_status}=    Json Parse    ${json_output}
    ${yaml_status}=    Yaml Parse    ${yaml_output}
    Validate Status Document    ${json_status}
    Validate Status Document    ${yaml_status}
    Should Be Equal    ${json_status}[config]    ${yaml_status}[config]
    Should Be Equal    ${json_status}[warnings]    ${yaml_status}[warnings]
    # Invocation times and remaining seconds may differ; certificate identity and validity must agree.
    ${json_items}=    Stable Certificate Fields    ${json_status}
    ${yaml_items}=    Stable Certificate Fields    ${yaml_status}
    Should Be Equal    ${json_items}    ${yaml_items}

Unknown Output Format Is Rejected
    [Documentation]    Unsupported formats fail without producing status on stdout.
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    microshift certs status -o xml
    ...    sudo=True    return_stdout=True    return_stderr=True    return_rc=True
    Should Not Be Equal As Integers    ${rc}    0
    Should Be Empty    ${stdout}
    Should Contain    ${stderr}    unsupported output format


*** Keywords ***
Setup
    [Documentation]    Connect to the host and wait for MicroShift to be ready.
    Check Required Env Variables
    Login MicroShift Host
    Wait For MicroShift Healthcheck Success

Validate Status Document
    [Documentation]    Check the versioned document structure and enum values.
    [Arguments]    ${status}
    Should Be Equal    ${status}[apiVersion]    microshift.openshift.io/v1alpha1
    Should Be Equal    ${status}[kind]    CertificateStatusList
    Should Not Be Empty    ${status}[generatedAt]
    Dictionary Should Contain Key    ${status}    warnings
    Should Be True    isinstance($status['warnings'], list)
    Should Be True    isinstance($status['config']['forceRestartOnRedZone'], bool)
    Should Not Be Empty    ${status}[items]
    FOR    ${item}    IN    @{status}[items]
        Validate Status Item    ${item}
    END
    Certificates Should Be Sorted    ${status}

Validate Status Item
    [Documentation]    Check the public fields for one managed certificate.
    [Arguments]    ${item}
    Should Not Be Empty    ${item}[service]
    Should Not Be Empty    ${item}[name]
    Should Be True    $item['role'] in ('ca', 'serving', 'client', 'peer')
    Should Be True    $item['rotationPolicy'] in ('standard', 'extended')
    Should Be True    $item['zone'] in ('green', 'yellow', 'red')
    Should Be True    isinstance($item['remainingSeconds'], int)
    Should Not Be Empty    ${item}[notBefore]
    Should Not Be Empty    ${item}[notAfter]

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
    ...    [{k: v for k, v in item.items() if k not in ('remainingSeconds', 'zone')} for item in $status['items']]
    RETURN    ${items}
