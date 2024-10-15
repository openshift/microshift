*** Settings ***
Documentation       Tests related to running MicroShift on a disconnected host.

Library             OperatingSystem
Library             String
Library             ../../resources/qemu-guest-agent.py
Library             ../../resources/DataFormats.py
Resource            ../../resources/libvirt.resource
Resource            ../../resources/microshift-config.resource

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           offline


*** Variables ***
${GUEST_NAME}           ${EMPTY}

${NODE_IP}              10.44.0.1

${SERVER_POD}           assets/hello-microshift.yaml
${SERVER_INGRESS}       assets/hello-microshift-ingress.yaml
${SERVER_SVC}           assets/hello-microshift-service.yaml
${TEST_NS}              ${EMPTY}


*** Test Cases ***
MicroShift Should Boot Into Healthy State When Host Is Offline
    [Documentation]    Test that MicroShift starts when the host is offline.
    Reboot MicroShift Host
    Wait For Greenboot Health Check To Exit

Load Balancer Services Should Be Running Correctly When Host Is Offline
    [Documentation]    Test that the load balancer services are running correctly when the host is offline.
    [Setup]    Setup Test

    Wait Until Keyword Succeeds    2m    10s
    ...    Pod Should Be Reachable Via Ingress

    [Teardown]    Teardown Test


*** Keywords ***
# Suite Setup / Teardown
Setup
    [Documentation]    Test suite setup.
    Should Not Be Equal As Strings    ${GUEST_NAME}    ''    The guest name must be set.
    Wait Until Keyword Succeeds    5m    10s
    ...    Guest Agent Is Ready    ${GUEST_NAME}
    Wait Until Keyword Succeeds    5x    1s
    ...    Wipe MicroShift
    Update MicroShift Config
    Update Guest Network Config
    Systemctl    enable    microshift.service    --now
    Systemctl    restart    greenboot-healthcheck.service
    Wait For Greenboot Health Check To Exit

Teardown
    [Documentation]    Test suite teardown
    Restore Default MicroShift Config
    Restore Guest Network Config
    Wipe MicroShift

Setup Test
    [Documentation]    Deploy cluster resources to provide an endpoint for testing. Wait for greenboot
    Wait For Greenboot Health Check To Exit
    ${ns}=    Create Test Namespace
    Send File    ${SERVER_SVC}    /tmp/hello-microshift-service.yaml
    Run With Kubeconfig    oc    create    -n\=${ns}    -f\=/tmp/hello-microshift-service.yaml
    Send File    ${SERVER_INGRESS}    /tmp/hello-microshift-ingress.yaml
    Run With Kubeconfig    oc    create    -n\=${ns}    -f\=/tmp/hello-microshift-ingress.yaml
    Send File    ${SERVER_POD}    /tmp/hello-microshift.yaml
    Run With Kubeconfig    oc    create    -n\=${ns}    -f\=/tmp/hello-microshift.yaml
    # Set this last, it's not available within this scope anyway
    Run With Kubeconfig    oc    wait    -n\=${ns}    --for\=condition=Ready    pod/hello-microshift
    Set Test Variable    ${TEST_NS}    ${ns}

Teardown Test
    [Documentation]    Test teardown
    Run With Kubeconfig    oc    delete    -n\=${TEST_NS}    -f\=/tmp/hello-microshift-service.yaml
    Run With Kubeconfig    oc    delete    -n\=${TEST_NS}    -f\=/tmp/hello-microshift-ingress.yaml
    Run With Kubeconfig    oc    delete    -n\=${TEST_NS}    -f\=/tmp/hello-microshift.yaml
    Run With Kubeconfig    oc    delete    namespace    ${TEST_NS}

Wipe MicroShift
    [Documentation]    Execute microshift-cleanup-data on the guest.
    ${result}    ${exited}=    Run Guest Process    ${GUEST_NAME}
    ...    bash    -c    microshift-cleanup-data --all --keep-images <<<1
    Log Many    ${result["stdout"]}    ${result["stderr"]}
    Should Be Equal As Integers    ${result["rc"]}    0

