# ansible-microshift

The purpose of this ansible automation is to help gain insight into the footprint and start times of MicroShift.
At this stage we are capturing the first start, the second start (with cached images), disk usage data from a
number of points in time as well as performance tool data with pbench for each of the two starts.

## Prereqs

We are using [pbench](https://github.com/distributed-system-analysis/pbench) to capture a wide array of performance tool data.
Pbench has Ansible playbooks available in Ansible Galaxy to install the pbench-agent on the MicroShift host.

- Ansible:
  - On Fedora: `dnf install ansible`

- Pbench agent installation role from ansible-galaxy:
  - `ansible-galaxy collection install pbench.agent`

Note: ensure your ansible galaxy user directory is exported:
```
export ANSIBLE_ROLES_PATH=$HOME/.ansible/collections/ansible_collections/pbench/agent/roles:$ANSIBLE_ROLES_PATH
```

## Vars

There are a few variables that may need user configuration and they are found in `vars/all.yml`.

| Variable Name  | Description | Default |
| -------------- | ----------- | ------- |
| `manage_subscription` | Use subscription-manager to entitle host and attach to pool | `false` |
| `rhel_username` | Red Hat subscription account username (string) | `null` | 
| `rhel_password` | Red Hat subscription account password (string) | `null` | 
| `rhel_pool_id` | Red Hat subscription pool id (string) | `null` | 
| `prometheus_logging` | Set up cadvisor and node-exporter on microshift node | `false` |
| `install_pbench` | Install pbench performance capture tooling (boolean) | `true` |
| `pbench_record_duration` | Duration of time for pbench recording (seconds) | `360` |
| `setup_microshift_host` | Complete initial setup of VM (subscription, packages, firewall, etc) | `false` |
| `build_etcd_binary` | Build and deploy a separate etcd process | `false` |


## Running

There are a few places to configure before execution.

As with any Ansible playbook the first file to configure is the inventory, located at `inventory/inventory`.


### Inventory Configuration

**Sample inventory file:**
```
[microshift]
192.168.122.48

[microshift:vars]
ansible_user=microshift

[logging]

[logging:vars]
ansible_user=centos
ansible_become=yes
```

The user needs to configure the microshift host IP address to match their environment, also if the configured
user on the node is not `microshift` that should be changed as well.

### Global Variables Configuration

The next location for configuration will be the global vars file, located at `vars/all.yml`.

#### Clean VM Example Vars

If the user has created a fresh RHEL host/VM and the scripts can take manage initial host setup and configuration in addition to the performance capture.
As we can read from this configuration we have selected to `manage_subscription` and this requires the subsequent `rhel_*` vars to be set.
The other initial configuration steps are toggled with the `setup_microshift_host` variable, whivh we have set to `true` for this config.

On a clean system the playbook will also copy the user provided `pull-secret` from `roles/install-microshift/files/pull-secret.txt` to the correct location
on the host.

**Sample vars file: (clean VM)**
```
manage_subscription: true
rhel_username: <subscription manager username>
rhel_password: <subscription manager password>
rhel_pool_id: <subscription manager pool-id>

prometheus_logging: false

install_pbench: true
pbench_configuration_environment: <pbench environment>
pbench_key_url: <location of pbench ssh key>
pbench_config_url: <location of pbench configuration file>
pbench_record_duration: 360

setup_microshift_host: true
build_etcd_binary: false
```

#### Existing Pre-configured VM Example vars

If the user has an existing and already configured host/VM the scripts can be run without the initial configuration steps.
Two variables have been set to `false`: `manage_subscription` and `setup_microshift_host`.
We still want to `install_pbench` which is the tool we are using to collect performance metrics.
The `pbench_*` variables are required, and the URLs will depend on the user environment.

**Sample vars file: (configured VM)**
```
manage_subscription: false
rhel_username:
rhel_password:
rhel_pool_id:

prometheus_logging: false

install_pbench: true
pbench_configuration_environment: <pbench environment>
pbench_key_url: <location of pbench ssh key>
pbench_config_url: <location of pbench configuration file>
pbench_record_duration: 360

setup_microshift_host: false
build_etcd_binary: false
```


### Run playbook

Once the `inventory` has been edited and the `vars` have been configured we can now run the playbook:

```
time ansible-playbook -v -i inventory/inventory run-perf.yml
```

### Output

There are a few text files that will be created on the local system running the playbook: `boot.txt, disk0.txt, disk1.txt, disk2.txt`.
The `boot.txt` will have the time that it took the microshift service to start and for all the pods to enter the `Running` state.
The three `disk*.txt` files will have a snapshot of the disk usage at different stages of installation & execution.
The performance metrics (if pbench was enabled) will be uploaded to the pbench server as configured, where the user can view all the captured performance data.
