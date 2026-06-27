# OpenShift CI for MicroShift

MicroShift uses [OpenShift CI](https://docs.ci.openshift.org/) with [Prow](https://github.com/kubernetes/test-infra/tree/master/prow) for continuous integration. CI configuration lives in the [openshift/release](https://github.com/openshift/release) repository.

For the test infrastructure itself (scenarios, VMs, image blueprints, Robot Framework), see [test_harness.md](./test_harness.md).

## How CI Runs

OpenShift CI conditionally runs testing whenever a PR is opened or updated. Depending on the changes introduced, only unit and validation checks may be run. Changes that affect MicroShift compile-time or runtime (source code, build scripts, container image definitions) trigger full E2E test runs.

CI jobs appear as GitHub status checks on PRs and report pending/pass/fail state. Failed jobs block the PR from merging (blocks may be overridden by maintainers).

E2E jobs are executed in parallel, each provisioning its own isolated MicroShift VM. A complete PR test run takes approximately 45 minutes.

## CI Configuration

CI configuration files are in `openshift/release`:

| File | Purpose |
|------|---------|
| `ci-operator/config/openshift/microshift/openshift-microshift-main.yaml` | Presubmit (PR) job definitions |
| `ci-operator/config/openshift/microshift/openshift-microshift-main__periodics.yaml` | Periodic (nightly) job definitions |
| `ci-operator/step-registry/openshift/microshift/` | Step definitions for test workflows |

### Job Types

| Type | Trigger | What it tests |
|------|---------|--------------|
| **Presubmit** | Every PR update | Current source against presubmit image blueprints |
| **Periodic** | Nightly/weekly (cron) | Extended testing with periodic image blueprints |
| **Release** | Before release cuts | Brew RPM-sourced images, upgrade paths |
| **Rebase** | Weekdays at 5 AM UTC | Automated dependency sync with OpenShift |

### CI Operator Concepts

CI configuration uses [steps, chains, and workflows](https://docs.ci.openshift.org/docs/architecture/step-registry/) to define execution trees. Each test job specifies a workflow that orchestrates pre-test infrastructure setup, test execution, and post-test teardown.

Key CI Operator references:

- [CI Operator architecture](https://docs.ci.openshift.org/docs/architecture/ci-operator/)
- [Step registry](https://docs.ci.openshift.org/docs/architecture/step-registry/)
- [API type definitions](https://github.com/openshift/ci-tools/blob/master/pkg/api/types.go)

## Running Tests Locally

Tests run against a remote MicroShift host via SSH using Robot Framework. See [test_harness.md](./test_harness.md#running-tests) for setup and invocation.

## Debugging CI Failures

### Accessing Job Logs

CI jobs are identified by their `ci/prow/*` GitHub status prefix. Click the `Details` link on a failed status to view the combined build log.

Use `/retest` in PR comments to rerun all failed jobs. See the [Prow bot commands](https://github.com/kubernetes/test-infra/blob/master/prow/jobs.md#triggering-jobs-with-comments) for the full list.

### Accessing CI Cluster

> Only the PR author is authorized to access a job's OCP namespace.

OpenShift GitHub org members can access the CI cluster namespace for a job. The build log prints a URL to the OCP Console namespace near the top. Log in with Company SSO.

> Test runtime objects (namespace, pods, build images) are garbage collected after 1 hour from test completion.

### Examining Build Artifacts

To pull CI build images locally:

1. Log in to the CI cluster console and copy the login command
2. Run `oc login --token=...` and `oc registry login`
3. Pull the image: `podman pull registry.build0X.ci.openshift.org/[PROJECT]/pipeline:bin`

### Diagnostics Collected by CI

CI jobs automatically collect diagnostics from test VMs on failure:

- **SOS reports**: system diagnostics via the `sosreport` utility (stored in job artifacts under `artifacts/<test_name>/`)
- **PCP metrics**: performance data collected via `pcp-zeroconf` (stored in `artifacts/<test_name>/`)

SOS archives need to be downloaded and unpacked with `tar xf`. PCP data can be visualized with `pmstat -a .` (CLI) or `pmchart -a .` (GUI) after installing the `pcp` and `pcp-gui` packages.

## Secret Management

Secrets are stored in [Vault](https://vault.ci.openshift.org/) and injected into step definitions via the `credentials` field. See [CI docs on adding secrets](https://docs.ci.openshift.org/docs/how-tos/adding-a-new-secret-to-ci/).

Current secrets used by MicroShift CI:

- **Pull secret**: for pulling images from CI and release registries
- **GitHub App credentials**: App ID and private key for the rebase automation (`microshift-rebase-script`)

## Contributing CI Changes

CI configuration changes are submitted as PRs to [openshift/release](https://github.com/openshift/release). Key points:

1. Run `make update` before pushing to regenerate Prow job specs from CI Operator config
2. CI will execute configuration patches as a `rehearsal` job against the target branch
3. Merge privilege is managed via `OWNERS` files in the CI config directories
4. Use `/retest` to rerun rehearsal jobs

For full details, see the [CI contribution docs](https://docs.ci.openshift.org/docs/how-tos/contributing-openshift-release/).

## Useful Links

- [Prow dashboard for MicroShift](https://prow.ci.openshift.org/?repo=openshift%2Fmicroshift) — historical job performance
- [CI Operator reference](https://steps.ci.openshift.org/ci-operator-reference)
- [MicroShift CI config](https://github.com/openshift/release/tree/master/ci-operator/config/openshift/microshift)
- [MicroShift step registry](https://github.com/openshift/release/tree/master/ci-operator/step-registry/openshift/microshift)
- Slack: `coreos#forum-testplatform` (ping `@dptp-helpdesk` for CI issues)
- Slack: `coreos#announce-testplatform` (CI outages and status)
