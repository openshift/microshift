# Architecture Overview

build-machinery-go provides reusable GNU Make fragments and helper scripts that
OpenShift Go repositories vendor and include in their own `Makefile`. This
document describes how the pieces fit together. Update it when the makefile
layout or verification model changes.

## 1. Project Structure

See also [AGENTS.md](AGENTS.md) for contributor-oriented rules. Layout:

```
build-machinery-go/
├── Makefile                 # Meta-verification of example makefiles and logs
├── make/
│   ├── golang.mk            # Entry: pure Go projects
│   ├── default.mk           # Entry: OpenShift Go (+ images, bindata, codegen)
│   ├── operator.mk          # Entry: OpenShift operators
│   ├── *.example.mk         # Copy-paste starting points for downstream repos
│   ├── *.example.mk.help.log  # Checked-in `make help` output (audit trail)
│   ├── targets/             # Composable make modules
│   └── examples/            # Integration tests for makefile fragments
├── scripts/                 # Shell helpers invoked by make targets
└── commitchecker/           # Small Go binary using golang.mk (dogfooding)
```

## 2. High-Level System Diagram

This repository is a **library**, not a deployed service. Component repos consume
it at build time:

```
┌─────────────────────────┐
│  OpenShift component    │
│  repo (operator, etc.)  │
└───────────┬─────────────┘
            │ go mod vendor
            ▼
┌─────────────────────────┐     include      ┌──────────────────────────┐
│ vendor/.../build-       │ ───────────────► │  Component Makefile      │
│ machinery-go/make/*.mk  │                  │  (build, test, verify,   │
└─────────────────────────┘                  │   images, codegen, ...)  │
            ▲                                └──────────────────────────┘
            │ verify via examples
┌───────────┴─────────────┐
│  build-machinery-go     │
│  (this repo)            │
│  make verify / update   │
└─────────────────────────┘
```

Data flow: makefile fragments define targets; component repos run those targets
locally and in CI. This repo validates fragment behavior through checked-in log
output from example makefiles.

## 3. Core Components

### 3.1. Make fragment stacks

Three predefined stacks layer on top of each other. Downstream repos include
exactly one entry file from their vendored copy:

| Stack    | Entry file         | Extends   | Typical targets                          |
|----------|--------------------|-----------|------------------------------------------|
| Golang   | `make/golang.mk`   | —         | `build`, `test-unit`, `verify-gofmt`     |
| Default  | `make/default.mk`  | Golang    | + `images`, `verify-codegen`, `bindata`  |
| Operator | `make/operator.mk` | Default   | + `test-operator-integration`, profiles  |

**Include chain:**

```
operator.mk
  └── default.mk
        ├── targets/openshift/deps.mk
        ├── targets/openshift/images.mk
        ├── targets/openshift/bindata.mk
        ├── targets/openshift/codegen.mk
        └── golang.mk
              ├── targets/help.mk
              └── targets/golang/*.mk

operator.mk also includes:
  └── targets/openshift/operator/*.mk
```

Entry files are thin wrappers that `include` modules from `make/targets/`. New
behavior belongs in `make/targets/`, not duplicated in entry files.

### 3.2. Target modules (`make/targets/`)

| Directory                     | Purpose                                           |
|-------------------------------|---------------------------------------------------|
| `targets/golang/`             | Build, test, fmt, vet, version, vulncheck         |
| `targets/openshift/`          | Images, bindata, codegen, deps, kustomize, yq, rpm |
| `targets/openshift/operator/` | Release, telepresence, profile manifests, MOM       |

Each `*.mk` file defines related targets and their `verify-*` / `update-*`
counterparts where applicable.

### 3.3. Scripts (`scripts/`)

Shell scripts hold logic too complex for inline make recipes:

| Script                         | Used by                            |
|--------------------------------|------------------------------------|
| `update-deps.sh`               | Dependency update targets          |
| `test-operator-integration.sh` | Operator integration test target     |
| `run-telepresence.sh`          | Telepresence development workflow  |
| `vulncheck.sh`                 | Vulnerability scanning target        |

### 3.4. commitchecker

A minimal Go package that includes `golang.mk` from the parent directory. It
dogfoods the Golang stack to confirm fragments still work for real Go builds.
See [`commitchecker/README.md`](commitchecker/README.md) for downstream CI usage.

## 4. Downstream Consumption

Component repos vendor this module and include one entry file:

```makefile
include $(addprefix vendor/github.com/openshift/build-machinery-go/make/, \
    default.mk \
)
```

