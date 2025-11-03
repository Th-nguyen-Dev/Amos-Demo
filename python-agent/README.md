# Python LangChain AI Agent

A strictly-typed Python conversational AI agent using LangChain, Gemini 2.5 Pro, and FastAPI. This agent integrates with the Go backend to provide intelligent Q&A capabilities using the knowledge base.

## Features

- **🤖 Gemini 2.5 Pro Integration**: Powered by Google's latest language model
- **🔧 LangChain Tools**: Automatic tool selection for knowledge base queries
- **💬 OpenAI Message Format**: Fully compatible with standard message formats
- **📦 Type-Safe**: Strict type hints throughout with Pydantic v2
- **🔄 Streaming Responses**: Real-time streaming for better UX
- **💾 Conversation Persistence**: All messages saved to PostgreSQL via Go backend
- **🌐 React-Ready**: CORS-enabled FastAPI endpoints for React frontend
- **🔍 Pinecone-Ready**: Stubs in place for future semantic search integration

## Architecture

The agent wraps Go backend endpoints as LangChain tools:
- **search_knowledge_base**: Full-text search of Q&A pairs
- **get_qa_by_ids**: Retrieve specific Q&A pairs by UUID
- **semantic_search_knowledge_base**: Pinecone vector search (stub - falls back to text search)

The AI agent automatically decides which tools to use based on user queries.

## Prerequisites

- Python 3.11+
- Go backend running on `http://localhost:8080`
- Google Gemini API key

## Installation

1. **Navigate to the python-agent directory**:
   ```bash
   cd python-agent
   ```

2. **Create a virtual environment**:
   ```bash
   python -m venv venv
   source venv/bin/activate  # On Windows: venv\Scripts\activate
   ```

3. **Install dependencies**:
   ```bash
   pip install -r requirements.txt
   ```

4. **Configure environment variables**:
   ```bash
   cp env.template .env
   # Edit .env and add your GEMINI_API_KEY
   ```

## Configuration

Edit `.env` file with your settings:

```env
# Required
GEMINI_API_KEY=your_actual_api_key_here

# Optional (defaults shown)
GEMINI_MODEL=gemini-2.0-flash-exp
BACKEND_URL=http://localhost:8080
USE_PINECONE=false
API_HOST=0.0.0.0
API_PORT=8000
CORS_ORIGINS=["http://localhost:3000"]
```

## Usage

### Start the FastAPI Server

```bash
python main.py
```

The API will be available at `http://localhost:8000`

### API Documentation

FastAPI provides automatic interactive documentation:
- Swagger UI: http://localhost:8000/docs
- ReDoc: http://localhost:8000/redoc

## API Endpoints

### 1. Health Check
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

### 2. Create Conversation
```http
POST /chat/conversations
Content-Type: application/json

{
  "title": "Customer Support Chat"
}
```

**Response**:
```json
{
  "conversation": {
    "id": "uuid-here",
    "title": "Customer Support Chat",
    "created_at": "2025-11-03T10:00:00Z",
    "updated_at": "2025-11-03T10:00:00Z"
  }
}
```

### 3. Send Message (Streaming)
```http
POST /chat/conversations/{conversation_id}/messages
Content-Type: application/json

{
  "message": "What is your refund policy?"
}
```

**Response**: Server-Sent Events (text/plain streaming)

### 4. Get Conversation Messages
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
      "content": "What is your refund policy?",
      "raw_message": {"role": "user", "content": "..."},
      "created_at": "2025-11-03T10:00:00Z"
    }
  ]
}
```

## React Integration

Example React component for streaming chat:

```typescript
import { useState, useEffect } from 'react';

function Chat({ conversationId }: { conversationId: string }) {
  const [message, setMessage] = useState('');
  const [response, setResponse] = useState('');

  const sendMessage = async () => {
    const res = await fetch(
      `http://localhost:8000/chat/conversations/${conversationId}/messages`,
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ message })
      }
    );

    const reader = res.body?.getReader();
    const decoder = new TextDecoder();

    while (true) {
      const { done, value } = await reader!.read();
      if (done) break;
      
      const chunk = decoder.decode(value);
      setResponse(prev => prev + chunk);
    }
  };

  return (
    <div>
      <input 
        value={message} 
        onChange={e => setMessage(e.target.value)} 
      />
      <button onClick={sendMessage}>Send</button>
      <div>{response}</div>
    </div>
  );
}
```

## Project Structure

```
python-agent/
├── agent/
│   ├── __init__.py
│   ├── config.py          # Pydantic Settings configuration
│   ├── models.py          # Type-safe Pydantic models (OpenAI format)
│   ├── client.py          # Go backend API client
│   ├── tools.py           # LangChain tools (wraps backend endpoints)
│   └── agent.py           # Main agent logic with Gemini
├── api/
│   ├── __init__.py
│   ├── schemas.py         # FastAPI request/response schemas
│   └── routes.py          # FastAPI routes
├── main.py                # FastAPI application entry point
├── requirements.txt       # Python dependencies
├── env.template          # Environment variable template
└── README.md             # This file
```

## How It Works

1. **User sends message** via FastAPI endpoint
2. **Agent loads conversation history** from Go backend
3. **Gemini processes the message** and decides which tools to use
4. **Tools call Go backend** to search knowledge base
5. **Agent streams response** back to client
6. **All messages saved** to PostgreSQL via Go backend in OpenAI format

## Pinecone Integration (Future)

Currently, `semantic_search_knowledge_base` falls back to full-text search. To enable Pinecone:

1. Configure Pinecone in Go backend
2. Set `USE_PINECONE=true` in `.env`
3. Update `agent/client.py` to call Pinecone endpoint

## Development

### Run with auto-reload:
```bash
uvicorn main:app --reload --host 0.0.0.0 --port 8000
```

### Type checking:
```bash
mypy agent/ api/
```

### Linting:
```bash
ruff check agent/ api/
```

## Troubleshooting

### "Failed to connect to backend"
- Ensure Go backend is running on `http://localhost:8080`
- Check `BACKEND_URL` in `.env`

### "Invalid API key"
- Verify your `GEMINI_API_KEY` in `.env`
- Ensure the API key has proper permissions

### CORS errors
- Add your frontend URL to `CORS_ORIGINS` in `.env`
- Format: `["http://localhost:3000", "http://localhost:5173"]`

## License

Same as parent project

