import os
import subprocess
from fastmcp import FastMCP
import yaml
import tempfile

# Initialize FastMCP server
mcp = FastMCP("microshift")

SSH_IP_ADDR = os.getenv('SSH_IP_ADDR')
SSH_USER = os.getenv('SSH_USER')
SSH_CONFIG_FILE = os.getenv('SSH_CONFIG_FILE')
KUBECONFIG_PATH = os.getenv('KUBECONFIG_PATH')


# MCP resources
@mcp.resource(uri='microshift://available_commands')
def get_microshift_available_commands() -> str:
    """Get the available commands for the Microshift cluster"""
    return _run_command_smart(['sudo', 'microshift', 'help'])


# MCP tools
@mcp.tool()
def microshift_systemctl_commands(action: str) -> str:
    """Run a systemctl command on the Microshift cluster to get status, start, stop, restart, etc."""
    return _run_command_smart(['sudo', 'systemctl', action, 'microshift'])


@mcp.tool()
def get_latest_microshift_service_logs(number_of_lines: int = 100) -> str:
    """Get the latest logs from the Microshift cluster"""
    return _run_command_smart(['sudo', 'journalctl', '-u', 'microshift', '-n', f'{number_of_lines}'])


@mcp.tool()
def get_pods(namespace: str = 'all') -> str:
    """Get the pods in the Microshift cluster, by default all namespaces are returned"""
    if namespace == 'all':
        return _run_oc_command(['get', 'pods', '--all-namespaces', '-ojson'])
    else:
        return _run_oc_command(['get', 'pods', '-n', namespace, '-ojson'])


@mcp.tool()
def run_microshift_commands(action: str) -> str:
    """Run a microshift command on the Microshift cluster"""
    return _run_command_smart(['sudo', 'microshift', action])


@mcp.tool()
def get_microshift_current_config() -> str:
    """Get the current microshift config"""
    return _run_command_smart(['sudo', 'microshift', 'show-config'])


@mcp.tool()
def get_default_config_yaml(component: str = 'microshift') -> str:
    """Get the default config.yaml file for microshift, lvmd, lvms, ovn"""
    if component == 'microshift':
        return _run_command_smart(['sudo', 'cat', '/etc/microshift/config.yaml.default'])
    elif component == 'lvmd' or component == 'lvms':
        return _run_command_smart(['sudo', 'cat', '/etc/microshift/lvmd.yaml.default'])
    elif component == 'ovn':
        return _run_command_smart(['sudo', 'cat', '/etc/microshift/ovn.yaml.default'])


@mcp.tool()
def get_microshift_observability_config() -> str:
    """Get the current observability config for the Microshift cluster"""
    return _run_command_smart(['sudo', 'cat', '/etc/microshift/observability/opentelemetry-collector.yaml'])


@mcp.tool()
def get_microshift_custom_config() -> str:
    """Get the current custom microshift from /etc/microshift/config.d/*"""
    return _run_command_smart(['sudo', 'cat', '/etc/microshift/config.d/*'])


@mcp.tool()
def get_microshift_custom_manifests() -> str:
    """Get the current custom manifests from /etc/microshift/manifests.d/* and /etc/microshift/manifests/*"""
    return _run_command_smart(['sudo', 'cat', '/etc/microshift/manifests.d/*', '/etc/microshift/manifests/*'])


@mcp.tool()
def override_microshift_config(component: str, override_config: str) -> str:
    """Override the Microshift cluster configuration for microshift, lvmd, lvms, ovn"""
    if component == 'microshift':
        config_file = '/etc/microshift/config.yaml'
    elif component == 'lvmd' or component == 'lvms':
        config_file = '/etc/microshift/lvmd.yaml'
    elif component == 'ovn':
        config_file = '/etc/microshift/ovn.yaml'
    else:
        return "Invalid component"

    config = yaml.safe_load(override_config)

    # Create a temporary file with the config content
    with tempfile.NamedTemporaryFile(mode='w', suffix='.yaml', delete=False) as temp_file:
        yaml.dump(config, temp_file, default_flow_style=False)
        temp_config_path = temp_file.name

    if _is_local_host():
        # Copy locally
        _run_command_direct(['sudo', 'cp', temp_config_path, config_file])
    else:
        # Copy via SSH
        _run_command_direct(['scp', '-F', SSH_CONFIG_FILE, temp_config_path, f'{SSH_USER}@{SSH_IP_ADDR}:/tmp/config.yaml'])
        _run_ssh_command(['sudo', 'cp', '/tmp/config.yaml', config_file])

    # Clean up the temporary file
    os.unlink(temp_config_path)

    return _run_command_smart(['sudo', 'cat', config_file])


# Private functions
def _run_oc_command(command: list[str]) -> str:
    """Run an oc command on the Microshift cluster"""
    return _run_command_direct(['oc', f'--kubeconfig={KUBECONFIG_PATH}'] + command)


def _run_command_direct(command: list[str]) -> str:
    """Run a command directly on the local system"""
    try:
        result = subprocess.run(
            command,
            capture_output=True,
            text=True,
            timeout=30
        )

        if result.returncode == 0:
            return f"command output:\n{result.stdout}"
        else:
            return f"Failed to execute command:\n{result.stderr}"

    except subprocess.TimeoutExpired:
        return "command timed out"
    except Exception as e:
        return f"Failed to execute command: {str(e)}"


def _run_ssh_command(command: list[str]) -> str:
    """Run a command on the remote server via SSH"""
    try:
        # Convert command list to a single string for SSH execution
        command_str = ' '.join(command)
        ssh_command = [
            'ssh',
            '-F', SSH_CONFIG_FILE,
            f'{SSH_USER}@{SSH_IP_ADDR}',
            command_str
        ]

        result = subprocess.run(ssh_command, capture_output=True, text=True, timeout=30)

        if result.returncode == 0:
            return f"command output:\n{result.stdout}"
        else:
            return f"Failed to execute command:\n{result.stderr}"

    except subprocess.TimeoutExpired:
        return "command timed out"
    except Exception as e:
        return f"Failed to execute command: {str(e)}"


def _is_local_host() -> bool:
    """Check if we're running on the same host as MicroShift"""
    try:
        result = subprocess.run(
            ['microshift', 'help'],
            capture_output=True,
            text=True,
            timeout=10
        )
        return result.returncode == 0
    except (subprocess.TimeoutExpired, FileNotFoundError, Exception):
        return False


def _run_local_command(command: list[str]) -> str:
    """Run a command locally (with sudo if needed)"""
    return _run_command_direct(command)


def _run_command_smart(command: list[str]) -> str:
    """Run a command on the MicroShift host, using local execution if possible"""
    if _is_local_host():
        return _run_local_command(command)
    else:
        return _run_ssh_command(command)


if __name__ == "__main__":
    mcp.run(transport='stdio')
