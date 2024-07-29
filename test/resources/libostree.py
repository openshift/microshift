import os.path
import DataFormats
import SSHLibrary
from robot.libraries.BuiltIn import BuiltIn

_log = BuiltIn().log

DATA_DIR = "/var/lib/microshift"
VERSION_FILE = f"{DATA_DIR}/version"

BACKUP_STORAGE = "/var/lib/microshift-backups"
HEALTH_FILE = f"{BACKUP_STORAGE}/health.json"


def remote_sudo_rc(cmd: str) -> tuple[str, int]:
    ssh = BuiltIn().get_library_instance("SSHLibrary")
    stdout, stderr, rc = SSHLibrary.SSHLibrary.execute_command(
        ssh, command=cmd, sudo=True, return_stderr=True, return_rc=True
    )
    BuiltIn().log(f"stdout:\n{stdout}")
    BuiltIn().log(f"stderr:\n{stderr}")
    BuiltIn().log(f"rc: {rc}")
    return stdout, rc


def remote_sudo(cmd: str) -> str:
    stdout, rc = remote_sudo_rc(cmd)
    BuiltIn().should_be_equal_as_integers(rc, 0)
    return stdout


def get_booted_deployment_id() -> str:
    """
    Get ID of currently booted deployment
    """
    stdout = remote_sudo("rpm-ostree status --booted --json")
    return DataFormats.json_parse(stdout)["deployments"][0]["id"]


def get_staged_deployment_id() -> str:
    """
    Get ID of a staged deployment
    """
    stdout = remote_sudo("rpm-ostree status --json")
    deploy = DataFormats.json_parse(stdout)["deployments"][0]
    BuiltIn().should_be_true(deploy["staged"])
    return deploy["id"]


def get_deployment_backup_prefix_path(deploy_id: str) -> str:
    """
    Get backup path prefix for current deployment

    Prefix path is BACKUP_STORAGE/{id}.
    Globbing directories starting with the prefix will yield
    list of backups for the deployment.
    """
    BuiltIn().should_not_be_empty(BACKUP_STORAGE)
    BuiltIn().should_not_be_empty(deploy_id)
    return os.path.join(BACKUP_STORAGE, deploy_id)


def create_fake_backups(count: int, type_unknown: bool = False) -> None:
    """
    Create number of fake Backup directories.
    Unknown types are directories that do not match automated backups naming
    convention of 'deploymentID_bootID' which can be described more
    specifically by (osname)-(64 chars).(int)_(32 chars).
    Such backups should not be automatically pruned by MicroShift.
    """
    deploy_id = get_booted_deployment_id()
    prefix_path = (
        f"{get_deployment_backup_prefix_path(deploy_id)}_fake000000000000000000000000"
    )

    if type_unknown:
        prefix_path = os.path.join(BACKUP_STORAGE, "unknown_")

    for number in range(0, count):
        remote_sudo(f"mkdir -p {prefix_path}{number:0>4}")


def remove_backups_for_deployment(deploy_id: str) -> None:
    """Remove any existing backup for specified deployment"""
    prefix_path = get_deployment_backup_prefix_path(deploy_id)
    remote_sudo(f"rm -rf {prefix_path}*")


def remove_backup_storage() -> None:
    """
    Removes entire backup storage directory (/var/lib/microshift-backups)
    which contains backups and health data
    """
    remote_sudo(f"rm -rf {BACKUP_STORAGE}")


def get_current_ref() -> str:
    """
    Get reference of current deployment
    """
    ref = remote_sudo("rpm-ostree status --json | jq -r '.deployments[0].origin'")
    _log(f"Current ref: {ref}")
    # remote and ref are separated with colon
    return ref.split(":")[1]


def get_persisted_system_health() -> str:
    """
    Get system health information from health.json file
    """
    return remote_sudo(f"jq -r '.health' {BACKUP_STORAGE}/health.json")


def rpm_ostree_rebase(ref: str) -> None:
    """
    Rebase system to given OSTRee ref
    """
    return remote_sudo(f"rpm-ostree rebase {ref}")


def rpm_ostree_rollback() -> None:
    """
    Rollback system
    """
    return remote_sudo("rpm-ostree rollback")


def rebase_system(ref: str) -> str:
    """
    Rebase system to given OSTRee ref and return its deployment ID
    """
    rpm_ostree_rebase(ref)
    return get_staged_deployment_id()


def get_current_boot_id() -> str:
    boot_id = remote_sudo("cat /proc/sys/kernel/random/boot_id")
    return boot_id.replace("-", "")


def does_backup_exist(deploy_id: str, boot_id: str = "") -> bool:
    prefix = get_deployment_backup_prefix_path(deploy_id)

    if boot_id != "":
        path = f"{prefix}_{boot_id}"
    else:
        path = f"{prefix}"

    return path_exists(path)


def path_exists(path: str) -> bool:
    out, rc = remote_sudo_rc(f"test -e {path}")
    return rc == 0


def path_should_exist(path: str) -> None:
    BuiltIn().should_be_true(path_exists(path))


def path_should_not_exist(path: str) -> None:
    BuiltIn().should_not_be_true(path_exists(path))


def cleanup_rpm_ostree() -> None:
    """Removes any pending or rollback deployments leaving only currently booted"""
    remote_sudo("rpm-ostree cleanup --pending --rollback")


def create_agent_config(cfg: str) -> None:
    remote_sudo(f"echo '{cfg}' | sudo tee /var/lib/microshift-test-agent.json")


def write_greenboot_microshift_wait_timeout(seconds: int) -> None:
    remote_sudo(
        f"echo 'MICROSHIFT_WAIT_TIMEOUT_SEC={seconds}' | sudo tee /etc/greenboot/greenboot.conf"
    )


def remove_greenboot_microshift_wait_timeout() -> None:
    remote_sudo("sed -i '/^MICROSHIFT_WAIT_TIMEOUT_SEC=/d' /etc/greenboot/greenboot.conf")


def no_transaction_in_progress() -> None:
    stdout = remote_sudo("rpm-ostree status --json")
    status = DataFormats.json_parse(stdout)
    key = "transaction"
    transaction_in_progress = key in status and status[key] is not None
    BuiltIn().should_not_be_true(transaction_in_progress)


def write_insecure_registry_url(url: str) -> None:
    remote_sudo(
        f"printf '[[registry]]\nlocation = \"{url}\"\ninsecure = true\n' |"
        "sudo tee /etc/containers/registries.conf.d/999-microshift-insecure-registry.conf"
    )


def remove_insecure_registry_url() -> None:
    remote_sudo('rm -f /etc/containers/registries.conf.d/999-microshift-insecure-registry.conf')


def rebase_bootc_system(ref: str) -> str:
    """
    Rebase system to given bootc image ref and return its deployment ID
    """
    remote_sudo(f"bootc switch --quiet {ref}")
    return get_staged_deployment_id()
