# Smart Company Discovery Assistant

A full-stack internal knowledge management tool that combines traditional database management with AI-powered natural language processing. Manage company Q&A information and query it using intelligent, context-aware AI conversations.

## 📋 Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Technology Stack](#technology-stack)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Detailed Setup](#detailed-setup)
- [Environment Configuration](#environment-configuration)
- [Using the Application](#using-the-application)
- [API Reference](#api-reference)
- [Project Structure](#project-structure)
- [Development](#development)
- [Troubleshooting](#troubleshooting)
- [Documentation](#documentation)

## Overview

**Smart Company Discovery Assistant** helps internal teams efficiently manage and query company knowledge through:

- **Q&A Knowledge Base Management**: Full CRUD operations for question-answer pairs
- **AI-Powered Chat**: Natural language queries using Google's Gemini 2.0 with LangChain
- **Semantic Search**: Vector-based similarity search with Pinecone
- **Full-Text Search**: PostgreSQL-based keyword search
- **Modern Web Interface**: React-based UI with real-time updates

### Key Features

✅ **Complete Q&A Management** - Create, read, update, and delete knowledge base entries  
✅ **Conversational AI** - Ask questions in natural language and get intelligent responses  
✅ **Dual Search Modes** - Combine keyword and semantic search for best results  
✅ **Automatic Indexing** - Q&A pairs are automatically vectorized and indexed  
✅ **Conversation History** - All chat interactions are persisted in PostgreSQL  
✅ **Tool-Based Agent** - LangChain agent with multiple tools for knowledge retrieval  
✅ **Streaming Responses** - Real-time AI response streaming for better UX

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Frontend (React)                        │
│                    http://localhost:5173                        │
│  • Q&A Management UI       • AI Chat Interface                 │
│  • Redux State Management  • Real-time Streaming                │
└────────────────┬────────────────────────────────────────────────┘
                 │
                 ├──────────────────┬────────────────────────────┐
                 ▼                  ▼                            ▼
    ┌────────────────────┐  ┌──────────────────┐   ┌──────────────────────┐
    │  Backend (Go)      │  │  AI Agent        │   │  Vector Search       │
    │  :8080             │  │  (Python)        │   │  (Pinecone)          │
    │                    │  │  :8000           │   │  :5081 (local)       │
    │  • REST API        │  │                  │   │                      │
    │  • Q&A Service     │◄─┤  • LangChain     │   │  • Semantic Search   │
    │  • Conversation    │  │  • Gemini 2.0    │   │  • Vector Embeddings │
    │  • Embeddings      │  │  • Tools/Agents  │   │  • Similarity Match  │
    └─────────┬──────────┘  └──────────────────┘   └──────────────────────┘
              │
              ▼
    ┌─────────────────────┐
    │  PostgreSQL         │
    │  :5432              │
    │                     │
    │  • qa_pairs         │
    │  • conversations    │
    │  • messages         │
    └─────────────────────┘
```

### Data Flow

1. **Q&A Management**: User creates/updates Q&A → Saved to PostgreSQL → Auto-embedded → Indexed in Pinecone
2. **AI Chat**: User asks question → Python Agent selects tools → Searches knowledge base → LLM generates response → Saved to PostgreSQL
3. **Search**: Query → Generate embedding → Find similar vectors → Fetch Q&A from database → Return results

## Technology Stack

### Backend (Go)
- **Framework**: Gin (HTTP router)
- **Database**: PostgreSQL 16 with sqlx
- **Vector DB**: Pinecone (Official Go SDK v4)
- **Embeddings**: Google Gemini Embedding API (google.golang.org/genai)
- **Architecture**: Clean layered architecture (handlers → services → repositories)

### Frontend (React)
- **Framework**: React 19 + TypeScript
- **Build Tool**: Vite 7
- **UI Components**: ShadCN UI (Radix UI + Tailwind CSS)
- **State Management**: Redux Toolkit + RTK Query
- **Routing**: React Router v7
- **Styling**: Tailwind CSS 4

### AI Agent (Python)
- **Framework**: FastAPI
- **AI Orchestration**: LangChain + LangGraph
- **LLM**: Google Gemini 2.0 Flash (via langchain-google-genai)
- **Tools**: Custom LangChain tools for knowledge base integration
- **Type Safety**: Pydantic v2 with strict typing

### Infrastructure
- **Containerization**: Docker + Docker Compose
- **Database**: PostgreSQL 16 Alpine
- **Vector Search**: Pinecone Local (testing) / Pinecone Cloud (production)

## Prerequisites

### Required
- **Docker** and **Docker Compose** (recommended for easiest setup)
- **Node.js 18+** and **npm** (for frontend development)
- **Go 1.24+** (for backend development)
- **Python 3.11+** (for AI agent development)

### API Keys
- **Google Gemini API Key** (required for AI agent) - Get it at https://makersuite.google.com/app/apikey
- **Pinecone API Key** (optional, for production vector search) - Get it at https://app.pinecone.io/

### Optional
- **Google Cloud Project** (optional, for production embeddings)
- **Make** (for convenience commands)

## Quick Start

The fastest way to run the entire application:

### 1. Clone and Navigate

```bash
cd "/home/electron/projects/Amos Demo"
```

### 2. Set Up Environment Variables

Create a `.env` file in the project root:

```bash
cat > .env << 'EOF'
# Required for AI Agent
GEMINI_API_KEY=your-gemini-api-key-here

# Pinecone (Local mode - no key needed)
PINECONE_API_KEY=pclocal
PINECONE_HOST=http://pinecone-local:5081
PINECONE_INDEX_NAME=qa-index
PINECONE_NAMESPACE=default

# Optional: Google Embeddings (or uses mock)
GOOGLE_API_KEY=your-google-api-key
GOOGLE_PROJECT_ID=your-project-id
EOF
```

### 3. Start Backend Services (Docker)

```bash
docker-compose up -d
```

This starts:
- PostgreSQL (port 5432)
- Pinecone Local (port 5081)
- Go Backend API (port 8080)
- Python AI Agent (port 8000)

### 4. Start Frontend

```bash
cd frontend
npm install
npm run dev
```

Frontend runs at: **http://localhost:5173**

### 5. Verify Everything Works

```bash
# Check backend
curl http://localhost:8080/health

# Check AI agent
curl http://localhost:8000/health

# Open browser
open http://localhost:5173
```

That's it! You now have:
- ✅ Backend API at http://localhost:8080
- ✅ AI Agent at http://localhost:8000
- ✅ Frontend UI at http://localhost:5173
- ✅ PostgreSQL at localhost:5432
- ✅ Pinecone Local at localhost:5081

## Detailed Setup

### Option A: Full Docker Setup (Recommended)

All services run in Docker containers:

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down

# Stop and remove all data
docker-compose down -v
```

### Option B: Hybrid Setup (Frontend Local)

Backend in Docker, frontend local (better for frontend development):

```bash
# Terminal 1: Start backend services
docker-compose up -d

# Terminal 2: Start frontend locally
cd frontend
npm install
npm run dev
```

### Option C: Full Local Development

Run all services locally (best for development):

**Backend:**
```bash
cd backend

# Set environment variables
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=smart_discovery
export DB_USER=postgres
export DB_PASSWORD=postgres

# Run migrations (first time only)
psql -U postgres -d smart_discovery -f migrations/001_init_schema.sql

# Build and run
go build -o bin/server cmd/server/main.go
./bin/server
```

**Frontend:**
```bash
cd frontend
npm install
npm run dev
```

**AI Agent:**
```bash
cd python-agent

# Create virtual environment
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate

# Install dependencies
pip install -r requirements.txt

# Set environment variables
export GEMINI_API_KEY=your-key
export BACKEND_URL=http://localhost:8080

# Run
python main.py
```

## Environment Configuration

### Backend (.env or environment variables)

```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=smart_discovery
DB_SSLMODE=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5

# Server Configuration
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
SERVER_ENVIRONMENT=development

# Pinecone Configuration (Local Mode)
PINECONE_API_KEY=pclocal
PINECONE_HOST=http://localhost:5081
PINECONE_INDEX_NAME=qa-index
PINECONE_NAMESPACE=default

# Pinecone Configuration (Cloud Mode)
# PINECONE_API_KEY=your-real-key
# PINECONE_HOST=
# PINECONE_ENVIRONMENT=us-west1-gcp
# PINECONE_INDEX_NAME=qa-index

# Google Embedding Configuration (Optional)
GOOGLE_API_KEY=your-google-api-key
GOOGLE_PROJECT_ID=your-project-id
GOOGLE_LOCATION=us-central1
GOOGLE_EMBEDDING_MODEL=gemini-embedding-001
```

### Python Agent (.env)

```bash
# Required: Gemini API Key
GEMINI_API_KEY=your-gemini-api-key

# Gemini Configuration
GEMINI_MODEL=gemini-2.0-flash-exp

# Backend Configuration
BACKEND_URL=http://localhost:8080

# API Configuration
API_HOST=0.0.0.0
API_PORT=8000

# Feature Flags
USE_PINECONE=false

# CORS (for frontend)
CORS_ORIGINS=["http://localhost:5173","http://localhost:3000"]
```

### Frontend (.env)

```bash
VITE_API_BASE_URL=http://localhost:8080
```

## Using the Application

### Web Interface

1. **Open the application**: Navigate to http://localhost:5173

2. **Q&A Management Page**:
   - View all Q&A pairs in a table
   - Search using keywords
   - Create new Q&A pairs (click "Create Q&A")
   - Edit existing pairs (click Edit icon)
   - Delete pairs (click Delete icon)
   - Navigate with pagination

3. **AI Chat Page**:
   - Type questions in natural language
   - AI agent searches the knowledge base
   - View streaming responses in real-time
   - See tool calls and reasoning
   - Conversation history is automatically saved

### API Usage

#### Create Q&A Pair

```bash
curl -X POST http://localhost:8080/api/qa-pairs \
  -H "Content-Type: application/json" \
  -d '{
    "question": "What is Docker?",
    "answer": "Docker is a containerization platform that packages applications and dependencies."
  }'
```

#### Search Q&A Pairs

```bash
curl "http://localhost:8080/api/qa-pairs?search=docker&limit=10"
```

#### Update Q&A Pair

```bash
curl -X PUT http://localhost:8080/api/qa-pairs/{id} \
  -H "Content-Type: application/json" \
  -d '{
    "question": "What is Docker?",
    "answer": "Updated answer here."
  }'
```

#### Delete Q&A Pair

```bash
curl -X DELETE http://localhost:8080/api/qa-pairs/{id}
```

#### Chat with AI Agent

```bash
# Create conversation
curl -X POST http://localhost:8000/chat/conversations \
  -H "Content-Type: application/json" \
  -d '{"title": "My Conversation"}'

# Send message (streaming response)
curl -X POST http://localhost:8000/chat/conversations/{id}/messages \
  -H "Content-Type: application/json" \
  -d '{"message": "What is Docker?"}'
```

## API Reference

### Backend API (Go) - Port 8080

#### Q&A Endpoints
- `GET /api/qa-pairs` - List Q&A pairs (with pagination and search)
- `GET /api/qa-pairs/:id` - Get specific Q&A pair
- `POST /api/qa-pairs` - Create new Q&A pair
- `PUT /api/qa-pairs/:id` - Update Q&A pair
- `DELETE /api/qa-pairs/:id` - Delete Q&A pair

#### Conversation Endpoints
- `POST /api/conversations` - Create conversation
- `GET /api/conversations` - List conversations
- `GET /api/conversations/:id` - Get conversation
- `DELETE /api/conversations/:id` - Delete conversation
- `POST /api/conversations/:id/messages` - Add message
- `GET /api/conversations/:id/messages` - Get messages

#### Tool Endpoints (for Python Agent)
- `POST /tools/search-qa` - Full-text search
- `POST /tools/semantic-search-qa` - Vector similarity search
- `POST /tools/get-qa-by-ids` - Get specific Q&A pairs by IDs
- `POST /tools/save-message` - Save conversation message

#### Health
- `GET /health` - Health check endpoint

### AI Agent API (Python) - Port 8000

- `POST /chat/conversations` - Create new conversation
- `POST /chat/conversations/{id}/messages` - Send message (streaming)
- `GET /chat/conversations/{id}/messages` - Get conversation messages
- `GET /health` - Health check

See component READMEs for detailed API documentation.

## Project Structure

```
.
├── backend/                    # Go backend service
│   ├── cmd/
│   │   ├── server/            # Main API server
│   │   └── batch-index/       # Batch indexing utility
│   ├── internal/
│   │   ├── api/               # HTTP handlers & middleware
│   │   ├── clients/           # External clients (Pinecone, Google)
│   │   ├── config/            # Configuration management
│   │   ├── models/            # Data models
│   │   ├── repository/        # Database layer
│   │   └── service/           # Business logic
│   ├── migrations/            # Database migrations
│   └── tests/                 # Integration tests
│
├── frontend/                   # React frontend
│   ├── src/
│   │   ├── features/
│   │   │   ├── qa/           # Q&A management feature
│   │   │   └── chat/         # AI chat feature
│   │   ├── components/       # Shared components
│   │   ├── app/              # Redux store
│   │   └── lib/              # Utilities
│   └── package.json
│
├── python-agent/              # Python AI agent
│   ├── agent/                # LangChain agent logic
│   │   ├── agent.py          # Main agent
│   │   ├── tools.py          # LangChain tools
│   │   ├── client.py         # Backend client
│   │   └── config.py         # Configuration
│   ├── api/                  # FastAPI routes
│   └── requirements.txt
│
├── docs/                      # Documentation
├── docker-compose.yml         # Docker orchestration
├── Makefile                   # Convenience commands
└── README.md                  # This file
```

## Development

### Backend Development

```bash
cd backend

# Run tests
go test ./...

# Integration tests
go test ./tests/...

# Build
go build -o bin/server cmd/server/main.go

# Run
./bin/server

# Format code
go fmt ./...

# Lint
golangci-lint run
```

### Frontend Development

```bash
cd frontend

# Install dependencies
npm install

# Run dev server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview

# Lint
npm run lint
```

### Python Agent Development

```bash
cd python-agent

# Activate virtual environment
source venv/bin/activate

# Install dependencies
pip install -r requirements.txt

# Run with hot reload
uvicorn main:app --reload --host 0.0.0.0 --port 8000

# Type checking
mypy agent/ api/
```

### Useful Make Commands

```bash
# Build backend
make build

# Run batch indexing
make batch-index

# Test run (dry run)
make batch-index-dry

# Clean build artifacts
make clean
```

## Troubleshooting

### Port Already in Use

```bash
# Check what's using the port
lsof -i :8080
lsof -i :5173
lsof -i :8000

# Kill the process
kill -9 <PID>
```

### Docker Issues

```bash
# Restart all services
docker-compose restart

# Rebuild containers
docker-compose up -d --build

# View logs
docker-compose logs -f backend
docker-compose logs -f python-agent

# Clean everything
docker-compose down -v
docker system prune -a
```

### Database Connection Errors

```bash
# Check if PostgreSQL is running
docker-compose ps postgres

# Connect to database
docker exec -it smart-discovery-db psql -U postgres -d smart_discovery

# Check tables
\dt

# View Q&A pairs
SELECT * FROM qa_pairs LIMIT 5;
```

### AI Agent Not Working

1. **Check Gemini API Key**: Ensure `GEMINI_API_KEY` is set correctly
2. **View logs**: `docker-compose logs -f python-agent`
3. **Test endpoint**: `curl http://localhost:8000/health`
4. **Check backend connection**: Agent needs backend at `http://backend:8080`

### Frontend Can't Connect to Backend

1. **CORS issues**: Check browser console (F12)
2. **Backend running**: `curl http://localhost:8080/health`
3. **Environment**: Check `VITE_API_BASE_URL` in `.env`
4. **Proxy**: Vite dev server proxies `/api/*` to backend

### Pinecone Local Issues

```bash
# Restart Pinecone Local
docker-compose restart pinecone-local

# Pull latest image
docker pull ghcr.io/pinecone-io/pinecone-local:latest

# Check if running
docker ps | grep pinecone
```

### Search Returns No Results

1. **Index Q&A pairs**: Run `make batch-index`
2. **Check embeddings**: Ensure Google API key is set
3. **Use mock client**: Set empty `GOOGLE_API_KEY` to use mock embeddings
4. **Check Pinecone**: Verify Pinecone Local is running

## Documentation

Detailed component documentation:

- **[Backend README](backend/README.md)** - Architecture, API docs, testing
- **[Frontend README](frontend/README.md)** - Components, state management, features
- **[Python Agent README](python-agent/README.md)** - LangChain, tools, agent flow

Additional documentation in `docs/`:

- [Startup Commands](docs/STARTUP_COMMANDS.md)
- [Frontend Setup](docs/FRONTEND_SETUP.md)
- [Docker Quick Start](docs/DOCKER_QUICKSTART.md)
- [Local Development](docs/LOCAL_DEVELOPMENT.md)
- [Pinecone Integration](docs/pinecone-integration.md)
- [Semantic Search Implementation](docs/SEMANTIC_SEARCH_IMPLEMENTATION.md)
- [Pagination Tests](docs/PAGINATION_TESTS_COMPLETE.md)

## License

This project is for demonstration purposes.

## Support

For issues or questions:
1. Check the [Troubleshooting](#troubleshooting) section
2. Review component-specific READMEs
3. Check documentation in `docs/` directory
4. Review Docker logs: `docker-compose logs -f`

---

**Built with ❤️ using Go, React, Python, LangChain, and Gemini 2.0**
