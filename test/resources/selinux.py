from robot.libraries.BuiltIn import BuiltIn
from libostree import remote_sudo_rc, remote_sudo
from typing import List

ACCESS_CHECK_MAP = {
    "/var/lib/microshift/version": ["cat"],
    "/etc/microshift/config.yaml.default": ["cat"],
    "/etc/microshift/lvmd.yaml.default": ["cat"],
    "/etc/microshift/ovn.yaml.default": ["cat"],
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
        "/etc/microshift/manifests",
        "/etc/microshift/manifests.d",
    ],
}

# This list should only ever change if we alter the SELinux policy or
# upstream container linux package changes something in these contexts,
# those events should be rare. However, if anything changes these contexts, the test should
# fail so we can decide what that means for MicroShift and update the list then.
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


# Here we care about matching what our SELinux policy says with what the host says for contexts.
# The contexts that effect us and OCP are `kubernetes_file_t|container_var_lib_t|kubelet_exec_t|container_t`
# we query and filter for those contexts to validate against our expected list.
def get_fcontext_list() -> List[str]:
    context_list = "kubernetes_file_t|container_var_lib_t|kubelet_exec_t|container_t"
    semanage_filter_cmd = f"semanage fcontext -l | grep -E  \"({context_list})\" | awk '{{print $1 }}'"
    return remote_sudo(semanage_filter_cmd).splitlines()


def get_denial_audit_log() -> List[str]:
    ausearch_filter_cmd = "ausearch --input-logs -m avc | grep microshift"
    stdout, rc = remote_sudo_rc(ausearch_filter_cmd)
    if rc == 0:
        return stdout.splitlines()
    return []


# In order to check access, we use the `runcon` command to initiate shell commands.
# We use the user and role of `system_u` and `system_r` with a context type of `container_t`
# because that is the context under which the container runtime will execute containers.
# So here we simulate a container that has broken out trying to access files on the host.
def run_access_check(access_check_map: dict[str, List[str]]) -> List[str]:
    runcon_cmd = "runcon -u system_u -r system_r -t container_t"
    allowed_access = []
    for file_path, commands in access_check_map.items():
        for command in commands:
            stdout, rc = remote_sudo_rc(f"{runcon_cmd} {command} {file_path} 2>&1")
            BuiltIn().should_not_match(stdout, "*No such file or directory*")
            if rc == 0:
                allowed_access.append(f"should not have been allowed access to {file_path} by running {command}")

    return allowed_access


# Gather a list of all files in the specified directory and call `run_access_check` to check `cat` access
def run_access_check_on_dir(directory: str) -> List[str]:
    du_ls_cmd = "sudo du -a"
    stdout, rc = remote_sudo_rc(f"{du_ls_cmd} {directory} | awk '{{print $2 }}'")
    BuiltIn().should_not_be_empty(stdout)
    BuiltIn().should_be_equal_as_integers(rc, 0)
    list_of_paths = stdout.splitlines()
    directory_dict = {file_path: ["cat"] for file_path in list_of_paths if "." in file_path[len(directory):]}
    return run_access_check(directory_dict)


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
            if context_do_not_match(stdout, context):
                incorrect_fcontext.append(f"expected {file_path} to have context of ({context}) but got ({stdout})")

    return incorrect_fcontext


def context_do_not_match(context_string: str, expected_context_string: str) -> bool:
    context_parts = context_string.split(":")
    context_only = ":".join(context_parts[1:])

    expected_context_parts = expected_context_string.split(":")
    expected_context_only = ":".join(expected_context_parts[1:])

    # Since in CI we sometimes manually create folders, the user context might be
    # unconfined_u instead of system_u. That is valid within SELinux practices, so if the context
    # don't match and they don't match with out the user, then we fail.
    if context_string != expected_context_string and context_only != expected_context_only:
        return True

    # If context do match but it's not a system_u and an unconfined_u user, that's not expected and should be looked at.
    if context_parts[0] != "system_u" and context_parts[0] != "unconfined_u":
        return True

    return False
