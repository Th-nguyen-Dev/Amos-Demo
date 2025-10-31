# Go Backend Architecture Design Document

**Version:** 1.0  
**Date:** October 30, 2025  
**Project:** Smart Company Discovery Assistant

---

## Table of Contents

1. [Introduction](#introduction)
2. [System Architecture](#system-architecture)
3. [Project Structure](#project-structure)
4. [Database Design](#database-design)
5. [Message Format Standard](#message-format-standard)
6. [Data Models](#data-models)
7. [Repository Layer](#repository-layer)
8. [Service Layer](#service-layer)
9. [API Specification](#api-specification)
10. [Full-Text Search](#full-text-search)
11. [Cursor Pagination](#cursor-pagination)
12. [UUIDv7 Implementation](#uuidv7-implementation)
13. [Pinecone Integration](#pinecone-integration)
14. [Configuration](#configuration)
15. [Error Handling](#error-handling)
16. [Middleware](#middleware)
17. [Dependencies](#dependencies)

---

## 1. Introduction

### 1.1 System Overview

The Go backend serves as the central data management layer for the Smart Company Discovery Assistant. It provides RESTful APIs for both the React UI and the Python AI service, managing Q&A knowledge base, conversation history, and integration with Pinecone vector database.

### 1.2 Design Principles

- **Separation of Concerns:** Clean layered architecture (Handler → Service → Repository)
- **Modern Identifiers:** UUIDv7 for time-ordered, distributed-ready IDs
- **Efficient Pagination:** Cursor-based pagination for scalable data access
- **Intelligent Search:** PostgreSQL full-text search with stemming and ranking
- **Provider-Agnostic:** OpenAI message format as universal standard
- **Type Safety:** Strong typing with sqlx and struct validation

### 1.3 Technology Stack

| Component | Technology | Version | Purpose |
|-----------|-----------|---------|---------|
| HTTP Framework | Gin | v1.9+ | Routing and middleware |
| Database Driver | sqlx | v1.3+ | Type-safe PostgreSQL access |
| UUID Generation | github.com/google/uuid | v1.4+ | UUIDv7 generation |
| Validation | go-playground/validator | v10 | Struct validation |
| Configuration | viper | v1.17+ | Environment management |
| Logging | zerolog | v1.31+ | Structured logging |
| Vector Database | Pinecone Go Client | latest | Vector operations |
| Database | PostgreSQL | 12+ | Primary data store |

---

## 2. System Architecture

### 2.1 High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         React UI                                │
│  - Q&A Management Interface                                     │
│  - Conversation Interface                                       │
└─────────────┬───────────────────────────────────┬───────────────┘
              │                                   │
              │ HTTP/REST                         │ HTTP/SSE
              │                                   │
              ▼                                   ▼
┌─────────────────────────┐         ┌────────────────────────────┐
│   Go Backend Server     │         │  Python AI Service         │
│                         │         │  - LLM Processing          │
│  ┌──────────────────┐   │         │  - Embeddings Generation   │
│  │  API Handlers    │   │         │  - Agent Orchestration     │
│  │  (Gin)           │   │         └──────────┬─────────────────┘
│  └────────┬─────────┘   │                    │
│           │             │                    │ HTTP/REST (Tools)
│  ┌────────▼─────────┐   │                    │
│  │  Service Layer   │   │◄───────────────────┘
│  │  - Business Logic│   │
│  └────────┬─────────┘   │
│           │             │
│  ┌────────▼─────────┐   │
│  │ Repository Layer │   │
│  │  - Data Access   │   │
│  └────────┬─────────┘   │
│           │             │
│  ┌────────▼─────────┐   │
│  │  Clients         │   │
│  │  - Pinecone      │   │
│  └────────┬─────────┘   │
└───────────┼─────────────┘
            │
    ┌───────┴────────┐
    │                │
    ▼                ▼
┌─────────┐    ┌──────────┐
│PostgreSQL│    │ Pinecone │
│         │    │  (Vector)│
└─────────┘    └──────────┘
```

### 2.2 Component Responsibilities

**Go Backend:**
- Serve React UI with CRUD APIs for Q&A management
- Provide conversation history management
- Expose tool APIs for Python AI service
- Manage PostgreSQL and Pinecone operations
- Handle authentication and authorization (future)

**React UI:**
- Q&A management interface (create, edit, delete, search)
- Conversation browsing and display
- Direct connection to Python for AI Q&A (SSE streaming)

**Python AI Service:**
- Process natural language questions
- Generate embeddings for vector search
- Call Go tool APIs for data operations
- Orchestrate LLM responses with function calling
- Stream responses via SSE to React

**PostgreSQL:**
- Store Q&A knowledge base
- Store conversation history with OpenAI format messages
- Full-text search indexing

**Pinecone:**
- Store and query vector embeddings
- Similarity search for semantic Q&A matching

---

## 3. Project Structure

```
smart-company-discovery/
├── cmd/
│   └── server/
│       └── main.go                    # Application entry point
│
├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   │   ├── qa_handler.go         # Q&A CRUD handlers
│   │   │   ├── conversation_handler.go # Conversation handlers
│   │   │   ├── tool_handler.go       # Tool API for Python
│   │   │   └── health_handler.go     # Health check
│   │   └── middleware/
│   │       ├── cors.go               # CORS configuration
│   │       ├── logger.go             # Request logging
│   │       ├── recovery.go           # Panic recovery
│   │       └── validator.go          # Request validation
│   │
│   ├── service/
│   │   ├── qa_service.go             # Q&A business logic
│   │   ├── conversation_service.go   # Conversation logic
│   │   └── vector_service.go         # Pinecone operations
│   │
│   ├── repository/
│   │   ├── qa_repository.go          # Q&A data access
│   │   ├── conversation_repository.go # Conversation data access
│   │   └── postgres.go               # Database connection
│   │
│   ├── models/
│   │   ├── qa.go                     # Q&A models and DTOs
│   │   ├── conversation.go           # Conversation models
│   │   ├── pagination.go             # Pagination models
│   │   └── error.go                  # Error models
│   │
│   ├── config/
│   │   └── config.go                 # Configuration struct and loading
│   │
│   └── clients/
│       └── pinecone.go               # Pinecone client wrapper
│
├── pkg/
│   ├── utils/
│   │   └── uuid.go                   # UUIDv7 utilities
│   └── validator/
│       └── custom.go                 # Custom validation rules
│
├── migrations/
│   ├── 001_init_schema.sql          # Initial database schema
│   ├── 002_add_indexes.sql          # Performance indexes
│   └── 003_add_fts.sql              # Full-text search setup
│
├── docs/
│   ├── go-backend-design.md          # This document
│   └── api.md                        # API documentation
│
├── .env.example                       # Environment variable template
├── go.mod                            # Go module definition
├── go.sum                            # Dependency checksums
└── README.md                         # Project documentation
```

### 3.1 Directory Descriptions

**`/cmd/server`**
- Application entry point
- Initializes configuration, database, services
- Sets up HTTP server with Gin
- Graceful shutdown handling

**`/internal/api/handlers`**
- HTTP request handlers (controllers)
- Request parsing and validation
- Response formatting
- Minimal business logic

**`/internal/api/middleware`**
- Cross-cutting concerns
- CORS, logging, recovery, authentication
- Request/response transformation

**`/internal/service`**
- Business logic layer
- Orchestrates repository calls
- Transaction management
- Complex operations

**`/internal/repository`**
- Data access layer
- Database queries (sqlx)
- CRUD operations
- No business logic

**`/internal/models`**
- Domain entities
- DTOs (Data Transfer Objects)
- Request/Response structures
- Validation tags

**`/internal/config`**
- Configuration loading
- Environment variable parsing
- Configuration validation

**`/internal/clients`**
- External service clients
- Pinecone wrapper
- HTTP clients for external APIs

**`/pkg`**
- Reusable packages
- Can be imported by external projects
- Utilities and helpers

**`/migrations`**
- SQL migration scripts
- Versioned database changes
- Up/down migrations

---

## 4. Database Design

### 4.1 Schema Overview

The database uses **UUIDv7** for all primary keys, providing time-ordered identifiers that enable efficient cursor pagination and distributed system compatibility.

### 4.2 Q&A Knowledge Base

```sql
CREATE TABLE qa_pairs (
    id UUID PRIMARY KEY,
    question TEXT NOT NULL,
    answer TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_qa_id_desc ON qa_pairs(id DESC);
CREATE INDEX idx_qa_fts ON qa_pairs 
    USING gin(to_tsvector('english', question || ' ' || answer));

-- Update timestamp trigger
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_qa_pairs_updated_at BEFORE UPDATE ON qa_pairs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

**Design Notes:**
- `id`: UUIDv7 generated in Go application
- `question`: Full-text indexed for search
- `answer`: Full-text indexed for search
- `created_at`: Auto-set, used for auditing
- `updated_at`: Auto-updated on modifications

### 4.3 Conversations

```sql
CREATE TABLE conversations (
    id UUID PRIMARY KEY,
    title TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_conv_id_desc ON conversations(id DESC);

-- Update timestamp trigger
CREATE TRIGGER update_conversations_updated_at BEFORE UPDATE ON conversations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

**Design Notes:**
- `id`: UUIDv7 for chronological sorting
- `title`: Optional, can be auto-generated from first message
- Timestamp indexes for sorting by recency

### 4.4 Messages (OpenAI Format)

```sql
CREATE TABLE messages (
    id UUID PRIMARY KEY,
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    
    -- Extracted fields for querying (OpenAI standard)
    role TEXT NOT NULL CHECK (role IN ('user', 'assistant', 'tool', 'system')),
    content TEXT,
    tool_call_id TEXT,
    
    -- Complete message in OpenAI format
    raw_message JSONB NOT NULL,
    
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_messages_conv ON messages(conversation_id, id ASC);
CREATE INDEX idx_messages_role ON messages(role);
CREATE INDEX idx_messages_content ON messages 
    USING gin(to_tsvector('english', content));
CREATE INDEX idx_messages_raw ON messages USING gin(raw_message);
```

**Design Notes:**
- `role`: Enum for message types (user, assistant, tool, system)
- `content`: Main text, NULL for assistant messages with tool_calls
- `tool_call_id`: Links tool responses to specific tool calls
- `raw_message`: Complete OpenAI format for LLM consumption
- Composite index `(conversation_id, id)` for efficient message loading

### 4.5 Entity Relationship Diagram

```
┌─────────────────┐
│    qa_pairs     │
│─────────────────│
│ id (UUID) PK    │
│ question        │
│ answer          │
│ created_at      │
│ updated_at      │
└─────────────────┘


┌─────────────────┐         ┌─────────────────┐
│ conversations   │         │    messages     │
│─────────────────│         │─────────────────│
│ id (UUID) PK    │◄────────│ id (UUID) PK    │
│ title           │       1:N│ conversation_id │
│ created_at      │         │ role            │
│ updated_at      │         │ content         │
└─────────────────┘         │ tool_call_id    │
                            │ raw_message     │
                            │ created_at      │
                            └─────────────────┘
```

---

## 5. Message Format Standard

### 5.1 OpenAI Message Format

The system uses **OpenAI's message format** as the universal standard for storing conversation history. This format is provider-agnostic and works with OpenAI, Anthropic (via conversion), local models, and LangChain.

### 5.2 Message Types

#### 5.2.1 User Message

```json
{
  "role": "user",
  "content": "What is your refund policy?"
}
```

**Storage in PostgreSQL:**
```sql
INSERT INTO messages (id, conversation_id, role, content, raw_message)
VALUES (
    '018c2f1e-3b4a-7d1c-9f2e-3a4b5c6d7e8f',
    '018c2f1e-3b4a-7d1c-9f2e-3a4b5c6d7e8e',
    'user',
    'What is your refund policy?',
    '{"role": "user", "content": "What is your refund policy?"}'::jsonb
);
```

#### 5.2.2 Assistant Message with Tool Calls

```json
{
  "role": "assistant",
  "content": null,
  "tool_calls": [
    {
      "id": "call_abc123",
      "type": "function",
      "function": {
        "name": "find_similar_qa",
        "arguments": "{\"question\": \"refund policy\", \"top_k\": 3}"
      }
    }
  ]
}
```

**Storage in PostgreSQL:**
```sql
INSERT INTO messages (id, conversation_id, role, content, tool_call_id, raw_message)
VALUES (
    '018c2f1e-3b4a-7d1c-9f2e-3a4b5c6d7e90',
    '018c2f1e-3b4a-7d1c-9f2e-3a4b5c6d7e8e',
    'assistant',
    NULL,
    NULL,
    '{
      "role": "assistant",
      "content": null,
      "tool_calls": [{
        "id": "call_abc123",
        "type": "function",
        "function": {
          "name": "find_similar_qa",
          "arguments": "{\"question\": \"refund policy\", \"top_k\": 3}"
        }
      }]
    }'::jsonb
);
```

**Note:** Multiple tool calls can exist in one message. Tool call IDs are inside the JSONB.

#### 5.2.3 Tool Response Message

```json
{
  "role": "tool",
  "tool_call_id": "call_abc123",
  "content": "[{\"id\": 5, \"question\": \"What is your refund policy?\", \"answer\": \"Refunds are processed within 5-7 business days.\"}]"
}
```

**Storage in PostgreSQL:**
```sql
INSERT INTO messages (id, conversation_id, role, content, tool_call_id, raw_message)
VALUES (
    '018c2f1e-3b4a-7d1c-9f2e-3a4b5c6d7e91',
    '018c2f1e-3b4a-7d1c-9f2e-3a4b5c6d7e8e',
    'tool',
    '[{"id": 5, "question": "What is your refund policy?", "answer": "Refunds are processed within 5-7 business days."}]',
    'call_abc123',
    '{
      "role": "tool",
      "tool_call_id": "call_abc123",
      "content": "[{\"id\": 5, \"question\": \"What is your refund policy?\", \"answer\": \"Refunds are processed within 5-7 business days.\"}]"
    }'::jsonb
);
```

**Note:** Each tool response is a separate message. If assistant calls 3 tools, there will be 3 tool response messages.

#### 5.2.4 Final Assistant Message

```json
{
  "role": "assistant",
  "content": "Based on our refund policy, refunds are processed within 5-7 business days. You can request a refund through your account dashboard."
}
```

**Storage in PostgreSQL:**
```sql
INSERT INTO messages (id, conversation_id, role, content, raw_message)
VALUES (
    '018c2f1e-3b4a-7d1c-9f2e-3a4b5c6d7e92',
    '018c2f1e-3b4a-7d1c-9f2e-3a4b5c6d7e8e',
    'assistant',
    'Based on our refund policy, refunds are processed within 5-7 business days.',
    '{"role": "assistant", "content": "Based on our refund policy, refunds are processed within 5-7 business days."}'::jsonb
);
```

### 5.3 Complete Conversation Example

A conversation with AI agent using multiple tools:

```sql
-- Message 1: User asks
INSERT INTO messages VALUES (
    gen_random_uuid(), 
    'conv_id', 
    'user', 
    'How many refund Q&As exist?',
    NULL,
    '{"role": "user", "content": "How many refund Q&As exist?"}'::jsonb
);

-- Message 2: Assistant calls two tools
INSERT INTO messages VALUES (
    gen_random_uuid(), 
    'conv_id', 
    'assistant', 
    NULL,
    NULL,
    '{
      "role": "assistant",
      "content": null,
      "tool_calls": [
        {"id": "call_search_123", "type": "function", "function": {"name": "search_qa", "arguments": "{\"query\": \"refund\"}"}},
        {"id": "call_count_456", "type": "function", "function": {"name": "count_qa", "arguments": "{}"}}
      ]
    }'::jsonb
);

-- Message 3: Tool response for search
INSERT INTO messages VALUES (
    gen_random_uuid(), 
    'conv_id', 
    'tool', 
    '[{...Q&A results...}]',
    'call_search_123',
    '{"role": "tool", "tool_call_id": "call_search_123", "content": "[...]"}'::jsonb
);

-- Message 4: Tool response for count
INSERT INTO messages VALUES (
    gen_random_uuid(), 
    'conv_id', 
    'tool', 
    '{"count": 5}',
    'call_count_456',
    '{"role": "tool", "tool_call_id": "call_count_456", "content": "{\"count\": 5}"}'::jsonb
);

-- Message 5: Final answer
INSERT INTO messages VALUES (
    gen_random_uuid(), 
    'conv_id', 
    'assistant', 
    'We have 5 Q&As about refunds.',
    NULL,
    '{"role": "assistant", "content": "We have 5 Q&As about refunds."}'::jsonb
);
```

### 5.4 Design Rationale

**Why OpenAI Format?**

1. **Industry Standard:** Most widely adopted message format
2. **Provider Agnostic:** Works with OpenAI, Azure OpenAI, OpenRouter, local models
3. **Framework Compatible:** LangChain has built-in conversion functions
4. **Type Safe:** Strict Pydantic models available
5. **Future Proof:** Easily extendable for new features

**Why Store `raw_message`?**

1. **Complete Data:** No information loss
2. **LLM Ready:** Can load directly into conversation context
3. **Queryable:** JSONB supports complex queries if needed
4. **Flexible:** Handles future OpenAI format changes

**Why Extract Fields (`role`, `content`, `tool_call_id`)?**

1. **Query Performance:** Fast filtering without JSONB parsing
2. **Indexing:** Can create indexes on extracted fields
3. **Simple Queries:** Easy to get user messages or filter by role
4. **UI Display:** Direct access to content for display

---

## 6. Data Models

### 6.1 Core Domain Models

```go
package models

import (
    "time"
    "github.com/google/uuid"
)

// QAPair represents a question-answer pair in the knowledge base
type QAPair struct {
    ID        uuid.UUID `db:"id" json:"id"`
    Question  string    `db:"question" json:"question"`
    Answer    string    `db:"answer" json:"answer"`
    CreatedAt time.Time `db:"created_at" json:"created_at"`
    UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// Conversation represents a chat conversation
type Conversation struct {
    ID        uuid.UUID `db:"id" json:"id"`
    Title     *string   `db:"title" json:"title,omitempty"`
    CreatedAt time.Time `db:"created_at" json:"created_at"`
    UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// Message represents a single message in a conversation
type Message struct {
    ID             uuid.UUID              `db:"id" json:"id"`
    ConversationID uuid.UUID              `db:"conversation_id" json:"conversation_id"`
    Role           string                 `db:"role" json:"role"`
    Content        *string                `db:"content" json:"content,omitempty"`
    ToolCallID     *string                `db:"tool_call_id" json:"tool_call_id,omitempty"`
    RawMessage     map[string]interface{} `db:"raw_message" json:"raw_message"`
    CreatedAt      time.Time              `db:"created_at" json:"created_at"`
}
```

### 6.2 Request/Response DTOs

```go
// CreateQARequest represents a request to create a Q&A pair
type CreateQARequest struct {
    Question string `json:"question" validate:"required,min=3,max=1000"`
    Answer   string `json:"answer" validate:"required,min=3,max=5000"`
}

// UpdateQARequest represents a request to update a Q&A pair
type UpdateQARequest struct {
    Question string `json:"question" validate:"required,min=3,max=1000"`
    Answer   string `json:"answer" validate:"required,min=3,max=5000"`
}

// CreateQAResponse represents the response after creating a Q&A pair
type CreateQAResponse struct {
    QAPair QAPair `json:"qa_pair"`
}

// UpdateQAResponse represents the response after updating a Q&A pair
type UpdateQAResponse struct {
    QAPair QAPair `json:"qa_pair"`
}

// ListQAResponse represents a paginated list of Q&A pairs
type ListQAResponse struct {
    Data       []QAPair         `json:"data"`
    Pagination CursorPagination `json:"pagination"`
}

// CreateConversationRequest represents a request to create a conversation
type CreateConversationRequest struct {
    Title string `json:"title" validate:"omitempty,max=200"`
}

// CreateConversationResponse represents the response after creating a conversation
type CreateConversationResponse struct {
    Conversation Conversation `json:"conversation"`
}

// ListConversationsResponse represents a paginated list of conversations
type ListConversationsResponse struct {
    Data       []Conversation   `json:"data"`
    Pagination CursorPagination `json:"pagination"`
}

// CreateMessageRequest represents a request to create a message
type CreateMessageRequest struct {
    ConversationID uuid.UUID              `json:"conversation_id" validate:"required"`
    Role           string                 `json:"role" validate:"required,oneof=user assistant tool system"`
    Content        *string                `json:"content,omitempty"`
    ToolCallID     *string                `json:"tool_call_id,omitempty"`
    RawMessage     map[string]interface{} `json:"raw_message" validate:"required"`
}

// CreateMessageResponse represents the response after creating a message
type CreateMessageResponse struct {
    Message Message `json:"message"`
}

// ListMessagesResponse represents a paginated list of messages
type ListMessagesResponse struct {
    Data       []Message        `json:"data"`
    Pagination CursorPagination `json:"pagination"`
}
```

### 6.3 Pagination Models

```go
// CursorParams represents cursor pagination parameters
type CursorParams struct {
    Limit     int    `form:"limit" validate:"min=1,max=100"`
    Cursor    string `form:"cursor"`
    Direction string `form:"direction" validate:"omitempty,oneof=next prev"`
    Search    string `form:"search" validate:"omitempty,max=200"`
}

// CursorPagination represents cursor pagination metadata
type CursorPagination struct {
    NextCursor string `json:"next_cursor,omitempty"`
    PrevCursor string `json:"prev_cursor,omitempty"`
    HasNext    bool   `json:"has_next"`
    HasPrev    bool   `json:"has_prev"`
}

// NewCursorParams creates default cursor params
func NewCursorParams() CursorParams {
    return CursorParams{
        Limit:     10,
        Direction: "next",
    }
}
```

### 6.4 Tool API Models

```go
// FindSimilarRequest represents a request to find similar Q&A pairs
type FindSimilarRequest struct {
    Embedding []float32 `json:"embedding" validate:"required,dive,number"`
    TopK      int       `json:"top_k" validate:"required,min=1,max=20"`
}

// SimilarityMatch represents a Q&A pair with similarity score
type SimilarityMatch struct {
    QAPair QAPair  `json:"qa_pair"`
    Score  float32 `json:"score"`
}

// FindSimilarResponse represents the response from similarity search
type FindSimilarResponse struct {
    Results []SimilarityMatch `json:"results"`
}

// GetQAByIDsRequest represents a request to get multiple Q&A pairs by IDs
type GetQAByIDsRequest struct {
    IDs []uuid.UUID `json:"ids" validate:"required,min=1,max=50,dive,required"`
}

// GetQAByIDsResponse represents the response with multiple Q&A pairs
type GetQAByIDsResponse struct {
    QAPairs []QAPair `json:"qa_pairs"`
}

// CreateQAWithEmbeddingRequest represents a request to create Q&A with embedding
type CreateQAWithEmbeddingRequest struct {
    Question  string    `json:"question" validate:"required,min=3,max=1000"`
    Answer    string    `json:"answer" validate:"required,min=3,max=5000"`
    Embedding []float32 `json:"embedding" validate:"required,dive,number"`
}

// UpdateQAWithEmbeddingRequest represents a request to update Q&A with embedding
type UpdateQAWithEmbeddingRequest struct {
    ID        uuid.UUID `json:"id" validate:"required"`
    Question  string    `json:"question" validate:"required,min=3,max=1000"`
    Answer    string    `json:"answer" validate:"required,min=3,max=5000"`
    Embedding []float32 `json:"embedding" validate:"required,dive,number"`
}

// DeleteQARequest represents a request to delete a Q&A pair
type DeleteQARequest struct {
    ID uuid.UUID `json:"id" validate:"required"`
}

// DeleteQAResponse represents the response after deleting a Q&A pair
type DeleteQAResponse struct {
    Success             bool `json:"success"`
    DeletedFromDB       bool `json:"deleted_from_db"`
    DeletedFromPinecone bool `json:"deleted_from_pinecone"`
}

// SearchQARequest represents a full-text search request
type SearchQARequest struct {
    Query string `json:"query" validate:"required,min=1,max=200"`
    Limit int    `json:"limit" validate:"required,min=1,max=100"`
}

// SearchQAResponse represents the search response
type SearchQAResponse struct {
    QAPairs []QAPair `json:"qa_pairs"`
    Count   int      `json:"count"`
}

// SaveMessageRequest represents a request to save a message from Python agent
type SaveMessageRequest struct {
    ConversationID uuid.UUID              `json:"conversation_id" validate:"required"`
    Role           string                 `json:"role" validate:"required,oneof=user assistant tool system"`
    Content        *string                `json:"content"`
    ToolCallID     *string                `json:"tool_call_id"`
    RawMessage     map[string]interface{} `json:"raw_message" validate:"required"`
}

// SaveMessageResponse represents the response after saving a message
type SaveMessageResponse struct {
    Message Message `json:"message"`
}
```

### 6.5 Configuration Models

```go
// Config represents the application configuration
type Config struct {
    Server    ServerConfig    `mapstructure:"server"`
    Database  DatabaseConfig  `mapstructure:"database"`
    Pinecone  PineconeConfig  `mapstructure:"pinecone"`
    Embedding EmbeddingConfig `mapstructure:"embedding"`
}

// ServerConfig represents HTTP server configuration
type ServerConfig struct {
    Port        int    `mapstructure:"port" validate:"required,min=1,max=65535"`
    Host        string `mapstructure:"host"`
    Environment string `mapstructure:"environment" validate:"required,oneof=development staging production"`
}

// DatabaseConfig represents PostgreSQL configuration
type DatabaseConfig struct {
    Host         string `mapstructure:"host" validate:"required"`
    Port         int    `mapstructure:"port" validate:"required,min=1,max=65535"`
    User         string `mapstructure:"user" validate:"required"`
    Password     string `mapstructure:"password" validate:"required"`
    DBName       string `mapstructure:"dbname" validate:"required"`
    SSLMode      string `mapstructure:"sslmode" validate:"required,oneof=disable require verify-ca verify-full"`
    MaxOpenConns int    `mapstructure:"max_open_conns" validate:"min=1"`
    MaxIdleConns int    `mapstructure:"max_idle_conns" validate:"min=1"`
}

// PineconeConfig represents Pinecone vector database configuration
type PineconeConfig struct {
    APIKey      string `mapstructure:"api_key" validate:"required"`
    Environment string `mapstructure:"environment" validate:"required"`
    IndexName   string `mapstructure:"index_name" validate:"required"`
    Dimension   int    `mapstructure:"dimension" validate:"required,min=1"`
}

// EmbeddingConfig represents embedding model configuration
type EmbeddingConfig struct {
    Dimension int    `mapstructure:"dimension" validate:"required,min=1"`
    Model     string `mapstructure:"model" validate:"required"`
}

// ConnectionString builds PostgreSQL connection string
func (c DatabaseConfig) ConnectionString() string {
    return fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
    )
}
```

### 6.6 Error Models

```go
// ErrorResponse represents a standardized error response
type ErrorResponse struct {
    Error   string                 `json:"error"`
    Code    string                 `json:"code"`
    Message string                 `json:"message"`
    Details map[string]interface{} `json:"details,omitempty"`
}

// ValidationError represents a field validation error
type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}

// Error codes
const (
    ErrCodeValidation      = "VALIDATION_ERROR"
    ErrCodeNotFound        = "NOT_FOUND"
    ErrCodeInternal        = "INTERNAL_ERROR"
    ErrCodeDatabaseError   = "DATABASE_ERROR"
    ErrCodePineconeError   = "PINECONE_ERROR"
    ErrCodeUnauthorized    = "UNAUTHORIZED"
    ErrCodeForbidden       = "FORBIDDEN"
    ErrCodeBadRequest      = "BAD_REQUEST"
)

// NewErrorResponse creates a new error response
func NewErrorResponse(code, message string, details map[string]interface{}) ErrorResponse {
    return ErrorResponse{
        Error:   "error",
        Code:    code,
        Message: message,
        Details: details,
    }
}
```

---

## 7. Repository Layer

### 7.1 QA Repository Interface

```go
package repository

import (
    "context"
    "github.com/google/uuid"
    "smart-company-discovery/internal/models"
)

// QARepository defines Q&A data access operations
type QARepository interface {
    // Create creates a new Q&A pair (generates UUIDv7)
    Create(ctx context.Context, qa *models.QAPair) error
    
    // GetByID retrieves a Q&A pair by UUID
    GetByID(ctx context.Context, id uuid.UUID) (*models.QAPair, error)
    
    // GetByIDs retrieves multiple Q&A pairs by UUIDs (batch)
    GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.QAPair, error)
    
    // Update updates an existing Q&A pair
    Update(ctx context.Context, qa *models.QAPair) error
    
    // Delete deletes a Q&A pair by UUID
    Delete(ctx context.Context, id uuid.UUID) error
    
    // List retrieves Q&A pairs with cursor pagination
    List(ctx context.Context, params models.CursorParams) ([]*models.QAPair, *models.CursorPagination, error)
    
    // SearchFullText performs full-text search on Q&A pairs
    SearchFullText(ctx context.Context, query string, params models.CursorParams) ([]*models.QAPair, *models.CursorPagination, error)
    
    // Count returns total count of Q&A pairs
    Count(ctx context.Context) (int, error)
}
```

### 7.2 QA Repository Implementation

```go
package repository

import (
    "context"
    "database/sql"
    "fmt"
    "github.com/google/uuid"
    "github.com/jmoiron/sqlx"
    "smart-company-discovery/internal/models"
)

type qaRepository struct {
    db *sqlx.DB
}

// NewQARepository creates a new QA repository
func NewQARepository(db *sqlx.DB) QARepository {
    return &qaRepository{db: db}
}

// Create creates a new Q&A pair
func (r *qaRepository) Create(ctx context.Context, qa *models.QAPair) error {
    // Generate UUIDv7
    qa.ID = uuid.Must(uuid.NewV7())
    
    query := `
        INSERT INTO qa_pairs (id, question, answer)
        VALUES ($1, $2, $3)
        RETURNING created_at, updated_at
    `
    
    return r.db.GetContext(ctx, qa, query, qa.ID, qa.Question, qa.Answer)
}

// GetByID retrieves a Q&A pair by UUID
func (r *qaRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.QAPair, error) {
    var qa models.QAPair
    
    query := `
        SELECT id, question, answer, created_at, updated_at
        FROM qa_pairs
        WHERE id = $1
    `
    
    err := r.db.GetContext(ctx, &qa, query, id)
    if err == sql.ErrNoRows {
        return nil, nil
    }
    return &qa, err
}

// GetByIDs retrieves multiple Q&A pairs by UUIDs
func (r *qaRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.QAPair, error) {
    if len(ids) == 0 {
        return []*models.QAPair{}, nil
    }
    
    query := `
        SELECT id, question, answer, created_at, updated_at
        FROM qa_pairs
        WHERE id = ANY($1)
        ORDER BY id DESC
    `
    
    var qaPairs []*models.QAPair
    err := r.db.SelectContext(ctx, &qaPairs, query, ids)
    return qaPairs, err
}

// Update updates an existing Q&A pair
func (r *qaRepository) Update(ctx context.Context, qa *models.QAPair) error {
    query := `
        UPDATE qa_pairs
        SET question = $2, answer = $3
        WHERE id = $1
        RETURNING updated_at
    `
    
    return r.db.GetContext(ctx, qa, query, qa.ID, qa.Question, qa.Answer)
}

// Delete deletes a Q&A pair
func (r *qaRepository) Delete(ctx context.Context, id uuid.UUID) error {
    query := `DELETE FROM qa_pairs WHERE id = $1`
    
    result, err := r.db.ExecContext(ctx, query, id)
    if err != nil {
        return err
    }
    
    rows, err := result.RowsAffected()
    if err != nil {
        return err
    }
    
    if rows == 0 {
        return sql.ErrNoRows
    }
    
    return nil
}

// List retrieves Q&A pairs with cursor pagination
func (r *qaRepository) List(ctx context.Context, params models.CursorParams) ([]*models.QAPair, *models.CursorPagination, error) {
    // Set defaults
    if params.Limit < 1 {
        params.Limit = 10
    }
    if params.Limit > 100 {
        params.Limit = 100
    }
    if params.Direction == "" {
        params.Direction = "next"
    }
    
    // Build query
    whereClauses := []string{}
    args := []interface{}{}
    argIdx := 1
    
    // Cursor filter
    if params.Cursor != "" {
        cursorID, err := uuid.Parse(params.Cursor)
        if err != nil {
            return nil, nil, fmt.Errorf("invalid cursor: %w", err)
        }
        
        if params.Direction == "prev" {
            whereClauses = append(whereClauses, fmt.Sprintf("id > $%d", argIdx))
        } else {
            whereClauses = append(whereClauses, fmt.Sprintf("id < $%d", argIdx))
        }
        args = append(args, cursorID)
        argIdx++
    }
    
    whereSQL := ""
    if len(whereClauses) > 0 {
        whereSQL = "WHERE " + whereClauses[0]
    }
    
    // Order
    order := "DESC"
    if params.Direction == "prev" {
        order = "ASC"
    }
    
    // Fetch one extra to check if there's more
    fetchLimit := params.Limit + 1
    
    query := fmt.Sprintf(`
        SELECT id, question, answer, created_at, updated_at
        FROM qa_pairs
        %s
        ORDER BY id %s
        LIMIT $%d
    `, whereSQL, order, argIdx)
    
    args = append(args, fetchLimit)
    
    var qaPairs []*models.QAPair
    err := r.db.SelectContext(ctx, &qaPairs, query, args...)
    if err != nil {
        return nil, nil, err
    }
    
    // Check if there are more results
    hasMore := len(qaPairs) > params.Limit
    if hasMore {
        qaPairs = qaPairs[:params.Limit]
    }
    
    // Reverse if prev direction
    if params.Direction == "prev" {
        for i, j := 0, len(qaPairs)-1; i < j; i, j = i+1, j-1 {
            qaPairs[i], qaPairs[j] = qaPairs[j], qaPairs[i]
        }
    }
    
    // Build pagination
    pagination := &models.CursorPagination{}
    
    if len(qaPairs) > 0 {
        pagination.NextCursor = qaPairs[len(qaPairs)-1].ID.String()
        pagination.PrevCursor = qaPairs[0].ID.String()
        pagination.HasNext = hasMore
        pagination.HasPrev = params.Cursor != ""
    }
    
    return qaPairs, pagination, nil
}

// SearchFullText performs full-text search
func (r *qaRepository) SearchFullText(ctx context.Context, searchQuery string, params models.CursorParams) ([]*models.QAPair, *models.CursorPagination, error) {
    // Set defaults
    if params.Limit < 1 {
        params.Limit = 10
    }
    if params.Limit > 100 {
        params.Limit = 100
    }
    
    // Build full-text search query
    query := `
        SELECT 
            id, question, answer, created_at, updated_at,
            ts_rank(
                to_tsvector('english', question || ' ' || answer),
                to_tsquery('english', $1)
            ) as rank
        FROM qa_pairs
        WHERE to_tsvector('english', question || ' ' || answer) 
              @@ to_tsquery('english', $1)
        ORDER BY rank DESC, id DESC
        LIMIT $2
    `
    
    fetchLimit := params.Limit + 1
    
    var qaPairs []*models.QAPair
    err := r.db.SelectContext(ctx, &qaPairs, query, searchQuery, fetchLimit)
    if err != nil {
        return nil, nil, err
    }
    
    // Check if there are more results
    hasMore := len(qaPairs) > params.Limit
    if hasMore {
        qaPairs = qaPairs[:params.Limit]
    }
    
    // Build pagination
    pagination := &models.CursorPagination{
        HasNext: hasMore,
        HasPrev: false, // Search doesn't support cursor navigation
    }
    
    return qaPairs, pagination, nil
}

// Count returns total count of Q&A pairs
func (r *qaRepository) Count(ctx context.Context) (int, error) {
    var count int
    query := `SELECT COUNT(*) FROM qa_pairs`
    err := r.db.GetContext(ctx, &count, query)
    return count, err
}
```

### 7.3 Conversation Repository Interface

```go
// ConversationRepository defines conversation data access operations
type ConversationRepository interface {
    // CreateConversation creates a new conversation
    CreateConversation(ctx context.Context, conv *models.Conversation) error
    
    // GetConversation retrieves a conversation by UUID
    GetConversation(ctx context.Context, id uuid.UUID) (*models.Conversation, error)
    
    // ListConversations retrieves conversations with cursor pagination
    ListConversations(ctx context.Context, params models.CursorParams) ([]*models.Conversation, *models.CursorPagination, error)
    
    // DeleteConversation deletes a conversation (cascades to messages)
    DeleteConversation(ctx context.Context, id uuid.UUID) error
    
    // CreateMessage creates a new message in a conversation
    CreateMessage(ctx context.Context, msg *models.Message) error
    
    // GetMessages retrieves messages for a conversation with cursor pagination
    GetMessages(ctx context.Context, conversationID uuid.UUID, params models.CursorParams) ([]*models.Message, *models.CursorPagination, error)
}
```

### 7.4 Conversation Repository Implementation

```go
type conversationRepository struct {
    db *sqlx.DB
}

// NewConversationRepository creates a new conversation repository
func NewConversationRepository(db *sqlx.DB) ConversationRepository {
    return &conversationRepository{db: db}
}

// CreateConversation creates a new conversation
func (r *conversationRepository) CreateConversation(ctx context.Context, conv *models.Conversation) error {
    conv.ID = uuid.Must(uuid.NewV7())
    
    query := `
        INSERT INTO conversations (id, title)
        VALUES ($1, $2)
        RETURNING created_at, updated_at
    `
    
    return r.db.GetContext(ctx, conv, query, conv.ID, conv.Title)
}

// GetConversation retrieves a conversation by UUID
func (r *conversationRepository) GetConversation(ctx context.Context, id uuid.UUID) (*models.Conversation, error) {
    var conv models.Conversation
    
    query := `
        SELECT id, title, created_at, updated_at
        FROM conversations
        WHERE id = $1
    `
    
    err := r.db.GetContext(ctx, &conv, query, id)
    if err == sql.ErrNoRows {
        return nil, nil
    }
    return &conv, err
}

// ListConversations retrieves conversations with cursor pagination
func (r *conversationRepository) ListConversations(ctx context.Context, params models.CursorParams) ([]*models.Conversation, *models.CursorPagination, error) {
    // Implementation similar to QA List
    // ...
    return nil, nil, nil
}

// DeleteConversation deletes a conversation
func (r *conversationRepository) DeleteConversation(ctx context.Context, id uuid.UUID) error {
    query := `DELETE FROM conversations WHERE id = $1`
    
    result, err := r.db.ExecContext(ctx, query, id)
    if err != nil {
        return err
    }
    
    rows, err := result.RowsAffected()
    if err != nil {
        return err
    }
    
    if rows == 0 {
        return sql.ErrNoRows
    }
    
    return nil
}

// CreateMessage creates a new message
func (r *conversationRepository) CreateMessage(ctx context.Context, msg *models.Message) error {
    msg.ID = uuid.Must(uuid.NewV7())
    
    query := `
        INSERT INTO messages (id, conversation_id, role, content, tool_call_id, raw_message)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING created_at
    `
    
    return r.db.GetContext(ctx, msg, query, 
        msg.ID, msg.ConversationID, msg.Role, msg.Content, msg.ToolCallID, msg.RawMessage)
}

// GetMessages retrieves messages for a conversation
func (r *conversationRepository) GetMessages(ctx context.Context, conversationID uuid.UUID, params models.CursorParams) ([]*models.Message, *models.CursorPagination, error) {
    // Set defaults
    if params.Limit < 1 {
        params.Limit = 50
    }
    if params.Limit > 100 {
        params.Limit = 100
    }
    if params.Direction == "" {
        params.Direction = "next"
    }
    
    // Build query
    whereClauses := []string{fmt.Sprintf("conversation_id = $1")}
    args := []interface{}{conversationID}
    argIdx := 2
    
    // Cursor filter
    if params.Cursor != "" {
        cursorID, err := uuid.Parse(params.Cursor)
        if err != nil {
            return nil, nil, fmt.Errorf("invalid cursor: %w", err)
        }
        
        if params.Direction == "prev" {
            whereClauses = append(whereClauses, fmt.Sprintf("id < $%d", argIdx))
        } else {
            whereClauses = append(whereClauses, fmt.Sprintf("id > $%d", argIdx))
        }
        args = append(args, cursorID)
        argIdx++
    }
    
    whereSQL := "WHERE " + whereClauses[0]
    if len(whereClauses) > 1 {
        whereSQL += " AND " + whereClauses[1]
    }
    
    // Order (messages are ASC by default for chronological display)
    order := "ASC"
    if params.Direction == "prev" {
        order = "DESC"
    }
    
    fetchLimit := params.Limit + 1
    
    query := fmt.Sprintf(`
        SELECT id, conversation_id, role, content, tool_call_id, raw_message, created_at
        FROM messages
        %s
        ORDER BY id %s
        LIMIT $%d
    `, whereSQL, order, argIdx)
    
    args = append(args, fetchLimit)
    
    var messages []*models.Message
    err := r.db.SelectContext(ctx, &messages, query, args...)
    if err != nil {
        return nil, nil, err
    }
    
    // Check if there are more results
    hasMore := len(messages) > params.Limit
    if hasMore {
        messages = messages[:params.Limit]
    }
    
    // Reverse if prev direction
    if params.Direction == "prev" {
        for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
            messages[i], messages[j] = messages[j], messages[i]
        }
    }
    
    // Build pagination
    pagination := &models.CursorPagination{}
    
    if len(messages) > 0 {
        pagination.NextCursor = messages[len(messages)-1].ID.String()
        pagination.PrevCursor = messages[0].ID.String()
        pagination.HasNext = hasMore
        pagination.HasPrev = params.Cursor != ""
    }
    
    return messages, pagination, nil
}
```

---

## 8. Service Layer

### 8.1 Service Layer Overview

The service layer contains business logic and orchestrates repository operations. It sits between HTTP handlers and repositories, providing transaction management and complex operations.

### 8.2 QA Service Interface

```go
package service

import (
    "context"
    "github.com/google/uuid"
    "smart-company-discovery/internal/models"
)

// QAService defines Q&A business logic operations
type QAService interface {
    // For React API
    CreateQA(ctx context.Context, req models.CreateQARequest) (*models.QAPair, error)
    GetQA(ctx context.Context, id uuid.UUID) (*models.QAPair, error)
    UpdateQA(ctx context.Context, id uuid.UUID, req models.UpdateQARequest) (*models.QAPair, error)
    DeleteQA(ctx context.Context, id uuid.UUID) error
    ListQA(ctx context.Context, params models.CursorParams) ([]*models.QAPair, *models.CursorPagination, error)
    SearchQA(ctx context.Context, query string, params models.CursorParams) ([]*models.QAPair, *models.CursorPagination, error)
    
    // For Python Tool API
    FindSimilar(ctx context.Context, embedding []float32, topK int) ([]models.SimilarityMatch, error)
    GetQAByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.QAPair, error)
    CreateQAWithEmbedding(ctx context.Context, req models.CreateQAWithEmbeddingRequest) (*models.QAPair, error)
    UpdateQAWithEmbedding(ctx context.Context, req models.UpdateQAWithEmbeddingRequest) (*models.QAPair, error)
    DeleteQAWithEmbedding(ctx context.Context, id uuid.UUID) (*models.DeleteQAResponse, error)
}
```

### 8.3 QA Service Implementation

```go
package service

import (
    "context"
    "fmt"
    "github.com/google/uuid"
    "smart-company-discovery/internal/models"
    "smart-company-discovery/internal/repository"
    "smart-company-discovery/internal/clients"
)

type qaService struct {
    qaRepo    repository.QARepository
    pinecone  clients.PineconeClient
}

// NewQAService creates a new QA service
func NewQAService(qaRepo repository.QARepository, pinecone clients.PineconeClient) QAService {
    return &qaService{
        qaRepo:   qaRepo,
        pinecone: pinecone,
    }
}

// CreateQA creates a new Q&A pair (without embedding)
func (s *qaService) CreateQA(ctx context.Context, req models.CreateQARequest) (*models.QAPair, error) {
    qa := &models.QAPair{
        Question: req.Question,
        Answer:   req.Answer,
    }
    
    err := s.qaRepo.Create(ctx, qa)
    if err != nil {
        return nil, fmt.Errorf("failed to create Q&A: %w", err)
    }
    
    return qa, nil
}

// GetQA retrieves a Q&A pair by UUID
func (s *qaService) GetQA(ctx context.Context, id uuid.UUID) (*models.QAPair, error) {
    qa, err := s.qaRepo.GetByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get Q&A: %w", err)
    }
    if qa == nil {
        return nil, fmt.Errorf("Q&A not found")
    }
    return qa, nil
}

// UpdateQA updates an existing Q&A pair
func (s *qaService) UpdateQA(ctx context.Context, id uuid.UUID, req models.UpdateQARequest) (*models.QAPair, error) {
    // Check if exists
    existing, err := s.qaRepo.GetByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get Q&A: %w", err)
    }
    if existing == nil {
        return nil, fmt.Errorf("Q&A not found")
    }
    
    // Update fields
    existing.Question = req.Question
    existing.Answer = req.Answer
    
    err = s.qaRepo.Update(ctx, existing)
    if err != nil {
        return nil, fmt.Errorf("failed to update Q&A: %w", err)
    }
    
    return existing, nil
}

// DeleteQA deletes a Q&A pair
func (s *qaService) DeleteQA(ctx context.Context, id uuid.UUID) error {
    err := s.qaRepo.Delete(ctx, id)
    if err != nil {
        return fmt.Errorf("failed to delete Q&A: %w", err)
    }
    return nil
}

// ListQA lists Q&A pairs with cursor pagination
func (s *qaService) ListQA(ctx context.Context, params models.CursorParams) ([]*models.QAPair, *models.CursorPagination, error) {
    return s.qaRepo.List(ctx, params)
}

// SearchQA performs full-text search
func (s *qaService) SearchQA(ctx context.Context, query string, params models.CursorParams) ([]*models.QAPair, *models.CursorPagination, error) {
    return s.qaRepo.SearchFullText(ctx, query, params)
}

// FindSimilar finds similar Q&A pairs using vector search
func (s *qaService) FindSimilar(ctx context.Context, embedding []float32, topK int) ([]models.SimilarityMatch, error) {
    // Query Pinecone
    matches, err := s.pinecone.Query(ctx, embedding, topK)
    if err != nil {
        return nil, fmt.Errorf("pinecone query failed: %w", err)
    }
    
    // Extract IDs
    ids := make([]uuid.UUID, 0, len(matches))
    scoreMap := make(map[uuid.UUID]float32)
    
    for _, match := range matches {
        id, err := uuid.Parse(match.ID)
        if err != nil {
            continue
        }
        ids = append(ids, id)
        scoreMap[id] = match.Score
    }
    
    // Fetch Q&A pairs from PostgreSQL
    qaPairs, err := s.qaRepo.GetByIDs(ctx, ids)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch Q&A pairs: %w", err)
    }
    
    // Build similarity matches
    results := make([]models.SimilarityMatch, 0, len(qaPairs))
    for _, qa := range qaPairs {
        results = append(results, models.SimilarityMatch{
            QAPair: *qa,
            Score:  scoreMap[qa.ID],
        })
    }
    
    return results, nil
}

// GetQAByIDs retrieves multiple Q&A pairs by UUIDs
func (s *qaService) GetQAByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.QAPair, error) {
    return s.qaRepo.GetByIDs(ctx, ids)
}

// CreateQAWithEmbedding creates a Q&A pair and stores embedding in Pinecone
func (s *qaService) CreateQAWithEmbedding(ctx context.Context, req models.CreateQAWithEmbeddingRequest) (*models.QAPair, error) {
    // Create in PostgreSQL
    qa := &models.QAPair{
        Question: req.Question,
        Answer:   req.Answer,
    }
    
    err := s.qaRepo.Create(ctx, qa)
    if err != nil {
        return nil, fmt.Errorf("failed to create Q&A: %w", err)
    }
    
    // Store embedding in Pinecone
    metadata := map[string]interface{}{
        "id":       qa.ID.String(),
        "question": qa.Question,
        "answer":   qa.Answer,
    }
    
    err = s.pinecone.Upsert(ctx, qa.ID.String(), req.Embedding, metadata)
    if err != nil {
        // Note: Q&A is already created in PostgreSQL
        // Consider implementing compensation logic or transaction
        return nil, fmt.Errorf("failed to store embedding: %w", err)
    }
    
    return qa, nil
}

// UpdateQAWithEmbedding updates Q&A pair and embedding
func (s *qaService) UpdateQAWithEmbedding(ctx context.Context, req models.UpdateQAWithEmbeddingRequest) (*models.QAPair, error) {
    // Get existing
    existing, err := s.qaRepo.GetByID(ctx, req.ID)
    if err != nil {
        return nil, fmt.Errorf("failed to get Q&A: %w", err)
    }
    if existing == nil {
        return nil, fmt.Errorf("Q&A not found")
    }
    
    // Update PostgreSQL
    existing.Question = req.Question
    existing.Answer = req.Answer
    
    err = s.qaRepo.Update(ctx, existing)
    if err != nil {
        return nil, fmt.Errorf("failed to update Q&A: %w", err)
    }
    
    // Update Pinecone
    metadata := map[string]interface{}{
        "id":       existing.ID.String(),
        "question": existing.Question,
        "answer":   existing.Answer,
    }
    
    err = s.pinecone.Upsert(ctx, existing.ID.String(), req.Embedding, metadata)
    if err != nil {
        return nil, fmt.Errorf("failed to update embedding: %w", err)
    }
    
    return existing, nil
}

// DeleteQAWithEmbedding deletes from both PostgreSQL and Pinecone
func (s *qaService) DeleteQAWithEmbedding(ctx context.Context, id uuid.UUID) (*models.DeleteQAResponse, error) {
    response := &models.DeleteQAResponse{}
    
    // Delete from PostgreSQL
    err := s.qaRepo.Delete(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to delete Q&A from database: %w", err)
    }
    response.DeletedFromDB = true
    
    // Delete from Pinecone
    err = s.pinecone.Delete(ctx, id.String())
    if err != nil {
        // Log error but don't fail the operation
        // Q&A is already deleted from PostgreSQL
        response.DeletedFromPinecone = false
    } else {
        response.DeletedFromPinecone = true
    }
    
    response.Success = response.DeletedFromDB
    return response, nil
}
```

### 8.4 Conversation Service

```go
// ConversationService defines conversation business logic operations
type ConversationService interface {
    CreateConversation(ctx context.Context, title string) (*models.Conversation, error)
    GetConversation(ctx context.Context, id uuid.UUID) (*models.Conversation, error)
    ListConversations(ctx context.Context, params models.CursorParams) ([]*models.Conversation, *models.CursorPagination, error)
    DeleteConversation(ctx context.Context, id uuid.UUID) error
    
    AddMessage(ctx context.Context, req models.CreateMessageRequest) (*models.Message, error)
    GetMessages(ctx context.Context, conversationID uuid.UUID, params models.CursorParams) ([]*models.Message, *models.CursorPagination, error)
}

type conversationService struct {
    convRepo repository.ConversationRepository
}

func NewConversationService(convRepo repository.ConversationRepository) ConversationService {
    return &conversationService{convRepo: convRepo}
}

func (s *conversationService) CreateConversation(ctx context.Context, title string) (*models.Conversation, error) {
    conv := &models.Conversation{
        Title: &title,
    }
    
    err := s.convRepo.CreateConversation(ctx, conv)
    if err != nil {
        return nil, fmt.Errorf("failed to create conversation: %w", err)
    }
    
    return conv, nil
}

func (s *conversationService) AddMessage(ctx context.Context, req models.CreateMessageRequest) (*models.Message, error) {
    // Validate conversation exists
    conv, err := s.convRepo.GetConversation(ctx, req.ConversationID)
    if err != nil {
        return nil, fmt.Errorf("failed to get conversation: %w", err)
    }
    if conv == nil {
        return nil, fmt.Errorf("conversation not found")
    }
    
    msg := &models.Message{
        ConversationID: req.ConversationID,
        Role:           req.Role,
        Content:        req.Content,
        ToolCallID:     req.ToolCallID,
        RawMessage:     req.RawMessage,
    }
    
    err = s.convRepo.CreateMessage(ctx, msg)
    if err != nil {
        return nil, fmt.Errorf("failed to create message: %w", err)
    }
    
    return msg, nil
}
```

---

## 9. API Specification

### 9.1 Base URL

```
Development: http://localhost:8080
Production: https://api.company.com
```

### 9.2 React UI Endpoints

#### 9.2.1 List Q&A Pairs

**Endpoint:** `GET /api/qa-pairs`

**Query Parameters:**
- `limit` (integer, optional): Number of items to return (1-100, default: 10)
- `cursor` (string, optional): UUID cursor for pagination
- `direction` (string, optional): `next` or `prev` (default: `next`)
- `search` (string, optional): Full-text search query

**Response:**
```json
{
  "data": [
    {
      "id": "018c2f1e-3b4a-7d1c-9f2e-3a4b5c6d7e8f",
      "question": "What is your refund policy?",
      "answer": "Refunds are processed within 5-7 business days.",
      "created_at": "2024-10-30T14:00:00Z",
      "updated_at": "2024-10-30T14:00:00Z"
    }
  ],
  "pagination": {
    "next_cursor": "018c2f1e-3b4a-7d1c-9f2e-3a4b5c6d7e90",
    "prev_cursor": "018c2f1e-3b4a-7d1c-9f2e-3a4b5c6d7e8f",
    "has_next": true,
    "has_prev": false
  }
}
```

**cURL Example:**
```bash
curl -X GET "http://localhost:8080/api/qa-pairs?limit=10&search=refund"
```

#### 9.2.2 Get Single Q&A

**Endpoint:** `GET /api/qa-pairs/:id`

**Response:**
```json
{
  "id": "018c2f1e-3b4a-7d1c-9f2e-3a4b5c6d7e8f",
  "question": "What is your refund policy?",
  "answer": "Refunds are processed within 5-7 business days.",
  "created_at": "2024-10-30T14:00:00Z",
  "updated_at": "2024-10-30T14:00:00Z"
}
```

**cURL Example:**
```bash
curl -X GET "http://localhost:8080/api/qa-pairs/018c2f1e-3b4a-7d1c-9f2e-3a4b5c6d7e8f"
```

#### 9.2.3 Create Q&A Pair

**Endpoint:** `POST /api/qa-pairs`

**Request Body:**
```json
{
  "question": "What is your refund policy?",
  "answer": "Refunds are processed within 5-7 business days."
}
```

**Response:**
```json
{
  "qa_pair": {
    "id": "018c2f1e-3b4a-7d1c-9f2e-3a4b5c6d7e8f",
    "question": "What is your refund policy?",
    "answer": "Refunds are processed within 5-7 business days.",
    "created_at": "2024-10-30T14:00:00Z",
    "updated_at": "2024-10-30T14:00:00Z"
  }
}
```

**cURL Example:**
```bash
curl -X POST "http://localhost:8080/api/qa-pairs" \
  -H "Content-Type: application/json" \
  -d '{
    "question": "What is your refund policy?",
    "answer": "Refunds are processed within 5-7 business days."
  }'
```

#### 9.2.4 Update Q&A Pair

**Endpoint:** `PUT /api/qa-pairs/:id`

**Request Body:**
```json
{
  "question": "What is your updated refund policy?",
  "answer": "Refunds are now processed within 3-5 business days."
}
```

**Response:**
```json
{
  "qa_pair": {
    "id": "018c2f1e-3b4a-7d1c-9f2e-3a4b5c6d7e8f",
    "question": "What is your updated refund policy?",
    "answer": "Refunds are now processed within 3-5 business days.",
    "created_at": "2024-10-30T14:00:00Z",
    "updated_at": "2024-10-30T15:00:00Z"
  }
}
```

**cURL Example:**
```bash
curl -X PUT "http://localhost:8080/api/qa-pairs/018c2f1e-3b4a-7d1c-9f2e-3a4b5c6d7e8f" \
  -H "Content-Type: application/json" \
  -d '{
    "question": "What is your updated refund policy?",
    "answer": "Refunds are now processed within 3-5 business days."
  }'
```

#### 9.2.5 Delete Q&A Pair

**Endpoint:** `DELETE /api/qa-pairs/:id`

**Response:**
```json
{
  "success": true
}
```

**cURL Example:**
```bash
curl -X DELETE "http://localhost:8080/api/qa-pairs/018c2f1e-3b4a-7d1c-9f2e-3a4b5c6d7e8f"
```

#### 9.2.6 Create Conversation

**Endpoint:** `POST /api/conversations`

**Request Body:**
```json
{
  "title": "Refund Policy Discussion"
}
```

**Response:**
```json
{
  "conversation": {
    "id": "018c2f1e-3b4a-7d1c-9f2e-3a4b5c6d7e8e",
    "title": "Refund Policy Discussion",
    "created_at": "2024-10-30T14:00:00Z",
    "updated_at": "2024-10-30T14:00:00Z"
  }
}
```

#### 9.2.7 List Conversations

**Endpoint:** `GET /api/conversations`

**Query Parameters:**
- `limit` (integer, optional): Number of items (default: 20)
- `cursor` (string, optional): UUID cursor
- `direction` (string, optional): `next` or `prev`

**Response:**
```json
{
  "data": [
    {
      "id": "018c2f1e-3b4a-7d1c-9f2e-3a4b5c6d7e8e",
      "title": "Refund Policy Discussion",
      "created_at": "2024-10-30T14:00:00Z",
      "updated_at": "2024-10-30T14:00:00Z"
    }
  ],
  "pagination": {
    "next_cursor": "...",
    "has_next": false,
    "has_prev": false
  }
}
```

#### 9.2.8 Get Conversation Messages

**Endpoint:** `GET /api/conversations/:id/messages`

**Query Parameters:**
- `limit` (integer, optional): Number of messages (default: 50)
- `cursor` (string, optional): UUID cursor
- `direction` (string, optional): `next` or `prev`

**Response:**
```json
{
  "data": [
    {
      "id": "018c2f1e-3b4a-7d1c-9f2e-3a4b5c6d7e8f",
      "conversation_id": "018c2f1e-3b4a-7d1c-9f2e-3a4b5c6d7e8e",
      "role": "user",
      "content": "What is your refund policy?",
      "raw_message": {
        "role": "user",
        "content": "What is your refund policy?"
      },
      "created_at": "2024-10-30T14:00:00Z"
    }
  ],
  "pagination": {
    "next_cursor": "...",
    "has_next": false,
    "has_prev": false
  }
}
```

### 9.3 Python Tool API Endpoints

#### 9.3.1 Find Similar Q&A

**Endpoint:** `POST /tools/find-similar`

**Request Body:**
```json
{
  "embedding": [0.1, 0.2, 0.3, ...],
  "top_k": 3
}
```

**Response:**
```json
{
  "results": [
    {
      "qa_pair": {
        "id": "018c2f1e-3b4a-7d1c-9f2e-3a4b5c6d7e8f",
        "question": "What is your refund policy?",
        "answer": "Refunds are processed within 5-7 business days.",
        "created_at": "2024-10-30T14:00:00Z",
        "updated_at": "2024-10-30T14:00:00Z"
      },
      "score": 0.95
    }
  ]
}
```

**cURL Example:**
```bash
curl -X POST "http://localhost:8080/tools/find-similar" \
  -H "Content-Type: application/json" \
  -d '{
    "embedding": [0.1, 0.2, 0.3],
    "top_k": 3
  }'
```

#### 9.3.2 Get Q&A by IDs

**Endpoint:** `POST /tools/get-qa-by-ids`

**Request Body:**
```json
{
  "ids": [
    "018c2f1e-3b4a-7d1c-9f2e-3a4b5c6d7e8f",
    "018c2f1e-3b4a-7d1c-9f2e-3a4b5c6d7e90"
  ]
}
```

**Response:**
```json
{
  "qa_pairs": [
    {
      "id": "018c2f1e-3b4a-7d1c-9f2e-3a4b5c6d7e8f",
      "question": "What is your refund policy?",
      "answer": "Refunds are processed within 5-7 business days.",
      "created_at": "2024-10-30T14:00:00Z",
      "updated_at": "2024-10-30T14:00:00Z"
    }
  ]
}
```

#### 9.3.3 Create Q&A with Embedding

**Endpoint:** `POST /tools/create-qa`

**Request Body:**
```json
{
  "question": "What is your shipping policy?",
  "answer": "We offer free shipping on orders over $50.",
  "embedding": [0.1, 0.2, 0.3, ...]
}
```

**Response:**
```json
{
  "qa_pair": {
    "id": "018c2f1e-3b4a-7d1c-9f2e-3a4b5c6d7e91",
    "question": "What is your shipping policy?",
    "answer": "We offer free shipping on orders over $50.",
    "created_at": "2024-10-30T14:00:00Z",
    "updated_at": "2024-10-30T14:00:00Z"
  }
}
```

#### 9.3.4 Search Q&A (Full-Text)

**Endpoint:** `POST /tools/search-qa`

**Request Body:**
```json
{
  "query": "refund",
  "limit": 10
}
```

**Response:**
```json
{
  "qa_pairs": [
    {
      "id": "018c2f1e-3b4a-7d1c-9f2e-3a4b5c6d7e8f",
      "question": "What is your refund policy?",
      "answer": "Refunds are processed within 5-7 business days.",
      "created_at": "2024-10-30T14:00:00Z",
      "updated_at": "2024-10-30T14:00:00Z"
    }
  ],
  "count": 1
}
```

#### 9.3.5 Save Message

**Endpoint:** `POST /tools/save-message`

**Request Body:**
```json
{
  "conversation_id": "018c2f1e-3b4a-7d1c-9f2e-3a4b5c6d7e8e",
  "role": "assistant",
  "content": "Based on our policy, refunds are processed within 5-7 business days.",
  "raw_message": {
    "role": "assistant",
    "content": "Based on our policy, refunds are processed within 5-7 business days."
  }
}
```

**Response:**
```json
{
  "message": {
    "id": "018c2f1e-3b4a-7d1c-9f2e-3a4b5c6d7e92",
    "conversation_id": "018c2f1e-3b4a-7d1c-9f2e-3a4b5c6d7e8e",
    "role": "assistant",
    "content": "Based on our policy, refunds are processed within 5-7 business days.",
    "raw_message": {
      "role": "assistant",
      "content": "Based on our policy, refunds are processed within 5-7 business days."
    },
    "created_at": "2024-10-30T14:01:00Z"
  }
}
```

---

## 10. Full-Text Search

### 10.1 PostgreSQL Full-Text Search Overview

PostgreSQL has built-in full-text search capabilities that enable intelligent text matching with stemming, ranking, and language-specific processing.

### 10.2 Key Concepts

**tsvector:** Document representation optimized for text search
```sql
SELECT to_tsvector('english', 'What is your refund policy?');
-- Result: 'polic':5 'refund':4
-- (stemmed words with positions)
```

**tsquery:** Search query representation
```sql
SELECT to_tsquery('english', 'refund');
-- Result: 'refund'
-- (automatically stems to match "refunds", "refunded", etc.)
```

**@@ operator:** Matches tsvector against tsquery
```sql
to_tsvector('english', text) @@ to_tsquery('english', 'refund')
```

### 10.3 Implementation

#### 10.3.1 Index Creation

```sql
CREATE INDEX idx_qa_fts ON qa_pairs 
USING gin(to_tsvector('english', question || ' ' || answer));
```

**GIN Index Benefits:**
- Fast lookups (milliseconds on millions of rows)
- Automatically maintained on INSERT/UPDATE
- Space-efficient compression

#### 10.3.2 Basic Search Query

```sql
SELECT id, question, answer
FROM qa_pairs
WHERE to_tsvector('english', question || ' ' || answer) 
      @@ to_tsquery('english', 'refund')
ORDER BY created_at DESC
LIMIT 10;
```

#### 10.3.3 Search with Ranking

```sql
SELECT 
    id, 
    question, 
    answer,
    ts_rank(
        to_tsvector('english', question || ' ' || answer),
        to_tsquery('english', 'refund')
    ) as relevance
FROM qa_pairs
WHERE to_tsvector('english', question || ' ' || answer) 
      @@ to_tsquery('english', 'refund')
ORDER BY relevance DESC, created_at DESC
LIMIT 10;
```

**Relevance Scoring:**
- Higher score = more relevant
- Based on term frequency and position
- Range: 0.0 to 1.0

#### 10.3.4 Go Implementation

```go
func (r *qaRepository) SearchFullText(ctx context.Context, searchQuery string, params models.CursorParams) ([]*models.QAPair, *models.CursorPagination, error) {
    query := `
        SELECT 
            id, question, answer, created_at, updated_at,
            ts_rank(
                to_tsvector('english', question || ' ' || answer),
                to_tsquery('english', $1)
            ) as rank
        FROM qa_pairs
        WHERE to_tsvector('english', question || ' ' || answer) 
              @@ to_tsquery('english', $1)
        ORDER BY rank DESC, id DESC
        LIMIT $2
    `
    
    var qaPairs []*models.QAPair
    err := r.db.SelectContext(ctx, &qaPairs, query, searchQuery, params.Limit)
    
    return qaPairs, nil, err
}
```

### 10.4 Search Features

#### 10.4.1 Stemming

Automatic word normalization:
- `refund` matches `refunds`, `refunded`, `refunding`
- `ship` matches `ships`, `shipping`, `shipped`
- `policy` matches `policies`

#### 10.4.2 Stop Words

Automatically filters common words:
- `the`, `is`, `and`, `or`, `a`, `an`
- These don't affect search results

#### 10.4.3 Multi-Word Search

```sql
-- AND operator (both words must appear)
to_tsquery('english', 'refund & policy')

-- OR operator (either word)
to_tsquery('english', 'refund | return')

-- NOT operator (exclude word)
to_tsquery('english', 'refund & !shipping')
```

### 10.5 Performance Characteristics

| Dataset Size | Query Time (with GIN index) |
|--------------|----------------------------|
| 1,000 rows   | < 1ms                      |
| 10,000 rows  | 1-5ms                      |
| 100,000 rows | 5-10ms                     |
| 1,000,000 rows | 10-20ms                  |

**Note:** Without index, query time grows linearly with dataset size (unacceptable for production).

---

## 11. Cursor Pagination

### 11.1 Cursor Pagination Overview

Cursor-based pagination uses the ID of the last item as a pointer to fetch the next batch. With UUIDv7, IDs are chronologically ordered, making cursor pagination simple and efficient.

### 11.2 Why Cursor Pagination?

**Problems with OFFSET/LIMIT:**
```sql
-- Page 1
SELECT * FROM qa_pairs ORDER BY id DESC LIMIT 10 OFFSET 0;
-- Fast (scans 10 rows)

-- Page 100
SELECT * FROM qa_pairs ORDER BY id DESC LIMIT 10 OFFSET 990;
-- Slow (scans 1000 rows, returns 10)
```

**Cursor-based approach:**
```sql
-- Page 1
SELECT * FROM qa_pairs ORDER BY id DESC LIMIT 10;
-- Fast (scans 10 rows)

-- Page 100 (using cursor)
SELECT * FROM qa_pairs WHERE id < 'cursor' ORDER BY id DESC LIMIT 10;
-- Still fast (index seek + scan 10 rows)
```

### 11.3 Implementation with UUIDv7

#### 11.3.1 Next Page

```sql
-- First page (no cursor)
SELECT * FROM qa_pairs 
ORDER BY id DESC 
LIMIT 10;

-- Returns: IDs ending with ...7e8f, ...7e90, ...7e91, ..., ...7e98
-- Next cursor: ...7e98 (last item)

-- Next page (with cursor)
SELECT * FROM qa_pairs 
WHERE id < '018c2f1e-3b4a-7d1c-9f2e-3a4b5c6d7e98'
ORDER BY id DESC 
LIMIT 10;

-- Returns: IDs ending with ...7e99, ...7e9a, ..., ...7ea0
```

#### 11.3.2 Previous Page

```sql
-- Previous page (reverse direction)
SELECT * FROM qa_pairs 
WHERE id > '018c2f1e-3b4a-7d1c-9f2e-3a4b5c6d7e99'
ORDER BY id ASC 
LIMIT 10;

-- Returns results in ASC order
-- Reverse in application code to maintain DESC order
```

### 11.4 Go Implementation

```go
func (r *qaRepository) List(ctx context.Context, params models.CursorParams) ([]*models.QAPair, *models.CursorPagination, error) {
    // Parse cursor
    var cursorID uuid.UUID
    var err error
    if params.Cursor != "" {
        cursorID, err = uuid.Parse(params.Cursor)
        if err != nil {
            return nil, nil, fmt.Errorf("invalid cursor: %w", err)
        }
    }
    
    // Build WHERE clause
    whereClause := ""
    args := []interface{}{}
    if params.Cursor != "" {
        if params.Direction == "prev" {
            whereClause = "WHERE id > $1"
        } else {
            whereClause = "WHERE id < $1"
        }
        args = append(args, cursorID)
    }
    
    // Order
    order := "DESC"
    if params.Direction == "prev" {
        order = "ASC"
    }
    
    // Fetch one extra to check if there's more
    fetchLimit := params.Limit + 1
    
    query := fmt.Sprintf(`
        SELECT id, question, answer, created_at, updated_at
        FROM qa_pairs
        %s
        ORDER BY id %s
        LIMIT %d
    `, whereClause, order, fetchLimit)
    
    var qaPairs []*models.QAPair
    if params.Cursor != "" {
        err = r.db.SelectContext(ctx, &qaPairs, query, args...)
    } else {
        err = r.db.SelectContext(ctx, &qaPairs, query)
    }
    
    if err != nil {
        return nil, nil, err
    }
    
    // Check if there are more results
    hasMore := len(qaPairs) > params.Limit
    if hasMore {
        qaPairs = qaPairs[:params.Limit]
    }
    
    // Reverse if prev direction
    if params.Direction == "prev" && len(qaPairs) > 0 {
        for i, j := 0, len(qaPairs)-1; i < j; i, j = i+1, j-1 {
            qaPairs[i], qaPairs[j] = qaPairs[j], qaPairs[i]
        }
    }
    
    // Build pagination metadata
    pagination := &models.CursorPagination{}
    if len(qaPairs) > 0 {
        pagination.NextCursor = qaPairs[len(qaPairs)-1].ID.String()
        pagination.PrevCursor = qaPairs[0].ID.String()
        pagination.HasNext = hasMore
        pagination.HasPrev = params.Cursor != ""
    }
    
    return qaPairs, pagination, nil
}
```

### 11.5 React Integration

```typescript
const [data, setData] = useState<QAPair[]>([]);
const [nextCursor, setNextCursor] = useState<string | null>(null);
const [prevCursor, setPrevCursor] = useState<string | null>(null);
const [hasNext, setHasNext] = useState(false);
const [hasPrev, setHasPrev] = useState(false);

const fetchData = async (cursor?: string, direction?: 'next' | 'prev') => {
  const params = new URLSearchParams({
    limit: '10',
    ...(cursor && { cursor }),
    ...(direction && { direction })
  });
  
  const response = await fetch(`/api/qa-pairs?${params}`);
  const result = await response.json();
  
  setData(result.data);
  setNextCursor(result.pagination.next_cursor);
  setPrevCursor(result.pagination.prev_cursor);
  setHasNext(result.pagination.has_next);
  setHasPrev(result.pagination.has_prev);
};

// Navigation
<button onClick={() => fetchData(prevCursor, 'prev')} disabled={!hasPrev}>
  Previous
</button>
<button onClick={() => fetchData(nextCursor, 'next')} disabled={!hasNext}>
  Next
</button>
```

### 11.6 Performance Comparison

| Method | Page 1 | Page 100 | Page 1000 |
|--------|--------|----------|-----------|
| OFFSET | 10ms | 100ms | 1000ms |
| Cursor | 10ms | 10ms | 10ms |

**Cursor pagination maintains constant performance regardless of page number.**

---

## 12. UUIDv7 Implementation

### 12.1 UUIDv7 Overview

UUIDv7 is a time-ordered UUID format that combines a timestamp with random data, making it ideal for distributed systems and cursor pagination.

**Structure:**
```
018c2f1e-3b4a-7d1c-9f2e-3a4b5c6d7e8f
└──┬───┘ └─┬─┘ └─┬─┘ └────┬──────────┘
  Timestamp   Random      Random
  (48 bits)   (12 bits)   (62 bits)
```

**First 48 bits:** Unix timestamp in milliseconds
- Provides chronological ordering
- Sortable by creation time
- ~8000 years before wrapping

### 12.2 Benefits

1. **Time-Ordered:** Natural chronological sorting
2. **Cursor Pagination:** Simple single-column cursors
3. **Distributed:** No coordination needed between servers
4. **B-Tree Friendly:** Sequential inserts (better than UUIDv4)
5. **Extractable Timestamp:** Can retrieve creation time if needed

### 12.3 Go Implementation

#### 12.3.1 Library

```go
import "github.com/google/uuid"

// Requires: github.com/google/uuid v1.4.0+
```

#### 12.3.2 Generation

```go
// Generate UUIDv7
id := uuid.Must(uuid.NewV7())

// Returns: uuid.UUID type
// Example: 018c2f1e-3b4a-7d1c-9f2e-3a4b5c6d7e8f
```

#### 12.3.3 Usage in Repository

```go
func (r *qaRepository) Create(ctx context.Context, qa *models.QAPair) error {
    // Generate UUIDv7
    qa.ID = uuid.Must(uuid.NewV7())
    
    query := `
        INSERT INTO qa_pairs (id, question, answer)
        VALUES ($1, $2, $3)
        RETURNING created_at, updated_at
    `
    
    return r.db.GetContext(ctx, qa, query, qa.ID, qa.Question, qa.Answer)
}
```

#### 12.3.4 Parsing

```go
// Parse string to UUID
id, err := uuid.Parse("018c2f1e-3b4a-7d1c-9f2e-3a4b5c6d7e8f")
if err != nil {
    // Invalid UUID
}

// Convert to string
idStr := id.String()
```

#### 12.3.5 Extracting Timestamp

```go
func ExtractTimestamp(id uuid.UUID) time.Time {
    // First 6 bytes contain timestamp
    timestampBytes := id[:6]
    
    // Convert to Unix milliseconds
    var timestamp uint64
    for _, b := range timestampBytes {
        timestamp = (timestamp << 8) | uint64(b)
    }
    
    return time.UnixMilli(int64(timestamp))
}

// Usage
id := uuid.Must(uuid.NewV7())
createdAt := ExtractTimestamp(id)
```

### 12.4 Database Configuration

```sql
-- PostgreSQL supports UUID type natively
CREATE TABLE qa_pairs (
    id UUID PRIMARY KEY,  -- No DEFAULT needed (generated in Go)
    question TEXT NOT NULL,
    answer TEXT NOT NULL
);

-- Index for cursor pagination (DESC order)
CREATE INDEX idx_qa_id_desc ON qa_pairs(id DESC);
```

**Note:** We generate UUIDs in the Go application rather than in PostgreSQL because `gen_random_uuid()` generates UUIDv4 (random), not UUIDv7 (time-ordered).

### 12.5 Migration from SERIAL

If migrating from `SERIAL` IDs:

```sql
-- Old schema
CREATE TABLE qa_pairs (
    id SERIAL PRIMARY KEY,
    ...
);

-- New schema
CREATE TABLE qa_pairs (
    id UUID PRIMARY KEY,
    ...
);

-- Migration strategy
-- 1. Create new UUID column
ALTER TABLE qa_pairs ADD COLUMN uuid_id UUID;

-- 2. Generate UUIDs for existing rows (in application)
UPDATE qa_pairs SET uuid_id = generate_uuidv7() WHERE uuid_id IS NULL;

-- 3. Switch primary key
ALTER TABLE qa_pairs DROP CONSTRAINT qa_pairs_pkey;
ALTER TABLE qa_pairs ADD PRIMARY KEY (uuid_id);
ALTER TABLE qa_pairs DROP COLUMN id;
ALTER TABLE qa_pairs RENAME COLUMN uuid_id TO id;
```

### 12.6 Comparison with Other ID Strategies

| Strategy | Sortable | Distributed | Size | Readability |
|----------|----------|-------------|------|-------------|
| SERIAL | ✅ | ❌ | 4-8 bytes | ⭐⭐⭐⭐⭐ |
| UUIDv4 | ❌ | ✅ | 16 bytes | ⭐⭐ |
| UUIDv7 | ✅ | ✅ | 16 bytes | ⭐⭐ |
| ULID | ✅ | ✅ | 16 bytes | ⭐⭐⭐ |

**UUIDv7 is the best choice for modern distributed systems with pagination requirements.**

---

## 13. Pinecone Integration

### 13.1 Pinecone Overview

Pinecone is a managed vector database optimized for similarity search and recommendations. It stores embeddings and enables fast nearest-neighbor queries.

### 13.2 Client Interface

```go
package clients

import (
    "context"
)

// PineconeMatch represents a similarity search result
type PineconeMatch struct {
    ID       string                 `json:"id"`
    Score    float32                `json:"score"`
    Metadata map[string]interface{} `json:"metadata"`
}

// PineconeClient defines vector database operations
type PineconeClient interface {
    // Upsert inserts or updates a vector
    Upsert(ctx context.Context, id string, vector []float32, metadata map[string]interface{}) error
    
    // Query performs similarity search
    Query(ctx context.Context, vector []float32, topK int) ([]PineconeMatch, error)
    
    // Delete removes a vector by ID
    Delete(ctx context.Context, id string) error
}
```

### 13.3 Client Implementation

```go
package clients

import (
    "context"
    "fmt"
    "github.com/pinecone-io/go-pinecone/pinecone"
)

type pineconeClient struct {
    client *pinecone.Client
    index  *pinecone.Index
}

// NewPineconeClient creates a new Pinecone client
func NewPineconeClient(apiKey, environment, indexName string) (PineconeClient, error) {
    client, err := pinecone.NewClient(pinecone.NewClientParams{
        ApiKey:      apiKey,
        Environment: environment,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create Pinecone client: %w", err)
    }
    
    index := client.Index(indexName)
    
    return &pineconeClient{
        client: client,
        index:  index,
    }, nil
}

// Upsert inserts or updates a vector
func (c *pineconeClient) Upsert(ctx context.Context, id string, vector []float32, metadata map[string]interface{}) error {
    vectors := []pinecone.Vector{
        {
            ID:       id,
            Values:   vector,
            Metadata: metadata,
        },
    }
    
    _, err := c.index.UpsertVectors(ctx, vectors)
    if err != nil {
        return fmt.Errorf("failed to upsert vector: %w", err)
    }
    
    return nil
}

// Query performs similarity search
func (c *pineconeClient) Query(ctx context.Context, vector []float32, topK int) ([]PineconeMatch, error) {
    results, err := c.index.Query(ctx, &pinecone.QueryRequest{
        Vector:          vector,
        TopK:            uint32(topK),
        IncludeMetadata: true,
    })
    
    if err != nil {
        return nil, fmt.Errorf("failed to query vectors: %w", err)
    }
    
    matches := make([]PineconeMatch, 0, len(results.Matches))
    for _, match := range results.Matches {
        matches = append(matches, PineconeMatch{
            ID:       match.ID,
            Score:    match.Score,
            Metadata: match.Metadata,
        })
    }
    
    return matches, nil
}

// Delete removes a vector
func (c *pineconeClient) Delete(ctx context.Context, id string) error {
    err := c.index.DeleteVectors(ctx, []string{id})
    if err != nil {
        return fmt.Errorf("failed to delete vector: %w", err)
    }
    
    return nil
}
```

### 13.4 Index Configuration

**Create Index (via Pinecone Console or API):**
```python
import pinecone

pinecone.create_index(
    name="qa-embeddings",
    dimension=1536,  # For OpenAI text-embedding-ada-002
    metric="cosine",
    pods=1,
    pod_type="p1.x1"
)
```

**Configuration:**
- **Dimension:** Must match embedding model (1536 for OpenAI, 384 for MiniLM)
- **Metric:** `cosine` for semantic similarity
- **Pods:** Start with 1, scale as needed

### 13.5 Usage in Service Layer

```go
// FindSimilar finds similar Q&A pairs using vector search
func (s *qaService) FindSimilar(ctx context.Context, embedding []float32, topK int) ([]models.SimilarityMatch, error) {
    // Query Pinecone
    matches, err := s.pinecone.Query(ctx, embedding, topK)
    if err != nil {
        return nil, fmt.Errorf("pinecone query failed: %w", err)
    }
    
    // Extract IDs
    ids := make([]uuid.UUID, 0, len(matches))
    scoreMap := make(map[uuid.UUID]float32)
    
    for _, match := range matches {
        id, err := uuid.Parse(match.ID)
        if err != nil {
            continue
        }
        ids = append(ids, id)
        scoreMap[id] = match.Score
    }
    
    // Fetch Q&A pairs from PostgreSQL
    qaPairs, err := s.qaRepo.GetByIDs(ctx, ids)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch Q&A pairs: %w", err)
    }
    
    // Build similarity matches
    results := make([]models.SimilarityMatch, 0, len(qaPairs))
    for _, qa := range qaPairs {
        results = append(results, models.SimilarityMatch{
            QAPair: *qa,
            Score:  scoreMap[qa.ID],
        })
    }
    
    return results, nil
}
```

### 13.6 Best Practices

1. **Batch Operations:** Use batch upsert for multiple vectors
2. **Error Handling:** Implement retries for transient failures
3. **Metadata:** Store minimal metadata (ID, essential fields)
4. **Consistency:** Keep Pinecone and PostgreSQL in sync
5. **Monitoring:** Track query latency and error rates

---

## 14. Configuration

### 14.1 Environment Variables

```bash
# Server Configuration
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
SERVER_ENVIRONMENT=development

# PostgreSQL Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=smart_discovery
DB_SSLMODE=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5

# Pinecone Configuration
PINECONE_API_KEY=your-api-key
PINECONE_ENVIRONMENT=us-west1-gcp
PINECONE_INDEX_NAME=qa-embeddings
PINECONE_DIMENSION=1536

# Embedding Configuration
EMBEDDING_DIMENSION=1536
EMBEDDING_MODEL=text-embedding-ada-002

# CORS Configuration
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Content-Type,Authorization

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

### 14.2 Configuration Loading

```go
package config

import (
    "fmt"
    "github.com/spf13/viper"
)

// LoadConfig loads configuration from environment variables
func LoadConfig() (*models.Config, error) {
    viper.SetConfigFile(".env")
    viper.AutomaticEnv()
    
    // Attempt to read config file (optional)
    _ = viper.ReadInConfig()
    
    config := &models.Config{
        Server: models.ServerConfig{
            Port:        viper.GetInt("SERVER_PORT"),
            Host:        viper.GetString("SERVER_HOST"),
            Environment: viper.GetString("SERVER_ENVIRONMENT"),
        },
        Database: models.DatabaseConfig{
            Host:         viper.GetString("DB_HOST"),
            Port:         viper.GetInt("DB_PORT"),
            User:         viper.GetString("DB_USER"),
            Password:     viper.GetString("DB_PASSWORD"),
            DBName:       viper.GetString("DB_NAME"),
            SSLMode:      viper.GetString("DB_SSLMODE"),
            MaxOpenConns: viper.GetInt("DB_MAX_OPEN_CONNS"),
            MaxIdleConns: viper.GetInt("DB_MAX_IDLE_CONNS"),
        },
        Pinecone: models.PineconeConfig{
            APIKey:      viper.GetString("PINECONE_API_KEY"),
            Environment: viper.GetString("PINECONE_ENVIRONMENT"),
            IndexName:   viper.GetString("PINECONE_INDEX_NAME"),
            Dimension:   viper.GetInt("PINECONE_DIMENSION"),
        },
        Embedding: models.EmbeddingConfig{
            Dimension: viper.GetInt("EMBEDDING_DIMENSION"),
            Model:     viper.GetString("EMBEDDING_MODEL"),
        },
    }
    
    // Validate configuration
    if err := validateConfig(config); err != nil {
        return nil, fmt.Errorf("configuration validation failed: %w", err)
    }
    
    return config, nil
}

func validateConfig(config *models.Config) error {
    // Add validation logic
    if config.Server.Port < 1 || config.Server.Port > 65535 {
        return fmt.Errorf("invalid server port: %d", config.Server.Port)
    }
    
    if config.Database.Host == "" {
        return fmt.Errorf("database host is required")
    }
    
    if config.Pinecone.APIKey == "" {
        return fmt.Errorf("Pinecone API key is required")
    }
    
    return nil
}
```

### 14.3 .env.example

```bash
# Copy this file to .env and update with your values

# Server Configuration
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
SERVER_ENVIRONMENT=development

# PostgreSQL Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your-password-here
DB_NAME=smart_discovery
DB_SSLMODE=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5

# Pinecone Configuration
PINECONE_API_KEY=your-pinecone-api-key
PINECONE_ENVIRONMENT=us-west1-gcp
PINECONE_INDEX_NAME=qa-embeddings
PINECONE_DIMENSION=1536

# Embedding Configuration
EMBEDDING_DIMENSION=1536
EMBEDDING_MODEL=text-embedding-ada-002

# CORS Configuration
CORS_ALLOWED_ORIGINS=http://localhost:3000
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Content-Type,Authorization

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

---

## 15. Error Handling

### 15.1 Error Types

```go
package models

import "errors"

var (
    ErrNotFound          = errors.New("resource not found")
    ErrValidation        = errors.New("validation error")
    ErrDatabase          = errors.New("database error")
    ErrPinecone          = errors.New("pinecone error")
    ErrInternal          = errors.New("internal server error")
    ErrUnauthorized      = errors.New("unauthorized")
    ErrForbidden         = errors.New("forbidden")
    ErrBadRequest        = errors.New("bad request")
    ErrConflict          = errors.New("conflict")
)
```

### 15.2 Error Response Format

```go
// ErrorResponse represents a standardized error response
type ErrorResponse struct {
    Error   string                 `json:"error"`
    Code    string                 `json:"code"`
    Message string                 `json:"message"`
    Details map[string]interface{} `json:"details,omitempty"`
}

// Example JSON response
{
  "error": "error",
  "code": "VALIDATION_ERROR",
  "message": "Question cannot be empty",
  "details": {
    "field": "question",
    "constraint": "required"
  }
}
```

### 15.3 Error Handling Middleware

```go
package middleware

import (
    "database/sql"
    "errors"
    "github.com/gin-gonic/gin"
    "net/http"
    "smart-company-discovery/internal/models"
)

// ErrorHandler handles errors and returns standardized responses
func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()
        
        // Check if there were any errors
        if len(c.Errors) > 0 {
            err := c.Errors.Last().Err
            
            var statusCode int
            var errorCode string
            var message string
            
            // Determine error type
            switch {
            case errors.Is(err, models.ErrNotFound) || errors.Is(err, sql.ErrNoRows):
                statusCode = http.StatusNotFound
                errorCode = models.ErrCodeNotFound
                message = "Resource not found"
                
            case errors.Is(err, models.ErrValidation):
                statusCode = http.StatusBadRequest
                errorCode = models.ErrCodeValidation
                message = err.Error()
                
            case errors.Is(err, models.ErrDatabase):
                statusCode = http.StatusInternalServerError
                errorCode = models.ErrCodeDatabaseError
                message = "Database operation failed"
                
            case errors.Is(err, models.ErrPinecone):
                statusCode = http.StatusInternalServerError
                errorCode = models.ErrCodePineconeError
                message = "Vector database operation failed"
                
            case errors.Is(err, models.ErrUnauthorized):
                statusCode = http.StatusUnauthorized
                errorCode = models.ErrCodeUnauthorized
                message = "Authentication required"
                
            case errors.Is(err, models.ErrForbidden):
                statusCode = http.StatusForbidden
                errorCode = models.ErrCodeForbidden
                message = "Access denied"
                
            default:
                statusCode = http.StatusInternalServerError
                errorCode = models.ErrCodeInternal
                message = "Internal server error"
            }
            
            errorResponse := models.ErrorResponse{
                Error:   "error",
                Code:    errorCode,
                Message: message,
            }
            
            c.JSON(statusCode, errorResponse)
        }
    }
}
```

### 15.4 Handler Error Handling

```go
func (h *QAHandler) GetQA(c *gin.Context) {
    idStr := c.Param("id")
    
    id, err := uuid.Parse(idStr)
    if err != nil {
        c.Error(models.ErrValidation)
        return
    }
    
    qa, err := h.qaService.GetQA(c.Request.Context(), id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            c.Error(models.ErrNotFound)
        } else {
            c.Error(models.ErrInternal)
        }
        return
    }
    
    c.JSON(http.StatusOK, qa)
}
```

---

## 16. Middleware

### 16.1 CORS Middleware

```go
package middleware

import (
    "github.com/gin-gonic/gin"
)

// CORS middleware for handling cross-origin requests
func CORS(allowedOrigins []string) gin.HandlerFunc {
    return func(c *gin.Context) {
        origin := c.Request.Header.Get("Origin")
        
        // Check if origin is allowed
        allowed := false
        for _, allowedOrigin := range allowedOrigins {
            if origin == allowedOrigin {
                allowed = true
                break
            }
        }
        
        if allowed {
            c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
            c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
            c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
            c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        }
        
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        
        c.Next()
    }
}
```

### 16.2 Logging Middleware

```go
package middleware

import (
    "github.com/gin-gonic/gin"
    "github.com/rs/zerolog/log"
    "time"
)

// Logger middleware for request logging
func Logger() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        query := c.Request.URL.RawQuery
        
        c.Next()
        
        end := time.Now()
        latency := end.Sub(start)
        
        log.Info().
            Str("method", c.Request.Method).
            Str("path", path).
            Str("query", query).
            Int("status", c.Writer.Status()).
            Dur("latency", latency).
            Str("ip", c.ClientIP()).
            Str("user-agent", c.Request.UserAgent()).
            Msg("HTTP request")
    }
}
```

### 16.3 Recovery Middleware

```go
package middleware

import (
    "github.com/gin-gonic/gin"
    "github.com/rs/zerolog/log"
    "net/http"
)

// Recovery middleware for panic recovery
func Recovery() gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                log.Error().
                    Interface("error", err).
                    Str("path", c.Request.URL.Path).
                    Msg("Panic recovered")
                
                c.JSON(http.StatusInternalServerError, gin.H{
                    "error":   "error",
                    "code":    "INTERNAL_ERROR",
                    "message": "Internal server error",
                })
                c.Abort()
            }
        }()
        
        c.Next()
    }
}
```

### 16.4 Validation Middleware

```go
package middleware

