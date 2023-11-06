*** Settings ***
Documentation       Tests for MicroShift service lifecycle

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-process.resource

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           restart


*** Variables ***
${RESTART_ATTEMPTS}     3


*** Test Cases ***
Restarting MicroShift In The Middle Of Startup Should Succeed
    [Documentation]    Checks if restarting MicroShift during startup fails.
    ...    (for example due to not stopped microshift-etcd.scope).
    [Template]    Restart MicroShift ${TIME_TO_WAIT} Seconds After Starting

    40
    30
    20
    15
    10
    5
    3
    1


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig    # for readiness checks

Teardown
    [Documentation]    Test suite teardown
    Restart MicroShift
    Logout MicroShift Host
    Remove Kubeconfig

Restart MicroShift ${time_to_wait} Seconds After Starting
    [Documentation]    Tests MicroShift's resilience when restarted during startup.

    FOR    ${attempt}    IN RANGE    ${RESTART_ATTEMPTS}
        # Make sure MicroShift is fully stopped first,
        # so it's startup can be interrupted after specified amount of time.
        Stop MicroShift
        Start MicroShift Without Waiting For Systemd Readiness
        Sleep    ${time_to_wait}s
        Restarting MicroShift Should Be Successful On First Try
    END

Restarting MicroShift Should Be Successful On First Try
    [Documentation]    Restarts MicroShift without additional retries, so
    ...    MicroShift will have one chance to shutdown and restart correctly.
    ...    Although normally (on boot) systemd attempts to start service
    ...    several times, here it is expected that single attempt is enough.

    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    systemctl restart microshift
    ...    sudo=True
    ...    return_stdout=True
    ...    return_stderr=True
    ...    return_rc=True
    Should Be Equal As Integers    0    ${rc}

Start MicroShift Without Waiting For Systemd Readiness
    [Documentation]    Starts MicroShift without waiting for daemon readiness
    ...    (which happens after all internal services/components declare ready
    ...    (close ready channel)), so it can be interrupted (restarted) mid startup.

    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    systemctl start microshift --no-block
    ...    sudo=True
    ...    return_stdout=True
    ...    return_stderr=True
    ...    return_rc=True
    Should Be Equal As Integers    0    ${rc}
