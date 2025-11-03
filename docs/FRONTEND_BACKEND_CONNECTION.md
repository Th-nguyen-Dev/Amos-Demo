# Frontend-Backend Connection Guide

## 🎯 Overview

This guide explains how to connect and run the frontend (React) and backend (Go) together.

## 🏗️ Architecture

```
┌─────────────────┐         ┌─────────────────┐         ┌──────────────┐
│   Frontend      │         │   Backend       │         │  PostgreSQL  │
│   (React/Vite)  │────────▶│   (Go/Gin)      │────────▶│  Database    │
│   Port: 5173    │  HTTP   │   Port: 8080    │         │  Port: 5432  │
└─────────────────┘         └─────────────────┘         └──────────────┘
                                     │
                                     ▼
                            ┌─────────────────┐
                            │  Pinecone       │
                            │  Vector DB      │
                            │  Port: 5081     │
                            └─────────────────┘
```

## 🚀 Quick Start (Both Services)

### Option 1: Using Docker + Local Frontend (Recommended for Development)

This runs the backend in Docker while you develop the frontend locally with hot-reload:

```bash
# Terminal 1: Start backend services (PostgreSQL + Pinecone + Backend API)
docker-compose up -d

# Terminal 2: Start frontend dev server
cd frontend
npm install
npm run dev
```

**Access:**
- Frontend: http://localhost:5173
- Backend API: http://localhost:8080
- API Health: http://localhost:8080/health

### Option 2: Everything Local (No Docker)

**Prerequisites:**
- Go 1.24+
- Node.js 18+
- PostgreSQL running locally
- (Optional) Pinecone Local running

```bash
# Terminal 1: Start backend
cd backend
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=smart_discovery
go run cmd/server/main.go

# Terminal 2: Start frontend
cd frontend
npm install
npm run dev
```

### Option 3: Everything in Docker

Add the frontend to `docker-compose.yml` (see section below).

---

## 🔧 Frontend Configuration

### Environment Variables

The frontend uses Vite environment variables. Create `.env` in the `frontend/` directory:

```env
# Frontend Environment Variables
VITE_API_BASE_URL=http://localhost:8080
```

**Default behavior (no .env needed):**
- If `VITE_API_BASE_URL` is not set, it defaults to `http://localhost:8080`
- This works perfectly for local development

### How the Frontend Connects

The frontend API files are configured to connect to the backend:

**Chat API** (`frontend/src/features/chat/api/chatApi.ts`):
```typescript
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
```

**QA API** (`frontend/src/features/qa/api/qaApi.ts`):
```typescript
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
```

### Vite Proxy Configuration

The frontend has a proxy configured in `vite.config.ts`:

```typescript
server: {
  port: 5173,
  proxy: {
    '/api': {
      target: 'http://localhost:8080',
      changeOrigin: true,
    },
  },
}
```

**What this means:**
- Frontend requests to `/api/*` are proxied to `http://localhost:8080/api/*`
- Helps avoid CORS issues during development
- The proxy is ONLY used in development mode (`npm run dev`)

---

## 🔧 Backend Configuration

### CORS Configuration

The backend already has CORS enabled in `backend/internal/api/middleware/cors.go`:

```go
func CORS() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        // ...
    }
}
```

**Current configuration:**
- ✅ Allows all origins (`*`)
- ✅ Allows required headers
- ✅ Handles preflight OPTIONS requests

**For production**, you should restrict origins:
```go
c.Writer.Header().Set("Access-Control-Allow-Origin", "https://yourdomain.com")
```

### Backend Endpoints

The backend exposes these endpoints on port 8080:

**Health Check:**
- `GET /health`

**Q&A Management:**
- `GET /api/qa-pairs` - List Q&A pairs
- `GET /api/qa-pairs/:id` - Get single Q&A pair
- `POST /api/qa-pairs` - Create Q&A pair
- `PUT /api/qa-pairs/:id` - Update Q&A pair
- `DELETE /api/qa-pairs/:id` - Delete Q&A pair

**Conversations:**
- `POST /api/conversations` - Create conversation
- `GET /api/conversations` - List conversations
- `GET /api/conversations/:id` - Get conversation
- `POST /api/conversations/:id/messages` - Add message
- `GET /api/conversations/:id/messages` - Get messages

**Python Agent Tools:**
- `POST /tools/search-qa` - Search Q&A pairs
- `POST /tools/get-qa-by-ids` - Get Q&A pairs by IDs
- `POST /tools/save-message` - Save conversation message

---

## 🐳 Adding Frontend to Docker Compose

To run the frontend in Docker alongside the backend, add this service to `docker-compose.yml`:

```yaml
  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    container_name: smart-discovery-frontend
    environment:
      - VITE_API_BASE_URL=http://localhost:8080
    ports:
      - "5173:5173"
    volumes:
      - ./frontend:/app
      - /app/node_modules
    depends_on:
      - backend
    restart: unless-stopped
```

**Then create** `frontend/Dockerfile`:

```dockerfile
FROM node:20-alpine

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci

# Copy source code
COPY . .

# Expose Vite dev server port
EXPOSE 5173

# Start development server
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]
```

---

## 🧪 Testing the Connection

### 1. Test Backend is Running

```bash
# Health check
curl http://localhost:8080/health

# Expected response:
# {"status":"healthy","database":"connected"}
```

### 2. Test Backend API

```bash
# List Q&A pairs
curl http://localhost:8080/api/qa-pairs

# Expected response:
# {"data":[...],"pagination":{...}}
```

