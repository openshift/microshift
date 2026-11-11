*** Settings ***
Documentation       Tests related to MicroShift automated certificate rotation.

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/microshift-host.resource
Resource            ../../resources/ostree-health.resource
Library             Collections
Library             DateTime

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           restart


*** Variables ***
${KUBE_SCHEDULER_CLIENT_CERT}       /var/lib/microshift/certs/kube-control-plane-signer/kube-scheduler/client.crt
${ETCD_SIGNER_CA}                   /var/lib/microshift/certs/etcd-signer/ca.crt
${ETCD_PEER_CERT}                   /var/lib/microshift/certs/etcd-signer/etcd-peer/peer.crt
${ETCD_SERVING_CERT}                /var/lib/microshift/certs/etcd-signer/etcd-serving/peer.crt
${ETCD_APISERVER_CLIENT_CERT}       /var/lib/microshift/certs/etcd-signer/apiserver-etcd-client/client.crt
${OSSL_CMD}                         openssl x509 -noout -dates -in
${OSSL_DATE_FORMAT}                 %b %d %Y
${TIMEDATECTL_DATE_FORMAT}          %Y-%m-%d %H:%M:%S
# A one-year standard certificate enters ExpiresSoon after about 153 days.
${FUTURE_DAYS}                      160


*** Test Cases ***
Manual Rotation Of Etcd Signer Certs
    [Documentation]    Verify that deleting etcd signer certificates and restarting
    ...    MicroShift causes them to be regenerated.
    [Tags]    etcd
    # Capture the original cert fingerprint before deletion
    ${original_fp}=    Get Cert Fingerprint    ${ETCD_SIGNER_CA}
    Command Should Work    bash -c 'rm -rf /var/lib/microshift/certs/etcd-signer/*'
    Restart MicroShift

    VAR    @{cert_files}=
    ...    ${ETCD_SIGNER_CA}
    ...    ${ETCD_PEER_CERT}
    ...    ${ETCD_SERVING_CERT}
    ...    ${ETCD_APISERVER_CLIENT_CERT}
    FOR    ${cert_file}    IN    @{cert_files}
        Verify Remote File Exists With Sudo    ${cert_file}
        Certificate Should Be Valid For Current Time    ${cert_file}
    END

    # Expiry comparison won't work: alignValidity() anchors all same-day
    # certs to the same notAfter (tomorrow_midnight + validity). Use
    # fingerprint instead — it covers key, serial, and all cert fields.
    ${new_fp}=    Get Cert Fingerprint    ${ETCD_SIGNER_CA}
    Should Not Be Equal As Strings    ${original_fp}    ${new_fp}
    ...    msg=CA cert fingerprint should differ after regeneration

Certificate Rotation
    [Documentation]    Performs Certificate Expiration Rotation test
    # Certificates expire at midnight of (tomorrow + validity). For short-lived certs,
    # validity is 365 days, so expiry = tomorrow + 365 days.
    ${first_cert_date}=    Compute Date After Days    365    ${OSSL_DATE_FORMAT}
    Certs Should Expire On    ${KUBE_SCHEDULER_CLIENT_CERT}    ${first_cert_date}
    # After moving the clock forward by FUTURE_DAYS, regenerated certs will expire at
    # (new tomorrow + 365). Since "new tomorrow" is (original tomorrow + FUTURE_DAYS + 1),
    # the expected expiry from the original date is: tomorrow + 366 + FUTURE_DAYS.
    ${cert_should_expire_in_days}=    Evaluate    366+${FUTURE_DAYS}
    ${cert_expiry_date}=    Compute Date After Days    ${cert_should_expire_in_days}    ${OSSL_DATE_FORMAT}
    ${future_date}=    Compute Date After Days    ${FUTURE_DAYS}    ${TIMEDATECTL_DATE_FORMAT}
    Change System Date To    ${future_date}
    Certs Should Expire On    ${KUBE_SCHEDULER_CLIENT_CERT}    ${cert_expiry_date}
    [Teardown]    Restore System Date

