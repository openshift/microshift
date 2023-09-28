*** Settings ***
Documentation       verify microshift with systemd-resolved installed

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-host.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/ostree-health.resource

Suite Setup         Setup
Suite Teardown      Teardown


*** Test Cases ***
Systemd-resolved installed
    [Documentation]    Verify kubelet uses the upstream resolve file of systemd-resolved
    Install Systemd-resolved RPM Packages
    Restart MicroShift
    Wait Until Greenboot Health Check Exited
    ${output}    ${rc}=    Execute Command
    ...    grep -q "resolvConf: /run/systemd/resolve/resolv.conf" \
    ...    /var/lib/microshift/resources/kubelet/config/config.yaml
    ...    sudo=True    return_rc=True
    Should Be Equal As Integers    0    ${rc}
    Uninstall Systemd-resolved RPM Packages
    Restart MicroShift


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host

Install Systemd-resolved RPM Packages
    [Documentation]    Installs Systemd-resolved RPM packages from the specified URL

    ${stdout}    ${stderr}    ${rc}=    SSHLibrary.Execute Command
    ...    dnf install -y systemd-resolved
    ...    sudo=True    return_rc=True    return_stderr=True    return_stdout=True
    Log    ${stderr}
    Should Be Equal As Integers    0    ${rc}

Uninstall Systemd-resolved RPM Packages
    [Documentation]    Uninstalls Systemd-resolved RPM packages

    ${stdout}    ${stderr}    ${rc}=    SSHLibrary.Execute Command
    ...    dnf remove -y systemd-resolved
    ...    sudo=True    return_rc=True    return_stderr=True    return_stdout=True
    Log    ${stderr}
    Should Be Equal As Integers    0    ${rc}
