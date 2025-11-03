"""Type-safe HTTP client for Go backend APIs."""
import httpx
from typing import Optional
from uuid import UUID
from agent.models import (
    Conversation, Message, QAPair,
    SearchQARequest, SearchQAResponse,
    GetQAByIDsRequest, GetQAByIDsResponse
)
from agent.config import settings


class BackendClient:
    """Type-safe HTTP client for Go backend APIs."""
    
    def __init__(self, base_url: str = settings.backend_url):
        self.base_url = base_url
        self.client = httpx.AsyncClient(base_url=base_url, timeout=30.0)
    
    async def create_conversation(
        self, 
        title: Optional[str] = None
    ) -> Conversation:
        """Create a new conversation."""
        response = await self.client.post(
            "/api/conversations",
            json={"title": title} if title else {}
        )
        response.raise_for_status()
        data = response.json()
        return Conversation(**data["conversation"])
    
    async def get_messages(
        self, 
        conversation_id: UUID
    ) -> list[Message]:
        """Get messages for a conversation."""
        response = await self.client.get(
            f"/api/conversations/{conversation_id}/messages"
        )
        response.raise_for_status()
        data = response.json()
        return [Message(**msg) for msg in data["data"]]
    
    async def search_qa(
        self, 
        query: str, 
        limit: int = 5
    ) -> SearchQAResponse:
        """Search QA pairs using full-text search."""
        request = SearchQARequest(query=query, limit=limit)
        response = await self.client.post(
            "/tools/search-qa",
            json=request.model_dump()
        )
        response.raise_for_status()
        return SearchQAResponse(**response.json())
    
    async def get_qa_by_ids(
        self, 
        ids: list[UUID]
    ) -> GetQAByIDsResponse:
        """Get specific QA pairs by their IDs."""
        request = GetQAByIDsRequest(ids=ids)
        response = await self.client.post(
            "/tools/get-qa-by-ids",
            json={"ids": [str(id) for id in ids]}
        )
        response.raise_for_status()
        return GetQAByIDsResponse(**response.json())
    
    async def save_message(
        self,
        conversation_id: UUID,
        role: str,
        content: Optional[str],
        tool_call_id: Optional[str],
        raw_message: dict
    ) -> Message:
        """Save a message to the backend."""
        payload = {
            "conversation_id": str(conversation_id),
            "role": role,
            "content": content,
            "tool_call_id": tool_call_id,
            "raw_message": raw_message
        }
        response = await self.client.post(
            "/tools/save-message",
            json=payload
        )
        response.raise_for_status()
        data = response.json()
        return Message(**data["message"])
    
    async def semantic_search_qa(
        self, 
        query: str, 
        top_k: int = 5
    ) -> SearchQAResponse:
        """
        Pinecone semantic search stub.
        TODO: Implement when Pinecone is configured.
        Currently falls back to full-text search.
        """
        if not settings.use_pinecone:
            # Fallback to full-text search
            return await self.search_qa(query, top_k)
        
        # TODO: Call Pinecone-enabled endpoint when available
        # This would be something like:
        # response = await self.client.post(
        #     "/tools/semantic-search-qa",
        #     json={"query": query, "top_k": top_k}
        # )
        raise NotImplementedError("Pinecone integration pending")
    
    async def close(self):
        """Close the HTTP client."""
        await self.client.aclose()

