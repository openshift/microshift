*** Settings ***
Documentation     Tests that MicroShift is healthy when no nic is up on MicroShift Host

Library           Process
Library           OperatingSystem
Library           SSHLibrary

Resource    ../../resources/libvirt.resource
Resource    ../../resources/microshift-host.resource
Resource    ../../resources/microshift-process.resource
Resource    ../../resources/microshift-config.resource
Resource    ../../resources/microshift-network.resource
Resource    ../../resources/common.resource

Suite Setup     Setup Test
Suite Teardown  Teardown Test


*** Variables ***
${VM_NAME}     ${EMPTY}    # passed by scenario.sh, also used as hostname
${VM_USERNAME}    redhat
${VM_PASSWORD}    redhat
${LO_IP}   10.44.0.1


*** Test Cases ***
MicroShift First Boot Is Healthy With No Gateway
    [Documentation]     Tests that MicroShift is healthy when no nic is present on MicroShift Host
    # Wait for the system to reboot, and for the test to run, and for the system to reboot again
    Log To Console    Start test case
    Log To Console    Logging into MicroShift host


    libvirt.Reboot MicroShift Host    ${VM_NAME}
    # The test script running on the MicroShift host will reboot the system after it completes.

    ${result}=  Run Process    virsh    --connect\=qemu:///system event --domain ${VM_NAME} --event reboot
    Should Be Equal As Integers    ${result.rc}    0
    Log Many   ${result.stdout}    ${result.stderr}

    Enable NIC For MicroShift Host    ${VM_NAME}    ${VNET_IFACE}

    Wait Until Keyword Succeeds    20m    15s
    ...    Is System Rebooted    ${BOOT_ID}

    libvirt.Reboot MicroShift Host    ${VM_NAME}
    Wait Until Keyword Succeeds    5m    15s
    ...    Make New SSH Connection
    ${stdout}  ${stderr}  ${rc}=    Execute Command    cat /var/lib/foobar.txt
    ...    sudo=True    return_rc=True    return_stderr=True    return_stdout=True
    Log Many    ${stdout}    ${stderr}
    Should Be Equal As Integers    ${rc}    0
    Log To Console    Logged in, checking script output
    Should Be Equal    ${stdout}    foobar

    Log To Console    Test Complete


*** Keywords ***
Setup Test
    [Documentation]    Setup the test environment, disabling the NIC, and rebooting the MicroShift host.
    Log To Console    Setting up test
    Should Not Be Empty    ${VM_NAME}
    Setup Suite With Namespace
    Initialize Suite Variables

    Save Default MicroShift Config
    ${boot_id}=    Get Current Boot Id
    Set Suite Variable    ${BOOT_ID}    ${boot_id}

    Setup Headless Testing Files
    Systemctl Daemon Reload
    ${stdout}  ${stderr}  ${rc}=    Execute Command
    ...    restorecon -Rv /etc/systemd/system
    ...    sudo=True    return_rc=True    return_stderr=True    return_stdout=True
    Log Many    ${stdout}    ${stderr}
    Should Be Equal As Integers    ${rc}    0

    Stop MicroShift
    Systemctl    stop    kubepods.slice
    Cleanup MicroShift    --all    --keep-images

    Systemctl    enable    test-microshift-network.service

    Update MicroShift Config For Loopback NIC
    Configure Loopback As NIC
    Disable NIC For MicroShift Host    ${VM_NAME}    ${VNET_IFACE}

    Log To Console    Setup Complete

Teardown Test
    # create a better detailed documentation for this keyword
    [Documentation]    Teardown the test environment, re-enabling the NIC, and rebooting the MicroShift host.
    Log to Console    Tearing down test
    Enable NIC For MicroShift Host    ${VM_NAME}    ${VNET_IFACE}
    libvirt.Reboot MicroShift Host    ${VM_NAME}
    Wait Until Keyword Succeeds    10x    15s
    ...     Make New SSH Connection
    ${stdout}  ${stderr}  ${rc}=    Execute Command
    ...    nmcli c del stable-microshift
    ...    sudo=True    return_rc=True    return_stderr=True    return_stdout=True
    Log Many    ${stdout}    ${stderr}    ${rc}
    Restore Default MicroShift Config
    Tear Down Suite With Namespace
    Logout MicroShift Host
    Log to Console    Teardown Complete

