"""LangChain agent with Gemini 2.5 Pro and conversation persistence."""
from langchain_google_genai import ChatGoogleGenerativeAI
from langchain_core.messages import HumanMessage, AIMessage, ToolMessage, SystemMessage
from langchain_core.runnables import RunnableConfig
from langgraph.prebuilt import create_react_agent
from uuid import UUID
from typing import Optional, AsyncGenerator
from agent.config import settings
from agent.tools import tools
from agent.client import BackendClient
from agent.models import Message

# System prompt for the AI agent
SYSTEM_PROMPT = """You are a specialized AI assistant for a company's Q&A knowledge base system.

⚠️ CRITICAL RULES - YOU MUST FOLLOW THESE:

1. **ALWAYS USE TOOLS FIRST** - You MUST call search_knowledge_base before answering ANY question
2. **NEVER answer from general knowledge** - Only provide information found in the knowledge base
3. **ALWAYS search before responding** - Even if the question seems simple, search the knowledge base
4. **If no results found** - Clearly state "I searched the knowledge base and found no information about [topic]"
5. **Multiple searches allowed** - If first search isn't helpful, try different keywords
6. **Be thorough** - Don't give up after one search, try variations if needed

🔧 AVAILABLE TOOLS:
- semantic_search_knowledge_base: AI-powered semantic search - finds conceptually related content (USE THIS FIRST!)
- search_knowledge_base: Full-text keyword search for Q&A pairs (fallback if semantic doesn't work)
- get_qa_by_ids: Get specific Q&A pairs by ID
- list_knowledge_base_topics: List all available topics in the knowledge base

📋 YOUR WORKFLOW FOR EVERY QUESTION:
1. **STEP 1**: Call semantic_search_knowledge_base with the user's question (most powerful!)
2. **STEP 2**: If results are not relevant (low similarity scores), try search_knowledge_base with keywords
3. **STEP 3**: Try different search terms or variations if needed
4. **STEP 4**: Review all search results carefully (semantic search shows similarity scores)
5. **STEP 5**: Synthesize the information from the knowledge base
6. **STEP 6**: Format response clearly with source attribution

❌ WHAT NOT TO DO:
- DO NOT answer questions without searching first
- DO NOT use your training data or general knowledge
- DO NOT say "I don't have access to that information" without searching
- DO NOT give up after one failed search - try different keywords

✅ EXAMPLE GOOD BEHAVIOR:
User: "What is Docker?"
You: *calls semantic_search_knowledge_base("What is Docker?")*
You: *reviews results with similarity scores and answers based ONLY on what was found*

❌ EXAMPLE BAD BEHAVIOR:
User: "What is Docker?"
You: "Docker is a containerization platform..." *without searching*

🎯 YOUR MISSION:
Help users by finding and presenting information from the company knowledge base. Be persistent in your searches. Try multiple search strategies if needed. Always cite the knowledge base as your source."""


