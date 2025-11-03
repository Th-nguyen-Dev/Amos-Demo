# Bug Fix: Missing Tool Calls in UI and Database

## Problem Summary

When the AI agent makes multiple tool calls using the ReAct pattern (e.g., trying semantic search first, then falling back to keyword search), **only the first tool call was being saved to the database and displayed in the UI**.

### What Was Happening

**User Question:** "What is your policy on refund"

**Expected Flow:**
1. User message saved ✅
2. **Assistant message** with `semantic_search_knowledge_base` tool call ✅
3. Tool result (500 error) ✅
4. **Assistant message** with `search_knowledge_base` tool call ❌ **MISSING!**
5. Tool result (success) ✅ (orphaned without matching assistant message)
6. Final assistant response ✅

**Actual Database State:**
```json
// Message 2: Assistant with first tool call
{
  "role": "assistant",
  "tool_calls": [{"id": "call_77ce494b", "function": {"name": "semantic_search_knowledge_base"}}]
}

// Message 3: Tool result for first call
{
  "role": "tool",
  "tool_call_id": "call_77ce494b",
  "content": "Error: 500..."
}

// Message 4: Tool result for SECOND call (orphaned!)
{
  "role": "tool",
  "tool_call_id": "call_65cce98b",  // <-- NO MATCHING ASSISTANT MESSAGE!
  "content": "Result 1: ... 30-day money-back guarantee ..."
}

// Message 5: Final response
{
  "role": "assistant",
  "content": "Our refund policy is a 30-day money-back guarantee..."
}
```

## Root Cause

In `/python-agent/agent/agent.py`, the variables tracking tool calls were initialized once at the start but **never reset between ReAct cycles**:

```python
current_tool_calls = []  # Initialized once
tool_call_assistant_saved = False  # Set to True after first tool call
```

**The Bug:**
1. First tool call starts → `tool_call_assistant_saved = False`
2. Assistant message with first tool call is saved
3. `tool_call_assistant_saved` set to `True` ✅
4. First tool completes
5. **Agent decides to try second tool** (new ReAct cycle)
6. Second tool call starts → `tool_call_assistant_saved` is still `True`! 
7. **Assistant message is NOT saved** ❌
8. Second tool result is saved with orphaned `tool_call_id`

## The Fix

**File:** `python-agent/agent/agent.py`

**Changed:** Lines 213-222

**Before:**
```python
elif kind == "on_tool_start":
    tool_name = event["name"]
    tool_input = event["data"].get("input", {})
    tool_call_id = f"call_{str(uuid.uuid4())[:8]}"
    
    # Create tool call in OpenAI format
    tool_call = {
        "id": tool_call_id,
        "type": "function",
        "function": {
            "name": tool_name,
            "arguments": json.dumps(tool_input)
        }
    }
    current_tool_calls.append(tool_call)
    
    # Save assistant message with tool_calls if not already saved
    if not tool_call_assistant_saved and current_tool_calls:
```

**After:**
```python
elif kind == "on_tool_start":
    tool_name = event["name"]
    tool_input = event["data"].get("input", {})
    tool_call_id = f"call_{str(uuid.uuid4())[:8]}"
    
    # If we've already saved an assistant message with tool calls,
    # this is a new ReAct cycle - reset for the new tool call
    if tool_call_assistant_saved:
        current_tool_calls = []
        tool_call_assistant_saved = False
    
    # Create tool call in OpenAI format
    tool_call = {
        "id": tool_call_id,
        "type": "function",
        "function": {
            "name": tool_name,
            "arguments": json.dumps(tool_input)
        }
    }
    current_tool_calls.append(tool_call)
    
    # Save assistant message with tool_calls if not already saved
    if not tool_call_assistant_saved and current_tool_calls:
```

**Key Change:** Added lines 218-222 to **reset the tracking variables** when starting a new tool call after a previous one has been saved.

## Expected Behavior After Fix

Now when the agent makes multiple tool calls:

1. ✅ First tool call → Assistant message saved
2. ✅ First tool completes
3. ✅ **Reset tracking variables** when second tool starts
4. ✅ Second tool call → **New assistant message saved**
5. ✅ Second tool completes (no longer orphaned)
6. ✅ Final response

**New Database State:**
```json
// Message 2: Assistant with first tool call
{
  "role": "assistant",
  "tool_calls": [{"id": "call_abc123", "function": {"name": "semantic_search_knowledge_base"}}]
}

// Message 3: Tool result for first call
{
  "role": "tool",
  "tool_call_id": "call_abc123"
}

// Message 4: Assistant with SECOND tool call (NEW - no longer missing!)
{
  "role": "assistant",
  "tool_calls": [{"id": "call_xyz789", "function": {"name": "search_knowledge_base"}}]
}

// Message 5: Tool result for second call (no longer orphaned!)
{
  "role": "tool",
  "tool_call_id": "call_xyz789"
}

// Message 6: Final response
{
  "role": "assistant",
  "content": "..."
}
```

## Testing the Fix

1. **Restart the Python agent service:**
   ```bash
   docker-compose restart python-agent
   ```

2. **Create a new conversation** (important - old conversations may have cached state)

3. **Ask a question that triggers multiple tool calls:**
   - "What is your refund policy?"
   - The agent should try semantic search (which may fail), then keyword search

4. **Check the message history:**
   ```bash
   curl http://localhost:8080/api/conversations/{conversation_id}/messages
   ```

5. **Verify:**
   - ✅ Each tool call has a matching assistant message with `tool_calls` array
   - ✅ Each tool result has a `tool_call_id` that matches an assistant message
   - ✅ No orphaned tool results
   - ✅ UI displays all tool calls

## Related Issue: Semantic Search 500 Error

The bug investigation also revealed that `semantic_search_knowledge_base` is returning a 500 error. This is a separate issue likely caused by:

1. **Pinecone not configured or not running** (check `.env` and `docker-compose.yml`)
2. **QA pairs not indexed in Pinecone** (need to run batch indexing)
3. **Google Embedding API key not configured** (falls back to mock client which may not work properly)

### To Fix Semantic Search:

Check if Pinecone is configured:
```bash
# Check environment variables
grep PINECONE backend/.env

# If using Pinecone Local, check if it's running
docker-compose ps pinecone
```

Index QA pairs:
```bash
# Run batch indexing script
cd backend
go run cmd/batch-index/main.go
```

## Impact

**Before Fix:**
- ❌ Only first tool call visible in UI
- ❌ Database had orphaned tool results
- ❌ Confusing user experience ("where did the second tool call go?")
- ❌ Could not properly track agent reasoning

**After Fix:**
- ✅ All tool calls visible in UI
- ✅ Database has complete conversation history
- ✅ Users can see full agent reasoning (tried A, then B)
- ✅ Proper tool call → result mapping

## Files Changed

1. `python-agent/agent/agent.py` - Fixed ReAct cycle tracking (lines 218-222)

## Files to Review (No Changes Needed)

1. `frontend/src/features/chat/pages/ChatPage.tsx` - Tool call rendering (already correct)
2. `frontend/src/features/chat/components/ChatMessage.tsx` - Tool call display (already correct)
3. `backend/cmd/server/main.go` - Semantic search endpoint (working correctly)

