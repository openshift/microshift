# Observability: Scraping MicroShift and Optional Metric Exporters

MicroShift ships the `microshift-observability` RPM, which runs the Red Hat build of the
OpenTelemetry Collector as a host `systemd` service, plus three optional metric-exporter
RPMs. When the collector and one or more exporters are installed together, the collector
automatically scrapes each exporter's `/metrics` endpoint and forwards the data to the
configured OTLP backend — with no manual configuration.

This document is **contributor-facing**: it describes the `scrape.d` drop-in integration so
maintainers can keep the collector presets, the drop-in files, and the RPM packaging in
sync. Admin-facing usage of the OpenTelemetry Collector lives in **openshift-docs**
(assembly *"Using MicroShift Observability"*,
`microshift_running_apps/microshift-observability-service.adoc`).

## The `scrape.d` drop-in mechanism

The observability RPM installs a scrape-config drop-in for MicroShift's local
endpoint. Each optional metrics RPM installs its own drop-in into
`/etc/microshift/observability/scrape.d/`. All three collector presets
(`packaging/observability/opentelemetry-collector-{small,medium,large}.yaml`) load whatever
drop-ins are present at startup via a `prometheus` receiver:

```yaml
receivers:
  prometheus:
    config:
      scrape_config_files:
        - /etc/microshift/observability/scrape.d/*.yaml
processors:
  resourcedetection/system:
    detectors: [ "system" ]
    system:
      hostname_sources: [ "os" ]
service:
  pipelines:
    metrics/scrape:
      receivers: [ prometheus ]
      processors: [ resourcedetection/system, batch ]
      exporters: [ otlp ]
```

Because the receiver globs only the files that are on disk, the design satisfies three
cases with no extra wiring:

- **Observability without optional exporters** — only the local MicroShift target
  is configured. If the endpoint is disabled or unavailable, its scrape reports `up=0`;
  this does not prevent other receivers or MicroShift from running.
- **Observability with a subset of exporters** — the local target and only the
  installed optional exporters are scraped.
- **Exporters without observability** — the drop-in files exist on disk but are never read,
  so no scraping occurs (the exporter workload still deploys and serves its endpoint).

### Package ownership

The `microshift-observability` subpackage owns the `scrape.d/` **directory** and
`microshift.yaml`; each optional metrics subpackage owns its **drop-in file**.
All drop-in files are packaged as `%config(noreplace)` (see
`packaging/rpm/microshift.spec`).

| Exporter | RPM package | Drop-in file (`.../scrape.d/`) | Service (ns `openshift-monitoring`) | Port(s) | Exposes | Auth |
|---|---|---|---|---|---|---|
| MicroShift | `microshift-observability` | `microshift.yaml` | host process (no Kubernetes Service) | 9090 (`http`, loopback only) | certificate status, pending renewal, collection freshness | none (local access) |
| node-exporter | `microshift-metrics-node-exporter` | `node-exporter.yaml` | `node-exporter` | 9100 (`https`) | host CPU/memory/disk/network/OS metrics | kube-rbac-proxy static cert |
| kube-state-metrics | `microshift-metrics-kube-state` | `kube-state-metrics.yaml` | `kube-state-metrics` | 8443 (`https-main`), 9443 (`https-self`) | Kubernetes object-state metrics | kube-rbac-proxy static cert |
| metrics-server | `microshift-metrics-server` | `metrics-server.yaml` | `metrics-server` | 443 (`https`) | metrics-server's own serving metrics | Kubernetes RBAC |

The exporter RPMs are `ExclusiveArch: x86_64 aarch64`.

### Authentication model

The optional-exporter drop-ins authenticate with the `openshift-observability-client` cert (issued from the
`admin-kubeconfig-signer` CA) against the service-CA-signed exporter endpoints.

- **kube-state-metrics** and **node-exporter** sit behind `kube-rbac-proxy` configured with
  static authorization — any client cert signed by the trusted CA is admitted, with no
  RBAC `SubjectAccessReview`. The observability client cert is sufficient.
- **metrics-server** delegates to the Kubernetes API for authorization, so it requires an
  RBAC rule. The observability `ClusterRole` therefore includes
  `nonResourceURLs: ["/metrics"]` — already shipped in
  `assets/optional/observability/02-cluster-role.yaml`. No additional RBAC is needed.

> **Note:** Scraping metrics-server's `/metrics` collects metrics-server's *own* serving
> metrics. Node/pod resource metrics for `kubectl top` are served through the aggregated
> Metrics API (`metrics.k8s.io`), **not** through this scrape.

### Local MicroShift target

`microshift.yaml` scrapes `http://127.0.0.1:9090/metrics` once per minute with a
five-second timeout. Enable the endpoint separately as described in
[Local Certificate Metrics](../user/howto_certificates.md#local-certificate-metrics).
Package installation does not enable metrics or restart MicroShift. No API
discovery, client certificate, CA bundle, or insecure TLS bypass is needed for
this HTTP loopback target.

All three presets add the OS hostname as the OTel resource attribute `host.name`
using the existing `resourcedetection/system` processor. Keep hostnames stable and
unique within the fleet, or add a customer-owned node identifier in the collection
pipeline. A backend must preserve resource identity: `127.0.0.1:9090` alone cannot
distinguish nodes. Prometheus remote-write exporters can use
`resource_to_telemetry_conversion.enabled: true` to retain `host_name` as a label.

For a custom port or IPv6 listener, edit the existing drop-in's target to match
`metrics.listenAddress` (for example, `'[::1]:9090'`) and restart
`microshift-observability`. To opt out of this scrape, keep the file with
`scrape_configs: []` and restart the collector; disabling the MicroShift listener
alone leaves an unavailable target. Avoid adding a second job for the same endpoint.

Upgrades preserve edited collector configuration and drop-ins. Existing custom
configurations must include the Prometheus receiver, the `scrape_config_files`
glob, and the `metrics/scrape` pipeline shown above, including host resource
detection. Merge these additions into the active configuration without replacing
customer exporters or credentials, then restart the collector. A customer-managed
host collector can use the same wiring and drop-in without the bundled service;
its distribution must include the Prometheus receiver and resource detection processor.

Certificate renewal does not introduce scrape credentials to refresh for this
target. Preparing renewal leaves active certificate series unchanged; activation
on MicroShift restart updates them on subsequent observations. Existing
authenticated receivers still depend on their own credentials. The bundled
collector retains its systemd dependency on MicroShift; it is not an independent
diagnostics service when MicroShift exits.

### Source-of-truth files

Keep the collector presets, the drop-ins, and the RPM packaging in sync with these:

- `packaging/observability/scrape.d/{microshift,node-exporter,kube-state-metrics,metrics-server}.yaml`
- `packaging/observability/opentelemetry-collector-{small,medium,large}.yaml`
- `assets/optional/{node-exporter,kube-state-metrics,metrics-server}/04-service.yaml`
- `assets/optional/observability/02-cluster-role.yaml`
- `packaging/rpm/microshift.spec` (subpackages and `scrape.d` file ownership)

End-to-end behavior is exercised by `test/suites/optional/observability.robot`
(USHIFT-7407), which enables the exporters alongside observability and asserts that
kube-state-metrics, node-exporter, metrics-server, and MicroShift certificate series
appear in the collector output and backend, with host identity.
