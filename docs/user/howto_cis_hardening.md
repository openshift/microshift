# CIS Level 2 Hardening for MicroShift

This guide describes how to run MicroShift on a CIS Level 2 hardened RHEL system. It covers which CIS rules MicroShift affects, the required hardening playbook overrides, and how to verify compliance.

## MicroShift's CIS Impact

MicroShift introduces changes to a hardened system in three categories. These rules will fail on any system running Kubernetes containers and cannot be avoided without disabling core functionality.

### IP Forwarding

OVN-Kubernetes requires IP forwarding for pod networking. MicroShift enables `net.ipv4.ip_forward` and `net.ipv6.conf.all.forwarding` at startup.

| Rule ID | OS |
|:--------|:---|
| `sysctl_net_ipv4_ip_forward` | RHEL 9, 10 |
| `sysctl_net_ipv6_conf_all_forwarding` | RHEL 9, 10 |
| `sysctl_net_ipv4_conf_all_forwarding` | RHEL 10 |
| `sysctl_net_ipv4_conf_default_forwarding` | RHEL 10 |
| `sysctl_net_ipv6_conf_default_forwarding` | RHEL 10 |

### Container File Ownership

Container images contain files owned by UIDs and GIDs that do not exist in the host's `/etc/passwd` and `/etc/group`. These are stored in `/var/lib/containers/storage/overlay/` and are an inherent characteristic of any container runtime.

| Rule ID | OS |
|:--------|:---|
| `file_permissions_ungroupowned` | RHEL 9 |
| `no_files_unowned_by_user` | RHEL 9 |
| `no_files_or_dirs_ungroupowned` | RHEL 10 |
| `no_files_or_dirs_unowned_by_user` | RHEL 10 |

### Kubelet Volume Permissions

Kubelet creates 0777 directories for ConfigMap and EmptyDir volumes so that pods running as any UID can access their projected configuration. It also creates 0666 container termination log files.

| Rule ID | OS |
|:--------|:---|
| `dir_perms_world_writable_sticky_bits` | RHEL 9, 10 |
| `file_permissions_unauthorized_world_writable` | RHEL 9, 10 |

### Audit Rules

CIS requires audit rules for all setuid/setgid binaries. If MicroShift is installed after running the CIS hardening role, the audit rules will not cover binaries added by MicroShift RPMs.

| Rule ID | OS |
|:--------|:---|
| `audit_rules_privileged_commands` | RHEL 9, 10 |

This rule can be resolved by regenerating audit rules after installing MicroShift:

```bash
sudo find / -xdev \( -perm -4000 -o -perm -2000 \) -type f 2>/dev/null | \
  awk '{print "-a always,exit -F path=" $1 " -F perm=x -F auid>=1000 -F auid!=unset -k privileged"}' | \
  sudo tee /etc/audit/rules.d/99-privileged-post-install.rules > /dev/null
sudo augenrules --load
```

## Hardening Playbook Overrides

When using the RedHatOfficial CIS ansible roles, the hardening playbook must override the IP forwarding variables to prevent the role from disabling forwarding before MicroShift starts.

For RHEL 9:

```yaml
---
- name: Apply CIS Level 2 hardening
  hosts: localhost
  connection: local
  become: true
  roles:
    - role: ansible-role-rhel9-cis
      vars:
        sysctl_net_ipv4_ip_forward: false
        sysctl_net_ipv6_conf_all_forwarding: false
```

For RHEL 10, the CIS benchmark adds per-interface forwarding checks:

```yaml
---
- name: Apply CIS Level 2 hardening
  hosts: localhost
  connection: local
  become: true
  roles:
    - role: ansible-role-rhel10-cis
      vars:
        sysctl_net_ipv4_ip_forward: false
        sysctl_net_ipv4_conf_all_forwarding: false
        sysctl_net_ipv4_conf_default_forwarding: false
        sysctl_net_ipv6_conf_all_forwarding: false
        sysctl_net_ipv6_conf_default_forwarding: false
```

> Each variable is set to `false` to tell the role **not** to disable forwarding. This does not enable forwarding — MicroShift does that at startup.

## Firewall Configuration

CIS hardening enables `firewalld` and locks down all traffic. MicroShift requires several ports and trusted network sources. See [Firewall Configuration](howto_firewall.md) for the full list.

At minimum, the pod and service networks must be trusted:

```bash
sudo firewall-cmd --permanent --zone=trusted --add-source=10.42.0.0/16
sudo firewall-cmd --permanent --zone=trusted --add-source=169.254.169.1
sudo firewall-cmd --permanent --zone=trusted --add-source=fd01::/48
sudo firewall-cmd --permanent --zone=public --add-port=6443/tcp
sudo firewall-cmd --reload
```

## References

- [RedHatOfficial ansible-role-rhel9-cis](https://github.com/RedHatOfficial/ansible-role-rhel9-cis)
- [RedHatOfficial ansible-role-rhel10-cis](https://github.com/RedHatOfficial/ansible-role-rhel10-cis)
- [OpenSCAP Project](https://www.open-scap.org/)
- [CIS Benchmarks](https://www.cisecurity.org/cis-benchmarks)
- [Firewall Configuration](howto_firewall.md)
