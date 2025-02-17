*** Settings ***
Documentation       Verify MicroShift host name resolution with and without systemd-resolved installed

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-host.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/ostree.resource
Resource            ../../resources/ostree-health.resource
Resource            ../../resources/systemd.resource

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           slow


*** Variables ***
${KUBELET_CONFIG_FILE}      /var/lib/microshift/resources/kubelet/config/config.yaml
${RESOLVE_CONF_FILE}        /run/systemd/resolve/resolv.conf


*** Test Cases ***
Verify Kubelet Config With Systemd-Resolved Running
    [Documentation]    Verify kubelet uses the upstream resolve file of
    ...    systemd-resolved when the service is running
    [Setup]    Run Keywords
    ...    Systemd-Resolved Must Be Installed And Disabled
    ...    Start Systemd-Resolved

    Delete Kubelet Configuration File
    Restart MicroShift

    # Verify the presence of the kubelet option
    ${rc}=    Check ResolveConf Option Presence
    Should Be Equal As Integers    0    ${rc}

    [Teardown]    Run Keywords
    ...    Stop Systemd-Resolved

Verify Kubelet Config With Systemd-Resolved Disabled
    [Documentation]    Verify kubelet does not use the upstream resolve file of
    ...    systemd-resolved when the service is disabled
    [Setup]    Run Keywords
    ...    Systemd-Resolved Must Be Installed And Disabled

    Delete Kubelet Configuration File
    Restart MicroShift

    # Verify the absence of the kubelet option
    ${rc}=    Check ResolveConf Option Presence
    Should Not Be Equal As Integers    0    ${rc}

Verify Kubelet Config With Systemd-Resolved Uninstalled
    [Documentation]    Verify kubelet does not use the upstream resolve file of
    ...    systemd-resolved when the package is not present
    [Setup]    Run Keywords
    ...    Uninstall Systemd-Resolved
    ...    Systemd-Resolved Must Not Be Installed

    Delete Kubelet Configuration File
    Restart MicroShift

    # Verify the absence of the kubelet option
    ${rc}=    Check ResolveConf Option Presence
    Should Not Be Equal As Integers    0    ${rc}

    # Revert the system to the original configuration
    [Teardown]    Run Keywords
    ...    Restore Systemd-Resolved
    ...    Systemd-Resolved Must Be Installed And Disabled


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host

Systemd-Resolved Must Be Installed And Disabled
    [Documentation]    Verify the systemd-resolved package is installed

    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    rpm -q systemd-resolved
    ...    sudo=False    return_rc=True    return_stdout=True    return_stderr=True
    Should Be Equal As Integers    0    ${rc}

    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    systemctl is-enabled -q systemd-resolved.service
    ...    sudo=False    return_rc=True    return_stdout=True    return_stderr=True
    Should Not Be Equal As Integers    0    ${rc}

    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    systemctl is-active -q systemd-resolved.service
    ...    sudo=False    return_rc=True    return_stdout=True    return_stderr=True
    Should Not Be Equal As Integers    0    ${rc}

Systemd-Resolved Must Not Be Installed
    [Documentation]    Verify the systemd-resolved package is installed

    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    rpm -q systemd-resolved
    ...    sudo=False    return_rc=True    return_stdout=True    return_stderr=True
    Should Not Be Equal As Integers    0    ${rc}

Delete Kubelet Configuration File
    [Documentation]    Delete the kubelet configuration file

    ${stderr}    ${rc}=    Execute Command
    ...    rm -f ${KUBELET_CONFIG_FILE}
    ...    sudo=True    return_rc=True    return_stderr=True    return_stdout=False
    Log    ${stderr}
    Should Be Equal As Integers    0    ${rc}

Start Systemd-Resolved
    [Documentation]    Start the systemd-resolved service

    Systemd-Resolved Must Be Installed And Disabled
    Systemctl    start    systemd-resolved

Stop Systemd-Resolved
    [Documentation]    Stop the systemd-resolved service

    Systemctl    stop    systemd-resolved
    Reboot MicroShift Host
    Wait Until Greenboot Health Check Exited

Uninstall Systemd-Resolved
    [Documentation]    Remove the systemd-resolved package

    ${is_ostree}=    Is System OSTree
    IF    ${is_ostree} == ${TRUE}
        ${stdout}    ${rc}=    Execute Command
        ...    rpm-ostree usroverlay
        ...    sudo=True    return_rc=True
        Should Be Equal As Integers    0    ${rc}
    END

    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    rpm -ev systemd-resolved
    ...    sudo=True    return_rc=True    return_stdout=True    return_stderr=True
    Should Be Equal As Integers    0    ${rc}

Restore Systemd-Resolved
    [Documentation]    Install the systemd-resolved package

    ${is_ostree}=    Is System OSTree
    IF    ${is_ostree} == ${TRUE}
        Reboot MicroShift Host
        Wait Until Greenboot Health Check Exited
    ELSE
        ${stdout}    ${stderr}    ${rc}=    Execute Command
        ...    dnf install -y systemd-resolved
        ...    sudo=True    return_rc=True    return_stdout=True    return_stderr=True
        Should Be Equal As Integers    0    ${rc}
    END

Check ResolveConf Option Presence
    [Documentation]    Check if the 'resolvConf' option is present in the kubelet
    ...    configuration file. Return a none-zero code if not present.

    ${rc}=    Execute Command
    ...    grep -qE "^resolvConf:.*${RESOLVE_CONF_FILE}" ${KUBELET_CONFIG_FILE}
    ...    sudo=True    return_rc=True    return_stdout=False    return_stderr=False
    RETURN    ${rc}
