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


*** Test Cases ***
Verify MicroShift Runs On Both NICs
    [Documentation]    Verify MicroShift can run in the default configuration

    # Wait for MicroShift API readiness and run verification
    Wait For MicroShift
    Verify Hello MicroShift LB
    Verify Hello MicroShift NodePort    ${USHIFT_HOST_IP1}    ${USHIFT_HOST_IP2}

Verify MicroShift Runs On First NIC
    [Documentation]    Verify MicroShift can run on the first NIC
    Verify MicroShift On One NIC    ${USHIFT_HOST_IP1}    ${NIC2_NAME}    ${USHIFT_HOST_IP1}    ${EMPTY}

Verify MicroShift Runs On Second NIC
    [Documentation]    Verify MicroShift can run on the second NIC
    Verify MicroShift On One NIC    ${USHIFT_HOST_IP2}    ${NIC1_NAME}    ${EMPTY}    ${USHIFT_HOST_IP2}


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Login MicroShift Host
    Initialize Global Variables
    Setup Suite With Namespace
    Verify Multiple NICs
    Wait For Healthy System

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host

Verify Multiple NICs
    [Documentation]    Verifies that the host has two Ethernet network interfaces

    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    ${NMCLI_CMD} | wc -l
    ...    return_stdout=True    return_stderr=True    return_rc=True
    Log    ${stderr}
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
    Log    ${stderr}
    Should Be Equal As Integers    ${rc}    0
    Set Suite Variable    \${NIC1_NAME}    ${stdout}

    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    ${NMCLI_CMD} | tail -1
    ...    return_stdout=True    return_stderr=True    return_rc=True
    Log    ${stderr}
    Should Be Equal As Integers    ${rc}    0
    Set Suite Variable    \${NIC2_NAME}    ${stdout}

Nmcli Connection Control
    [Documentation]    Run nmcli connection command with arguments
    [Arguments]    ${command}    ${conn_name}
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    nmcli connection ${command} ${conn_name}
    ...    sudo=True return_stdout=False    return_stderr=True    return_rc=True
    Log    ${stderr}
    Should Be Equal As Integers    ${rc}    0

Login Switch To IP
    [Documentation]    Switch to using the specified IP for SSH connections
    [Arguments]    ${new_ip}

    IF    '${USHIFT_HOST}'!='${new_ip}'
        Logout MicroShift Host
        Set Global Variable    \${USHIFT_HOST}    ${new_ip}
        Login MicroShift Host
    END

Login Switch To IP1
    [Documentation]    Switch to using the first IP for SSH connections
    Login Switch To IP    ${USHIFT_HOST_IP1}

Login Switch To IP2
    [Documentation]    Switch to using the second IP for SSH connections
    Login Switch To IP    ${USHIFT_HOST_IP2}

Verify Hello MicroShift NodePort
    [Documentation]    Run Hello MicroShift NodePort verification
    [Arguments]    ${ip1}    ${ip2}
    Create Hello MicroShift Pod
    Expose Hello MicroShift Pod Via NodePort

    IF    '${ip1}'!='${EMPTY}'
        Wait Until Keyword Succeeds    30x    10s
        ...    Access Hello Microshift    ${NP_PORT}    ${ip1}
    END
    IF    '${ip2}'!='${EMPTY}'
        Wait Until Keyword Succeeds    30x    10s
        ...    Access Hello Microshift    ${NP_PORT}    ${ip2}
    END

    [Teardown]    Run Keywords
    ...    Delete Hello MicroShift Pod And Service

Verify MicroShift On One NIC
    [Documentation]    Generic procedure to verify MicroShift network
    ...    functionality while one of the network interfaces is down.
    [Arguments]    ${login_ip}    ${down_nic}    ${verify_ip1}    ${verify_ip2}

    ${cur_pid}=    MicroShift Process ID

    Login Switch To IP    ${login_ip}
    Nmcli Connection Control    down    ${down_nic}

    # MicroShift should restart due to IP change
    Wait Until MicroShift Process ID Changes    ${cur_pid}
    Wait For MicroShift Service
    Setup Kubeconfig

    # Wait for MicroShift API readiness and run verification
    Wait For MicroShift
    Verify Hello MicroShift LB
    Verify Hello MicroShift NodePort    ${verify_ip1}    ${verify_ip2}

    # Rebooting MicroShift host restores the network configuration
    [Teardown]    Run Keywords
    ...    Reboot MicroShift Host
    ...    Login Switch To IP1
    ...    Wait For Healthy System
