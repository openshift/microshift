# Auto-recovery from manual backups

The auto-recovery from manual backups feature provides a building blocks for automatically restoring backups when MicroShift fails to start.

The feature introduces new command-line options for the existing `backup` and `restore` commands:
- `--auto-recovery`: Changes the behavior of both commands.
  The `PATH` argument is no longer treated as final backup path, instead it's treated as directory holding all backups for auto-recovery.
- `--dont-save-failed` (only `restore`): Opt-out from backing up failed MicroShift data.

## Creating backups

To create backups compatible with the auto-recovery feature, use the following command, specifying `--auto-recovery` option
and a destination where backups will be stored. For example:
```
$ sudo microshift backup --auto-recovery /var/lib/microshift-auto-recovery
```

Backups are created with predefined schema recognized by the restore command:
- For ostree/bootc systems: `$dateTime_deploymentId`
- For RPM systems: `$dateTime_microshiftVersion`

## Restore backups

To restore a backup, run following command:
```
$ sudo microshift restore --auto-recovery /var/lib/microshift-auto-recovery
```

The process will:
- Copy MicroShift data to `/var/lib/microshift-auto-recovery/failed/` for later investigation (opt-out using `--dont-save-failed`)
- Select the most recent, compatible backup and restore it.

The next time the command is executed, the previously restored backup will be moved to `/var/lib/microshift-auto-recovery/restored/`.

Note that the `restore --auto-recovery` command does not attempt to stop MicroShift.
It is assumed that when the command is executed, MicroShift service already failed or it is user's responsibility to stop it.

## User responsibilities

- Creating backups: Backups require stopping MIcroShift. Only the user can determine the best time to perform this.
- Restarting MicroShift: The `restore --auto-recovery` command does not start MicroShift after restoring; it is the responsibility of user automation.
- Disk space monitoring: MicroShift does not monitor the disk space of any filesystem. Users must ensure their automation handles old backup removal.

## Example of an automation - integration with systemd

Here's an example of how to automate this process using systemd for RPM systems where `/usr` is writable.
For ostree/bootc systems, it's recommended to include these changes in the blueprint/Containerfile (see the next section for Containerfile examples).

This method utilizes systemd's `OnFailure` functionality which specifies services to run if the service fails.
In this example, if `microshift.service` enters a failed state, systemd starts the `microshift-auto-recovery.service` unit,
executing the auto-recovery restore process and restarting MicroShift.

The automation includes a guard to stop while Greenboot is still testing the newly staged deployment - there is another automation in place for system rolling back.

1. Create a directory for the `microshift.service` drop-in:
   ```
   $ sudo mkdir -p /usr/lib/systemd/system/microshift.service.d
   ```
1. Create the `10-auto-recovery.conf` file to instruct systemd to run `microshift-auto-recovery.service` if `microshift.service` fails:
   ```
   sudo tee /usr/lib/systemd/system/microshift.service.d/10-auto-recovery.conf > /dev/null <<'EOF'
   [Unit]
   OnFailure=microshift-auto-recovery.service
   EOF
   ```
1. Create the `microshift-auto-recovery.service` file:
   ```
   sudo tee /usr/lib/systemd/system/microshift-auto-recovery.service > /dev/null <<'EOF'
   [Unit]
   Description=MicroShift auto-recovery
   
   [Service]
   Type=oneshot
   ExecStart=/usr/bin/microshift-auto-recovery
   
   [Install]
   WantedBy=multi-user.target
   EOF
   ```
1. Create the `microshift-auto-recovery` script:
   ```
   #!/usr/bin/env bash
   set -xeuo pipefail
   
   # If greenboot uses non-default file for clearing boot_counter, use boot_success instead.
   if grep -q  "/boot/grubenv" /usr/libexec/greenboot/greenboot-grub2-set-success; then
       if grub2-editenv - list | grep -q ^boot_success=0; then
           echo "Greenboot didn't decide the system is healthy after staging new deployment."
           echo "Quiting to not interfere with the process"
           exit 0
       fi
   else
       if grub2-editenv - list | grep -q ^boot_counter=; then
           echo "Greenboot didn't decide the system is healthy after staging a new deployment."
           echo "Quiting to not interfere with the process"
           exit 0
       fi
   fi
   
   /usr/bin/microshift restore --auto-recovery /var/lib/microshift-auto-recovery
   /usr/bin/systemctl reset-failed microshift
   /usr/bin/systemctl start microshift
   
   echo "DONE"
   ```
1. Make the script executable:
   ```
   sudo chmod +x /usr/bin/microshift-auto-recovery 
   ```
1. Reload the system configuration:
   ```
   sudo systemctl daemon-reload
   ```

### Containerfile example

For bootc systems, the following Containerfile example demostrates how to put necessary files in place:
```
RUN mkdir -p /usr/lib/systemd/system/microshift.service.d
COPY ./auto-rec/10-auto-recovery.conf /usr/lib/systemd/system/microshift.service.d/10-auto-recovery.conf
COPY ./auto-rec/microshift-auto-recovery.service /usr/lib/systemd/system/
COPY ./auto-rec/microshift-auto-recovery /usr/bin/
RUN chmod +x /usr/bin/microshift-auto-recovery && systemctl daemon-reload
```
