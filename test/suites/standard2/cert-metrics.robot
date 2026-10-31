*** Settings ***
Documentation       Local certificate metrics smoke test using the actual MicroShift service.

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-host.resource
Resource            ../../resources/microshift-process.resource
Library             String
Library             ../../resources/DataFormats.py

Suite Setup         Setup
Suite Teardown      Teardown


*** Test Cases ***
Local Metrics Report The Managed Certificate Inventory
    [Documentation]    The local endpoint publishes all CLI inventory identities and bounded status series.
    Wait Until Keyword Succeeds    90s    5s    Certificate Collection Should Succeed
    ${metrics}=    Read Local Metrics
    ${stdout}=    Command Should Work    microshift certs status -o json
    ${status}=    Json Parse    ${stdout}
    ${expiry_lines}=    Get Lines Matching Regexp    ${metrics}    microshift_certificate_not_after_timestamp_seconds.*
    ${expiry_count}=    Get Line Count    ${expiry_lines}
    Length Should Be    ${status}[items]    ${expiry_count}
    FOR    ${item}    IN    @{status}[items]
        Should Contain    ${metrics}    name="${item}[name]"
    END
    Should Contain    ${metrics}    microshift_certificate_renewal_pending 0
    Should Not Contain    ${metrics}    PRIVATE KEY

Metrics Listen Only On Loopback
    [Documentation]    The service binds its configured IPv4 loopback address, not a wildcard.
    ${listeners}=    Command Should Work    ss -H -lnt 'sport = :9090'
    ${count}=    Get Line Count    ${listeners}
    Should Be Equal As Integers    ${count}    1
    Should Contain    ${listeners}    127.0.0.1:9090


*** Keywords ***
Setup
    [Documentation]    Enable the endpoint on the test host without installing a collector.
    Setup Suite
    Wait For MicroShift
    Drop In MicroShift Config
    ...    metrics: {status: Enabled, listenAddress: '127.0.0.1:9090'}
    ...    99-cert-metrics
    Restart MicroShift

Teardown
    [Documentation]    Restore the original configuration and close the host connection.
    Remove Drop In MicroShift Config    99-cert-metrics
    Restart MicroShift
    Remove Kubeconfig
    Logout MicroShift Host

Read Local Metrics
    [Documentation]    Scrape with a bounded timeout and without using a proxy.
    ${output}=    Command Should Work
    ...    curl --noproxy '*' --fail --silent --show-error --max-time 5 http://127.0.0.1:9090/metrics
    RETURN    ${output}

Certificate Collection Should Succeed
    [Documentation]    Wait for the initial background observation to complete.
    ${output}=    Read Local Metrics
    Should Contain    ${output}    microshift_certificate_inventory_collection_success 1