Configure Loopback As NIC
    [Documentation]    Configure the loopback device as the NIC for the MicroShift host.
    ${cmd}=    Catenate    sudo /bin/bash -c "
    ...    nmcli c add type loopback con-name stable-microshift ifname lo ip4 ${LO_IP}/32 &&
    ...    nmcli c modify stable-microshift ipv4.ignore-auto-dns yes &&
    ...    nmcli c modify stable-microshift ipv4.dns ${LO_IP} &&
    ...    bash -c 'cat >> /etc/hosts' <<<${LO_IP} ${VM_NAME}"

    ${stdout}  ${stderr}  ${rc}=    Execute Command
    ...    ${cmd}    return_rc=True  return_stdout=True  return_stderr=True
    Log Many    ${stdout}    ${stderr}
    Should Be Equal As Integers     ${rc}   0

Update MicroShift Config For Loopback NIC
    [Documentation]    Configure MicroShift to use the loopback device as the NIC.
    ${patch}=   Catenate    SEPARATOR=\n
    ...    node:
    ...    \  hostnameOverride: ${VM_NAME}
    ...    \  nodeIP: ${LO_IP}
    ${new_config}=    Extend MicroShift Config    ${patch}
    Upload MicroShift Config    ${new_config}

Initialize Suite Variables
    [Documentation]    Gets the vnet device name for the VM name.
    ${device}=    Get Vnet Device For MicroShift Host    ${VM_NAME}
    Set Suite Variable    $VNET_IFACE    ${device}

    ${nic}=    Get MicroShift Host Active NIC
    Set Suite Variable    $NIC    ${nic}

Setup Headless Testing Files
    [Documentation]    Uploads the test files to the MicroShift host.

    ${test_dir}=    Set Variable    /var/lib/microshift-headless-test

    ${stdout}  ${stderr}  ${rc}=    Execute Command
    ...     mkdir -p ${test_dir}
    ...     sudo=True   return_rc=True  return_stdout=True  return_stderr=True
    Log Many    ${test_dir}    ${stderr}
    Should Be Equal As Integers     ${rc}   0

    FOR  ${file}    IN
    ...    assets/hello-microshift.yaml
    ...    assets/hello-microshift-ingress.yaml
    ...    assets/isolated-lb-service.yaml
    ...    assets/headless-tests/test-microshift-network.sh
        ${content}=    OperatingSystem.Get File    ${file}
        Should Not Be Empty    ${content}
        ${_}  ${file}=    Split Path    ${file}
        Upload String To File    ${content}    ${test_dir}/${file}
    END
    Deploy Test Service    ${test_dir}/test-microshift-network.sh

Deploy Test Service
    [Documentation]    Returns a string of a systemd unit that executes the test-microshift-network.sh script only once
    [Arguments]    ${script_path}
    ${unit}=    Catenate    SEPARATOR=\n
    ...    [Unit]
    ...    Description=Test MicroShift Network
    ...    After=network-online.target
    ...    Wants=network-online.target
    ...    After=greenboot-healthcheck.service
    ...    Requires=greenboot-healthcheck.service
    ...    ConditionPathExists=!/var/lib/SCRIPT_RAN
    ...
    ...    [Service]
    ...    Type=oneshot
    ...    ExecStart=/bin/bash ${script_path}
    ...    RemainAfterExit=yes
    ...
    ...    [Install]
    ...    WantedBy=multi-user.target
    Upload String To File    ${unit}    /etc/systemd/system/test-microshift-network.service

Get MicroShift Host Active NIC
    [Documentation]    Returns the active NIC for the MicroShift host.
    ${stdout}  ${stderr}  ${rc}=     Execute Command
    ...    ip -o a | awk '/${USHIFT_HOST}/{sub(/:/,"", $2); print $2}'
    ...    sudo=True   return_rc=True    return_stderr=True    return_stdout=True
    Log Many    ${stdout}    ${stderr}
    Should Be Equal As Integers    ${rc}    0
    Should Not Be Empty    ${stdout}
    RETURN    ${stdout}
