"""FastAPI routes for chat endpoints."""
from fastapi import APIRouter, HTTPException
from fastapi.responses import StreamingResponse
from uuid import UUID
from api.schemas import (
    ChatRequest, ChatResponse, 
    CreateConversationRequest, ConversationResponse
)
from agent.agent import ConversationalAgent
from agent.client import BackendClient

router = APIRouter(prefix="/chat", tags=["chat"])
agent = ConversationalAgent()
backend_client = BackendClient()


@router.post("/conversations", response_model=ConversationResponse)
async def create_conversation(request: CreateConversationRequest):
    """Create a new conversation."""
    try:
        conv = await backend_client.create_conversation(request.title)
        return ConversationResponse(conversation=conv)
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.post("/conversations/{conversation_id}/messages")
async def send_message(conversation_id: UUID, request: ChatRequest):
    """Send a message and stream the agent's response as JSON events."""
    import json
    import logging
    
    logger = logging.getLogger(__name__)
    
    try:
        async def generate():
            try:
                async for event in agent.chat(conversation_id, request.message):
                    # Stream as newline-delimited JSON
                    yield f"{event}\n"
            except Exception as e:
                logger.error(f"Error in chat stream: {e}")
                error_event = json.dumps({
                    "type": "error",
                    "data": {"message": str(e)}
                })
                yield f"{error_event}\n"
        
        return StreamingResponse(generate(), media_type="application/x-ndjson")
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/conversations/{conversation_id}/messages")
async def get_messages(conversation_id: UUID):
    """Get conversation messages."""
    try:
        messages = await backend_client.get_messages(conversation_id)
        return {"messages": messages}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
