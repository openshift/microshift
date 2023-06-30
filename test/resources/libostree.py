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
