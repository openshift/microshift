# Automated Functional and Integration Tests

This directory includes tests for MicroShift that go beyond unit
tests. They exercise the system more fully and verify behaviors
end-to-end.

The tests are written using [Robot
Framework](https://robotframework.org), a test automation framework
that separates the description of the test from the implementation of
the test.

## Test suites

Groups of tests are saved to `.robot` files in the `suites/`
directory. The tests in a suite should have the same basic
prerequisites and should test related functionality.

## Setting up

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

## Running the tests

Use `run.sh` to run the tests. It will create a Python virtual
environment in `$reporoot/_output` and install Robot Framework
automatically.

The `-h` option prints usage instructions:

```
$ ./test/run.sh -h
run.sh [-h] [-n] [-o output_dir] [test suite files]

Options:

  -h       Print this help text.

  -n       Dry-run, do not run the tests.

  -o DIR   The output directory.
```

### Running a Single Suite

By default, all of the test suites will be run. To run a subset,
specify the filenames as arguments on the command line.

```
$ ./test/run.sh suite/show-config.robot
```

### Running a Subset of Tests

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

## Test Scenarios in CI

The test scenario tools in the `bin` directory are useful for running
more complex test cases that require VMs and different images.

### Package Sources

In order to build different images, Composer needs to be configured
with all of the right package sources to pull the required RPMs. The
sources used by all scenarios are in the `package-sources` directory.

Package source definition files are templates using `envsubst`,
which means the files can have shell variables embedded.

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

Refer to `./bin/build_images.sh` for the set of known variables that can
be expanded.

## Image Blueprints

The image blueprints are independent of the test scenarios so that
images can be reused. Be careful making changes to specific images in
case the change affects the way the image is used in different
scenarios. Be careful adding unnecessary new images, since each image
takes time to build and may slow down the overall test job.

Add blueprints as TOML files in the `image-blueprints` directory, then
add a short description of the image here for reference.

Name | Purpose
---- | -------
rhel-9.2 | A simple RHEL image without MicroShift.
rhel-9.2-microshift-4.13 | A RHEL 9.2 image with the latest MicroShift 4.13 z-stream installed and enabled.
rhel-9.2-microshift-source | A RHEL 9.2 image with the RPMs built from source.

## Preparing to run test scenarios

### Creating the local RPM repository

After running `make rpm` at the top of the source tree, run
`./bin/create_local_repo.sh` from this directory to copy the necessary
files into a location that can be used as an RPM repository by
Composer. If you build new RPMs, you need to re-run
`create_local_repo.sh` *and* build new images.

### Creating the images

Use `./bin/start_osbuild_workers.sh` to create multiple workers for
building images in parallel. This is optional, and not necessarily
recommended on a laptop.

Use `./bin/build_images.sh` to build all of the images for all of the
blueprints available.

### Downloading the images for use by the test scenarios

Use `./bin/download_images.sh` to download all of the images from
Composer's cache and set up the directory for the web server to host
the files needed to launch VMs and run test scenarios.

### Global settings

The test scenario tool uses several global settings that may be
configured before the tool is used in `./scenario_settings.sh`. You
can copy `./scenario_settings.sh.example` as a starting point.

`PUBLIC_IP` -- The public IP of the hypervisor, when accessing VMs
remotely through port-forwarded connections.

`SSH_PUBLIC_KEY` -- The name of the public key file to use for
providing password-less access to the VMs.

`SSH_PRIVATE_KEY` -- The name of the private key file to use for
providing password-less access to the VMs. Set to an empty string to
use ssh-agent.

### Creating test infrastructure

Use `./bin/start_webserver.sh` to run a caddy web server to serve the
images needed for the test scenarios.

Use `./bin/configure_hypervisor_firewall.sh` to set up the firewall
rules that allow VMs to access the web server on the hypervisor.

Use `./bin/scenario.sh` to create test infrastructure for a scenario
with the `create` argument and a scenario directory name as input.

```
$ ./bin/scenario.sh create ./scenarios/rhel-9.2-microshift-source-standard-suite.sh
```

### Enabling connections to the VMs

Run `./bin/manage_vm_connections.sh` on the hypervisor to set up the API
server and ssh port of each VM. The appropriate connection ports are
written to the `$SCENARIO_INFO_DIR` directory (refer to `common.sh`
for the setting for the variable), depending on the mode used.

For CI integration or a remote hypervisor, use `remote` and pass the
starting ports for the API server and ssh server port forwarding.

```
$ ./bin/manage_vm_connections.sh remote -a 7000 -s 6000 -l 7500
```

To run the tests from a local hypervisor, as in a local developer
configuration, use `local` with no other arguments.

```
$ ./bin/manage_vm_connections.sh local
```

### Run a scenario

Use `./bin/scenario.sh run` with a scenario file to run the tests for
the scenario.

```
$ ./bin/scenario.sh run ./scenarios/rhel-9.2-microshift-source-standard-suite.sh
```

## Scenario definitions

Scenarios are saved as shell scripts under `scenarios`. Each
scenario includes several functions that are combined
with the framework scripts to take the specific actions for the
combination of images and tests that make up the scenario.

The scenario script should be defined with a combination of the RHEL
version(s), MicroShift version(s), and an indication of what sort of
tests are being run. For example,
`rhel-9.2-microshift-source-standard-suite.sh` runs the standard test
suite (not the ostree upgrade tests) against MicroShift built from
source running on a RHEL 9.2 image.

Scenarios define VMs using short names, like `host1`, which are made
unique across the entire set of scenarios. VMs are not reused across
scenarios.

Scenarios use images defined by the blueprints created
earlier. Blueprints and images are reused between scenarios. Refer to
"Image Blueprints" above for details.

Scenarios use kickstart templates from the `kickstart-templates`
directory. Kickstart templates are reused between scenarios.

All of the functions that act as the scenario API are run in the
context of `scenario.sh` and can therefore use any functions defined
there.

### scenario_create_vms

This function should do any work needed to boot all of the VMs needed
for the scenario, including producing kickstart files and launching
the VMs themselves.

The `prepare_kickstart` function takes as input the VM name, kickstart
template file relative to the `kickstart-templates` directory, and the
initial edge image to boot. It produces a unique kickstart file _for
that VM_ in the `${SCENARIO_INFO_DIR}` directory. Use the function
multiple times to create kickstart files for additional VMs.

The `launch_vm` function takes as input the VM name. It expects a
kickstart file to already exist, and it defines a new VM configured to
boot from the installer ISO and the kickstart file.

### scenario_remove_vms

This functions is used to remove any VMs defined by the scenario. It
should call `remove_vm` for each VM created by `scenario_create_vms`,
and take any other cleanup actions that might be unique to the
scenario based on other steps taken in `scenario_create_vms`.

It is not necessary to explicitly clean up the scenario metadata in
`${SCENARIO_INFO_DIR}`.

### scenario_run_tests

This function runs the tests. It is invoked separately because the
same host may be used with multiple test runs when working on a local
developer system.

The function `run_tests` should be invoked exactly one time, passing
the primary host to use for connecting and any arguments to be given
to `robot`, including the test suites and any unique variables that
are not saved to the default variables file by the framework.

## Troubleshooting

Use `./bin/scenario.sh login` to login to a VM for a scenario as the
`redhat` user.

## Cleaning up

Use `./bin/composer_cleanup.sh` to stop any running jobs, remove
everything from the queue, and delete existing builds.

Use `./bin/scenario.sh cleanup` to remove the test infrastructure for a
scenario.

```
$ ./bin/scenario.sh cleanup  ./scenarios/rhel-9.2-microshift-source-standard-suite/
```

## CI Integration Scripts

### ci_phase_iso_build.sh

Runs on the hypervisor. Responsible for all of the setup to build all
needed images. Rebuilds MicroShift RPMs from source, sets up RPM repo,
sets up osbuild workers, builds the images, and creates the web server
to host the images.

### ci_phase_iso_boot.sh

Runs on the hypervisor. Responsible for launching all of the VMs that
are used in the test step.
