# Observability: Scraping the Optional Metric Exporters

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

Each optional metrics RPM installs a Prometheus scrape-config drop-in into
`/etc/microshift/observability/scrape.d/`. All three collector presets
(`packaging/observability/opentelemetry-collector-{small,medium,large}.yaml`) load whatever
drop-ins are present at startup via a `prometheus` receiver:

```yaml
receivers:
  prometheus:
    config:
      scrape_config_files:
        - /etc/microshift/observability/scrape.d/*.yaml
service:
  pipelines:
    metrics/scrape:
      receivers: [ prometheus ]
      processors: [ batch ]
      exporters: [ otlp ]
```

Because the receiver globs only the files that are on disk, the design satisfies three
cases with no extra wiring:

- **Observability without metrics** — `scrape.d/` is empty; the receiver loads no configs
  and reports no error.
- **Observability with a subset of exporters** — only the installed exporters are scraped.
- **Exporters without observability** — the drop-in files exist on disk but are never read;
  no side effects.

### Package ownership

The `microshift-observability` subpackage owns the `scrape.d/` **directory**; each metrics
subpackage owns its **drop-in file** as `%config(noreplace)` (see
`packaging/rpm/microshift.spec`).

| Exporter | RPM package | Drop-in file (`.../scrape.d/`) | Service (ns `openshift-monitoring`) | Port(s) | Exposes | Auth |
|---|---|---|---|---|---|---|
| node-exporter | `microshift-metrics-node-exporter` | `node-exporter.yaml` | `node-exporter` | 9100 (`https`) | host CPU/memory/disk/network/OS metrics | kube-rbac-proxy static cert |
| kube-state-metrics | `microshift-metrics-kube-state` | `kube-state-metrics.yaml` | `kube-state-metrics` | 8443 (`https-main`), 9443 (`https-self`) | Kubernetes object-state metrics | kube-rbac-proxy static cert |
| metrics-server | `microshift-metrics-server` | `metrics-server.yaml` | `metrics-server` | 443 (`https`) | metrics-server's own serving metrics | Kubernetes RBAC |

The exporter RPMs are `ExclusiveArch: x86_64 aarch64`.

### Authentication model

The drop-ins authenticate with the `openshift-observability-client` cert (issued from the
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

### Source-of-truth files

Keep the collector presets, the drop-ins, and the RPM packaging in sync with these:

- `packaging/observability/scrape.d/{node-exporter,kube-state-metrics,metrics-server}.yaml`
- `packaging/observability/opentelemetry-collector-{small,medium,large}.yaml`
- `assets/optional/{node-exporter,kube-state-metrics,metrics-server}/04-service.yaml`
- `assets/optional/observability/02-cluster-role.yaml`
- `packaging/rpm/microshift.spec` (subpackages and `scrape.d` file ownership)

End-to-end behavior is exercised by `test/suites/optional/observability.robot`
(USHIFT-7407), which enables the exporters alongside observability and asserts that
kube-state-metrics and node-exporter series appear in the collector output and that
metrics-server is scraped without `403` errors.
