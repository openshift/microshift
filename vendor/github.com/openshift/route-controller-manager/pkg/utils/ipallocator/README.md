# ipallocator

Minimal in-memory IP range allocator for assigning IPs from a CIDR block.

## Provenance

This code is copied from two packages in **k8s.io/kubernetes v1.35.0**:

- `k8s.io/kubernetes/pkg/registry/core/service/allocator` — bitmap allocator
- `k8s.io/kubernetes/pkg/registry/core/service/ipallocator` — IP range wrapper

Only the code paths used by the IngressIP controller (`pkg/route/ingressip/`) are
included. The following features from the original were removed:

- Metrics recording (`metricsRecorderInterface`, Prometheus counters)
- Dry-run support (`DryRun()`, `dryRunRange`)
- Snapshot/restore (`Snapshot()`, `Restore()`, `NewFromSnapshot`)
- IPFamily tracking (`IPFamily()`, `api.IPFamily` dependency)
- Factory pattern (`New()` with `AllocatorWithOffsetFactory`)
- Unused methods: `ForEach()`, `Destroy()`, `CIDR()`, `Used()`, `EnableMetrics()`

This eliminates the dependency on `k8s.io/kubernetes` (the monorepo).
IP math uses `k8s.io/utils/net` which is a standalone module.
