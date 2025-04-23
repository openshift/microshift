*** Settings ***
Documentation       This suite performs basic log scans to determine if the opentelemetry-collector is exporting data
...                 and in a healthy state

Library             OperatingSystem
Library             SSHLibrary
Library             ../../resources/journalctl.py
Resource            ../../resources/kubeconfig.resource
Resource            ../../resources/common.resource
Resource            ../../resources/systemd.resource

Suite Setup         Setup Suite And Prepare Test Host
Suite Teardown      Teardown Suite And Revert Test Host


*** Variables ***
${JOURNAL_CUR}          ${EMPTY}
${CONFIG_PATH}          /etc/microshift/opentelemetry-collector.yaml
${CONFIG_ORIGINAL}      ${EMPTY}


*** Test Cases ***
Logs Should Contain Exported Metric Data
    [Documentation]    OpenTelemetry logs should contain the data from the MetricsExporter

    ${pattern}    Catenate    SEPARATOR=
    ...    info\\s+MetricsExporter\\s+{"kind": "exporter", "data_type": "metrics", "name": "debug", "resource metrics":
    ...    \ [0-9]+, "metrics": [0-9]+, "data points": [0-9]+}
    Pattern Should Appear In Log Output    ${JOURNAL_CUR}    ${pattern}    unit="microshift-observability"

Logs Should Contain Exported Log Data
    [Documentation]    OpenTelemetry logs should contain the data from the LogsExporter

    ${pattern}    Catenate
    ...    SEPARATOR=
    ...    info\\s+LogsExporter\\s+\\{"kind": "exporter", "data_type": "logs", "name": "debug", "resource logs": [0-9]+,
    ...    \ "log records": [0-9]+\\}
    Pattern Should Appear In Log Output    ${JOURNAL_CUR}    ${pattern}    unit="microshift-observability"

Logs Should Not Contain Receiver Errors
    [Documentation]    Internal receiver errors are not treated as fatal. Typically these are due to a miconfiguration
    ...    and thus indicate the provided default config should be reviewed.

    ${pattern}    Catenate    SEPARATOR=
    ...    \\s+\\{"error":.*\\}
    Pattern Should Not Appear In Log Output    ${JOURNAL_CUR}    ${pattern}    unit="microshift-observability"


*** Keywords ***
Setup Suite And Prepare Test Host
    [Documentation]    The service starts after MicroShift starts and thus will start generating pertinant log data
    ...    right away. When the suite is executed, immediately get the cursor for the microshift-observability unit.
    Setup Suite
    Deploy Debug Config
    ${cur}    Get Journal Cursor    unit=microshift-observability
    Set Suite Variable    ${JOURNAL_CUR}    ${cur}
    Wait Until Keyword Succeeds    1 min    5 sec
    ...    Journal Contains Enough Lines To Test

Journal Contains Enough Lines To Test
    [Documentation]    Execution should wait until there are at least 20 lines of journal data to process. This is
    ...    necessary because opentelemetry-collector will write debug output in batches. Thus, it can often happen
    ...    that there is not enough log data to gain an accurate read of the process's state, resulting in a false
    ...    negative signal (the opentelemetry-collector is healthy, but did not yet write data to journal.

    ${output}    ${rc}    Get Log Output With Pattern    ${JOURNAL_CUR}    .*    microshift-observability
    ${line_cnt}    Get Line Count    ${output}
    Should Be True    ${line_cnt} > 20

Deploy Debug Config
    [Documentation]    Dynamically set all exporters in the opentelemetry-collector config to "debug", otherwise the
    ...    process will attempt to stream data to a non-existent service, resulting in the log being spammed with
    ...    errors.
    ${cfg}    Create Random Temp File
    SSHLibrary.Get File    ${CONFIG_PATH}    ${cfg}
    Set Suite Variable    ${CONFIG_ORIGINAL}    ${cfg}
    ${yq_cmd}    Catenate
    ...    yq '.service.pipelines |= with_entries(.value.exporters |= [ "debug" ])
    ...    | .exporters["debug"] = {}'
    ${rc}    ${cfg_patched}    OperatingSystem.Run And Return Rc And Output    ${yq_cmd} ${cfg}
    Should Be Equal As Integers    ${rc}    0
    Upload String To File    ${cfg_patched}    ${CONFIG_PATH}
    Systemctl    restart    microshift-observability

Teardown Suite And Revert Test Host
    [Documentation]    Calls the global Teardown Suite keyword and restores the opentelemetry-collector.yaml to its
    ...    original state.
    ${exists}    Run Keyword And Return Status
    ...    OperatingSystem.File Should Exist    ${CONFIG_ORIGINAL}

    IF    ${exists}
        ${cfg}    OperatingSystem.Get File    ${CONFIG_ORIGINAL}
        Upload String To File    ${cfg}    ${CONFIG_PATH}
        Systemctl    restart    microshift-observability
    END
    Teardown Suite
