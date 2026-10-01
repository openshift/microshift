# CA Consolidation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Reduce MicroShift's CA count from 12 to 6 by restructuring `certSetup()` and updating all callers, with an atomic backup-and-regenerate migration path for upgrades.

**Architecture:** Four stacked commits, each compilable independently. Commit 1 replaces path helpers in `certinfo.go`; Commit 2 restructures `certSetup()` and adds `migrateLegacyCertLayout()`; Commits 3–4 update the two controllers that reference old CA paths.

**Tech Stack:** Go 1.21+, `pkg/util/cryptomaterial/certchains` builder framework (already in repo), `os.Rename` for atomic migration.

## Global Constraints

- All 4 commits must compile: `go build ./pkg/... ./vendor/...` must pass after each.
- No Robot Framework tests in scope — Go changes only.
- Leaf certificate CN/O fields, validity periods, and key sizes are **unchanged**.
- `certchains` builder framework API is **not modified**.
- Work in worktree: `/home/eslutsky/dev/microshift-dev/repos/microshift/.worktrees/ocpstrat-2900-cert-consolidation`
- Run all commands from inside that worktree path.
- Design spec: `docs/superpowers/specs/2026-10-01-ca-consolidation-design.md`

---

## File Map

| File | Commit | Change |
|------|--------|--------|
| `pkg/util/cryptomaterial/certinfo.go` | 1 | Add 3 new CA dir helpers + `KASServingCertDir` + migration sentinel; re-root all leaf helpers; delete 12 old root-CA helpers |
| `pkg/cmd/init.go` | 2 | Add `migrateLegacyCertLayout()`; restructure `certSetup()` from 9 signers to 6; update `WithCABundle` and `initKubeconfigs()` |
| `pkg/controllers/kube-apiserver.go` | 3 | Update `--etcd-cafile`, `--kubelet-certificate-authority`, `--tls-cert-file`; collapse `namedCerts` 3→1; fix `discoverEtcdServers` `TrustedCAFile` |
| `pkg/controllers/etcd.go` | 4 | Update `getEtcdClient` `TrustedCAFile` |

---

## Task 1: Commit 1 — `certinfo.go` path helpers

**Files:**
- Modify: `pkg/util/cryptomaterial/certinfo.go`

**Interfaces:**
- Produces: `ClientCADir(certsDir)`, `ServingCADir(certsDir)`, `PeerCADir(certsDir)`, `KASServingCertDir(certsDir)`, `legacyKubeControlPlaneSignerDir(certsDir)` (unexported); all existing leaf helpers re-rooted.

- [ ] **Step 1: Replace the entire section of old root-CA and leaf helpers**

Replace everything from `func KubeControlPlaneSignerCertDir` through `func KubeAPIServerServiceNetworkServingCertDir` (lines 49–157 in the current file) with the new helpers below. Leave the bundle-path functions at the bottom (`TotalClientCABundlePath`, `UltimateTrustBundlePath`, etc.) **unchanged**.

New content for the replaced section:

