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
