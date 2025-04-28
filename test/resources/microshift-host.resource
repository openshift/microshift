*** Settings ***
Documentation       Keywords for working with the MicroShift host

Library             OperatingSystem
Library             SSHLibrary
Library             libostree.py


*** Keywords ***
Login MicroShift Host
    [Documentation]    Login to the MicroShift host via ssh and leave the connection open
    ...
    ...    This keyword is meant to be used at the suite level. This ensures
    ...    most tests already have an open connection. Any tests that will take
    ...    action that disrupt that connection are responsible for restoring it.

    Log    Host: ${USHIFT_HOST}
    IF    '${SSH_PORT}'
        SSHLibrary.Open Connection    ${USHIFT_HOST}    port=${SSH_PORT}
    ELSE
        SSHLibrary.Open Connection    ${USHIFT_HOST}
    END
    # If there is an ssh key set in the global configuration, use that to
    # login to the host. Otherwise assume that the ssh agent is running
    # and configured properly.
    IF    '${SSH_PRIV_KEY}'
        SSHLibrary.Login With Public Key    ${USHIFT_USER}    ${SSH_PRIV_KEY}    keep_alive_interval=30
    ELSE
        SSHLibrary.Login    ${USHIFT_USER}    allow_agent=True    keep_alive_interval=30
    END

Logout MicroShift Host
    [Documentation]    Close the open ssh connection to the MicroShift host
    SSHLibrary.Close Connection

Make New SSH Connection
    [Documentation]    Closes all SSH connections and makes a new one.
    # Staging deployments potentially introduces multiple reboots
    # which could break existing SSH connection

    Logout MicroShift Host
    Login MicroShift Host

SSH Connection To MicroShift Host Should Be Functional
    [Documentation]    Creates new SSH connection and checks if is working
    Make New SSH Connection
    ${rc}=    SSHLibrary.Execute Command    true    return_stdout=False    return_rc=True
    Should Be Equal As Integers    0    ${rc}

Reboot MicroShift Host
    [Documentation]    Reboot the MicroShift host and wait until
    ...    SSH connection is working again and boot identifier changes
    ...
    ...    Expects that initial SSH connection to MicroShift host is active.

    ${bootid}=    Get Current Boot Id
    SSHLibrary.Start Command    reboot    sudo=True
    Sleep    30s

    Wait Until Keyword Succeeds    5m    15s
    ...    System Should Be Rebooted    ${bootid}

Create Thin Storage Pool
    [Documentation]    Create a new thin storage pool
    ${lvmd_vg}=    Set Variable If    '${LVMD_VG_OVERRIDE}' != ''    ${LVMD_VG_OVERRIDE}    rhel
    ${stderr}    ${rc}=    SSHLibrary.Execute Command    lvcreate --size 2G --thin /dev/${lvmd_vg} --name thin
    ...    sudo=True    return_rc=True    return_stderr=True    return_stdout=False
    Log    ${stderr}
    Should Be Equal As Integers    0    ${rc}

Delete Thin Storage Pool
    [Documentation]    Delete the thin storage pool as part of restoring test env
    ${lvmd_vg}=    Set Variable If    '${LVMD_VG_OVERRIDE}' != ''    ${LVMD_VG_OVERRIDE}    rhel
    ${stderr}    ${rc}=    SSHLibrary.Execute Command    lvremove -y ${lvmd_vg}/thin
    ...    sudo=True    return_rc=True    return_stderr=True    return_stdout=False
    Log    ${stderr}
    Should Be Equal As Integers    0    ${rc}

Is System Bootc
    [Documentation]    Returns true if system is bootc-based.

    ${stdout}    ${stderr}    ${rc}=    SSHLibrary.Execute Command
    ...    bootc status --booted --json
    ...    sudo=True    return_stderr=True    return_stdout=True    return_rc=True
    IF    ${rc} != 0    RETURN    ${FALSE}

    ${json_status}=    Json Parse    ${stdout}
    ${bootc_type}=    Set Variable    ${json_status}[status][type]
    IF    '${bootc_type}' == 'bootcHost'
        RETURN    ${TRUE}
    ELSE
        RETURN    ${FALSE}
    END

