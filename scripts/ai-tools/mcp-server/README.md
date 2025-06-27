# MCP Server for MicroShift

A Model Context Protocol (MCP) server that exposes MicroShift management operations as standardized tools for AI models. This server enables AI assistants to perform actual MicroShift operations through a secure, standardized interface.

## 🚀 Overview

The MCP server bridges the gap between AI models and MicroShift cluster management by exposing common operations as MCP tools. It provides secure, authenticated access to MicroShift clusters through SSH and kubectl/oc commands.

## 📁 Files

```
mcp-server/
├── mcp-server.py          # Main MCP server implementation
├── start-mcp-server.sh    # Server startup script with auto-setup
├── requirements.txt       # Python dependencies
└── README.md             # This file
```

## 🎯 Features

- **FastMCP Implementation**: Built using the FastMCP framework for efficient MCP protocol handling
- **Secure Access**: Uses SSH key-based authentication and kubeconfig for secure cluster access
- **Auto-Setup**: Startup script automatically creates virtual environment and installs dependencies
- **Comprehensive Tools**: Six core tools covering most common MicroShift operations
- **Error Handling**: Robust error handling with timeouts and detailed error messages

## 🛠️ Available MCP Tools

The server exposes the following tools that AI models can use:

### 1. Service Management
```python
@mcp.tool()
def run_systemctl_microshift_commands(action: str) -> str:
    """Run a systemctl command on the Microshift cluster to start, stop, restart, etc."""
```
**Usage**: Control the MicroShift service
**Parameters**: `action` - start, stop, restart, status, enable, disable
**Example**: "Start the MicroShift service"

### 2. Log Retrieval
```python
@mcp.tool()
def get_latest_microshift_service_logs(number_of_lines: int = 100) -> str:
    """Get the latest logs from the Microshift cluster"""
```
**Usage**: Retrieve service logs for troubleshooting
**Parameters**: `number_of_lines` - Number of log lines to retrieve (default: 100)
**Example**: "Show me the latest 50 lines from MicroShift logs"

### 3. Pod Management
```python
@mcp.tool()
def get_pods(namespace: str = 'all') -> str:
    """Get the pods in the Microshift cluster, by default all namespaces are returned"""
```
**Usage**: List pods in the cluster
**Parameters**: `namespace` - Specific namespace or 'all' for all namespaces (default: 'all')
**Example**: "List all pods in the kube-system namespace"

### 4. MicroShift Commands
```python
@mcp.tool()
def run_microshift_commands(action: str) -> str:
    """Run a microshift command on the Microshift cluster"""
```
**Usage**: Execute MicroShift-specific commands
**Parameters**: `action` - MicroShift command to run
**Example**: "Run microshift version"

### 5. Configuration Retrieval
```python
@mcp.tool()
def get_config_yaml_default() -> str:
    """Get the default config.yaml file for the Microshift cluster"""
```
**Usage**: Retrieve the default MicroShift configuration
**Parameters**: None
**Example**: "Show me the default MicroShift configuration"

### 6. Configuration Management
```python
@mcp.tool()
def override_microshift_config(override_config: str) -> str:
    """Override the Microshift cluster configuration on config.yaml"""
```
**Usage**: Update MicroShift configuration
**Parameters**: `override_config` - YAML configuration string
**Example**: "Update the configuration with custom settings"

## 🔧 Environment Variables

The server requires these environment variables for operation:

### Required Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `SSH_IP_ADDR` | MicroShift host IP/hostname | `10.1.235.14` |
| `SSH_USER` | SSH username for MicroShift host | `microshift` |
| `SSH_CONFIG_FILE` | Path to SSH config file | `/Users/username/.ssh/config` |
| `KUBECONFIG_PATH` | Path to kubeconfig file | `/path/to/kubeconfig` |

### Setting Environment Variables

```bash
# MicroShift connection details
export SSH_IP_ADDR="your-microshift-host-ip"
export SSH_USER="your-ssh-username"
export SSH_CONFIG_FILE="/path/to/your/ssh/config"
export KUBECONFIG_PATH="/path/to/your/kubeconfig"
```

## 🚀 Quick Start

### Option 1: Using Startup Script (Recommended)

The startup script automatically sets up the environment:

```bash
# From the microshift workspace root
./scripts/ai-tools/mcp-server/start-mcp-server.sh
```

### Option 2: Manual Setup

```bash
cd scripts/ai-tools/mcp-server

# Create virtual environment
python3 -m venv .venv
source .venv/bin/activate

# Install dependencies
pip install -r requirements.txt

# Set environment variables (see above)

# Run the server
python mcp-server.py
```

## 🧪 Testing

### Test Server Startup

```bash
# Test basic server functionality
./scripts/ai-tools/mcp-server/start-mcp-server.sh
```

The server should start and display MCP protocol messages if working correctly.

### Test with LLM CLI

```bash
# Test the server with the LLM CLI
./scripts/ai-tools/llm-cli/llm-cli.sh "List all pods from all namespaces"
```

### Manual Testing

You can test individual components:

```bash
# Test SSH connectivity
ssh -F $SSH_CONFIG_FILE $SSH_USER@$SSH_IP_ADDR "sudo systemctl status microshift"

# Test kubectl/oc access
oc --kubeconfig=$KUBECONFIG_PATH get pods --all-namespaces
```

## 🏗️ Architecture

### MCP Protocol Flow

1. **Client Connection**: AI model connects to MCP server via stdio transport
2. **Tool Discovery**: Server advertises available tools to the client
3. **Tool Execution**: Client requests tool execution with parameters
4. **Command Execution**: Server executes commands on MicroShift host
5. **Response**: Server returns formatted results to client