```go
// ClientCADir returns the path to the consolidated client CA directory.
func ClientCADir(certsDir string) string { return filepath.Join(certsDir, "client-ca") }

// ServingCADir returns the path to the consolidated serving CA directory.
func ServingCADir(certsDir string) string { return filepath.Join(certsDir, "serving-ca") }

// PeerCADir returns the path to the peer (etcd) CA directory.
func PeerCADir(certsDir string) string { return filepath.Join(certsDir, "peer-ca") }

// legacyKubeControlPlaneSignerDir returns the old CA path used only for migration detection.
func legacyKubeControlPlaneSignerDir(certsDir string) string {
	return filepath.Join(certsDir, "kube-control-plane-signer")
}

// Client CA leaf directories (all under client-ca/).
func KubeSchedulerClientCertDir(certsDir string) string {
	return filepath.Join(ClientCADir(certsDir), "kube-scheduler")
}
func KubeControllerManagerClientCertDir(certsDir string) string {
	return filepath.Join(ClientCADir(certsDir), "kube-controller-manager")
}
func KubeAPIServerToKubeletClientCertDir(certsDir string) string {
	return filepath.Join(ClientCADir(certsDir), "kube-apiserver-to-kubelet-client")
}
func MetricsServerKubeletClientCertDir(certsDir string) string {
	return filepath.Join(ClientCADir(certsDir), "metrics-server-kubelet-client")
}
func AdminKubeconfigClientCertDir(certsDir string) string {
	return filepath.Join(ClientCADir(certsDir), "admin-kubeconfig-client")
}
func KubeletClientCertDir(certsDir string) string {
	return filepath.Join(ClientCADir(certsDir), "kubelet-client")
}

// Serving CA leaf directories (all under serving-ca/).
func KASServingCertDir(certsDir string) string {
	return filepath.Join(ServingCADir(certsDir), "kube-apiserver-serving")
}
func KubeletServingCertDir(certsDir string) string {
	return filepath.Join(ServingCADir(certsDir), "kubelet-server")
}

// Unchanged CA directories.
func ServiceCADir(certsDir string) string { return filepath.Join(certsDir, "service-ca") }
func RouteControllerManagerServingCertDir(certsDir string) string {
	return filepath.Join(ServiceCADir(certsDir), "route-controller-manager-serving")
}
func IngressCADir(certsDir string) string { return filepath.Join(certsDir, "ingress-ca") }
func AggregatorSignerDir(certsDir string) string {
	return filepath.Join(certsDir, "aggregator-signer")
}
func AggregatorClientCertDir(certsDir string) string {
	return filepath.Join(AggregatorSignerDir(certsDir), "aggregator-client")
}

// Peer CA leaf directories (all under peer-ca/).
func EtcdAPIServerClientCertDir(certsDir string) string {
	return filepath.Join(PeerCADir(certsDir), "apiserver-etcd-client")
}
func EtcdPeerCertDir(certsDir string) string {
	return filepath.Join(PeerCADir(certsDir), "etcd-peer")
}
func EtcdServingCertDir(certsDir string) string {
	return filepath.Join(PeerCADir(certsDir), "etcd-serving")
}
```

- [ ] **Step 2: Verify build fails with expected errors**

```bash
cd /home/eslutsky/dev/microshift-dev/repos/microshift/.worktrees/ocpstrat-2900-cert-consolidation
go build ./pkg/... 2>&1 | grep -v "^#" | head -40
```

Expected: errors referencing deleted function names (`KubeControlPlaneSignerCertDir`, `EtcdSignerDir`, `KubeAPIServerExternalSigner`, etc.) in `init.go` and `kube-apiserver.go`. These are fixed in later commits — this step confirms the deletions are visible.

- [ ] **Step 3: Confirm no other callers of deleted helpers outside pkg/cmd and pkg/controllers**

```bash
cd /home/eslutsky/dev/microshift-dev/repos/microshift/.worktrees/ocpstrat-2900-cert-consolidation
grep -rn "KubeControlPlaneSignerCertDir\|KubeAPIServerToKubeletSignerCertDir\|AdminKubeconfigSignerDir\|KubeletCSRSignerSignerCertDir\|CSRSignerCertDir\|EtcdSignerDir\|KubeAPIServerExternalSigner\|KubeAPIServerLocalhostSigner\|KubeAPIServerServiceNetworkSigner\|KubeAPIServerExternalServingCertDir\|KubeAPIServerLocalhostServingCertDir\|KubeAPIServerServiceNetworkServingCertDir" \
  --include="*.go" . | grep -v "pkg/cmd/init.go" | grep -v "pkg/controllers/kube-apiserver.go" | grep -v "_test.go"
```

Expected: no output (all remaining callers are in the two files addressed in Tasks 2 and 3).

If any other files appear, fix those callers before committing.

- [ ] **Step 4: Commit**

