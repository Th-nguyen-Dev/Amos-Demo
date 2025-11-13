# Backend - Smart Company Discovery Assistant

Go-based REST API backend providing Q&A knowledge management, semantic search, and conversation management capabilities. Built with clean architecture principles and modern Go practices.

## 📋 Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Technology Stack](#technology-stack)
- [Project Structure](#project-structure)
- [Database Schema](#database-schema)
- [API Documentation](#api-documentation)
- [Services Architecture](#services-architecture)
- [Development Setup](#development-setup)
- [Testing](#testing)
- [Building and Running](#building-and-running)
- [Configuration Reference](#configuration-reference)
- [Integration Details](#integration-details)

## Overview

The backend service is a RESTful API built in Go that manages the core functionality of the Smart Company Discovery Assistant. It handles data persistence, search operations, embeddings generation, and provides endpoints for both the React frontend and Python AI agent.

### Key Responsibilities

- **Q&A Management**: CRUD operations for question-answer pairs
- **Vector Embeddings**: Generate embeddings using Google Gemini API
- **Semantic Search**: Vector similarity search via Pinecone
- **Full-Text Search**: PostgreSQL-based keyword search
- **Conversation Management**: Store and retrieve AI chat conversations
- **Automatic Indexing**: Q&A pairs are automatically vectorized on create/update
- **Tool Endpoints**: Specialized endpoints for the Python AI agent

## Architecture

The backend follows a clean, layered architecture:

```
┌─────────────────────────────────────────────────────────────┐
│                      HTTP Layer (Gin)                       │
│  • Routing          • Middleware         • Request Binding  │
└───────────────────────────┬─────────────────────────────────┘
                            │
┌───────────────────────────┴─────────────────────────────────┐
│                        Handlers Layer                        │
│  • qa_handler.go         • conversation_handler.go          │
│  • Request validation    • Response formatting              │
└───────────────────────────┬─────────────────────────────────┘
                            │
┌───────────────────────────┴─────────────────────────────────┐
│                       Services Layer                         │
│  • qa_service.go         • conversation_service.go          │
│  • embedding_service.go  • Business logic                   │
│  • Automatic indexing    • Search orchestration             │
└─────────────┬───────────────────────────────┬───────────────┘
              │                               │
    ┌─────────┴──────────┐         ┌─────────┴──────────┐
    │  Repository Layer  │         │   Clients Layer    │
    │  • qa_repository   │         │  • Pinecone        │
    │  • conv_repository │         │  • Google Gemini   │
    │  • SQL queries     │         │  • Mock clients    │
    └─────────┬──────────┘         └────────────────────┘
              │
    ┌─────────┴──────────┐
    │    PostgreSQL      │
    │  • qa_pairs        │
    │  • conversations   │
    │  • messages        │
    └────────────────────┘
```

### Architecture Principles

1. **Separation of Concerns**: Each layer has a single, well-defined responsibility
2. **Dependency Injection**: Dependencies are injected, making testing easier
3. **Interface-Based Design**: Services and repositories use interfaces
4. **Clean Error Handling**: Consistent error responses across the API
5. **Context Propagation**: Request context flows through all layers

## Technology Stack

### Core Framework
- **Go**: 1.24
- **Gin**: HTTP web framework with excellent performance
- **sqlx**: Enhanced SQL operations with struct scanning
- **pq**: PostgreSQL driver

### External Services
- **PostgreSQL 16**: Primary data store
- **Pinecone**: Vector database (Official Go SDK v4.1.4)
- **Google Gemini**: Embedding generation (google.golang.org/genai v1.33.0)

### Testing
- **testify**: Assertions and mocking
- **txdb**: Transactional test database
- **httptest**: HTTP handler testing

### Utilities
- **uuid**: UUID generation (google/uuid)
- **protobuf**: For Google API communication

## Project Structure

```
backend/
├── cmd/
│   ├── server/
│   │   └── main.go                 # Main API server entry point
│   └── batch-index/
│       └── main.go                 # Batch indexing utility
│
├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   │   ├── qa_handler.go       # Q&A CRUD handlers
│   │   │   └── conversation_handler.go  # Conversation handlers
│   │   └── middleware/
│   │       └── cors.go             # CORS middleware
│   │
│   ├── clients/
│   │   ├── google_embedding.go     # Google Gemini client
│   │   ├── pinecone_client.go      # Pinecone client (local + cloud)
│   │   └── pinecone_mock.go        # Mock Pinecone for testing
│   │
│   ├── config/
│   │   └── config.go               # Configuration loading
│   │
│   ├── models/
│   │   ├── config.go               # Configuration models
│   │   ├── qa.go                   # Q&A models
│   │   ├── conversation.go         # Conversation models
│   │   ├── pagination.go           # Pagination models
│   │   └── error.go                # Error models
│   │
│   ├── repository/
│   │   ├── qa_repository.go        # Q&A data access
│   │   └── conversation_repository.go  # Conversation data access
│   │
│   ├── service/
│   │   ├── qa_service.go           # Q&A business logic
│   │   ├── conversation_service.go # Conversation business logic
│   │   └── embedding_service.go    # Embedding generation
│   │
│   └── testutil/
│       └── db.go                   # Test database utilities
│
├── migrations/
│   └── 001_init_schema.sql         # Database schema
│
├── tests/
│   ├── qa/
│   │   ├── qa_integration_test.go  # Q&A integration tests
│   │   └── README.md
│   ├── conversation/
│   │   ├── conversation_integration_test.go
│   │   └── README.md
│   └── pagination/
│       ├── pagination_test.go
│       └── README.md
│
├── Dockerfile                       # Docker image definition
├── go.mod                          # Go module definition
└── go.sum                          # Dependency checksums
```

### Directory Explanations

- **cmd/**: Application entry points
  - `server/`: Main API server
  - `batch-index/`: Utility to re-index all Q&A pairs

- **internal/api/**: HTTP layer
  - `handlers/`: Request handlers and response formatting
  - `middleware/`: HTTP middleware (CORS, auth, logging)

- **internal/clients/**: External service clients
  - Pinecone vector database client
  - Google Gemini embedding client
  - Mock implementations for testing

- **internal/config/**: Configuration management
  - Environment variable loading
  - Configuration validation

- **internal/models/**: Data models
  - Request/response DTOs
  - Domain models
  - Database models

- **internal/repository/**: Data access layer
  - SQL queries
  - Database operations
  - Cursor-based pagination

- **internal/service/**: Business logic layer
  - Orchestrates repositories and clients
  - Implements business rules
  - Handles automatic indexing

- **migrations/**: Database migrations
  - SQL schema definitions
  - Database initialization

- **tests/**: Integration tests
  - Feature-based test suites
  - Uses real PostgreSQL (txdb for transactions)

## Database Schema

### qa_pairs Table

```sql
CREATE TABLE qa_pairs (
    id UUID PRIMARY KEY,
    question TEXT NOT NULL,
    answer TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Full-text search index
CREATE INDEX idx_qa_fts ON qa_pairs 
    USING gin(to_tsvector('english', question || ' ' || answer));

-- Pagination indexes
CREATE INDEX idx_qa_id_desc ON qa_pairs(id DESC);
CREATE INDEX idx_qa_covering ON qa_pairs(created_at DESC) 
    INCLUDE (question, answer);
```

### conversations Table

```sql
CREATE TABLE conversations (
    id UUID PRIMARY KEY,
    title TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_conv_id_desc ON conversations(id DESC);
```

### messages Table

```sql
CREATE TABLE messages (
    id UUID PRIMARY KEY,
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    
    -- Extracted fields (OpenAI standard)
    role TEXT NOT NULL CHECK (role IN ('user', 'assistant', 'tool', 'system')),
    content TEXT,
    tool_call_id TEXT,
    
    -- Complete message in OpenAI format
    raw_message JSONB NOT NULL,
    
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Optimized indexes for conversation queries
CREATE INDEX idx_messages_conv_time ON messages(conversation_id, created_at DESC, id);
CREATE INDEX idx_messages_role ON messages(role);
CREATE INDEX idx_messages_content ON messages 
    USING gin(to_tsvector('english', content));
CREATE INDEX idx_messages_raw ON messages USING gin(raw_message);

-- Partial indexes for common queries
CREATE INDEX idx_messages_user ON messages(conversation_id, created_at DESC) 
    WHERE role = 'user';
CREATE INDEX idx_messages_assistant ON messages(conversation_id, created_at DESC) 
    WHERE role = 'assistant';
```

### Indexes Strategy

1. **Full-text search**: GIN indexes on text columns for fast keyword search
2. **Cursor pagination**: Descending ID indexes for efficient pagination
3. **Covering indexes**: Include frequently accessed columns
4. **Partial indexes**: Optimize specific query patterns
5. **JSONB indexes**: Enable efficient querying of OpenAI message format

## API Documentation

### Q&A Endpoints

#### List Q&A Pairs

```http
GET /api/qa-pairs?limit=20&cursor=&search=docker
```

**Query Parameters:**
- `limit` (int, optional): Number of results (default: 20, max: 100)
- `cursor` (string, optional): Pagination cursor from previous response
- `search` (string, optional): Full-text search query

**Response:**
```json
{
  "data": [
    {
      "id": "uuid",
      "question": "What is Docker?",
      "answer": "Docker is a containerization platform...",
      "created_at": "2025-11-03T10:00:00Z",
      "updated_at": "2025-11-03T10:00:00Z"
    }
  ],
  "pagination": {
    "next_cursor": "encoded-cursor",
    "has_more": true,
    "limit": 20
  }
}
```

#### Get Q&A Pair

```http
GET /api/qa-pairs/:id
```

**Response:**
```json
{
  "qa_pair": {
    "id": "uuid",
    "question": "What is Docker?",
    "answer": "Docker is a containerization platform...",
    "created_at": "2025-11-03T10:00:00Z",
    "updated_at": "2025-11-03T10:00:00Z"
  }
}
```

#### Create Q&A Pair

```http
POST /api/qa-pairs
Content-Type: application/json

{
  "question": "What is Docker?",
  "answer": "Docker is a containerization platform..."
}
```

**Response:** `201 Created`
```json
{
  "qa_pair": {
    "id": "uuid",
    "question": "What is Docker?",
    "answer": "Docker is a containerization platform...",
    "created_at": "2025-11-03T10:00:00Z",
    "updated_at": "2025-11-03T10:00:00Z"
  }
}
```

**Side Effects:**
- Q&A pair is saved to PostgreSQL
- Embedding is generated via Google Gemini
- Vector is indexed in Pinecone (automatic)

#### Update Q&A Pair

```http
PUT /api/qa-pairs/:id
Content-Type: application/json

{
  "question": "What is Docker?",
  "answer": "Updated answer..."
}
```

**Response:** `200 OK`

**Side Effects:**
- Q&A pair is updated in PostgreSQL
- New embedding is generated
- Vector is re-indexed in Pinecone (automatic)

#### Delete Q&A Pair

```http
DELETE /api/qa-pairs/:id
```

**Response:** `200 OK`
```json
{
  "message": "QA pair deleted successfully"
}
```

**Side Effects:**
- Q&A pair is deleted from PostgreSQL
- Vector is removed from Pinecone (automatic)

### Conversation Endpoints

#### Create Conversation

```http
POST /api/conversations
Content-Type: application/json

{
  "title": "Docker Questions"
}
```

#### List Conversations

```http
GET /api/conversations?limit=20&cursor=
```

#### Get Conversation

```http
GET /api/conversations/:id
```

#### Delete Conversation

```http
DELETE /api/conversations/:id
```

Note: Deletes all messages via CASCADE constraint.

#### Add Message to Conversation

```http
POST /api/conversations/:id/messages
Content-Type: application/json

{
  "role": "user",
  "content": "What is Docker?",
  "raw_message": {
    "role": "user",
    "content": "What is Docker?"
  }
}
```

#### Get Conversation Messages

```http
GET /api/conversations/:id/messages?limit=50&cursor=
```

### Tool Endpoints (for Python AI Agent)

#### Search Q&A (Full-Text)

```http
POST /tools/search-qa
Content-Type: application/json

{
  "query": "docker container",
  "limit": 5
}
```

**Response:**
```json
{
  "qa_pairs": [...],
  "count": 5
}
```

#### Semantic Search Q&A (Vector)

```http
POST /tools/semantic-search-qa
Content-Type: application/json

{
  "query": "containerization platforms",
  "top_k": 5
}
```

**Response:**
```json
{
  "results": [
    {
      "qa_pair": {...},
      "score": 0.89
    }
  ],
  "count": 5
}
```

#### Get Q&A by IDs

```http
POST /tools/get-qa-by-ids
Content-Type: application/json

{
  "ids": ["uuid1", "uuid2"]
}
```

#### Save Message (for Agent)

```http
POST /tools/save-message
Content-Type: application/json

{
  "conversation_id": "uuid",
  "role": "assistant",
  "content": "Answer...",
  "raw_message": {...}
}
```

### Health Check

```http
GET /health
```

**Response:**
```json
{
  "status": "healthy",
  "database": "connected"
}
```

## Services Architecture

### QA Service (`qa_service.go`)

Handles Q&A business logic with automatic vector indexing.

**Key Methods:**
- `CreateQA(ctx, req)` - Create and auto-index Q&A pair
- `GetQA(ctx, id)` - Retrieve single Q&A pair
- `ListQA(ctx, params)` - List with cursor pagination
- `SearchQA(ctx, query, params)` - Full-text search
- `UpdateQA(ctx, id, req)` - Update and re-index
- `DeleteQA(ctx, id)` - Delete and remove from vector DB
- `SearchSimilarByText(ctx, text, topK)` - Semantic search
- `GetQAByIDs(ctx, ids)` - Batch retrieval

**Automatic Indexing Flow:**

```go
// Create Q&A
qa := CreateQA(ctx, request)
  ↓
Save to PostgreSQL (repository)
  ↓
Generate embedding (embeddingService)
  ↓
Index vector in Pinecone (pineconeClient)
  ↓
Return Q&A pair
```

### Conversation Service (`conversation_service.go`)

Manages conversations and messages in OpenAI format.

**Key Methods:**
- `CreateConversation(ctx, title)` - New conversation
- `GetConversation(ctx, id)` - Retrieve conversation
- `ListConversations(ctx, params)` - List with pagination
- `DeleteConversation(ctx, id)` - Delete conversation (cascades to messages)
- `AddMessage(ctx, req)` - Add message to conversation
- `GetMessages(ctx, conversationID, params)` - List messages with pagination

**Message Storage:**
- Stores complete OpenAI-format message in `raw_message` JSONB
- Extracts common fields (role, content, tool_call_id) for efficient querying
- Supports all OpenAI roles: user, assistant, tool, system

### Embedding Service (`embedding_service.go`)

Orchestrates embedding generation and vector indexing.

**Key Methods:**
- `GenerateAndIndexQA(ctx, qaPair)` - Generate embedding and index vector
- `UpdateQAIndex(ctx, qaPair)` - Update existing vector
- `DeleteQAIndex(ctx, id)` - Remove vector from index

**Process:**
```go
GenerateAndIndexQA(qaPair)
  ↓
Combine question + answer into text
  ↓
Generate embedding via Google Gemini (768 dimensions)
  ↓
Upsert vector to Pinecone with metadata
  ↓
Handle errors gracefully
```

## Development Setup

### Prerequisites

- Go 1.24+
- PostgreSQL 16
- Docker and Docker Compose (recommended)
- Make (optional)

### Local Development (with Docker for services)

```bash
# Start PostgreSQL and Pinecone Local
docker-compose up -d postgres pinecone-local

# Set environment variables
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=smart_discovery
export DB_USER=postgres
export DB_PASSWORD=postgres
export PINECONE_API_KEY=pclocal
export PINECONE_HOST=http://localhost:5081
export PINECONE_INDEX_NAME=qa-index

# Optional: Set Google API key for real embeddings
export GOOGLE_API_KEY=your-key
export GOOGLE_PROJECT_ID=your-project

# Run migrations (first time only)
psql -h localhost -U postgres -d smart_discovery -f migrations/001_init_schema.sql

# Install dependencies
go mod download

# Run the server
go run cmd/server/main.go
```

Server starts at: **http://localhost:8080**

### Local Development (Full)

If you want to run PostgreSQL locally without Docker:

```bash
# Install PostgreSQL 16
# Create database
createdb -U postgres smart_discovery

# Run migrations
psql -U postgres -d smart_discovery -f migrations/001_init_schema.sql

# Set environment and run
export DB_HOST=localhost
# ... other env vars ...
go run cmd/server/main.go
```

### Development with Hot Reload

Use `air` for hot reload during development:

```bash
# Install air
go install github.com/cosmtrek/air@latest

# Run with hot reload
air
```

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific test suite
go test -v ./tests/qa/
go test -v ./tests/conversation/
go test -v ./tests/pagination/

# Run with coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Integration Tests

Integration tests use a real PostgreSQL database with txdb for transaction isolation.

**Test Setup:**
```go
// Each test runs in a transaction that's rolled back after the test
db := testutil.SetupTestDB(t)
defer db.Close()

// Tests have a clean database state
// No need to manually clean up
```

**Test Structure:**
```
tests/
├── qa/
│   └── qa_integration_test.go         # Q&A CRUD tests
├── conversation/
│   └── conversation_integration_test.go # Conversation tests
└── pagination/
    └── pagination_test.go             # Pagination tests
```

### Test Database

Integration tests require PostgreSQL:

```bash
# Using Docker
docker-compose up -d postgres

# Or use test database
createdb -U postgres smart_discovery_test
psql -U postgres -d smart_discovery_test -f migrations/001_init_schema.sql
```

### Mock Clients

For testing without external dependencies:

```go
// Use mock Pinecone client
pineconeClient := clients.NewMockPineconeClient()

// Use mock embedding client
embeddingClient := clients.NewMockEmbeddingClient(768)
```

## Building and Running

### Build Binary

```bash
# Build server
go build -o bin/server cmd/server/main.go

# Build batch indexer
go build -o bin/batch-index cmd/batch-index/main.go

# Run
./bin/server
```

### Using Make

```bash
# Build
make build

# Run batch indexing
make batch-index

# Clean
make clean
```

### Docker Build

```bash
# Build image
docker build -t smart-discovery-backend .

# Run container
docker run -p 8080:8080 \
  -e DB_HOST=host.docker.internal \
  -e DB_NAME=smart_discovery \
  smart-discovery-backend
```

### Docker Compose

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f backend

# Restart backend
docker-compose restart backend
```

## Configuration Reference

### Environment Variables

#### Database Configuration

```bash
DB_HOST=localhost                 # Database host
DB_PORT=5432                      # Database port
DB_USER=postgres                  # Database user
DB_PASSWORD=postgres              # Database password
DB_NAME=smart_discovery           # Database name
DB_SSLMODE=disable               # SSL mode (disable, require, verify-ca, verify-full)
DB_MAX_OPEN_CONNS=25             # Max open connections
DB_MAX_IDLE_CONNS=5              # Max idle connections
```

#### Server Configuration

```bash
SERVER_PORT=8080                  # API server port
SERVER_HOST=0.0.0.0              # Server host (0.0.0.0 for Docker)
SERVER_ENVIRONMENT=development    # Environment (development, production)
```

#### Pinecone Configuration (Local Mode)

```bash
PINECONE_API_KEY=pclocal         # Use "pclocal" for local mode
PINECONE_HOST=http://localhost:5081   # Pinecone Local URL
PINECONE_INDEX_NAME=qa-index     # Index name
PINECONE_NAMESPACE=default       # Namespace (optional)
```

#### Pinecone Configuration (Cloud Mode)

```bash
PINECONE_API_KEY=your-real-key   # Your Pinecone API key
PINECONE_HOST=                    # Empty for cloud mode
PINECONE_ENVIRONMENT=us-west1-gcp # Your Pinecone environment
PINECONE_INDEX_NAME=qa-index     # Index name
PINECONE_NAMESPACE=default       # Namespace (optional)
```

#### Google Embedding Configuration

```bash
GOOGLE_API_KEY=your-api-key      # Google Gemini API key
GOOGLE_PROJECT_ID=your-project   # Google Cloud project ID
GOOGLE_LOCATION=us-central1      # Google Cloud region
GOOGLE_EMBEDDING_MODEL=gemini-embedding-001  # Embedding model
```

**Note**: If Google credentials are not provided, the system uses a mock embedding client (random 768-dimensional vectors).

### Configuration Loading

Configuration is loaded from environment variables via `config.LoadConfig()`:

```go
cfg, err := config.LoadConfig()
if err != nil {
    // Falls back to defaults
}
```

## Integration Details

### Pinecone Integration

**Local Mode (Testing)**:
- Uses Pinecone Local Docker container
- Data is in-memory (not persistent)
- No API key required (use "pclocal")
- Perfect for development and testing

```bash
PINECONE_API_KEY=pclocal
PINECONE_HOST=http://localhost:5081
```

**Cloud Mode (Production)**:
- Uses Pinecone Cloud
- Data is persistent
- Requires API key and account
- Pay-as-you-go pricing

```bash
PINECONE_API_KEY=your-real-key
PINECONE_HOST=
PINECONE_ENVIRONMENT=us-west1-gcp
```

**Switching Modes**:
```go
// System automatically detects mode based on PINECONE_HOST
if cfg.Pinecone.Host != "" {
    // Local mode
} else {
    // Cloud mode
}
```

### Google Gemini Embeddings

Uses the official `google.golang.org/genai` SDK for embeddings:

```go
client, err := clients.NewGoogleEmbeddingClient(ctx, clients.GoogleEmbeddingConfig{
    APIKey:    cfg.GoogleEmbedding.APIKey,
    ProjectID: cfg.GoogleEmbedding.ProjectID,
    Location:  cfg.GoogleEmbedding.Location,
    Model:     "gemini-embedding-001",
})
```

**Features**:
- 768-dimensional embeddings
- Text normalization
- Batch processing support
- Graceful fallback to mock client

### Database Connection Pool

Optimized connection pool settings:

```go
db.SetMaxOpenConns(25)        // Max simultaneous connections
db.SetMaxIdleConns(5)         // Idle connections in pool
db.SetConnMaxLifetime(5 * time.Minute)  // Connection lifetime
```

### CORS Configuration

CORS is enabled for frontend access:

```go
// Allows all origins in development
router.Use(middleware.CORS())

// Configure for production:
// - Specific origins
// - Credentials support
// - Method restrictions
```

### Graceful Shutdown

Server supports graceful shutdown:

```go
// Waits for existing requests to complete
// Timeout: 5 seconds
// Triggered by: SIGINT, SIGTERM
```

## Performance Considerations

### Indexing Strategy

1. **Cursor Pagination**: Efficient for large datasets
2. **Covering Indexes**: Reduce disk I/O
3. **Partial Indexes**: Optimize specific queries
4. **GIN Indexes**: Fast full-text search

### Query Optimization

- Use prepared statements (sqlx handles this)
- Limit result sets with cursor pagination
- Batch operations where possible
- Connection pooling

### Embedding Performance

- Embeddings are generated asynchronously
- Batch indexing available for bulk operations
- Mock client for testing (no API calls)

### Caching (Future Enhancement)

Consider adding:
- Redis for frequently accessed Q&A pairs
- Embedding cache to avoid re-generation
- Search result caching

## Monitoring and Logging

### Logging

```go
log.Printf("✓ Successfully connected to PostgreSQL")
log.Printf("⚠ Warning: Using mock client")
log.Printf("✗ Error: %v", err)
```

### Health Check

```bash
curl http://localhost:8080/health
```

Monitor:
- Database connectivity
- API responsiveness
- Service availability

### Metrics (Future Enhancement)

Consider adding:
- Request duration
- Error rates
- Database query performance
- Embedding generation time

## API Versioning

Current: **v1** (implicit)

Future: Add `/v1/` prefix for versioning:
```
/api/v1/qa-pairs
/api/v2/qa-pairs
```

## Security Considerations

### Current Implementation

- CORS enabled for frontend
- Input validation on all endpoints
- SQL injection prevention (parameterized queries)
- UUID for identifiers (not sequential)

### Production Recommendations

1. **Authentication**: Add JWT or API key authentication
2. **Authorization**: Role-based access control
3. **Rate Limiting**: Prevent abuse
4. **HTTPS**: TLS encryption in production
5. **Input Sanitization**: Additional validation
6. **Secrets Management**: Use secret managers (not env vars)

## Troubleshooting

### Database Connection Issues

```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Test connection
psql -h localhost -U postgres -d smart_discovery

# View logs
docker-compose logs postgres
```

### Pinecone Issues

```bash
# Check Pinecone Local
curl http://localhost:5081/

# Restart
docker-compose restart pinecone-local

# Re-index all Q&A pairs
make batch-index
```

### Embedding Errors

```bash
# Check Google API key
echo $GOOGLE_API_KEY

# Use mock client (remove API key)
unset GOOGLE_API_KEY
```

### Build Errors

```bash
# Clean and rebuild
go clean
go mod tidy
go build -o bin/server cmd/server/main.go
```

## Related Documentation

- [Main README](../README.md) - Full application setup
- [Frontend README](../frontend/README.md) - React frontend
- [Python Agent README](../python-agent/README.md) - AI agent
- [Pinecone Integration](../docs/pinecone-integration.md)
- [Database Migrations](../docs/MIGRATION_SUMMARY.md)

---

**Built with clean architecture principles and modern Go practices** 🚀






