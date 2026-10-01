# CA Consolidation Implementation Design

**Date:** 2026-10-01
**Jira:** OCPSTRAT-2900
**EP:** https://github.com/openshift/enhancements/pull/2107
**Branch:** ocpstrat-2900-cert-consolidation

---

## Goal

Reduce MicroShift's CA count from 12 (11 roots + 1 sub-CA) to 6 by restructuring
`certSetup()` and updating all callers. Leaf certificate identities, SANs (except
KAS serving consolidation), and validity periods are unchanged.

## Scope

- Go implementation only — no Robot Framework tests in this PR
- Four stacked commits (each must compile and pass existing tests)
- OCPSTRAT-2899 (PKI inventory) is already merged; no dependency gap

## Non-Goals

- Changing leaf certificate CN/O fields
- Changing certificate validity periods
- Modifying the `certchains` builder framework API
- Robot Framework integration tests (follow-up)
- `microshift certs` CLI changes (already handled by OCPSTRAT-2899 inventory)

---

## Target CA Structure

```
certs/
  client-ca/          ← replaces 5 CA objects (4 signer entries incl. two-level
  |                     kubelet-csr-signer-signer/kube-csr-signer pair)
  serving-ca/         ← replaces 3 KAS serving CAs; kubelet-server moves here too
  peer-ca/            ← renamed from etcd-signer
  aggregator-signer/  ← unchanged
  service-ca/         ← unchanged
  ingress-ca/         ← unchanged
  ca-bundle/          ← same filenames, updated content
```

### CA Bundle Changes

| File | Old content | New content |
|------|-------------|-------------|
| `ca-bundle/client-ca.crt` | 5 client CA certs | single `client-ca` cert |
| `ca-bundle/kubelet-ca.crt` | 5 client CA certs | single `client-ca` cert |
| `ca-bundle/kubelet-serving-ca.crt` | kubelet-signer + kube-csr-signer | single `serving-ca` cert |
| `ca-bundle/service-account-token-ca.crt` | 3 KAS serving CAs | single `serving-ca` cert |

### KAS Serving Consolidation

3 serving certs (external, localhost, service-network) → 1 `kube-apiserver-serving`
cert under `serving-ca/` with all SANs merged. The `dynamiccertificates` SNI
selection logic is unchanged — it always matches the single cert.

---

## Commit Plan

### Commit 1: `pkg/util/cryptomaterial/certinfo.go` — path helpers

**Add** three new CA directory helpers:

```go
func ClientCADir(certsDir string) string  { return filepath.Join(certsDir, "client-ca") }
func ServingCADir(certsDir string) string { return filepath.Join(certsDir, "serving-ca") }
func PeerCADir(certsDir string) string    { return filepath.Join(certsDir, "peer-ca") }
```

**Add** one new KAS serving helper (replaces three old serving cert helpers):

```go
func KASServingCertDir(certsDir string) string {
    return filepath.Join(ServingCADir(certsDir), "kube-apiserver-serving")
}
```

**Re-root** existing leaf helpers under new CA dirs (same function names, new paths):

| Helper | Old root | New root |
|--------|----------|----------|
| `KubeSchedulerClientCertDir` | `KubeControlPlaneSignerCertDir` | `ClientCADir` |
| `KubeControllerManagerClientCertDir` | `KubeControlPlaneSignerCertDir` | `ClientCADir` |
| `KubeAPIServerToKubeletClientCertDir` | `KubeAPIServerToKubeletSignerCertDir` | `ClientCADir` |
| `MetricsServerKubeletClientCertDir` | `KubeAPIServerToKubeletSignerCertDir` | `ClientCADir` |
| `AdminKubeconfigClientCertDir` | `AdminKubeconfigSignerDir` | `ClientCADir` |
| `KubeletClientCertDir` | `CSRSignerCertDir` | `ClientCADir` |
| `KubeletServingCertDir` | `CSRSignerCertDir` | `ServingCADir` |
| `EtcdAPIServerClientCertDir` | `EtcdSignerDir` | `PeerCADir` |
| `EtcdPeerCertDir` | `EtcdSignerDir` | `PeerCADir` |
| `EtcdServingCertDir` | `EtcdSignerDir` | `PeerCADir` |

