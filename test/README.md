# Automated Functional and Integration Tests

This directory includes tests for MicroShift that go beyond unit
tests. They exercise the system more fully and verify behaviors
end-to-end.

The tests are written using [Robot Framework](https://robotframework.org),
a test automation framework that separates the description of the test
from the implementation of the test.

The document is organized in the following sections:
* [Test Suites](#test-suites) run against an existing instance of MicroShift
* [Test Scenarios](#test-scenarios) provide tools for complex test environments
* [Troubleshooting](#troubleshooting) techniques for the test environments
* [CI Integration Scripts](#ci-integration-scripts) run as part of CI flow

## Test Suites

Groups of tests are saved to `.robot` files in the `suites/`
directory. The tests in a suite should have the same basic
prerequisites and should test related functionality.

### Setting Up

To run the tests, you need an existing host running the version of
MicroShift to be tested. Most suites expect the service to be running,
although some may restart the service as part of testing
behavior. Most of the test suites expect the firewall on the host to
allow remote connections to the MicroShift API.

The test tool uses a configuration file, `variables.yaml`, saved in
the same directory as this README.md file. The variables in the file
tell the test suites how to find the MicroShift host and connect to
it.

To create the variables file, copy `variables.yaml.example` and edit
the settings:

```
# USHIFT_HOST: MicroShift host, can be a name or IP
USHIFT_HOST: microshift
# USHIFT_USER: User to log into MicroShift's host
USHIFT_USER: microshift
# SSH Private key to use when logging into MicroShift's host.
# Unset this variable to use ssh agent.
SSH_PRIV_KEY: /home/microshift/.ssh/id_rsa
# SSH port to use when connecting to MicroShift's host.
SSH_PORT: 22
# API port, in case the connection is through a forwarded port.
# Defaults to whatever is in the kubeconfig file.
#API_PORT: 6443
```

`USHIFT_HOST` should be the host name or IP of the VM. The value must
match one of the `kubeconfig` files generated for the host because
many of the tests will try to download that file to connect to the API
remotely.

`USHIFT_USER` should be the username for logging in to `USHIFT_HOST`
remotely via ssh and have sudo permission without password.

`SSH_PRIV_KEY` should be an ssh key file to use for authenticating as
`USHIFT_USER`. The key must not require a password. To connect to
hosts using a key with a password, leave `SSH_PRIV_KEY` set to an
empty string and the tests will connect as `USHIFT_USER` and rely on
the ssh agent to provide the correct credentials.

`SSH_PORT` should be the port used for an ssh connection.

`API_PORT` should be set when connections are performed through a
forwarded port.

To ensure the router smoke test works properly on a Microshift host,
you need to install the `avahi-resolve-host-name` command, which is
essential for hostname resolution. You can install it using the following command:

```
sudo dnf install -y avahi-tools
```

### Running Tests

Use `run.sh` to run the tests. It will create a Python virtual
environment in the `_output` directory and install Robot Framework
automatically.

The `-h` option prints usage instructions:

```
$ ./test/run.sh -h
run.sh [-h] [-n] [-o output_dir] [-v venv_dir] [-i var_file] [test suite files]

Options:

  -h       Print this help text.
  -n       Dry-run, do not run the tests.
  -o DIR   The output directory.
  -v DIR   The venv directory.
  -i PATH  The variables file.
```

#### Running a Single Suite

By default, all of the test suites will be run. To run a subset,
specify the filenames as arguments on the command line.

```
$ ./test/run.sh suites/standard/show-config.robot
```

> The test suite file names should be relative to the `test` directory.

#### Running a Subset of Tests

To run a single test, or subset of tests, use the `robot` command's
features for filtering the tests. Use `--` to separate arguments to
`run.sh` from arguments to be passed unaltered to `robot`. The input
files must also be specified explicitly in all of these cases.

To run tests with names matching a pattern, use the `-t` option:

```
$ ./test/run.sh -- -t '*No Mode*' suites/standard/show-config.robot
```

To run tests with specific tags, use the `-i` option:

```
$ ./test/run.sh -- -i etcd suites/*/*.robot
```

For more options, see the help output for the `robot` command.

```
$ ./_output/robotenv/bin/robot -h
```

## Test Scenarios

The test scenario tools in the `./test/bin` directory are useful for running
more complex test cases that require VMs and different images.

### Package Sources

In order to build different images, Image Builder needs to be configured
with all of the right package sources to pull the required RPMs. The
sources used by all scenarios are in the `./test/package-sources` directory.

Package source definition files are templates that support embedded shell
variables. Image build procedure expands those variables using `envsubst`.

> Refer to `./test/bin/build_images.sh` for the set of known variables that
> can be expanded.

```
id = "fast-datapath"
name = "Fast Datapath for RHEL 9"
type = "yum-baseurl"
url = "https://cdn.redhat.com/content/dist/layered/rhel9/${UNAME_M}/fast-datapath/os"
check_gpg = true
check_ssl = true
system = false
rhsm = true
```

### Image Blueprints

The image blueprints are independent of the test scenarios so that
images can be reused.

Images are organized into "layers" that contain "groups".

Each layer has its designation:
- `base` layer is a prerequisite for all the subsequent ones
- `presubmit` layer contains images used in presubmit CI jobs
- `periodic` layer contains images used in periodic CI jobs

Layers are built one after the other. Each group in a given layer is built
sequentially and all of the images in a group are built in parallel.

New blueprints can be added as TOML files in the appropriate layer and
group directories under `./test/image-blueprints`.

> **Warning**
> - Making changes to specific images may affect the way the image is used
> in different scenarios.
> - Each image takes time to build and adding unnecessary new images may
> slow down the overall test job.

#### Blueprint Customization

Image blueprint definition files are templates that support embedded shell
variables. Image build procedure expands those variables using `envsubst`.

> Refer to `./test/bin/build_images.sh` for the set of known variables that
> can be expanded.

Image blueprint templates also support the `RENDER_CONTAINER_IMAGES=version`
macro for rendering the embedded container image references into blueprints.
The `version` value is mandatory and it can be hardcoded, or referenced by a
shell variable. Version is used to uniquely identify the `microshift-release-info`
RPM file name for extracting the container image reference information.

```
RENDER_CONTAINER_IMAGES="${SOURCE_VERSION}"
```

#### Blueprint Naming

Blueprint names must be globally unique, regardless of the layer and group
that contains the blueprint template.

Blueprint names should clearly identify the combination of operating
system and MicroShift versions. The convention is to put the operating
system first, followed by `microshift` (i.e `rhel-9.2-microshift-4.13`).

Regardless of the branch, the blueprints using MicroShift built from
the source PR should use `source` in the name to facilitate rebuilding
only the source-based images (i.e `rhel-9.2-microshift-source`).

To make it easy to include the right image in a test scenario, each
blueprint produces an edge-commit image identified with a `ref` that
is the same as the blueprint name contained within the blueprint file.

#### Image Parents

`ostree` images work better when the `parent` image is clearly defined
for variants derived from it. The automation relies on file naming
conventions to determine the parent blueprint by extracting the prefix
before the first dash (`-`) in the filename and then using that to find
the blueprint **template** file, and ultimately the blueprint **name**.

For example, `rhel92-microshift-source` has prefix `rhel92`. There is
a blueprint template `./test/image-blueprints/layer1-base/group1/rhel92.toml`
that contains the name `rhel-9.2`, so when the image for
`rhel92-microshift-source` is built, the parent is configured as the
`rhel-9.2` image.

To support complex dependencies, a special `# parent = parent_name` directive
can be added to the blueprint files to override the parent specification. This
directive must be commented out as it is not recognized by `osbuild-composer`.

```
# Parent specification directive recognized by test/bin/build_images.sh to be
# used with the '--parent' argument of 'osbuild-composer'
# parent = "rhel-9.4-microshift-4.{{ .Env.PREVIOUS_MINOR_VERSION }}"
```

#### Image Reference Aliases

Sometimes it is useful to use the same image via a different
reference. To define an alias from one ref to another, create a file
with the name of the desired alias and the extension `.alias`
containing the name of the reference the alias should point to. For
example, to create an alias `rhel-9.2-microshift-source-aux` for
`rhel-9.2-microshift-source`:

```
$ cat ./test/image-blueprints/layer2-presubmit/group1/rhel-9.2-microshift-source-aux.alias
rhel-9.2-microshift-source
```

#### Installer ISO Images

To create an ISO with an installer image from a blueprint, create a
file with the extension `.image-installer` containing the name of the
blueprint to base the image on. For example, to create a `rhel92.iso`
file from the `rhel-9.2` blueprint:

```
$ cat ./test/image-blueprints/layer1-base/group1/rhel92.image-installer
rhel-9.2
```

### Downloaded ISO Images

To download a pre-built ISO image from the Internet, create a file with
the extension `.image-fetcher` containing the URL of the image to be fetched.
The name of the downloaded image will be derived from the base name of the
file with the `.iso` extension.

For example, create a `centos9.image-fetcher` file with a link to the CentOS 9
ISO image. The downloaded file will be named `centos9.iso`.

```
$ cat ./test/image-blueprints/layer1-base/group1/centos9.image-fetcher
https://mirrors.centos.org/mirrorlist?path=/9-stream/BaseOS/{{ .Env.UNAME_M }}/iso/CentOS-Stream-9-latest-{{ .Env.UNAME_M }}-dvd1.iso&redirect=1&protocol=https
```

> Note that the `.image-fetcher` file contents may contain Go template expressions
> that will be expanded at runtime.

### Preparing to Run Test Scenarios

The steps in this section need to be executed on a `development host`.

#### Building RPMs

The upgrade and rollback test scenarios use multiple builds of
MicroShift to create images with different versions. Use
`./test/bin/build_rpms.sh` to build all of the necessary packages and
copy the necessary files into locations that can be used as RPM repositories
by Image Builder.

#### Creating Images

Use `./test/bin/manage_composer_config.sh` to set up the system for building
images. Create the configuration and start the webserver using `create`.

```
$ ./test/bin/manage_composer_config.sh create
```

Optionally, use `create-workers [num_workers]` to create multiple workers for building
images in parallel. The image build process is mostly CPU and I/O
intensive. For a development environment, setting the number of workers to
half of the CPU number may be a good starting point. If no `num_workers` is set, the script
determines the ideal number of workers based on the number of CPU cores available.

```
$ ./test/bin/manage_composer_config.sh create-workers
```

> This setting is optional and not necessarily recommended for configurations
> with small number of CPUs and limited disk performance.

Use `./test/bin/build_images.sh` to build all of the images for all of the
blueprints available.

Run `./test/bin/build_images.sh -h` to see all the supported modes for
building images. For example, run the following command to only rebuild the
images that use RPMs created from source (not already published releases).

```
./test/bin/build_images.sh -s
```

### Configuring Test Scenarios

The steps in this section need to be executed on a `hypervisor host`.

If the `hypervisor host` is different from the `development host`,
copy the contents of the `_output/test-images` directory generated
in the [Preparing to Run Test Scenarios](#preparing-to-run-test-scenarios)
section to the `hypervisor host`.

```
MICROSHIFT_HOST=microshift-dev

mkdir -p _output/test-images
scp -r microshift@${MICROSHIFT_HOST}:microshift/_output/test-images/ _output/
```
#### Mirroring the container registry
In order to avoid possible disruptions from external sources such as container
registries, container registry is mirrored locally for all the images MicroShift
requires.

The registry will contain all images extracted from previously built MicroShift
RPMs in the `build` phase (taken from `microshift-release-info`).
A quay mirror is configured in the hypervisor, all images are downloaded from
their original registry and pushed into the new one. Each of the scenarios
will inject the mirror's configuration automatically. Access to the mirror
uses credentials and TLS.

This is enabled by default in `./test/bin/ci_phase_iso_boot.sh`, as it makes
pipelines more robust.
It is disabled by default in each of the scenarios to ease development cycle.
It is uncommon to run the full `./test/bin/ci_phase_iso_boot.sh` outside of
CI as it requires a powerful machine.

#### Global Settings

The test scenario tool uses several global settings that may be
configured before the tool is used in `./test/scenario_settings.sh`.
You can copy `./test/scenario_settings.sh.example` as a starting point.

`SSH_PUBLIC_KEY` -- The name of the public key file to use for
providing password-less access to the VMs.

`SSH_PRIVATE_KEY` -- The name of the private key file to use for
providing password-less access to the VMs. Set to an empty string to
use ssh-agent.

`SKIP_SOS` -- Disable sos report collection. This can speed up setup
in a local environment where sos can be run manually when
necessary. Do not enable this option in CI.

`VNC_CONSOLE` -- Whether to add a VNC graphics console to hosts. This
is useful in local developer settings where cockpit can be used to
login to the host. Set to `true` to enable. Defaults to `false`.

`SUBSCRIPTION_MANAGER_PLUGIN` -- Should be the full path to a bash
script that can be sourced to provide a function called
`subscription_manager_register`. The function must take 1 argument,
the name of the VM within the current scenario. It should update that
VM so that it is registered with a Red Hat software subscription to
allow packages to be installed. The default implementation handles the
automated workflow used in CI and a manual workflow useful for
developers running a single scenario interactively.

#### Configuring Hypervisor

Use `./test/bin/manage_hypervisor_config.sh` to manage the following
hypervisor settings:
- Firewall
- Storage pools used for VM images and disks
- Isolated networks
- Nginx webserver for serving images used in scenarios

Create the necessary configuration using `create`.

```
$ ./test/bin/manage_hypervisor_config.sh create
```

To cleanup after execution and teardown of VMs use `cleanup`.

```
$ ./test/bin/manage_hypervisor_config.sh cleanup
```

> For a full cleanup, the storage pool directory may need to be deleted manually.
> ```
> sudo rm -rf _output/test-images/vm-storage
> ```

#### Creating Test Infrastructure

Use `./test/bin/scenario.sh` to create test infrastructure for a scenario
with the `create` argument and a scenario directory name as input.

```
$ ./test/bin/scenario.sh create \
      ./test/scenarios/el92-src@standard-suite.sh
```

#### Enabling Connections to VMs

Run `./test/bin/manage_vm_connections.sh` on the hypervisor to set up the API
server and ssh port of each VM. The appropriate connection ports are
written to the `$SCENARIO_INFO_DIR` directory (refer to `common.sh`
for the setting for the variable), depending on the mode used.

For CI integration or a remote hypervisor, use `remote` and pass the
starting ports for the API server and ssh server port forwarding.

```
$ ./test/bin/manage_vm_connections.sh remote -a 7000 -s 6000 -l 7500
```

To run the tests from a local hypervisor, as in a local developer
configuration, use `local` with no other arguments.

```
$ ./test/bin/manage_vm_connections.sh local
```

### Run a Scenario

Use `./test/bin/scenario.sh run` with a scenario file to run the tests for
the scenario.

```
$ ./scripts/fetch_tools.sh robotframework
$ ./test/bin/scenario.sh run \
      ./test/scenarios/el92-src@standard-suite.sh
```

### Scenario Definitions

Scenarios are saved as shell scripts under `./test/scenarios` and
`./test/scenarios-periodics`.
Each scenario includes several functions that are combined
with the framework scripts to take the specific actions for the
combination of images and tests that make up the scenario.

The scenario script should be defined with a combination of the RHEL
version(s), MicroShift version(s), and an indication of what sort of
tests are being run. For example, `el92-src@standard-suite.sh` runs
the standard test suite (not the `ostree` upgrade tests) against
MicroShift built from source running on a RHEL 9.2 image.

Scenarios define VMs using short names, like `host1`, which are made
unique across the entire set of scenarios. VMs are not reused across
scenarios.

Scenarios use images defined by the blueprints created earlier. Blueprints
and images are reused between scenarios. Refer to "Image Blueprints" above
for details.

Scenarios use kickstart templates from the `./test/kickstart-templates`
directory. Kickstart templates are reused between scenarios.

All of the functions that act as the scenario API are run in the
context of `./test/bin/scenario.sh` and can therefore use any functions
defined there.

#### Scenarios coverage

Scenarios utilize following distinct MicroShift sources:
- `src`: built from source (code in PR)
- `base`: built from base branch (PR's target branch)
- `prel`: previous MicroShift minor release
- `crel`: current MicroShift minor release (already built and released
   RPMs like ECs, RCs, Z-stream). It is optional meaning that shortly after
   branch cut, before first EC is released, it will be skipped.

| Starting ref | End ref | Successful upgrade scenario | Failed upgrade scenario |
|--------------|---------|-----------------------------|-------------------------|
| `base` | `src` |`el92-base@upgrade-ok.sh` | **MISSING** |
| `prel` | `src` |`el92-prel@upgrade-ok.sh` | **MISSING** |
| `src` | `src` | **MISSING** | `el92-src@upgrade-failing-cannot-backup.sh` |
| `crel` | `src` | `el92-crel@upgrade-ok.sh` | `el92-crel@upgrade-fails.sh` |

#### scenario_create_vms

This function should do any work needed to boot all of the VMs needed
for the scenario, including producing kickstart files and launching
the VMs themselves.

The `prepare_kickstart` function takes as input the VM name, kickstart
template file relative to the `./test/kickstart-templates` directory, and
the initial edge image to boot. It produces a unique kickstart file _for
that VM_ in the `${SCENARIO_INFO_DIR}` directory. Use the function
multiple times to create kickstart files for additional VMs.

The `launch_vm` function takes as input the VM name. It expects a
kickstart file to already exist, and it defines a new VM configured to
boot from the installer ISO and the kickstart file.

The `launch_vm` function also accepts two optional arguments:
- The image blueprint used to create the ISO that should
be used to boot the VM (default to `$DEFAULT_BOOT_BLUEPRINT`).
- The name of the network used when creating the VM
(defaults to `default`).

#### scenario_remove_vms

This functions is used to remove any VMs defined by the scenario. It
should call `remove_vm` for each VM created by `scenario_create_vms`,
and take any other cleanup actions that might be unique to the
scenario based on other steps taken in `scenario_create_vms`.

It is not necessary to explicitly clean up the scenario metadata in
`${SCENARIO_INFO_DIR}`.

#### scenario_run_tests

This function runs the tests. It is invoked separately because the
same host may be used with multiple test runs when working on a local
developer system.

The function `run_tests` should be invoked exactly one time, passing
the primary host to use for connecting and any arguments to be given
to `robot`, including the test suites and any unique variables that
are not saved to the default variables file by the framework.

The `robot` command uses the following options that can be overridden
as an environment setting in scenario files:
* `TEST_RANDOMIZATION=all` for running the tests in a random order
* `TEST_EXECUTION_TIMEOUT=60m` for timing out on tests that run longer than expected

> Execution timeout is disabled with running the scenario script in the
> interactive mode to allow convenient interruption from the terminal.

## Troubleshooting

### Accessing VMs

Use `./test/bin/scenario.sh login` to login to a VM for a scenario as the
`redhat` user.

### Cleaning Up

On a `development host`, use `./test/bin/cleanup_composer.sh` to fully
clean composer jobs and cache, also restarting its services.

On a `hypervisor host`, use `./test/bin/cleanup_hypervisor.sh` to remove
the test infrastructure for all scenarios, undo the hypervisor configuration
and kill the web server process.

## CI Integration Scripts

### ci_phase_iso_build.sh

Runs on the hypervisor. Responsible for all of the setup to build all
needed images. Rebuilds MicroShift RPMs from source, sets up RPM repo,
sets up Image Builder workers, builds the images, and creates the web
server to host the images.

The script implements the following build time optimizations depending
on the runtime environment:
* When the `CI_JOB_NAME` environment variable is defined
  * The `layer3-periodic` groups are built only if the job name contains
    the`periodic` substring.
* When access to the `microshift-build-cache` AWS S3 Bucket is configured
  * If the script is run with the `-update_cache` command line argument, it
    builds `layer1-base` groups and uploads them to the S3 bucket.
  * If the script is run without command line arguments, it attempts to
    download cached `layer1-base` artifacts instead of building them.
* In any case, the fallback is to perform full builds on all layers.

> See [Image Caching in AWS S3 Bucket](#image-caching-in-aws-s3-bucket)
> for more information.

### ci_phase_iso_boot.sh

Runs on the hypervisor. Responsible for launching all of the VMs that
are used in the test step. Scenarios are taken from `SCENARIO_SOURCES`
variable, which defaults to `./test/scenarios`.

### ci_phase_iso_test.sh

Runs on the hypervisor. Responsible for running all of the scenarios from
`SCENARIO_SOURCES`, waiting for them to complete and exiting with an
error code if at least one test failed.

### Image Caching in AWS S3 Bucket

The `ci_phase_iso_build.sh` script attempts to optimizes image build times
when access to the `microshift-build-cache` AWS S3 Bucket is configured in
the current environment.

```
$ ./scripts/fetch_tools.sh awscli
You can now run: /home/microshift/microshift/_output/bin/aws --version

$ ./_output/bin/aws configure list
      Name                    Value             Type    Location
      ----                    -----             ----    --------
   profile                <not set>             None    None
access_key     ****************TCBI shared-credentials-file
secret_key     ****************pc5/ shared-credentials-file
    region                eu-west-1      config-file    ~/.aws/config

$ ./_output/bin/aws s3 ls
2023-10-29 08:38:40 microshift-build-cache
```

#### manage_build_cache.sh

The script abstracts build cache manipulation by implementing an interface
allowing to `upload`, `download`, `verify` and `cleanup` image build artifact
data using AWS S3 for storage.

The default name of the bucket is `microshift-build-cache` and it can be
overriden by setting the `AWS_BUCKET_NAME` environment variable.

The script uses `branch` and `tag` arguments to determine the sub-directories
for storing the data at `${AWS_BUCKET_NAME}/<branch>/${UNAME_M}/<tag>`. Those
arguments are set in the `common.sh` script using the following environment
variables:

* `SCENARIO_BUILD_BRANCH`: the name of the current branch, i.e `main`,
  `release-4.14`, etc.
* `SCENARIO_BUILD_TAG`: the current build tag using the `yymmdd` format.
* `SCENARIO_BUILD_TAG_PREV`: the previous build tag using yesterday's date.

A special `${AWS_BUCKET_NAME}/<branch>/${UNAME_M}/last` file can be set and
retrieved using `setlast` and `getlast` operations. The file contains the
name of the last tag with the valid cached data.

The cleanup operation is implemented using the `keep` operation, which deletes
data from all the tags in the current branch and architecture, except those
pointed by the `last` file and the `tag` argument.

> Deleting the `last` file and data pointed by it is possible by manipulating
> the AWS S3 bucket contents directly, using other tools.

#### Update CI Cache Job

Cache update is scheduled as periodic CI jobs named `microshift-metal-cache-nightly`
and `microshift-metal-cache-nightly-arm`. The jobs run nightly by executing the
`test/bin/ci_phase_iso_build.sh -update_cache` command.

There are also pre-submit CI jobs named `microshift-metal-cache` and
`microshift-metal-cache-arm`. The jobs are executed automatically in pull requests
when any scripts affecting the caching are updated.

> The cache upload operation does not overwrite existing valid cache to avoid
> race conditions with running jobs. The procedure exits normally in this case.
> Manual deletion of a cache tag in the AWS S3 bucket is required for forcing
> an existing cache update.

The job cannot use its current AWS CI account because it is regularly purged of
all objects older than 1-3 days. Instead, the cached data is stored in the
MicroShift development AWS account, which does not have the automatic cleanup
scheduled.

Environment variables are set to affect the AWS CLI command behavior:
- `AWS_BUCKET_NAME` is set to `microshift-build-cache-${EC2_REGION}` for
  ensuring that S3 traffic is local to a given region for saving costs.
- `AWS_PROFILE` is set to `microshift-ci` to ensure that all the AWS CLI
  commands are using the right credendials and region.

The credentials for accessing the MicroShift development AWS account are stored
in the vault as described at [Adding a New Secret to CI](https://docs.ci.openshift.org/docs/how-tos/adding-a-new-secret-to-ci/).
The keys are then mounted for the job and copied to the `~/.aws/config` and
`~/.aws/credentials` files under the `microshift-ci` profile name.

> The procedure assumes that the `microshift-build-cache-<region>` bucket exists
> in the appropriate region.

### Local Developer Overrides

In some cases, it is necessary to override default values used by the
CI scripts to make them work in the local environment. If the file
`./test/dev_overrides.sh` exists, it is sourced by the test framework
scripts after initializing the common defaults and before computing
any per-script defaults.

For example, creating the file with this content

```
#!/bin/bash
export AWS_BUCKET_NAME=microshift-build-cache-us-west-2
export CI_JOB_NAME=local-dev
```

will ensure that a valid AWS bucket is used as the source of image
cache data and that scripts that check for the CI job name will have a
name pattern set to compare.

NOTE: Include the shebang line with the shell set and export variables
to avoid issues with the shellcheck linter.
