import asyncio
import sys
import os
import argparse

from agno.agent import Agent
from agno.models.google import Gemini
from agno.tools.mcp import MCPTools
from agno.models.base import Model
from ibm_model import FullyCustomModel as Granite


async def run_agent(llm_model: Model, prompt_message: str, use_mcp: bool = True, debug: bool = False) -> None:
    if use_mcp:
        if debug:
            print("[DEBUG] Using MCP tools")
        if not os.getenv("START_MCP_SERVER_COMMAND"):
            print("ERROR: START_MCP_SERVER_COMMAND environment variable is not set")
            sys.exit(1)
        async with MCPTools(
            f"{os.getenv('START_MCP_SERVER_COMMAND')}",
            timeout_seconds=60,
        ) as mcp_tools:
            agent = Agent(
                model=llm_model,
                tools=[mcp_tools],
                show_tool_calls=True,
                debug_mode=debug,
                markdown=True,
            )
            await agent.aprint_response(prompt_message, stream=True)
    else:
        if debug:
            print("[DEBUG] Running without MCP tools")
        agent = Agent(
            model=llm_model,
            debug_mode=debug,
            markdown=True,
        )
        await agent.aprint_response(prompt_message, stream=True)


def parse_arguments():
    """Parse command line arguments"""
    parser = argparse.ArgumentParser(
        description="LLM CLI with optional MCP tools support",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  # Use Granite with MCP tools (default)
  python llm-cli.py "List all pods from all namespaces"

  # Use Gemini with MCP tools
  python llm-cli.py --llm-model gemini "List all pods from all namespaces"

  # Use Granite without MCP tools
  python llm-cli.py --no-mcp "What is MicroShift?"

  # Use Gemini without MCP tools
  python llm-cli.py --llm-model gemini --no-mcp "What is MicroShift?"

  # Use with debug mode
  python llm-cli.py --llm-model granite --debug "Show me the microshift service status"
        """
    )

    parser.add_argument(
        "prompt_message",
        nargs="+",
        help="Message/prompt to send to the model"
    )

    parser.add_argument(
        "--llm-model",
        choices=["gemini", "granite"],
        default="granite",
        help="LLM model to use (default: granite)"
    )

    parser.add_argument(
        "--no-mcp",
        action="store_true",
        help="Disable MCP tools (run model without tools)"
    )

    parser.add_argument(
        "--debug",
        action="store_true",
        help="Enable debug mode"
    )

    return parser.parse_args()


if __name__ == "__main__":
    args = parse_arguments()

    # Join message parts
    prompt_message = " ".join(args.prompt_message)

    # Determine if MCP tools should be used
    use_mcp = not args.no_mcp

    # Initialize the model based on selection
    if args.llm_model == "gemini":
        if not os.getenv("GOOGLE_API_KEY"):
            print("ERROR: GOOGLE_API_KEY environment variable is not set")
            sys.exit(1)
        llm_model = Gemini(id="gemini-2.0-flash-001")
    elif args.llm_model == "granite":
        # Import Granite model
        try:
            if not os.getenv("GRANITE_API_KEY"):
                print("ERROR: GRANITE_API_KEY environment variable is not set")
                sys.exit(1)
            llm_model = Granite(
                id="/data/granite-3.2-8b-instruct",
                name="IBM Granite",
                api_key=os.getenv("GRANITE_API_KEY"),
                verify_ssl=False,
                temperature=0.7,
                max_tokens=1000,
                debug_mode=args.debug
            )
        except ImportError:
            print("ERROR: Granite model (IBM_model.py) not found")
            print("Please ensure IBM_model.py is available in the parent directory")
            sys.exit(1)
    else:
        print(f"ERROR: Not supported model: {args.llm_model}")
        sys.exit(1)

    if args.debug:
        print(f"[DEBUG] Model: {args.llm_model}")
        print(f"[DEBUG] Message: {prompt_message}")

    asyncio.run(run_agent(llm_model, prompt_message, use_mcp, args.debug))
