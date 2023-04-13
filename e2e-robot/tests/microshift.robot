*** Settings ***
Documentation       MicroShift e2e test suite

Library             SSHLibrary
Library             String
Library             OperatingSystem
Library             Process
Library             RequestsLibrary

Suite Setup         Get Kubeconfig
Suite Teardown      Remove Kubeconfig


*** Variables ***
${USHIFT_IP}        ${EMPTY}
${USHIFT_USER}      ${EMPTY}


*** Test Cases ***
Router Smoke Test
    [Documentation]    Verify that Router correctly exposes HTTP service
    [Tags]    smoke
    [Setup]    Run Keywords
    ...    Create Hello MicroShift Pod    AND
    ...    Expose Hello MicroShift Pod Via Router    AND
    ...    Open Port    80    tcp

    Wait Until Keyword Succeeds    3x    3s    Access Hello Microshift via Router

    [Teardown]    Run Keywords
    ...    Delete Hello MicroShift Pod Route And Service    AND
    ...    Close Port    80    tcp

Load Balancer Smoke Test
    [Documentation]    Verify that Load Balancer correctly exposes HTTP service
    [Tags]    smoke
    [Setup]    Run Keywords
    ...    Create Hello MicroShift Pod    AND
    ...    Expose Hello MicroShift Pod Via LB    AND
    ...    Open Port    5678    tcp

    Wait Until Keyword Succeeds    3x    3s    Access Hello Microshift via LB

    [Teardown]    Run Keywords
    ...    Delete Hello MicroShift Pod Route And Service    AND
    ...    Close Port    5678    tcp

Reboot Test
    [Documentation]    Verify that MicroShift starts successfully after reboot
    [Setup]    Run Keywords    Create Pod With PVC

    Open Connection    ${USHIFT_IP}
    Login    ${USHIFT_USER}    allow_agent=True
    Execute Command    reboot now    sudo=True
    Close Connection

    Sleep    15s

    Open Connection    ${USHIFT_IP}
    Set Client Configuration    timeout=600s
    Wait Until Keyword Succeeds    10x    10s    Login    ${USHIFT_USER}    allow_agent=True

    Wait Until Keyword Succeeds
    ...    10x
    ...    10s
    ...    Execute Command
    ...    [ $(systemctl show -p SubState --value microshift) = "running" ]
    ...    timeout=10s    return_stdout=True    return_stderr=True

    ${stdout}    ${rc}=    Execute Command
    ...    /etc/greenboot/check/required.d/40_microshift_running_check.sh | tee /tmp/asd.log
    ...    sudo=True
    ...    timeout=600s
    ...    return_stdout=True
    ...    return_rc=True
    Log    ${stdout}
    Should Be Equal As Integers    ${rc}    0
    Close Connection

    Run With Kubeconfig    oc wait --for=condition=Ready --timeout=120s pod/test-pod

    [Teardown]    Delete Pod With PVC

Failed Test
    Fail    Let's see how it looks in Prow


*** Keywords ***
Get Kubeconfig
    [Documentation]    X
    Open Connection    ${USHIFT_IP}
    Login    ${USHIFT_USER}    allow_agent=True
    ${konfig}=    Execute Command
    ...    cat /var/lib/microshift/resources/kubeadmin/${USHIFT_IP}/kubeconfig
    ...    sudo=True
    Should Not Be Empty    ${konfig}
    ${rand}=    Generate Random String
    ${path}=    Join Path    /tmp    ${rand}
    Create File    ${path}    ${konfig}
    Close Connection
    Set Suite Variable    \${KUBECONFIG}    ${path}

Remove Kubeconfig
    Remove File    ${KUBECONFIG}

Access Hello Microshift via Router
    ${result}=    Run Process
    ...    curl -i http://hello-microshift.cluster.local --resolve "hello-microshift.cluster.local:80:${USHIFT_IP}"
    ...    shell=True    timeout=15s
    Check HTTP Response    ${result}

Access Hello Microshift via LB
    ${result}=    Run Process    curl -i ${USHIFT_IP}:5678    shell=True    timeout=15s
    Check HTTP Response    ${result}

Check HTTP Response
    [Arguments]    ${result}
    Log    ${result.stdout}
    Log    ${result.stderr}
    Should Be Equal As Integers    ${result.rc}    0
    Should Match Regexp    ${result.stdout}    HTTP.*200
    Should Match    ${result.stdout}    *Hello MicroShift*

Create Pod With PVC
    Run With Kubeconfig    oc create -f ../e2e/tests/assets/pod-with-pvc.yaml
    Run With Kubeconfig    oc wait --for=condition=Ready --timeout=120s pod/test-pod

Delete Pod With PVC
    Run With Kubeconfig    oc delete -f ../e2e/tests/assets/pod-with-pvc.yaml    True

Create Hello MicroShift Pod
    Run With Kubeconfig    oc create -f ../e2e/tests/assets/hello-microshift.yaml
    Run With Kubeconfig    oc wait pods -l app\=hello-microshift --for condition\=Ready --timeout\=60s

Expose Hello MicroShift Pod Via Router
    Run With Kubeconfig    oc expose pod hello-microshift
    Run With Kubeconfig    oc expose svc hello-microshift --hostname hello-microshift.cluster.local

Expose Hello MicroShift Pod Via LB
    Run With Kubeconfig    oc create service loadbalancer hello-microshift --tcp=5678:8080

Delete Hello MicroShift Pod Route And Service
    Run With Kubeconfig    oc delete route hello-microshift    True
    Run With Kubeconfig    oc delete service hello-microshift    True
    Run With Kubeconfig    oc delete -f ../e2e/tests/assets/hello-microshift.yaml    True

Run With Kubeconfig
    [Arguments]    ${cmd}    ${allow_fail}=False
    ${result}=    Run Process    ${cmd}    env:KUBECONFIG=${KUBECONFIG}    stderr=STDOUT    shell=True
    Log    ${result.stdout}
    IF    ${allow_fail} == False
        Should Be Equal As Integers    ${result.rc}    0
    END

Open Port
    [Arguments]    ${number}    ${protocol}
    ${res}=    Run Process    bash    -c
    ...    if declare -F firewall::open_port; then firewall::open_port ${number} ${protocol}; fi    stderr=STDOUT
    Log    ${res.stdout}
    Should Be Equal As Integers    ${res.rc}    0

Close Port
    [Arguments]    ${number}    ${protocol}
    ${res}=    Run Process    bash    -c
    ...    if declare -F firewall::close_port; then firewall::close_port ${number} ${protocol}; fi    stderr=STDOUT
    Log    ${res.stdout}
    Should Be Equal As Integers    ${res.rc}    0
