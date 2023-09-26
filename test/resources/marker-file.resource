*** Settings ***
Documentation       Keywords for managing marker file used for validating data management.

Library             SSHLibrary
Resource            ostree-data.resource


*** Variables ***
${DATA_DIR}     /var/lib/microshift


*** Keywords ***
Create Marker File
    [Documentation]    Creates a marker file used for validating data persistence.
    ...    For example:
    ...    - if created in the backup directory, then it can be used to validate if the backup was restored,
    ...    - if created in data dir, can be used that data wasn't cleaned or overwritten with backup.
    [Arguments]    ${destination_dir}

    ${mark_stdout}    ${mark_stderr}    ${mark_rc}=    Execute Command
    ...    touch ${destination_dir}/marker
    ...    sudo=True    return_stderr=True    return_rc=True
    Should Be Equal As Integers    0    ${mark_rc}

Marker Should Exist
    [Documentation]    Checks if marker file exists inside provided directory
    [Arguments]    ${destination_dir}

    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    stat ${destination_dir}/marker
    ...    sudo=True    return_stderr=True    return_rc=True
    Should Be Equal As Integers    0    ${rc}

Remove Marker
    [Documentation]    Removes marker file from given data directory
    [Arguments]    ${destination_dir}

    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    rm -fv ${destination_dir}/marker
    ...    sudo=True    return_stderr=True    return_rc=True
    Should Be Equal As Integers    0    ${rc}

Create Marker File In Data Dir
    [Documentation]    Creates a marker file in MicroShift data directory
    Create Marker File    ${DATA_DIR}

Marker Should Exist In Data Dir
    [Documentation]    Checks if marker file exists in MicroShift data directory

    Marker Should Exist    ${DATA_DIR}

Remove Marker From Data Dir
    [Documentation]    Removes marker file from MicroShift data directory

    Remove Marker    ${DATA_DIR}
