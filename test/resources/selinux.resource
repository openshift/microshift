*** Settings ***
Documentation       Keywords for OSTree-based systems

Resource            ostree-data.resource
Library             Collections
Library             selinux.py


*** Keywords ***
Validate SELinux With Backup
    [Documentation]    Wrapper call for all SELinux checks with extra check inside backup folder
    [Arguments]    ${backup_name}

    ${result}=    Run Container Access Check On    "${BACKUP_STORAGE}/${backup_name}/resources/kubeadmin/kubeconfig"
    Should Be Empty    ${result}

    Validate SELinux

Validate SELinux
    [Documentation]    Wrapper call for all SELinux checks

    Containers Should Not Have Access To Container Var Lib Labels
    Folders Should Have Expected Fcontext Types
    Semanage Fcontext Should Have Combined List Of OCP And MicroShift Rules
    Audit Log Should Be Empty For MicroShift

Containers Should Not Have Access To Container Var Lib Labels
    [Documentation]    Performs a check to make sure containers can not access
    ...    files or folders that are labeled with container var lib as well as the
    ...    generated backup file

    ${default_result}=    Run Default Access Check
    Should Be Empty    ${default_result}

Folders Should Have Expected Fcontext Types
    [Documentation]    Performs a check to make sure the folders created during rpm install
    ...    have the expected fcontext values

    ${err_list}=    Run Fcontext Check
    Should Be Empty    ${err_list}

Semanage Fcontext Should Have Combined List Of OCP And MicroShift Rules
    [Documentation]    Validates that the fcontext data is the combined set for
    ...    OCP and MicroShift

    ${result}=    Get Fcontext List
    ${expected}=    Get Expected Ocp Microshift Fcontext List
    Lists Should Be Equal    ${result}    ${expected}

Audit Log Should Be Empty For MicroShift
    [Documentation]    Checks that no permission denials have occured during running MicroShift

    ${result}=    Get Denial Audit Log
    Should Be Empty    ${result}

# Helper Functions

Run Container Access Check On
    [Documentation]    Given a file path or optional cmd, runcon command as a container_t,
    ...    returns a list of errors if access was granted
    [Arguments]    ${file_path}    ${cmd}=cat

    ${cmd_list}=    Create List    ${cmd}
    ${file_paths_map}=    Create Dictionary    ${file_path}=${cmd_list}
    ${err_list}=    Run Access Check    ${file_paths_map}
    RETURN    ${err_list}
