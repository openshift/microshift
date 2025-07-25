*** Settings ***
Documentation       Generic Device Plugin

Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/ostree-health.resource
Variables           strings.py
Library             strings.py

Suite Setup         Setup Suite With Namespace
Suite Teardown      Teardown Suite With Namespace

Test Tags           generic-device-plugin


*** Variables ***
${NAMESPACE}    ${EMPTY}


*** Test Cases ***
Sanity Test
    [Documentation]    Performs a simple test of Generic Device Plugin
    [Setup]    Run Keywords
    ...    Enable And Configure GDP
    ...    Enable Serialsim
    ...    Copy Script To Host

    Wait Until Device Is Allocatable

    Command Should Work    crictl pull registry.access.redhat.com/ubi9/ubi:9.6
    Start Script On Host
    Create Test Job

    Wait For Job Completion And Check Logs

    [Teardown]    Run Keywords
    ...    Stop Script On Host
    ...    Disable GDP


*** Keywords ***
Enable And Configure GDP
    [Documentation]    Enables GDP and adds fake device path in MicroShift configuration
    Drop In MicroShift Config    ${GDP_CONFIG_DROPIN}    10-gdp
    Restart MicroShift

Disable GDP
    [Documentation]    Removes GDP configuration drop-in
    Remove Drop In MicroShift Config    10-gdp
    Restart MicroShift
    Restart Greenboot And Wait For Success

Enable Serialsim
    [Documentation]    Enables the serialsim kernel module.
    ...    serialsim creates echo and pipe devices.
    Command Should Work    modprobe serialsim

Copy Script To Host
    [Documentation]    Starts fake serial communication script to the host
    Put File
    ...    ./assets/generic-device-plugin/fake-serial-communication.py
    ...    /tmp/fake-serial-communication.py

Start Script On Host
    [Documentation]    Starts fake serial communication script on the host in the background
    ${cmd}=    Catenate
    ...    systemd-run -u gdp-test-comm
    ...    python /tmp/fake-serial-communication.py host
    Command Should Work    ${cmd}

Stop Script On Host
    [Documentation]    Attempts to stop the fake serial communication script on the host.
    ...    If it was successful, the unit is deleted automatically.
    Command Execution    journalctl -u gdp-test-comm.service
    Command Execution    systemctl stop gdp-test-comm
    Command Execution    systemctl reset-failed gdp-test-comm

Create Test Job
    [Documentation]    Creates Job that spawns test Pod running to completion.
    ${script}=    OperatingSystem.Get File    ./assets/generic-device-plugin/fake-serial-communication.py
    ${configmap}=    Append To Preamble    ${script}
    Log    ${configmap}
    ${path}=    Create Random Temp File    ${configmap}
    Oc Create    -f ${path} -n ${NAMESPACE}
    Oc Create    -f ./assets/generic-device-plugin/job.yaml -n ${NAMESPACE}

Wait Until Device Is Allocatable
    [Documentation]    Waits until device device.microshift.io/fakeserial is allocatable
    ${node}=    Run With Kubeconfig    oc get node -o=name
    ${node_name}=    Remove String    ${node}    node/

    Wait Until Keyword Succeeds    60s    5s
    ...    Device Should Be Allocatable    ${node_name}

Device Should Be Allocatable
    [Documentation]    Checks if device device.microshift.io/fakeserial is allocatable
    [Arguments]    ${node_name}
    ${device_amount}=    Oc Get JsonPath
    ...    node
    ...    ${EMPTY}
    ...    ${node_name}
    ...    .status.allocatable.device\\.microshift\\.io/fakeserial
    Should Be Equal As Integers    ${device_amount}    1

Wait For Job Completion And Check Logs
    [Documentation]    Waits for Job completion and checks Pod logs looking for 'Test successful' message

    Oc Wait    -n ${NAMESPACE} job/gdp-test    --for=condition=complete --timeout=120s
    ${pod}=    Oc Get JsonPath
    ...    pod
    ...    ${NAMESPACE}
    ...    --selector=batch.kubernetes.io/job-name=gdp-test
    ...    .items[*].metadata.name
    ${logs}=    Oc Logs    ${pod}    ${NAMESPACE}
    Should Contain    ${logs}    Test successful
