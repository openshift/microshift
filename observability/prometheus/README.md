# MicroShift Prometheus exporter

A MicroShift Prometheus exporter to collect unique metrics from a MicroShift cluster

## Prerequisites
 - MicroShift installed and running
 - Go version >= 1.21.11

## Steps to run it
1. Run `go run .`
2. You can check exported metric on port `9090`.

   Example: `curl -s http://localhost:9090/metrics`
    ```
    # HELP microshift_info MicroShift info
    # TYPE microshift_info gauge
    microshift_info{buildDate="2024-06-14T21:18:09Z",gitCommit="c5a37dff79a7a4428e94f7da448569821cdc2970",gitVersion="4.16.0~rc.5"} 1
    # HELP microshift_version MicroShift version
    # TYPE microshift_version gauge
    microshift_version{level="major"} 4
    microshift_version{level="minor"} 16
    microshift_version{level="patch"} 0
    ```

## Grafana dashboard
You can get a [Grafana dashboard example here](grafana_dashboard.json):
![alt text](grafana_dashboard.png)
