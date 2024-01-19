*** Settings ***
Documentation       Keywords for OSTree-based systems

Resource            ostree-data.resource
Library             Collections
Library             SSHLibrary
Library             selinux.py


*** Keywords ***
Validate SELinux With Backup
    [Documentation]    Wrapper call for all SELinux checks with extra check inside backup folder
    [Arguments]    ${backup_name}

    ${result}=    Run Access Check On Dir    "${BACKUP_STORAGE}/${backup_name}"
    Should Be Empty    ${result}

    Validate SELinux

Validate SELinux
    [Documentation]    Wrapper call for all SELinux checks

    Containers Should Not Have Access To Container Var Lib Labels
    Context Traversal Should Not Gain More Access
    Folders Should Have Expected Fcontext Types
    Semanage Fcontext Should Have MicroShift Rules
    Audit Log Should Be Empty For MicroShift

Containers Should Not Have Access To Container Var Lib Labels
    [Documentation]    Performs a check to make sure containers can not access
    ...    files or folders that are labeled with container var lib as well as the
    ...    generated backup file

    ${default_result}=    Run Default Access Check
    Should Be Empty    ${default_result}

    # Copy over a test file that forks and execs the given command
    SSHLibrary.Put File    ./assets/selinux-test-exec.sh    /tmp/selinux-test-exec.sh

    # Here we check that given a priveleged context, it can not domain transition to a context with privelage.
    #
    # Validate that the given path `kubelet_exec_t -> kubelet_t -> container_var_lib_t` should not be possible
    # when running under context container_t
    ${default_transition_result}=    Run Default Access Binary Transition Check    /tmp/selinux-test-exec.sh
    Should Be Empty    ${default_transition_result}

Folders Should Have Expected Fcontext Types
    [Documentation]    Performs a check to make sure the folders created during rpm install
    ...    have the expected fcontext values

    ${err_list}=    Run Fcontext Check
    Should Be Empty    ${err_list}

Semanage Fcontext Should Have MicroShift Rules
    [Documentation]    Validates that the fcontext data includes the MicroShift rules
    # Refer to ${reporoot}/packaging/selinux/microshift.fc
    ${result}=    Get Fcontext List
    List Should Contain Value    ${result}    /etc/microshift(/.*)?
    List Should Contain Value    ${result}    /usr/bin/microshift
    List Should Contain Value    ${result}    /usr/bin/microshift-etcd
    List Should Contain Value    ${result}    /usr/lib/microshift(/.*)?
    List Should Contain Value    ${result}    /usr/local/bin/microshift
    List Should Contain Value    ${result}    /usr/local/bin/microshift-etcd
    List Should Contain Value    ${result}    /var/lib/microshift(/.*)?
    List Should Contain Value    ${result}    /var/lib/microshift-backups(/.*)?
    List Should Contain Value    ${result}    /var/lib/microshift\\.saved(/.*)?

Audit Log Should Be Empty For MicroShift
    [Documentation]    Checks that no permission denials have occured during running MicroShift

    ${result}=    Get Denial Audit Log
    Should Be Empty    ${result}

Context Traversal Should Not Gain More Access
    [Documentation]    Checks that no extra permissions are gained via domain and context changes.

    ${result}=    Run Default Traversal Access Check
    Should Be Empty    ${result}

# Helper Functions

Run Container Access Check On File
    [Documentation]    Given a file path or optional cmd, runcon command as a container_t,
    ...    returns a list of errors if access was granted
    [Arguments]    ${file_path}    ${cmd}=cat

    ${cmd_list}=    Create List    ${cmd}
    ${file_paths_map}=    Create Dictionary    ${file_path}=${cmd_list}
    ${err_list}=    Run Access Check    ${file_paths_map}
    RETURN    ${err_list}
