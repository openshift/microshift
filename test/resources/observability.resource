*** Settings ***
Documentation       Variables common to observability test suites


*** Variables ***
${PROMETHEUS_PORT}          9091
${LOKI_PORT}                3100
${PROM_EXPORTER_PORT}       8889
${LOCAL_METRICS_URL}        http://127.0.0.1:9090/metrics
