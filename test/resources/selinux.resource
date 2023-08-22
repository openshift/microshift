*** Settings ***
Documentation       Keywords for OSTree-based systems

Library             Collections
Library             selinux.py


*** Keywords ***
Run Container Access Check On
    [Documentation]    Given a file path or optional cmd, runcon command as a container_t,
    ...    returns a list of errors if access was granted
    [Arguments]    ${file_path}    ${cmd}=cat

    ${cmd_list}=    Create List    ${cmd}
    ${file_paths_map}=    Create Dictionary    ${file_path}=${cmd_list}
    ${err_list}=    Run Access Check    ${file_paths_map}
    RETURN    ${err_list}

Run Default Container Access Check
    [Documentation]    Runs default file path checks using runcon command as a container_t,
    ...    returns a list of errors if access was granted

    ${err_list}=    Run Default Access Check
    RETURN    ${err_list}

Run Fcontext Types Check
    [Documentation]    Checks main MicroShift directories paths for expected fcontext,
    ...    returns a list of errors for fcontexts that did not match
    ${err_list}=    Run Fcontext Check
    RETURN    ${err_list}

Get Current Fcontext List
    [Documentation]    Queries the system for a list of fcontext for
    ...    kubernetes_file_t container_var_lib_t kubelet_exec_t container_t
    ${list}=    Get Fcontext List
    RETURN    ${list}

Get Expected Fcontext List
    [Documentation]    Returns the expected fcontext list for OCP and MicroShift
    ${list}=    Get Expected Ocp Microshift Fcontext List
    RETURN    ${list}

Get Denial Audit Log List
    [Documentation]    Quries for denial logs for MicroShift and returns it as a list
    ${list}=    Get Denial Audit Log
    RETURN    ${list}
