from agno.agent import Agent
from agno.models.google import Gemini
from agno.tools.mcp import MCPTools

import asyncio
import sys

async def run_agent(message: str) -> None:
    async with (
        MCPTools(
            f"bash /Users/agullon/workspace/microshift/scripts/mcp-server/start-mcp-server.sh",
        ) as mcp_tools,
    ):
        agent = Agent(
            model=Gemini(id="gemini-2.0-flash-001"),
            tools=[mcp_tools],
            show_tool_calls=True,
            debug_mode=False,
        )
        await agent.aprint_response(message)

# Example usage
if __name__ == "__main__":    
    if len(sys.argv) < 2:
        print("Usage: python agno_microshift.py '<your message>'")
        sys.exit(1)
    
    user_input = sys.argv[1]
    asyncio.run(run_agent(user_input))
