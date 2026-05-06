*** Settings ***
Documentation       Tests that the API server rejects requests during startup

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-host.resource
Resource            ../../resources/microshift-process.resource

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           restart


*** Variables ***
${READINESS_HEADER}     X-OpenShift-Internal-If-Not-Ready: reject
${APIS_ENDPOINT}        https://localhost:6443/apis


*** Test Cases ***
API Server Rejects Requests During Startup
    [Documentation]    When the X-OpenShift-Internal-If-Not-Ready: reject header is sent
    ...    and the API server is not yet ready, it should return HTTP 429.
    ...    This prevents internal clients from using a partially initialized server.
    Stop MicroShift
    Start MicroShift Without Waiting For Systemd Readiness

    ${found_429}=    Poll For 429 During Startup
    Should Be True    ${found_429}    API server did not return 429 during startup

    [Teardown]    Restart MicroShift


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host
    Remove Kubeconfig

Poll For 429 During Startup
    [Documentation]    Poll the API server until we observe a 429 response or the server becomes ready.
    ...    Returns True if 429 was observed, False otherwise.
    VAR    ${found}=    ${FALSE}
    FOR    ${i}    IN RANGE    300
        ${stdout}    ${stderr}    ${rc}=    Execute Command
        ...    curl -sk -o /dev/null -w "%{http_code}" -H "${READINESS_HEADER}" ${APIS_ENDPOINT}
        ...    sudo=True    return_stdout=True    return_stderr=True    return_rc=True
        IF    "${stdout}" == "429"
            VAR    ${found}=    ${TRUE}
            BREAK
        END
        IF    "${stdout}" == "200" or "${stdout}" == "401" or "${stdout}" == "403"
            BREAK
        END
        Sleep    0.2s
    END
    RETURN    ${found}
