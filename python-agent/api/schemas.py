"""FastAPI request/response schemas."""
from pydantic import BaseModel, Field
from typing import Optional
from agent.models import Conversation, Message


class CreateConversationRequest(BaseModel):
    """Request to create a new conversation."""
    title: Optional[str] = Field(None, max_length=200)


class ConversationResponse(BaseModel):
    """Response containing a conversation."""
    conversation: Conversation


class ChatRequest(BaseModel):
    """Request to send a chat message."""
    message: str = Field(..., min_length=1, max_length=2000)


class ChatResponse(BaseModel):
    """Response containing chat message."""
    response: str
    conversation_id: str

