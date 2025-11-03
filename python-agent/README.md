# Python AI Agent - Smart Company Discovery Assistant

A type-safe Python conversational AI agent using LangChain, LangGraph, Google Gemini 2.0, and FastAPI. The agent intelligently searches the company knowledge base using multiple tools and provides context-aware answers with streaming responses.

## 📋 Table of Contents

- [Overview](#overview)
- [Technology Stack](#technology-stack)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [How the Agent Works](#how-the-agent-works)
- [LangChain Tools](#langchain-tools)
- [Setup Instructions](#setup-instructions)
- [Configuration](#configuration)
- [API Endpoints](#api-endpoints)
- [Integration with Go Backend](#integration-with-go-backend)
- [Conversation Persistence](#conversation-persistence)
- [Development](#development)
- [Testing](#testing)
- [Troubleshooting](#troubleshooting)

## Overview

The Python AI Agent is a sophisticated conversational assistant that helps users query the company knowledge base using natural language. Built on LangChain and powered by Google's Gemini 2.0, it automatically selects and executes the appropriate tools to search the knowledge base and provide accurate, source-backed answers.

### Key Features

✅ **Intelligent Tool Selection** - Agent automatically decides which search methods to use  
✅ **Dual Search Strategy** - Combines semantic (vector) and keyword (text) search  
✅ **Streaming Responses** - Real-time token-by-token response delivery  
✅ **Conversation Persistence** - All messages stored in PostgreSQL via Go backend  
✅ **Tool Call Transparency** - Users see which tools are being executed  
✅ **Type-Safe** - Full type hints throughout with Pydantic v2  
✅ **ReAct Pattern** - Reasoning and acting loop for complex queries  
✅ **Automatic Retry** - Graceful error handling with retries

## Technology Stack

### Core Framework
- **Python 3.11+** - Modern Python with type hints
- **FastAPI 0.109+** - Modern, fast web framework for APIs
- **Uvicorn** - ASGI server with HTTP/2 support

### AI & Orchestration
- **LangChain 0.1+** - Framework for LLM applications
- **LangChain Google GenAI 0.0.11+** - Google Gemini integration
- **LangGraph 0.0.20+** - State machine for complex agent workflows
- **Google Gemini 2.0 Flash** - Latest LLM from Google

### Type Safety & Validation
- **Pydantic v2** - Data validation and settings management
- **Pydantic Settings** - Environment variable management
- **Type Hints** - Full static typing throughout

### HTTP Client
- **httpx** - Modern async HTTP client for Python
- **python-dotenv** - Environment variable loading

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     FastAPI Application                          │
│  • CORS middleware        • Route handlers                      │
│  • Request validation     • Response streaming                  │
└────────────────┬────────────────────────────────────────────────┘
                 │
┌────────────────┴────────────────────────────────────────────────┐
│                   Conversational Agent                           │
│  • LangGraph ReAct Agent     • Gemini 2.0 Flash                │
│  • Conversation history      • Message persistence              │
│  • Tool orchestration        • Streaming events                 │
└────────────────┬────────────────────────────────────────────────┘
                 │
┌────────────────┴────────────────────────────────────────────────┐
│                      LangChain Tools                             │
│  ┌────────────────────┐  ┌────────────────────┐                │
│  │ semantic_search    │  │ search_knowledge   │                │
│  │ (Vector/Pinecone)  │  │ (Full-text/PG)     │                │
│  └────────────────────┘  └────────────────────┘                │
│  ┌────────────────────┐  ┌────────────────────┐                │
│  │ get_qa_by_ids      │  │ list_topics        │                │
│  └────────────────────┘  └────────────────────┘                │
└────────────────┬────────────────────────────────────────────────┘
                 │
┌────────────────┴────────────────────────────────────────────────┐
│                      Backend Client                              │
│  • HTTP client to Go backend                                    │
│  • Tool endpoint wrapper                                         │
│  • Message persistence                                           │
└────────────────┬────────────────────────────────────────────────┘
                 │
         ┌───────┴────────┐
         │   Go Backend   │
         │   :8080        │
         │                │
         │  • PostgreSQL  │
         │  • Pinecone    │
         │  • Embeddings  │
         └────────────────┘
```

### Agent Flow

```
User Question
     ↓
FastAPI receives request
     ↓
ConversationalAgent.chat()
     ↓
Load conversation history from backend
     ↓
Add system prompt + user message
     ↓
LangGraph ReAct Agent starts
     ↓
┌────────────────────────────────────┐
│  ReAct Loop (Reasoning + Acting)  │
│                                    │
│  1. Agent reasons about question   │
│  2. Decides which tool(s) to use   │
│  3. Executes tool(s)               │
│  4. Reviews tool output            │
│  5. Decides next action            │
│     - Call more tools?             │
│     - Generate final answer?       │
│  6. Repeat if needed               │
└────────────────────────────────────┘
     ↓
Stream response tokens to frontend
     ↓
Save all messages to backend
     ↓
Return complete response
```

## Project Structure

```
python-agent/
├── agent/
│   ├── __init__.py
│   ├── agent.py              # Main ConversationalAgent class
│   ├── client.py             # Backend HTTP client
│   ├── config.py             # Configuration and settings
│   ├── models.py             # Pydantic models
│   └── tools.py              # LangChain tool definitions
│
├── api/
│   ├── __init__.py
│   ├── routes.py             # FastAPI route handlers
│   └── schemas.py            # Request/response schemas
│
├── main.py                   # Application entry point
├── requirements.txt          # Python dependencies
├── Dockerfile               # Docker image definition
├── env.template             # Environment variable template
└── README.md                # This file
```

### Module Explanations

- **agent/agent.py**: Core agent logic with LangGraph integration
- **agent/tools.py**: LangChain tool definitions that wrap backend endpoints
- **agent/client.py**: HTTP client for communicating with Go backend
- **agent/config.py**: Pydantic settings for configuration management
- **agent/models.py**: Type-safe data models for messages and conversations
- **api/routes.py**: FastAPI endpoints for chat and conversations
- **api/schemas.py**: Request/response schemas for API validation

## How the Agent Works

### System Prompt Strategy

The agent uses a carefully designed system prompt that instructs it to:

1. **ALWAYS use BOTH search methods** - Semantic AND keyword search
2. **Never answer from general knowledge** - Only use knowledge base
3. **Always search before responding** - Even for simple questions
4. **Be transparent** - Explain when information is not found
5. **Try multiple strategies** - Rephrase searches if needed

**Key Instruction**:
```python
"⚠️ CRITICAL RULES - YOU MUST FOLLOW THESE:

1. **ALWAYS USE BOTH SEARCH METHODS** - You MUST call BOTH 
   semantic_search_knowledge_base AND search_knowledge_base for 
   EVERY question
   
2. **NEVER answer from general knowledge** - Only provide information 
   found in the knowledge base
   
3. **ALWAYS search before responding** - Even if the question seems 
   simple, search the knowledge base with both methods"
```

### ReAct Pattern (Reasoning + Acting)

The agent uses LangGraph's ReAct pattern:

```
┌─────────────────────────────────────────┐
│  1. REASON: "User asks about Docker"    │
│     "I need to search the knowledge     │
│      base with both methods"            │
└─────────────────┬───────────────────────┘
                  ↓
┌─────────────────┴───────────────────────┐
│  2. ACT: Call semantic_search(Docker)   │
│     Result: 3 relevant Q&A pairs        │
└─────────────────┬───────────────────────┘
                  ↓
┌─────────────────┴───────────────────────┐
│  3. ACT: Call search_knowledge(Docker)  │
│     Result: 2 relevant Q&A pairs        │
└─────────────────┬───────────────────────┘
                  ↓
┌─────────────────┴───────────────────────┐
│  4. REASON: "I have 5 unique results"   │
│     "I can now answer the question"     │
└─────────────────┬───────────────────────┘
                  ↓
┌─────────────────┴───────────────────────┐
│  5. RESPOND: Synthesize answer from     │
│     knowledge base with source citation │
└─────────────────────────────────────────┘
```

### Streaming Implementation

The agent streams responses as NDJSON (newline-delimited JSON):

```python
# Event types streamed to frontend
{
  "type": "content",
  "data": {"content": "Docker is..."}
}

{
  "type": "tool_call_start",
  "data": {"id": "call_123", "name": "semantic_search", "args": {...}}
}

{
  "type": "tool_call_end",
  "data": {"id": "call_123", "status": "success", "output_preview": "..."}
}

{
  "type": "error",
  "data": {"message": "Error details..."}
}
```

### Message Persistence

All messages are saved in **OpenAI format** to PostgreSQL:

```python
# User message
{
  "role": "user",
  "content": "What is Docker?"
}

# Assistant message with tool calls
{
  "role": "assistant",
  "content": null,
  "tool_calls": [
    {
      "id": "call_abc",
      "type": "function",
      "function": {
        "name": "semantic_search_knowledge_base",
        "arguments": "{\"query\": \"Docker\", \"top_k\": 5}"
      }
    }
  ]
}

# Tool result message
{
  "role": "tool",
  "content": "Result 1:\nQuestion: What is Docker?\nAnswer: ...",
  "tool_call_id": "call_abc",
  "name": "semantic_search_knowledge_base"
}

# Final assistant message
{
  "role": "assistant",
  "content": "Based on the knowledge base, Docker is..."
}
```

## LangChain Tools

### 1. semantic_search_knowledge_base

**Purpose**: AI-powered semantic similarity search using vector embeddings

```python
@tool
async def semantic_search_knowledge_base(
    query: str,
    top_k: int = 5
) -> str:
    """
    🎯 SEMANTIC SEARCH: Search using AI-powered semantic similarity
    (Pinecone vector search).
    """
```

**How it works**:
1. Query sent to Go backend `/tools/semantic-search-qa`
2. Backend generates embedding for query
3. Pinecone finds similar vectors
4. Returns Q&A pairs with similarity scores

**Output format**:
```
Result 1 (Similarity: 89.5%):
Question: What is Docker?
Answer: Docker is a containerization platform...
ID: uuid
```

### 2. search_knowledge_base

**Purpose**: Full-text keyword search for exact matches

```python
@tool
async def search_knowledge_base(
    query: str,
    limit: int = 5
) -> str:
    """
    🔍 KEYWORD SEARCH: Search using full-text keyword matching.
    """
```

**How it works**:
1. Query sent to Go backend `/tools/search-qa`
2. PostgreSQL performs full-text search
3. Returns matching Q&A pairs

**Output format**:
```
Result 1:
Question: What is Docker?
Answer: Docker is a containerization platform...
ID: uuid
```

### 3. get_qa_by_ids

**Purpose**: Retrieve specific Q&A pairs by UUID

```python
@tool
async def get_qa_by_ids(
    qa_ids: list[str]
) -> str:
    """
    Retrieve specific Q&A pairs by their IDs.
    """
```

**Use case**: When agent needs to reference previously found Q&A pairs

### 4. list_knowledge_base_topics

**Purpose**: List all available topics in the knowledge base

```python
@tool
async def list_knowledge_base_topics() -> str:
    """
    📋 List all available Q&A pairs in the knowledge base.
    """
```

**Use case**: When user asks "what topics do you cover?" or similar

### Tool Priority

Tools are ordered by importance:
```python
tools = [
    semantic_search_knowledge_base,  # Primary - AI-powered
    search_knowledge_base,           # Secondary - exact matches
    list_knowledge_base_topics,      # Helper - exploration
    get_qa_by_ids,                   # Utility - specific retrieval
]
```

The agent is instructed to use **BOTH** semantic and keyword search for every query.

## Setup Instructions

### Prerequisites

- **Python 3.11+** (3.12 recommended)
- **Google Gemini API Key** (required)
- **Go Backend running** at `http://localhost:8080`

### Option 1: Docker (Recommended)

```bash
# From project root
cd "/home/electron/projects/Amos Demo"

# Set Gemini API key in .env
echo "GEMINI_API_KEY=your-key-here" >> .env

# Start all services (includes python-agent)
docker-compose up -d

# View logs
docker-compose logs -f python-agent
```

Agent runs at: **http://localhost:8000**

### Option 2: Local Development

```bash
# Navigate to python-agent directory
cd python-agent

# Create virtual environment
python -m venv venv

# Activate virtual environment
source venv/bin/activate  # Linux/Mac
# OR
venv\Scripts\activate     # Windows

# Install dependencies
pip install -r requirements.txt

# Set environment variables
export GEMINI_API_KEY=your-gemini-api-key
export BACKEND_URL=http://localhost:8080

# Run the application
python main.py

# OR use uvicorn with hot reload
uvicorn main:app --reload --host 0.0.0.0 --port 8000
```

### Get Gemini API Key

1. Go to https://makersuite.google.com/app/apikey
2. Sign in with Google account
3. Click "Create API Key"
4. Copy the key

## Configuration

### Environment Variables

Create a `.env` file or set environment variables:

```bash
# Required: Gemini API Key
GEMINI_API_KEY=your-gemini-api-key-here

# Gemini Model Configuration
GEMINI_MODEL=gemini-2.0-flash-exp

# Go Backend URL
BACKEND_URL=http://localhost:8080

# API Server Configuration
API_HOST=0.0.0.0
API_PORT=8000

# Feature Flags
USE_PINECONE=false

# CORS Origins (for frontend)
CORS_ORIGINS=["http://localhost:5173","http://localhost:3000"]
```

### Configuration Management

The agent uses Pydantic Settings for type-safe configuration:

```python
# agent/config.py
from pydantic_settings import BaseSettings

class Settings(BaseSettings):
    gemini_api_key: str
    gemini_model: str = "gemini-2.0-flash-exp"
    backend_url: str = "http://localhost:8080"
    api_host: str = "0.0.0.0"
    api_port: int = 8000
    cors_origins: list[str] = ["http://localhost:5173"]
    use_pinecone: bool = False
    
    class Config:
        env_file = ".env"
        case_sensitive = False
```

## API Endpoints

### Health Check

```http
GET /health
```

**Response**:
```json
{
  "status": "healthy",
  "model": "gemini-2.0-flash-exp"
}
```

### Create Conversation

```http
POST /chat/conversations
Content-Type: application/json

{
  "title": "My Questions"
}
```

**Response**:
```json
{
  "conversation": {
    "id": "uuid",
    "title": "My Questions",
    "created_at": "2025-11-03T10:00:00Z",
    "updated_at": "2025-11-03T10:00:00Z"
  }
}
```

### Send Message (Streaming)

```http
POST /chat/conversations/{conversation_id}/messages
Content-Type: application/json

{
  "message": "What is Docker?"
}
```

**Response**: `application/x-ndjson` (streaming)

Each line is a JSON event:
```json
{"type":"tool_call_start","data":{"id":"call_123","name":"semantic_search_knowledge_base","args":{"query":"Docker","top_k":5}}}
{"type":"tool_call_end","data":{"id":"call_123","status":"success","output_preview":"Result 1..."}}
{"type":"content","data":{"content":"Based"}}
{"type":"content","data":{"content":" on"}}
{"type":"content","data":{"content":" the"}}
{"type":"content","data":{"content":" knowledge"}}
{"type":"content","data":{"content":" base,"}}
{"type":"content","data":{"content":" Docker"}}
{"type":"content","data":{"content":" is..."}}
```

### Get Messages

```http
GET /chat/conversations/{conversation_id}/messages
```

**Response**:
```json
{
  "messages": [
    {
      "id": "uuid",
      "conversation_id": "uuid",
      "role": "user",
      "content": "What is Docker?",
      "raw_message": {...},
      "created_at": "2025-11-03T10:00:00Z"
    },
    {
      "id": "uuid",
      "conversation_id": "uuid",
      "role": "assistant",
      "content": "Docker is...",
      "raw_message": {...},
      "created_at": "2025-11-03T10:00:01Z"
    }
  ]
}
```

## Integration with Go Backend

### Backend Client

The agent communicates with the Go backend via HTTP:

```python
# agent/client.py
class BackendClient:
    """HTTP client for Go backend communication."""
    
    async def search_qa(self, query: str, limit: int) -> SearchResponse:
        """Full-text search."""
        response = await self.client.post(
            f"{self.base_url}/tools/search-qa",
            json={"query": query, "limit": limit}
        )
        return SearchResponse(**response.json())
    
    async def semantic_search_qa(self, query: str, top_k: int) -> SemanticSearchResponse:
        """Semantic vector search."""
        response = await self.client.post(
            f"{self.base_url}/tools/semantic-search-qa",
            json={"query": query, "top_k": top_k}
        )
        return SemanticSearchResponse(**response.json())
    
    async def save_message(self, conversation_id: UUID, ...) -> Message:
        """Save message to database."""
        response = await self.client.post(
            f"{self.base_url}/tools/save-message",
            json={...}
        )
        return Message(**response.json()["message"])
```

### Backend Endpoints Used

| Endpoint | Purpose |
|----------|---------|
| `POST /tools/search-qa` | Full-text search |
| `POST /tools/semantic-search-qa` | Vector search |
| `POST /tools/get-qa-by-ids` | Get specific Q&A pairs |
| `POST /tools/save-message` | Save conversation message |
| `GET /api/conversations/{id}/messages` | Load conversation history |

## Conversation Persistence

### Message Flow

1. **User sends message** → Frontend
2. **Frontend POSTs** → Python Agent
3. **Agent saves user message** → Go Backend → PostgreSQL
4. **Agent processes** → Calls tools → Gets results
5. **Agent saves tool calls** → Go Backend → PostgreSQL
6. **Agent saves tool results** → Go Backend → PostgreSQL
7. **Agent generates response** → Streams to frontend
8. **Agent saves final response** → Go Backend → PostgreSQL

### Message Format

All messages stored in **OpenAI-compatible format** in the `messages` table:

```sql
CREATE TABLE messages (
    id UUID PRIMARY KEY,
    conversation_id UUID REFERENCES conversations(id),
    role TEXT CHECK (role IN ('user', 'assistant', 'tool', 'system')),
    content TEXT,
    tool_call_id TEXT,
    raw_message JSONB NOT NULL,
    created_at TIMESTAMPTZ
);
```

This allows:
- ✅ Replay conversations
- ✅ Continue conversations across sessions
- ✅ Analyze tool usage
- ✅ Debug agent behavior
- ✅ Export to other LLM platforms

## Development

### Project Setup

```bash
# Clone and setup
cd python-agent
python -m venv venv
source venv/bin/activate
pip install -r requirements.txt
```

### Development Server

```bash
# With auto-reload
uvicorn main:app --reload --host 0.0.0.0 --port 8000

# View API docs
open http://localhost:8000/docs
```

### Code Style

```bash
# Format code
black agent/ api/

# Sort imports
isort agent/ api/

# Type checking
mypy agent/ api/

# Linting
pylint agent/ api/
```

### Adding a New Tool

1. **Define tool** in `agent/tools.py`:
```python
@tool
async def my_new_tool(
    param: Annotated[str, "Parameter description"]
) -> str:
    """Tool description for the agent."""
    # Implementation
    return result
```

2. **Add to tools list**:
```python
tools = [
    semantic_search_knowledge_base,
    search_knowledge_base,
    my_new_tool,  # Add here
]
```

3. **Update system prompt** if needed to instruct agent on when to use it

### Debugging

```bash
# Enable debug logging
import logging
logging.basicConfig(level=logging.DEBUG)

# View streaming events
docker-compose logs -f python-agent

# Test tool directly
python -c "
from agent.tools import semantic_search_knowledge_base
import asyncio
result = asyncio.run(semantic_search_knowledge_base('Docker', 5))
print(result)
"
```

## Testing

### Manual Testing

```bash
# Test health endpoint
curl http://localhost:8000/health

# Create conversation
curl -X POST http://localhost:8000/chat/conversations \
  -H "Content-Type: application/json" \
  -d '{"title": "Test"}'

# Send message (streaming)
curl -X POST http://localhost:8000/chat/conversations/{id}/messages \
  -H "Content-Type: application/json" \
  -d '{"message": "What is Docker?"}'
```

### Unit Tests (Future)

```bash
# Install test dependencies
pip install pytest pytest-asyncio pytest-mock

# Run tests
pytest tests/ -v

# With coverage
pytest tests/ --cov=agent --cov=api
```

### Integration Tests (Future)

Test with real backend:

```python
# tests/test_integration.py
@pytest.mark.asyncio
async def test_chat_flow():
    # Create conversation
    conv = await create_conversation("Test")
    
    # Send message
    response = await send_message(conv.id, "What is Docker?")
    
    # Verify response
    assert "Docker" in response.content
```

## Troubleshooting

### Agent Not Starting

```bash
# Check Gemini API key
echo $GEMINI_API_KEY

# Check dependencies
pip list | grep langchain

# Reinstall
pip install --upgrade -r requirements.txt
```

### No Response from Agent

1. **Check backend is running**:
   ```bash
   curl http://localhost:8080/health
   ```

2. **Check agent logs**:
   ```bash
   docker-compose logs -f python-agent
   ```

3. **Verify tools are working**:
   ```bash
   curl -X POST http://localhost:8080/tools/search-qa \
     -H "Content-Type: application/json" \
     -d '{"query": "test", "limit": 5}'
   ```

### Agent Not Using Tools

1. **Check system prompt** - Ensure instructions are clear
2. **Lower temperature** - More deterministic behavior
3. **Check tool descriptions** - Should be clear and specific
4. **View agent reasoning** - Check logs for decision process

### Streaming Not Working

1. **Check CORS** - Frontend must be in allowed origins
2. **Check SSE support** - Browser must support EventSource
3. **Check response format** - Must be `application/x-ndjson`
4. **Test with curl** - Isolate frontend vs backend issues

### Memory Issues

```bash
# Limit conversation history
# agent/agent.py
history = history[-20:]  # Keep only last 20 messages

# Use smaller model
GEMINI_MODEL=gemini-1.5-flash
```

## Performance Optimization

### Current Optimizations

1. **Async HTTP** - All backend calls are async
2. **Streaming** - Responses stream immediately
3. **Connection pooling** - httpx client reuses connections
4. **Temperature tuning** - 0.3 for consistent tool usage

### Future Enhancements

1. **Caching** - Cache search results
2. **Parallel tools** - Call multiple tools simultaneously
3. **Smart history** - Summarize old messages
4. **Rate limiting** - Protect against abuse

## Related Documentation

- [Main README](../README.md) - Full application setup
- [Backend README](../backend/README.md) - Go backend API
- [Frontend README](../frontend/README.md) - React frontend
- [LangChain Docs](https://python.langchain.com/) - LangChain framework
- [LangGraph Docs](https://langchain-ai.github.io/langgraph/) - Agent workflows
- [Gemini API Docs](https://ai.google.dev/docs) - Google Gemini

---

**Built with LangChain, LangGraph, and Google Gemini 2.0** 🤖
