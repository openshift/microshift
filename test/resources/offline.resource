*** Settings ***
Documentation       Keyword for offline (isolated) test suites

Library             OperatingSystem
Library             qemu-guest-agent.py
Resource            libvirt.resource


*** Variables ***
${GUEST_NAME}       ${EMPTY}
${NODE_IP}          10.44.0.1


*** Keywords ***
Setup Suite
    [Documentation]    Setup offline test suite.
    Should Not Be Equal As Strings    ${GUEST_NAME}    ''    The guest name must be set.
    Wait Until Keyword Succeeds    5m    10s
    ...    Guest Agent Is Ready    ${GUEST_NAME}

    # Verify that no active ethernet connections exist. The grep
    # command returns 1 when no lines are selected.
    ${result}    ${ignore}=    Run Guest Process    ${GUEST_NAME}
    ...    bash
    ...    -c
    ...    nmcli connection show | grep -q ethernet
    Should Be Equal As Integers    ${result["rc"]}    1

    # Verify MicroShift successful start on first boot
    Wait For Greenboot Health Check To Exit

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
