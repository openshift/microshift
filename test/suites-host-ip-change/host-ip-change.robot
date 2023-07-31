*** Settings ***
Documentation       Host IP and Certificate change test Suite

Resource            ../resources/common.resource
Resource            ../resources/systemd.resource
Resource            ../resources/microshift-config.resource
Resource            ../resources/microshift-process.resource

Library             OperatingSystem
Library             BuiltIn
Library             String
Library             SSHLibrary
Library             random

Suite Setup         Setup Suite
Suite Teardown      Teardown Suite

Test Tags           restart    slow


*** Variables ***
${HOSTS_FILE_PATH}  /etc/hosts
${EXTERNAL_SERVER_CERT}    /var/lib/microshift/certs/kube-apiserver-external-signer/kube-external-serving/server.crt
${DEPLOY_POD_YAML}    assets/hello-microshift.yaml
${NAMESPACE}    test-ocp64902
${SCC_UID}    '{"metadata": {"annotations": {"openshift.io/sa.scc.uid-range": "1000-2000"}}}'
${SCC_MCS}    '{"metadata": {"annotations": {"openshift.io/sa.scc.mcs": "s0:c11,c5"}}}'


*** Test Cases ***
Verify Certificate After Host IP Change
    [Documentation]  Ensure certificate updates when the host IP changes
    # Generate new IP in same subnet
    Generate New IP

    # Take backup of /etc/hosts and /etc/NetworkManager/system-connections/8.nmconnection
    ${hosts_backup}=  Save Backup Default File    ${HOSTS_FILE_PATH}

    # Take initial checksum of the external server cert ffile
    # Save the initial checksum of the file
    ${initial_checksum}=    Get File Checksum    ${EXTERNAL_SERVER_CERT}

    # Add node name and IP to /etc/hosts
    ${node_name}=    Get NodeName
    Add Node To Hosts File    ${NEW_HOST_IP}    ${node_name}

    # Stop MicroShift
    Stop MicroShift

    # Modify the NetworkManager configuration using nmcli commands
    Modify NetworkManager Configuration   ${NEW_HOST_IP}/24   ${GATEWAY}   ${DEVICE}    manual

    # Reboot the MicroShift system vm
    Reboot MicroShift System    ${NEW_HOST_IP}

    # Login to the MicroShift host with IP via ssh and leave the connection open
    Login MicroShift Host With New IP   ${NEW_HOST_IP}

    # Verify changed IP in Microshift config
    Verify Microshift Config With New IP   ${NEW_HOST_IP}

    # Verify new Kubeconfig generated for new IP
    Check New Kubeconfig Generated For New IP

    # Compare the checksum of the external server cert file after the operations
    File Checksum Should Not Match    ${EXTERNAL_SERVER_CERT}    ${initial_checksum}

    # Rollback configuration files after the test case completes
    Restore Backup File    ${HOSTS_FILE_PATH}
    Modify NetworkManager Configuration   ${OLD_HOST_IP}/24   ${GATEWAY}   ${DEVICE}    auto

    # Reboot the MicroShift system vm after rollback
    Reboot MicroShift System    ${USHIFT_HOST}
    Wait For MicroShift

    # Verify microshift cluster work fine after rollback
    Deploy Pod In Namespace    ${NAMESPACE}
    Delete Namespace     ${NAMESPACE}