import (
    "github.com/gin-gonic/gin"
    "github.com/go-playground/validator/v10"
    "net/http"
)

var validate = validator.New()

// ValidateRequest validates request body against struct tags
func ValidateRequest(obj interface{}) gin.HandlerFunc {
    return func(c *gin.Context) {
        if err := c.ShouldBindJSON(obj); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{
                "error":   "error",
                "code":    "VALIDATION_ERROR",
                "message": "Invalid request body",
                "details": err.Error(),
            })
            c.Abort()
            return
        }
        
        if err := validate.Struct(obj); err != nil {
            validationErrors := err.(validator.ValidationErrors)
            errors := make([]map[string]string, 0, len(validationErrors))
            
            for _, fieldError := range validationErrors {
                errors = append(errors, map[string]string{
                    "field":   fieldError.Field(),
                    "message": fieldError.Tag(),
                })
            }
            
            c.JSON(http.StatusBadRequest, gin.H{
                "error":   "error",
                "code":    "VALIDATION_ERROR",
                "message": "Validation failed",
                "details": errors,
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

---

## 17. Dependencies

### 17.1 go.mod

```go
module smart-company-discovery

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/google/uuid v1.4.0
    github.com/jmoiron/sqlx v1.3.5
    github.com/lib/pq v1.10.9
    github.com/pinecone-io/go-pinecone v0.1.0
    github.com/spf13/viper v1.17.0
    github.com/rs/zerolog v1.31.0
    github.com/go-playground/validator/v10 v10.16.0
)
```

### 17.2 Installation Commands

```bash
# Initialize Go module
go mod init smart-company-discovery

# Install core dependencies
go get github.com/gin-gonic/gin@v1.9.1
go get github.com/google/uuid@v1.4.0
go get github.com/jmoiron/sqlx@v1.3.5
go get github.com/lib/pq@v1.10.9

# Install Pinecone client
go get github.com/pinecone-io/go-pinecone

# Install configuration and logging
go get github.com/spf13/viper@v1.17.0
go get github.com/rs/zerolog@v1.31.0

# Install validation
go get github.com/go-playground/validator/v10@v10.16.0

# Download dependencies
go mod download

# Verify dependencies
go mod verify

# Tidy up dependencies
go mod tidy
```

### 17.3 Dependency Descriptions

| Package | Purpose | Version |
|---------|---------|---------|
| `gin-gonic/gin` | HTTP web framework | v1.9.1 |
| `google/uuid` | UUIDv7 generation | v1.4.0+ |
| `jmoiron/sqlx` | PostgreSQL database access | v1.3.5 |
| `lib/pq` | PostgreSQL driver | v1.10.9 |
| `pinecone-io/go-pinecone` | Pinecone vector database client | latest |
| `spf13/viper` | Configuration management | v1.17.0 |
| `rs/zerolog` | Structured logging | v1.31.0 |
| `go-playground/validator` | Struct validation | v10.16.0 |

---

## Appendix A: Quick Start Guide

### A.1 Prerequisites

- Go 1.21 or higher
- PostgreSQL 12 or higher
- Pinecone account and API key

### A.2 Setup Steps

```bash
# 1. Clone repository
git clone <repository-url>
cd smart-company-discovery

# 2. Install dependencies
go mod download

# 3. Configure environment
cp .env.example .env
# Edit .env with your values

# 4. Run database migrations
psql -U postgres -f migrations/001_init_schema.sql

# 5. Run the server
go run cmd/server/main.go
```

### A.3 Verify Installation

```bash
# Health check
curl http://localhost:8080/health

# Create Q&A
curl -X POST http://localhost:8080/api/qa-pairs \
  -H "Content-Type: application/json" \
  -d '{
    "question": "Test question",
    "answer": "Test answer"
  }'

# List Q&A
curl http://localhost:8080/api/qa-pairs
```

---

## Appendix B: Common Operations

### B.1 Database Operations

```bash
# Connect to PostgreSQL
psql -U postgres -d smart_discovery

# View Q&A pairs
SELECT id, question, created_at FROM qa_pairs ORDER BY id DESC LIMIT 10;

# Full-text search test
SELECT question, answer 
FROM qa_pairs 
WHERE to_tsvector('english', question || ' ' || answer) @@ to_tsquery('english', 'refund')
LIMIT 5;

# Check indexes
\d+ qa_pairs
```

### B.2 Pinecone Operations

```python
import pinecone

# Initialize
pinecone.init(api_key="your-key", environment="us-west1-gcp")

# Check index stats
index = pinecone.Index("qa-embeddings")
print(index.describe_index_stats())

# Query
results = index.query(vector=[...], top_k=3, include_metadata=True)
```

---

## Appendix C: Troubleshooting

### C.1 Common Issues

**Issue: UUIDs not sorting chronologically**
- Solution: Ensure you're using `uuid.NewV7()` not `uuid.New()` (which generates UUIDv4)

**Issue: Full-text search not working**
- Solution: Verify GIN index exists: `\d+ qa_pairs`
- Recreate index: `CREATE INDEX idx_qa_fts ON qa_pairs USING gin(...)`

**Issue: Cursor pagination returning duplicates**
- Solution: Ensure index on `id DESC` exists for performance

**Issue: Pinecone connection fails**
- Solution: Verify API key and environment in `.env`
- Check Pinecone console for index status

---

## Document Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2024-10-30 | System | Initial comprehensive design document |

---

**End of Document**
