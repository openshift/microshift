# Cursor configuration

````
❯ cat .cursor/mcp.json
{
    "mcpServers": {
        "microshift": {
            "command": "<full path to start-mcp-server.sh>",
            "env": {
                "KUBECONFIG_PATH": "${KUBECONFIG_PATH}",
                "SSH_IP_ADDR": "${SSH_IP_ADDR}",
                "SSH_USER": "${SSH_USER}",
                "SSH_CONFIG_FILE": "${SSH_CONFIG_FILE}"
            }
        },
        "kubernetes": {
            "command": "npx",
            "args": ["mcp-server-kubernetes"],
            "env": {
                "KUBECONFIG_PATH": "${KUBECONFIG_PATH}"
            }
        }
    }
}
```
