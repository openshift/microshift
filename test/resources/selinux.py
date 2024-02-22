import re
from typing import List

from robot.libraries.BuiltIn import BuiltIn

from libostree import remote_sudo, remote_sudo_rc

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
    "system_u:object_r:container_log_t:s0": [
        "/var/log/kube-apiserver",
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

SOURCE_TARGET_TRANSITION = {
    "container_t": ["container_var_lib_t"]
}

# Ignoring things that are common of less privileged and thus not a concern.
# https://github.com/fedora-selinux/selinux-policy/blob/b56ae8af271091a8eb8d1fe0421f4f8ffa151f2d/policy/modules/kernel/files.te#L216-L226
DOMAIN_PERMISSION_IGNORE_REGEX = {
    "^file_type [a-z_]+:filesystem$": ["associate"],
}


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


# We want to validate that when a binary that forks a process or runs as a privileged process, the SELinux transition
# for that binary can not gain access to specific files.
# Here we test that first the provided bin DOES have access, then we validate that when running under the
# container_t it no longer has access to those files.
def run_binary_domain_transition_check(test_bin: str, access_check_map: dict[str, List[str]]) -> List[str]:
    runcon_cmd = "runcon -u system_u -r system_r -t container_t"
    error_list = []

    # When running in the container_t context, there is no binary transition that will allow it access
    for file_path, commands in access_check_map.items():
        for command in commands:
            stdout_access, rc_access = remote_sudo_rc(f"{test_bin} {command} {file_path} 2>&1")
            BuiltIn().should_not_match(stdout_access, "*No such file or directory*")
            if rc_access != 0:
                error_list.append(f"test bin should have been allowed access to {file_path} when not under container_t")

            stdout_denied, rc_denied = remote_sudo_rc(f"{runcon_cmd} {test_bin} {command} {file_path} 2>&1")
            BuiltIn().should_not_match(stdout_denied, "*No such file or directory*")
            if rc_denied == 0:
                error_list.append(f"should not have been allowed access to {file_path} by running {command}")

    return error_list


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


# In order to test binary transition escapes, we give a test script the kubelet_exec_t context which has a
# kubelet_t entrypoint, meaning that kubelet_exec_t will allow a forked process a transition to a kubelet_t context.
# The kubelet_t context has access to the container_var_lib_t context.
def run_default_access_binary_transition_check(script_file_path: str) -> List[str]:
    chcon_cmd = "chcon -t kubelet_exec_t"
    chmod_cmd = "chmod +x"
    error_list = []
    stdout, rc = remote_sudo_rc(f"{chcon_cmd} {script_file_path} && {chmod_cmd} {script_file_path} 2>&1")
    if rc != 0:
        error_list.append(f"failed to chcon of test file {script_file_path} to kubelet_exec_t")
        return error_list
    return run_binary_domain_transition_check(script_file_path, ACCESS_CHECK_MAP)


# For a given root source, query transition process types to recursivly get all
# contexts
def get_all_traversal_transition_contexts(source: str) -> List[str]:
    sesearch_cmd = "sesearch --type_trans -c process -s"
    accum = []
    # Search for a process transition from the specified source
    stdout, rc = remote_sudo_rc(f"{sesearch_cmd} {source} | awk '{{ print $3 }}' 2>&1")
    BuiltIn().should_be_equal_as_integers(rc, 0)

    transition_contexts = stdout.splitlines()
    if len(transition_contexts) == 0:
        return accum

    for context in transition_contexts:
        source = context.split(":")[0]
        accum.append(source)
        result = get_all_traversal_transition_contexts(source)
        accum.extend(result)

    # Remove dups
    return list(set(accum))


# Check if we should ignore this specific domain check
def ignore_domain_check(source: str, key: str, domain_permissions: List[str]) -> bool:
    # Ignore self refernce to filesystem associatation
    found = re.search(f"{source} {source}:filesystem", key)
    if found:
        if len(domain_permissions) == 1 and "associate" in domain_permissions:
            print(f"match on {source} key {key}")
            return True

    # Iterate over ignore rules
    for reg, permissions in DOMAIN_PERMISSION_IGNORE_REGEX.items():
        found = re.search(reg, key)
        if found and all(perm in permissions for perm in domain_permissions):
            print(f"match on {reg} key {key}")
            return True
    return False


# Given a parent context and all of its possible transitional context sources check to make sure
# each one of the transitional sources does not contain permissions that are not expected or more
# than the parent permissions.
# i.e. a permission to a context that the parent does not have.
def sources_have_less_permission_than_parent(parent: str, all_sources: List[str]) -> List[str]:
    allow_cmd = "sesearch -S -A"
    semanage_cmd = "semanage fcontext -l"

    parent_number_of_permissions, rc = remote_sudo_rc(f"{allow_cmd} -s {parent} | wc -l")
    BuiltIn().should_be_equal_as_integers(rc, 0)
    error_list = []

    for source in all_sources:
        domain_context = source
        if "_exec_t" in domain_context:
            domain_context = source.replace('_exec', '')

        source_number_of_permissions, rc = remote_sudo_rc(f"{allow_cmd} -s {domain_context} | wc -l")
        BuiltIn().should_be_equal_as_integers(rc, 0)

        if source_number_of_permissions != parent_number_of_permissions:
            file_paths, rc = remote_sudo_rc(f"{semanage_cmd} | grep {source} | awk '{{ print $1 }}'")
            BuiltIn().should_be_equal_as_integers(rc, 0)
            for file_path in file_paths.splitlines():
                file_exists, rc = remote_sudo_rc(f"[ -e {file_path} ] && echo found || echo safe")
                BuiltIn().should_be_equal_as_integers(rc, 0)

                if file_exists == "found":
                    error_list.append(f"source: {source} has differnt permission count than parent: {parent} with file({file_path}), needs investigation.")

    return error_list


# Given a source context and target source target, try and traverse the process transiton path to see
# if it's possible to gain access to the target context.
def run_traversal_access_check(source_target_map: dict[str, List[str]]) -> List[str]:
    allow_cmd = "sesearch -S -A"
    error_list = []

    for parent_source, targets in source_target_map.items():
        # For a parent source, follow to get all possible sources
        # i.e source_a --transition--> source_b --transition--> source_c
        all_sources = get_all_traversal_transition_contexts(parent_source)
        if len(all_sources) == 0:
            return error_list

        errs = sources_have_less_permission_than_parent(parent_source, all_sources)
        if errs and len(errs) > 0:
            error_list.extend(errs)

        # for all desired targets check if the identified sources have allow rules to our target
        # currently, we fail on any allow rules existing
        for target in targets:
            for source in all_sources:
                stdout, rc = remote_sudo_rc(f"{allow_cmd} -s {source} -t {target} 2>&1")
                if rc != 0:
                    error_list.append(f"failed to run command {allow_cmd} on source: {source} and target: {target}")
                if stdout != "":
                    error_list.append(f"found access from source: {source} to target: {target}")

    return error_list


def run_default_traversal_access_check() -> List[str]:
    return run_traversal_access_check(SOURCE_TARGET_TRANSITION)


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