Renewal Recovers From An Expired CA
    [Documentation]    Verify that when an admin misses the renewal window and the CA has expired,
    ...    leaf-only renewal is refused with advice to use --ca, CA renewal succeeds, and the
    ...    renewed certificates are active and healthy after MicroShift starts. MicroShift is
    ...    stopped before the clock change because startup regenerates expired certificates by
    ...    itself, so a running service would test startup instead of the renew command.
    ...
    ...    Verifies the expired-state half of the OCPEDGE-2998 acceptance criteria
    ...    ("renew --serving regenerates serving leaves only (not CAs)" via its
    ...    refusal, "renew --ca regenerates CAs and all descendants" via the
    ...    recovery), and the OCPSTRAT-2899 goal that renewal is the admin's
    ...    recovery path when the rotation window was missed.
    ...    OCPSTRAT-2899 TC-24
    [Tags]    qe    ocpstrat-2899
    # No [Timeout] here: Robot timeouts are wall-clock based and this test
    # jumps the wall clock ten years forward, which would blow any timeout
    # instantly (false failure observed 2026-11-04). Hang protection comes
    # from the reconnecting teardown instead.
    Move Clock Past CA Expiry With MicroShift Stopped

    ${stdout}    ${stderr}    ${rc}=    Command Execution    microshift certs renew --serving -o json
    Should Not Be Equal As Integers    ${rc}    0
    ${error}=    Evaluate    json.loads($stderr)    modules=json
    Should Be Equal    ${error}[code]    RenewalFailed
    Should Contain    ${error}[message]    --ca

    ${output}=    Command Should Work    microshift certs renew --ca -o json
    ${result}=    Evaluate    json.loads($output)    modules=json
    Should Be Equal    ${result}[status]    pending

    Start MicroShift
    Wait For MicroShift
    Renewed Chain Should Be Active And Healthy    ${result}
    [Teardown]    Run Keywords    Start MicroShift    AND    Restore Date With Fresh SSH Connection


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig    # for readiness checks
    Wait For MicroShift Healthcheck Success

Teardown
    [Documentation]    Test suite teardown
    Restore System Date
    Logout MicroShift Host

Restore System Date
    [Documentation]    Reset MicroShift date to current date.
    ...    Skips if chronyd is already active (date was never changed).
    ${stdout}    ${rc}=    Execute Command
    ...    systemctl is-active chronyd
    ...    sudo=True    return_rc=True
    IF    '${stdout.strip()}' == 'active'    RETURN
    ${ushift_pid}=    MicroShift Process ID
    Start Chronyd Without Wall Clock Sleep
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
    All Certificates Should Be Valid For Current Time
    Setup Kubeconfig
    Wait For MicroShift Healthcheck Success

Compute Date After Days
    [Documentation]    return system date after number of days elapsed from midnight tomorrow
    [Arguments]    ${number_of_days}    ${date_format}
    # Certificates are aligned to expire at midnight of the next day + validity
    ${future_date}=    Command Should Work    TZ=UTC date "+${date_format}" -d "tomorrow + ${number_of_days} day"
    RETURN    ${future_date}

Certs Should Expire On
    [Documentation]    verify if the certificate expires at given date.
    [Arguments]    ${cert_file}    ${cert_expected_date}
    ${expiration_date}=    Command Should Work
    ...    ${OSSL_CMD} ${cert_file} | grep notAfter | cut -f2 -d'=' | awk '{printf ("%s %02d %d",$1,$2,$4)}'
    Should Be Equal As Strings    ${cert_expected_date}    ${expiration_date}

All Certificates Should Be Valid For Current Time
    [Documentation]    Wait for multiple certificate files to be regenerated and valid

    ${kubeconfig}=    Get Kubeconfig    ${USHIFT_HOST}
    ${kubeconfig_cert_file}=    Create Random Temp File
    Command Should Work
    ...    echo '${kubeconfig}' | grep client-certificate-data | cut -d: -f2 | tr -d ' ' | base64 -d > ${kubeconfig_cert_file}

    VAR    @{cert_files}=
    ...    /var/lib/microshift/certs/kube-control-plane-signer/kube-scheduler/client.crt
    ...    /var/lib/microshift/certs/kube-control-plane-signer/kube-controller-manager/client.crt
    ...    /var/lib/microshift/certs/admin-kubeconfig-signer/admin-kubeconfig-client/client.crt
    ...    ${kubeconfig_cert_file}
    FOR    ${cert_file}    IN    @{cert_files}
        Wait Until Keyword Succeeds    30x    5s
        ...    Certificate Should Be Valid For Current Time    ${cert_file}
    END

Certificate Should Be Valid For Current Time
    [Documentation]    Verify that certificate is valid for the current system time
    [Arguments]    ${cert_file}

    ${cert_not_before}=    Command Should Work
    ...    ${OSSL_CMD} ${cert_file} | grep notBefore | cut -f2 -d'=' | xargs -I {} date -d "{}" +%s
    ${cert_not_after}=    Command Should Work
    ...    ${OSSL_CMD} ${cert_file} | grep notAfter | cut -f2 -d'=' | xargs -I {} date -d "{}" +%s
    ${current_time}=    Command Should Work    date +%s
    Should Be True    ${cert_not_before} <= ${current_time}
    ...    msg=Certificate NotBefore (${cert_not_before}) is after current time (${current_time})
    Should Be True    ${current_time} <= ${cert_not_after}
    ...    msg=Certificate NotAfter (${cert_not_after}) is before current time (${current_time})

Get Cert Fingerprint
    [Documentation]    Return the SHA256 fingerprint of a certificate
    [Arguments]    ${cert_file}
    ${fingerprint}=    Command Should Work
    ...    openssl x509 -noout -fingerprint -sha256 -in ${cert_file}
    RETURN    ${fingerprint}