*** Keywords ***
Generate New IP
    [Documentation]    Generate new IP
    ${hostip}     ${rc1}=    Execute Command    microshift show-config | grep -i "nodeIP:" | awk '{print $2}'
    ...    sudo=True    return_rc=True
    Should Be Equal As Integers    ${rc1}    0
    ${gateway}    ${rc2}=    Execute Command    ip route show | grep -i ${hostip} | grep -i default | awk '{print $3}'
    ...    return_stdout=True    return_rc=True
    Should Be Equal As Integers    ${rc2}    0
    ${device}     ${rc3}=    Execute Command    ip route show | grep -i ${hostip} | grep -i default | awk '{print $5}'
    ...    return_stdout=True    return_rc=True
    Should Be Equal As Integers    ${rc3}    0

    FOR    ${last_digit}    IN RANGE    2    255
        ${digit}=    Evaluate    int($hostip.split('.')[-1])
        ${new_host_ip}=    Evaluate    '${hostip}'.rsplit('.', 1)[0] + '.' + str(${last_digit})
        ${ping_output}     ${rc4}=    Execute Command    ping -c 1 ${new_host_ip}
        ...    return_stdout=True    return_rc=True
        IF     ${rc4} > 0    BREAK
    END

    # Add one more check to ensure the generated IP is not the same as the current IP
    Should Not Be Equal    ${new_host_ip}    ${hostip}

    # Set the global variables
    Set Global Variable    ${NEW_HOST_IP}    ${new_host_ip}
    Set Global Variable    ${GATEWAY}       ${gateway}
    Set Global Variable    ${DEVICE}        ${device}
    Set Global Variable    ${OLD_HOST_IP}    ${hostip}

Save Backup Default File
    [Arguments]    ${file_path}
    [Documentation]    Takes a backup of the given file and returns the path to the backup file.
    ...    This keyword is meant to be used from a Setup step.
    ${stdout}    ${rc}=    Execute Command
    ...    cp ${file_path} ${file_path}.backup
    ...    sudo=True    return_rc=True
    Should Be Equal As Integers    ${rc}    0

Restore Backup File
    [Arguments]    ${backup_file}
    [Documentation]    Restores the original file from the backup file.
    ${stdout}    ${rc}=    Execute Command
    ...    cp ${backup_file}.backup ${backup_file}
    ...    sudo=True    return_rc=True
    Should Be Equal As Integers    ${rc}    0

Setup Suite
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig    # for readiness checks

Add Node To Hosts File
    [Arguments]    ${ip}    ${node_name}
    [Documentation]    Add the node name and IP address to the /etc/hosts file
    ${stdout}    ${rc}=    Execute Command
    ...    /bin/bash -c 'echo "${ip} ${node_name}" | tee -a /etc/hosts'
    ...    sudo=True    return_rc=True
    Should Be Equal As Integers    ${rc}    0

Get NodeName
    [Documentation]    Get and Return nodename
    ${nodeName}=    Run With Kubeconfig    oc get node -o name | cut -d "/" -f2
    Should Not Be Empty    ${nodeName}
    RETURN    ${nodeName}

Modify NetworkManager Configuration
    [Arguments]    ${ip}    ${gateway}   ${device}    ${mode}
    [Documentation]    Modify NetworkManager configuration using nmcli commands
    ${stdout}   ${rc1}=    Execute Command    nmcli con modify ${device} ipv4.addresses ${ip}
    ...    sudo=True    return_rc=True
    ${stdout}   ${rc2}=    Execute Command    nmcli con mod ${device} ipv4.gateway ${gateway}
    ...    sudo=True    return_rc=True
    ${stdout}   ${rc3}=    Execute Command    nmcli con mod ${device} ipv4.method ${mode}
    ...    sudo=True    return_rc=True
    Should Be Equal As Integers    ${rc1}     0
    Should Be Equal As Integers    ${rc2}     0
    Should Be Equal As Integers    ${rc3}     0

Login MicroShift Host With New IP
  [Arguments]  ${ip}
  [Documentation]  Login to the MicroShift host via ssh and leave the connection open
  ...  This keyword is meant to be used at the suite level. This ensures
  ...  most tests already have an open connection. Any tests that will take
  ...  action that disrupt that connection are responsible for restoring it.

  IF    '${SSH_PORT}'
        SSHLibrary.Open Connection    ${ip}    port=${SSH_PORT}
  ELSE
        SSHLibrary.Open Connection    ${ip}
  END
  Wait And Retry Login MicroShift Host  ${ip}
  # If there is an ssh key set in the global configuration, use that to
  # login to the host. Otherwise, assume that the ssh agent is running
  # and configured properly.
  IF  '${SSH_PRIV_KEY}'
    SSHLibrary.Login With Public Key  ${USHIFT_USER}  ${SSH_PRIV_KEY}
  ELSE
    SSHLibrary.Login  ${USHIFT_USER}  allow_agent=True
  END

