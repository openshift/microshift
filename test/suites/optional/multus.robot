*** Settings ***
Documentation       Tests for Multus and Bridge plugin on MicroShift

Resource            ../../resources/common.resource
Resource            ../../resources/multus.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/microshift-rpm.resource

Suite Setup         Setup
Suite Teardown      Teardown Suite With Namespace


*** Variables ***
${BRIDGE_INTERFACE}         br-test
${BRIDGE_NAD_YAML}          ./assets/multus/bridge-nad.yaml
${BRIDGE_POD_YAML}          ./assets/multus/bridge-pod.yaml
${BRIDGE_POD_NAME}          test-bridge
${BRIDGE_IP}                10.10.0.10/24

${PE_BRIDGE_INTERFACE}      br-preexisting
${PE_BRIDGE_NAD_YAML}       ./assets/multus/bridge-preexisting-nad.yaml
${PE_BRIDGE_POD_YAML}       ./assets/multus/bridge-preexisting-pod.yaml
${PE_BRIDGE_POD_NAME}       test-bridge-preexisting
${PE_BRIDGE_IP}             10.10.1.10/24

${MACVLAN_NAD_YAML}         ./assets/multus/macvlan-nad.yaml
${MACVLAN_POD_YAML}         ./assets/multus/macvlan-pod.yaml
${MACVLAN_POD_NAME}         test-macvlan
${MACVLAN_MASTER}           ${EMPTY}

${IPVLAN_NAD_YAML}          ./assets/multus/ipvlan-nad.yaml
${IPVLAN_POD_YAML}          ./assets/multus/ipvlan-pod.yaml
${IPVLAN_POD_NAME}          test-ipvlan
${IPVLAN_MASTER}            ${EMPTY}


*** Test Cases ***
Pre-Existing Bridge Interface
    [Documentation]    Test verifies if Bridge CNI plugin will work correctly with pre-existing interface.
    [Setup]    Run Keywords
    ...    Interface Should Not Exist    ${PE_BRIDGE_INTERFACE}
    ...    AND
    ...    Create Interface    ${PE_BRIDGE_INTERFACE}
    ...    AND
    ...    Create NAD And Pod    ${PE_BRIDGE_NAD_YAML}    ${PE_BRIDGE_POD_YAML}
    ...    AND
    ...    Named Pod Should Be Ready    ${PE_BRIDGE_POD_NAME}    ${NAMESPACE}
    ...    AND
    ...    Interface Should Exist    ${PE_BRIDGE_INTERFACE}
    ...    AND
    ...    Set IP For Host Interface    ${PE_BRIDGE_INTERFACE}    ${PE_BRIDGE_IP}

    Connect To Pod Over Local Interface    ${PE_BRIDGE_POD_NAME}    ${NAMESPACE}    ${PE_BRIDGE_INTERFACE}

    [Teardown]    Cleanup Bridge Test
    ...    ${PE_BRIDGE_NAD_YAML}
    ...    ${PE_BRIDGE_POD_YAML}
    ...    ${PE_BRIDGE_INTERFACE}

No Pre-Existing Bridge Interface
    [Documentation]    Test verifies if Bridge CNI plugin will
    ...    work correctly if there is no pre-existing bridge
    ...    interface - it needs to be created.
    [Setup]    Run Keywords
    ...    Interface Should Not Exist    ${BRIDGE_INTERFACE}
    ...    AND
    ...    Create NAD And Pod    ${BRIDGE_NAD_YAML}    ${BRIDGE_POD_YAML}
    ...    AND
    ...    Named Pod Should Be Ready    ${BRIDGE_POD_NAME}    ${NAMESPACE}
    ...    AND
    ...    Interface Should Exist    ${BRIDGE_INTERFACE}
    ...    AND
    ...    Set IP For Host Interface    ${BRIDGE_INTERFACE}    ${BRIDGE_IP}

    Connect To Pod Over Local Interface    ${BRIDGE_POD_NAME}    ${NAMESPACE}    ${BRIDGE_INTERFACE}

    [Teardown]    Cleanup Bridge Test
    ...    ${BRIDGE_NAD_YAML}
    ...    ${BRIDGE_POD_YAML}
    ...    ${BRIDGE_INTERFACE}

