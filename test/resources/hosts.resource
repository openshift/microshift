*** Settings ***
Documentation       Keywords for interacting with the hosts file

Library             Process
Library             String
Resource            common.resource


*** Keywords ***
Generate Random HostName
    [Documentation]    Generate Random Hostname
    ${rand}=    Generate Random String
    ${rand}=    Convert To Lower Case    ${rand}
    RETURN    ${rand}.api.com

Add Entry To Hosts
    [Documentation]    Add new entry to local /etc/hosts
    [Arguments]    ${ip}    ${host}    ${hostsfile}=/etc/hosts
    VAR    ${ttt}=    ${ip}\t${host} # RF test marker\n
    ${stdout}    ${stderr}    ${rc}=    SSHLibrary.Execute Command
    ...    echo "${ttt}" | sudo tee -a ${hostsfile}
    ...    sudo=True    return_rc=True    return_stderr=True    return_stdout=True
    Should Be Equal As Integers    ${rc}    0

Remove Entry From Hosts
    [Documentation]    Removes entry from local /etc/hosts
    [Arguments]    ${host}    ${hostsfile}=/etc/hosts
    ${stdout}    ${stderr}    ${rc}=    SSHLibrary.Execute Command
    ...    sed -i "/${host} # RF test marker/d" ${hostsfile}
    ...    sudo=True    return_rc=True    return_stderr=True    return_stdout=True
    Should Be Equal As Integers    ${rc}    0

Add Entry To Local Hosts
    [Documentation]    Add new entry to local /etc/hosts
    [Arguments]    ${ip}    ${host}
    VAR    ${ttt}=    ${ip}\t${host} # RF test marker\n
    ${result}=    Run Process    sudo tee -a /etc/hosts    shell=True    stdin=${ttt}
    Should Be Equal As Integers    ${result.rc}    0

Remove Entry From Local Hosts
    [Documentation]    Removes entry from local /etc/hosts
    [Arguments]    ${host}
    ${result}=    Run Process    sudo sed -i "/${host} # RF test marker/d" /etc/hosts    shell=True
    Should Be Equal As Integers    ${result.rc}    0