Wait And Retry Login MicroShift Host
  [Arguments]  ${ip}
  [Documentation]  Wait and retry logging in to the MicroShift host via SSH
  ...  with a specified maximum number of retries and interval between retries.
  Wait Until Keyword Succeeds  10m  15s  SSHLibrary.Login With Public Key  ${USHIFT_USER}  ${SSH_PRIV_KEY}

Reboot Microshift System
    [Arguments]    ${host_ip}
    [Documentation]    Reboot the system using "systemctl reboot" command

    SSHLibrary.Start Command    reboot    sudo=True
    Sleep    30s

    # Wait until the system is reachable after the reboot
    Wait Until Keyword Succeeds    10m    10s    System Should Be Reachable   ${host_ip}
    # Add a sleep here (if needed) to allow the system to fully reboot and become reachable again
    Sleep    30s

Verify Microshift Config With New IP
    [Arguments]  ${new_host_ip}
    [Documentation]   Verify that 'microshift show-config' command output contains the IP address
    ${stdout}    ${rc}=    Execute Command    microshift show-config
    ...    sudo=True    return_rc=True
    Should Be Equal As Integers    0    ${rc}
    ${config}=    Yaml Parse    ${stdout}
    Should Be Equal    ${new_host_ip}    ${config.node.nodeIP}

Check New Kubeconfig Generated For New IP
    [Documentation]    Check the new kubeconfig
    ${hostname}   ${rc1}=    Execute Command    hostname
    ...    sudo=True    return_rc=True
    ${output}    ${rc2}=     Execute Command    ls /var/lib/microshift/resources/kubeadmin/${hostname}/
    ...    sudo=True    return_rc=True
    Should Contain    ${output}    kubeconfig
    Should Be Equal As Integers    ${rc1}    0
    Should Be Equal As Integers    ${rc2}    0

Get File Checksum
    [Arguments]    ${file_path}
    [Documentation]    Generate File Checksum
    ${stdout}    ${rc}=    Execute Command     sha256sum ${file_path}
    ...    sudo=True    return_rc=True
    IF    '${rc}' != '0'    Fail    Failed to get the checksum of the file: ${file_path}
    RETURN    ${stdout.split()[0]}

File Checksum Should Not Match
    [Arguments]    ${file_path}    ${expected_checksum}
    [Documentation]    File Checksum Comparison
    ${current_checksum}=    Get File Checksum    ${file_path}
    Should Not Be Equal    ${current_checksum}    ${expected_checksum}

Deploy Pod In Namespace
    [Arguments]    ${namespace}
    [Documentation]    Deploy pod with namespace
    ${stdout}=    Run With Kubeconfig    oc create ns ${namespace}
    Should Contain    ${stdout}    created
    Run With Kubeconfig    oc patch namespace ${namespace} -p ${SCC_UID}
    Run With Kubeconfig    oc patch namespace ${namespace} -p ${SCC_MCS}
    Run With Kubeconfig    oc create -f ${DEPLOY_POD_YAML} -n ${namespace}
    Run With Kubeconfig    oc wait pods -l app\=hello-microshift --for condition\=Ready --timeout\=60s -n ${namespace}

Delete Namespace
    [Arguments]    ${namespace}
    [Documentation]    Delete  namespace
    Run With Kubeconfig    oc delete ns ${namespace} --ignore-not-found

System Should Be Reachable
    [Arguments]    ${host_ip}
    [Documentation]    Check System Reachable
    # Put the logic to check if the system is reachable here.
    Wait Until Keyword Succeeds    10m    10s
    ...    SSHLibrary.Open Connection    ${host_ip}