Macvlan
    [Documentation]    Tests if Pod with macvlan plugin interface is accessible
    ...    from outside the MicroShift host.
    [Setup]    Run Keywords
    ...    Template And Create NAD And Pod    ${MACVLAN_NAD_YAML}    ${MACVLAN_POD_YAML}
    ...    AND
    ...    Named Pod Should Be Ready    ${MACVLAN_POD_NAME}    ${NAMESPACE}

    Wait Until Keyword Succeeds    5x    5s
    ...    Connect To Pod From The Hypervisor    ${MACVLAN_POD_NAME}    ${NAMESPACE}    ${NAMESPACE}/macvlan-conf

    [Teardown]    Remove NAD And Pod    ${MACVLAN_NAD_YAML}    ${MACVLAN_POD_YAML}

Ipvlan
    [Documentation]    Tests if Pod with ipvlan plugin interface is accessible
    ...    from outside the MicroShift host.
    [Setup]    Run Keywords
    ...    Template And Create NAD And Pod    ${IPVLAN_NAD_YAML}    ${IPVLAN_POD_YAML}
    ...    AND
    ...    Named Pod Should Be Ready    ${IPVLAN_POD_NAME}    ${NAMESPACE}

    Wait Until Keyword Succeeds    5x    5s
    ...    Connect To Pod From The Hypervisor    ${IPVLAN_POD_NAME}    ${NAMESPACE}    ${NAMESPACE}/ipvlan-conf

    [Teardown]    Remove NAD And Pod    ${IPVLAN_NAD_YAML}    ${IPVLAN_POD_YAML}


*** Keywords ***
Setup
    [Documentation]    Setup test suite
    Setup Suite With Namespace

    ${out}=    Command Should Work    ip route list default | cut -d' ' -f5
    @{enps}=    String.Split To Lines    ${out}
    ${len}=    Get Length    ${enps}
    Should Be True    ${len}>=2
    Set Suite Variable    ${MACVLAN_MASTER}    ${enps[0]}
    Set Suite Variable    ${IPVLAN_MASTER}    ${enps[1]}
    Verify MicroShift RPM Install

Template And Create NAD And Pod
    [Documentation]    Template NAD and create it along with Pod
    [Arguments]    ${nad_tpl}    ${pod}
    ${rand}=    Generate Random String
    ${nad_tpl_output}=    Join Path    /tmp    multus-templates-${rand}.yaml
    ${template}=    OperatingSystem.Get File    ${nad_tpl}
    ${contents}=    Replace Variables    ${template}
    OperatingSystem.Append To File    ${nad_tpl_output}    ${contents}
    Create NAD And Pod    ${nad_tpl_output}    ${pod}
    OperatingSystem.Remove File    ${nad_tpl_output}

Cleanup Bridge Test
    [Documentation]    Removes provided NetworkAttachmentDefinition, Pod and network interface to allow for test rerun.
    [Arguments]    ${nad}    ${pod}    ${if}
    Remove NAD And Pod    ${nad}    ${pod}
    Command Should Work    ip link delete ${if}

Connect To Pod From The Hypervisor
    [Documentation]    Makes a HTTP request to port 8080 of a given Pod from the hypervisor machine.
    ...    This is a limitation of macvlan devices - virtual devices cannot communicate with the master interface.
    [Arguments]    ${pod}    ${ns}    ${extra_cni_name}

    ${networks}=    Get And Verify Pod Networks    ${pod}    ${ns}    ${extra_cni_name}
    ${extra_ip}=    Set Variable    ${networks}[1][ips][0]
    Should Contain    ${extra_ip}    192.168.112

    ${result}=    Process.Run Process    curl    -v    ${extra_ip}:8080
    Should Contain    ${result.stdout}    Hello MicroShift

Interface Should Not Exist
    [Documentation]    Verifies that network interface does not exist.
    [Arguments]    ${if}
    Command Should Fail    ip link show ${if}

Create Interface
    [Documentation]    Creates network interface.
    [Arguments]    ${if}    ${type}=bridge
    Command Should Work    ip link add dev ${if} type ${type}

Interface Should Exist
    [Documentation]    Verifies that interface exists on the host.
    [Arguments]    ${if}
    Command Should Work    ip link show ${if}
