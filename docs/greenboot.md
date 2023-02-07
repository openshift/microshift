# Integrating MicroShift with Greenboot

## Motivation

Serviceability of Edge Devices is often limited or non-existent, which makes it
challenging to troubleshoot device problems following a failed software or
operating system upgrade. 

To mitigate these problems, MicroShift uses [greenboot](https://github.com/fedora-iot/greenboot),
the Generic Health Check Framework for `systemd` on `rpm-ostree` based systems.
If a failure is detected, the system is expected to boot into the last known
working configuration using `rpm-ostree` rollback facilities.

This functionality benefits the users by reducing the risk of being locked out
of an Edge Device when upgrades take place. Users should not experience significant
interruption of service in case of a failed upgrade.

## Implementation Details

### MicroShift Validation

MicroShift includes the `40_microshift_running_check.sh` health check script to
validate that all the required MicroShift services are up and running. There is
no explicit dependency on the `greenboot` RPM from MicroShift core packages, or
a requirement to use health check procedures. The health check script is packaged
in the separate `microshift-greenboot` RPM, which may be installed on the system
if the `greenboot` facilities are to be used.

The health check script is installed into the `/etc/greenboot/check/required.d`
directory and it is not executed during the system boot in case the `greenboot`
package is not present.

In addition, if the `greenboot-default-health-check` RPM subpackage is installed,
it already includes
[health check scripts](https://github.com/fedora-iot/greenboot#health-checks-included-with-subpackage-greenboot-default-health-checks)
verifying that DNS and `ostree` services can be accessed.

Exiting the script with a non-zero status will have the boot declared as failed.
Greenboot redirects all the script output to the system log, accessible via the
`journalctl -u greenboot-healthcheck.service` command.

|Validation                                           |Pass  |Fail  |
|-----------------------------------------------------|------|------|
|Check the script runs with 'root' permissions        |Next  |exit 0|
|Check microshift.service is enabled                  |Next  |exit 0|
|Wait for microshift.service to be active (!failed)   |Next  |exit 1|
|Wait for Kubernetes API health endpoints to be OK    |Next  |exit 1|
|Wait for any Pod to start                            |Next  |exit 1|
|For each core namespace, wait for images to be pulled|Next  |exit 1|
|For each core namespace, wait for Pods to be ready   |Next  |exit 1|
|For each core namespace, check Pods not restarting   |exit 0|exit 1|

> If the system is not booted using the `ostree` file system, the health check
> procedures still run, but no rollback would be possible in case of an upgrade
> failure.

The wait period in each validation starts from 5 minutes base time and it is
incremented by the base wait period after each boot in the verification loop.
It is possible to override the base time wait period setting with the
`MICROSHIFT_WAIT_TIMEOUT_SEC` environment variable in the `/etc/greenboot/greenboot.conf`
configuration file alongside other [Greenboot Configuration](https://github.com/fedora-iot/greenboot#configuration)
settings.

### User Workloads Validation

Some 3rd party user workloads may become operational before the upgrade is
declared valid and potentially create or update data on the device. If a
rollback is performed subsequentially, there is a risk of data loss because the
file system is reverted to its state before the upgrade. One of the ways to
mitigate this problem is to have 3rd party workloads wait until a boot is
declared successful.

```bash
$ sudo grub2-editenv - list | grep ^boot_success
boot_success=1
```

> Note that the MicroShift health check script only performs validation of the
> core MicroShift services. Users should install their own workload validation
> scripts using `greenboot` facilities to ensure the successful operation after
> system upgrades.
