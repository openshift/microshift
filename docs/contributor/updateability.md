# MicroShift Updateability

Following document is meant to be a contributor's introduction
to the feature of MicroShift's updateability.

## What is and isn't "MicroShift updateability"

Updateability of MicroShift is set of features and components that allow
MicroShift to be updated from version to another.

It includes following areas:
- backing up and restoring MicroShift's data
- persisting and verifying version of MicroShift's data and binary
- migrating Kubernetes objects to newer versions (e.g. `v1beta1` to `v1beta2`)

On ostree-based systems, backups and restores are automated, tied to the
lifecycle of ostree deployments. On regular RPM systems (non-ostree), the process
is manual using `microshift backup` and `microshift restore` commands.

Updateability is not a disaster recovery. Backups are created to allow
rolling back to a healthy system after failed upgrade.

## High level flows

### Preparing ostree commit (image) with MicroShift for initial install

To build an ostree commit you need:
- builder host with RHEL9, active Red Hat Subscription, and osbuild installed
- MicroShift RPMs
- blueprint for creating ostree commit (and, optionally, installer ISO)

For instructions on installing MicroShift on R4E see 
[Install MicroShift on RHEL for Edge](/docs/contributor/rhel4edge_iso.md).

To learn about MicroShift's test harness for testing on ostree systems see
[test/README.md](/test/README.md).

### Updating system to a new ostree commit

