# Docker Quick Start Guide

This guide shows how to run the complete stack including the Python LangChain AI Agent.

## Services Overview

- **postgres**: PostgreSQL database (port 5432)
- **pinecone-local**: Local Pinecone vector database for testing (ports 5081-6000)
- **backend**: Go API server (port 8080)
- **python-agent**: Python LangChain AI agent with Gemini (port 8000)

## Prerequisites

1. Docker and Docker Compose installed
2. Google Gemini API key

## Quick Start

### 1. Set Environment Variables

Create a `.env` file in the project root:

```bash
# Create .env file with your Gemini API key
cat > .env << EOF
GEMINI_API_KEY=your_actual_gemini_api_key_here
EOF
```

### 2. Start All Services

```bash
docker-compose up -d
```

This will start all four services. First run will take a few minutes to build images.

### 3. Verify Services

```bash
# Check all services are running
docker-compose ps

# Should show:
# - postgres (healthy)
# - pinecone-local (healthy)
# - backend (running)
# - python-agent (running)
```

### 4. Test the Services

```bash
# Test Go backend
curl http://localhost:8080/health

# Test Python agent
curl http://localhost:8000/health

# View API documentation
open http://localhost:8000/docs
```

## Using the Python AI Agent

### Create a Conversation

```bash
curl -X POST http://localhost:8000/chat/conversations \
  -H "Content-Type: application/json" \
  -d '{"title": "My First Chat"}'
```

This returns a conversation ID. Use it to send messages.

### Send a Message

```bash
curl -X POST http://localhost:8000/chat/conversations/{conversation_id}/messages \
  -H "Content-Type: application/json" \
  -d '{"message": "What is your refund policy?"}' \
  --no-buffer
```

The agent will:
1. Load conversation history
2. Use LangChain tools to search the knowledge base
3. Stream the response back
4. Save all messages to PostgreSQL

### View Conversation History

```bash
curl http://localhost:8000/chat/conversations/{conversation_id}/messages
```

## Service Management

### View Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f python-agent
docker-compose logs -f backend
```

### Restart Services

```bash
# Restart all
docker-compose restart

# Restart specific service
docker-compose restart python-agent
```

### Rebuild After Code Changes

```bash
# Rebuild Go backend
docker-compose up -d --build backend

# Rebuild Python agent
docker-compose up -d --build python-agent
```

### Stop Services

```bash
# Stop but keep data
docker-compose stop

# Stop and remove containers (keeps data volumes)
docker-compose down

# Stop and remove everything including data
docker-compose down -v
```

## Architecture

```
┌─────────────────┐
│  React Client   │
│ (Port 3000)     │
└────────┬────────┘
         │
         ├──────────────────┐
         │                  │
         ▼                  ▼
┌─────────────────┐  ┌──────────────────┐
│  Python Agent   │  │   Go Backend     │
│  (Port 8000)    │◄─┤   (Port 8080)    │
│                 │  │                  │
│  - Gemini 2.5   │  │  - REST API      │
│  - LangChain    │  │  - QA Management │
│  - Tools        │  │                  │
└─────────────────┘  └────────┬─────────┘
                              │
                    ┌─────────┴─────────┐
                    │                   │
                    ▼                   ▼
            ┌──────────────┐    ┌──────────────┐
            │  PostgreSQL  │    │  Pinecone    │
            │  (Port 5432) │    │  (Port 5081) │
            └──────────────┘    └──────────────┘
```

## Communication Flow

1. **User → Python Agent**: React client sends message to FastAPI
2. **Python Agent → Go Backend**: Agent calls REST APIs for tools
   - `POST /tools/search-qa` - Search knowledge base
   - `POST /tools/save-message` - Save messages
3. **Go Backend → PostgreSQL**: Store/retrieve data
4. **Go Backend → Pinecone**: Vector search (when configured)
5. **Python Agent → User**: Stream response back

## Environment Variables

### Python Agent (set in docker-compose.yml)

```yaml
GEMINI_API_KEY: ${GEMINI_API_KEY}        # From root .env
GEMINI_MODEL: gemini-2.0-flash-exp       # Model to use
BACKEND_URL: http://backend:8080         # Docker service name
USE_PINECONE: false                      # Not yet configured
API_HOST: 0.0.0.0
API_PORT: 8000
CORS_ORIGINS: ["http://localhost:3000"]
```

### Go Backend (set in docker-compose.yml)

```yaml
DB_HOST: postgres                        # Docker service name
BACKEND_PORT: 8080
PINECONE_HOST: http://pinecone-local:5081
```

## Networking

All services run on the default Docker Compose network. Services communicate using service names:
- Python agent calls `http://backend:8080`
- Go backend calls `postgres:5432`
- Go backend calls `http://pinecone-local:5081`

## Data Persistence

- **PostgreSQL data**: Persisted in `postgres_data` Docker volume
- **Pinecone Local**: NOT persisted (testing only, data lost on restart)

## Troubleshooting

### Python agent can't connect to backend

```bash
# Check backend is running
docker-compose ps backend

# Check backend logs
docker-compose logs backend

# Test backend from inside python-agent container
docker-compose exec python-agent curl http://backend:8080/health
```

### Backend can't connect to database

```bash
# Check postgres is healthy
docker-compose ps postgres

# Check database logs
docker-compose logs postgres

# Wait for database to be ready
docker-compose up -d --wait
```

### "Invalid API key" errors

```bash
# Check environment variable is set
docker-compose config | grep GEMINI_API_KEY

# If empty, add to root .env file
echo "GEMINI_API_KEY=your_key_here" >> .env

# Restart services
docker-compose restart python-agent
```

### Port already in use

```bash
# Check what's using the port
lsof -i :8000  # or :8080, :5432

# Change port in docker-compose.yml
# ports:
#   - "8001:8000"  # Map to different host port
```

## Next Steps

1. **Add QA pairs** to the knowledge base via Go backend API
2. **Configure Pinecone Cloud** for production vector search
3. **Build React frontend** to interact with the Python agent
4. **Deploy to production** using the same Docker setup

## Resources

- Python Agent API Docs: http://localhost:8000/docs
- Go Backend: http://localhost:8080/api/qa-pairs
- Database: `postgresql://postgres:postgres@localhost:5432/smart_discovery`

