import os
from typing import Any
import httpx
import subprocess
from fastmcp import FastMCP
import json,yaml

# Initialize FastMCP server
mcp = FastMCP("microshift")

SSH_IP_ADDR = os.getenv('SSH_IP_ADDR')
SSH_USER = os.getenv('SSH_USER')
SSH_CONFIG_FILE = os.getenv('SSH_CONFIG_FILE')
KUBECONFIG_PATH = os.getenv('KUBECONFIG_PATH')

# MCP resources
@mcp.resource(uri='microshift:available_commands')
def get_microshift_available_commands() -> str:
    """Get the available commands for the Microshift cluster"""
    return _run_ssh_command(['sudo', 'microshift', 'help'])

# MCP tools
@mcp.tool()
def run_systemctl_microshift_commands(action: str) -> str:
    """Run a systemctl command on the Microshift cluster to start, stop, restart, etc."""
    return _run_ssh_command(['sudo', 'systemctl', action, 'microshift'])

@mcp.tool()
def get_latest_microshift_service_logs(number_of_lines: int = 100) -> str:
    """Get the latest logs from the Microshift cluster"""
    return _run_ssh_command(['sudo', 'journalctl', '-u', 'microshift', '-n', f'{number_of_lines}'])

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
    return _run_ssh_command(['sudo', 'microshift', action])

@mcp.tool()
def get_config_yaml_default() -> str:
    """Get the default config.yaml file for the Microshift cluster"""
    return _run_ssh_command(['sudo', 'cat', '/etc/microshift/config.yaml.default'])

@mcp.tool()
def override_microshift_config(override_config: str) -> str:
    """Override the Microshift cluster configuration on config.yaml"""
    config = yaml.safe_load(override_config)
    # Create a temporary file with the config content
    import tempfile
    with tempfile.NamedTemporaryFile(mode='w', suffix='.yaml', delete=False) as temp_file:
        yaml.dump(config, temp_file, default_flow_style=False)
        temp_config_path = temp_file.name
    
    # Copy the temp file to the remote server and then to the config location
    _run_command('scp', ['-F', SSH_CONFIG_FILE, temp_config_path, f'{SSH_USER}@{SSH_IP_ADDR}:/tmp/config.yaml'])
    result = _run_ssh_command(['sudo', 'cp', '/tmp/config.yaml', '/etc/microshift/config.yaml'])
    
    # Clean up the temporary file
    os.unlink(temp_config_path)

    final_config = _run_ssh_command(['sudo', 'cat', '/etc/microshift/config.yaml'])
    return f"{final_config}"

# Private functions
def _run_oc_command(command: list[str]) -> str:
    """Run an oc command on the Microshift cluster"""
    return _run_command('oc', [f'--kubeconfig={KUBECONFIG_PATH}', *command])

def _run_ssh_command(command: list[str]) -> str:
    """Run a command on the remote server"""
    return _run_command('ssh', ['-F', f'{SSH_CONFIG_FILE}', f'{SSH_USER}@{SSH_IP_ADDR}', *command])

def _run_command(command: str, args: list[str]) -> str:
    """Run a command"""
    try:
        result = subprocess.run([
            f'{command}', 
            *args
        ], capture_output=True, text=True, timeout=30)
        
        if result.returncode == 0:
            return f"command output:\n{result.stdout}"
        else:
            return f"Failed to execute command: :\n{result.stderr}"

    except subprocess.TimeoutExpired:
        return "command timed out"
    except Exception as e:
        return f"Failed to execute command: {str(e)}"

if __name__ == "__main__":
    # Initialize and run the server
    mcp.run(transport='stdio')