### 3. Test Frontend

1. Open browser: http://localhost:5173
2. Navigate to "Q&A Management" page
3. You should see the list of Q&A pairs
4. Try creating a new Q&A pair

### 4. Check Browser Console

Open DevTools (F12) and look for:
- ✅ No CORS errors
- ✅ Successful API requests to `http://localhost:8080/api/*`
- ✅ 200 OK responses

---

## 🔍 Troubleshooting

### Frontend can't connect to backend

**Symptom:** Network errors, CORS errors, or "Failed to fetch"

**Solutions:**

1. **Check backend is running:**
   ```bash
   curl http://localhost:8080/health
   ```

2. **Check CORS headers:**
   ```bash
   curl -I http://localhost:8080/api/qa-pairs
   # Look for: Access-Control-Allow-Origin: *
   ```

3. **Check frontend environment:**
   ```bash
   cd frontend
   echo $VITE_API_BASE_URL  # Should be http://localhost:8080 or empty
   ```

4. **Clear browser cache** and hard reload (Ctrl+Shift+R)

### CORS errors in browser

**Symptom:** "CORS policy: No 'Access-Control-Allow-Origin' header"

**Solutions:**

1. Make sure backend CORS middleware is applied in `backend/cmd/server/main.go`:
   ```go
   router.Use(middleware.CORS())
   ```

2. Restart the backend service:
   ```bash
   docker-compose restart backend
   ```

### Frontend shows wrong API URL

**Check the environment:**

```bash
# In frontend directory
cat .env

# Should show:
# VITE_API_BASE_URL=http://localhost:8080
```

**Note:** Vite only loads `.env` files at build time. After changing `.env`, restart the dev server:
```bash
# Stop with Ctrl+C, then:
npm run dev
```

### Port already in use

**Backend (8080):**
```bash
# Find process using port 8080
lsof -i :8080

# Kill process
kill -9 <PID>
```

**Frontend (5173):**
```bash
# Find process using port 5173
lsof -i :5173

# Or change port in vite.config.ts:
server: {
  port: 3000,  // Use different port
}
```

### Database connection errors

**Check PostgreSQL is running:**
```bash
docker-compose ps postgres

# Should show "Up" status
```

**Check database logs:**
```bash
docker-compose logs postgres
```

**Reset database:**
```bash
docker-compose down -v  # Remove volumes
docker-compose up -d    # Recreate with fresh data
```

---

## 🌐 Production Deployment

### Environment-Specific Configuration

**Development:**
```env
# Frontend .env.development
VITE_API_BASE_URL=http://localhost:8080
```

**Production:**
```env
# Frontend .env.production
VITE_API_BASE_URL=https://api.yourdomain.com
```

### Build Frontend for Production

```bash
cd frontend
npm run build

# Output in: frontend/dist/
# Serve with nginx, Apache, or any static server
```

### Update Backend CORS for Production

In `backend/internal/api/middleware/cors.go`:

```go
func CORS() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Only allow your frontend domain
        c.Writer.Header().Set("Access-Control-Allow-Origin", "https://yourdomain.com")
        c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
        // ... rest of the code
    }
}
```

---

## 📝 Quick Reference

| Service | Port | URL | Status |
|---------|------|-----|--------|
| Frontend | 5173 | http://localhost:5173 | ✅ Configured |
| Backend | 8080 | http://localhost:8080 | ✅ Configured |
| PostgreSQL | 5432 | localhost:5432 | ✅ Via Docker |
| Pinecone Local | 5081 | http://localhost:5081 | ✅ Via Docker |
| Python Agent | 8000 | http://localhost:8000 | ✅ Via Docker |

**Connection Status:**
- ✅ Frontend → Backend: Configured and working
- ✅ Backend → Database: Configured and working
- ✅ Backend → Pinecone: Configured and working
- ✅ Python Agent → Backend: Configured and working
- ✅ CORS: Enabled on backend

---

## 🎓 Understanding the Connection Flow

### 1. User Action in Browser
```
User clicks "Create Q&A" button in React app
↓
```

### 2. Frontend API Call
```typescript
// frontend/src/features/qa/api/qaApi.ts
const response = await fetch('http://localhost:8080/api/qa-pairs', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ question, answer, category }),
})
```

### 3. Vite Proxy (Development Only)
```
Request: http://localhost:5173/api/qa-pairs
↓ (proxied to)
↓ http://localhost:8080/api/qa-pairs
```

### 4. Backend Receives Request
```go
// backend/cmd/server/main.go
api.POST("/qa-pairs", qaHandler.CreateQA)
↓
```

### 5. Backend Processes
```
1. Validate request
2. Save to PostgreSQL
3. Generate embedding
4. Index in Pinecone
5. Return response
```

### 6. Frontend Updates UI
```typescript
// React component updates with new data
// Redux state is updated
// UI re-renders automatically
```

---

## 🎉 You're All Set!

The frontend and backend are already configured to work together. Just run:

```bash
# Terminal 1: Backend
docker-compose up -d

# Terminal 2: Frontend
cd frontend && npm install && npm run dev
```

Then open http://localhost:5173 in your browser!

---

## 📚 Related Documentation

- [Main README](./README.md) - Project overview
- [Docker Quick Start](./DOCKER_QUICKSTART.md) - Docker setup
- [Local Development](./docs/LOCAL_DEVELOPMENT.md) - Local development guide
- [Frontend Setup](./FRONTEND_SETUP.md) - Frontend-specific setup

