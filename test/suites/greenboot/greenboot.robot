*** Settings ***
Documentation       Tests related to greenboot

Resource            ../../resources/systemd.resource
Resource            ../../resources/ostree.resource
Resource            ../../resources/ostree-health.resource
Resource            ../../resources/microshift-config.resource

Suite Setup         Setup Suite
Suite Teardown      Teardown Suite

Test Tags           restart    slow


*** Variables ***
${LOCAL_WORKLOAD_FILE}          ../docs/config/busybox_running_check.sh
${GREENBOOT_WORKLOAD_FILE}      /etc/greenboot/check/required.d/50_busybox_running_check.sh
${GREENBOOT_CONFIG_FILE}        /etc/greenboot/greenboot.conf
${GREENBOOT_CONFIG_CONTENT}     MICROSHIFT_WAIT_TIMEOUT_SEC=180
${MANIFEST_SOURCE_DIR}          ./assets/kustomizations/greenboot/
${MANIFEST_DIR}                 /etc/microshift/manifests
${HOSTNAME_BIN_PATH}            ""


*** Test Cases ***
Run with User Workload
    [Documentation]    Add user's workload and verify that the workload starts and greenboot is successful.
    Restart Greenboot And Wait For Success
    Add User Workload
    Cleanup And Start

    [Teardown]    Run Keywords
    ...    Cleanup User Workload
    ...    Cleanup And Start

Simulate Service Failure
    [Documentation]    Simulate Service failure
    Restart Greenboot And Wait For Success
    Disrupt Service
    Cleanup MicroShift    --all    --keep-images
    # not using 'Start MicroShift' kw because it retries
    Run Keyword And Expect Error    0 != 1
    ...    Systemctl    start    microshift
    Run Keyword And Expect Error    0 != 1
    ...    Restart Greenboot And Wait For Success

    [Teardown]    Run Keywords
    ...    Restore Service
    ...    Cleanup And Start

Simulate Pod Failure
    [Documentation]    Simulate pod network failure
    Restart Greenboot And Wait For Success
    Disrupt Pod Network
    Restart MicroShift
    Run Keyword And Expect Error    0 != 1
    ...    Restart Greenboot And Wait For Success

    [Teardown]    Run Keywords
    ...    Remove Drop In MicroShift Config    10-svcNetwork
    ...    AND
    ...    Cleanup And Start


*** Keywords ***
Setup Suite
    [Documentation]    Set up all of the tests in this suite
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig
    Upload String To File    ${GREENBOOT_CONFIG_CONTENT}    ${GREENBOOT_CONFIG_FILE} # change the default timeout

Add User Workload
    [Documentation]    Upload User workload files to the MicroShift host
    Put File    ${LOCAL_WORKLOAD_FILE}    /tmp/busybox_running_check.sh
    ${stdout}    ${rc}=    Execute Command
    ...    [ -f ${GREENBOOT_WORKLOAD_FILE} ] || sudo mv /tmp/busybox_running_check.sh ${GREENBOOT_WORKLOAD_FILE}
    ...    sudo=True    return_rc=True
    Should Be Equal As Integers    0    ${rc}

    Put Directory    ${MANIFEST_SOURCE_DIR}    /tmp/manifests
    ${stdout}    ${rc}=    Execute Command
    ...    cp -f /tmp/manifests/* ${MANIFEST_DIR} && sudo rm -rf /tmp/manifests
    ...    sudo=True    return_rc=True
    Should Be Equal As Integers    0    ${rc}

Cleanup User Workload
    [Documentation]    Remove User workload files from the MicroShift host
    ${stdout}    ${rc}=    Execute Command
    ...    rm -rf ${GREENBOOT_WORKLOAD_FILE} ${MANIFEST_DIR}/*
    ...    sudo=True    return_rc=True
    Should Be Equal As Integers    0    ${rc}

Disrupt Service
    [Documentation]    Prevent Microshift service from starting correctly.

    ${stdout}    ${rc}=    Execute Command
    ...    which hostname
    ...    sudo=False    return_rc=True
    IF    ${rc} == 0    Set Suite Variable    \${HOSTNAME_BIN_PATH}    ${stdout}

    # This covers both ostree and bootc systems
    ${is_ostree}=    Is System OSTree
    IF    ${is_ostree}    Create Usr Directory Overlay

    ${rc}=    Execute Command
    ...    chmod 000 ${HOSTNAME_BIN_PATH}
    ...    sudo=True    return_rc=True
    Should Be Equal As Integers    0    ${rc}

Restore Service
    [Documentation]    Restore Microshift service to the correct form.
    ${stdout}    ${rc}=    Execute Command
    ...    chmod 755 ${HOSTNAME_BIN_PATH}
    ...    sudo=True    return_rc=True
    Should Be Equal As Integers    0    ${rc}

    # Reboot to regain ostree deployment (revert usroverlay)
    ${is_ostree}=    Is System OSTree
    IF    ${is_ostree}    Reboot MicroShift Host

Disrupt Pod Network
    [Documentation]    Prevent Microshift pods From starting correctly
    ${configuration}=    Catenate    SEPARATOR=\n
    ...    network:
    ...    \ clusterNetwork:
    ...    \ - 10.42.0.0/16
    ...    \ serviceNetwork:
    ...    \ - 10.66.0.0/16
    ...
    Drop In MicroShift Config    ${configuration}    10-svcNetwork

Cleanup And Start
    [Documentation]    Wipe Microshift data and start it.
    Cleanup MicroShift    --all    --keep-images
    Start MicroShift
    Setup Kubeconfig
    Restart Greenboot And Wait For Success
