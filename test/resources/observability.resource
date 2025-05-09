*** Settings ***
Documentation       Keywords common to observability test suites


*** Keywords ***
Start Prometheus Server
    [Documentation]    Start a Prometheus Server on port 9092
    Local Command Should Work    bash -x ./bin/manage_prometheus.sh start

Stop Prometheus Server
    [Documentation]    Stop the Prometheus Server
    Local Command Should Work    bash -x ./bin/manage_prometheus.sh stop

Start Loki Server
    [Documentation]    Start a Loki Server on port 3100
    Local Command Should Work    bash -x ./bin/manage_loki.sh start

Stop Loki Server
    [Documentation]    Stop the Loki Server
    Local Command Should Work    bash -x ./bin/manage_loki.sh stop
