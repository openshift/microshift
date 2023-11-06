*** Settings ***
Documentation       Keywords for working with the MicroShift on a non-ostree system

Resource            common.resource
Library             SSHLibrary


*** Variables ***
${REPO_FILE_NAME}       "/etc/yum.repos.d/microshift-local.repo"


*** Keywords ***
Install MicroShift RPM Packages From System Repo
    [Documentation]    Installs MicroShift RPM packages from a system repository
    ${stdout}    ${stderr}    ${rc}=    SSHLibrary.Execute Command
    ...    dnf install -y microshift
    ...    sudo=True    return_rc=True    return_stderr=True    return_stdout=True
    Log    ${stderr}
    Should Be Equal As Integers    0    ${rc}

Install MicroShift RPM Packages From Repo
    [Documentation]    Installs MicroShift RPM packages from the specified URL
    [Arguments]    ${repo_url}

    Configure MicroShift Repository    ${repo_url}

    ${stdout}    ${stderr}    ${rc}=    SSHLibrary.Execute Command
    ...    dnf install -y microshift
    ...    sudo=True    return_rc=True    return_stderr=True    return_stdout=True
    Log    ${stderr}
    Should Be Equal As Integers    0    ${rc}

    # Look for all warnings and errors before testing so that the log
    # shows the output for both.
    ${warnings}=    Get Lines Containing String    ${stdout}    warning    ignore_case=True
    ${errors}=    Get Lines Containing String    ${stdout}    error    ignore_case=True
    Should Not Contain    ${stdout}    warning    ignore_case=True
    Should Not Contain    ${stdout}    error    ignore_case=True

    Unconfigure MicroShift Repository

Configure MicroShift Repository
    [Documentation]    Sets up repository
    [Arguments]    ${repo_url}
    ${repo_file_contents}=    Catenate    SEPARATOR=\n
    ...    [microshift-local]
    ...    name=Microshift Local Repository
    ...    baseurl=${repo_url}
    ...    enabled=1
    ...    gpgcheck=0

    Upload String To File    ${repo_file_contents}    ${REPO_FILE_NAME}
    Update Dnf Cache

Unconfigure MicroShift Repository
    [Documentation]    Remove repo
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
