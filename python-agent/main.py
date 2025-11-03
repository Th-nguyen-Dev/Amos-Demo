"""FastAPI application entry point for LangChain AI Agent."""
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from agent.config import settings
from api.routes import router as chat_router
import uvicorn

app = FastAPI(
    title="LangChain AI Agent",
    description="Conversational AI agent using Gemini 2.5 Pro",
    version="1.0.0"
)

# CORS configuration for React frontend
app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.cors_origins,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include routes
app.include_router(chat_router)


@app.get("/health")
async def health_check():
    """Health check endpoint."""
    return {"status": "healthy", "model": settings.gemini_model}


if __name__ == "__main__":
    uvicorn.run(
        "main:app",
        host=settings.api_host,
        port=settings.api_port,
        reload=True
    )