### Security Model

- **SSH Key Authentication**: Uses SSH keys for secure host access
- **Kubeconfig Authentication**: Uses kubeconfig for Kubernetes API access
- **Sudo Privileges**: Requires sudo access on MicroShift host for service management
- **Command Validation**: All commands are predefined and validated
- **Timeout Protection**: 30-second timeout prevents hanging operations

### Command Execution Flow

```
AI Model → MCP Server → SSH/OC Command → MicroShift Host → Response
```

## 🔍 Available Resources

The server also provides MCP resources for discovery:

### microshift:available_commands
```python
@mcp.resource(uri='microshift:available_commands')
def get_microshift_available_commands() -> str:
    """Get the available commands for the Microshift cluster"""
```

This resource provides help information about available MicroShift commands.

## 🐛 Troubleshooting

### Common Issues

#### 1. Server Won't Start

```bash
Error: Environment variables not set
```

**Solution**: Ensure all required environment variables are set:
```bash
echo $SSH_IP_ADDR
echo $SSH_USER
echo $SSH_CONFIG_FILE
echo $KUBECONFIG_PATH
```

#### 2. SSH Connection Failed

```bash
Failed to execute command: Permission denied
```

**Solutions**:
- Verify SSH key is added to the MicroShift host
- Test SSH connectivity manually
- Check SSH config file permissions
- Ensure SSH user has sudo privileges

#### 3. Kubeconfig Issues

```bash
Failed to execute command: Unable to connect to the server
```

**Solutions**:
- Verify kubeconfig file exists and is readable
- Test oc/kubectl connectivity manually
- Check kubeconfig file permissions
- Ensure cluster is accessible

#### 4. Command Timeouts

```bash
command timed out
```

**Solutions**:
- Check network connectivity to MicroShift host
- Increase timeout in `_run_command` function if needed
- Verify MicroShift service is responsive

### Debug Mode

Enable verbose logging by modifying the server:

```python
import logging
logging.basicConfig(level=logging.DEBUG)
```

### Manual Command Testing

Test individual commands manually:

```bash
# Test SSH command
ssh -F $SSH_CONFIG_FILE $SSH_USER@$SSH_IP_ADDR "sudo systemctl status microshift"

# Test oc command
oc --kubeconfig=$KUBECONFIG_PATH get pods --all-namespaces

# Test MicroShift command
ssh -F $SSH_CONFIG_FILE $SSH_USER@$SSH_IP_ADDR "sudo microshift version"
```

## 📚 Advanced Configuration

### Custom SSH Configuration

Create a custom SSH config file for MicroShift access:

```bash
# ~/.ssh/config or custom config file
Host microshift-host
    HostName your-microshift-ip
    User your-username
    IdentityFile ~/.ssh/your-private-key
    StrictHostKeyChecking no
    UserKnownHostsFile /dev/null
```

### Extending the Server

To add new MCP tools, modify `mcp-server.py`:

```python
@mcp.tool()
def your_custom_tool(parameter: str) -> str:
    """Description of your custom tool"""
    # Implementation here
    return _run_ssh_command(['your', 'command', parameter])
```

### Custom Startup Configuration

Modify `start-mcp-server.sh` to set custom defaults:

```bash
# Custom environment variable defaults
export KUBECONFIG_PATH="${KUBECONFIG_PATH:-/your/custom/kubeconfig}"
export SSH_IP_ADDR="${SSH_IP_ADDR:-your-default-host}"
export SSH_USER="${SSH_USER:-your-default-user}"
export SSH_CONFIG_FILE="${SSH_CONFIG_FILE:-/your/custom/ssh/config}"
```

## 🔗 Integration

### Using with AI Models

The server is designed to work with any MCP-compatible AI model:

```python
# Example with agno framework
from agno.tools.mcp import MCPTools

async with MCPTools("bash /path/to/start-mcp-server.sh") as mcp_tools:
    # AI model can now use MicroShift tools
    pass
```

### Protocol Compliance

The server implements the MCP 1.0 specification:
- **Transport**: stdio (standard input/output)
- **Tools**: Synchronous tool execution
- **Resources**: Static resource discovery
- **Error Handling**: Standardized error responses

## 🎯 Best Practices

1. **Security**: Always use SSH keys, never passwords
2. **Permissions**: Use dedicated service accounts with minimal required privileges
3. **Monitoring**: Monitor server logs for security and performance
4. **Updates**: Keep dependencies updated for security patches
5. **Testing**: Test all tools individually before deployment
6. **Backup**: Backup configurations before making changes

## 🤝 Contributing

When adding new tools:

1. Follow the existing naming convention
2. Add comprehensive docstrings
3. Include error handling
4. Test with various parameters
5. Update this README

## 📞 Support

For issues:
- **SSH connectivity**: Check SSH keys and network connectivity
- **Kubeconfig issues**: Verify cluster access and permissions
- **Tool failures**: Check MicroShift host status and logs
- **Server errors**: Enable debug logging and check server logs

## 📝 Example Usage

```bash
# Start the server
./scripts/ai-tools/mcp-server/start-mcp-server.sh

# In another terminal, use with LLM CLI
./scripts/ai-tools/llm-cli/llm-cli.sh "What is the status of MicroShift?"
./scripts/ai-tools/llm-cli/llm-cli.sh "List all pods in the cluster"
./scripts/ai-tools/llm-cli/llm-cli.sh "Show me the latest 20 lines from MicroShift logs"
```

The MCP server provides a robust, secure interface for AI-powered MicroShift management!
