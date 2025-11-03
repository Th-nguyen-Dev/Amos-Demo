"""LangChain tools for the AI agent to interact with Go backend."""
from langchain_core.tools import tool
from agent.client import BackendClient
from typing import Annotated
from uuid import UUID

# Shared backend client instance
backend_client = BackendClient()


@tool
async def search_knowledge_base(
    query: Annotated[str, "The search query to find relevant Q&A pairs. Be specific with keywords."],
    limit: Annotated[int, "Number of results to return (1-10)"] = 5
) -> str:
    """
    🔍 PRIMARY TOOL: Search the company knowledge base for relevant Q&A pairs using full-text search.
    
    ⚠️ USE THIS TOOL FIRST for ANY user question before responding!
    
    This searches through all available Q&A pairs and returns the most relevant matches.
    If you don't find what you need, try different keywords or search terms.
    
    Returns: Matching question-answer pairs from the knowledge base with their IDs.
    """
    try:
        response = await backend_client.search_qa(query, min(limit, 10))
        
        if response.count == 0:
            return "No relevant information found in the knowledge base."
        
        results = []
        for i, qa in enumerate(response.qa_pairs, 1):
            results.append(
                f"Result {i}:\n"
                f"Question: {qa.question}\n"
                f"Answer: {qa.answer}\n"
                f"ID: {qa.id}"
            )
        
        return "\n\n".join(results)
    except Exception as e:
        return f"Error searching knowledge base: {str(e)}"


@tool
async def get_qa_by_ids(
    qa_ids: Annotated[list[str], "List of QA pair UUIDs to retrieve"]
) -> str:
    """
    Retrieve specific Q&A pairs by their IDs.
    Use this when you need to fetch exact Q&A pairs that were previously referenced.
    
    Each ID should be a valid UUID string.
    """
    try:
        # Convert string IDs to UUID objects
        uuids = [UUID(qa_id) for qa_id in qa_ids]
        
        # Call backend endpoint
        response = await backend_client.get_qa_by_ids(uuids)
        
        if not response.qa_pairs:
            return "No Q&A pairs found for the provided IDs."
        
        results = []
        for qa in response.qa_pairs:
            results.append(
                f"ID: {qa.id}\n"
                f"Question: {qa.question}\n"
                f"Answer: {qa.answer}"
            )
        
        return "\n\n".join(results)
    except ValueError as e:
        return f"Invalid UUID format: {str(e)}"
    except Exception as e:
        return f"Error retrieving Q&A pairs: {str(e)}"


@tool
async def semantic_search_knowledge_base(
    query: Annotated[str, "The semantic search query"],
    top_k: Annotated[int, "Number of semantically similar results (1-20)"] = 5
) -> str:
    """
    Search the knowledge base using semantic similarity (Pinecone vector search).
    Use this for finding conceptually related content, even if keywords don't match exactly.
    
    NOTE: Currently falls back to full-text search until Pinecone is configured.
    """
    try:
        response = await backend_client.semantic_search_qa(query, top_k)
        
        if response.count == 0:
            return "No semantically similar information found."
        
        results = []
        for i, qa in enumerate(response.qa_pairs, 1):
            results.append(
                f"Result {i}:\n"
                f"Question: {qa.question}\n"
                f"Answer: {qa.answer}"
            )
        
        return "\n\n".join(results)
    except NotImplementedError:
        # Fallback to regular search
        return await search_knowledge_base(query, top_k)
    except Exception as e:
        return f"Error in semantic search: {str(e)}"


@tool
async def list_knowledge_base_topics() -> str:
    """
    📋 List all available Q&A pairs in the knowledge base to see what topics are covered.
    
    Use this tool to:
    - See what information is available before searching
    - Understand what topics you can help users with
    - Get an overview of the knowledge base contents
    
    Returns: A list of all Q&A pair questions/topics currently in the knowledge base.
    """
    try:
        # Get all QA pairs with a high limit
        response = await backend_client.search_qa("", limit=100)
        
        if response.count == 0:
            return "The knowledge base is currently empty. No Q&A pairs have been added yet."
        
        topics = []
        for i, qa in enumerate(response.qa_pairs, 1):
            topics.append(f"{i}. {qa.question} (ID: {qa.id})")
        
        header = f"📚 Knowledge Base Contents ({response.count} Q&A pairs):\n\n"
        return header + "\n".join(topics)
    except Exception as e:
        return f"Error listing topics: {str(e)}"


# Complete tool list - Agent will choose which tools to use based on context
# ORDER MATTERS: Most important tools first
tools = [
    search_knowledge_base,          # Primary tool - use this first!
    list_knowledge_base_topics,     # Helper to see what's available
    semantic_search_knowledge_base, # Alternative search method
    get_qa_by_ids,                  # For retrieving specific items
]

