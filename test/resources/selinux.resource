*** Settings ***
Documentation       Keywords for OSTree-based systems

Library             Collections
Library             selinux.py


*** Keywords ***
Run Container File Access
    [Documentation]    Given a file path, converts it to a list and runs an access check With
    ...    default list of files to check the access of, runcon command as a container_t
    ...    returns a list of errors if access was granted
    [Arguments]    ${file_path}
    ${file_paths}=    Create List    ${file_path}
    ${err_list}=    Run Access Check    ${file_paths}
    RETURN    ${err_list}

Run Fcontext Types Check
    [Documentation]    Validates file access by running a runcon command as a container_t
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