Is System OSTree
    [Documentation]    Returns true if system is ostree-based.

    ${rc}=    Execute Command    stat /run/ostree-booted
    ...    sudo=True    return_stderr=False    return_stdout=False    return_rc=True
    IF    ${rc} == 0    RETURN    ${TRUE}    ELSE    RETURN    ${FALSE}

System Should Be Rebooted
    [Documentation]    Assert if the system rebooted comparing the current and provided boot identifier
    [Arguments]    ${old_bootid}
    ${rebooted}=    Is System Rebooted    ${old_bootid}
    Should Be True    ${rebooted}

Is System Rebooted
    [Documentation]    Check if the system rebooted comparing the current and provided boot identifier
    [Arguments]    ${old_bootid}

    Make New SSH Connection
    ${cur_bootid}=    Get Current Boot Id
    ${len}=    Get Length    ${cur_bootid}
    IF    ${len} == 0
        RETURN    False
    ELSE
        ${system_rebooted}=    Evaluate    '${old_bootid}' != '${cur_bootid}'
        RETURN    ${system_rebooted}
    END

Change Hostname
    [Documentation]    Changes hostname to the argument. Returns old hostname
    [Arguments]    ${hostname}

    ${old_hostname}    ${stderr}    ${rc}=    SSHLibrary.Execute Command    hostname
    ...    sudo=True    return_rc=True    return_stdout=True    return_stderr=True
    Log    ${stderr}
    Should Be Equal As Integers    0    ${rc}

    ${stderr}    ${rc}=    SSHLibrary.Execute Command    hostnamectl hostname ${hostname}
    ...    sudo=True    return_rc=True    return_stderr=True    return_stdout=False
    Log    ${stderr}
    Should Be Equal As Integers    0    ${rc}

    RETURN    ${old_hostname}

Command Should Work
    [Documentation]    Run a command remotely, log stderr, fail if RC is not 0, return stdout
    [Arguments]    ${command}    ${sudo_mode}=True
    ${stdout}    ${stderr}    ${rc}=    Command Execution    ${command}    ${sudo_mode}
    Should Be Equal As Integers    ${rc}    0
    RETURN    ${stdout}

Command Should Fail
    [Documentation]    Run a command remotely, log stderr, fail if RC is 0, return stdout
    [Arguments]    ${command}    ${sudo_mode}=True
    ${stdout}    ${stderr}    ${rc}=    Command Execution    ${command}    ${sudo_mode}
    Should Not Be Equal As Integers    ${rc}    0
    RETURN    ${stderr}

Command Execution
    [Documentation]    Run a command remotely, log stdout and stderr. Don't check RC.
    [Arguments]    ${command}    ${sudo_mode}=True
    ${stdout}    ${stderr}    ${rc}=    SSHLibrary.Execute Command
    ...    ${command}
    ...    sudo=${sudo_mode}
    ...    return_rc=True
    ...    return_stderr=True
    ...    return_stdout=True
    Log    ${stdout}
    Log    ${stderr}
    RETURN    ${stdout}    ${stderr}    ${rc}

Get System Architecture
    [Documentation]    Returns the system architecture
    ${arch}=    Command Should Work    uname -m
    RETURN    ${arch}

Get MicroShift Cluster ID From File
    [Documentation]    Read and return the cluster ID from the file.
    ${id}=    Command Should Work    cat /var/lib/microshift/cluster-id

    Should Not Be Empty    ${id}
    RETURN    ${id}

Get Deployment Type
    [Documentation]    Returns whether the target system is bootc, ostree or rpm
    ${is_bootc}=    Is System Bootc
    ${is_ostree}=    Is System Ostree
    ${cluster_id}=    Get MicroShift Cluster ID From File

    IF    ${is_bootc} == ${TRUE}
        RETURN    bootc
    ELSE IF    ${is_ostree} == ${TRUE}
        RETURN    ostree
    END
    RETURN    rpm

Get Host OS Id
    [Documentation]    Returns the host OS id
    ${os_id}=    Command Should Work    bash -c '. /etc/os-release && echo "\${ID}"'
    Should Not Be Empty    ${os_id}
    RETURN    ${os_id}

Get Host OS Version
    [Documentation]    Returns the host OS version
    ${os_version}=    Command Should Work    bash -c '. /etc/os-release && echo "\${VERSION_ID}"'
    RETURN    ${os_version}
