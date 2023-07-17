*** Settings ***
Documentation       Keywords for working with the MicroShift host

Library             SSHLibrary


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
        SSHLibrary.Login With Public Key    ${USHIFT_USER}    ${SSH_PRIV_KEY}
    ELSE
        SSHLibrary.Login    ${USHIFT_USER}    allow_agent=True
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
    ...    SSH connection is working again
    ...
    ...    Expects that intial SSH connection to MicroShift host is active.

    SSHLibrary.Start Command    reboot    sudo=True
    Sleep    30s
    Wait Until Keyword Succeeds    5m    15s
    ...    SSH Connection To MicroShift Host Should Be Functional

Create Thin Storage Pool
    [Documentation]         Create a new thin storage pool
    ${stderr}   ${rc}=      SSHLibrary.Execute Command    lvcreate --size 4G --thin /dev/rhel --name thin
    ...                     sudo=True   return_rc=True  return_stderr=True  return_stdout=False
    Log                     ${stderr}
    Should Be Equal As Integers     0   ${rc}


Delete Thin Storage Pool
    [Documentation]         Delete the thin storage pool as part of restoring test env
    ${stderr}   ${rc}=       SSHLibrary.Execute Command    lvremove -y rhel\thin
    ...                     sudo=True   return_rc=True  return_stderr=True  return_stdout=False
    Log                     ${stderr}
    Should Be Equal As Integers    0    ${rc}