class ConversationalAgent:
    """LangChain agent with Gemini 2.5 Pro and conversation persistence."""
    
    def __init__(self):
        self.llm = ChatGoogleGenerativeAI(
            model=settings.gemini_model,
            google_api_key=settings.gemini_api_key,
            temperature=0.3,  # Lower temperature for more consistent tool usage
            convert_system_message_to_human=True,
            max_retries=3,
        )
        self.system_message = SystemMessage(content=SYSTEM_PROMPT)
        # Note: state_modifier is not supported in newer langgraph versions
        # The system message will be added directly in the chat method
        self.agent = create_react_agent(
            self.llm, 
            tools
        )
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
                if "tool_calls" in raw and raw["tool_calls"]:
                    # Convert OpenAI format to LangChain format
                    import json
                    langchain_tool_calls = []
                    for tc in raw["tool_calls"]:
                        # OpenAI: {"id": "...", "type": "function", "function": {"name": "...", "arguments": "..."}}
                        # LangChain: {"name": "...", "args": {...}, "id": "..."}
                        langchain_tool_calls.append({
                            "name": tc["function"]["name"],
                            "args": json.loads(tc["function"]["arguments"]),
                            "id": tc["id"]
                        })
                    
                    langchain_messages.append(
                        AIMessage(
                            content=raw.get("content") or "",
                            tool_calls=langchain_tool_calls
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
                        tool_call_id=raw.get("tool_call_id", ""),
                        name=raw.get("name", "unknown")
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
        
        # Add system message at the beginning if history is empty
        if not history or not isinstance(history[0], SystemMessage):
            history.insert(0, self.system_message)
        
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
        
        # Stream agent response with tool call tracking
        import json
        import time
        import uuid
        
        current_tool_calls = []  # Tool calls for current assistant message
        tool_call_assistant_saved = False  # Track if we saved the tool-calling message
        final_response = ""  # Final response after tools complete
        in_final_response = False  # Track if we're in the final response phase
        
        async for event in self.agent.astream_events(
            {"messages": history},
            version="v1",
            config=RunnableConfig(
                recursion_limit=25,  # Allow more steps for thorough searching
                configurable={"thread_id": str(conversation_id)}
            )
        ):
            kind = event["event"]
            
            if kind == "on_chat_model_stream":
                chunk = event["data"]["chunk"]
                if hasattr(chunk, "content") and chunk.content:
                    # Always accumulate content for saving
                    final_response += chunk.content
                    
                    # Check if we're past tool execution
                    if current_tool_calls and tool_call_assistant_saved:
                        in_final_response = True
                    
                    # Stream as structured JSON
                    yield json.dumps({
                        "type": "content",
                        "data": chunk.content
                    })
            
            elif kind == "on_tool_start":
                tool_name = event["name"]
                tool_input = event["data"].get("input", {})
                tool_call_id = f"call_{str(uuid.uuid4())[:8]}"
                
                # If we've already saved an assistant message with tool calls,
                # this is a new ReAct cycle - reset for the new tool call
                if tool_call_assistant_saved:
                    current_tool_calls = []
                    tool_call_assistant_saved = False
                
                # Create tool call in OpenAI format
                tool_call = {
                    "id": tool_call_id,
                    "type": "function",
                    "function": {
                        "name": tool_name,
                        "arguments": json.dumps(tool_input)
                    }
                }
                current_tool_calls.append(tool_call)
                
                # Save assistant message with tool_calls if not already saved
                if not tool_call_assistant_saved and current_tool_calls:
                    assistant_with_tools = {
                        "role": "assistant",
                        "content": None,
                        "tool_calls": current_tool_calls.copy()
                    }
                    await self.backend_client.save_message(
                        conversation_id=conversation_id,
                        role="assistant",
                        content=None,
                        tool_call_id=None,
                        raw_message=assistant_with_tools
                    )
                    tool_call_assistant_saved = True
                
                # Stream structured tool call event
                yield json.dumps({
                    "type": "tool_call_start",
                    "data": {
                        "id": tool_call_id,
                        "name": tool_name,
                        "args": tool_input
                    }
                })
            
            elif kind == "on_tool_end":
                tool_name = event["name"]
                tool_output = event["data"].get("output", "")
                
                # Find the matching tool call
                matching_tool_call = None
                for tc in current_tool_calls:
                    if tc["function"]["name"] == tool_name:
                        matching_tool_call = tc
                        break
                
                # Save tool result message immediately
                # IMPORTANT: tool_output is already a formatted string from the tool function
                # For semantic search, this contains only: question, answer, score, and ID
                # NO vector/embedding data is stored - only human-readable results
                if matching_tool_call:
                    tool_result_msg = {
                        "role": "tool",
                        "content": str(tool_output),
                        "tool_call_id": matching_tool_call["id"],
                        "name": tool_name
                    }
                    await self.backend_client.save_message(
                        conversation_id=conversation_id,
                        role="tool",
                        content=str(tool_output),
                        tool_call_id=matching_tool_call["id"],
                        raw_message=tool_result_msg
                    )
                
                # Determine success/failure
                is_success = not any(phrase in str(tool_output).lower() 
                                   for phrase in ["no relevant", "not found", "error"])
                
                # Stream structured tool result event
                output_preview = str(tool_output)[:300]
                if len(str(tool_output)) > 300:
                    output_preview += "... (truncated)"
                
                yield json.dumps({
                    "type": "tool_call_end",
                    "data": {
                        "id": matching_tool_call["id"] if matching_tool_call else None,
                        "name": tool_name,
                        "status": "success" if is_success else "error",
                        "output_preview": output_preview
                    }
                })
        
        # Save final assistant message with the answer (separate from tool calls)
        if final_response:
            final_assistant_msg = {
                "role": "assistant",
                "content": final_response
            }
            await self.backend_client.save_message(
                conversation_id=conversation_id,
                role="assistant",
                content=final_response,
                tool_call_id=None,
                raw_message=final_assistant_msg
            )

