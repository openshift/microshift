# LLM CLI - MicroShift AI Assistant

A command-line interface for interacting with Large Language Models (LLMs) with optional Model Context Protocol (MCP) tools integration for MicroShift management.

## 🚀 Overview

The LLM CLI provides a unified interface to interact with different AI models:
- **IBM Granite** (default) - Enhanced custom implementation with MCP support for internal Red Hat LLMs from Model.corp
- **Google Gemini** - Google's Gemini 2.0 Flash model

With optional MCP tools integration, you can perform actual MicroShift operations or ask general AI questions.

## 📁 Files

```
llm-cli/
├── llm-cli.py          # Main CLI application
├── llm-cli.sh          # Shell wrapper with auto-setup
├── ibm_model.py        # Enhanced IBM Granite model implementation
├── requirements.txt    # Python dependencies
└── README.md          # This file
```

## 🎯 Features

- **Two AI Models**: IBM Granite (default) and Google Gemini
- **MCP Tools Integration**: Optional tools for MicroShift management
- **Debug Mode**: Detailed logging for troubleshooting
- **Auto-Setup**: Shell wrapper automatically creates virtual environment
- **Flexible Usage**: With or without MCP tools

## 🚀 Quick Start

### Option 1: Using Shell Wrapper (Recommended)

The shell wrapper (`llm-cli.sh`) automatically sets up the environment:

```bash
# From the microshift workspace root
./scripts/ai-tools/llm-cli/llm-cli.sh "List all pods from all namespaces"
```

### Option 2: Manual Setup

```bash
cd scripts/ai-tools/llm-cli

# Create virtual environment
python3 -m venv .venv
source .venv/bin/activate

# Install dependencies
pip install -r requirements.txt

# Run the CLI
python llm-cli.py "List all pods from all namespaces"
```

## 🔧 Environment Variables

Set these environment variables before using the CLI:

### Required for Models

```bash
# For IBM Granite model (default)
export GRANITE_API_KEY="your-granite-api-key"

# For Google Gemini model
export GOOGLE_API_KEY="your-google-api-key"
```

### Required for MCP Tools

```bash
# MicroShift connection details
export SSH_IP_ADDR="your-microshift-host"
export SSH_USER="your-ssh-user"
export SSH_CONFIG_FILE="path/to/ssh/config"
export KUBECONFIG_PATH="path/to/kubeconfig"

# MCP server startup command
export START_MCP_SERVER_COMMAND="bash /full/path/to/scripts/ai-tools/mcp-server/start-mcp-server.sh"
```

## 📋 Usage Examples

### Basic Usage (Default Granite with MCP tools)

```bash
# List all pods
./scripts/ai-tools/llm-cli/llm-cli.sh "List all pods from all namespaces"

# Check service status
./scripts/ai-tools/llm-cli/llm-cli.sh "What is the status of the MicroShift service?"

# Get logs
./scripts/ai-tools/llm-cli/llm-cli.sh "Show me the latest 20 lines from MicroShift logs"
```

### Using Different Models

```bash
# Use Gemini model
./scripts/ai-tools/llm-cli/llm-cli.sh --llm-model gemini "Show me the MicroShift service status"

# Use Granite model (explicit)
./scripts/ai-tools/llm-cli/llm-cli.sh --llm-model granite "List all pods"
```

### Without MCP Tools (General AI Questions)

```bash
# Ask general questions without MCP tools
./scripts/ai-tools/llm-cli/llm-cli.sh --no-mcp "What is MicroShift?"

# Explain concepts
./scripts/ai-tools/llm-cli/llm-cli.sh --no-mcp "Explain Kubernetes pods and services"

# Get help with troubleshooting
./scripts/ai-tools/llm-cli/llm-cli.sh --no-mcp "How do I troubleshoot MicroShift networking issues?"
```

### Debug Mode

```bash
# Enable debug mode for detailed logging
./scripts/ai-tools/llm-cli/llm-cli.sh --debug "List all pods"

# Debug with specific model
./scripts/ai-tools/llm-cli/llm-cli.sh --llm-model granite --debug "Show me the microshift service status"

# Debug without MCP tools
./scripts/ai-tools/llm-cli/llm-cli.sh --no-mcp --debug "What is container orchestration?"
```

## 🛠️ Command Line Options

```bash
./scripts/ai-tools/llm-cli/llm-cli.sh [OPTIONS] "MESSAGE"
```

### Options

| Option | Description | Default |
|--------|-------------|---------|
| `--llm-model {gemini,granite}` | Choose AI model | `granite` |
| `--no-mcp` | Disable MCP tools | MCP enabled |
| `--debug` | Enable debug mode | Debug disabled |
| `--help` | Show help message | - |

### Examples

```bash
# Show help
./scripts/ai-tools/llm-cli/llm-cli.sh --help

# Use all options together
./scripts/ai-tools/llm-cli/llm-cli.sh --llm-model gemini --no-mcp --debug "Explain Kubernetes concepts"
```

## 🔍 Available MCP Tools

When MCP tools are enabled, the following MicroShift operations are available:

| Tool | Description | Example Usage |
|------|-------------|---------------|
| `run_systemctl_microshift_commands` | Control MicroShift service | "Start the MicroShift service" |
| `get_latest_microshift_service_logs` | Get service logs | "Show me the latest logs" |
| `get_pods` | List cluster pods | "List all pods" |
| `run_microshift_commands` | Run MicroShift commands | "Run microshift version" |
| `get_config_yaml_default` | Get default configuration | "Show me the default config" |
| `override_microshift_config` | Update configuration | "Update the config with..." |

