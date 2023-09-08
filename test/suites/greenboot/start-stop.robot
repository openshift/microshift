*** Settings ***
Documentation       Tests related to measuring MicroShift start and stop times

Resource            ../../resources/systemd.resource
Resource            ../../resources/ostree.resource
Resource            ../../resources/microshift-config.resource

Suite Setup         Setup Suite

# Suite Teardown    Cleanup And Start MicroShift
Test Tags           restart    slow


*** Variables ***
${START_TIME_LIMIT}                 70
${START_TIME_DELTA}                 15

${STOP_TIME_LIMIT}                  15
${STOP_TIME_DELTA}                  5

# The actual service readiness is 70s because Greenboot
# script spends 1m on checking pods are not restarting
${CLEAN_GREENBOOT_TIME_LIMIT}       130
${CLEAN_GREENBOOT_TIME_DELTA}       15

# The actual service readiness is 20s because Greenboot
# script spends 1m on checking pods are not restarting
${SECOND_GREENBOOT_TIME_LIMIT}      80
${SECOND_GREENBOOT_TIME_DELTA}      10


*** Test Cases ***
Verify MicroShift Clean Start Times
    [Documentation]    Verify MicroShift service start performance in a clean environment.

    Cleanup And Enable MicroShift    ${TRUE}
    Verify Crio Resource Count    ${TRUE}

    Measure Keyword Runtime With Limits    Start MicroShift
    ...    ${START_TIME_LIMIT}    ${START_TIME_DELTA}
    Measure Keyword Runtime With Limits    Restart Greenboot And Wait For Success
    ...    ${CLEAN_GREENBOOT_TIME_LIMIT}    ${CLEAN_GREENBOOT_TIME_DELTA}
    Measure Keyword Runtime With Limits    Stop MicroShift
    ...    ${STOP_TIME_LIMIT}    ${STOP_TIME_DELTA}

Verify MicroShift Start Times With Container Images Present
    [Documentation]    Verify MicroShift service start performance with container
    ...    images present on the local host.

    Cleanup And Enable MicroShift    ${FALSE}
    Verify Crio Resource Count    ${FALSE}

    Measure Keyword Runtime With Limits    Start MicroShift
    ...    ${START_TIME_LIMIT}    ${START_TIME_DELTA}
    Measure Keyword Runtime With Limits    Restart Greenboot And Wait For Success
    ...    ${SECOND_GREENBOOT_TIME_LIMIT}    ${SECOND_GREENBOOT_TIME_DELTA}
    Measure Keyword Runtime With Limits    Stop MicroShift
    ...    ${STOP_TIME_LIMIT}    ${STOP_TIME_DELTA}

Verify MicroShift Restart Times With Container Images Present
    [Documentation]    Verify MicroShift service restart performance with container
    ...    images present on the local host.

    Cleanup And Enable MicroShift    ${FALSE}
    Verify Crio Resource Count    ${FALSE}

    Measure Keyword Runtime With Limits    Start MicroShift
    ...    ${START_TIME_LIMIT}    ${START_TIME_DELTA}
    Measure Keyword Runtime With Limits    Restart Greenboot And Wait For Success
    ...    ${SECOND_GREENBOOT_TIME_LIMIT}    ${SECOND_GREENBOOT_TIME_DELTA}
    Measure Keyword Runtime With Limits    Stop MicroShift
    ...    ${STOP_TIME_LIMIT}    ${STOP_TIME_DELTA}

    Measure Keyword Runtime With Limits    Start MicroShift
    ...    ${START_TIME_LIMIT}    ${START_TIME_DELTA}
    Measure Keyword Runtime With Limits    Restart Greenboot And Wait For Success
    ...    ${SECOND_GREENBOOT_TIME_LIMIT}    ${SECOND_GREENBOOT_TIME_DELTA}


*** Keywords ***
Setup Suite
    [Documentation]    Set up all of the tests in this suite
    Check Required Env Variables
    Login MicroShift Host

Measure Keyword Runtime With Limits
    [Documentation]    Run a keyword and verify its run time is within the
    ...    specified limits.
    [Arguments]    ${keyword_name}    ${runtime_limit}    ${runtime_delta}

    ${start_time}=    Get Time    epoch
    Run Keyword
    ...    ${keyword_name}
    ${end_time}=    Get Time    epoch
    ${elapsed_time}=    Evaluate    ${end_time} - ${start_time}

    # If the keyword took longer than expected to execute,
    # make sure the delta is within the expected limits
    IF    ${elapsed_time} > ${runtime_limit}
        ${current_delta}=    Evaluate    ${elapsed_time} - ${runtime_limit}
        Should Be True    ${runtime_delta} >= ${current_delta}
    END

Restart Greenboot And Wait For Success
    [Documentation]    Restart the greenboot-healthcheck service and check its status
    Systemctl    restart    greenboot-healthcheck.service
    Wait Until Greenboot Health Check Exited

Cleanup And Enable MicroShift
    [Documentation]    Wipe Microshift data and enable the service
    [Arguments]    ${full_clean}

    IF    ${full_clean}
        Cleanup MicroShift
    ELSE
        Cleanup MicroShift    "--keep-images"
    END
    Systemctl    enable    microshift

Cleanup And Start MicroShift
    [Documentation]    Wipe Microshift data and start the service
    Cleanup MicroShift    "--keep-images"
    Systemctl    enable    --now microshift
    Restart Greenboot And Wait For Success

Verify Crio Resource Count
    [Documentation]    Make sure that cri-o pod, container
    ...    and image counts are correct
    [Arguments]    ${after_full_clean}

    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    crictl pods | wc -l
    ...    sudo=True    return_stdout=True    return_stderr=True    return_rc=True
    Should Be Equal As Integers    ${rc}    0
    Should Be Equal As Integers    ${stdout}    1

    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    crictl ps | wc -l
    ...    sudo=True    return_stdout=True    return_stderr=True    return_rc=True
    Should Be Equal As Integers    ${rc}    0
    Should Be Equal As Integers    ${stdout}    1

    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    crictl images | wc -l
    ...    sudo=True    return_stdout=True    return_stderr=True    return_rc=True
    Should Be Equal As Integers    ${rc}    0
    IF    ${after_full_clean}
        Should Be Equal As Integers    ${stdout}    1
    ELSE
        Should Be True    ${stdout} > 1
    END
