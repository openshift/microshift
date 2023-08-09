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
remotely via ssh.

`SSH_PRIV_KEY` should be an ssh key file to use for authenticating as
`USHIFT_USER`. The key must not require a password. To connect to
hosts using a key with a password, leave `SSH_PRIV_KEY` set to an
empty string and the tests will connect as `USHIFT_USER` and rely on
the ssh agent to provide the correct credentials.

`SSH_PORT` should be the port used for an ssh connection.

`API_PORT` should be set when connections are performed through a
forwarded port.

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
$ ./test/run.sh suites/show-config.robot
```

> The test suite file names should be relative to the `test` directory.

#### Running a Subset of Tests

To run a single test, or subset of tests, use the `robot` command's
features for filtering the tests. Use `--` to separate arguments to
`run.sh` from arguments to be passed unaltered to `robot`. The input
files must also be specified explicitly in all of these cases.

To run tests with names matching a pattern, use the `-t` option:

```
$ ./test/run.sh -- -t '*No Mode*' suites/show-config.robot
```

To run tests with specific tags, use the `-i` option:

```
$ ./test/run.sh -- -i etcd suites/*.robot
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
images can be reused. Be careful making changes to specific images in
case the change affects the way the image is used in different
scenarios. Be careful adding unnecessary new images, since each image
takes time to build and may slow down the overall test job.

Images are organized into "groups". Each group is built in order and
all of the images in a group are built in parallel.

Add blueprints as TOML files in the appropriate group directory in the
`./test/image-blueprints` directory, then add a short description of the
image here for reference.

Blueprint | Group | Image Name | Purpose
--------- | ----- | ---------- | -------
el92.toml | group1 | el92 | A simple RHEL image without MicroShift.
el92-prev.toml | group2 | el92-prev | A RHEL 9.2 image with the latest MicroShift from the previous y-stream installed and enabled.
el92-base.toml | group2 | el92-base | A RHEL 9.2 image with the RPMs built from base release branch (PR's merge target).
el92-src.toml | group2 | el92-src | A RHEL 9.2 image with the RPMs built from source.
el92-src-fake-y1.toml | group2 | el92-src-fake-y1 | A RHEL 9.2 image with the RPMs built from source from the current PR but with the _version_ set to the next y-stream.
el92-src-fake-y2.toml | group2 | el92-src-fake-y2 | A RHEL 9.2 image with the RPMs built from source from the current PR but with the _version_ set to the current+2 y-stream.

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

Blueprint names must be globally unique, regardless of the group that
contains the blueprint template.

Blueprint names should clearly identify the combination of operating
system and MicroShift versions. The convention is to put the operating
system first, followed by `microshift`. For example,
`el92-prev`.

Regardless of the branch, the blueprints using MicroShift built from
the source PR should use `src` in the name to facilitate rebuilding
only the source-based images. For example,
`el92-src`.

To make it easy to include the right image in a test scenario, each
blueprint produces an edge-commit image identified with a `ref` that
is the same as the blueprint name contained within the blueprint file.

#### Image Parents

`ostree` images work better when the `parent` image is clearly defined
for variants derived from it. The automation relies on file naming
conventions to determine the parent blueprint by extracting the prefix
before the first dash (`-`) in the filename and then using that to find
the blueprint **template** file, and ultimately the blueprint **name**.

For example, `el92-src` has prefix `el92`. There is
a blueprint template `./test/image-blueprints/group1/el92.toml` that
contains the name `el92`, so when the image for
`el92-src` is built, the parent is configured as the
`el92` image.

#### Image Reference Aliases

Sometimes it is useful to use the same image via a different
reference. To define an alias from one ref to another, create a file
with the name of the desired alias and the extension `.alias`
containing the name of the reference the alias should point to. For
example, to create an alias `el92-src-aux` for
`el92-src`:

```
$ cat ./test/image-blueprints/group2/el92-src-aux.alias
el92-src
```

#### Installer ISO Images

To create an ISO with an installer image from a blueprint, create a
file with the extension `.image-installer` containing the name of the
blueprint to base the image on. For example, to create a `el92.iso`
file from the `el92` blueprint:

```
$ cat ./test/image-blueprints/group1/el92.image-installer
rhel-9.2
```

### Preparing to Run Test Scenarios

The steps in this section need to be executed on a `development host`.

#### Building RPMs

The upgrade and rollback test scenarios use multiple builds of
MicroShift to create images with different versions. Use
`./test/bin/build_rpms.sh` to build all of the necessary packages.

#### Creating Local RPM Repositories

After building RPMs, run `./test/bin/create_rpm_repos.sh` to copy the
necessary files into locations that can be used as RPM repositories by
Image Builder.

#### Creating Images

Use `./test/bin/start_osbuild_workers.sh` to create multiple workers for
building images in parallel. The image build process is mostly CPU and I/O
intensive. For a development environment, setting the number of workers to
half of the CPU number may be a good starting point.

```
NCPUS=$(lscpu | grep '^CPU(s):' | awk '{print $2}')
./test/bin/start_osbuild_workers.sh $((NCPUS / 2))
```

> This setting is optional and not necessarily recommended for configurations
> with small number of CPUs and limited disk performance.

Use `./test/bin/start_webserver.sh` to run an `nginx` web server to serve the
images needed for the build.

Use `./test/bin/build_images.sh` to build all of the images for all of the
blueprints available.

Run `./test/bin/build_images.sh -h` to see all the supported modes for
building images. For example, run the following command to only rebuild the
images that use RPMs created from source (not already published releases).

```
./test/bin/build_images.sh -s
```

#### Rebuilding from Sources Easily

If you build new RPMs, you need to re-run several steps (build the
RPMs, build the local repos, build the images, download the images).
Use `./bin/rebuild_source_images.sh` to automate all of those
steps with one script while only rebuilding the images that use RPMs
created from source (not already published releases).

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

#### Global Settings

The test scenario tool uses several global settings that may be
configured before the tool is used in `./test/scenario_settings.sh`.
You can copy `./test/scenario_settings.sh.example` as a starting point.

`PUBLIC_IP` -- The public IP of the hypervisor, when accessing VMs
remotely through port-forwarded connections.

`SSH_PUBLIC_KEY` -- The name of the public key file to use for
providing password-less access to the VMs.

`SSH_PRIVATE_KEY` -- The name of the private key file to use for
providing password-less access to the VMs. Set to an empty string to
use ssh-agent.

`SKIP_SOS` -- Disable sos report collection. This can speed up setup
in a local environment where sos can be run manually when
necessary. Do not enable this option in CI.

#### Configuring Hypervisor

Use `./test/bin/manage_hypervisor_config.sh` to manage the following
hypervisor settings:
- Firewall
- Storage pools used for VM images and disks
- Isolated networks

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

Use `./test/bin/start_webserver.sh` to run an `nginx` web server to serve the
images needed for the test scenarios.

Use `./test/bin/scenario.sh` to create test infrastructure for a scenario
with the `create` argument and a scenario directory name as input.

```
$ ./test/bin/scenario.sh create \
      ./test/scenarios/el92_src_standard-suite.sh
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
      ./test/scenarios/el92_src_standard-suite.sh
```

### Scenario Definitions

Scenarios are saved as shell scripts under `./test/scenarios` and
`./test/scenarios-periodics`.
Each scenario includes several functions that are combined
with the framework scripts to take the specific actions for the
combination of images and tests that make up the scenario.

The scenario script should be defined with a combination of the RHEL
version(s), MicroShift version(s), and an indication of what sort of
tests are being run. For example,
`el92_src_standard-suite.sh` runs the standard test
suite (not the `ostree` upgrade tests) against MicroShift built from
source running on a RHEL 9.2 image.

Scenarios define VMs using short names, like `host1`, which are made
unique across the entire set of scenarios. VMs are not reused across
scenarios.

Scenarios use images defined by the blueprints created
earlier. Blueprints and images are reused between scenarios. Refer to
"Image Blueprints" above for details.

Scenarios use kickstart templates from the `./test/kickstart-templates`
directory. Kickstart templates are reused between scenarios.

All of the functions that act as the scenario API are run in the
context of `./test/bin/scenario.sh` and can therefore use any functions
defined there.

#### Scenarios testing between different MicroShift sources

Scenarios utilize following distinct MicroShift sources:
- `src`: built from source (code in PR)
- `base`: built from base branch (PR's target branch)
- `prev-minor`: previous MicroShift minor release

| Starting ref | End ref | Successful upgrade scenario | Failed upgrade scenario |
|--------------|---------|-----------------------------|-------------------------|
| `base` | `src` |`el92_base_upgrade-ok.sh` | `el92_base_upgrade-failing.sh` |
| `prev-minor` | `src` |`el92_prev-minor_upgrade-ok.sh` | `el92_prev-minor_upgrade-failing.sh` |
| `src` | `src` |`el92_src_upgrade-ok.sh` | `el92_src_upgrade-failing.sh` |

In future, another source of MicroShift should be added which is
most recent MicroShift RPMs built by ART (EC, then RC, and finally
Z stream releases matching version of currently tested code).
Both successful and failed upgrades scenarios should be added:
- `released` to `src` (presubmit)
- `released` to `main` / `release-4.YY` (periodic)

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

### ci_phase_iso_boot.sh

Runs on the hypervisor. Responsible for launching all of the VMs that
are used in the test step. Scenarios are taken from `SCENARIO_SOURCES`
variable, which defaults to `./test/scenarios`.

### ci_phase_iso_test.sh

Runs on the hypervisor. Responsible for running all of the scenarios from
`SCENARIO_SOURCES`, waiting for them to complete and exiting with an
error code if at least one test failed.
