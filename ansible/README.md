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
| `install_pbench` | Install pbench performance capture tooling (boolean) | `false` |
| `pbench_record_duration` | Duration of time for pbench recording (seconds) | `360` |

## Running

To run this playbook:
```
time ansible-playbook -v -i inventory/inventory setup-node.yml
```
