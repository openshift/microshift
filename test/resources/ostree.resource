*** Settings ***
Documentation       Keywords for OSTree-based systems

Library             SSHLibrary
Library             DataFormats.py


*** Variables ***
${BACKUP_STORAGE}       /var/lib/microshift-backups
${HEALTH_FILE}          /var/lib/microshift-backups/health.json


*** Keywords ***
Get Booted Deployment Status
    [Documentation]    Get rpm-ostree status of booted deployment as a json

    ${ostree_status}    ${rc}=    Execute Command
    ...    rpm-ostree status --booted --json
    ...    return_rc=True
    Should Be Equal As Integers    0    ${rc}
    ${status}=    Json Parse    ${ostree_status}
    RETURN    ${status}

Get Booted Deployment ID
    [Documentation]    Get ID of currently booted deployment

    ${status}=    Get Booted Deployment Status
    RETURN    ${status.deployments[0].id}

Get Booted Deployment Backup Prefix Path
    [Documentation]    Get backup path prefix for current deployment
    ...
    ...    Prefix path is BACKUP_STORAGE/deployment-id.
    ...    Globbing directories starting with the prefix will yield
    ...    list of backups for the deployment.

    ${current_deployment_id}=    Get Booted Deployment ID
    Should Not Be Empty    ${current_deployment_id}
    Should Not Be Empty    ${BACKUP_STORAGE}
    ${backup_path}=    Catenate    SEPARATOR=/    ${BACKUP_STORAGE}    ${current_deployment_id}
    RETURN    ${backup_path}

Get System Health
    [Documentation]    Get system health information from health.json file

    ${health}    ${stderr}    ${rc}=    Execute Command
    ...    jq -r '.health' ${BACKUP_STORAGE}/health.json
    ...    sudo=True    return_rc=True    return_stderr=True
    Should Be Equal As Integers    0    ${rc}
    RETURN    ${health}

Does Backup For Booted Deployment Exist
    [Documentation]    Checks if any backup for current deployment exists
    ...    by finding paths starting with "backup prefix path"

    ${deploy_id}=    Get Booted Deployment ID
    ${output}    ${rc}=    Execute Command
    ...    find ${BACKUP_STORAGE} -name '${deploy_id}*' | wc -l
    ...    sudo=True    return_rc=True
    Should Be Equal As Integers    0    ${rc}

    IF    "${output}" == "0"    RETURN    ${FALSE}
    RETURN    ${TRUE}
