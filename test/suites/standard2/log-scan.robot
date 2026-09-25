*** Settings ***
Documentation       Tests for various log messages we do or do not want.

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/ostree-health.resource
Library             ../../resources/journalctl.py

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           restart    slow


*** Variables ***
${CURSOR}                   ${EMPTY}    # Journal cursor for the current service startup; set by Service Startup And Scan Journal

# Known-benign "forbidden" log lines to ignore during the log scan.
# On a fresh/clean start the cert-manager operator creates the cainjector,
# controller and webhook Deployments and their ServiceAccounts dynamically.
# The kube-controller-manager ReplicaSet controller can briefly try to create
# a pod before the matching ServiceAccount is observed, logging a transient
# "forbidden: error looking up service account" that it retries away once the
# SA lands. These pods are operator-managed (not MicroShift manifests), so the
# ordering is not under MicroShift's control and this is a benign startup race.
@{FORBIDDEN_EXCEPTIONS}
...                         is forbidden: error looking up service account cert-manager/[^ ]+: serviceaccount .* not found


*** Test Cases ***
Log Scan
    [Documentation]    Scan the journal of a clean first service startup and then
    ...    of a restart. Both service startups are checked, but the "forbidden"
    ...    check runs only on the restart: a clean first service startup logs a
    ...    benign "forbidden" while components initialize (see
    ...    ${FORBIDDEN_EXCEPTIONS}), whereas a restart must be free of it.
    Cleanup MicroShift    --all    --keep-images
    Enable MicroShift

    # Clean first service startup: skip the forbidden check.
    Service Startup And Scan Journal    check_forbidden=False
    # Restart: forbidden messages must not reappear.
    Service Startup And Scan Journal


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host

Teardown
    [Documentation]    Test suite teardown
    Restart MicroShift
    Wait For MicroShift Healthcheck Success

    Logout MicroShift Host
    Remove Kubeconfig

Service Startup And Scan Journal
    [Documentation]    Record the journal cursor, start MicroShift, wait until it
    ...    is initialized, stop it, and scan this service startup's journal for
    ...    wanted and unwanted messages.
    [Arguments]    ${check_forbidden}=True
    ${cursor}=    Get Journal Cursor
    VAR    ${CURSOR}=    ${cursor}    scope=SUITE
    Start MicroShift
    Setup Kubeconfig

    Wait For MicroShift Healthcheck Success
    Stop MicroShift

    Scan Service Startup Journal    check_forbidden=${check_forbidden}

Scan Service Startup Journal
    [Documentation]    Assert this service startup's journal contains the expected
    ...    readiness messages and none of the unwanted ones.
    [Arguments]    ${check_forbidden}=True
    IF    ${check_forbidden}    Should Not Find Forbidden
    Should Not Find Cannot Patch Resource
    Services Should Not Timeout When Stopping
    Should Find Etcd Is Ready
    Should Find MicroShift Is Ready

Should Not Find Forbidden
    [Documentation]    Logs should not say "forbidden", excluding known-benign
    ...    startup races listed in ${FORBIDDEN_EXCEPTIONS}.
    Pattern Should Not Appear In Log Output    ${CURSOR}    forbidden    exceptions=${FORBIDDEN_EXCEPTIONS}

Should Not Find Cannot Patch Resource
    [Documentation]    Logs should not say "cannot patch resource"
    Pattern Should Not Appear In Log Output    ${CURSOR}    cannot patch resource

Services Should Not Timeout When Stopping
    [Documentation]    Logs should not say "Timed out waiting for services to stop"
    Pattern Should Not Appear In Log Output    ${CURSOR}    MICROSHIFT STOP TIMED OUT

Should Find Etcd Is Ready
    [Documentation]    Logs should say "etcd is ready"
    Pattern Should Appear In Log Output    ${CURSOR}    etcd is ready

Should Find MicroShift Is Ready
    [Documentation]    Logs should say "MICROSHIFT READY"
    Pattern Should Appear In Log Output    ${CURSOR}    MICROSHIFT READY
