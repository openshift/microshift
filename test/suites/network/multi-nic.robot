*** Settings ***
Documentation       Tests related to MicroShift running in a multiple NIC environment

Resource            ../../resources/ostree-health.resource
Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/microshift-host.resource
Resource            ../../resources/microshift-network.resource
Resource            ../../resources/kubeconfig.resource
Library             Collections

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           network


*** Variables ***
${USHIFT_HOST_IP1}      ${EMPTY}
${USHIFT_HOST_IP2}      ${EMPTY}
${NIC1_NAME}            ${EMPTY}
${NIC2_NAME}            ${EMPTY}
${NICS_COUNT}           2
${NMCLI_CMD}            nmcli -f name,type connection | awk '$2 == "ethernet" {print $1}' | sort
${OSSL_CMD}             openssl x509 -text -noout -in
${CERT_FILE}            /var/lib/microshift/certs/kube-apiserver-external-signer/kube-external-serving/server.crt
${GREP_SUBJ_IPS}        grep -A1 'Subject Alternative Name:' | tail -1


*** Test Cases ***
Verify MicroShift Runs On Both NICs
    [Documentation]    Verify MicroShift can run in the default configuration

    # Wait for MicroShift API readiness and run verification
    Wait For MicroShift
    Verify Hello MicroShift LB
    Verify Hello MicroShift NodePort    ${USHIFT_HOST_IP1}
    Verify Hello MicroShift NodePort    ${USHIFT_HOST_IP2}
    IP Should Be Present In External Certificate    ${USHIFT_HOST_IP1}
    IP Should Be Present In External Certificate    ${USHIFT_HOST_IP2}

Verify MicroShift Runs Only On Primary NIC
    [Documentation]    Verify MicroShift can run only on the primary NIC. The
    ...    node IP will be taken from this NIC by default. When disabling the
    ...    secondary NIC nothing will happen in MicroShift, as the IP did not
    ...    change. A restart is forced so that MicroShift picks up the new
    ...    configuration (without the secondary IP) and regenerates the
    ...    certificates, which will be lacking the IP from secondary NIC.
    [Setup]    Save Default MicroShift Config

    Configure Subject Alternative Name    ${USHIFT_HOST_IP1}

    ${cur_pid}=    MicroShift Process ID

    Login Switch To IP    ${USHIFT_HOST_IP1}
    Disable Interface    ${NIC2_NAME}

    Restart MicroShift
    Restart Greenboot And Wait For Success

    Verify MicroShift On Single NIC    ${USHIFT_HOST_IP1}    ${USHIFT_HOST_IP2}

    [Teardown]    Restore Network Configuration By Rebooting Host

Verify MicroShift Runs Only On Secondary NIC
    [Documentation]    Verify MicroShift can run only on the secondary NIC. The
    ...    node IP will change when disabling the primary interface, triggering
    ...    an automatic restart of the service. After restarting, the node IP will
    ...    be that of the secondary NIC, and certificates will be updated according
    ...    to the new configuration (which includes only the secondary IP).
    [Setup]    Save Default MicroShift Config

    Configure Subject Alternative Name    ${USHIFT_HOST_IP2}

    ${cur_pid}=    MicroShift Process ID

    Login Switch To IP    ${USHIFT_HOST_IP2}
    Disable Interface    ${NIC1_NAME}

    Wait Until MicroShift Process ID Changes    ${cur_pid}
    Wait For MicroShift Service
    Setup Kubeconfig
    Wait For MicroShift

    Verify MicroShift On Single NIC    ${USHIFT_HOST_IP2}    ${USHIFT_HOST_IP1}

    [Teardown]    Restore Network Configuration By Rebooting Host


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Login MicroShift Host
    Initialize Global Variables
    Setup Suite With Namespace
    Verify Multiple NICs
    Wait Until Greenboot Health Check Exited

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host

Verify Multiple NICs
    [Documentation]    Verifies that the host has two Ethernet network interfaces

    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    ${NMCLI_CMD} | wc -l
    ...    return_stdout=True    return_stderr=True    return_rc=True
    Should Be Equal As Integers    ${rc}    0
    Should Be Equal As Strings    ${stdout}    ${NICS_COUNT}