**Add** unexported migration sentinel (used only in Commit 2):

```go
func legacyKubeControlPlaneSignerDir(certsDir string) string {
    return filepath.Join(certsDir, "kube-control-plane-signer")
}
```

**Delete** all old root-CA directory helpers:
`KubeControlPlaneSignerCertDir`, `KubeAPIServerToKubeletSignerCertDir`,
`AdminKubeconfigSignerDir`, `KubeletCSRSignerSignerCertDir`, `CSRSignerCertDir`,
`EtcdSignerDir`, `KubeAPIServerExternalSigner`, `KubeAPIServerLocalhostSigner`,
`KubeAPIServerServiceNetworkSigner`, `KubeAPIServerExternalServingCertDir`,
`KubeAPIServerLocalhostServingCertDir`, `KubeAPIServerServiceNetworkServingCertDir`.

---

### Commit 2: `pkg/cmd/init.go` — migration + certSetup() restructuring

**New function** `migrateLegacyCertLayout(certsDir, version string) error`:

```go
func migrateLegacyCertLayout(certsDir, version string) error {
    if _, err := os.Stat(legacyKubeControlPlaneSignerDir(certsDir)); os.IsNotExist(err) {
        return nil // fresh install or already migrated
    }
    ts := strconv.FormatInt(time.Now().Unix(), 10)
    backupDir := certsDir + ".backup." + version + "." + ts
    return os.Rename(certsDir, backupDir) // atomic on same filesystem
}
```

Called at the top of `initCerts()`, before `certSetup()`. Failure behaviour:
- `os.Rename` fails → `initCerts()` returns error, old `certs/` intact, MicroShift does not start
- `certSetup()` fails after successful rename → existing fresh-start fallback removes partial `certs/` and retries; if retry fails, MicroShift does not start and greenboot triggers rollback

**`certSetup()` restructuring** — replace 9 old `NewCertificateSigner` calls with 3 new + 3 unchanged:

```
client-ca  (LongLived)
  └─ WithClientCertificates: kube-controller-manager, kube-scheduler,
       cluster-policy-controller, route-controller-manager,
       kube-apiserver-to-kubelet-client, metrics-server-kubelet-client,
       admin-kubeconfig-client, openshift-observability-client, kubelet-client

serving-ca (LongLived)
  └─ WithServingCertificates: kube-apiserver-serving [all SANs merged]
  └─ WithServingCertificates: kubelet-server

peer-ca    (LongLived)
  └─ WithClientCertificates: apiserver-etcd-client
  └─ WithPeerCertificates:   etcd-peer, etcd-serving

aggregator-signer (ShortLived) ← unchanged
service-ca        (LongLived)  ← unchanged
ingress-ca        (LongLived)  ← unchanged
```

**`WithCABundle` updates:**

```go
.WithCABundle(TotalClientCABundlePath(certsDir),  []string{"client-ca"})
.WithCABundle(KubeletClientCAPath(certsDir),       []string{"client-ca"})
.WithCABundle(KubeletServingCAPath(certsDir),      []string{"serving-ca"})
.WithCABundle(ServiceAccountTokenCABundlePath(certsDir), []string{"serving-ca"})
```

**`initKubeconfigs()` call sites** — all `GetCertKey` calls that reference old signer names:

| Line | Old call | New call |
|------|----------|----------|
| `cfg.Ingress.ServingCertificate` | `GetCertKey("ingress-ca", "router-default-serving")` | unchanged |
| admin kubeconfig | `GetCertKey("admin-kubeconfig-signer", "admin-kubeconfig-client")` | `GetCertKey("client-ca", "admin-kubeconfig-client")` |
| KCM kubeconfig | `GetCertKey("kube-control-plane-signer", "kube-controller-manager")` | `GetCertKey("client-ca", "kube-controller-manager")` |
| scheduler kubeconfig | `GetCertKey("kube-control-plane-signer", "kube-scheduler")` | `GetCertKey("client-ca", "kube-scheduler")` |
| kubelet cert | `GetCertKey("kubelet-signer", "kube-csr-signer", "kubelet-client")` | `GetCertKey("client-ca", "kubelet-client")` |
| cluster-policy-controller | `GetCertKey("kube-control-plane-signer", "cluster-policy-controller")` | `GetCertKey("client-ca", "cluster-policy-controller")` |
| route-controller-manager | `GetCertKey("kube-control-plane-signer", "route-controller-manager")` | `GetCertKey("client-ca", "route-controller-manager")` |
| observability client | `GetCertKey("admin-kubeconfig-signer", "openshift-observability-client")` | `GetCertKey("client-ca", "openshift-observability-client")` |

