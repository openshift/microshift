*** Settings ***
Documentation       Keywords for working with the MicroShift on a non-ostree system

Resource            common.resource
Library             String
Library             SSHLibrary


*** Variables ***
${REPO_FILE_NAME}       "/etc/yum.repos.d/microshift-local.repo"


*** Keywords ***
Install MicroShift RPM Packages
    [Documentation]    Installs MicroShift RPM packages from the specified URL
    [Arguments]    ${repo_url}

    ${repo_file_contents}=    Catenate    SEPARATOR=\n
    ...    [microshift-local]
    ...    name=Microshift Local Repository
    ...    baseurl=${repo_url}
    ...    enabled=1
    ...    gpgcheck=0

    Upload String To File    ${repo_file_contents}    ${REPO_FILE_NAME}
    Update Dnf Cache

    ${stdout}    ${stderr}    ${rc}=    SSHLibrary.Execute Command
    ...    dnf install -y microshift
    ...    sudo=True    return_rc=True    return_stderr=True    return_stdout=True
    Log    ${stderr}
    Should Be Equal As Integers    0    ${rc}

    ${rc}=    SSHLibrary.Execute Command    rm -f "${REPO_FILE_NAME}"
    ...    sudo=True    return_rc=True    return_stderr=False    return_stdout=False
    Should Be Equal As Integers    0    ${rc}

    Update Dnf Cache

Uninstall MicroShift RPM Packages
    [Documentation]    Uninstalls MicroShift RPM packages

    ${stdout}    ${stderr}    ${rc}=    SSHLibrary.Execute Command
    ...    dnf remove -y --exclude=microshift-test-agent microshift\*
    ...    sudo=True    return_rc=True    return_stderr=True    return_stdout=True
    Log    ${stderr}
    Should Be Equal As Integers    0    ${rc}

Update Dnf Cache
    [Documentation]    Updates dnf cache

    ${stderr}    ${rc}=    SSHLibrary.Execute Command    dnf makecache
    ...    sudo=True    return_rc=True    return_stderr=True    return_stdout=False
    Log    ${stderr}
    Should Be Equal As Integers    0    ${rc}

Get Version Of MicroShift RPM
    [Documentation]    Returns the version of the installed MicroShift RPM as a string
    ${rpm_cmd_output}    ${stderr}    ${rc}=    SSHLibrary.Execute Command
    ...    rpm -q microshift
    ...    sudo=True    return_rc=True    return_stdout=True    return_stderr=True
    Log    ${stderr}
    Should Be Equal As Integers    0    ${rc}
    # rpm_cmd_output will look like:
    # microshift-4.15.0_0.nightly_2023_11_01_080931_20231103152813_3f5593fca_dirty-1.el9.x86_64
    ${ignored}    ${version_string_raw}=    Split String    ${rpm_cmd_output}    -    1
    # version_string_raw should be:
    # 4.15.0_0.nightly_2023_11_01_080931_20231103152813_3f5593fca_dirty-1.el9.x86_64
    ${version_string}=    Strip String    ${version_string_raw}
    RETURN    ${version_string}
