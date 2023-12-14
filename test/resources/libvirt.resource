*** Settings ***
Documentation       Keywords for interacting with the libvirt host via the virsh command.

Library             Process
Library             String


*** Keywords ***
Reboot MicroShift Host
    [Documentation]    Reboot the MicroShift host.
    [Arguments]    ${vm_name}
    ${result}=    Virsh    ${vm_name}    reboot
    Wait For MicroShift Host To Reboot    ${vm_name}
    RETURN    ${result}

Shutdown MicroShift Host
    [Documentation]    Shutdown the MicroShift host.
    [Arguments]    ${vm_name}
    ${result}=    Virsh    ${vm_name}    shutdown
    RETURN    ${result}

Start MicroShift Host
    [Documentation]    Start the MicroShift host.
    [Arguments]    ${vm_name}
    ${result}=    Virsh    ${vm_name}    start
    RETURN    ${result}

Wait For MicroShift Host To Reboot
    [Documentation]    Wait for the MicroShift host to reboot.
    [Arguments]    ${vm_name}    ${timeout}=300
    Wait Until Keyword Succeeds    ${timeout}s    5s
    ...    MicroShift Host Should Be Running    ${vm_name}

MicroShift Host Should Be Running
    [Documentation]    Verify that the MicroShift host rebooted successfully.
    [Arguments]    ${vm_name}
    ${result}=    Virsh    ${vm_name}    domstate
    ${state}=    Strip String    ${result.stdout}
    Should Be Equal As Strings    ${state}    running

MicroShift Host Should Be Shutdown
    [Documentation]    Verify that the MicroShift host shutdown successfully.
    [Arguments]    ${vm_name}
    ${result}=    Virsh    ${vm_name}    domstate
    ${state}=    Strip String    ${result.stdout}
    Should Be Equal As Strings    ${state}    shut off

Get Vnet Devices For MicroShift Host
    [Documentation]    Get the vnet device for the MicroShift host.
    [Arguments]    ${vm_name}
    ${result}=    Virsh    ${vm_name}    domifaddr
    ${vnet_ifaces}=    Get Regexp Matches    ${result.stdout}    vnet[0-9]+    flags=MULTILINE
    RETURN    ${vnet_ifaces}

Enable NIC For MicroShift Host
    [Documentation]    Enable the NIC on the MicroShift host.
    ...    ${vnet_iface} is the vnet device for the MicroShift host, found via the dom. It can be either name of the
    ...    device, i.e. vnet123, or the device's MAC address.    It is not the interface name on the host, i.e. eth0.
    [Arguments]    ${vm_name}    ${vnet_iface}
    ${result}=    Virsh    ${vm_name}    domif-setlink --state\=up --interface\=${vnet_iface}

Disable NIC For MicroShift Host
    [Documentation]    Disable the NIC on the MicroShift host.
    ...    ${vnet_iface} is the vnet device for the MicroShift host, found via the dom. It can be either name of the
    ...    device, i.e. vnet123, or the device's MAC address.    It is not the interface name on the host, i.e. eth0.
    [Arguments]    ${vm_name}    ${vnet_iface}
    ${result}=    Virsh    ${vm_name}    domif-setlink --state\=down --interface\=${vnet_iface}

Get MicroShift Host NIC State
    [Documentation]    Get the state of the NIC on the MicroShift host.
    ...    ${vnet_iface} is the vnet device for the MicroShift host, found via the dom. It can be either name of the
    ...    device, i.e. vnet123, or the device's MAC address.    It is not the interface name on the host, i.e. eth0.
    [Arguments]    ${vm_name}    ${vnet_iface}
    ${result}=    Virsh    ${vm_name}    domif-getlink --interface\=${vnet_iface}
    ${state}=    Remove String    ${result.stdout}    ${vnet_iface}    \:    \"
    ${state}=    Strip String    ${state}
    RETURN    ${state}

MicroShift Host NIC Has State
    [Documentation]    Verify that host's NIC has the specified state.
    ...    ${state} valid states are 'up' and 'down'.
    ...    ${vnet_iface} is the vnet device for the MicroShift host, found via the dom. It can be either name of the
    ...    device, i.e. vnet123, or the device's MAC address.    It is not the interface name on the host, i.e. eth0.
    [Arguments]    ${vm_name}    ${vnet_iface}    ${state}
    ${state_actual}=    Get MicroShift Host NIC State    ${vm_name}    ${vnet_iface}
    Should Be Equal As Strings    ${state_actual}    ${state}

Virsh
    [Documentation]    Run a virsh command.
    [Arguments]    ${domain}    ${command_with_args}
    ${result}=    Run Process
    ...    virsh --connect\=qemu:///system ${command_with_args} --domain\=${domain}
    ...    shell=True
    Log Many    ${result.stdout}    ${result.stderr}
    Should Be Equal As Integers    ${result.rc}    0
    RETURN    ${result}