```bash
cd /home/eslutsky/dev/microshift-dev/repos/microshift/.worktrees/ocpstrat-2900-cert-consolidation
git add pkg/util/cryptomaterial/certinfo.go
git commit -m "$(cat <<'EOF'
OCPSTRAT-2900: add consolidated CA path helpers to certinfo.go

Replace old per-signer root helpers with ClientCADir, ServingCADir,
PeerCADir; re-root all leaf helpers under new CA directories; keep
unexported legacyKubeControlPlaneSignerDir for migration detection.

Co-Authored-By: Claude Sonnet 4.6 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 2: Commit 2 — `init.go` migration + `certSetup()` restructuring

**Files:**
- Modify: `pkg/cmd/init.go`

**Interfaces:**
- Consumes: `ClientCADir`, `ServingCADir`, `PeerCADir`, `KASServingCertDir`, `legacyKubeControlPlaneSignerDir` from Task 1.
- Produces: `migrateLegacyCertLayout(certsDir, version string) error` (called from `initCerts`); restructured `certSetup()` returning a 6-CA chain; updated `initKubeconfigs()`.

- [ ] **Step 1: Add `migrateLegacyCertLayout` and call it from `initCerts`**

Add the new function immediately before `func certSetup`:

```go
// migrateLegacyCertLayout detects the pre-consolidation cert layout and atomically
// backs it up so certSetup can regenerate a fresh 6-CA hierarchy.
func migrateLegacyCertLayout(certsDir, version string) error {
	if _, err := os.Stat(legacyKubeControlPlaneSignerDir(certsDir)); os.IsNotExist(err) {
		return nil // fresh install or already migrated
	}
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	backupDir := certsDir + ".backup." + version + "." + ts
	klog.Infof("Migrating cert layout: backing up %s → %s", certsDir, backupDir)
	return os.Rename(certsDir, backupDir)
}
```

Then, add a call at the top of `initCerts`, right before the `certSetup` call.
`version.Get().GitVersion` is the correct version string (from `pkg/version`);
add `"github.com/openshift/microshift/pkg/version"` to the imports if not already present.

```go
func initCerts(cfg *config.Config) (*certchains.CertificateChains, error) {
	certsDir := cryptomaterial.CertsDirectory(config.DataDir)
	if err := migrateLegacyCertLayout(certsDir, version.Get().GitVersion); err != nil {
		return nil, fmt.Errorf("cert layout migration failed: %w", err)
	}

	certChains, err := certSetup(cfg)
	// ... rest unchanged
```

- [ ] **Step 2: Replace the `certchains.NewCertificateChains(...)` block in `certSetup()`**

Replace everything from `certChains, err := certchains.NewCertificateChains(` through `.Complete()` (lines 101–378) with:

```go
	certChains, err := certchains.NewCertificateChains(
		//------------------------------
		// CLIENT CA
		//------------------------------
		certchains.NewCertificateSigner(
			"client-ca",
			cryptomaterial.ClientCADir(certsDir),
			alignValidity(cryptomaterial.LongLivedCertificateValidity),
		).WithClientCertificates(
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta:  certchains.CSRMeta{Name: "kube-controller-manager", Validity: alignValidity(cryptomaterial.ShortLivedCertificateValidity)},
				UserInfo: &user.DefaultInfo{Name: "system:kube-controller-manager"},
			},
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta:  certchains.CSRMeta{Name: "kube-scheduler", Validity: alignValidity(cryptomaterial.ShortLivedCertificateValidity)},
				UserInfo: &user.DefaultInfo{Name: "system:kube-scheduler"},
			},
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta:  certchains.CSRMeta{Name: "cluster-policy-controller", Validity: alignValidity(cryptomaterial.ShortLivedCertificateValidity)},
				UserInfo: &user.DefaultInfo{Name: "system:kube-controller-manager"},
			},
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta:  certchains.CSRMeta{Name: "route-controller-manager", Validity: alignValidity(cryptomaterial.ShortLivedCertificateValidity)},
				UserInfo: serviceaccount.UserInfo("openshift-route-controller-manager", "route-controller-manager-sa", ""),
			},
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta:  certchains.CSRMeta{Name: "kube-apiserver-to-kubelet-client", Validity: alignValidity(cryptomaterial.ShortLivedCertificateValidity)},
				UserInfo: &user.DefaultInfo{Name: "system:kube-apiserver", Groups: []string{"kube-master"}},
			},
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta:  certchains.CSRMeta{Name: "metrics-server-kubelet-client", Validity: alignValidity(cryptomaterial.ShortLivedCertificateValidity)},
				UserInfo: &user.DefaultInfo{Name: "system:metrics-server"},
			},
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta:  certchains.CSRMeta{Name: "admin-kubeconfig-client", Validity: alignValidity(cryptomaterial.LongLivedCertificateValidity)},
				UserInfo: &user.DefaultInfo{Name: "system:admin", Groups: []string{"system:masters"}},
			},
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta:  certchains.CSRMeta{Name: "openshift-observability-client", Validity: alignValidity(cryptomaterial.ShortLivedCertificateValidity)},
				UserInfo: &user.DefaultInfo{Name: "openshift-observability-client"},
			},
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta:  certchains.CSRMeta{Name: "kubelet-client", Validity: alignValidity(cryptomaterial.ShortLivedCertificateValidity)},
				UserInfo: &user.DefaultInfo{Name: "system:node:" + cfg.CanonicalNodeName(), Groups: []string{"system:nodes"}},
			},
		),

		//------------------------------
		// SERVING CA
		//------------------------------
		certchains.NewCertificateSigner(
			"serving-ca",
			cryptomaterial.ServingCADir(certsDir),
			alignValidity(cryptomaterial.LongLivedCertificateValidity),
		).WithServingCertificates(
			&certchains.ServingCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{Name: "kube-apiserver-serving", Validity: alignValidity(cryptomaterial.ShortLivedCertificateValidity)},
				Hostnames: append([]string{
					"kubernetes", "kubernetes.default", "kubernetes.default.svc",
					"kubernetes.default.svc.cluster.local",
					"openshift", "openshift.default", "openshift.default.svc",
					"openshift.default.svc.cluster.local",
					"api." + cfg.DNS.BaseDomain,
					"api-int." + cfg.DNS.BaseDomain,
					cfg.ApiServer.AdvertiseAddress,
					apiServerServiceIP.String(),
					"localhost", "127.0.0.1", "::1",
					cfg.Node.HostnameOverride,
					cfg.Node.NodeIP,
				}, cfg.ApiServer.SubjectAltNames...),
			},
			&certchains.ServingCertificateSigningRequestInfo{
				CSRMeta:   certchains.CSRMeta{Name: "kubelet-server", Validity: alignValidity(cryptomaterial.ShortLivedCertificateValidity)},
				Hostnames: []string{cfg.Node.HostnameOverride, cfg.Node.NodeIP},
			},
		),

		//------------------------------
		// PEER CA
		//------------------------------
		certchains.NewCertificateSigner(
			"peer-ca",
			cryptomaterial.PeerCADir(certsDir),
			alignValidity(cryptomaterial.LongLivedCertificateValidity),
		).WithClientCertificates(
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta:  certchains.CSRMeta{Name: "apiserver-etcd-client", Validity: alignValidity(cryptomaterial.LongLivedCertificateValidity)},
				UserInfo: &user.DefaultInfo{Name: "etcd", Groups: []string{"etcd"}},
			},
		).WithPeerCertificiates(
			&certchains.PeerCertificateSigningRequestInfo{
				CSRMeta:   certchains.CSRMeta{Name: "etcd-peer", Validity: alignValidity(cryptomaterial.LongLivedCertificateValidity)},
				UserInfo:  &user.DefaultInfo{Name: "system:etcd-peer:etcd-client", Groups: []string{"system:etcd-peers"}},
				Hostnames: []string{"localhost", cfg.Node.HostnameOverride, cfg.Node.NodeIP},
			},
			&certchains.PeerCertificateSigningRequestInfo{
				CSRMeta:   certchains.CSRMeta{Name: "etcd-serving", Validity: alignValidity(cryptomaterial.LongLivedCertificateValidity)},
				UserInfo:  &user.DefaultInfo{Name: "system:etcd-server:etcd-client", Groups: []string{"system:etcd-servers"}},
				Hostnames: []string{"localhost", cfg.Node.HostnameOverride, cfg.Node.NodeIP},
			},
		),

		//------------------------------
		// UNCHANGED CAs
		//------------------------------
		certchains.NewCertificateSigner(
			"aggregator-signer",
			cryptomaterial.AggregatorSignerDir(certsDir),
			alignValidity(cryptomaterial.ShortLivedCertificateValidity),
		).WithClientCertificates(
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta:  certchains.CSRMeta{Name: "aggregator-client", Validity: alignValidity(cryptomaterial.ShortLivedCertificateValidity)},
				UserInfo: &user.DefaultInfo{Name: "system:openshift-aggregator"},
			},
		),

		certchains.NewCertificateSigner(
			"service-ca",
			cryptomaterial.ServiceCADir(certsDir),
			alignValidity(cryptomaterial.LongLivedCertificateValidity),
		).WithServingCertificates(
			&certchains.ServingCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{Name: "route-controller-manager-serving", Validity: alignValidity(cryptomaterial.ShortLivedCertificateValidity)},
				Hostnames: []string{
					"route-controller-manager.openshift-route-controller-manager.svc",
					"route-controller-manager.openshift-route-controller-manager.svc.cluster.local",
				},
			},
		),

		certchains.NewCertificateSigner(
			"ingress-ca",
			cryptomaterial.IngressCADir(certsDir),
			alignValidity(cryptomaterial.LongLivedCertificateValidity),
		).WithServingCertificates(
			&certchains.ServingCertificateSigningRequestInfo{
				CSRMeta:   certchains.CSRMeta{Name: "router-default-serving", Validity: alignValidity(cryptomaterial.ShortLivedCertificateValidity)},
				Hostnames: []string{"*.apps." + cfg.DNS.BaseDomain},
			},
		),
	).WithCABundle(
		cryptomaterial.TotalClientCABundlePath(certsDir),
		[]string{"client-ca"},
	).WithCABundle(
		cryptomaterial.KubeletClientCAPath(certsDir),
		[]string{"client-ca"},
	).WithCABundle(
		cryptomaterial.KubeletServingCAPath(certsDir),
		[]string{"serving-ca"},
	).WithCABundle(
		cryptomaterial.ServiceAccountTokenCABundlePath(certsDir),
		[]string{"serving-ca"},
	).Complete()
