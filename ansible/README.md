# Ansible MicroShift

The purpose of this Ansible automation is to help gain insight into the footprint and start times of MicroShift.
At present, we are capturing the first start, the second start (with cached images), disk usage data from a
number of points in time as well as performance data for each of the two starts.

## Prerequisites

We are using [Prometheus](https://prometheus.io/), [process-exporter](https://github.com/ncabatoff/process-exporter) and [cAdvisor](https://github.com/google/cadvisor) to capture a wide array of performance tool data.

It is necessary to have two hosts configured for running the benchmarks:
* Ansible server used to start the automation scripts
* MicroShift server used to execute the performance tests

Run the following command on the Ansible server to install the Ansible package.
```
sudo dnf install -y ansible
```

## Running

Follow the instructions below depending on whether the MicroShift server is set up as a clean, development or RHEL for Edge host.

### Clean Host

If the user has a fresh RHEL host to be used for running MicroShift performance benchmarks, it is necessary to first follow the instructions in [Clean Host Example Variables](#clean-host-example-variables) for editing the `vars/all.yml` file.

Run the playbook using the following commands, making sure that the values of the MicroShift server host and user variables are appropriate for your environment.
```bash
USHIFT_HOST=microshift-clean
USHIFT_USER=microshift

ssh-copy-id ${USHIFT_USER}@${USHIFT_HOST}
time ansible-playbook -v \
    -e ansible_host_var=${USHIFT_HOST} -e ansible_user_var=${USHIFT_USER} \
    -i inventory/inventory run-perf.yml
```

### Development Host

If the user has an existing host used as the MicroShift development environment, the default settings in the `vars/all.yml` file can be used.
> Review the instructions in [Development Host Example Variables](#development-host-example-variables) for more information on the configuration settings.

Run the playbook using the following commands, making sure that the values of the MicroShift server host and user variables are appropriate for your environment.
```bash
USHIFT_HOST=microshift-dev
USHIFT_USER=microshift

ssh-copy-id ${USHIFT_USER}@${USHIFT_HOST}
time ansible-playbook -v \
    -e ansible_host_var=${USHIFT_HOST} -e ansible_user_var=${USHIFT_USER} \
    -e install_microshift_arg=true \
    -i inventory/inventory run-perf.yml
```

### RHEL for Edge Host

If the user has deployed a RHEL for Edge image built with MicroShift, the default settings in the `vars/all.yml` file can be used. The image must be build with `-prometheus` option to enable benchmark information collection as described in the [Building Installer](../docs/rhel4edge_iso.md#building-installer) section.
> Review the instructions in [RHEL for Edge Host Example Variables](#rhel-for-edge-host-example-variables) for more information on the configuration settings.

Run the playbook using the following commands, making sure that the values of the MicroShift server host and user variables are appropriate for your environment.
```bash
USHIFT_HOST=microshift-edge
USHIFT_USER=redhat

ssh-copy-id ${USHIFT_USER}@${USHIFT_HOST}
time ansible-playbook -v \
    -e ansible_host_var=${USHIFT_HOST} -e ansible_user_var=${USHIFT_USER} \
    -e prometheus_logging_arg=false \
    -i inventory/inventory run-perf.yml
```

## Output

The following text files will be created on the Ansible server running the playbook:
- The `boot.txt` will have the time that it took the microshift service to start and for all the pods to enter the `Running` state
- The `disk0.txt, disk1.txt, disk2.txt` files will have a snapshot of the disk usage at different stages of installation and execution.

If Prometheus was enabled, the performance metrics will be uploaded to the Ansible server, where users can view all the captured performance data using the `http://<ansible-server-ip>:9091` URL. These metrics can be conveniently visualized in the Grafana dashboards accessible at the `http://<ansible-server-ip>:3000` site.
> Use `admin:admin` credentials to log into the Grafana server.

## Configuration Overview

There are a few configuration files that can be configured before execution of the playbook.
>Certain Ansible playbook variables can be overriden using the `--extra-args` or `-e` command line argument.

### Variables

The following variables found in `vars/all.yml` may need user configuration.

| Variable Name  | Description | Default |
| -------------- | ----------- | ------- |
| `manage_subscription` | Use subscription-manager to entitle host and attach to pool | `false` |
| `rhel_username` | Red Hat subscription account username (string) | `null` | 
| `rhel_password` | Red Hat subscription account password (string) | `null` | 
| `rhel_pool_id` | Red Hat subscription pool id (string) | `null` | 
| `manage_repos` | Enable necessary repos to build MicroShift and install necessary components | `false` |
| `setup_microshift_host` | Complete initial setup of MicroShift host (subscription, packages, firewall, etc) | `false` |
| `prometheus_logging` | Set up logging and exporters on the nodes | `true` |
| `install_microshift` | Build and install MicroShift from git | `false` |
| `build_etcd_binary` | Build and deploy a separate etcd process | `false` |

### Inventory Configuration

The inventory configuration file is located at `inventory/inventory`.

The MicroShift host name or IP address must be configured to match the current environment. If the user account available on the MicroShift server is not `microshift`, it should be changed as well.

> The following Ansible command line options can be used instead of editing the `inventory/inventory` file.
> - `-e ansible_host_var=<hostname>`
> - `-e ansible_user_var=<username>`
> - `-e private_ip_var=<ipaddress>`

**Sample Inventory File**
```
[microshift]
microshift-dev ansible_host=microshift-dev private_ip=

[microshift:vars]
ansible_user=microshift

[logging]
localhost ansible_connection=local
```

### Global Variables Configuration

The global variables file is located at `vars/all.yml`.
> The following Ansible command line options can be used instead of editing the `vars/all.yml` file.
> - `-e install_microshift_arg=<true | false>`
> - `-e prometheus_logging_arg=<true | false>`

#### Clean Host Example Variables

If the user has a fresh RHEL host to be used for running MicroShift performance benchmarks, the scripts can manage the initial host setup and configuration in addition to the performance capture.

As we can read from the following configuration, we have selected to `manage_subscription` and this requires the subsequent `rhel_*` vars to be set.
The other initial configuration steps are toggled with the `setup_microshift_host` variable, which we have set to `true` for this configuration.

On a clean system, the playbook will also copy the user-provided `pull-secret` from `roles/install-microshift/files/pull-secret.txt` to the correct location on the host.

**Sample vars/all.yml File**
```
manage_subscription: true
rhel_username: <subscription manager username>
rhel_password: <subscription manager password>
rhel_pool_id: <subscription manager pool-id>
manage_repos: true

setup_microshift_host: true
prometheus_logging: true
install_microshift: true
build_etcd_binary: false
```

#### Development Host Example Variables

If the user has an existing host used as the MicroShift development environment, the scripts can be run without the initial configuration steps.
> Such host can be created using the instructions from the [MicroShift Development Environment on RHEL 8](../docs/devenv_rhel8.md) document.

The `manage_subscription`, `manage_repos` and `setup_microshift_host` variables have been set to `false`.

**Sample vars/all.yml File**
```
manage_subscription: false
rhel_username:
rhel_password:
rhel_pool_id:
manage_repos: false

setup_microshift_host: false
prometheus_logging: true
install_microshift: true
build_etcd_binary: false
```

#### RHEL for Edge Host Example Variables

If the user has deployed a RHEL for Edge image built with MicroShift, the scripts can be run with a few more variables disabled as the host is fully configured out of the box.
> Such host can be created using the instructions from the [Install MicroShift on RHEL for Edge](../docs/rhel4edge_iso.md) document.

The `manage_subscription`, `manage_repos`, `setup_microshift_host`, `prometheus_logging` and `install_microshift` variables have been set to `false`.

**Sample vars/all.yml File**
```
manage_subscription: false
rhel_username:
rhel_password:
rhel_pool_id:
manage_repos: false

setup_microshift_host: false
prometheus_logging: false
install_microshift: false
build_etcd_binary: false
```
