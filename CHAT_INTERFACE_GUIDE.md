# Chat Interface Guide - Multi-Conversation Support

## ✅ Status: All Services Running!

```
✅ Go Backend       → http://localhost:8080 (Q&A data, conversations)
✅ Python Agent     → http://localhost:8000 (AI chat agent)
✅ Frontend         → http://localhost:5173 (React UI)
✅ PostgreSQL       → localhost:5432 (Database)
✅ Pinecone Local   → localhost:5081 (Vector search)
```

---

## 🎯 What Was Changed

### Backend Architecture
- **Go Backend (port 8080)**: Manages Q&A pairs and conversation storage
- **Python Agent (port 8000)**: Handles AI chat with Gemini and streaming responses
- **Frontend connects to BOTH**:
  - Port 8080 for Q&A management and conversation CRUD operations
  - Port 8000 for sending messages and getting AI responses

### Frontend Features Added

#### ✨ Multi-Conversation Support
- **Sidebar**: Shows all your conversations
- **Create New**: Start multiple independent chats
- **Switch Between**: Click any conversation to view its history
- **Delete**: Remove conversations you don't need
- **Streaming Responses**: See AI responses appear in real-time

#### 🔄 Architecture Flow
```
User Question
     ↓
Frontend (React)
     ↓
Python Agent (port 8000) ← Streams AI response
     ↓
Go Backend (port 8080) ← Stores messages & fetches Q&A data
     ↓
PostgreSQL ← Persists all data
```

---

## 🚀 How to Use the Chat Interface

### 1. Access the Chat Page
Open http://localhost:5173 and navigate to the **Chat** page from the navigation menu.

### 2. Create Your First Conversation
Click the **"New Chat"** button in the left sidebar. This creates a new conversation.

### 3. Ask Questions
Type your question in the input box at the bottom and press Enter or click Send. The AI will:
- Search the Q&A knowledge base
- Use tools to find relevant information
- Stream the response back to you in real-time

### 4. Manage Multiple Conversations
- **Switch**: Click any conversation in the sidebar to view it
- **Delete**: Hover over a conversation and click the trash icon
- **History**: All messages are saved and will reappear when you select a conversation

---

## 🛠️ Technical Details

### Frontend API Configuration

The frontend uses these environment variables (with defaults):

```bash
# Backend API (Go server)
VITE_API_BASE_URL=http://localhost:8080

# Python Agent (AI chat)
VITE_PYTHON_AGENT_URL=http://localhost:8000
```

**Note:** These are already set to the correct defaults in the code, so no `.env` file is needed for local development!

### API Endpoints Used

#### Go Backend (port 8080)
```bash
# Conversations
GET    /api/conversations              # List all
POST   /api/conversations              # Create new
GET    /api/conversations/:id          # Get one
DELETE /api/conversations/:id          # Delete
GET    /api/conversations/:id/messages # Get messages

# Q&A Management (separate page)
GET    /api/qa-pairs                   # List Q&A pairs
POST   /api/qa-pairs                   # Create
PUT    /api/qa-pairs/:id               # Update
DELETE /api/qa-pairs/:id               # Delete
```

#### Python Agent (port 8000)
```bash
# Chat
POST   /chat/conversations              # Create conversation
POST   /chat/conversations/:id/messages # Send message (streaming)
GET    /chat/conversations/:id/messages # Get messages

# Health
GET    /health                          # Check status
```

---

## 🧪 Testing the Chat

### 1. Test Python Agent Directly
```bash
# Check health
curl http://localhost:8000/health

# Expected:
# {"status":"healthy","model":"gemini-2.0-flash-exp"}
```

### 2. Test Creating a Conversation via Go Backend
```bash
curl -X POST http://localhost:8080/api/conversations \
  -H "Content-Type: application/json" \
  -d '{"title":"Test Chat"}'
```

