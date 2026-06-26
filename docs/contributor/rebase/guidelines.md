# Rebase Guidelines

Conventions, common pitfalls, and tribal knowledge for rebasing MicroShift. For the step-by-step procedure, see [procedure.md](./procedure.md). For the high-level overview, see [rebase.md](./rebase.md).

## When Manual Intervention Is Needed

The nightly automation handles z-stream rebases within a minor version. Manual intervention is needed for:

- **Minor version bumps** (e.g. 4.17 → 4.18) — often introduces new APIs, manifest changes, or dependency conflicts that the automation cannot resolve
- **LVMS rebases** — separate cadence from OCP, run via `rebase-lvms.sh`
- **Failed nightly rebases** — the automation creates a PR with the failure output; debugging requires manual steps
- **Embedded component config drift** — the non-automated step of comparing embedded component configs with upstream templates

## Commit Conventions

Rebase PRs follow a specific commit structure. Each step produces its own commit:

```text
update go.mod
update vendoring
update component images
update embedded manifests
update LVMS images
update LVMS manifests
update buildfiles
```

The PR description should include the changelog showing what upstream commits are being pulled in.

## What to Review in a Rebase PR

- **Manifest diffs** (`assets/`): look for new/removed resources, changed fields, namespace changes
- **New or removed APIs**: check `pkg/config/` for new configuration parameters that need defaults
- **Image reference changes** (`pkg/release/`): verify digests match the target release
- **go.mod changes**: verify replace directives resolved correctly, no unexpected version bumps
- **Build file changes** (`Makefile.*.var`): version strings should match the target release

## Handling go.mod Conflicts

When the rebase script's hint system cannot resolve module version conflicts:

1. Identify which components disagree on the module version (check error output from `go mod tidy`)
2. Determine which component's version is authoritative (usually `openshift/kubernetes` for k8s modules)
3. If the conflict cannot be resolved by hints, add a patch to `scripts/auto-rebase/rebase_patches/`

There are two patch directories:

- `scripts/auto-rebase/rebase_patches/` — applied to vendored Go code via `make patch-deps` during the vendor step
- `scripts/auto-rebase/manifests_patches/` — applied to upstream manifests in `assets/` after `rebase.sh manifests`

Both should be minimal and well-documented with the reason for the patch.

## Common Pitfalls

### yq Version

MicroShift requires `yq` from [mikefarah/yq](https://github.com/mikefarah/yq), not the Python-based `yq`. The wrong version produces silent failures in manifest processing. Check the [CI config](https://github.com/openshift/release/blob/master/ci-operator/config/openshift/microshift/openshift-microshift-main__periodics.yaml) for the authoritative version.

### Pull Secrets

Rebasing against nightly or CI images requires a CI registry pull secret. Rebasing against released images requires an OpenShift pull secret. The secrets are different and can be combined:

```bash
jq -c -s '.[0] * .[1]' app.ci_registry.json openshift-pull-secret.json > ~/.pull-secret.json
```

Missing or expired pull secrets cause silent download failures in the rebase script.

### Stale Staging Directory

The `_output/staging/` directory caches downloaded release content. When switching between target releases, stale content can cause incorrect rebases. Clean it with `rm -rf _output/staging/` before retargeting.

### Architecture-Specific Image Mismatches

The rebase takes both AMD64 and ARM64 release images as input. These must be from the same logical release. Mixing images from different builds (e.g. different nightly timestamps) causes digest mismatches in `pkg/release/`.

## LVMS Rebase Specifics

- LVMS is not integrated with the OCP release image — it comes from a separate operator bundle
- The script reverse-engineers the OLM bundle: it extracts RBAC, deployments, webhooks, and services from the ClusterServiceVersion
- Namespace patching is required because OLM-based installation uses operator-managed namespaces, while MicroShift installs manifests directly
- The LVMS rebase can be run independently of the OCP rebase

## Embedded Component Config Drift

This is the one rebase step that is not automated. When rebasing to a new minor version, compare the configuration of embedded components with their upstream templates:

- **Kubelet config**: compare `writeConfig(...)` in `pkg/node/kubelet.go` with MCO's template in `_output/staging/machine-config-operator/templates/master/01-master-kubelet/_base/files/kubelet.yaml`
- **OVN-K config**: compare startup flags in `pkg/controllers/` with the OVN-K operator's default configuration
- **Other components**: check for new flags, removed options, or changed defaults in each embedded component

## Testing Rebase Changes

### Local Dry Run

```bash
AMD64_RELEASE=registry.ci.openshift.org/ocp/release:<tag> \
ARM64_RELEASE=registry.ci.openshift.org/ocp-arm64/release-arm64:<tag> \
LVMS_RELEASE=registry.redhat.io/lvms4/lvms-operator-bundle:<tag> \
DRY_RUN=y \
TOKEN=ghp_... \
ORG=openshift \
REPO=microshift \
./scripts/auto-rebase/rebase.py
```

### Fork-Based Testing

Run against a personal fork to verify branch creation and PR workflow:

```bash
TOKEN=ghp_... \
ORG=<your-github-username> \
REPO=microshift \
AMD64_RELEASE=... \
ARM64_RELEASE=... \
LVMS_RELEASE=... \
./scripts/auto-rebase/rebase.py
```

The `microshift-rebase-script` GitHub App must be [installed](https://github.com/apps/microshift-rebase-script/installations/new) on your fork for push and PR creation to work.

### CI Rehearsal

Create a dummy PR in `openshift/release` that changes `ci-operator/step-registry/openshift/microshift/rebase/openshift-microshift-rebase-commands.sh` to target your fork or enable `DRY_RUN`. This lets you test the rebase in the production CI environment without affecting the main repo.
