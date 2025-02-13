*** Settings ***
Documentation    This suite performs basic log scans to determine if the opentelemetry-collector is exporting data and
...    in a healthy state

Library    OperatingSystem
Library    ../../resources/journalctl.py

Resource   ../resources/kubeconfig.resource
Resource   ../resources/common.resource
Resource    ../../resources/common.resource

Suite Setup    Setup Suite And Set Journal Cursor
Suite Teardown    Teardown

*** Variables ***
${JOURNAL_CUR}    ${EMPTY}


*** Test Cases ***
MicroShift Observability Logs Should Contain Export Debug Data
    [Documentation]    OpenTelemetry logs should contain the debug data from the LogsExporter and MetricsExporter

    Should Contain Metric Exporter Debug Data
    Should Contain Log Exporter Debug Data

MicroShift Observability Logs Should Not Contain Errors
    [Documentation]    OpenTelemetry logs should not contain errors by any component (reciver, processor,exporter).

    Should Not Contain Component Errors


*** Keywords ***
Should Contain Metric Exporter Debug Data
    [Documentation]    Scan the microshift-observability journal for metric debugging info

    ${pattern}   Set Variable    info\ \ \ \ \ \ \ \ MetricsExporter\ \ \ \ \ \ \ \ {\"kind\": \"exporter\", \"data_
    ...    type\": \"metrics\", \"name\": \"debug\", \"resource metrics\": [0-9]{1,4}, \"metrics\": [0-9]{1,4},\
    ...    \"data points\": [0-9]{1,4}}
    Pattern Should Appear In Log Output    cursor    pattern    unit="microshift-observability"

Should Contain Log Exporter Debug Data
    [Documentation]    Scan the microshift-observability journal for metric debugging info

    ${pattern}   Set Variable    info\ \ \ \ \ \ \ \ LogsExporter\ \ \ \ \ \ \ \ {\"kind\": \"exporter\", \"data_type\
    ...    ": \"logs\", \"name\": \"debug\", \"resource logs\": [0-9]{1,4}, \"log records\": [0-9]{1,4}}
    Pattern Should Appear In Log Output    cursor    pattern    unit="microshift-observability"

Should Not Contain Component Errors
    [Documentation]    Scan the microshift-observability journal for errors reported by collector components
    ${pattern}   Set Variable    info\ \ \ \ \ \ \ \ LogsExporter\ \ \ \ \ \ \ \ {\"kind\": \"exporter\", \"data_type\
    ...    ": \"logs\", \"name\": \"debug\", \"resource logs\": [0-9]{1,4}, \"log records\": [0-9]{1,4}}
    Pattern Should Appear In Log Output    cursor    pattern    unit="microshift-observability"

Setup Suite And Set Journal Cursor
    [Documentation]    The service starts after MicroShift starts and thus will start generating pertinant log data
    ...    right away. When the suite is executed, immediately get the cursor for the current
    Setup Suite
    ${cur}    Get Journal Cursor
    Set Suite Variable    ${JOURNAL_CUR}   ${cur}

Teardown
    [Documentation]    Test suite teardown
    ${JOURNAL_CUR}    Set Variable    ${EMPTY}
    Teardown Suite
