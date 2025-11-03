"""LangChain agent with Gemini 2.5 Pro and conversation persistence."""
from langchain_google_genai import ChatGoogleGenerativeAI
from langchain_core.messages import HumanMessage, AIMessage, ToolMessage
from langchain_core.runnables import RunnableConfig
from langgraph.prebuilt import create_react_agent
from uuid import UUID
from typing import Optional, AsyncGenerator
from agent.config import settings
from agent.tools import tools
from agent.client import BackendClient
from agent.models import Message


class ConversationalAgent:
    """LangChain agent with Gemini 2.5 Pro and conversation persistence."""
    
    def __init__(self):
        self.llm = ChatGoogleGenerativeAI(
            model=settings.gemini_model,
            google_api_key=settings.gemini_api_key,
            temperature=0.7,
            convert_system_message_to_human=True
        )
        self.agent = create_react_agent(self.llm, tools)
        self.backend_client = BackendClient()
    
    async def load_conversation_history(
        self, 
        conversation_id: UUID
    ) -> list:
        """Load conversation history and convert to LangChain format."""
        messages = await self.backend_client.get_messages(conversation_id)
        
        langchain_messages = []
        for msg in messages:
            raw = msg.raw_message
            
            if msg.role == "user":
                langchain_messages.append(
                    HumanMessage(content=raw.get("content", ""))
                )
            elif msg.role == "assistant":
                if "tool_calls" in raw:
                    langchain_messages.append(
                        AIMessage(
                            content=raw.get("content", ""),
                            tool_calls=raw["tool_calls"]
                        )
                    )
                else:
                    langchain_messages.append(
                        AIMessage(content=raw.get("content", ""))
                    )
            elif msg.role == "tool":
                langchain_messages.append(
                    ToolMessage(
                        content=raw.get("content", ""),
                        tool_call_id=raw.get("tool_call_id", "")
                    )
                )
        
        return langchain_messages
    
    def langchain_to_openai_format(self, message) -> dict:
        """Convert LangChain message to OpenAI format."""
        if isinstance(message, HumanMessage):
            return {"role": "user", "content": message.content}
        elif isinstance(message, AIMessage):
            if message.tool_calls:
                return {
                    "role": "assistant",
                    "content": message.content or None,
                    "tool_calls": message.tool_calls
                }
            return {"role": "assistant", "content": message.content}
        elif isinstance(message, ToolMessage):
            return {
                "role": "tool",
                "content": message.content,
                "tool_call_id": message.tool_call_id
            }
        return {"role": "system", "content": str(message)}
    
    async def chat(
        self,
        conversation_id: UUID,
        user_message: str
    ) -> AsyncGenerator[str, None]:
        """
        Process user message and stream agent response.
        Saves all messages to backend.
        """
        # Load history
        history = await self.load_conversation_history(conversation_id)
        
        # Save user message
        user_msg_dict = {"role": "user", "content": user_message}
        await self.backend_client.save_message(
            conversation_id=conversation_id,
            role="user",
            content=user_message,
            tool_call_id=None,
            raw_message=user_msg_dict
        )
        
        # Add user message to history
        history.append(HumanMessage(content=user_message))
        
        # Stream agent response
        full_response = ""
        tool_calls = []
        
        async for event in self.agent.astream_events(
            {"messages": history},
            version="v1",
            config=RunnableConfig(recursion_limit=10)
        ):
            kind = event["event"]
            
            if kind == "on_chat_model_stream":
                chunk = event["data"]["chunk"]
                if hasattr(chunk, "content") and chunk.content:
                    full_response += chunk.content
                    yield chunk.content
            
            elif kind == "on_tool_start":
                tool_name = event["name"]
                tool_input = event["data"].get("input", {})
                yield f"\n[Using tool: {tool_name}]\n"
            
            elif kind == "on_tool_end":
                tool_output = event["data"].get("output", "")
                yield f"\n[Tool result received]\n"
        
        # Save assistant message
        assistant_msg_dict = {"role": "assistant", "content": full_response}
        await self.backend_client.save_message(
            conversation_id=conversation_id,
            role="assistant",
            content=full_response,
            tool_call_id=None,
            raw_message=assistant_msg_dict
        )

