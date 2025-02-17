# Rebasing MicroShift

## Overview

MicroShift repackages a minimal core of OpenShift: A given MicroShift release is built from the same content as the corresponding OpenShift release, adding a small amount of "glue logic".

Therefore, on a high level a rebase of MicroShift onto a newer OpenShift release starts from a given OpenShift release image and involves the following steps:

* vendoring the source code of the embedded OpenShift components at the same commit as used in OpenShift,
* embedding the resource manifests of the hosted OpenShift components,
* updating the image references (digests) of the hosted OpenShift components, and
* updating build files (Makefile, Containerfile, ...).

This process is supported by the `scripts/auto-rebase/rebase.sh` script, which automates generating the changes for updating MicroShift to the given target OpenShift release.

The following describes the current rebase process in more detail.

## Process

### Prerequisites

Rebasing requires the following tools to be installed, which is already the case when running in a [MicroShift development environment](./devenv_setup.md):

* `git` >= 2.3
* `golang` >= 1.18
* `oc` (latest)
* `jq` >= 1.6
* `yq` >= 4.26

> There are multiple tools called `yq`, please make sure you install the one from [mikefarah](https://github.com/mikefarah/yq). See the [CI config](https://github.com/openshift/release/blob/master/ci-operator/config/openshift/microshift/openshift-microshift-main__periodics.yaml) (in the `dockerfile_literal` of `yq-cli` image) for the authoritative version and how to install it.

When rebasing onto an OpenShift version whose release image requires a pull secret, place that secret in `~/.pull-secret.json`.

> For rebasing onto nightlies and other versions on CI, see the [OpenShift CI docs](https://docs.ci.openshift.org/docs/how-tos/use-registries-in-build-farm/#how-do-i-log-in-to-pull-images-that-require-authentication) for how to obtain a pull secret for the CI registry. The gist is: Get a CLI login token from the [app.ci web console](https://console-openshift-console.apps.ci.l2s4.p1.openshiftapps.com/), use it to log in from the CLI (`oc login https://app.ci.openshift.org --token=<token>`), and download the registry pull-secret using `oc registry login --to=app.ci_registry.json`.
>
> For rebasing onto released OpenShift versions, use the regular [OpenShift pull secret](https://cloud.redhat.com/openshift/install).
>
> You can combine both pull secrets using `jq -c -s '.[0] * .[1]' app.ci_registry.json openshift-pull-secret.json > ~/.pull-secret.json`.

Finally. git clone your personal fork of microshift and `cd` into it.

### A Note on Rebasing MicroShift's CSI Plugin

The Logical Volume Manager Service is not integrated with the ocp release image and must be passed explicitly to the rebase script as its 4th argument (including the sub-command). Images can be found at [Red Hat Container Catalog](https://catalog.redhat.com/software/containers/lvms4/lvms-operator-bundle/63972de4d8764b33ec4dbf79?tag=v4.12.0-4&architecture=amd64&push_date=1673885582000&container-tabs=gti).

### Fully automatic rebasing of OCP components

The following command attempts a fully automatic rebase to a given target upstream OpenShift release. It is what is run nighly from CI and should work for most cases within a z-stream. It creates a new branch named after the target release, then runs the individual steps described in the following sections, including creating the respective commits.

```shell
./scripts/auto-rebase/rebase.sh to quay.io/openshift-release-dev/ocp-release:4.10.25-x86_64 quay.io/openshift-release-dev/ocp-release:4.10.25-aarch64
```

### Manual rebasing of OCP components

#### Downloading the target OpenShift release

Run the following to download the OpenShift release to rebase to, specifying the target release images for _both_ Intel and Arm architectures, e.g.:

```shell
./scripts/auto-rebase/rebase.sh download quay.io/openshift-release-dev/ocp-release:4.10.25-x86_64 quay.io/openshift-release-dev/ocp-release:4.10.25-aarch64
```

This will create a directory `_output/staging`, download the specified release images' metadata (`release_{amd64,arm64}.json`) and manifests, git-clone (only) the repos of the embedded components and the operators of the loaded components, and check out the commit used by that OpenShift release.

#### Updating the changelog

The rebase process tracks the git commits used for the embedded code and images to build a changelog of the updates in each rebase. Having the changelog in the pull request makes it easier for reviewers to understand what changes are being pulled in as part of the rebase.

```shell
./scripts/auto-rebase/rebase.sh changelog
```

#### Rebasing the go.mod file and vendoring

In MicroShift's `go.mod` file, we only explicitly add the `require` directives needed by MicroShift itself (marked with the comment `// microshift`) whereas we let `go mod tidy` figure out the direct and indirect requirements for the embedded components. For the rebase, the focus is therefore on the `replace` directives.

When resolving version mismatches between modules used by different embedded components, the general heuristic is to start from a minimal subset of ReplaceDirectives from `o/k` (only those `go mod tidy` actually tries to find) to ensure these are consistent, then add etcd's and route-controller-manager's dependencies.

The `rebase.sh` script automates updating the modulepaths (e.g. rewriting local paths like `./staging` to the global paths) and versions, but it needs hints which component's version to pick. We encode these hints as keywords in comments in MicroShift's `go.mod` file. Each replacement modulepath and versions are picked as follows:

```shell
// from $COMPONENT                  picks from the $COMPONENT's go.mod
// staging kubernetes               picks the version in openshift/kubernetes's staging directory
// release $COMPONENT [via $REPO]   picks from the OpenShift's release image
// override                         keep the current version / do not replace
```

The optional `via $REPO` argument to `release` can be used when the git repository name does not match the component image name.

Run the following to rebase the `go.mod`:

```shell
./scripts/auto-rebase/rebase.sh go.mod
go mod tidy
git add go.mod go.sum
git commit -m "update go.mod"
```

As we're vendoring from multiple OpenShift components, there may be a situation in which we need to pick a module version for one component that is not completely aligned with another component's version (like we've had when still vendoring `openshift-apiserver`). In this case, it may be necessary to resolve the conflict through patches to the vendored modules. These patches would then be stored in `scripts/auto-rebase/rebase_patches`.

To update the vendoring run:

```shell
make vendor
git add vendor
git commit -m "update vendoring"
```

#### Rebasing the hosted component images

To update the image references for the hosted components of the MicroShift release (incl. the pause image for CRI-O), run:

```shell
./scripts/auto-rebase/rebase.sh images
git add pkg/release
git commit -m "update component images"
```

#### Rebasing the hosted component manifests

The next step is to update the manifests of embedded and hosted components in the asset directory:

```shell
./scripts/auto-rebase/rebase.sh manifests
```

For each component, this performs the following high-level steps:

1. From that component's Operator, copy the operand manifests / manifest templates into the component's asset directory. Typically, we just replace the complete directory, so removed manifests are handled correctly, but in some cases we need to preserve MicroShift-specific assets.
2. Render operand manifest templates like the operator would, e.g. filling in names, namespaces, and labels.
3. Make MicroShift-specific changes, e.g. changing the replica count to 1.
4. Replace values that MicroShift needs to fill in at runtime with the corresponding templating var. As `yq`, which is used for transforming manifests, trips over Go templating vars, this step neeeds to be done last.

Each step consistes of zero or more transformations. These transformations should remain relatively stable, but at least when rebasing to a new minor version of OpenShift, the output produced by the components Operator and that of the rebase script should be compared to make the necessary updates.

#### Rebasing the embedded component configs

The step isn't automated at all yet, which is to compare whether the config parameters of embedded components changed, for example the kubelet configuration in `writeConfig(...)` of `pkg/node/kubelet.go` with OpenShift MCO's template (which the `rebase.sh` script downloads into `_output/staging/machine-config-operator/templates/master/01-master-kubelet/_base/files/kubelet.yaml`).

#### Rebasing the build files

The final step is to update build files which include following:
* `Makefile.kube_git.var` contains information about Kubernetes version:
  major, minor, full version, commit, and git tree state.
  > Note: Kubernetes version is sourced from `openshift/kubernetes`'
  > `openshift-hack/images/hyperkube/Dockerfile.rhel` file.
  >
  > If OCP/Kubernetes rebase process changes, MicroShift rebase tooling may require an update.
* `Makefile.version.x86_64.var` and `Makefile.version.aarch64.var` are updated with the OCP version.

```shell
./scripts/auto-rebase/rebase.sh buildfiles
```

When updating to a new minor version of OpenShift, you may also need to update other locations, for example:

* [microshift.spec](/packaging/rpm/microshift.spec)
  * the Golang version (`golang_version`)

Commit the changes:

```shell
git add Makefile* packaging
git commit -m "update buildfiles"
```

### Fully Automated Update of LVMS

The following command attempts a fully automatic rebase to a given target LVMS release.  It creates a new branch named after the LVMS release, then runs the individual steps described in the following sections, including creating the respective commits.

```shell
./scripts/auto-rebase/rebase.sh lvms-to registry.redhat.io/lvms4/lvms-operator-bundle:[TAG || DIGEST]
```

### Manual Update of LVMS

#### Downloading the target LVMS release

Run the following to download the LVMS release to update to, specifying the multi-arch target release image, e.g.:

```shell
./scripts/auto-rebase/rebase.sh lvms-download registry.redhat.io/lvms4/lvms-operator-bundle:v4.12
```

This will create a directory `_output/staging`, download the operator bundle for the specified LVMS release.

#### Updating the LVMS component images

To update the image references for LVMS, run:

```shell
./scripts/auto-rebase/rebase.sh lvms-images
git add pkg/release
git commit -m "update LVMS images"
```

#### Updating the LVMS manifests

To update the manifests for LVMS, run:

```shell
./scripts/auto-rebase/rebase.sh lvms-manifests
git add assets
git commit -m "update LVMS manifests"
```

## Rebasing a Community OKD Release

The rebase script can also be used to build a community version of MicroShift that does not require access to internal image registries. To build the community version, rebase against an [OKD image found here](https://origin-release.ci.openshift.org/#4-stable) by following the same rebase script, and downloading the OKD release image [like so](#downloading-the-target-openshift-release) instead of an OpenShift release. This build is not supported or maintained and there is no guarantee that OKD releases will continue to mirror OpenShift releases.

## Rebase Prow Job

MicroShift's Go dependencies, manifests, and images are kept in sync with OpenShift by a [rebase](https://steps.ci.openshift.org/workflow/openshift-microshift-rebase) Prow Job that runs on weekdays at night (5 AM UTC) which executes a rebase procedure and creates a Pull Request if needed.
Rebase Prow Job is set up in release repository but scripts are kept in the Microshift repository in `scripts/auto-rebase/` directory.
This allows us to keep all logic in one place for easier testing and development, and to minimize problems due to synchronizing PR merges across different repositories (first make changes to Prow Job configuration to expose additional files, resources, etc., then make changes to the scripts to utilize them).

Following scripts are revelant for the rebase job (explained in detail below):
- `rebase_job_entrypoint.sh`
- `rebase.py`
- `rebase.sh`

### Credentials

Rebase job requires following credentials:
- **Pull Secret** used to pull OpenShift release images.
  - It belongs to `system-serviceaccount-microshift-image-puller` which is configured [here](https://github.com/openshift/release/blob/master/clusters/app.ci/registry-access/microshift/admin_manifest.yaml).
- **GitHub App ID and private key** used to interact with `openshift/microshift` repository and GitHub API.
  - They're obtained from GH App's Settings page - private key must be generated and stored in a safe location.
    - For more information about GitHub Apps see [GitHub Docs: About apps](https://docs.github.com/en/developers/apps/getting-started-with-apps/about-apps)
  - App ID and key are used to obtain an Installation Access Token (similar to Personal Access Token) which can be used to interact with remote git repository and GitHub API.
    - Installation Access Token is only valid for one hour, so it needs to be generated on each job run.
    - See [GitHub documentation about authentication as an installation](https://docs.github.com/en/developers/apps/building-github-apps/authenticating-with-github-apps#authenticating-as-an-installation) for more information.


Credentials are stored in [CI's Vault](https://vault.ci.openshift.org/) and made available to the job's `openshift-microshift-rebase` step by [this configuration](https://github.com/openshift/release/blob/d50f86ece447e22011acf0bdddc7ff4af7124953/ci-operator/step-registry/openshift/microshift/rebase/openshift-microshift-rebase-ref.yaml#L16-L22).
See [OpenShift CI Docs: Adding a New Secret to CI](https://docs.ci.openshift.org/docs/how-tos/adding-a-new-secret-to-ci/) for more details.

### Getting latest release tags

Rebase Prow Job is configured to leverage CI's ability to provide the job with latest nightly tag inside a ConfigMap that scripts later access.
[Here](https://github.com/openshift/release/blob/d50f86ece447e22011acf0bdddc7ff4af7124953/ci-operator/config/openshift/microshift/openshift-microshift-main__periodics.yaml#L95-L106) is a configuration to populate the ConfigMaps.
[rebase_job_entrypoint.sh](https://github.com/openshift/microshift/blob/3255867882b6b7da6c411449c52ae50c610a9dc7/scripts/auto-rebase/rebase_job_entrypoint.sh#L15-L20) accesses these ConfigMaps to obtain tags and create URIs which `rebase.sh` uses to download release images.

### Setting up the App as trusted

Installing an App for specific repository does not make it part of the organisation which means that PRs created by the App are not tested automatically requiring team members to apply `ok-to-test` label.

To make the App trusted and have presubmit jobs executed without any intervention it needs to be added to Prow's Plugin Config for the repository.
This configuration is performed in `openshift/release` repository in file `core-services/prow/02_config/ORG/REPOSITORY/_pluginconfig.yaml`
([openshift/microshift example](https://github.com/openshift/release/blob/70057044c4088c19fb89fa051a8d51e21a49e4e3/core-services/prow/02_config/openshift/microshift/_pluginconfig.yaml#L54-L58)):
```yaml
triggers:
- repos:
  - openshift/microshift
  trusted_apps:
  - microshift-rebase-script
```

### rebase_job_entrypoint.sh

Entrypoint of the Rebase Prow Job.
It expects to be executed in the job's container as it looks for files and objects defined in the job's configuration in (`openshift/release` repository):
pull secret file, latest release ConfigMaps existing in CI's cluster, GH App ID and key files.
Can serve as an example of what arguments `rebase.py` expects.

### rebase.sh

Primary consumer of release image references.
Script responsible for updating MicroShift's manifests, image references, and go.mod references based on contents of release images.
It's executed by `rebase.py` to capture output and exit code.

It also creates a git branch which name consists of prefix `rebase-`, version and stream (e.g. `4.13.0-0.nightly`), and creation datetimes of AMD and ARM nightly releases (e.g. `_amd64-2023-01-31-072358_arm64-2023-01-31-233557`).

### rebase.py

Python script responsible for interacting with remote `openshift/microshift` repository and communicating with GitHub API.
Uses GH App's ID and key to generate IAT (Installation Access Token) which is like Personal Access Token except its lifetime is 1 hour.

It also executes `rebase.sh` to capture its output and exit code.
In case of error, it will save `rebase.sh` output to a file, commit it together with any modified files, push changes, and create a PR to inform about the failure in visible way.

`rebase.py` interacts with remote git repository:
- Creates new, temporary `git remote`: `https://x-access-token:{token}@github.com/openshift/microshift`.
- Fetches remote references (branches) to check if rebase branch already exists, compare it with local, and decide if it needs update.
- (Force) pushes local branch to `openshift/microshift`

`rebase.py` interacts with GitHub API:
- Lists existing PRs to find one matching local branch name, so it'll force push changes and comment under PR that it was refreshed (can happen if no new nightlies were produced, but another PR was merged on microshift repository, so what happens is job is rebasing the rebase PR on newer main branch)
- Creates a PR if needed.
- Posts a PR comment which extra information that is important, but shouldn't be part of PR's description.
- Cleans up branches of closed PRs (branches of merges PRs are automatically deleted, otherwise we need to clean the up).


### Testing changes to rebase automation

To reduce a need of Pull Request synchronization between repositories, keep logic together, and allow for easier testing and developing, all scripts related to rebasing are residing in `scripts/auto-rebase/` directory in this repository.

[Job's logic is to only execute](https://github.com/openshift/release/blob/1976df80259f0357c0842700a70a64433988892a/ci-operator/step-registry/openshift/microshift/rebase/openshift-microshift-rebase-commands.sh)
`scripts/auto-rebase/rebase_job_entrypoint.sh` which gathers necessary arguments and passes them to `rebase.py` and is implemented in a way expecting to be executed in the job's container but can be a guidance on what arguments provide to `rebase.py` locally.


#### Getting release image build references

Rebase procedure expects references to AMD64 and ARM64 OpenShift release images, and LVM Storage (LVMS) Operator bundle image.
OpenShift release images can be obtained from Release Status pages: [AMD64](https://amd64.ocp.releases.ci.openshift.org/) and [ARM64](https://arm64.ocp.releases.ci.openshift.org/) - navigate to section with nightly image builds for version that is currently worked on and pick latest approved for both architectures.
LVMS Operator bundle image can be obtained from [Red Hat's catalog](https://catalog.redhat.com/software/containers/lvms4/lvms-operator-bundle/63972de4d8764b33ec4dbf79) - tag can be just appended to following URI: `registry.redhat.io/lvms4/lvms-operator-bundle:`.
These references are passed to `rebase.py` using `AMD64_RELEASE`, `ARM64_RELEASE`, and `LVMS_RELEASE` environment variables, for example:
```
AMD64_RELEASE=registry.ci.openshift.org/ocp/release:4.13.0-0.nightly-2023-01-27-165107 \
ARM64_RELEASE=registry.ci.openshift.org/ocp-arm64/release-arm64:4.13.0-0.nightly-arm64-2023-01-30-010253 \
LVMS_RELEASE=registry.redhat.io/lvms4/lvms-operator-bundle:v4.12 \
./scripts/auto-rebase/rebase.py
```

#### Testing locally

For testing `rebase.py` locally, following env vars can be useful:
- `TOKEN` expects a GitHub Personal Access Token which can be generated [here](https://github.com/settings/tokens).
   Use it instead of `APP_ID` and `KEY`.
- `DRY_RUN` instructs script to not make any changes on the repo (i.e. git push, create PR, post comment, etc.) - instead actions are logged and script continues.
- `BASE_BRANCH` forces script to diff results against different branch than what was checked out when script started running
   (useful when testing changes exist on local branch that is not `main` - otherwise, script would want to create a PR with base being branch does not exists on remote).

```shell
TOKEN=ghp_... \
ORG=openshift \
DRY_RUN=y \
REPO=microshift \
AMD64_RELEASE=registry.ci.openshift.org/ocp/release:4.13.0-0.nightly-2023-01-27-165107 \
ARM64_RELEASE=registry.ci.openshift.org/ocp-arm64/release-arm64:4.13.0-0.nightly-arm64-2023-01-30-010253 \
LVMS_RELEASE=registry.redhat.io/lvms4/lvms-operator-bundle:v4.12 \
./scripts/auto-rebase/rebase.py
```

Script also can be ran against a private fork in non-dry run which is helpful to verify that the communication with remote repository is as expected and so is created PR.
In such case `ORG` env var needs to be set to GitHub username:

```shell
TOKEN=ghp_... \
ORG=USER_NAME \
REPO=microshift \
AMD64_RELEASE=registry.ci.openshift.org/ocp/release:4.13.0-0.nightly-2023-01-27-165107 \
ARM64_RELEASE=registry.ci.openshift.org/ocp-arm64/release-arm64:4.13.0-0.nightly-arm64-2023-01-30-010253 \
LVMS_RELEASE=registry.redhat.io/lvms4/lvms-operator-bundle:v4.12 \
./scripts/auto-rebase/rebase.py
```

#### Testing in CI

> Note: Rehearsing Rebase Prow Job without dry run can result in force pushing rebase branch, creation of rebase PR, and deleting stale rebase branches in openshift/microshift.

To test changes in context of "production" (CI Prow Job) environment it's recommended to first set and export one of two environment variables either in `openshift-microshift-rebase-commands.sh` or `rebase_job_entrypoint.sh`:
  - `ORG=GITHUB_USERNAME` to make the Job target specific fork of `microshift` like [pmtk/microshift](https://github.com/pmtk/microshift)
     > This requires [installing](https://github.com/apps/microshift-rebase-script/installations/new) the [microshift-rebase-script](https://github.com/apps/microshift-rebase-script) for the fork (this will allow the job to push branches and create PRs)
  - `DRY_RUN=y` to not push branches or create PRs on `$ORG/microshift` - job will just log what it would do and continue the execution

Then, create a dummy PR in [openshift/release](https://github.com/openshift/release) repository for rehearsing the rebase job that switches from `openshift/microshift` `main` branch to `org/microshift` for testing.

Example of `ci-operator/step-registry/openshift/microshift/rebase/openshift-microshift-rebase-commands.sh` (tweaked [PR](https://github.com/openshift/release/pull/35875/files):
```bash
#!/bin/bash

# These will be picked up in rebase_job_entrypoint.sh as well
export ORG=pmtk
# export DRY_RUN=y

git remote add TEST https://github.com/${ORG}/microshift.git
git fetch TEST
git switch --track TEST/csi-rebase-script

./scripts/auto-rebase/rebase_job_entrypoint.sh

git diff csi-rebase-script
```