Paths resolve relative to the included file via
`$(dir $(lastword $(MAKEFILE_LIST)))`, so fragments work regardless of vendored
path depth.

For a starting point, copy the matching `*.example.mk` into the component repo's
`Makefile` and adjust `GO_BUILD_PACKAGES`, image names, and codegen paths.

## 5. Verification Model

This repo does not build OpenShift operators. It verifies makefile fragments
through:

1. **Example makefiles** (`make/*.example.mk`) — `make help` output captured in
   checked-in `*.help.log` files.
2. **Integration examples** (`make/examples/*/Makefile.test`) — exercise
   specific targets (codegen, profile manifests, golang version checks).
   Output captured in `Makefile.test.log` files.
3. **Root `Makefile`** — runs all examples and diffs output via
   `make verify` / `make update`.

Log files are the audit trail: any change to makefile behavior must be visible
in regenerated log diffs.

## 6. External Integrations

| Integration | Purpose | How |
|-------------|---------|-----|
| OpenShift release images | CI build root; reproducible `make update` | `registry.ci.openshift.org/openshift/release` (see `.ci-operator.yaml`) |
| `govulncheck` | Dependency vulnerability scanning | Invoked by `scripts/vulncheck.sh` via `targets/golang/vulncheck.mk` |
| Codegen / image tooling | Bindata, CRD schema, controller-gen, imagebuilder | Referenced by `targets/openshift/*.mk`; run in downstream repos |
| Telepresence | Local operator development | `scripts/run-telepresence.sh` (operator stack only) |

Downstream repos may integrate additional external tools through their own
Makefile variables; this repo provides the make targets that invoke them.

## 7. Deployment & Infrastructure

**Distribution:** Published as the Go module
`github.com/openshift/build-machinery-go`. Downstream repos pin a version in
`go.mod` and copy fragments into `vendor/` via `go mod vendor`. There is no
runtime deployment of this repository itself.

**CI/CD:** OpenShift Prow via ci-operator. Build root image is defined in
[`.ci-operator.yaml`](.ci-operator.yaml) (`rhel-9-release-golang-1.23-openshift-4.19`).
The primary CI check is `make verify`, which diffs example makefile output
against checked-in logs.

**Infrastructure owned by this repo:** None. Make targets in downstream repos may
build container images, generate manifests, or interact with clusters, but that
machinery runs in consumer repositories, not here.

## 8. Security Considerations

| Area | Practice |
|------|----------|
| Dependency scanning | `vulncheck` target runs `govulncheck`; fails on module vulnerabilities |
| Vendor integrity | `verify-deps` / `update-deps` targets in Default stack validate dependency state in downstream repos |
| Script execution | Helper scripts use `bash -e` and clean up temp files (`trap` in `vulncheck.sh`) |
| Secrets | No credentials or cluster access in this repo; downstream targets that need `KUBECONFIG` run in component repos |
| Supply chain | Changes to makefile fragments are auditable via checked-in `*.log` diffs in PRs |

This repo does not implement authentication or authorization. Security-sensitive
operations (image pushes, cluster deploys) are gated by CI and credentials in
downstream repositories.

## 9. Development & Testing Environment

**Local setup:** Clone the repo and run `make verify`. See
[CONTRIBUTING.md](CONTRIBUTING.md) for the full workflow, including the
container command to match CI log output.

**Testing approach:**

| Layer | Mechanism |
|-------|-----------|
| Makefile fragments | `make/examples/*/Makefile.test` integration examples |
| Help output | `make/*.example.mk.help.log` snapshot tests |
| Golang stack | `commitchecker/` dogfooding build |

**Code quality:** `make verify` is the gate. Downstream stacks additionally
expose `verify-gofmt`, `verify-govet`, `verify-codegen`, and related targets.

## 10. Project Identification

| Field | Value |
|-------|-------|
| Project name | build-machinery-go |
| Repository | https://github.com/openshift/build-machinery-go |
| Module path | `github.com/openshift/build-machinery-go` |
| Maintainers | See [OWNERS](OWNERS) (`control-plane-approvers`, `jsafrane`, `sanchezl`) |

## 11. Glossary

| Term | Definition |
|------|------------|
| **Stack** | One of Golang, Default, or Operator entry-file layers |
| **Fragment** | A `*.mk` file included into a downstream `Makefile` |
| **Target module** | A composable `make/targets/**/*.mk` file defining related targets |
| **Log audit** | Checked-in `*.log` files that snapshot makefile output for `git diff` verification |
| **Downstream repo** | An OpenShift component repo that vendors and includes these make fragments |