To update system to a newer commit (that may include newer MicroShift,
but doesn't have to) follow the same procedure on preparing ostree commit,
but this time provide new MicroShift RPMs.

When new ostree commit is part of repository known to the system you can use
of the following commands:
- `rpm-ostree upgrade` - if new commit has the same reference
  (e.g. `rhel/9/x86_64/edge/microshift`), ostree should detect that reference
  points to a new commit, fetch it, and update the system
- `rpm-ostree rebase` - if new commit has different reference,
  `rebase` must be used to point ostree to a new reference to use for the system.

After rpm-ostree staged new deployment (commit), restart the machine
(e.g. `systemctl reboot`) to boot the new image and make system run using it.
MicroShift will detect that the deployment (image) changed and will perform
steps needed to synchronize on-disk data with the MicroShift executable.

### Scenarios

MicroShift uses Robot Framework to test updateability in automated fashion.
Here's a list of existing scenario:

[Healthy upgrade of MicroShift with one minor version bump](/test/suites/upgrade/upgrade-successful.robot):
- System must be healthy
- New deployment is staged and host is rebooted
- System, using new image, should end up healthy and not roll back.
- MicroShift should:
  - Create a backup for previous deployment
  - Migrate all Kubernetes resources to most recent versions

[System rolls back because MicroShift fails to create a backup resulting in unhealthy system](/test/suites/upgrade/upgrade-fails-and-rolls-back.robot):
- Initially system must be healthy
- New deployment is staged and host is rebooted
- MicroShift never succeeds to make a backup of existing data:
  - MicroShift never attempts to start the cluster
  - Health check fails, i.e. system is unhealthy
  - Greenboot restarts the host several times attempting to remediate the situation
    but MicroShift consistently fails to back up
- System rolls back to previous deployment
- This time MicroShift does not have problems backing up the data
- MicroShift starts fully
- System is healthy
- We check if MicroShift created the backup matching information of previous healthy boot

> Note: test makes use of `microshift-test-agent` to make `/var/lib/microshift-backups`
> immutable resulting in MicroShift not being able to create a backup directory.

[Initially system does not feature MicroShift, first attempt to deploy image with MicroShift fails and second attempt succeeds (FDO)](/test/suites/upgrade/fdo.robot):
- Make sure system does not have `microshift` binary and neither
  `/var/lib/microshift` nor `/var/lib/microshift-backups` does not exists
- Deploy a new image with MicroShift and reboot the host
- New deployment starts, but for any reason it never becomes healthy
- After greenboot "healing" reboots are exhausted, system rolls back to initial
  deployment that does not have `microshift` binary, but now it has
  `/var/lib/microshift` and `/var/lib/microshift-backups`
  (leftover from failed deployment)
- Stage a new image with MicroShift and reboot the system
- MicroShift starts, inspects existing on-disk artifacts and decides to remove
  existing data to start from scratch.
- MicroShift successfully starts, system is healthy and does not roll back to
  initial deployment (without MicroShift)

> Note: test makes use of `microshift-test-agent` to create additional
> greenboot healthcheck script that always fails resulting in system rolling
> back after all greenboot reboots are exhausted.

[When system is healthy, rebooting it should trigger MicroShift to create a new backup](/test/suites/backup/backup-restore-on-reboot.robot):
- System is healthy
- Host is rebooted manually
- MicroShift starts and creates a new backup named: `currentDeploymentID_previousBootID`

[When system is unhealthy, rebooting it should trigger MicroShift to restore data from a backup](/test/suites/backup/backup-restore-on-reboot.robot):
- Precondition: backup exists for current deployment
- System is unhealthy
- Host is rebooted manually
- MicroShift restores existing backup matching currently booted deployment
  and starts successfully (system should be healthy)

[When new image contains older minor version of MicroShift, it will refuse to run resulting in a rollback](/test/suites/upgrade/downgrade-block.robot)
- System must be healthy
- A new image is staged on the system which contains older minor version of
  Microshift than the version that is currently running
- System is rebooted
- MicroShift starts and notices that current binary is "older" than
  version of the data, so it refused to run
- System becomes unhealthy, reboots couple of times, eventually rolls back to
  previous deployment.
- MicroShift's journal should contain an information about the failure:
  `checking version compatibility failed`

[When system is manually rolled back to a deployment that features older MicroShift, it should restore the backup for that version and run successfully](/test/suites/upgrade/rollback-manual.robot):
- System should be running deployment with older MicroShift initially.
- Stage new deployment with newer MicroShift and reboot the host.
- MicroShift should create a backup for the old deployment,
  then successfully start
- When greenboot finishes run and determines the system to be healthy,
  roll back the system: `rpm-ostree rollback` and reboot it.
- When system starts, it should run initial deployment (with older MicroShift)
- MicroShift should start successfully
  - It should include an information that a backup was restored
  - It should not attempt to perform an "upgrade"

**Test ideas for implementation**
- Testing that upgrading from unhealthy system is blocked
  (this probably needs to be implemented)
  - Make system with MicroShift unhealthy,
    so that information is persisted in the `health.json`
  - Stage a new deployment and reboot the host
  - MicroShift should detect that it's a new deployment and previous one was unhealthy,
    it should refuse to start, thus render system unhealthy
  - System should roll back to "original" unhealthy deployment
  - Greenboot should declare "system need manual intervention"

- Testing "blocking upgrade from version"
  - Set up a list of blocked upgrades with fake versions
  - Build RPMs with these fake versions
  - Deploy "from" MicroShift, then try upgrading to a version that has "from"
    on block list

## Context: RHEL For Edge

MicroShift's updateability primarily target integration with RHEL For Edge,
which is RHEL with some unique technologies and tools such as ostree
(system immutability) and greenboot (health check framework for systems).

This section provides an overview of this environment.

### Greenboot

[Greenboot](https://github.com/fedora-iot/greenboot) is a health check framework,
primarily for ostree-based systems.

For more in-depth information about greenboot and how MicroShift integrates
with it see following enhancements:
[Integrating MicroShift with Greenboot](https://github.com/openshift/enhancements/blob/master/enhancements/microshift/microshift-greenboot.md)
[MicroShift updateability in ostree based systems: integration with greenboot](https://github.com/openshift/enhancements/blob/master/enhancements/microshift/microshift-updateability-ostree.md#integration-with-greenboot)

In short, on system boot, greenboot will run health check scripts
(residing in `/etc/greenboot/check/{required.d,wanted.d}/`)
and, depending on result of *required* scripts, will run either "green" 
(healthy) or "red" (unhealthy) scripts. 

Greenboot strongly integrates with grub.
When new deployment is staged, greenboot sets `boot_counter` variable to a 
specific value (default: 3). Then, when host boots, it is grub that decrements
that variable. If `boot_counter` falls down to 0, grub will select alternate
boot entry (i.e. rollback deployment). See [Grub](#grub-ostree) section for
more information.

MicroShift delivers health check, red, and green scripts
in `microshift-greenboot` RPM package.

Health check script for MicroShift resides in the repository
as `packaging/greenboot/microshift-running-check.sh`.

Before the updateability feature was implemented, `microshift-greenboot` already
provided `packaging/greenboot/microshift-pre-rollback.sh` file as a "red" script:
if `boot_counter` equals `0`, the script will run `microshift-cleanup-data --ovn`.

> Note: It's a red script, so it runs after healthchecks decided that
> system is unhealthy, and cleanup only happens when `boot_counter`
> already fell down to `0`.
>
> So it does not happen on each reboot of unhealthy system.

#### Updateability additions to `microshift-greenboot`

Updateability feature added two new files - one "green" and one "red" script:
- `packaging/greenboot/microshift_set_healthy.sh` -> `/etc/greenboot/green.d/40_microshift_set_healthy.sh`
- `packaging/greenboot/microshift_set_unhealthy.sh` -> `/etc/greenboot/red.d/40_microshift_set_unhealthy.sh`

These scripts are only functional on ostree-based systems[^1] for following reasons:
- greenboot on non-ostree systems is only informatory, not functional
- `health.json` is more of a implementation detail to handle automated backups
  - journal logs (`greenboot-healthcheck`, `microshift`) are recommended
    for debugging system health issues
- automated backup names include deployment ID, which is not available on non-ostree
  - using deployment ID as a backup prefix serves as organization and allows to
    relatively easy match backups to deployments (and in extension, specific ostree commits)

[^1]: By checking if `/run/ostree-booted` file exists.

Primary responsibility of "set healthy" and "set unhealthy" scripts
is to create or update `health.json`.

`microshift_set_unhealthy.sh` will not overwrite `health.json` if a backup for
information in that file was not created (deployment and boot IDs).
This behavior ensures that MicroShift does not lose information that it should
perform a backup.

<details>
<summary>Example of scenario that behavior is needed</summary>

- System is booted to newer deployment
- MicroShift fails to create a backup
- System ends up unhealthy
- After reboot, we still want MicroShift to make an attempt to back up the data,
  which would not happen if we'd write "unhealthy" to the file.
  - Otherwise, after rollback, MicroShift would want to restore:
    - Outdated backup (because the most recent one that should be created failed)
    - No backup at all (because previous deployment was not rebooted,
      so the one and only backup procedure failed).
      - *Pending implementation: actually this should be handled by the MicroShift*
        *when it will start comparing "backup-to-restore" with version metadata*
        *extended with deployment and boot IDs.*
</details>

#### Greenboot on non-ostree systems

Although greenboot can be installed and used on non-ostree system, its
usefulness is greatly diminished. It does not perform automated reboots that
try to help system get healthy, nor does it cause a rollback to previously
working system or MicroShift version.

### (rpm-)ostree [ostree-only]

ostree is technology for creating git-like repositories for filesystems.

It allows to create whole operating system images and switch between them
at boot time.
It also provides an immutability - only handful of directories are writeable,
core of the system should be updated by creating new images.

For information:
- [ostree](https://ostreedev.github.io/ostree/)
- [rpm-ostree](https://coreos.github.io/rpm-ostree/)

#### `rpm-ostree` vs `ostree`

`ostree` (also referred to as OSTree or libostree) is a shared library and
set of tools for creating the repositories (which commits and branches).

`rpm-ostree` is built on top of ostree to create "full hybrid image/package system".
It connects power of ostree and dnf libraries to provide deployment upgrades
and rollbacks and RPM package layering.

#### Commits and deployments

Using osbuild to create new system images produces "ostree commits". Commit
contains timestamp, log message, and checksums. Each commit can have a parent,
so it can be used as a history of the builds.

When rpm-ostree is provided a commit to deploy, that commit becomes a
deployment - a bootable root filesystem.

> Note: ostree can be used to hold not only OS filesystems but any file structure.
> One example of such project is flatpak which uses ostree internally,
>
> For example: `flatpak info -r com.slack.Slack` will print very familiar
> looking reference `app/com.slack.Slack/x86_64/stable`.

More information:
- [Anatomy of an OSTree repository](https://ostreedev.github.io/ostree/repo/)
- [rpm-ostree: client administration](https://coreos.github.io/rpm-ostree/administrator-handbook/)
- [Flatpak: under the hood](https://docs.flatpak.org/en/latest/under-the-hood.html)

#### Useful `ostree` commands

> Hint: By default, `ostree` searches for repository in current directory.
>
> Repository can be explicitly specified with option `--repo=PATH`.

`ostree` is more about managing ostree repositories, but it also provides
some functionality to manage ostree-booted systems with `ostree admin`.

- List references in ostree repository: `ostree refs`
- Create new reference pointing to a commit: `ostree refs --create=NEWREF REVISION`
- List registered remotes: `ostree remote list`
- List url of a remote: `ostree remote show-url REMOTE`
- List references present on the remote: `ostree remote refs REMOTE`
- Add references to local repository from another local repository
  (for example after producing a new commit using osbuild): `ostree pull-local`
- Clean repository from old data that is no longer accessible
  directly by references: `ostree prune --refs-only`
- Pin deployment so it's not deleted when new deployments
  are added to the system: `ostree admin pin`

More info:
- [ostree: repository management](https://ostreedev.github.io/ostree/repository-management/)

#### Useful `rpm-ostree` commands


- Get list of deployments on the system: `rpm-ostree status`
- Switch to a different ostree reference, for example from `rhel-9.2-microshift-source` to 
  `rhel-9.2-microshift-source-fake-next-minor`: `rpm-ostree rebase REF`
- Rollback the system (sets previous deployment as active/current,
  so it's default after reboot): `rpm-ostree rollback`
- Update system to a newer version of current reference: `rpm-ostree upgrade`
- Install additional RPMs by layering them on top of current deployment: `rpm-ostree install`
- Remove layered RPMs: `rpm-ostree uninstall`
- Replace base image RPM with different version: `rpm-ostree override replace`
- Remove base image RPM: `rpm-ostree override remove`
- Reset overrides: `rpm-ostree override reset`
- Get kernel boot parameters: `rpm-ostree kargs`
  - Edit parameters: `rpm-ostree kargs --editor`
  - See also: `--append`, `--replace`, `--delete`

More info:
- [rpm-ostree: administrator's handbook](https://coreos.github.io/rpm-ostree/administrator-handbook/)

### Grub [ostree]

Greenboot sets and clears grub env vars such as `boot_counter` and `boot_success`.
But it's the grub decrementing the `boot_counter` and selecting alternate boot entry
(rollback deployment) when that variable falls down to 0.

Grub scripts involved in this process are delivered in `grub2-tools` package.

#### 08_fallback_counting

File resides in `/etc/grub.d`.

This file contains logic to do the actual counting down of the `boot_counter`.
Functionality is only executed if `boot_counter` exists and `boot_success` equals `0`.

> This is important because there is a `grub-boot-success.{timer,service}`
> which, after 2 minutes from user's logon, sets `boot_success` to `1`
> and thus causing `boot_counter` to not be decremented.

If `boot_counter` is `0` or `-1`, then variable `default` is set to `1`.
Otherwise, `boot_counter` is decremented
(*it's set by the greenboot to a configurable value, by default `3`, when staging new deployment*).

Setting `default` to `1` effectively changes which boot entry (1st one) will boot.

#### 10_reset_boot_success and 12_menu_auto_hide

Files reside in `/etc/grub.d`.

`10_reset_boot_success` contains logic for deciding to hide the boot menu
depending on values of `boot_success` (*last boot was ok*) and `boot_indeterminate` 
(*first boot attempt to boot the entry*) and may change values of these to
variables (and `menu_hide_ok`).

`12_menu_auto_hide` makes use of `menu_hide_ok` variable.

These scripts were not yet observed to have a big impact on efforts related to
updateability feature, but are most likely important part of grub and ostree integration.

#### grub-boot-success.{timer,service}

Both files are installed into `/usr/lib/systemd/user/` meaning they are intended
for a user, rather than a system.

`grub-boot-success.timer` is a 2 minute timer that starts when user logs in.
When user session is active for 2 minutes, it triggers `grub-boot-success.server`
which sets `boot_success`.

> **IMPORTANT**: Above means that active user session influences whole
> ostree+greenboot+grub integration. By setting `boot_success` to `1` it causes
> grub to not decrement the `boot_counter`.
>
> [Discussion on fedora-iot/greenboot about this particular issue](https://github.com/fedora-iot/greenboot/issues/108).

It is recommended that the timer is masked, so user session does not interfere
with the greenboot's process of assessing health, rebooting, and rolling back
the system.
This can be done during system installation in kickstart by including following
command in the post install section:
```
ln -sf /dev/null /etc/systemd/user/grub-boot-success.timer
```

#### Other files

There are other services that are involved in grub's envvar management: 
- `/usr/lib/systemd/system/grub-boot-indeterminate.service`
- `/usr/lib/systemd/system/grub2-systemd-integration.service`

Their impact on MicroShift was not investigated and is currently unknown,
however it is not expected to be significant.

### "grub * greenboot * rpm-ostree" integration summary

Here's a brief summary on how grub, greenboot, and rpm-ostree work together.

- User deploys a new deployment, e.g. using `rpm-ostree rebase` or `upgrade`
- User issues a command to reboot the host
- Just before shutting down, `greenboot-grub2-set-counter.service` followed by
  `ostree-finalize-staged.service` run.
  - `greenboot-grub2-set-counter.service` sets `boot_counter` to specified value
    (by default 3)
  - `ostree-finalize-staged.service` executes internal commands to finalize
    staged deployment
- System shuts down and boots
- Grub inspects `boot_counter` and `boot_success`
  - If `boot_counter` is set and `boot_success` is 0, the counter is decremented
- Grub boots recently staged deployment because it is first on the
  list (set as default)
- System starts, `greenboot-healthcheck.service` starts
  - `greenboot-healthcheck.service` runs healthcheck scripts and
    depending on the result runs green or red scripts
  - If system was unhealthy, `redboot-auto-reboot.service` reboots the host
    to give it another chance for a successful boot
- System reboots, grub decrements the counter
- Let's assume that new deployment continuously fails,
  greenboot reboots the system several times
- Grub sees that `boot_counter` is 0:
  - It changes the default boot from 0 to 1,
    that is: from failing deployment to the rollback deployment.
  - Sets `boot_counter` to -1
- Rollback deployment starts
- `greenboot-rpm-ostree-grub2-check-fallback.service` runs and inspects 
  `boot_counter` which is -1, so it executes `rpm-ostree rollback` command
  without rebooting the system. This syncs the rpm-ostree state with grub's
  new default boot. Then, it also removes the `boot_counter`.
- If system ends up unhealthy again, `redboot-auto-reboot.service`:
  - If `boot_counter` does not exist: print information that system is unhealthy
    and requires manual intervention
  - If `boot_counter` is `-1`: prints information that fallback boot was detect,
    but system is still unhealthy and requires manual intervention

## MicroShift Updateability Implementation

Upgradeability introduces following features that run in following order:
- backup management,
- version metadata management,
- Kubernetes storage migration.

These three areas are mostly separate, but backup management
may inspect version metadata to handle some corner cases.
Backup management and version metadata management run
before MicroShift attempts to start all the components and cluster itself.
Storage migration runs when cluster is up and running 
(and in 10 minutes intervals afterwards) because it
requires runtime components to correctly perform its task.

### On-disk artifacts

New on-disk artifacts were added:
- `/var/lib/microshift-backups/`
    - `health.json`
    - backup directories
- `/var/lib/microshift/version`

#### `/var/lib/microshift-backups/`

New, supporting directory for files and directories related to updateability.

It is non-configurable directory where automated backups are created on
ostree-based system[^2].

Users can also create manual backups in that location by including it
in the argument for `microshift backup` command, for example:
```
$ sudo microshift backup /var/lib/microshift-backups/manual-backup
```

[^2]: MicroShift's data and backup directories need to
reside on the same filesystem to leverage Copy-on-Write.

##### `health.json` [ostree-only]

`health.json` holds information about health of the system.

When system starts, the file contents refer to previous boot.
Only after `greenboot-healthcheck` completes health check procedure, the file is
updated with information about current boot by green and red scripts supplied
by MicroShift as part of greenboot integration.

> Note: Health assessment of MicroShift takes several minutes.
>
> To follow health check progress use command: `journalctl -fu greenboot-healthcheck`

<details>
<summary>Schema of health.json file</summary>

```json
# /var/lib/microshift-backups/health.json
{
    "health": "",
    "deployment_id": "",
    "boot_id": ""
}
```

- `health`: System status according to the greenboot health checks.
  - Expected values: `healthy` or `unhealthy`
- `deployment_id`: OSTree deployment ID.
  - Obtained with a command: `rpm-ostree status --json | jq '.deployments[].id'`
- `boot_id`: Boot ID generated by kernel.
  - Obtained with command: `tr -d '-' < /proc/sys/kernel/random/boot_id`.
  - Hyphens are removed to match format used by `journalctl` (see `journalctl --list-boots --reverse`)

Example file:
```json
{
  "health": "healthy",
  "deployment_id": "rhel-fe6192b549e3a787baa0d146dfc078ec4274e16fe42e7017ffecc6153dc473a6.0",
  "boot_id": "08f7e67d736e49b08402d0782a605b81"
}
```
</details>

##### Backups

> Note: Automated backing up and restoring is only on ostree-based systems
> as part of deployment staging and rolling back integration.

`/var/lib/microshift-backups/` also stores backups of MicroShift data.
Backups are created by copying data directory with `--reflink`
option to leverage Copy-on-Write functionality.


<details>
<summary>Backup naming schema</summary>

Naming schema of backup directories: `deploymentID_bootID[_unhealthy]`.
For example: 
```
rhel-fe6192b549e3a787baa0d146dfc078ec4274e16fe42e7017ffecc6153dc473a6.0_d5c48cf07f4442d1af593944789fb232_unhealthy
rhel-027a0e8a3be037246cc3eb8d1a81f55305f7a7e3e501d0108898766273481748.0_ebeedaa333364d81aa1b0a6c5d0a4bf0
|---------------------------------------------------------------------| |------------------------------|
                          deployment ID                                             boot ID
```

Suffix `_unhealthy` is used to mark backups of unhealthy systems.
Unhealthy backups are not taken into consideration during restore process.

Unhealthy backups are performed when all existing metadata suggests
that MicroShift should wipe the data and start from clean state
to improve chances for successful operation, but we want to err on a safe side
and keep the data if admin wishes to access it.

Example of such scenario is so called "FDO" where unhealthy deployment leaves
stale data.
</details>

#### `/var/lib/microshift/version`

This file holds version of last MicroShift executable that opened and used the data
together with the deployment and boot IDs of when that happened.
If MicroShift fails to perform backup or version metadata management and decides to
exit without attempting to start, the file will not be updated.

- On first start (data does not exist yet) MicroShift will create
  the file with executable's version as a content.
- If the file already exists, MicroShift will compare it against version of executable.
  And only after successful checks, the file is updated.

For more details about version comparison see 
[Version metadata management](#version-metadata-management).


<details>
<summary>Schema of version file</summary>

```
{
  "version": "MAJOR.MINOR.PATCH",
  "deployment_id": "rhel-...",
  "boot_id": ""
}
```

- All of MAJOR, MINOR, and PATCH are unsigned integers. For example: `4.14.0`
- `deployment_id` is an optional field - it's not used on regular RPM (non-ostree) systems.

> Note: MicroShift writes to a file without trailing newline.
> However, it should successfully handle a file with newline.

</details>


### Automated backup management [ostree-only]

It runs very close to the start of MicroShift's `run` command.

It checks for existence of `/run/ostree-booted` file.
If the file is not present, it is assumed that the system is regular RPM based
and this stage is skipped without an error.

#### Handling unusual scenarios

Then, procedure checks existence of three on-disk artifacts:
- `/var/lib/microshift` (ignores `.nodename`[0]) - referred to as data below
- `/var/lib/microshift/version`
- `/var/lib/microshift-backups/health.json`

[0] `.nodename` might be created by the "effective config generation" phase
which happens before the prerun. This can create `.nodename` and result
in `/var/lib/microshift` being present but, beside that one file, empty.

##### `/var/lib/microshift` (data) does not exist

> Lack of `/var/lib/microshift` implies that `/var/lib/microshift/version` does not exist.

If both data and `health.json` are missing, then we assume it's a first run of MicroShift.

If data does not exist but `health.json` is present, it means that MicroShift ran on the system already.
Perhaps it was unhealthy and admin decided to delete the data dir to try clean start.
In such case MicroShift will inspect `health.json` and depending of value of `health`:
- `healthy` - last boot with MicroShift was OK, but there is no data?
  There's nothing to back up, so MicroShift will continue start up.
- `unhealthy` - MicroShift will look for a healthy backup for currently
  booted deployment 
  - if backup is found, it'll be restored
  - if there is no backup to restore from, MicroShift will continue start up (no data = clean start).


##### Data exists, but `version` does not

If `health.json` exists, then what could have happened is failed upgrade from 4.13,
resulting in system rollback, followed by admin's manual intervention by removing
`/var/lib/microshift`, but keeping `/var/lib/microshift-backups`.

That's why, ignoring completely existence of `health.json`,
if `/var/lib/microshift` exists but `version` file does not,
MicroShift will treat the data as an upgrade from MicroShift 4.13 
(because it's the last release not featuring that file), create a backup
named `4.13` and proceed with the start up.

> TODO: if system rolls back to 4.13 deployment, and an upgrade is attempted again
> without manually deleting `4.13` backup after manually restoring it,
> MicroShift will fail to make a backup and exit with an error
> (because backup with such name already exists).
> It feels like we should handle this more gracefully.

##### Both data and `version` exist, but `health.json` does not

On the very first run of a deployment featuring MicroShift,
none of the `/var/lib/microshift` or `/var/lib/microshift-backups` exist.
In such case `health.json` will exists only after greenboot-healthcheck finishes
the assessment and executes green or red scripts.

Possible, common scenarios that can get MicroShift into that state:
- Restart of microshift.service to reload the config
- Power loss or unexpected reboot of the machine before end of greenboot-healthcheck

To gracefully handle such scenarios, MicroShift will skip backup management
and proceed with the start up.

> Side note on non-first boots where `health.json` exists,
> but actions derived from its data were already performed:
>
> In case of microshift.service restart (e.g. to reload the config),
> it will re-attempt to perform the same actions it did on first start, for example:
> - `healthy`: back up the data - which won't happen, because backup already exists (not an error) 
> - `unhealthy`: attempt an "unhealthy system procedure" (described later).


#### Regular, expected in most cases, backup management procedure

##### Boot ID stored in `health.json` matches current boot's ID: skip backup management and continue with start up

It means that information in `health.json` is intended for next boot
(be acted upon after system reboot) - it is a consequence of decision to
perform backup management on system boot.


##### `health.json` contains `healthy`

First, MicroShift will attempt to create a backup of the data named
with deployment and boot IDs from the `health.json`.

Then, if deployment ID stored in `health.json` differs from currently 
booted deployment's ID, it means different deployment was booted.

This can happen when:
- system is upgraded to a new deployment,
- system was rolled back,
- admin manually booted rollback deployment in grub menu.

To handle "going back to rollback deployment" scenario, MicroShift will check
the list of existing backups for one suitable for the currently booted deployment,
and, if found, it will restore that backup.
If no backup is found (e.g. upgrade), then MicroShift will continue start up.


##### `health.json` contains `unhealthy`

> Following steps might not be in the same order as implementation
> but should be functionally the same.

Summary:
- Try to restore backup for current deployment and continue start up
- If rollback deployment doesn't exist: continue start up
- Try to restore backup for rollback deployment and continue start up
- If there is no backup for neither the current nor rollback deployments:
  remove the data and continue start up (fresh)
- If `health.json` contains rollback deployment ID: exit with failure
- If `deployment_id` in `health.json` is neither current nor rollback deployments' ID:
  make "unhealthy" backup, remove the data and continue start up (fresh).

**If a backup for current deployment exists**: MicroShift will try to restore it and continue start up.

**If such backup does not exist**: MicroShift will try to restore backup for the rollback deployment.

**If rollback deployment does not exist** (i.e. there's only one deployment on the system):
it might be that system was manually rebooted or microshift.service manually restarted.
(It's not a reboot initiated by the greenboot because it only happens when `boot_counter` is set
and it's only set when deployment is staged.)
MicroShift will skip the backup management and continue with start up.

**If rollback deployment matches deployment ID persisted in the `health.json`**:
it means that it's an attempt to upgrade from unhealthy system which is unsupported -
if admin wants to perform such action, they should either get system to a healthy state,
or (if the MicroShift is the culprit) clear MicroShift data to start new
(and also get system to a healthy state).
In such case MicroShift will exit with an error that should render system unhealthy and eventually roll back.

> Extra comment:
>
> I think that on 2nd boot we'll lose that information resulting in different route:
> `health.json`.deployment_id == current deployment, health is `unhealthy`. 
> no backup to restore (neither current nor rollback deployment),
> so MicroShift will delete the data to start fresh.
> I think this might be confusing and unexpected (especially old data will be deleted).
> See [Test idea: upgrading from unhealthy system is blocked](#test-idea-upgrading-from-unhealthy-system-is-blocked)

**If deployment ID in `health.json` matches current deployment and backup for rollback deployment exists**:
restore and continue start up.
When upgrading the system fails during first boot of new deployement and renders system unhealthy,
on subsequent boots there is no backup for new deployment that could be restored.
Also, during subsequent boots MicroShift will consider the data unhealthy
(because that's the status in the `health.json`).

To gracefully handle such scenario, MicroShift will try to restore a backup
of the rollback deployment. This means that subsequent boots of new deployment
will have the same starting point as the first boot
(data belongs to previous deployment and is healthy).

> Extra comment:
> 
> This behaviour was previously designed so in case of failure to migrate the storage.
> It gives another chance to perform the migration starting from the initial state
> (that is starting with the data of previous deployment;
> starting just like it's first boot of new deployment).
>
> Since migration happens when full cluster is up and running,
> we might need to verify if this approach is still valid:
> - Does failure to migrate some CR will affect MicroShift's and in consequence system's health?

**If there is no backup for the rollback deployment**: remove the data and start fresh.
No backup of MicroShift data for rollback deployment might happen when rollback deployment
didn't feature MicroShift at all.
There is no backup to restore and current deployment is unhealthy -
we are not sure if MicroShift is the culprit, but being so early into
"MicroShift's lifetime" on the host we are not risking much when removing the data.
It's also a symmetrical with "restoring rollback deployment's backup"
where the data is restored so subsequent boots would start from the same point as first boot.

**If deployment ID in `health.json` doesn't match neither the rollback nor currently booted deployment**:
make an unhealthy backup and remove the data to allow for fresh start.


#### More details on backing up and restoring the data

Unhealthy backups were mentioned couple of times.
It must be noted that, when restoring the data, they are not considered a
candidates for the restore.

MicroShift will try its best to only keep one backup for particular deployment.
It means that before backing up, a list of backup for the deployment is obtained
and, after backing up, the list is used for deleting
"prior existing backups for the deployment that we
no longer need since we just backed up most up to date data".
This includes unhealthy backups.

If a particular backup (i.e. `deploymentID_bootID` combination),
there will be no new attempt to create a backup with that name again.
This is not considered an error - MicroShift will continue start up.

After successfully backing up the data, MicroShift compares list of
existing backups with deployments present on the system and removes any backup
which deployment is no longer present on the system.

### Version metadata management

This stage is all about comparing executable's and data's versions
to allow or block an upgrade and, if allowed, updating data version.

#### Version of the executable (`microshift` binary)

Version of executable is obtained from values embedded during built time:
```
# Makefile
 -X github.com/openshift/microshift/pkg/version.majorFromGit=$(MAJOR) \
 -X github.com/openshift/microshift/pkg/version.minorFromGit=$(MINOR) \
 -X github.com/openshift/microshift/pkg/version.patchFromGit=$(PATCH) \
```

> Hint:
>
> Values in Makefile can be overridden which was used to create
> "fake-next-minor" RPMs and commit - which is current code with
> newer minor version used in some tests to verify that version
> management works as expected.

#### Version of the data

Version of the data is loaded from `/var/lib/microshift/version` file.
If data exists, but the file does not, MicroShift assumes version
of the data is 4.13.

If real (from file) or assumed (4.13.0) data version is known, it will be
compared against version of the executable.
If data does not exist, MicroShift skips the checks and
proceeds to creation of the file.

#### Version compatibility

Following procedures compares the two versions making sure that:
- major versions are the same
- version of the data is not newer than version of the executable (e.g. downgrade)
- if executable is newer, then it's only by one minor version

#### Blocking certain upgrade paths

Executable and data versions are compared against a list of "blocked upgrades".
The list does not exist yet, but it is expected to be placed in 
`assets/release/upgrade-blocks.json` file with following schema:

```json
{
  "to-binary-version": ["blocked", "from", "data", "versions"],
  "to-binary-version-2": ["blocked", "from", "data", "versions"]
}
```
For example:
```json
{
		"4.14.10": ["4.14.5", "4.14.6"],
		"4.15.5":  ["4.15.2"]
}
```

Mechanism searches for the "top-level" version that matches executable's version
and, if one is found, checks if version of the data is present in the associated
list of version from which upgrade is blocked.
For example (using json data from above), if executable's version if "4.14.10"
and `version` file contains "4.14.5", then MicroShift will refuse to run with 
an error: "upgrade from '4.14.5' to '4.14.10' is blocked".

#### Updating `/var/lib/microshift/version`

If the data does not exist (i.e. it is a first run of MicroShift), then
file is created with version of the executable.

If file existed and version checks were successful, the file is updated 
with version of the executable.

### Kubernetes Storage Migration

Storage migration is process of updating objects to newer versions,
for example from `v1beta1` to `v1beta2`.

Initially, the plan was to perform the migration with only etcd and 
kube-apiserver running.
This turned out to not be a viable options because migrating
CustomResourceDefinitions might require webhooks which run as a Pods, 
therefore requiring kubelet, CNI, DNS, scheduler, and other components.

Right now, storage migration is comprised of an additional controller embedded
into MicroShift: Kube Storage Version Migrator. It makes use of the
`StorageVersionMigration` CR to update a given resource by its group and
version.

### `StorageVersionMigration` Custom Resources

This is the main API used by the Migrator Controller. It is a Custom Resource
that is created manually by a user or another controller and picked up by the
Migrator Controller. It specifies which resource to upgrade and to which
version, the progress of the migration is posted to the
`StorageVersionMigration` status.

#### Kube Storage Version Migrator

Migrator watches for `StorageVersionMigration` CRs and for each
will attempt to migrate related Kubernetes object.

The very core of the Migrator's implementation is simply: `GET` resource from
the API Server and `PUT` it back.
(Almost) all the heavy lifting is done by the API Server.

If schema of a CustomResourceDefinition changed substantially and requires custom logic
to perform the conversion between the versions, a conversion webhook needs to be provided.
If there are no schema changes between versions, API Server should handle
the conversion by itself.

For more information on migrating CRDs see 
[Versions in CustomResourceDefinitions](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definition-versioning/)
from Kubernetes documentation.

### Manual backup management (non-ostree)

Automated backup management is intended only for ostree-based systems.
For other, "regular RPM-based", systems `microshift` binary exposes following
commands aiming to help with creating and restoring backups of the MicroShift data:
- `microshift backup`
- `microshift restore`

Both commands require MicroShift to be stopped (to avoid corrupting data
if etcd was running during the process and modifying the files).

It is possible to use `microshift restore` command if `microshift.service`'s
status is `failed`, however the `microshift backup` will refuse to proceed.
This is based on assumption that `failed` means the MicroShift stopped running 
due to some runtime error and systemd gave up on restarting the service.
This suggests that MicroShift's data might not be healthy and thus should not
be backed up (mere existence of a backup suggests it contains valid data,
just like it is on ostree-based system with some exceptions that are explicitly
marked).

##### `microshift backup`

Command creates a backup of current MicroShift data (`/var/lib/microshift`).
It expects a full path of new backup directory - there is no default value.
The directory must not exist - MicroShift will create it.

```sh
$ sudo microshift backup /var/lib/microshift-backups/my-manual-backup
# or
$ sudo microshift backup /mnt/other-backups-location/another-manual-backup
```

##### `microshift restore`

Command to restore a MicroShift backup.
It expects a full path to an existing backup directory.

```sh
$ sudo microshift restore /var/lib/microshift-backups/my-manual-backup
# or
$ sudo microshift restore /mnt/other-backups-location/another-manual-backup
```

> Note: this command does not verify if provided path is really MicroShift's
> backup, so practically it could be used to copy any directory into
> `/var/lib/microshift`.