Initialize Global Variables
    [Documentation]    Initializes global variables.

    Log    Host: ${USHIFT_HOST_IP1} ${USHIFT_HOST_IP2}
    Should Not Be Empty    ${USHIFT_HOST_IP1}    USHIFT_HOST_IP1 variable is required
    Should Not Be Empty    ${USHIFT_HOST_IP2}    USHIFT_HOST_IP2 variable is required
    Initialize Nmcli Variables

Initialize Nmcli Variables
    [Documentation]    Initialize the variables on the host

    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    ${NMCLI_CMD} | head -1
    ...    return_stdout=True    return_stderr=True    return_rc=True
    Should Be Equal As Integers    ${rc}    0
    Set Suite Variable    \${NIC1_NAME}    ${stdout}

    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    ${NMCLI_CMD} | tail -1
    ...    return_stdout=True    return_stderr=True    return_rc=True
    Should Be Equal As Integers    ${rc}    0
    Set Suite Variable    \${NIC2_NAME}    ${stdout}

Disable Interface
    [Documentation]    Disable NIC given in ${conn_name}. Change is not persistent. On
    ...    the next reboot the interface will have its original status again.
    [Arguments]    ${conn_name}
    ${stderr}    ${rc}=    Execute Command
    ...    nmcli connection down ${conn_name}
    ...    sudo=True    return_stdout=False    return_stderr=True    return_rc=True
    Should Be Equal As Integers    ${rc}    0

Login Switch To IP
    [Documentation]    Switch to using the specified IP for SSH connections
    [Arguments]    ${new_ip}

    IF    '${USHIFT_HOST}'!='${new_ip}'
        Logout MicroShift Host
        Set Global Variable    \${USHIFT_HOST}    ${new_ip}
        Login MicroShift Host
    END

Verify Hello MicroShift NodePort
    [Documentation]    Run Hello MicroShift NodePort verification
    [Arguments]    ${ip}
    Create Hello MicroShift Pod
    Expose Hello MicroShift Pod Via NodePort

    Wait Until Keyword Succeeds    30x    10s
    ...    Access Hello Microshift    ${NP_PORT}    ${ip}

    [Teardown]    Run Keywords
    ...    Delete Hello MicroShift Pod And Service

Verify MicroShift On Single NIC
    [Documentation]    Generic procedure to verify MicroShift network
    ...    functionality while one of the network interfaces is down.
    [Arguments]    ${login_ip}    ${removed_ip}

    Verify Hello MicroShift LB
    Verify Hello MicroShift NodePort    ${login_ip}
    IP Should Be Present In External Certificate    ${login_ip}
    IP Should Not Be Present In External Certificate    ${removed_ip}

Configure Subject Alternative Name
    [Documentation]    Replace subjectAltNames entries in the configuration
    ...    to include only the one provided in ${ip}.
    [Arguments]    ${ip}

    ${subject_alt_names}=    CATENATE    SEPARATOR=\n
    ...    ---
    ...    apiServer:
    ...    \ \ subjectAltNames:
    ...    \ \ - ${ip}

    ${replaced}=    Replace MicroShift Config    ${subject_alt_names}
    Upload MicroShift Config    ${replaced}

Check IP Certificate
    [Documentation]    Checks whether the ${ip} is present in the subject
    ...    alternative names in ${CERT_FILE}.
    [Arguments]    ${ip}    ${grep_opt}

    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    ${OSSL_CMD} ${CERT_FILE} | ${GREP_SUBJ_IPS} | grep -w ${grep_opt} 'DNS:${ip}'
    ...    sudo=True    return_stdout=True    return_stderr=True    return_rc=True
    Should Be Equal As Integers    ${rc}    0

IP Should Be Present In External Certificate
    [Documentation]    Check the specified IP presence in the external
    ...    certificate file
    [Arguments]    ${ip}

    Check IP Certificate    ${ip}    ${EMPTY}

IP Should Not Be Present In External Certificate
    [Documentation]    Check the specified IP absence in the external
    ...    certificate file
    [Arguments]    ${ip}

    Check IP Certificate    ${ip}    '--invert-match'

Restore Network Configuration By Rebooting Host
    [Documentation]    Restores network interface initial configuration
    ...    by rebooting the host.
    Restore Default MicroShift Config
    Reboot MicroShift Host
    Login Switch To IP    ${USHIFT_HOST_IP1}
    Wait Until Greenboot Health Check Exited
