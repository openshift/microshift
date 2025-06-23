import os
from typing import Any
import httpx
import subprocess
from mcp.server.fastmcp import FastMCP

# Initialize FastMCP server
mcp = FastMCP("microshift")

SSH_IP_ADDR = os.getenv('SSH_IP_ADDR')
SSH_USER = os.getenv('SSH_USER')
SSH_CONFIG_FILE = os.getenv('SSH_CONFIG_FILE')
KUBECONFIG_PATH = os.getenv('KUBECONFIG_PATH')

# MCP tools
@mcp.tool()
async def get_microshift_service_status() -> str:
    """Get the status of the Microshift systemd service running on the remote server"""
    return await _run_ssh_command('sudo systemctl status microshift')

@mcp.tool()
async def get_running_pods() -> str:
    """Get the running pods in the Microshift cluster"""
    return await _run_oc_command('get pods --all-namespaces')

@mcp.tool()
async def summarize_microshift_cluster_config() -> str:
    """Summarize the Microshift cluster configuration"""
    return await _run_ssh_command('sudo microshift show-config')

# Private functions
async def _run_oc_command(command: str) -> str:
    """Run an oc command on the Microshift cluster"""
    return await _run_command('oc', [f'--kubeconfig={KUBECONFIG_PATH}', command])

async def _run_ssh_command(command: str) -> str:
    """Run a command on the remote server"""
    return await _run_command('ssh', ['-F', f'{SSH_CONFIG_FILE}', f'{SSH_USER}@{SSH_IP_ADDR}', command])

async def _run_command(command: str, args: list[str]) -> str:
    """Run a command"""
    try:
        result = subprocess.run([
            f'{command}', 
            *args
        ], capture_output=True, text=True, timeout=30)
        
        if result.returncode == 0:
            return f"command output:\n{result.stdout}"
        else:
            return f"command output:\n{result.stderr}"

    except subprocess.TimeoutExpired:
        return "command timed out"
    except Exception as e:
        return f"Failed to execute command: {str(e)}"

if __name__ == "__main__":
    # Initialize and run the server
    mcp.run(transport='stdio')
