"""Pydantic models matching Go backend OpenAI message format."""
from pydantic import BaseModel, Field, ConfigDict
from typing import Optional, Any, Literal
from uuid import UUID
from datetime import datetime


class ToolCall(BaseModel):
    """OpenAI tool call format."""
    model_config = ConfigDict(strict=True)
    
    id: str
    type: Literal["function"] = "function"
    function: "FunctionCall"


class FunctionCall(BaseModel):
    """OpenAI function call format."""
    model_config = ConfigDict(strict=True)
    
    name: str
    arguments: str  # JSON string


class Message(BaseModel):
    """OpenAI message format matching Go backend."""
    model_config = ConfigDict(strict=True)
    
    id: UUID
    conversation_id: UUID
    role: Literal["user", "assistant", "tool", "system"]
    content: Optional[str] = None
    tool_call_id: Optional[str] = None
    raw_message: dict[str, Any]
    created_at: datetime


class Conversation(BaseModel):
    """Conversation model matching Go backend."""
    model_config = ConfigDict(strict=True)
    
    id: UUID
    title: Optional[str] = None
    created_at: datetime
    updated_at: datetime


class QAPair(BaseModel):
    """QA pair model matching Go backend."""
    model_config = ConfigDict(strict=True)
    
    id: UUID
    question: str
    answer: str
    created_at: datetime
    updated_at: datetime


class SearchQARequest(BaseModel):
    """Search QA request."""
    model_config = ConfigDict(strict=True)
    
    query: str = Field(..., min_length=1, max_length=200)
    limit: int = Field(default=5, ge=1, le=100)


class SearchQAResponse(BaseModel):
    """Search QA response."""
    model_config = ConfigDict(strict=True)
    
    qa_pairs: list[QAPair]
    count: int


class GetQAByIDsRequest(BaseModel):
    """Get QA by IDs request."""
    model_config = ConfigDict(strict=True)
    
    ids: list[UUID] = Field(..., min_length=1, max_length=50)


class GetQAByIDsResponse(BaseModel):
    """Get QA by IDs response."""
    model_config = ConfigDict(strict=True)
    
    qa_pairs: list[QAPair]

