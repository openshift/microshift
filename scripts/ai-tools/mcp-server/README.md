# MCP Server for MicroShift

A Model Context Protocol (MCP) server that exposes MicroShift management operations as standardized tools for AI models.

## Features

- **Smart Execution**: Automatically detects local vs remote execution to optimize performance
- **Service Management**: Control MicroShift service (start, stop, restart, status)
- **Log Retrieval**: Get service logs for troubleshooting
- **Pod Management**: List and inspect cluster pods
- **Configuration Management**: View and update MicroShift configuration files
- **Multi-Component Support**: Handle microshift, lvmd, and ovn configurations

## Quick Start

### 1. Set Environment Variables
```bash
export SSH_IP_ADDR="your-microshift-host"
export SSH_USER="microshift"
export SSH_CONFIG_FILE="~/.ssh/config"
export KUBECONFIG_PATH="/path/to/kubeconfig"
```

### 2. Start the Server
```bash
./start-mcp-server.sh
```

## Available MCP Tools

- `microshift_systemctl_commands(action)` - Control MicroShift service
- `get_latest_microshift_service_logs(number_of_lines)` - Get service logs
- `get_pods(namespace)` - List cluster pods
- `run_microshift_commands(action)` - Run MicroShift commands
- `get_microshift_current_config()` - Get current MicroShift configuration
- `get_default_config_yaml(component)` - Get default configuration for microshift, lvmd, lvms, ovn
- `get_microshift_observability_config()` - Get current observability configuration
- `get_microshift_custom_config()` - Get custom configuration from config.d directory
- `get_microshift_custom_manifests()` - Get custom manifests from manifests.d and manifests directories
- `override_microshift_config(component, override_config)` - Update configuration for microshift, lvmd, lvms, ovn

## Files

- `mcp-server.py` - Main server implementation using FastMCP
- `start-mcp-server.sh` - Startup script with auto-setup
- `requirements.txt` - Python dependencies
- `mcp-server-vars.env` - Environment configuration

## Smart Execution

The server automatically detects if it's running on the same host as MicroShift:
- **Local execution**: Commands run directly for better performance
- **Remote execution**: Uses SSH when running on a different machine

## Requirements

- Python 3.11+
- SSH access to MicroShift host (for remote operations)
- Valid kubeconfig file
- Sudo privileges on MicroShift host
