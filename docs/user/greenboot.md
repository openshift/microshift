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

### Greenboot Configuration

MicroShift includes the `40_microshift_running_check.sh` health check script to
validate that all the required MicroShift services are up and running. The health
check script is packaged in the separate mandatory `microshift-greenboot` RPM,
which has an explicit dependency on the `greenboot` RPM.

The health check script is installed into the `/etc/greenboot/check/required.d`
directory and it is not executed during the system boot in case the `greenboot`
package is not present.

The `40_microshift_pre_rollback.sh` pre-rollback script is installed into the
`/etc/greenboot/red.d` directory, to be executed right before the system rollback
takes place. The script performs MicroShift pod and OVN cleanup to avoid potential
conflicts with the software rolled back to a previous version.

> The existing MicroShift data and container images are not affected by this operation.

In addition, if the `greenboot-default-health-check` RPM subpackage is installed,
it already includes [health check scripts](https://github.com/fedora-iot/greenboot#health-checks-included-with-subpackage-greenboot-default-health-checks)
verifying that DNS and `ostree` services can be accessed.

Greenboot redirects all the script output to the system log, accessible via the
following commands:
* `journalctl -u greenboot-healthcheck.service` for the health check procedure
* `journalctl -u redboot-task-runner.service` for the pre-rollback procedure

### MicroShift Validation

Exiting the health check script with a non-zero status will have the boot declared
as failed. The following validations are performed by the script.

| Validation                                                  | Pass | Fail   |
|-------------------------------------------------------------|------|--------|
| Check the script runs with 'root' permissions               | Next | exit 0 |
| Check microshift.service is enabled                         | Next | exit 0 |
| Wait for microshift.service to be active (!failed)          | Next | exit 1 |
| For each core namespace, wait for readiness of the workload | Next | exit 1 |

The pre-rollback script runs the `sudo microshift-cleanup-data --ovn` command
to prepare the system for a potential software downgrade.

> If the system is not booted using the `ostree` file system, the health check
> and pre-rollback procedures still run, but no rollback would be possible in
> case of an upgrade failure.

The wait period in each health check validation starts from 5 minutes base time
and it is incremented by the base wait period after each boot in the verification
loop. It is possible to override the base time wait period setting with the
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

## The `systemd` Journal Service Configuration

The default configuration of the `systemd` journal service stores the data in
the volatile `/run/log/journal` directory, which does not persist after a system
boot. To monitor `greenboot` activities across system boots, it is recommended to
enable the journal data persistency by creating the `/var/log/journal` directory
and setting limits on the maximal journal data size.

Run the following commands to configure the journal data persistency and limits.
```bash
sudo mkdir -p /etc/systemd/journald.conf.d
cat <<EOF | sudo tee /etc/systemd/journald.conf.d/microshift.conf &>/dev/null
[Journal]
Storage=persistent
SystemMaxUse=1G
RuntimeMaxUse=1G
EOF
```
