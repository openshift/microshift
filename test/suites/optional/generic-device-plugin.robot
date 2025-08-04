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
    [Setup]    GDP Test Setup    ${GDP_CONFIG_DROPIN}

    Create Test Job
    Wait For Job Completion And Check Logs

    [Teardown]    GDP Test Teardown

Verify that mountPath correctly renames the device within the container
    [Documentation]    Performs a test of Generic Device Plugin with custom mountPath configuration
    [Setup]    GDP Test Setup    ${GDP_CONFIG_DROPIN_WITH_MOUNT}

    Create Test Job With Modified Script
    Wait For Job Completion And Check Logs

    [Teardown]    GDP Test Teardown


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
    Oc Delete    job/gdp-test -n ${NAMESPACE}
    Oc Delete    configmap/gdp-script -n ${NAMESPACE}

Create Test Job
    [Documentation]    Creates Job that spawns test Pod running to completion.
    ${script}=    OperatingSystem.Get File    ./assets/generic-device-plugin/fake-serial-communication.py
    ${configmap}=    Append To Preamble    ${script}
    Log    ${configmap}
    ${path}=    Create Random Temp File    ${configmap}
    Oc Create    -f ${path} -n ${NAMESPACE}
    Oc Create    -f ./assets/generic-device-plugin/job.yaml -n ${NAMESPACE}

Create Test Job With Modified Script
    [Documentation]    Creates Job that spawns test Pod running to completion with modified script.
    ${script}=    OperatingSystem.Get File    ./assets/generic-device-plugin/fake-serial-communication.py
    ${modified_script}=    Replace String
    ...    ${script}
    ...    DEVICE_POD = "/dev/ttyPipeB0"
    ...    DEVICE_POD = "/dev/myrenamedserial"
    ${configmap}=    Append To Preamble    ${modified_script}
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

Teardown Suite With GDP Cleanup
    [Documentation]    Suite teardown that cleans up GDP configuration and restarts MicroShift
    # Clean up any remaining GDP configuration
    Remove Drop In MicroShift Config    10-gdp
    # Restart MicroShift to clean state for next suite
    Restart MicroShift
    Restart Greenboot And Wait For Success
    # Call original suite teardown
    Teardown Suite With Namespace