```

- [ ] **Step 3: Update `initKubeconfigs()` — trust PEM loading**

Replace the two separate `externalTrustPEM` / `internalTrustPEM` reads (lines 404–411) with a single `servingCAPEM` read:

```go
// before:
//   externalTrustPEM, err := os.ReadFile(cryptomaterial.CACertPath(cryptomaterial.KubeAPIServerExternalSigner(...)))
//   internalTrustPEM, err := os.ReadFile(cryptomaterial.CACertPath(cryptomaterial.KubeAPIServerLocalhostSigner(...)))

// after:
certsDir := cryptomaterial.CertsDirectory(config.DataDir)
servingCAPEM, err := os.ReadFile(cryptomaterial.CACertPath(cryptomaterial.ServingCADir(certsDir)))
if err != nil {
    return fmt.Errorf("failed to load serving CA: %v", err)
}
```

Then replace every use of `externalTrustPEM` and `internalTrustPEM` in `initKubeconfigs()` with `servingCAPEM`.

- [ ] **Step 4: Update `initKubeconfigs()` — `GetCertKey` call sites**

Replace all 7 old-signer `GetCertKey` calls with the new signer name `"client-ca"`. The `kubelet-signer`/`kube-csr-signer` call also drops the middle path segment:

```go
// KCM (line ~505)
kcmCertPEM, kcmKeyPEM, err := certChains.GetCertKey("client-ca", "kube-controller-manager")