Certificate Status
    [Documentation]    Return the parsed `microshift certs status` JSON document.
    ${output}=    Command Should Work    microshift certs status -o json
    ${status}=    Evaluate    json.loads($output)    modules=json
    RETURN    ${status}

Renewed Chain Should Be Active And Healthy
    [Documentation]    After recovery startup every certificate is Healthy and the
    ...    active files are exactly the ones prepared by the pending renewal.
    [Arguments]    ${result}
    ${status}=    Certificate Status
    ${unhealthy}=    Evaluate    [i['name'] for i in $status['items'] if i['status'] != 'Healthy']
    Should Be Empty    ${unhealthy}    msg=Certificates not healthy after recovery: ${unhealthy}
    ${active}=    Evaluate    {(i['service'], i['name']): i['notAfter'] for i in $status['items']}
    # $active is passed via a namespace because the special `$var` syntax is not
    # visible inside the condition of a generator expression.
    VAR    &{ns}=    active=${active}
    ${not_renewed}=    Evaluate
    ...    [i['name'] for i in $result['items'] if active[(i['service'], i['name'])] != i['newNotAfter']]
    ...    namespace=${ns}
    Should Be Empty    ${not_renewed}    msg=Active certificates are not the ones prepared by renew: ${not_renewed}

Move Clock Past CA Expiry With MicroShift Stopped
    [Documentation]    Stop MicroShift and time sync, then set the clock two days after the latest
    ...    CA expiry date. Every CA is then expired, and startup cannot regenerate them first.
    ${status}=    Certificate Status
    ${latest_ca_expiry}=    Evaluate    max(i['notAfter'] for i in $status['items'] if i['role'] == 'ca')
    ${expired_date}=    Evaluate
    ...    (datetime.datetime.fromisoformat($latest_ca_expiry.replace('Z', '+00:00')) + datetime.timedelta(days=2)).strftime('%Y-%m-%d %H:%M:%S')
    ...    modules=datetime
    Stop MicroShift
    Systemctl    stop    chronyd
    Command Should Work    TZ=UTC timedatectl set-time "${expired_date}"

Restore Date With Fresh SSH Connection
    [Documentation]    Restore the system date after a large clock jump without
    ...    relying on any SSH connection surviving the time step. A multi-year
    ...    backward step breaks connections under SSHLibrary and the next
    ...    command on the dead channel blocks forever (observed 2026-11-04,
    ...    and again 2026-11-05: the step also killed a connection opened
    ...    immediately after chronyd started — chronyd selects a source and
    ...    steps within about five seconds, so a prompt reconnect races the
    ...    step and loses). Start chronyd, drop the connection at once
    ...    (Close All Connections is local and needs no remote round trip),
    ...    wait out the step with nothing connected, and only then log back
    ...    in. Small steps (days) do not break the connection, which is why
    ...    the other date-restoring keywords do not need this.
    Start Chronyd Without Wall Clock Sleep
    Close All Connections
    # Wait out chronyd's clock step with nothing connected. This must be a
    # monotonic sleep (Python time.sleep), not Robot's Sleep keyword: Robot
    # measures Sleep against the wall clock, so the backward step landing
    # mid-sleep turns "sleep 60s" into "sleep ten years" (observed 2026-11-11:
    # the teardown hung forever inside Sleep).
    Evaluate    time.sleep(60)    modules=time
    Login MicroShift Host
    Wait Until Keyword Succeeds    30x    5s    System Date Should Be Synchronized
    # The restart that follows the backward step regenerates certificates and
    # rewrites the kubeadmin kubeconfig (TC-24's own CA renewal also invalidates
    # the copy saved at suite setup). Re-read it before any healthcheck.
    Setup Kubeconfig
    Wait For MicroShift

System Date Should Be Synchronized
    [Documentation]    Fail unless NTP reports the system clock as synchronized.
    ${output}=    Command Should Work    timedatectl show -p NTPSynchronized --value
    Should Be Equal As Strings    ${output.strip()}    yes

Start Chronyd Without Wall Clock Sleep
    [Documentation]    Start chronyd without going through the shared Systemctl
    ...    keyword. Systemctl sleeps 5s on the wall clock after any start, and
    ...    chronyd steps the clock back (possibly by years) within seconds of
    ...    starting, turning that sleep into a multi-year hang (observed
    ...    2026-11-11). A bare command returns before the step lands; the
    ...    caller must tolerate the connection breaking or avoid using it
    ...    until the step completes.
    ${stdout}    ${rc}=    Execute Command
    ...    systemctl start chronyd
    ...    sudo=True    return_rc=True
    Should Be Equal As Integers    ${rc}    0    msg=Failed to start chronyd: ${stdout}
