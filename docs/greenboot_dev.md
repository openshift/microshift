# Testing MicroShift Integration with Greenboot

## Motivation

[Integrating MicroShift with Greenboot](./greenboot.md) allows for automatic
software upgrade rollbacks in case of a failure. The current document describes
a few techniques for simulating software upgrade failures in a development
environment. These guidelines can be used by developers for implementing CI/CD
pipelines testing MicroShift integration with Greenboot.

## MicroShift Service Failure

To simulate a situation with the MicroShift service failure after an upgrade,
one can remove the `hostname` RPM package used by MicroShift during its startup.

Run the following command to remove the package and reboot.

```bash
sudo rpm-ostree override remove hostname -r
```

Examine the system log to monitor the upgrade verification procedure that will fail
and reboot a few times before rolling back into the previous state. Note the values
of `boot_success` and `boot_counter` GRUB variables.

```
$ sudo journalctl -o cat -u greenboot-healthcheck.service
...
...
Running Required Health Check Scripts...
STARTED
GRUB boot variables:
boot_success=0
boot_indeterminate=0
boot_counter=2
...
...
Waiting 300s for MicroShift service to be active and not failed
FAILURE
...
...
```

After a few failed verification attempts, the system should roll back into the
previous state. Use the `rpm-ostree` command to verify the current deployment.

```bash
$ rpm-ostree status
State: idle
Deployments:
  edge:rhel/8/x86_64/edge
                  Version: 8.7 (2022-12-26T10:28:32Z)
                     Diff: 1 removed
      RemovedBasePackages: hostname 3.20-6.el8

* edge:rhel/8/x86_64/edge
                  Version: 8.7 (2022-12-26T10:28:32Z)
```

Finish by checking that all MicroShift pods run normally and cleaning up
the failed rollback deployment.

```bash
sudo rpm-ostree cleanup -b -r
```

## MicroShift Pod Failure

To simulate a situation with the MicroShift pod failure after an upgrade, 
one can set the `network.serviceNetwork` MicroShift configuration option to a
non-default `10.66.0.0/16` value without resetting the MicroShift data at the
`/var/lib/microshift` directory.

Start by checking out the current file system state and modifying it by adding
the `config.yaml` file to its `usr/etc/microshift` directory.

```bash
NEW_BRANCH="microshift-config"
OSTREE_REF=$(ostree refs | head -1)
OSTREE_COM=$(ostree log ${OSTREE_REF} | grep ^commit | awk '{print $2}')

sudo ostree checkout ${OSTREE_COM} ${NEW_BRANCH}
pushd ${NEW_BRANCH}

sudo tee usr/etc/microshift/config.yaml &>/dev/null <<EOF
network:
  serviceNetwork:
  - 10.66.0.0/16
EOF
```

Commit the updated file system, clean up the checked out tree and compare the
base reference with the newly created one to verify that the `config.yaml` file
was added at the `/usr/etc/microshift` directory.

```bash
sudo ostree commit --subject="MicroShift config.yaml update" --branch="${NEW_BRANCH}"
popd && sudo rm -rf ${NEW_BRANCH}

ostree diff ${OSTREE_REF} ${NEW_BRANCH}
```

Switch to the new branch as the default deployment and reboot for changes to
become effective.

```bash
BRANCH_COM=$(ostree log ${NEW_BRANCH} | grep ^commit | awk '{print $2}')
sudo rpm-ostree rebase --branch=${NEW_BRANCH} ${BRANCH_COM} -r
```

Examine the system log to monitor the upgrade verification procedure that will fail
and reboot a few times before rolling back into the previous state. Note the pod
readiness failure in the `openshift-ingress` namespace.

```
$ sudo journalctl -o cat -u greenboot-healthcheck.service
...
...
Running Required Health Check Scripts...
STARTED
...
...
Waiting 300s for 1 pod(s) from the 'openshift-ingress' namespace to be in 'Ready' state
FAILURE
...
...
```

After a few failed verification attempts, the system should roll back into the
previous state. Use the `rpm-ostree` command to verify the current deployment.

```bash
$ rpm-ostree status
State: idle
Deployments:
* edge:rhel/8/x86_64/edge
                  Version: 8.7 (2022-12-28T16:50:54Z)

  edge:eae8486a204bd72eb56ac35ca9c911a46aff3c68e83855f377ae36a3ea4e87ef
                Timestamp: 2022-12-29T14:44:48Z
```

Finish by checking that all MicroShift pods run normally and cleaning up
the failed rollback deployment.

```bash
NEW_BRANCH="microshift-config"
sudo rpm-ostree cleanup -b -r
sudo ostree refs --delete ${NEW_BRANCH}
```