// scheduler (line ~519)
schedulerCertPEM, schedulerKeyPEM, err := certChains.GetCertKey("client-ca", "kube-scheduler")

// kubelet — note: 2 args now, not 3 (no sub-CA)
kubeletCertPEM, kubeletKeyPEM, err := certChains.GetCertKey("client-ca", "kubelet-client")

// cluster-policy-controller (line ~544)
clusterPolicyControllerCertPEM, clusterPolicyControllerKeyPEM, err := certChains.GetCertKey("client-ca", "cluster-policy-controller")

// route-controller-manager (line ~557)
routeControllerManagerCertPEM, routeControllerManagerKeyPEM, err := certChains.GetCertKey("client-ca", "route-controller-manager")

// observability client (line ~569)
observabilityClientCertPEM, observabilityClientKeyPEM, err := certChains.GetCertKey("client-ca", "openshift-observability-client")

// admin kubeconfig (line ~413)
adminKubeconfigCertPEM, adminKubeconfigKeyPEM, err := certChains.GetCertKey("client-ca", "admin-kubeconfig-client")
```

- [ ] **Step 5: Verify build passes (init.go only, controllers still broken)**

```bash
cd /home/eslutsky/dev/microshift-dev/repos/microshift/.worktrees/ocpstrat-2900-cert-consolidation
go build ./pkg/cmd/... 2>&1
```

Expected: no errors from `pkg/cmd/`. Errors may still come from `pkg/controllers/` — those are fixed in Tasks 3–4.

- [ ] **Step 6: Commit**

```bash
cd /home/eslutsky/dev/microshift-dev/repos/microshift/.worktrees/ocpstrat-2900-cert-consolidation
git add pkg/cmd/init.go
git commit -m "$(cat <<'EOF'
OCPSTRAT-2900: restructure certSetup() to 6-CA hierarchy, add migration

