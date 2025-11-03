# Environment Configuration

## Required Environment Variables

Create these environment variables in your deployment environment or `.env` file:

### Server Configuration
```bash
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
SERVER_ENVIRONMENT=development
```

### Database Configuration
```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=smart_discovery
DB_SSLMODE=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
```

### Pinecone Configuration
```bash
# Get your API key from: https://app.pinecone.io/
PINECONE_API_KEY=your-pinecone-api-key-here
PINECONE_ENVIRONMENT=your-environment  # e.g., us-west1-gcp, us-east-1-aws
PINECONE_INDEX_NAME=qa-index
PINECONE_NAMESPACE=
```

### Google Embedding Configuration
```bash
# Get your API key from: https://console.cloud.google.com/
GOOGLE_API_KEY=your-google-api-key-here
GOOGLE_PROJECT_ID=your-google-project-id
GOOGLE_LOCATION=us-central1
GOOGLE_EMBEDDING_MODEL=text-embedding-004
```

## Quick Setup Guide

### 1. Pinecone Setup

1. Sign up at https://app.pinecone.io/
2. Create a new index with these settings:
   - **Name**: `qa-index`
   - **Dimensions**: 768
   - **Metric**: Cosine
3. Copy your API key and environment name

### 2. Google Cloud Setup

1. Create a project at https://console.cloud.google.com/
2. Enable the Vertex AI API
3. Create an API key in **APIs & Services** > **Credentials**
4. Note your Project ID

### 3. Set Environment Variables

#### Linux/Mac
```bash
export PINECONE_API_KEY="your-key"
export PINECONE_ENVIRONMENT="us-west1-gcp"
export PINECONE_INDEX_NAME="qa-index"
export GOOGLE_API_KEY="your-key"
export GOOGLE_PROJECT_ID="your-project"
```

#### Windows (PowerShell)
```powershell
$env:PINECONE_API_KEY="your-key"
$env:PINECONE_ENVIRONMENT="us-west1-gcp"
$env:PINECONE_INDEX_NAME="qa-index"
$env:GOOGLE_API_KEY="your-key"
$env:GOOGLE_PROJECT_ID="your-project"
```

### 4. Start the Server

```bash
cd backend
go run cmd/server/main.go
```

You should see:
```
✓ Successfully connected to PostgreSQL database
✓ Successfully initialized Google Embedding client
✓ Successfully initialized Pinecone client
🚀 Server starting on http://0.0.0.0:8080
```

## Development Mode

For development without API keys, the system will automatically use mock clients:

```
ℹ Google Embedding not configured. Using mock embedding client.
ℹ Pinecone not configured. Using mock Pinecone client.
```

This allows you to develop and test without incurring API costs.