Pod Should Be Reachable Via Ingress
    [Documentation]    Checks if a load balancer service can be accessed. The curl command must be executed in a
    ...    subshell.
    ${result}    ${ignore}=    Wait Until Keyword Succeeds    5x    1s
    ...    Run Guest Process    ${GUEST_NAME}
    ...        bash
    ...        -c
    ...        curl --fail -I -H \"Host: hello-microshift.cluster.local\" ${NODE_IP}:80/principal
    Log Many    ${result["stdout"]}    ${result["stderr"]}
    Should Be Equal As Integers    ${result["rc"]}    0

Update MicroShift Config
    [Documentation]    Update the MicroShift config to use the offline-network endpoint
    ${patch}=    Catenate    SEPARATOR=\n
    ...    node:
    ...    \ \ hostnameOverride: ${GUEST_NAME}
    ...    \ \ nodeIP: ${NODE_IP}
    Write To File    ${GUEST_NAME}    /etc/microshift/config.d/10-hostname.yaml    ${patch}

Update Guest Network Config
    [Documentation]    Setup the guest's loopback with IP address, hostname, and DNS
    ${data}=    Read From File    ${GUEST_NAME}    /etc/hosts
    Set Suite Variable    ${DEFAULT_HOSTS_FILE}    ${data}

    ${data}=    Set Variable    ${EMPTY}
    TRY
        ${data}=    Read From File    ${GUEST_NAME}    /etc/resolv.conf
    EXCEPT    *No such file or directory*    type=glob
        Log    Default MicroShift config not found, ignoring
    END
    Set Suite Variable    ${DEFAULT_RESOLV_FILE}    ${data}

    Write To File    ${GUEST_NAME}    /etc/hosts    \n${NODE_IP} ${GUEST_NAME}\n    append=${TRUE}
    Write To File    ${GUEST_NAME}    /etc/resolv.conf    \nnameserver 10.44.1.1\n    append=${TRUE}
    Configure Network Manager

Configure Network Manager
    [Documentation]    Configure Network Manager to use the loopback interface
    ${result}    ${exited}=    Run Guest Process    ${GUEST_NAME}
    ...    nmcli    connection    modify    lo    +ipv4.addresses    ${NODE_IP}/32
    Should Be Equal As Integers    ${result["rc"]}    0
    ${result}    ${exited}=    Run Guest Process    ${GUEST_NAME}
    ...    nmcli    connection    modify    lo    autoconnect    yes
    Should Be Equal As Integers    ${result["rc"]}    0
    ${result}    ${exited}=    Run Guest Process    ${GUEST_NAME}    hostnamectl    set-hostname    ${GUEST_NAME}
    Should Be Equal As Integers    ${result["rc"]}    0

Restore Guest Network Config
    [Documentation]    Restore the guest's configuration files
    Write To File    ${GUEST_NAME}    /etc/hosts    ${DEFAULT_HOSTS_FILE}

    ${len}=    Get Length    ${DEFAULT_RESOLV_FILE}
    IF    ${len} > 0
        Write To File    ${GUEST_NAME}    /etc/resolv.conf    ${DEFAULT_RESOLV_FILE}
    ELSE
        ${result}    ${exited}=    Run Guest Process    ${GUEST_NAME}    rm    /etc/resolv.conf
        Should Be Equal As Integers    ${result["rc"]}    0
    END

    ${result}    ${exited}=    Run Guest Process    ${GUEST_NAME}
    ...    nmcli    connection    modify    lo    -ipv4.addresses    ${NODE_IP}/32
    Should Be Equal As Integers    ${result["rc"]}    0
    ${result}    ${exited}=    Run Guest Process    ${GUEST_NAME}
    ...    nmcli    connection    modify    lo    connection.autoconnect    no
    Should Be Equal As Integers    ${result["rc"]}    0

Restore Default MicroShift Config
    [Documentation]    Restore the default MicroShift config for offline start.
    ${result}    ${ignore}=    Run Guest Process
    ...    ${GUEST_NAME}
    ...    rm
    ...    -f
    ...    /etc/microshift/config.d/10-hostname.yaml
    Log Many    ${result["stdout"]}    ${result["stderr"]}
    Should Be Equal As Integers    ${result["rc"]}    0