## 🏗️ Models

### IBM Granite (Default)

- **Enhanced custom implementation** with MCP support
- **Designed for internal Red Hat LLMs** from Model.corp
- **Features**: Tool calling, streaming, debug mode, structured outputs
- **API**: Requires `GRANITE_API_KEY`

### Google Gemini

- **Google's Gemini 2.0 Flash** model
- **General-purpose AI model** with excellent capabilities
- **Features**: Tool calling, multimodal support
- **API**: Requires `GOOGLE_API_KEY`

## 🧪 Testing

### Test Basic Functionality

```bash
# Test without MCP tools (should work with just API keys)
./scripts/ai-tools/llm-cli/llm-cli.sh --no-mcp "Hello, how are you?"

# Test with Gemini
./scripts/ai-tools/llm-cli/llm-cli.sh --llm-model gemini --no-mcp "What is MicroShift?"
```

### Test MCP Integration

```bash
# Test MCP tools (requires full environment setup)
./scripts/ai-tools/llm-cli/llm-cli.sh --debug "List all pods"

# Test specific operations
./scripts/ai-tools/llm-cli/llm-cli.sh "What is the status of MicroShift?"
```

## 🐛 Troubleshooting

### Common Issues

#### 1. Environment Variables Not Set

```bash
ERROR: GRANITE_API_KEY environment variable is not set
ERROR: START_MCP_SERVER_COMMAND environment variable is not set
```

**Solution**: Set the required environment variables as shown in the configuration section.

#### 2. MCP Server Connection Failed

```bash
[ERROR] Failed to connect to MCP server
```

**Solutions**:
- Check SSH connectivity to MicroShift host
- Verify `START_MCP_SERVER_COMMAND` points to correct script
- Ensure MCP server environment variables are set
- Test SSH access manually

#### 3. Model Import Errors

```bash
ERROR: Granite model (ibm_model.py) not found
```

**Solution**: Ensure `ibm_model.py` is in the same directory as `llm-cli.py`.

#### 4. Virtual Environment Issues

```bash
Permission denied or module not found errors
```

**Solution**: 
- Use the shell wrapper (`llm-cli.sh`) for automatic setup
- Or manually recreate the virtual environment:
  ```bash
  rm -rf .venv
  python3 -m venv .venv
  source .venv/bin/activate
  pip install -r requirements.txt
  ```

### Debug Mode

Enable debug mode for detailed logging:

```bash
./scripts/ai-tools/llm-cli/llm-cli.sh --debug "Your query here"
```

Debug output includes:
- Model initialization details
- MCP tool interactions
- API request/response information
- Tool call processing

## 📚 Advanced Usage

### Custom Model Configuration

The IBM Granite model supports extensive configuration. You can modify `llm-cli.py` to customize:

```python
llm_model = Granite(
    id="/data/granite-3.2-8b-instruct",
    name="Custom Granite",
    api_key=os.getenv("GRANITE_API_KEY"),
    temperature=0.3,        # Lower for more consistent results
    max_tokens=2000,        # Increase for longer responses
    debug_mode=True,        # Enable detailed logging
    mcp_tools_enabled=True  # Enable MCP tool support
)
```

### Adding New Models

To add support for new models:

1. Import the model in `llm-cli.py`
2. Add it to the `choices` in argument parser
3. Add a new case in the model selection logic
4. Update this README

### Batch Operations

You can create scripts for common operations:

```bash
#!/bin/bash
# check-microshift.sh

echo "=== MicroShift Status ==="
./scripts/ai-tools/llm-cli/llm-cli.sh "What is the status of MicroShift?"

echo "=== Pod List ==="
./scripts/ai-tools/llm-cli/llm-cli.sh "List all pods"

echo "=== Recent Logs ==="
./scripts/ai-tools/llm-cli/llm-cli.sh "Show me the latest 10 lines from MicroShift logs"
```

## 🔗 Related Files

- `../mcp-server/` - MCP server implementation
- `../README.md` - Main AI tools documentation
- `ibm_model.py` - Enhanced IBM Granite model implementation

## 🎯 Best Practices

1. **Start with `--no-mcp`** to test basic model functionality
2. **Use debug mode** (`--debug`) when troubleshooting
3. **Set appropriate timeouts** for long-running MCP operations
4. **Use the shell wrapper** (`llm-cli.sh`) for automatic environment setup
5. **Keep API keys secure** in environment variables, not in code
6. **Test connectivity** to MicroShift host before using MCP tools

## 🤝 Contributing

When modifying the CLI:

1. Test with both models (Granite and Gemini)
2. Test with and without MCP tools
3. Update help text and examples
4. Add appropriate error handling
5. Update this README

## 📞 Support

For issues:
- **Model API errors**: Check API keys and network connectivity
- **MCP tools errors**: Verify MicroShift host connectivity and permissions
- **General errors**: Use `--debug` flag for detailed logging

## 📝 Examples Summary

```bash
# Quick start - list pods with default Granite model
./scripts/ai-tools/llm-cli/llm-cli.sh "List all pods from all namespaces"

# Use different model
./scripts/ai-tools/llm-cli/llm-cli.sh --llm-model gemini "Show MicroShift status"

# General AI questions (no MCP tools needed)
./scripts/ai-tools/llm-cli/llm-cli.sh --no-mcp "What is Kubernetes?"

# Debug mode for troubleshooting
./scripts/ai-tools/llm-cli/llm-cli.sh --debug "Get MicroShift logs"

# All options combined
./scripts/ai-tools/llm-cli/llm-cli.sh --llm-model gemini --no-mcp --debug "Explain containers"
```
