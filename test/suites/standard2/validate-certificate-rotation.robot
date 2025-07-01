*** Settings ***
Documentation       Tests related to MicroShift automated certificate rotation.

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/microshift-host.resource
Resource            ../../resources/ostree-health.resource
Library             DateTime
Library             Collections

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           restart


*** Variables ***
${KUBE_SCHEDULER_CLIENT_CERT}       /var/lib/microshift/certs/kube-control-plane-signer/kube-scheduler/client.crt
${OSSL_CMD}                         openssl x509 -noout -dates -in
${OSSL_DATE_FORMAT}                 %b %d %Y
${TIMEDATECTL_DATE_FORMAT}          %Y-%m-%d %H:%M:%S
${FUTURE_DAYS}                      150


*** Test Cases ***
Certificate Rotation
    [Documentation]    Performs Certificate Expiration Rotation test
    ${first_cert_date}=    Compute Date After Days    365    ${OSSL_DATE_FORMAT}
    Certs Should Expire On    ${KUBE_SCHEDULER_CLIENT_CERT}    ${first_cert_date}
    ${cert_should_expire_in_days}=    Evaluate    365+${FUTURE_DAYS}
    ${cert_expiry_date}=    Compute Date After Days    ${cert_should_expire_in_days}    ${OSSL_DATE_FORMAT}
    ${future_date}=    Compute Date After Days    ${FUTURE_DAYS}    ${TIMEDATECTL_DATE_FORMAT}
    Change System Date To    ${future_date}
    Certs Should Expire On    ${KUBE_SCHEDULER_CLIENT_CERT}    ${cert_expiry_date}


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig    # for readiness checks
    Restart Greenboot And Wait For Success

Teardown
    [Documentation]    Test suite teardown
    Restore System Date
    Logout MicroShift Host

Restore System Date
    [Documentation]    Reset Microshift date to current date
    ${ushift_pid}=    MicroShift Process ID
    Systemctl    start    chronyd
    Wait Until MicroShift Process ID Changes    ${ushift_pid}
    Wait For MicroShift

Change System Date To
    [Documentation]    Move the system to a future date.
    [Arguments]    ${future_date}
    ${ushift_pid}=    MicroShift Process ID
    Systemctl    stop    chronyd
    Command Should Work    TZ=UTC timedatectl set-time "${future_date}"

    Sleep    5s
    Wait Until MicroShift Process ID Changes    ${ushift_pid}

    Wait For Multiple Certificates Regeneration
    Wait Until Keyword Succeeds    30x    5s
    ...    Kubeconfig Should Be Updated    ${USHIFT_HOST}
    Setup Kubeconfig

    Wait For MicroShift

Compute Date After Days
    [Documentation]    return system date after number of days elapsed
    [Arguments]    ${number_of_days}    ${date_format}
    # date command is used here because we need to consider the remote vm timezone .
    ${future_date}=    Command Should Work    TZ=UTC date "+${date_format}" -d "$(date) + ${number_of_days} day"
    RETURN    ${future_date}

Certs Should Expire On
    [Documentation]    verify if the ceritifate expires at given date.
    [Arguments]    ${cert_file}    ${cert_expected_date}
    ${expiration_date}=    Command Should Work
    ...    ${OSSL_CMD} ${cert_file} | grep notAfter | cut -f2 -d'=' | awk '{printf ("%s %02d %d",$1,$2,$4)}'
    Should Be Equal As Strings    ${cert_expected_date}    ${expiration_date}

Wait For Multiple Certificates Regeneration
    [Documentation]    Wait for multiple certificate files to be regenerated and valid

    @{cert_files}=    Create List
    ...    /var/lib/microshift/certs/kube-control-plane-signer/kube-scheduler/client.crt
    ...    /var/lib/microshift/certs/kube-control-plane-signer/kube-controller-manager/client.crt
    ...    /var/lib/microshift/certs/admin-kubeconfig-signer/admin-kubeconfig-client/client.crt
    FOR    ${cert_file}    IN    @{cert_files}
        Wait Until Keyword Succeeds    30x    5s
        ...    Certificate Should Be Valid For Current Time    ${cert_file}
    END

Certificate Should Be Valid For Current Time
    [Documentation]    Verify that certificate is valid for the current system time
    [Arguments]    ${cert_file}

    ${cert_not_before}=    Command Should Work
    ...    ${OSSL_CMD} ${cert_file} | grep notBefore | cut -f2 -d'=' | xargs -I {} date -d "{}" +%s
    ${current_time}=    Command Should Work    date +%s
    Should Be True    ${cert_not_before} <= ${current_time}
    ...    msg=Certificate NotBefore (${cert_not_before}) is after current time (${current_time})

Kubeconfig Should Be Updated
    [Documentation]    Verify that kubeconfig contains valid certificates for current time
    [Arguments]    ${host}

    ${kubeconfig}=    Get Kubeconfig    ${host}
    ${cert_data}=    Command Should Work
    ...    echo '${kubeconfig}' | grep client-certificate-data | cut -d: -f2 | tr -d ' ' | base64 -d | openssl x509 -noout -dates
    ${cert_not_before}=    Command Should Work
    ...    echo '${cert_data}' | grep notBefore | cut -f2 -d'=' | xargs -I {} date -d "{}" +%s

    ${current_time}=    Command Should Work    date +%s

    Should Be True    ${cert_not_before} <= ${current_time}
    ...    msg=Kubeconfig client certificate NotBefore (${cert_not_before}) is after current time (${current_time})