Start MicroShift
    [Documentation]    Restart the MicroShift systemd service
    ${result}    ${ignore}=    Wait Until Keyword Succeeds    5m    5s
    ...    Systemctl    enable    microshift.service    --now
    Log Many    ${result["stdout"]}    ${result["stderr"]}
    Should Be Equal As Integers    ${result["rc"]}    0

Wait For Greenboot Health Check To Exit
    [Documentation]    Wait for the Greenboot Health Check systemd service to exit
    Wait Until Keyword Succeeds    10m    15s
    ...    Greenboot Health Check Exited

Greenboot Health Check Exited
    [Documentation]    Check that the Greenboot Health Check systemd service has state "exited"
    ${result}    ${exited}=    Wait Until Keyword Succeeds    5x    5s
    ...    Run Guest Process    ${GUEST_NAME}
    ...        systemctl
    ...        show
    ...        --property\=SubState
    ...        --value
    ...        greenboot-healthcheck.service
    Should Be Equal As Integers    ${result["rc"]}    0
    Should Be Equal As Strings    ${result["stdout"]}    exited

Reboot MicroShift Host
    [Documentation]    Reboot the MicroShift host and wait for the boot ID to change. This has the intended
    ...    side effect of waiting for the qemu-guest-agent service to come back online after the reboot.
    ${boot_id}=    Guest Has Boot ID
    libvirt.Reboot MicroShift Host    ${GUEST_NAME}
    Wait Until Keyword Succeeds    5m    1s
    ...    System Is Rebooted    ${boot_id}

System Is Rebooted
    [Documentation]    Verify that the guest's boot ID matches the specified boot ID.
    [Arguments]    ${boot_id}
    ${boot_id_actual}=    Guest Has Boot ID
    Should Not Be Equal As Strings    ${boot_id}    ${boot_id_actual}

Guest Has Boot ID
    [Documentation]    Get the boot ID of the guest.
    ${boot_id}=    Read From File    ${GUEST_NAME}    /proc/sys/kernel/random/boot_id
    ${len}=    Get Length    ${boot_id}
    Should Not Be Equal As Integers    ${len}    0
    RETURN    ${boot_id}

Send File
    [Documentation]    Send a file to the guest. Does not retain permissions or ownership.
    [Arguments]    ${src}    ${dest}
    ${data}=    OperatingSystem.Get File    ${src}
    ${len}=    Get Length    ${data}
    ${w_len}=    Write To File    ${GUEST_NAME}    ${dest}    ${data}
    Should Be Equal As Integers    ${len}    ${w_len}

Run With Kubeconfig
    [Documentation]    Run a guest-command with the KUBECONFIG environment variable set
    ...    ${command}    The command to run. Should but `oc` or `kubectl` but this is not enforced
    ...    @{args}    The arguments to pass to the command. See ../../resources/qemu-guest-agent.py for syntax
    [Arguments]    ${command}    @{args}
    ${env}=    Create Dictionary    KUBECONFIG=/var/lib/microshift/resources/kubeadmin/kubeconfig
    ${result}    ${ignore}=    Wait Until Keyword Succeeds    5x    2s
    ...    Run Guest Process    ${GUEST_NAME}    ${command}    @{args}    env=&{env}
    Log Many    ${result["stdout"]}    ${result["stderr"]}
    Should Be Equal As Integers    ${result["rc"]}    0
    RETURN    ${result}

Systemctl
    [Documentation]    Run systemctl on the guest
    [Arguments]    ${verb}    ${service}    @{args}
    ${result}    ${exited}=    Wait Until Keyword Succeeds    5m    10s
    ...    Run Guest Process    ${GUEST_NAME}    systemctl    ${verb}    ${service}    @{args}
    Log Many    ${result["stdout"]}    ${result["stderr"]}
    Should Be Equal As Integers    ${result["rc"]}    0

Create Test Namespace
    [Documentation]    Create a test namespace with a random name

    ${rand}=    Generate Random String
    ${rand}=    Convert To Lower Case    ${rand}
    ${ns}=    Set Variable    test-${rand}
    Wait Until Keyword Succeeds    5m    10s
    ...    Run With Kubeconfig    oc    create    namespace    ${ns}
    RETURN    ${ns}
