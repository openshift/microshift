*** Settings ***
Documentation       Keywords common to observability test suites


*** Keywords ***
Start Prometheus Server
    [Documentation]    Start a Prometheus Server on port 9092
    [Arguments]    ${port}=9092
    Local Command Should Work    bash -x ./bin/manage_prometheus.sh start ${port}

Stop Prometheus Server
    [Documentation]    Stop the Prometheus Server
    [Arguments]    ${port}=9092
    Local Command Should Work    bash -x ./bin/manage_prometheus.sh stop ${port}

Start Loki Server
    [Documentation]    Start a Loki Server on port 3100
    [Arguments]    ${port}=3100
    Local Command Should Work    bash -x ./bin/manage_loki.sh start ${port}

Stop Loki Server
    [Documentation]    Stop the Loki Server
    [Arguments]    ${port}=3100
    Local Command Should Work    bash -x ./bin/manage_loki.sh stop ${port}