- Add migrateLegacyCertLayout(): atomically renames old certs/ to
  certs.backup.<version>.<timestamp>/ on first boot after upgrade
- Replace 9 old NewCertificateSigner calls with client-ca, serving-ca,
  peer-ca (3 KAS certs merged into 1 SAN-based kube-apiserver-serving)
- Update WithCABundle references to new signer names
- Update initKubeconfigs() GetCertKey calls; replace per-signer trust
  PEM reads with single serving-ca read

Co-Authored-By: Claude Sonnet 4.6 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 3: Commit 3 — `kube-apiserver.go` flag updates

**Files:**
- Modify: `pkg/controllers/kube-apiserver.go`

**Interfaces:**
- Consumes: `PeerCADir`, `KASServingCertDir`, `KubeletServingCAPath` from Task 1.

- [ ] **Step 1: Update variables at the top of `configure()`**

In `configure()` (around line 97), the old variable assignments reference deleted helpers. Replace with:

```go
certsDir := cryptomaterial.CertsDirectory(config.DataDir)
// remove: kubeCSRSignerDir, kubeletClientDir (re-derive below)
// remove: serviceNetworkServingCertDir, servingCert, servingKey (replaced by KASServingCertDir)
kubeletClientDir := cryptomaterial.KubeAPIServerToKubeletClientCertDir(certsDir)
clientCABundlePath := cryptomaterial.TotalClientCABundlePath(certsDir)
aggregatorCAPath := cryptomaterial.CACertPath(cryptomaterial.AggregatorSignerDir(certsDir))
aggregatorClientCertDir := cryptomaterial.AggregatorClientCertDir(certsDir)
etcdClientCertDir := cryptomaterial.EtcdAPIServerClientCertDir(certsDir)
kasServingCertDir := cryptomaterial.KASServingCertDir(certsDir)
```

- [ ] **Step 2: Collapse `namedCerts` from 3 entries to 1**

Replace the entire `namedCerts` slice literal (lines ~120–136) with:

```go
namedCerts := []configv1.NamedCertificate{
    {
        CertInfo: configv1.CertInfo{
            CertFile: cryptomaterial.ServingCertPath(kasServingCertDir),
            KeyFile:  cryptomaterial.ServingKeyPath(kasServingCertDir),
        },
    },
}
```

The user-cert prepend loop below (`cfg.ApiServer.NamedCertificates`) is **unchanged**.

- [ ] **Step 3: Update the three flags in `APIServerArguments`**

```go
// was: "etcd-cafile": {cryptomaterial.CACertPath(cryptomaterial.EtcdSignerDir(certsDir))}
"etcd-cafile": {cryptomaterial.CACertPath(cryptomaterial.PeerCADir(certsDir))},

// was: "kubelet-certificate-authority": {cryptomaterial.CABundlePath(kubeCSRSignerDir)}
"kubelet-certificate-authority": {cryptomaterial.KubeletServingCAPath(certsDir)},

// was: "tls-cert-file": {servingCert}
"tls-cert-file": {cryptomaterial.ServingCertPath(kasServingCertDir)},

// was: "tls-private-key-file": {servingKey}
"tls-private-key-file": {cryptomaterial.ServingKeyPath(kasServingCertDir)},
```

`--etcd-certfile`, `--etcd-keyfile`, `--kubelet-client-certificate`, `--kubelet-client-key` are **unchanged** — they derive from `EtcdAPIServerClientCertDir` and `KubeAPIServerToKubeletClientCertDir`, both already re-rooted in Task 1.

- [ ] **Step 4: Update `discoverEtcdServers` TrustedCAFile (line ~433)**

```go
// was: TrustedCAFile: cryptomaterial.CACertPath(cryptomaterial.EtcdSignerDir(certsDir))
TrustedCAFile: cryptomaterial.CACertPath(cryptomaterial.PeerCADir(certsDir)),
```

- [ ] **Step 5: Verify build passes**

```bash
cd /home/eslutsky/dev/microshift-dev/repos/microshift/.worktrees/ocpstrat-2900-cert-consolidation
go build ./pkg/controllers/... 2>&1
```

