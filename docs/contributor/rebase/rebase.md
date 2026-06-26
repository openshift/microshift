# Rebasing MicroShift

## Upstream Relationship

MicroShift vendors OpenShift source code without modification. A given MicroShift release is built from the same content as the corresponding OpenShift release, adding a small amount of glue logic. MicroShift follows OCP's versioning scheme.

This means MicroShift must periodically sync with OpenShift to pick up new features, bug fixes, and security patches. This process is called a "rebase."

## What Gets Rebased

A rebase updates the following:

| Artifact | Source | Script Step |
|----------|--------|-------------|
| Go dependencies | Vendored k8s/OCP source in `vendor/`, `deps/` | `rebase.sh go.mod` + `make vendor` |
| Component manifests | OpenShift operator manifests in `assets/` | `rebase.sh manifests` |
| Container image references | OCP release image digests in `pkg/release/` | `rebase.sh images` |
| Build files | Kubernetes/OCP version in `Makefile.*.var` | `rebase.sh buildfiles` |
| Changelog | Git commit tracking in `scripts/auto-rebase/` | `rebase.sh changelog` |

## Rebase Automation

A Prow job runs `scripts/auto-rebase/rebase.sh` on weekdays at 5 AM UTC. It creates a PR automatically against `openshift/microshift` via the `microshift-rebase-script` GitHub App.

Rebases on `main` track the next minor OCP version (nightly images). Release branches (`release-X.Y`) receive z-stream rebases targeting the corresponding OCP z-stream releases.

The automation flow:

1. `rebase_job_entrypoint.sh` — gathers CI credentials and release image tags from ConfigMaps
2. `rebase.py` — manages GitHub interaction (branch creation, PR, comments, cleanup)
3. `rebase.sh` — performs the actual rebase operations (download, go.mod, images, manifests, buildfiles)

## Optional Component Rebases

Several components are not part of the OCP release image and have their own rebase scripts and cadences:

| Component | Script | Source |
|-----------|--------|--------|
| LVMS | `rebase-lvms.sh` | LVMS operator bundle (Red Hat catalog or CPaaS) |
| Gateway API | `rebase_gateway_api.sh` | OSSM (OpenShift Service Mesh) operator |
| SR-IOV | `rebase_sriov.sh` | SR-IOV network operator bundle |
| AI Model Serving | `rebase_ai_model_serving.sh` | RHOAI KServe operator |
| Cert Manager | `rebase_cert_manager.sh` | Cert Manager operator |

Each script reverse-engineers the upstream operator bundle to extract manifests, images, RBAC, and deployments into `assets/optional/`.

## Dependency Management

MicroShift's `go.mod` uses replace directives with hint comments that tell the rebase script where to source each module:

```text
// from $COMPONENT           — picks from the component's go.mod
// staging kubernetes         — picks from openshift/kubernetes staging directory
// release $COMPONENT         — picks from the OpenShift release image
// override                   — keep current version, do not replace
```

All `k8s.io/*` packages are replaced with local `./deps/github.com/openshift/kubernetes` staging directories, ensuring consistency with a pinned Kubernetes version.

## Further Reading

- [Rebase Procedure](./procedure.md) — step-by-step instructions for running a rebase
- [Rebase Guidelines](./guidelines.md) — conventions, common pitfalls, and tribal knowledge
