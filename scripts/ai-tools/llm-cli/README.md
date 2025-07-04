# LLM CLI for MicroShift

Command-line interface for interacting with Large Language Models using optional MCP tools for MicroShift management.

## Features

- **Multiple Models**: Support for IBM Granite and Google Gemini
- **MCP Integration**: Optional MCP tools for cluster operations
- **Debug Mode**: Detailed logging for troubleshooting
- **Async Execution**: Efficient async/await implementation
- **Flexible Usage**: Can run with or without MCP tools

## Quick Start

### 1. Set Environment Variables
```bash
# Required for LLM models
export GRANITE_API_KEY="your-granite-api-key"
export GOOGLE_API_KEY="your-google-api-key"  # Optional, for Gemini

# Required for MCP tools
export START_MCP_SERVER_COMMAND="bash /path/to/mcp-server/start-mcp-server.sh"
```

### 2. Usage Examples
```bash
# Use Granite with MCP tools (default)
./llm-cli.sh "List all pods from all namespaces"

# Use Gemini model
./llm-cli.sh --llm-model gemini "Show me the MicroShift service status"

# Ask questions without MCP tools
./llm-cli.sh --no-mcp "What is MicroShift?"

# Enable debug mode
./llm-cli.sh --debug "Get the latest MicroShift logs"
```

## Command Options

- `--llm-model {gemini,granite}` - Choose LLM model (default: granite)
- `--no-mcp` - Disable MCP tools (run model without cluster access)
- `--debug` - Enable debug mode for detailed logging

## Files

- `llm-cli.py` - Main CLI application
- `llm-cli.sh` - Shell wrapper script with auto-setup
- `ibm_model.py` - IBM Granite model implementation
- `requirements.txt` - Python dependencies

## Supported Models

### IBM Granite
- Custom implementation with MCP support
- Requires `GRANITE_API_KEY` environment variable
- Optimized for technical tasks and MicroShift operations

### Google Gemini
- Uses Gemini 2.0 Flash model
- Requires `GOOGLE_API_KEY` environment variable
- General-purpose conversational AI

## Requirements

- Python 3.11+
- Valid API keys for chosen LLM models
- MCP server setup (for cluster operations)
