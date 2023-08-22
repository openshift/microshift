from robot.libraries.BuiltIn import BuiltIn
from libostree import remote_sudo_rc, remote_sudo
from typing import List

ACCESS_CHECK_MAP = {
    "/var/lib/microshift/version": ["cat"],
    "/etc/microshift/config.yaml.default": ["cat"],
    "/var/lib/microshift-backups/health.json": ["cat"],
}

CONTEXT_CHECK_MAP = {
    "system_u:object_r:container_var_lib_t:s0": [
        "/var/lib/microshift",
        "/var/lib/microshift-backups",
    ],
    "system_u:object_r:kubelet_exec_t:s0": [
        "/usr/bin/microshift",
        "/usr/bin/microshift-etcd",
    ],
    "system_u:object_r:kubernetes_file_t:s0": [
        "/etc/microshift",
    ],
}

EXPECTED_FCONTEXT_LIST = [
    "/etc/kubernetes(/.*)?",
    "/etc/microshift(/.*)?",
    "/exports(/.*)?",
    "/usr/bin/microshift",
    "/usr/bin/microshift-etcd",
    "/usr/lib/microshift(/.*)?",
    "/usr/local/bin/microshift",
    "/usr/local/bin/microshift-etcd",
    "/usr/local/s?bin/hyperkube.*",
    "/usr/local/s?bin/kubelet.*",
    "/usr/s?bin/hyperkube.*",
    "/usr/s?bin/kubelet.*",
    "/var/lib/buildkit(/.*)?",
    "/var/lib/cni(/.*)?",
    "/var/lib/containerd(/.*)?",
    "/var/lib/containers(/.*)?",
    "/var/lib/docker(/.*)?",
    "/var/lib/docker-latest(/.*)?",
    "/var/lib/kubelet(/.*)?",
    "/var/lib/lxc(/.*)?",
    "/var/lib/lxd(/.*)?",
    "/var/lib/microshift(/.*)?",
    "/var/lib/microshift-backups(/.*)?",
    "/var/lib/microshift.saved(/.*)?",
    "/var/lib/ocid(/.*)?",
    "/var/lib/registry(/.*)?",
]


def get_expected_ocp_microshift_fcontext_list() -> List[str]:
    return EXPECTED_FCONTEXT_LIST


def get_fcontext_list() -> List[str]:
    context_list = "kubernetes_file_t|container_var_lib_t|kubelet_exec_t|container_t"
    semanage_filter_cmd = f"semanage fcontext -l | grep -E  \"({context_list})\" | awk '{{print $1 }}'"
    return remote_sudo(semanage_filter_cmd).split('\n')


def get_denial_audit_log() -> List[str]:
    ausearch_filter_cmd = "ausearch --input-logs -m avc | grep microshift"
    stdout, rc = remote_sudo_rc(ausearch_filter_cmd)
    if rc == 0:
        return stdout.split('\n')
    return []


def run_access_check(access_check_map: dict[str, List[str]]) -> List[str]:
    runcon_cmd = "runcon -u system_u -r system_r -t container_t"
    allowed_access = []
    for file_path, commands in access_check_map.items():
        for command in commands:
            stdout, rc = remote_sudo_rc(f"{runcon_cmd} {command} {file_path}")
            if rc == 0:
                allowed_access.append(f"should not have been allowed access to {file_path} by running {command}")

    return allowed_access


def run_default_access_check() -> List[str]:
    return run_access_check(ACCESS_CHECK_MAP)


def run_fcontext_check() -> List[str]:
    ls_cmd = "ls -Zd"
    incorrect_fcontext = []
    for context, file_paths in CONTEXT_CHECK_MAP.items():
        for file_path in file_paths:
            stdout, rc = remote_sudo_rc(f"{ls_cmd} {file_path} | awk '{{print $1 }}'")
            BuiltIn().should_not_be_empty(stdout)
            BuiltIn().should_be_equal_as_integers(rc, 0)
            if stdout != context:
                incorrect_fcontext.append(f"should have the same context as {context} but got {stdout}")

    return incorrect_fcontext
