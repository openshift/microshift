*** Settings ***
Documentation       Keywords for running the microshift command line.

Library             OperatingSystem
Library             Process
Library             String
Library             SSHLibrary
Resource            oc.resource
Resource            systemd.resource
Resource            microshift-host.resource
Resource            microshift-etcd-process.resource
Library             DataFormats.py


*** Keywords ***
MicroShift Version
    [Documentation]    Run the version command
    ${version_text}    ${rc}=    Execute Command
    ...    microshift version -o yaml
    ...    sudo=True    return_rc=True
    Should Be Equal As Integers    ${rc}    0
    Should Not Be Empty    ${version_text}
    ${version}=    Yaml Parse    ${version_text}
    RETURN    ${version}

MicroShift Is Ready
    [Documentation]    Check the /readyz endpoint
    ${stdout}=    Run With Kubeconfig    oc get --raw='/readyz'
    Should Be Equal As Strings    ${stdout}    ok    strip_spaces=True

MicroShift Is Live
    [Documentation]    Check the /livez endpoint
    [Arguments]    ${extra_args}=${EMPTY}
    ${stdout}=    Run With Kubeconfig    oc get --raw='/livez' ${extra_args}
    Should Be Equal As Strings    ${stdout}    ok    strip_spaces=True

Wait For MicroShift
    [Documentation]    Wait for various checks to ensure MicroShift is online.
    Wait For MicroShift Service
    Wait Until Keyword Succeeds    30x    10s
    ...    MicroShift Is Ready
    Wait Until Keyword Succeeds    30x    10s
    ...    MicroShift Is Live
    # We could also wait for relevant pods. Can we restructure the
    # greenboot check script to let us use it even when not on a
    # greenboot host?

MicroShift Process ID
    [Documentation]    Return the current MicroShift process ID
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    pidof microshift
    ...    return_stdout=True    return_stderr=True    return_rc=True
    Log    ${stderr}
    RETURN    ${stdout}

MicroShift Process ID Changed
    [Documentation]    Compare the specified and current MicroShift process ID
    [Arguments]    ${old_pid}
    ${cur_pid}=    MicroShift Process ID
    Should Not Be Equal    ${cur_pid}    ${old_pid}

Wait Until MicroShift Process ID Changes
    [Documentation]    Wait until the current MicroShift process ID changes
    [Arguments]    ${old_pid}

    Wait Until Keyword Succeeds    30x    10s
    ...    MicroShift Process ID Changed    ${old_pid}

MicroShift Service Is Active
    [Documentation]    Check if microshift.service is active
    ...    (process announced readiness)
    ${state}=    Get Systemd Setting    microshift.service    ActiveState
    Should Be Equal As Strings    ${state}    active

Wait For MicroShift Service
    [Documentation]    Wait for MicroShift service to be active.
    ...    Budget covers TimeoutStartSec=4m (240s) with margin.
    Wait Until Keyword Succeeds    30x    10s
    ...    MicroShift Service Is Active

