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
${NAMESPACE}        ${EMPTY}
${DEVICE_COUNT}     5
${WAIT_TIMEOUT}     5m


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

Verify ttyUSB glob pattern device discovery and allocation
    [Documentation]    Tests GDP with glob pattern to discover multiple ttyUSB devices and verify device allocation across multiple pods
    [Tags]    glob-pattern
    [Setup]    TtyUSB Glob Test Setup

    # Create and verify pods with device allocation
    Create Pod And Verify Allocation    serial-test-pod    2    2
    Create Pod And Verify Allocation    serial-test-pod1    1    3
    Create Pod And Verify Allocation    serial-test-pod2    2    5

    [Teardown]    TtyUSB Glob Test Teardown

Verify FUSE device allocation and accessibility
    [Documentation]    Verifies FUSE device configuration, allocation, and accessibility in pods
    [Tags]    fuse-device
    [Setup]    Enable And Configure GDP    ${GDP_CONFIG_FUSE_COUNT}
    Wait Until Device Is Allocatable    10    fuse
    Oc Create    -f ./assets/generic-device-plugin/fuse-test-pod.yaml -n ${NAMESPACE}

    Oc Wait    -n ${NAMESPACE} pod/fuse-test-pod    --for=condition=Ready --timeout=${WAIT_TIMEOUT}

    # Verify /dev/fuse is accessible in the pod
    ${fuse_device}=    Oc Exec    fuse-test-pod    ls -l /dev/fuse
    Should Contain    ${fuse_device}    /dev/fuse

    # Verify node allocation shows 4 FUSE devices allocated
    ${node}=    Run With Kubeconfig    oc get node -o=name
    ${node_name}=    Remove String    ${node}    node/
    ${describe_output}=    Run With Kubeconfig    oc describe node ${node_name}
    ${allocated_line}=    Get Lines Containing String    ${describe_output}    device.microshift.io/fuse
    ${allocation_matches}=    Get Regexp Matches
    ...    ${allocated_line}
    ...    device\\.microshift\\.io/fuse\\s+(\\d+)\\s+(\\d+)
    ...    1    2
    Should Be Equal As Integers    ${allocation_matches}[0][0]    4
    Should Be Equal As Integers    ${allocation_matches}[0][1]    4


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
    [Arguments]    ${expected_count}=1    ${device_type}=fakeserial
    ${node}=    Run With Kubeconfig    oc get node -o=name
    ${node_name}=    Remove String    ${node}    node/
    Wait Until Keyword Succeeds    60s    5s
    ...    Device Should Be Allocatable    ${node_name}    ${expected_count}    ${device_type}

Device Should Be Allocatable
    [Documentation]    Checks if specified device is allocatable
    [Arguments]    ${node_name}    ${expected_count}=1    ${device_type}=fakeserial
    ${device_amount}=    Oc Get JsonPath
    ...    node
    ...    ${EMPTY}
    ...    ${node_name}
    ...    .status.allocatable.device\\.microshift\\.io/${device_type}
    Should Be Equal As Integers    ${device_amount}    ${expected_count}

Wait For Job Completion And Check Logs
    [Documentation]    Waits for Job completion and checks Pod logs looking for 'Test successful' message

    Oc Wait    -n ${NAMESPACE} job/gdp-test    --for=condition=complete --timeout=${WAIT_TIMEOUT}
    ${pod}=    Oc Get JsonPath
    ...    pod
    ...    ${NAMESPACE}
    ...    --selector=batch.kubernetes.io/job-name=gdp-test
    ...    .items[*].metadata.name
    ${logs}=    Oc Logs    ${pod}    ${NAMESPACE}
    Should Contain    ${logs}    Test successful

Create TtyUSB Devices
    [Documentation]    Creates dummy ttyUSB devices
    FOR    ${i}    IN RANGE    ${DEVICE_COUNT}
        Command Execution    sudo mknod --mode=666 /dev/ttyUSB${i} c 166 ${i}
    END

Remove Dummy TtyUSB Devices
    [Documentation]    Removes dummy ttyUSB devices
    FOR    ${i}    IN RANGE    ${DEVICE_COUNT}
        Command Execution    sudo rm -f /dev/ttyUSB${i}
    END

TtyUSB Glob Test Setup
    [Documentation]    Setup for ttyUSB glob pattern test - creates devices, configures GDP, waits for allocation
    Create TtyUSB Devices
    Drop In MicroShift Config    ${GDP_CONFIG_SERIAL_GLOB}    10-gdp
    Restart MicroShift
    Wait Until Device Is Allocatable    ${DEVICE_COUNT}

TtyUSB Glob Test Teardown
    [Documentation]    Cleanup for ttyUSB glob pattern test
    Oc Delete    pod --all -n ${NAMESPACE}
    Remove Dummy TtyUSB Devices
    Disable GDP

Verify Node Device Allocation
    [Documentation]    Verifies device allocation on the node matches expected total
    [Arguments]    ${expected_total_allocated}
    ${node}=    Run With Kubeconfig    oc get node -o=name
    ${node_name}=    Remove String    ${node}    node/
    ${describe_output}=    Run With Kubeconfig    oc describe node ${node_name}
    ${allocated_line}=    Get Lines Containing String    ${describe_output}    device.microshift.io/fakeserial
    ${allocated_count}=    Get Regexp Matches
    ...    ${allocated_line}
    ...    device\\.microshift\\.io/fakeserial\\s+(\\d+)
    ...    1
    VAR    ${allocated_count}=    ${allocated_count[0]}
    Should Be Equal As Integers    ${allocated_count}    ${expected_total_allocated}

Create Pod And Verify Allocation
    [Documentation]    Creates a pod using dynamic spec generation and verifies device allocation
    [Arguments]    ${pod_name}    ${requested_devices}    ${expected_total_allocated}

    # Generate pod spec dynamically
    ${pod_spec}=    Get Ttyusb Pod Definition    ${pod_name}    ${requested_devices}

    # Create and wait for pod
    ${path}=    Create Random Temp File    ${pod_spec}
    Oc Create    -f ${path} -n ${NAMESPACE}
    Oc Wait    -n ${NAMESPACE} pod/${pod_name}    --for=condition=Ready --timeout=${WAIT_TIMEOUT}

    # Verify correct number of devices allocated to pod
    ${devices_in_pod}=    Oc Exec    ${pod_name}    ls /dev/ | grep ttyUSB | wc -l    ${NAMESPACE}
    Should Be Equal As Integers    ${devices_in_pod}    ${requested_devices}

    # Verify node shows correct allocation
    Verify Node Device Allocation    ${expected_total_allocated}

Teardown Suite With GDP Cleanup
    [Documentation]    Suite teardown that cleans up GDP configuration and restarts MicroShift
    # Clean up any remaining GDP configuration
    Remove Drop In MicroShift Config    10-gdp
    # Restart MicroShift to clean state for next suite
    Restart MicroShift
    Restart Greenboot And Wait For Success
    Teardown Suite With Namespace
