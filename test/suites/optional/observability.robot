*** Settings ***
Documentation    This suite performs basic log scans to determine if the opentelemetry-collector is exporting data and
...    in a healthy state

Library    OperatingSystem
Library    String
Library    ../../resources/journalctl.py

Resource    ../../resources/kubeconfig.resource
Resource    ../../resources/common.resource

Suite Setup    Setup Suite And Set Journal Cursor
Suite Teardown    Teardown Suite


*** Variables ***
${JOURNAL_CUR}    ${EMPTY}


*** Test Cases ***
Logs Should Contain Exported Metric Data
    [Documentation]    OpenTelemetry logs should contain the data from the LogsExporter and MetricsExporter

    ${pattern}   Catenate    SEPARATOR=
    ...    info\\s+MetricsExporter\\s+{"kind": "exporter", "data_type": "metrics", "name": "debug", "resource metrics":
    ...    \ [0-9]+, "metrics": [0-9]+, "data points": [0-9]+}
    Pattern Should Appear In Log Output    ${JOURNAL_CUR}    ${pattern}    unit="microshift-observability"

Logs Should Contain Exported Log Data
    [Documentation]    OpenTelemetry logs should not contain errors by any component (reciver, processor,exporter).

    ${pattern}   Catenate    SEPARATOR=
    ...    info\\s+LogsExporter\\s+\\{"kind": "exporter", "data_type": "logs", "name": "debug", "resource logs": [0-9]+,
    ...    \ "log records": [0-9]+\\}
    Pattern Should Appear In Log Output    ${JOURNAL_CUR}    ${pattern}    unit="microshift-observability"

Logs Should Not Contain Reciever Errors
    [Documentation]    Internal receiver errors are not treated as fatal. Typically these are due to a miconfiguration
    ...    and thus indicate the provided default config should be reviewed.

    ${pattern}   Catenate    SEPARATOR= \\s+\\{"error":.*\\}
    Pattern Should Not Appear In Log Output    ${JOURNAL_CUR}    ${pattern}    unit="microshift-observability"


*** Keywords ***
Setup Suite And Set Journal Cursor
    [Documentation]    The service starts after MicroShift starts and thus will start generating pertinant log data
    ...    right away. When the suite is executed, immediately get the cursor for the current
    Setup Suite
    ${cur}    Get Journal Cursor
    Set Suite Variable    ${JOURNAL_CUR}   ${cur}
    Wait Until Keyword Succeeds    1 min    5 sec
    ...    Journal Contains Enough Lines To Test

Journal Contains Enough Lines To Test
    [Documentation]    Execution should wait until there are at least 10 lines of journal data to process. This is
    ...    necessary because opentelemetry-collector will write debug output in batches. Thus, it can often happen
    ...    that there is not enough log data to gain an accurate read of the process's state, resulting in a false
    ...    negative signal (the opentelemetry-collector is healthy, but did not yet write data to journal.

    ${output}    ${rc}    Get Log Output With Pattern    ${JOURNAL_CUR}    .*    microshift-observability
    ${lineCnt}    Get Line Count    ${output}
    Should Be True    ${lineCnt} > 10