Restart MicroShift    # robocop: off=LEN01
    [Documentation]    Restart the MicroShift service safely, handling all systemd
    ...    states including auto-restart loops after failures.
    ...
    ...    Background: microshift.service has Restart=always. After a startup
    ...    failure (e.g. invalid config), systemd auto-restarts the process
    ...    repeatedly. Each attempt fails and increments the rate-limit counter.
    ...    Once StartLimitBurst (default 5) is exceeded within
    ...    StartLimitIntervalSec (default 10s), systemd triggers start-limit-hit
    ...    and the service enters the "failed" state, refusing further starts.
    ...
    ...    Step 1 — Wait for transition to settle (30x 10s = 300s).
    ...    Poll ActiveState until it leaves activating/deactivating/reloading.
    ...    This lets the auto-restart loop run its course and hit start-limit-hit
    ...    before we proceed. Without this wait, reset-failed (step 2) clears the
    ...    rate-limit counter, but pending auto-restarts keep firing and
    ...    re-increment it — our restart in step 4 would be rejected.
    ...    Budget covers TimeoutStartSec=4m (240s) for legitimate slow startups.
    ...
    ...    Step 2 — Clear start-limit-hit via reset-failed.
    ...    Safe to call now because the auto-restart loop has settled (step 1
    ...    confirmed the service is no longer transitioning). Resets the
    ...    rate-limit counter and clears the "failed" state so the next start
    ...    is accepted.
    ...
    ...    Step 3 — Capture current PIDs for both microshift and microshift-etcd.
    ...    Taken after stability is confirmed (step 1) so the PIDs are stable
    ...    and will not change before our restart in step 4.
    ...
    ...    Step 4 — Issue systemctl restart (synchronous).
    ...    The Systemctl keyword runs the command, checks RC=0, and waits for
    ...    SubState=running. For Type=notify, this blocks until the new process
    ...    sends READY=1 or TimeoutStartSec expires.
    ...
    ...    Step 5 — Verify PID changes for microshift and microshift-etcd.
    ...    Confirms both processes actually restarted as new PIDs. Passes
    ...    instantly since step 4 already waited for the service to be running.
    ...    Defense-in-depth against Restart=always auto-restart racing with our
    ...    restart — ensures the running process is ours, not a stale
    ...    auto-restarted instance that happened to reach SubState=running.
    ...
    ...    Step 6 — Refresh kubeconfig and wait for readiness.
    ...    Certificates may rotate across restarts, so kubeconfig must be
    ...    re-fetched. Wait For MicroShift polls /readyz and /livez as a
    ...    final sanity check that the API surface is fully available.
    ...
    ...    All wait budgets are 30x 10s (300s), covering TimeoutStartSec=4m
    ...    (240s) with margin.
    #
    # Step 1: wait for auto-restart loop to settle
    Wait Until Keyword Succeeds    30x    10s
    ...    MicroShift Service Is Not Transitioning
    # Step 2: clear start-limit-hit
    Execute Command    systemctl reset-failed microshift.service    sudo=True
    # Step 3: snapshot PIDs before restart
    ${ushift_pid}=    MicroShift Process ID
    ${ushift_etcd_pid}=    MicroShift Etcd Process ID
    # Step 4: synchronous restart
    Systemctl    restart    microshift.service
    # Step 5: confirm PID changes
    Wait Until MicroShift Process ID Changes    ${ushift_pid}
    Wait Until MicroShift Etcd Process ID Changes    ${ushift_etcd_pid}
    # Step 6: refresh kubeconfig and verify readiness
    Setup Kubeconfig
    Wait For MicroShift

MicroShift Service Is Not Transitioning
    [Documentation]    Verify microshift.service is not in a transitional ActiveState
    ...    (activating, deactivating, reloading). This covers all transient
    ...    sub-states including auto-restart and start-pre/start/stop phases.
    ${state}=    Get Systemd Setting    microshift.service    ActiveState
    Should Not Be Equal As Strings    ${state}    activating
    Should Not Be Equal As Strings    ${state}    deactivating
    Should Not Be Equal As Strings    ${state}    reloading

Stop MicroShift
    [Documentation]    Stop the MicroShift service
    Systemctl    stop    microshift.service

Start MicroShift
    [Documentation]    Start the MicroShift service
    Systemctl    start    microshift.service
    # Kubeconfig may change between restarts, ensure we are taking the latest one.
    Setup Kubeconfig

Enable MicroShift
    [Documentation]    Enable the MicroShift service
    Systemctl    enable    microshift.service

Disable MicroShift
    [Documentation]    Disable the MicroShift service
    Systemctl    disable    microshift.service

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
    Log Many    ${stdout}    ${stderr}    ${rc}
    Should Be Equal As Integers    ${rc}    0

Cleanup MicroShift
    [Documentation]    Cleanup MicroShift data
    [Arguments]    ${cmd}="--all"    ${opt}=${EMPTY}

    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    echo 1 | sudo microshift-cleanup-data ${cmd} ${opt}
    ...    return_stdout=True    return_stderr=True    return_rc=True
    Log Many    ${stdout}    ${stderr}    ${rc}
    Should Be Equal As Integers    ${rc}    0

    Enable MicroShift
