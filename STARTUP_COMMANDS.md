# How to Start Frontend and Backend

## ✅ Current Status

**All services are now running!**

```
✅ Backend API       → http://localhost:8080 (Running in Docker)
✅ Frontend          → http://localhost:5173 (Running locally)
✅ PostgreSQL        → localhost:5432 (Running in Docker)
✅ Pinecone Local    → localhost:5081 (Running in Docker)
⚠️  Python Agent     → localhost:8000 (Restarting - needs GEMINI_API_KEY)
```

---

## 🚀 Quick Start Commands

### To Start Everything:

**Terminal 1 - Backend Services:**
```bash
cd "/home/electron/projects/Amos Demo"
docker-compose up -d
```

**Terminal 2 - Frontend:**
```bash
cd "/home/electron/projects/Amos Demo/frontend"
npm run dev
```

---

## 🛑 To Stop Everything:

**Stop Backend:**
```bash
cd "/home/electron/projects/Amos Demo"
docker-compose down
```

**Stop Frontend:**
Press `Ctrl+C` in the terminal where `npm run dev` is running

---

## ✅ Verify Everything is Working:

### 1. Test Backend API
```bash
curl http://localhost:8080/health
# Expected: {"status":"healthy","database":"connected"}
```

### 2. Test Frontend
Open your browser and go to: **http://localhost:5173**

### 3. Test Connection
1. Navigate to "Q&A Management" in the frontend
2. You should see a list of Q&A pairs
3. Try creating a new Q&A pair

---

## 📊 Service Details:

| Service | Port | URL | Command |
|---------|------|-----|---------|
| **Backend API** | 8080 | http://localhost:8080 | `docker-compose up -d` |
| **Frontend** | 5173 | http://localhost:5173 | `npm run dev` |
| **PostgreSQL** | 5432 | localhost:5432 | (via Docker) |
| **Pinecone Local** | 5081 | http://localhost:5081 | (via Docker) |
| **Python Agent** | 8000 | http://localhost:8000 | (via Docker) |

---

## 🔍 Checking Service Status:

### Backend Status
```bash
docker-compose ps
```

### Frontend Status
```bash
ps aux | grep vite
```

### Backend Logs
```bash
docker-compose logs -f backend
```

### Frontend Logs
Check the terminal where `npm run dev` is running

---

## 🐛 Troubleshooting:

### Backend won't start?
```bash
# Check if ports are in use
lsof -i :8080
lsof -i :5432

# Restart services
docker-compose restart
```

### Frontend won't start?
```bash
# Check if port is in use
lsof -i :5173

# Kill existing process if needed
kill -9 <PID>

# Try starting again
cd frontend
npm run dev
```

### Port conflicts?
If you get "address already in use" errors:
```bash
# Find what's using the port
lsof -i :<PORT_NUMBER>

# Stop the conflicting service or change the port
```

### Cannot connect to backend from frontend?
1. Check backend is running: `curl http://localhost:8080/health`
2. Check browser console for CORS errors (F12)
3. Make sure `.env` in frontend has: `VITE_API_BASE_URL=http://localhost:8080`

---

## 🔄 Restart Services:

### Restart Backend
```bash
docker-compose restart backend
```

### Restart All Docker Services
```bash
docker-compose down
docker-compose up -d
```

### Restart Frontend
Press `Ctrl+C` in the terminal, then:
```bash
npm run dev
```

---

## 🎯 API Endpoints:

### Health Check
```bash
curl http://localhost:8080/health
```

### List Q&A Pairs
```bash
curl http://localhost:8080/api/qa-pairs
```

### Create Q&A Pair
```bash
curl -X POST http://localhost:8080/api/qa-pairs \
  -H "Content-Type: application/json" \
  -d '{
    "question": "What is Docker?",
    "answer": "Docker is a containerization platform.",
    "category": "technology"
  }'
```

---

## 📝 Python Agent Setup (Optional)

The Python Agent requires a Gemini API key. To enable it:

1. Get your Gemini API key from: https://makersuite.google.com/app/apikey

2. Create a `.env` file in the project root:
```bash
cat > .env << 'EOF'
GEMINI_API_KEY=your-gemini-api-key-here
EOF
```

3. Restart Docker services:
```bash
docker-compose down
docker-compose up -d
```

---

## 🎉 You're All Set!

**Frontend:** http://localhost:5173  
**Backend API:** http://localhost:8080  
**API Health:** http://localhost:8080/health  

The frontend and backend are configured to work together automatically!

---

## 📚 Related Documentation:

- [Frontend-Backend Connection Guide](./FRONTEND_BACKEND_CONNECTION.md)
- [Main README](./README.md)
- [Docker Quick Start](./DOCKER_QUICKSTART.md)
- [Local Development](./docs/LOCAL_DEVELOPMENT.md)

