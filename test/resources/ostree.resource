*** Settings ***
Documentation       Keywords for OSTree-based systems

Resource            systemd.resource
Resource            microshift-process.resource
Library             libostree.py


*** Variables ***
${DATA_DIR}             "/var/lib/microshift"
${BACKUP_STORAGE}       "/var/lib/microshift-backups"


*** Keywords ***
Wait Until Greenboot Health Check Exited
    [Documentation]    Wait until greenboot healthchecks are done

    Wait Until Keyword Succeeds    10m    15s
    ...    Greenboot Health Check Exited

Greenboot Health Check Exited
    [Documentation]    Checks if greenboot-healthcheck finished running successfully (exited)

    ${value}=    Get Systemd Setting    greenboot-healthcheck.service    SubState
    Should Be Equal As Strings    ${value}    exited

Get Future Backup Name For Current Boot
    [Documentation]    Provides a name for future backup:
    ...    "4.13" for MicroShift 4.13, otherwise deployment ID + boot ID

    # When system is upgrading from 4.13, there is no metadata regarding previous boot,
    # therefore, when naming backup for existing data, "4.13" will be used as deployment ID,
    # i.e. it should reside in `/var/lib/microshift-backups/4.13`
    ${version}=    MicroShift Version
    IF    ${version.minor} == 13    RETURN    4.13

    ${deploy_id}=    Get Booted Deployment ID
    ${boot_id}=    Get Current Boot Id
    ${backup_name}=    Catenate    SEPARATOR=_    ${deploy_id}    ${boot_id}
    RETURN    ${backup_name}

Backup Should Exist
    [Documentation]    Checks if backup identified by deployment ID
    ...    and (optionally) boot ID exists
    [Arguments]    ${deployment_id}    ${boot_id}=${EMPTY}
    ${exists}=    Does Backup Exist    ${deployment_id}    ${boot_id}
    Should Be True    ${exists}

Remove Existing Backup For Current Deployment
    [Documentation]    Remove any existing backup for currently running deployment
    ${deploy_id}=    Get Booted Deployment Id
    Remove Backups For Deployment    ${deploy_id}

Rebase System And Verify
    [Documentation]    Deploys given ostree ref on the host, reboots it,
    ...    and checks if the host is healthy.
    [Arguments]    ${ref}

    Rebase System    ${ref}
    Reboot MicroShift Host

    Wait Until Greenboot Health Check Exited
    System Should Be Healthy

    ${ref_after_reboot}=    Get Current Ref
    Should Be Equal As Strings    ${ref_after_reboot}    ${ref}

System Should Be Healthy
    [Documentation]    Verifies if system is healthy by looking at health.json.
    ...    If this runs after greenboot-healthcheck finishes, it verifies
    ...    health of current boot.

    # Health of system with MicroShift 4.13 is implicitly checked by "Greenboot Health Check Exited"
    ${version}=    MicroShift Version
    IF    ${version.minor} == 13    RETURN

    ${health}=    Get Persisted System Health
    Should Be Equal As Strings    ${health}    healthy

MicroShift 413 Should Not Have Upgrade Artifacts
    [Documentation]    Verifies that host running MicroShift 4.13
    ...    does not have upgrade related artifacts.

    ${version}=    MicroShift Version
    IF    ${version.minor} == 13
        Path Should Not Exist    ${BACKUP_STORAGE}
        Path Should Not Exist    /var/lib/microshift/version
    END