Expected: errors only from `pkg/controllers/etcd.go` (fixed in Task 4). `kube-apiserver.go` compiles cleanly.

- [ ] **Step 6: Commit**

```bash
cd /home/eslutsky/dev/microshift-dev/repos/microshift/.worktrees/ocpstrat-2900-cert-consolidation
git add pkg/controllers/kube-apiserver.go
git commit -m "$(cat <<'EOF'
OCPSTRAT-2900: update kube-apiserver controller for 6-CA layout

- Collapse namedCerts from 3 serving cert entries to 1 (KASServingCertDir)
- Update --etcd-cafile, --tls-cert-file, --kubelet-certificate-authority
  to reference new CA paths
- Fix discoverEtcdServers TrustedCAFile: EtcdSignerDir → PeerCADir

Co-Authored-By: Claude Sonnet 4.6 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 4: Commit 4 — `etcd.go` TrustedCAFile + final build

**Files:**
- Modify: `pkg/controllers/etcd.go`

**Interfaces:**
- Consumes: `PeerCADir` from Task 1; `EtcdAPIServerClientCertDir` (re-rooted in Task 1).

- [ ] **Step 1: Update `getEtcdClient()` TrustedCAFile**

In `getEtcdClient()` (around line 236):

```go
// was:
TrustedCAFile: cryptomaterial.CACertPath(cryptomaterial.EtcdSignerDir(certsDir)),

// after:
TrustedCAFile: cryptomaterial.CACertPath(cryptomaterial.PeerCADir(certsDir)),
```

`CertFile` and `KeyFile` derive from `EtcdAPIServerClientCertDir(certsDir)`, which is already re-rooted under `PeerCADir` in Task 1 — no change needed there.

- [ ] **Step 2: Scan for any remaining old CA references**

```bash
cd /home/eslutsky/dev/microshift-dev/repos/microshift/.worktrees/ocpstrat-2900-cert-consolidation
grep -rn "EtcdSignerDir\|KubeControlPlaneSignerCertDir\|KubeAPIServerToKubeletSignerCertDir\|AdminKubeconfigSignerDir\|KubeletCSRSignerSignerCertDir\|CSRSignerCertDir\|KubeAPIServerExternalSigner\|KubeAPIServerLocalhostSigner\|KubeAPIServerServiceNetworkSigner" \
  --include="*.go" . | grep -v "_test.go"
```

Expected: no output. If any remain, fix them before committing.

- [ ] **Step 3: Full build of all packages**

```bash
cd /home/eslutsky/dev/microshift-dev/repos/microshift/.worktrees/ocpstrat-2900-cert-consolidation
go build ./pkg/... 2>&1
```

Expected: no errors.

- [ ] **Step 4: Run existing tests**

```bash
cd /home/eslutsky/dev/microshift-dev/repos/microshift/.worktrees/ocpstrat-2900-cert-consolidation
go test ./pkg/... 2>&1 | tail -30
```

Expected: all tests pass. If any test references old CA names (e.g. `"etcd-signer"`, `"kube-control-plane-signer"`), update the test to use the new name.

- [ ] **Step 5: Commit**

```bash
cd /home/eslutsky/dev/microshift-dev/repos/microshift/.worktrees/ocpstrat-2900-cert-consolidation
git add pkg/controllers/etcd.go
git commit -m "$(cat <<'EOF'
OCPSTRAT-2900: update etcd controller TrustedCAFile to peer-ca

getEtcdClient: EtcdSignerDir → PeerCADir for TrustedCAFile.
CertFile/KeyFile paths auto-updated via re-rooted EtcdAPIServerClientCertDir.

Co-Authored-By: Claude Sonnet 4.6 (1M context) <noreply@anthropic.com>
EOF
)"
```

- [ ] **Step 6: Push the branch**

```bash
cd /home/eslutsky/dev/microshift-dev/repos/microshift/.worktrees/ocpstrat-2900-cert-consolidation
git push
```

---

## Post-Implementation Checklist

- [ ] All 4 commits on branch, each independent and compilable
- [ ] `go build ./pkg/...` passes on the final commit
- [ ] `go test ./pkg/...` passes with no failures
- [ ] No references to deleted helpers remain in non-test Go files
- [ ] Branch pushed to fork for PR creation