### 3. Test in the UI
1. Open http://localhost:5173
2. Go to "Chat" page
3. Click "New Chat"
4. Ask: "What Q&A pairs do you have?"
5. Watch the AI search your knowledge base and respond!

---

## 🔍 How the AI Agent Works

The Python agent has access to tools:

### Available Tools
1. **search_qa**: Full-text search across Q&A pairs
2. **get_qa_by_ids**: Retrieve specific Q&A pairs
3. **Semantic search**: Vector-based similarity search (when configured)

### Example Interaction
```
User: "Tell me about Docker"
  ↓
AI Agent:
  1. Uses search_qa tool to find relevant Q&A pairs
  2. Finds Q&A about Docker
  3. Synthesizes response from the knowledge base
  4. Streams response to frontend
```

---

## ⚙️ Configuration

### Python Agent Setup (Optional)

The Python agent requires a Gemini API key. To set it up:

1. **Get API Key**: https://makersuite.google.com/app/apikey

2. **Set Environment Variable**:
```bash
# Option 1: Create .env in project root
echo "GEMINI_API_KEY=your-key-here" > /home/electron/projects/Amos\ Demo/.env

# Option 2: Export in shell
export GEMINI_API_KEY=your-key-here
```

3. **Restart Services**:
```bash
docker-compose down
docker-compose up -d
```

**Current Status:** The agent will use a mock/test mode if no API key is provided (for development).

---

## 🐛 Troubleshooting

### Python Agent Not Responding?
```bash
# Check logs
docker logs smart-discovery-python-agent --tail 50

# Restart it
docker-compose restart python-agent
```

### Frontend Can't Connect?
1. Check both backends are running:
```bash
curl http://localhost:8080/health  # Go backend
curl http://localhost:8000/health  # Python agent
```

2. Check browser console (F12) for errors

3. Make sure CORS is working (you should see these headers in network tab):
   - `Access-Control-Allow-Origin: *`
   - `Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS`

### Conversations Not Saving?
```bash
# Check database connection
docker logs smart-discovery-backend --tail 30

# Check if PostgreSQL is running
docker-compose ps postgres
```

### Streaming Not Working?
The frontend uses the Fetch API with ReadableStream. This requires:
- Modern browser (Chrome 80+, Firefox 75+, Safari 14.1+)
- No proxy blocking streaming responses

---

## 📊 Architecture Diagram

```
┌─────────────────────────────────────────────────┐
│           Frontend (React + Vite)               │
│              Port: 5173                         │
└───────────┬─────────────────────┬───────────────┘
            │                     │
   ┌────────▼────────┐   ┌────────▼─────────┐
   │   Go Backend    │   │  Python Agent    │
   │   Port: 8080    │   │   Port: 8000     │
   │                 │   │                  │
   │ • Q&A CRUD      │   │ • AI Chat        │
   │ • Conversations │   │ • Tool Calls     │
   │ • Messages      │   │ • Streaming      │
   └────────┬────────┘   └────────┬─────────┘
            │                     │
            └──────────┬──────────┘
                       │
            ┌──────────▼───────────┐
            │    PostgreSQL        │
            │     Port: 5432       │
            │                      │
            │ • Conversations      │
            │ • Messages           │
            │ • Q&A Pairs          │
            └──────────────────────┘
```

---

## 🎉 Ready to Chat!

Your multi-conversation chat interface is now fully functional!

**Quick Start:**
1. Open http://localhost:5173
2. Navigate to "Chat"
3. Click "New Chat"
4. Start asking questions!

**Tips:**
- Create separate conversations for different topics
- The AI will use your Q&A knowledge base to answer
- All conversations are saved in the database
- You can have unlimited conversations

---

## 📚 Related Documentation

- [Startup Commands](./STARTUP_COMMANDS.md)
- [Frontend-Backend Connection](./FRONTEND_BACKEND_CONNECTION.md)
- [Main README](./README.md)

