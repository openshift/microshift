*** Settings ***
Documentation       Networking smoke tests

Resource            ../../resources/common.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/ostree-health.resource
Resource            ../../resources/microshift-network.resource

Suite Setup         Setup Suite With Namespace
Suite Teardown      Teardown Suite With Namespace

Test Tags           smoke


*** Test Cases ***
Router Smoke Test
    [Documentation]    Run a router smoke test
    [Setup]    Run Keywords
    ...    Create Hello MicroShift Pod
    ...    AND
    ...    Expose Hello MicroShift Service Via Route

    Wait Until Keyword Succeeds    10x    6s
    ...    Access Hello Microshift    ${HTTP_PORT}

    Resolve Hello MicroShift MDNS

    [Teardown]    Run Keywords
    ...    Delete Hello MicroShift Route
    ...    Delete Hello MicroShift Pod And Service

Load Balancer Smoke Test
    [Documentation]    Verify that Load Balancer correctly exposes HTTP service
    Verify Hello MicroShift LB


*** Keywords ***
Expose Hello MicroShift Service Via Route
    [Documentation]    Expose the "hello microshift" application through the Route
    Oc Expose    pod hello-microshift -n ${NAMESPACE}
    Oc Expose    svc hello-microshift --hostname hello-microshift.cluster.local -n ${NAMESPACE}

Delete Hello MicroShift Route
    [Documentation]    Delete route for cleanup.
    Oc Delete    route/hello-microshift -n ${NAMESPACE}

Resolve Hello MicroShift MDNS
    [Documentation]    Resolve hello-microshift route via mDNS from the hypervisor/RF runner.

    # Open port 5353 on the hypervisor to be able to resolve mDNS hostnames
    ${result}=    Run Process
    ...    sudo firewall-cmd --zone\=libvirt --add-service\=mdns
    ...    shell=True
    ...    timeout=15s
    Log Many    ${result.rc}    ${result.stdout}    ${result.stderr}
    Should Be Equal As Integers    0    ${result.rc}

    ${result}=    Run Process
    ...    avahi-resolve-host-name hello-microshift.cluster.local
    ...    shell=True
    ...    timeout=15s
    Log Many    ${result.stdout}    ${result.stderr}

    # avahi-resolve-host-name always returns rc=0, even if it failed to resolve.
    # In case of failure, stdout will be empty and stderr will contain error.
    # Expected success stdout:
    # > hello-microshift.cluster.local 192.168.124.5
    # Possible stderr:
    # > Failed to resolve host name 'hello-microshift.cluster.local': Timeout reached
    Should Not Be Empty    ${result.stdout}
    Should Be Empty    ${result.stderr}
