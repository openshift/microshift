# Contributing to build-machinery-go

build-machinery-go ships reusable GNU Make fragments consumed by many OpenShift
Go projects via `go mod vendor`. Changes here have broad downstream impact.

Read [ARCHITECTURE.md](ARCHITECTURE.md) for how the make stacks and verification
model work.

## Related guidelines

This repository does not define org-wide OpenShift or Go conventions. Use these
instead:

| Topic | Where |
|-------|-------|
| Control plane code conventions, testing, PR process, review expectations | [openshift/service-ca-operator/CONTRIBUTING.md](https://github.com/openshift/service-ca-operator/blob/main/CONTRIBUTING.md) |
| OpenShift CI / Prow / Jira integration | [docs.ci.openshift.org](https://docs.ci.openshift.org/) |
| Commit signature verification | [OpenShift contribution policy](https://docs.google.com/document/d/1184EPSGunUkcSQYUK8T4a6iyawwi6f2zxdbB2jtG9nQ/edit?usp=sharing) |
| AI code review configuration | [openshift/coderabbit](https://github.com/openshift/coderabbit) |

For reviews, reach out via [OWNERS](OWNERS) or the control plane Slack channels
listed in the service-ca-operator contributing guide.

## Development workflow

1. Fork the repo and clone your fork.
2. Create a feature branch from `master`.
3. Make your changes. When makefile behavior changes, add or update examples under
   `make/examples/`.
4. Run `make verify` locally before pushing.
5. Open a PR against `openshift/build-machinery-go:master`.

Functional changes that regenerate logs should use two commits when applicable:
code first, then `update generated` for `*.log` files only.

## Verification

This repo validates makefile fragments through **checked-in log snapshots**, not
unit tests. See [ARCHITECTURE.md §5](ARCHITECTURE.md#5-verification-model) for
details.

- Run `make update` after changing `make/targets/` or examples, then commit the
  regenerated `*.log` files.
- Never hand-edit `*.example.mk.help.log` or `Makefile.test.log`.

### Matching CI output

Local `make update` output may differ across distributions. To match CI, run
update in the same build root image as Prow (defined in
[`.ci-operator.yaml`](.ci-operator.yaml)):

```bash
podman run -it --rm --pull=always \
  -v "$(pwd)":/go/src/$(go list -m) \
  --workdir=/go/src/$(go list -m) \
  registry.ci.openshift.org/openshift/release:rhel-9-release-golang-1.23-openshift-4.19 \
  make update
```

## Make fragment changes

- Add new behavior in `make/targets/`, not in entry files (`golang.mk`,
  `default.mk`, `operator.mk`).
- Keep backward compatibility unless a breaking change is explicitly agreed.
- Place complex shell logic in `scripts/` — follow the
  [shell styleguide](https://google.github.io/styleguide/shellguide.html).
- Do not add Go dependencies without justification in the PR description.

## Pull requests

Follow the linked control plane and OpenShift CI guidelines for Jira titles
(`CNTRLPLANE-XXXX:` or `NO-JIRA:`), `/lgtm`, `/approve`, `/verified`, and Prow
retests.

Repository-specific expectations:

- `make verify` must pass in CI.
- Changes to `make/targets/` must include updated examples and regenerated logs.
- Breaking fragment interface changes need maintainer agreement and a migration
  note for downstream repos.
- Do not modify `OWNERS` or `OWNERS_ALIASES` without explicit direction.

For makefile-only changes, `/verified by ci` is typically sufficient when
`make verify` passes.

## Areas requiring extra care

- Entry file changes (`golang.mk`, `default.mk`, `operator.mk`) affect every
  downstream repo on that stack.
- Target module interface changes (variables, target names, defaults) must stay
  backward compatible or document migration.
- Log normalization sed filters in the root `Makefile` must not hide real
  behavior changes.
