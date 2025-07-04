"""
Complete example of custom model implementations that use your own API endpoints.

This example shows:
1. Using OpenAILike for OpenAI-compatible APIs (simpler approach)
2. Creating a fully custom model implementation for non-OpenAI APIs with enhanced MCP support
"""

from dataclasses import dataclass
from os import getenv
from typing import Any, AsyncGenerator, Dict, Iterator, List, Optional, Type, Union
import json

import httpx
from pydantic import BaseModel

from agno.exceptions import ModelProviderError
from agno.models.base import Model
from agno.models.message import Message

from agno.models.response import ModelResponse
from agno.utils.log import log_error, log_warning


@dataclass
class FullyCustomModel(Model):
    """
    A fully custom model implementation for APIs that don't follow the OpenAI format.
    Enhanced with better MCP tool support.

    This is more complex but gives you complete control over the API interaction.
    """

    id: str = "/data/granite-3.2-8b-instruct"
    name: str = "/data/granite-3.2-8b-instruct"
    provider: str = "IBMGranite"

    # API configuration
    api_key: Optional[str] = getenv("GRANITE_API_KEY")
    base_url: str = "https://granite-3-2-8b-instruct--apicast-production.apps.int.stc.ai.prod.us-east-1.aws.paas.redhat.com/v1"
    verify_ssl: bool = False  # Skip SSL certificate verification
    timeout: float = 60.0

    # Model capabilities - Enhanced for MCP tool support
    supports_native_structured_outputs: bool = True
    supports_json_schema_outputs: bool = True
    supports_tool_calls: bool = True
    supports_function_calling: bool = True

    # Custom parameters for your model
    temperature: Optional[float] = None
    max_tokens: Optional[int] = None

    # HTTP client for API calls
    http_client: Optional[httpx.Client] = None

    # MCP-specific enhancements
    mcp_tools_enabled: bool = True
    tool_call_format: str = "openai"  # or "custom" for non-OpenAI format
    debug_mode: bool = False  # Keep debug mode off by default to reduce output noise

    def _get_headers(self) -> Dict[str, str]:
        """Get headers for API requests"""
        headers = {"Content-Type": "application/json"}

        if self.api_key:
            headers["Authorization"] = f"Bearer {self.api_key}"

        return headers

    def _format_messages(self, messages: List[Message]) -> List[Dict[str, Any]]:
        """Format messages for your custom API with enhanced tool support"""
        formatted_messages = []

        for message in messages:
            formatted = {
                "role": message.role,
                "content": message.content or "",
            }

            # Handle tool calls in messages
            if hasattr(message, 'tool_calls') and message.tool_calls:
                formatted["tool_calls"] = message.tool_calls

            # Handle tool call results
            if hasattr(message, 'tool_call_id') and message.tool_call_id:
                formatted["tool_call_id"] = message.tool_call_id
                formatted["name"] = getattr(message, 'name', '')

            formatted_messages.append(formatted)

        return formatted_messages

    def _format_tools_for_api(self, tools: List[Dict[str, Any]]) -> List[Dict[str, Any]]:
        """Format tools for the API request"""
        if not tools:
            return []

        formatted_tools = []
        for tool in tools:
            # Ensure tools are in the correct format for the API
            if self.tool_call_format == "openai":
                # Standard OpenAI format
                if "function" in tool:
                    # Already in OpenAI format
                    formatted_tool = {
                        "type": "function",
                        "function": tool["function"]
                    }
                else:
                    # Convert from MCP format to OpenAI format
                    formatted_tool = {
                        "type": "function",
                        "function": {
                            "name": tool.get("name", ""),
                            "description": tool.get("description", ""),
                            "parameters": tool.get("parameters", {})
                        }
                    }
            else:
                # Custom format - adjust as needed for your API
                formatted_tool = tool

            formatted_tools.append(formatted_tool)

        return formatted_tools

    def _create_request_body(self, messages: List[Message], tools: Optional[List[Dict[str, Any]]] = None, tool_choice: Optional[Union[str, Dict[str, Any]]] = None) -> Dict[str, Any]:
        """Create the request body for your API with enhanced tool support"""
        request_body = {
            "model": self.id,
            "messages": self._format_messages(messages),
        }

        # Add optional parameters if specified
        if self.temperature is not None:
            request_body["temperature"] = self.temperature

        if self.max_tokens is not None:
            request_body["max_tokens"] = self.max_tokens

        # Enhanced tool support
        if tools and len(tools) > 0 and self.supports_tool_calls:
            request_body["tools"] = self._format_tools_for_api(tools)

            # Add tool choice if specified
            if tool_choice:
                if isinstance(tool_choice, str):
                    request_body["tool_choice"] = tool_choice
                elif isinstance(tool_choice, dict):
                    request_body["tool_choice"] = tool_choice
                else:
                    # Default to auto if tool_choice is not recognized
                    request_body["tool_choice"] = "auto"
            else:
                # Default tool choice when tools are provided
                request_body["tool_choice"] = "auto"

        return request_body

    def invoke(
        self,
        messages: List[Message],
        response_format: Optional[Union[Dict, Type[BaseModel]]] = None,
        tools: Optional[List[Dict[str, Any]]] = None,
        tool_choice: Optional[Union[str, Dict[str, Any]]] = None,
    ) -> Any:
        """Send a request to your custom API with enhanced MCP tool support"""
        try:
            # Create HTTP client if not provided, with SSL verification disabled
            client = self.http_client or httpx.Client(timeout=self.timeout, verify=self.verify_ssl)

            # Prepare request
            url = f"{self.base_url}/chat/completions"
            headers = self._get_headers()
            request_body = self._create_request_body(messages, tools, tool_choice)

            # Debug prints for MCP tool debugging (only if debug_mode is explicitly enabled)
            if tools and self.mcp_tools_enabled and self.debug_mode:
                print(f"[MCP] Tools provided: {len(tools)} tools")
                print(f"[MCP] Tool choice: {tool_choice}")
                for i, tool in enumerate(tools):
                    if "function" in tool:
                        tool_name = tool["function"].get("name", f"tool_{i}")
                    else:
                        tool_name = tool.get("name", f"tool_{i}")
                    print(f"[MCP] Tool {i+1}: {tool_name}")

            # Only show detailed request info in debug mode
            if self.debug_mode:
                print("[DEBUG] Request URL:", url)
                print("[DEBUG] Request Body:", json.dumps(request_body, indent=2))
                print("[DEBUG] Headers:", {k: v if k != "Authorization" else "Bearer ***" for k, v in headers.items()})

            # Make API request
            response = client.post(url, json=request_body, headers=headers)

            if self.debug_mode:
                print("[DEBUG] Response Status:", response.status_code)
                print("[DEBUG] Response Body:", response.text)

            # Check for errors
            if not response.is_success:
                error_msg = f"API request failed with status {response.status_code}: {response.text}"
                log_error(error_msg)
                raise ModelProviderError(
                    message=error_msg,
                    status_code=response.status_code,
                    model_name=self.name,
                    model_id=self.id,
                )

            # Return raw response for parsing
            response_data = response.json()

            # Enhanced tool call logging
            if "choices" in response_data and len(response_data["choices"]) > 0:
                message = response_data["choices"][0].get("message", {})
                if message.get("tool_calls") and self.mcp_tools_enabled and self.debug_mode:
                    print(f"[MCP] Model requested {len(message['tool_calls'])} tool calls")
                    for i, tool_call in enumerate(message["tool_calls"]):
                        tool_name = tool_call.get("function", {}).get("name", "unknown")
                        print(f"[MCP] Tool call {i+1}: {tool_name}")

            return response_data

        except httpx.HTTPError as e:
            error_msg = f"HTTP error: {str(e)}"
            log_error(error_msg)
            raise ModelProviderError(message=error_msg, model_name=self.name, model_id=self.id) from e

        except Exception as e:
            error_msg = f"Unexpected error: {str(e)}"
            log_error(error_msg)
            raise ModelProviderError(message=error_msg, model_name=self.name, model_id=self.id) from e

    def parse_provider_response(
        self,
        response: Dict[str, Any],
        response_format: Optional[Union[Dict, Type[BaseModel]]] = None,
    ) -> ModelResponse:
        """
        Parse the API response into a ModelResponse with enhanced MCP tool support.
        """
        model_response = ModelResponse()

        try:
            # Get response message - handle different response formats
            if "choices" in response and len(response["choices"]) > 0:
                response_message = response["choices"][0]["message"]

                # Parse structured outputs if enabled
                try:
                    if (
                        response_format is not None
                        and isinstance(response_format, type)
                        and issubclass(response_format, BaseModel)
                    ):
                        parsed_object = response_message.get("parsed")
                        if parsed_object is not None:
                            model_response.parsed = parsed_object
                except Exception as e:
                    log_warning(f"Error retrieving structured outputs: {e}")

                # Add role
                if response_message.get("role") is not None:
                    model_response.role = response_message["role"]

                # Add content
                if response_message.get("content") is not None:
                    model_response.content = response_message["content"]

                # Enhanced tool call processing for MCP
                if response_message.get("tool_calls") is not None and len(response_message["tool_calls"]) > 0:
                    try:
                        tool_calls = response_message["tool_calls"]

                        # Validate and process tool calls
                        processed_tool_calls = []
                        for tool_call in tool_calls:
                            # Ensure tool call has required fields
                            if not tool_call.get("id"):
                                tool_call["id"] = f"call_{len(processed_tool_calls)}"

                            if not tool_call.get("type"):
                                tool_call["type"] = "function"

                            # Validate function call structure
                            if "function" in tool_call:
                                func = tool_call["function"]
                                if not func.get("name"):
                                    log_warning(f"Tool call missing function name: {tool_call}")
                                    continue

                                # Ensure arguments is a string (JSON)
                                if "arguments" in func and not isinstance(func["arguments"], str):
                                    try:
                                        func["arguments"] = json.dumps(func["arguments"])
                                    except Exception as e:
                                        log_warning(f"Failed to serialize tool call arguments: {e}")
                                        func["arguments"] = "{}"

                            processed_tool_calls.append(tool_call)

                        model_response.tool_calls = processed_tool_calls

                        if self.mcp_tools_enabled and self.debug_mode:
                            print(f"[MCP] Processed {len(processed_tool_calls)} tool calls")

                    except Exception as e:
                        log_warning(f"Error processing tool calls: {e}")

            # Handle alternative response formats (like direct text responses)
            elif "text" in response:
                model_response.content = response["text"]
                model_response.role = "assistant"
            elif "generated_text" in response:
                model_response.content = response["generated_text"]
                model_response.role = "assistant"
            else:
                # Fallback: try to extract content from the response
                log_warning(f"Unknown response format: {response}")
                model_response.content = str(response)
                model_response.role = "assistant"

            # Add usage information if available
            if response.get("usage") is not None:
                model_response.response_usage = response["usage"]

            return model_response

        except Exception as e:
            error_msg = f"Error parsing response: {str(e)}"
            log_error(error_msg)
            raise ModelProviderError(message=error_msg, model_name=self.name, model_id=self.id) from e

    async def ainvoke(
        self,
        messages: List[Message],
        response_format: Optional[Union[Dict, Type[BaseModel]]] = None,
        tools: Optional[List[Dict[str, Any]]] = None,
        tool_choice: Optional[Union[str, Dict[str, Any]]] = None,
    ) -> Any:
        """Send an asynchronous request to your custom API with enhanced MCP tool support"""
        try:
            # Create async HTTP client with SSL verification disabled
            async with httpx.AsyncClient(timeout=self.timeout, verify=self.verify_ssl) as client:
                # Prepare request
                url = f"{self.base_url}/chat/completions"
                headers = self._get_headers()
                request_body = self._create_request_body(messages, tools, tool_choice)

                # Debug prints for MCP tool debugging
                if tools and self.mcp_tools_enabled and self.debug_mode:
                    print(f"[MCP] Async tools provided: {len(tools)} tools")
                    print(f"[MCP] Async tool choice: {tool_choice}")

                # Make async API request
                response = await client.post(url, json=request_body, headers=headers)

                # Check for errors
                if not response.is_success:
                    error_msg = f"API request failed with status {response.status_code}: {response.text}"
                    log_error(error_msg)
                    raise ModelProviderError(
                        message=error_msg,
                        status_code=response.status_code,
                        model_name=self.name,
                        model_id=self.id,
                    )

                # Return raw response for parsing
                response_data = response.json()

                # Enhanced tool call logging for async
                if "choices" in response_data and len(response_data["choices"]) > 0:
                    message = response_data["choices"][0].get("message", {})
                    if message.get("tool_calls") and self.mcp_tools_enabled and self.debug_mode:
                        print(f"[MCP] Async model requested {len(message['tool_calls'])} tool calls")

                return response_data

        except httpx.HTTPError as e:
            error_msg = f"HTTP error: {str(e)}"
            log_error(error_msg)
            raise ModelProviderError(message=error_msg, model_name=self.name, model_id=self.id) from e

        except Exception as e:
            error_msg = f"Unexpected error: {str(e)}"
            log_error(error_msg)
            raise ModelProviderError(message=error_msg, model_name=self.name, model_id=self.id) from e

    def invoke_stream(
        self,
        messages: List[Message],
        response_format: Optional[Union[Dict, Type[BaseModel]]] = None,
        tools: Optional[List[Dict[str, Any]]] = None,
        tool_choice: Optional[Union[str, Dict[str, Any]]] = None,
    ) -> Iterator[Any]:
        """Stream responses from your custom API with MCP tool support"""
        try:
            # Create HTTP client if not provided, with SSL verification disabled
            client = self.http_client or httpx.Client(timeout=self.timeout, verify=self.verify_ssl)

            # Prepare request
            url = f"{self.base_url}/chat/completions"
            headers = self._get_headers()
            request_body = self._create_request_body(messages, tools, tool_choice)

            # Add streaming parameter
            request_body["stream"] = True

            # Make streaming API request
            with client.stream("POST", url, json=request_body, headers=headers) as response:
                if not response.is_success:
                    error_msg = f"API request failed with status {response.status_code}"
                    log_error(error_msg)
                    raise ModelProviderError(
                        message=error_msg,
                        status_code=response.status_code,
                        model_name=self.name,
                        model_id=self.id,
                    )

                # Process streaming response
                for chunk in response.iter_lines():
                    if not chunk:
                        continue

                    # Parse chunk (adjust based on your API's streaming format)
                    try:
                        # Handle server-sent events format
                        if chunk.startswith("data: "):
                            chunk_data = chunk[6:]  # Remove "data: " prefix
                            if chunk_data.strip() == "[DONE]":
                                break
                            yield json.loads(chunk_data)
                        else:
                            yield json.loads(chunk)
                    except Exception as e:
                        log_warning(f"Error parsing chunk: {e}")
                        continue

        except httpx.HTTPError as e:
            error_msg = f"HTTP error: {str(e)}"
            log_error(error_msg)
            raise ModelProviderError(message=error_msg, model_name=self.name, model_id=self.id) from e

        except Exception as e:
            error_msg = f"Unexpected error: {str(e)}"
            log_error(error_msg)
            raise ModelProviderError(message=error_msg, model_name=self.name, model_id=self.id) from e

    async def ainvoke_stream(
        self,
        messages: List[Message],
        response_format: Optional[Union[Dict, Type[BaseModel]]] = None,
        tools: Optional[List[Dict[str, Any]]] = None,
        tool_choice: Optional[Union[str, Dict[str, Any]]] = None,
    ) -> AsyncGenerator[Any, None]:
        """Stream responses asynchronously from your custom API with MCP tool support"""
        try:
            # Create async HTTP client with SSL verification disabled
            async with httpx.AsyncClient(timeout=self.timeout, verify=self.verify_ssl) as client:
                # Prepare request
                url = f"{self.base_url}/chat/completions"
                headers = self._get_headers()
                request_body = self._create_request_body(messages, tools, tool_choice)

                # Add streaming parameter
                request_body["stream"] = True

                # Make async streaming API request
                async with client.stream("POST", url, json=request_body, headers=headers) as response:
                    if not response.is_success:
                        error_msg = f"API request failed with status {response.status_code}"
                        log_error(error_msg)
                        raise ModelProviderError(
                            message=error_msg,
                            status_code=response.status_code,
                            model_name=self.name,
                            model_id=self.id,
                        )

                    # Process streaming response
                    async for line in response.aiter_lines():
                        if not line:
                            continue

                        # Parse chunk (adjust based on your API's streaming format)
                        try:
                            # Handle server-sent events format
                            if line.startswith("data: "):
                                chunk_data = line[6:]  # Remove "data: " prefix
                                if chunk_data.strip() == "[DONE]":
                                    break
                                yield json.loads(chunk_data)
                            else:
                                yield json.loads(line)
                        except Exception as e:
                            log_warning(f"Error parsing chunk: {e}")
                            continue

        except httpx.HTTPError as e:
            error_msg = f"HTTP error: {str(e)}"
            log_error(error_msg)
            raise ModelProviderError(message=error_msg, model_name=self.name, model_id=self.id) from e

        except Exception as e:
            error_msg = f"Unexpected error: {str(e)}"
            log_error(error_msg)
            raise ModelProviderError(message=error_msg, model_name=self.name, model_id=self.id) from e

    def parse_provider_response_delta(self, response_delta: Any) -> ModelResponse:
        """Parse streaming response chunks from your custom API with enhanced MCP tool support"""
        model_response = ModelResponse()

        try:
            # Extract content from streaming response (adjust based on your API's streaming format)
            if "choices" in response_delta and len(response_delta["choices"]) > 0:
                choice = response_delta["choices"][0]

                if "delta" in choice:
                    delta = choice["delta"]
                    model_response.content = delta.get("content", "")

                    if "role" in delta:
                        model_response.role = delta["role"]

                    # Enhanced tool call handling in streaming
                    if "tool_calls" in delta and delta["tool_calls"]:
                        try:
                            # Process streaming tool calls
                            tool_calls = delta["tool_calls"]
                            processed_tool_calls = []

                            for tool_call in tool_calls:
                                # Handle partial tool calls in streaming
                                if isinstance(tool_call, dict):
                                    # Ensure required fields exist
                                    if not tool_call.get("id") and tool_call.get("index") is not None:
                                        tool_call["id"] = f"call_stream_{tool_call['index']}"

                                    processed_tool_calls.append(tool_call)

                            model_response.tool_calls = processed_tool_calls

                            if self.mcp_tools_enabled and self.debug_mode:
                                print(f"[MCP] Streaming tool calls: {len(processed_tool_calls)}")

                        except Exception as e:
                            log_warning(f"Error processing streaming tool calls: {e}")

            # Add usage information if available
            if "usage" in response_delta:
                model_response.response_usage = response_delta["usage"]

            return model_response

        except Exception as e:
            error_msg = f"Error parsing response delta: {str(e)}"
            log_warning(error_msg)
            # Return empty response on error to continue streaming
            return model_response

    # MCP-specific utility methods
    def is_mcp_compatible(self) -> bool:
        """Check if the model is compatible with MCP tools"""
        return self.mcp_tools_enabled and self.supports_tool_calls

    def get_mcp_capabilities(self) -> Dict[str, Any]:
        """Get MCP-related capabilities of the model"""
        return {
            "supports_tool_calls": self.supports_tool_calls,
            "supports_function_calling": self.supports_function_calling,
            "supports_structured_outputs": self.supports_native_structured_outputs,
            "tool_call_format": self.tool_call_format,
            "mcp_tools_enabled": self.mcp_tools_enabled,
            "provider": self.provider,
            "model_id": self.id,
            "debug_mode": self.debug_mode
        }

    def enable_mcp_debug(self) -> None:
        """Enable MCP debugging mode"""
        self.debug_mode = True
        self.mcp_tools_enabled = True
        print("[MCP] Debug mode enabled")

    def disable_mcp_debug(self) -> None:
        """Disable MCP debugging mode"""
        self.debug_mode = False
        print("[MCP] Debug mode disabled")
