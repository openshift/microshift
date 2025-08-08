*** Settings ***
Documentation       Generic Device Plugin

Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/ostree-health.resource
Variables           strings.py
Library             strings.py

Suite Setup         Setup Suite With Namespace
Suite Teardown      Teardown Suite With GDP Cleanup

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

Verify GDP handles hot plugging of Devices
    [Documentation]    Verify that GDP reacts to devices being added or removed after MicroShift has started
    [Tags]    hot-plug
    Enable And Configure GDP
    Wait Until Device Is Allocatable    0

    Enable Serialsim
    Wait Until Device Is Allocatable    1


*** Keywords ***
GDP Test Setup
    [Documentation]    Common setup for GDP tests - configures GDP, enables serialsim, prepares environment
    [Arguments]    ${config_content}
    Enable And Configure GDP    ${config_content}
    Enable Serialsim
    Copy Script To Host
    Wait Until Device Is Allocatable
    Command Should Work    crictl pull registry.access.redhat.com/ubi9/ubi:9.6
    Start Script On Host

GDP Test Teardown
    [Documentation]    Common teardown for GDP tests - stops script, cleans up resources, disables GDP
    Stop Script On Host
    Cleanup Test Resources
    Disable GDP

Enable And Configure GDP
    [Documentation]    Enables GDP and adds fake device path in MicroShift configuration
    [Arguments]    ${config_content}=${GDP_CONFIG_DROPIN}    ${dropin_name}=10-gdp
    Drop In MicroShift Config    ${config_content}    ${dropin_name}
    Restart MicroShift

Disable GDP
    [Documentation]    Removes GDP configuration drop-in (without restart)
    [Arguments]    ${dropin_name}=10-gdp
    Remove Drop In MicroShift Config    ${dropin_name}

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

Cleanup Test Resources
    [Documentation]    Cleans up test resources including configmap and job
    Run Keyword And Ignore Error
    ...    Oc Delete    job/gdp-test -n ${NAMESPACE}
    Run Keyword And Ignore Error
    ...    Oc Delete    configmap/gdp-script -n ${NAMESPACE}

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
    [Arguments]    ${expected_count}=1
    ${node}=    Run With Kubeconfig    oc get node -o=name
    ${node_name}=    Remove String    ${node}    node/
    Wait Until Keyword Succeeds    60s    5s
    ...    Device Should Be Allocatable    ${node_name}    ${expected_count}

Device Should Be Allocatable
    [Documentation]    Checks if device device.microshift.io/fakeserial is allocatable
    [Arguments]    ${node_name}    ${expected_count}=1
    ${device_amount}=    Oc Get JsonPath
    ...    node
    ...    ${EMPTY}
    ...    ${node_name}
    ...    .status.allocatable.device\\.microshift\\.io/fakeserial
    Should Be Equal As Integers    ${device_amount}    ${expected_count}

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

Teardown Suite With GDP Cleanup
    [Documentation]    Suite teardown that cleans up GDP configuration and restarts MicroShift
    # Clean up any remaining GDP configuration
    Run Keyword And Ignore Error
    ...    Remove Drop In MicroShift Config    10-gdp
    # Restart MicroShift to clean state for next suite
    Restart MicroShift
    Restart Greenboot And Wait For Success
    # Call original suite teardown
    Teardown Suite With Namespace