Note: the kubelet call drops the `"kube-csr-signer"` middle arg because `kubelet-client` is now a direct child of `client-ca`, not a grandchild via a sub-CA.

---

### Commit 3: `pkg/controllers/kube-apiserver.go` — flag updates

| Flag | Old | New |
|------|-----|-----|
| `--etcd-cafile` | `CACertPath(EtcdSignerDir(certsDir))` | `CACertPath(PeerCADir(certsDir))` |
| `--etcd-certfile` | `ClientCertPath(EtcdAPIServerClientCertDir(certsDir))` | unchanged (helper re-rooted in Commit 1) |
| `--etcd-keyfile` | `ClientKeyPath(EtcdAPIServerClientCertDir(certsDir))` | unchanged |
| `--kubelet-certificate-authority` | `CABundlePath(CSRSignerCertDir(certsDir))` | `KubeletServingCAPath(certsDir)` |
| `--tls-cert-file` | `ServingCertPath(KubeAPIServerServiceNetworkServingCertDir(certsDir))` | `ServingCertPath(KASServingCertDir(certsDir))` |
| `--tls-private-key-file` | `ServingKeyPath(KubeAPIServerServiceNetworkServingCertDir(certsDir))` | `ServingKeyPath(KASServingCertDir(certsDir))` |

`namedCerts` collapses from 3 entries to 1:

```go
namedCerts := []configv1.NamedCertificate{{
    CertInfo: configv1.CertInfo{
        CertFile: cryptomaterial.ServingCertPath(cryptomaterial.KASServingCertDir(certsDir)),
        KeyFile:  cryptomaterial.ServingKeyPath(cryptomaterial.KASServingCertDir(certsDir)),
    },
}}
```

User-supplied `cfg.ApiServer.NamedCertificates` are prepended as before.

---

### Commit 4: `pkg/controllers/etcd.go` — TrustedCAFile

In `getEtcdClient()`:

```go
// before
TrustedCAFile: cryptomaterial.CACertPath(cryptomaterial.EtcdSignerDir(certsDir))
// after
TrustedCAFile: cryptomaterial.CACertPath(cryptomaterial.PeerCADir(certsDir))
```

`CertFile`/`KeyFile` derive from `EtcdAPIServerClientCertDir()` which is already
re-rooted under `PeerCADir` in Commit 1 — no further change needed.

Note: `kube-apiserver.go` has a second `EtcdSignerDir` reference (the etcd
health-check client TrustedCAFile at line 433). That is updated in Commit 3
alongside the other kube-apiserver changes, not here.

---

## Files Changed Summary

| File | Commit | Nature of change |
|------|--------|-----------------|
| `pkg/util/cryptomaterial/certinfo.go` | 1 | Add new helpers, re-root leaf helpers, delete old root helpers |
| `pkg/cmd/init.go` | 2 | New `migrateLegacyCertLayout()`, restructure `certSetup()`, update `initKubeconfigs()` |
| `pkg/controllers/kube-apiserver.go` | 3 | Flag path updates, namedCerts 3→1 |
| `pkg/controllers/etcd.go` | 4 | TrustedCAFile path update |

---

## Open Items

- Verify no other callers of deleted `certinfo.go` helpers exist outside `pkg/` (e.g. in `test/`)
- Confirm `kube-apiserver.go:433` TrustedCAFile reference is the only remaining `EtcdSignerDir` after Commit 1
- Check `pkg/components/metrics.go` — uses `MetricsServerKubeletClientCertDir` (re-rooted in Commit 1, should require no direct change)
