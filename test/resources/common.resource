*** Settings ***
Documentation       Keywords common to many test suites

Library             OperatingSystem
Library             String
Library             SSHLibrary
Resource            ../resources/kubeconfig.resource


*** Keywords ***
Setup Suite
    [Documentation]    Setup a namespace-less test suite by configuring Kubeconfig
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig

Setup Suite With Namespace
    [Documentation]    Setup a test suite by creating a unique namespace and configuring Kubeconfig
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig
    ${ns}=    Create Unique Namespace
    Set Suite Variable    \${NAMESPACE}    ${ns}

Create Unique Namespace
    [Documentation]    Creates a namespace with a unique suffix
    ${rand}=    Generate Random String
    ${rand}=    Convert To Lower Case    ${rand}
    ${ns}=    Catenate    SEPARATOR=-    test    ${rand}

    Create Namespace    ${ns}
    RETURN    ${ns}

Teardown Suite
    [Documentation]    Teardown the namespace-less test suite by removing the Kubeconfig.
    Remove Kubeconfig
    Login MicroShift Host

Teardown Suite With Namespace
    [Documentation]    Teardown the namespaced test suite by removing the namespace and removing the Kubeconfig.
    Remove Namespace    ${NAMESPACE}
    Remove Kubeconfig

Check Required Env Variables
    [Documentation]    Fail if any of the required environment variables is missing.
    IF    "${USHIFT_HOST}"=="${EMPTY}"
        Fatal Error    USHIFT_HOST variable is required
    END
    IF    "${USHIFT_USER}"=="${EMPTY}"
        Fatal Error    USHIFT_USER variable is required
    END

Create Random Temp File
    [Documentation]    Create a random temp file
    [Arguments]    ${content}=""
    ${rand}=    Generate Random String
    ${path}=    Join Path    /tmp    ${rand}
    Create File    ${path}    ${content}
    RETURN    ${path}

Upload String To File    # robocop: disable=too-many-calls-in-keyword
    [Documentation]    Write the string to a remote file
    [Arguments]    ${content}    ${remote_filename}
    ${rand}=    Generate Random String
    ${local_tmp}=    Join Path    /tmp    ${rand}
    Create File    ${local_tmp}    ${content}
    ${rand}=    Generate Random String
    ${remote_tmp}=    Join Path    /tmp    ${rand}
    Put File    ${local_tmp}    ${remote_tmp}    mode=0644
    Remove File    ${local_tmp}
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    mv -Z ${remote_tmp} ${remote_filename}
    ...    sudo=True    return_rc=True    return_stdout=True    return_stderr=True
    Should Be Equal As Integers    0    ${rc}
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    chown root:root ${remote_filename}
    ...    sudo=True    return_rc=True    return_stdout=True    return_stderr=True
    Should Be Equal As Integers    0    ${rc}

Verify Remote Directory Exists With Sudo
    [Documentation]    Use sudo to verify that the specified directory exists
    [Arguments]    ${remote_dir}

    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    test -d ${remote_dir}
    ...    sudo=True    return_rc=True    return_stdout=True    return_stderr=True
    Should Be Equal As Integers    0    ${rc}

Verify Remote Directory Does Not Exist With Sudo
    [Documentation]    Use sudo to verify that the specified directory does not exist
    [Arguments]    ${remote_dir}

    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    test -d ${remote_dir}
    ...    sudo=True    return_rc=True    return_stdout=True    return_stderr=True
    Should Not Be Equal As Integers    0    ${rc}

Verify Remote File Exists With Sudo
    [Documentation]    Use sudo to verify that the specified file exists
    [Arguments]    ${remote_file}

    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    test -f ${remote_file}
    ...    sudo=True    return_rc=True    return_stdout=True    return_stderr=True
    Should Be Equal As Integers    0    ${rc}

Verify Remote File Does Not Exist With Sudo
    [Documentation]    Use sudo to verify that the specified file does not exist
    [Arguments]    ${remote_file}

    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    test -f ${remote_file}
    ...    sudo=True    return_rc=True    return_stdout=True    return_stderr=True
    Should Not Be Equal As Integers    0    ${rc}

Create Random Temp Directory
    [Documentation]    Create a random temp file
    [Arguments]    ${root_path}=/tmp/certs
    ${rand}=    Generate Random String
    ${path}=    Join Path    ${root_path}    ${rand}
    Create Directory    ${path}
    RETURN    ${path}

Create Remote Dir For Path
    [Documentation]    Create parent dirname if doesnt exist
    [Arguments]    ${file_path}
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    mkdir -pZ $(dirname ${file_path})
    ...    sudo=True    return_rc=True    return_stdout=True    return_stderr=True
    Should Be Equal As Integers    0    ${rc}

Local Command Should Work
    [Documentation]    Run a command localy, log stdout, fail if RC is not 0, return stdout.
    ...    stderr is redirected to the stdout
    [Arguments]    ${command}
    ${rc}    ${stdout}=    Run And Return RC And Output
    ...    ${command} 2>&1
    Log    ${stdout}
    Should Be Equal As Integers    0    ${rc}
    RETURN    ${stdout}
