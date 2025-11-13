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

1. **ALWAYS USE BOTH SEARCH METHODS** - You MUST call BOTH semantic_search_knowledge_base AND search_knowledge_base for EVERY question
2. **NEVER answer from general knowledge** - Only provide information found in the knowledge base
3. **ALWAYS search before responding** - Even if the question seems simple, search the knowledge base with both methods
4. **If no results found** - Clearly state "I searched the knowledge base using both semantic and text search and found no information about [topic]"
5. **Multiple searches allowed** - Try variations with both search types if needed
6. **Be thorough** - Use both search methods to ensure comprehensive coverage

🔧 AVAILABLE TOOLS:
- semantic_search_knowledge_base: AI-powered semantic search - finds conceptually related content (ALWAYS USE THIS!)
- search_knowledge_base: Full-text keyword search for Q&A pairs (ALWAYS USE THIS TOO!)
- get_qa_by_ids: Get specific Q&A pairs by ID
- list_knowledge_base_topics: List all available topics in the knowledge base

📋 YOUR WORKFLOW FOR EVERY QUESTION:
1. **STEP 1**: Call semantic_search_knowledge_base with the user's question (AI-powered understanding!)
2. **STEP 2**: Call search_knowledge_base with relevant keywords (ensures exact matches aren't missed!)
3. **STEP 3**: Review all search results from BOTH methods carefully (semantic shows similarity scores)
4. **STEP 4**: Combine and deduplicate results from both searches
5. **STEP 5**: Synthesize the information from the knowledge base
6. **STEP 6**: Format response clearly with source attribution

🎯 WHY USE BOTH METHODS:
- **Semantic search** finds conceptually similar content even with different wording
- **Text search** finds exact keyword matches that semantic might miss
- **Using BOTH** maximizes recall and ensures comprehensive results
- **Belt and suspenders approach** - don't rely on just one method!

❌ WHAT NOT TO DO:
- DO NOT answer questions without using BOTH search methods first
- DO NOT use your training data or general knowledge
- DO NOT say "I don't have access to that information" without trying BOTH searches
- DO NOT give up after one search type - ALWAYS use both!
- DO NOT skip text search if semantic finds results - ALWAYS use both!

✅ EXAMPLE GOOD BEHAVIOR:
User: "What is Docker?"
You: *calls semantic_search_knowledge_base("What is Docker?")*
You: *calls search_knowledge_base("Docker")*
You: *reviews results from both methods and answers based ONLY on what was found*

❌ EXAMPLE BAD BEHAVIOR:
User: "What is Docker?"
You: *calls semantic_search_knowledge_base("What is Docker?")*
You: *answers immediately without also calling search_knowledge_base*

🎯 YOUR MISSION:
Help users by finding and presenting information from the company knowledge base. ALWAYS use both semantic and text search methods to ensure maximum coverage. Be persistent in your searches. Try multiple search strategies if needed. Always cite the knowledge base as your source."""


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
        
        # First pass: collect all tool_call_ids that have responses
        completed_tool_call_ids = set()
        for msg in messages:
            if msg.role == "tool":
                tool_call_id = msg.raw_message.get("tool_call_id")
                if tool_call_id:
                    completed_tool_call_ids.add(tool_call_id)
        
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
                        # Only include tool calls that have corresponding ToolMessages
                        if tc["id"] in completed_tool_call_ids:
                            # OpenAI: {"id": "...", "type": "function", "function": {"name": "...", "arguments": "..."}}
                            # LangChain: {"name": "...", "args": {...}, "id": "..."}
                            langchain_tool_calls.append({
                                "name": tc["function"]["name"],
                                "args": json.loads(tc["function"]["arguments"]),
                                "id": tc["id"]
                            })
                    
                    # Only add the AIMessage if it has valid tool calls or content
                    if langchain_tool_calls:
                        langchain_messages.append(
                            AIMessage(
                                content=raw.get("content") or "",
                                tool_calls=langchain_tool_calls
                            )
                        )
                    elif raw.get("content"):
                        # If no valid tool calls but has content, add as regular message
                        langchain_messages.append(
                            AIMessage(content=raw.get("content", ""))
                        )
                    # Otherwise skip this orphaned tool-calling message
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
        
        # Stream agent response
        import json
        
        async for event in self.agent.astream_events(
            {"messages": history},
            version="v1",
            config=RunnableConfig(
                recursion_limit=25,  # Allow more steps for thorough searching
                configurable={"thread_id": str(conversation_id)}
            )
        ):
            kind = event["event"]
            
            # Stream only essential LangChain events to frontend
            if kind == "on_chat_model_stream":
                chunk = event["data"]["chunk"]
                # Serialize the chunk data
                serializable_event = {
                    "event": kind,
                    "data": {
                        "chunk": {
                            "content": chunk.content if hasattr(chunk, "content") else "",
                            "tool_calls": chunk.tool_calls if hasattr(chunk, "tool_calls") and chunk.tool_calls else []
                        }
                    }
                }
                yield json.dumps({
                    "type": "langchain_event",
                    "data": serializable_event
                })
            
            elif kind == "on_tool_start":
                # Already serializable
                yield json.dumps({
                    "type": "langchain_event",
                    "data": {
                        "event": kind,
                        "name": event["name"],
                        "data": event["data"]
                    }
                })
            
            elif kind == "on_tool_end":
                # Extract ToolMessage and make it serializable
                tool_message = event["data"]["output"]
                yield json.dumps({
                    "type": "langchain_event",
                    "data": {
                        "event": kind,
                        "name": event["name"],
                        "data": {
                            "output": {
                                "tool_call_id": tool_message.tool_call_id,
                                "name": tool_message.name,
                                "content": tool_message.content
                            }
                        }
                    }
                })
            
            # Save messages to DB in OpenAI format
            if kind == "on_chat_model_end":
                # Extract the complete AIMessage
                ai_message = event["data"]["output"]["generations"][0][0]["message"]
                
                # Convert to OpenAI format with LLM's native tool_call IDs
                openai_msg = {
                    "role": "assistant",
                    "content": ai_message.content or "",
                }
                
                # Add tool_calls if present (uses LLM's UUIDs)
                if hasattr(ai_message, "tool_calls") and ai_message.tool_calls:
                    openai_msg["tool_calls"] = [
                        {
                            "id": tc["id"],
                            "type": "function",
                            "function": {
                                "name": tc["name"],
                                "arguments": json.dumps(tc["args"])
                            }
                        }
                        for tc in ai_message.tool_calls
                    ]
                
                # Save to database
                await self.backend_client.save_message(
                    conversation_id=conversation_id,
                    role="assistant",
                    content=openai_msg.get("content"),
                    tool_call_id=None,
                    raw_message=openai_msg
                )
            
            elif kind == "on_tool_end":
                # Extract the ToolMessage
                tool_message = event["data"]["output"]
                
                # Convert to OpenAI format with LLM's native tool_call ID
                openai_msg = {
                    "role": "tool",
                    "tool_call_id": tool_message.tool_call_id,
                    "name": tool_message.name,
                    "content": tool_message.content
                }
                
                # Save to database
                await self.backend_client.save_message(
                    conversation_id=conversation_id,
                    role="tool",
                    content=tool_message.content,
                    tool_call_id=tool_message.tool_call_id,
                    raw_message=openai_msg
                )

