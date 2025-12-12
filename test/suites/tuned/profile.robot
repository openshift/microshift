*** Settings ***
Documentation       Tests for verification on MicroShift's Tuned profile

Resource            ../../resources/microshift-host.resource
Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/ostree-health.resource
Resource            ../../resources/ostree.resource

Suite Setup         Setup
Suite Teardown      Teardown


*** Test Cases ***
Host Should Have MicroShift-Baseline TuneD Profile Enabled
    [Documentation]    Verifies that microshift-tuned.service was successful in activating the profile
    ...    and rebooting the host.

    Kernel Arguments Should Exist    nohz=on    nohz_full=2,4-5    cu_nocbs=2,4-5    tuned.non_isolcpus=0000000b
    ...    hugepagesz=2M    hugepages=10    test1=on    test2=true    dummy
    CPUs Should Be    0    3
    CPUs Should Be    1    1    2    4    5    # 0 is implicitly online

X86 64 Should Run RT Kernel
    [Documentation]    If system under test is x86_64, assert it's running RT kernel.

    ${arch}=    Command Should Work    uname -m
    IF    "${arch}" == "x86_64"
        ${kernel}=    Command Should Work    uname -r
        Should End With    ${kernel}    +rt
    END

Kubelet Resources Are As Expected
    [Documentation]    Validates that kubelet detected right amount of online CPUs and hugepages.

    Wait Until Greenboot Health Check Exited
    Setup Kubeconfig
    Verify Node Resources    hugepages-2Mi    20Mi
    Verify Node Resources    cpu    5

Created Pod Is Guaranteed And Has Correct CPU Set
    [Documentation]    Verify that Pod has guaranteed QoS and correct CPU set.
    [Setup]    Setup Namespace

    Oc Create    -n ${NAMESPACE} -f ./assets/tuned/pod.yaml
    Named Pod Should Be Ready    oslat

    ${qos_class}=    Oc Get JsonPath    pod    ${NAMESPACE}    oslat    .status.qosClass
    Should Be Equal    ${qos_class}    Guaranteed

    ${cpus}=    Oc Exec    oslat    cat /proc/self/status | grep Cpus_allowed_list: | cut -f 2
    Should Be Equal    ${cpus}    2,4

    Wait Until Oslat Completed Testing

Make Sure Everything Works After Reboot
    [Documentation]    Verify that after reboot MicroShift is starting and our low latency Pod is running.

    Reboot MicroShift Host
    Wait Until Greenboot Health Check Exited
    Named Pod Should Be Ready    oslat

    Wait Until Oslat Completed Testing

    [Teardown]    Remove Namespace    ${NAMESPACE}

Reactivate Offline CPU
    [Documentation]    Verify if reactivating previously offlined CPU is successful.
    [Setup]    Backup MicroShift Baseline Variables

    Command Should Work
    ...    sed -i 's/^offline_cpu_set=.*$/offline_cpu_set=/' /etc/tuned/microshift-baseline-variables.conf
    Command Should Work    tuned-adm profile microshift-baseline
    CPUs Should Be    1    1    2    3    4    5    # 0 is implicitly online

    [Teardown]    Restore MicroShift Baseline Variables


*** Keywords ***
Setup
    [Documentation]    Setup test for the test suite
    Login MicroShift Host

Teardown
    [Documentation]    Teardown test after the test suite
    Logout MicroShift Host

Setup Namespace
    [Documentation]    Setup unique namespace with elevated privileges
    ${ns}=    Create Unique Namespace
    VAR    ${NAMESPACE}=    ${ns}    scope=SUITE
    Run With Kubeconfig    oc label ns ${ns} --overwrite pod-security.kubernetes.io/audit=privileged
    Run With Kubeconfig    oc label ns ${ns} --overwrite pod-security.kubernetes.io/enforce=privileged
    Run With Kubeconfig    oc label ns ${ns} --overwrite pod-security.kubernetes.io/warn=privileged

Verify Node Resources
    [Documentation]    Checks if node resources are as expected
    [Arguments]    ${resource}    ${expected}
    ${is}=    Oc Get JsonPath    node    ${EMPTY}    ${EMPTY}    .items[0].status.capacity.${resource}
    Should Be Equal As Strings    ${is}    ${expected}

Kernel Arguments Should Exist
    [Documentation]    Verify that given kernel arguments are present in the kernel's command line.
    [Arguments]    @{kargs}
    FOR    ${karg}    IN    @{kargs}
        Command Should Work    grep ${karg} /proc/cmdline
    END

CPUs Should Be
    [Documentation]    Verify that given CPUs are offline
    [Arguments]    ${expected}    @{cpus}
    FOR    ${cpu}    IN    @{cpus}
        ${state}=    Command Should Work    cat /sys/devices/system/cpu/cpu${cpu}/online
        Should Be Equal    ${state}    ${expected}
    END

Wait Until Oslat Completed Testing
    [Documentation]    Wait until oslat container finished testing.
    Wait Until Keyword Succeeds    30s    5s
    ...    Oslat Completed Testing

Oslat Completed Testing
    [Documentation]    Check logs of oslat container looking for "Test completed." message.
    ...    We run oslat just to make sure it successfully runs, not for the results.
    ${logs}=    Oc Logs    oslat    ${NAMESPACE}
    Should Contain    ${logs}    Test completed.

Backup MicroShift Baseline Variables
    [Documentation]    Backs up microshift-baseline-variables.conf
    Command Should Work
    ...    cp /etc/tuned/microshift-baseline-variables.conf /etc/tuned/microshift-baseline-variables.conf.bak

Restore MicroShift Baseline Variables
    [Documentation]    Restores up microshift-baseline-variables.conf
    Command Should Work
    ...    mv /etc/tuned/microshift-baseline-variables.conf.bak /etc/tuned/microshift-baseline-variables.conf
