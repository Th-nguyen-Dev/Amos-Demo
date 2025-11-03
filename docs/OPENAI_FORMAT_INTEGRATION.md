# 🔧 OpenAI Format Integration - Tool Calls in JSONB

## ✅ **What Was Implemented**

We've migrated from markdown-embedded tool calls to **proper OpenAI-format messages** stored as structured JSONB in the database. This provides a much cleaner, more maintainable architecture.

---

## 🎯 **Architecture Overview**

### **Message Flow:**

```
User sends message
     ↓
Python Agent runs tools
     ↓
Saves messages in OpenAI format:
  1. Assistant message with tool_calls array
  2. Tool result messages (one per tool call)
     ↓
Go Backend stores in PostgreSQL:
  - content: Text content
  - raw_message: Full OpenAI-format JSONB
     ↓
Frontend fetches and parses JSONB
     ↓
Renders tool calls with structured components
```

---

## 📦 **Database Schema**

### **messages table:**

```sql
CREATE TABLE messages (
    id UUID PRIMARY KEY,
    conversation_id UUID REFERENCES conversations(id),
    role TEXT,              -- 'user', 'assistant', 'tool', 'system'
    content TEXT,           -- Plain text content  
    tool_call_id TEXT,      -- For tool result messages
    raw_message JSONB,      -- Full OpenAI-format message
    created_at TIMESTAMP
);
```

### **OpenAI Message Formats:**

#### **1. User Message:**
```json
{
  "role": "user",
  "content": "What is Docker?"
}
```

#### **2. Assistant Message with Tool Calls:**
```json
{
  "role": "assistant",
  "content": "Let me search for that information.",
  "tool_calls": [
    {
      "id": "call_abc123",
      "type": "function",
      "function": {
        "name": "search_knowledge_base",
        "arguments": "{\"query\": \"Docker\", \"limit\": 5}"
      }
    }
  ]
}
```

#### **3. Tool Result Message:**
```json
{
  "role": "tool",
  "tool_call_id": "call_abc123",
  "name": "search_knowledge_base",
  "content": "Result 1:\nQuestion: What is Docker?\nAnswer: Docker is..."
}
```

---

## 🐍 **Python Agent Changes**

### **File: `python-agent/agent/agent.py`**

#### **Before (Markdown embedded):**
```python
# Streamed text with embedded markdown
yield f"\n\n{'─' * 50}\n"
yield f"🔧 **Tool Call: {tool_name}**\n\n"
# ... more formatting ...

# Saved everything as plain text
await self.backend_client.save_message(
    content=complete_response,  # All text
    raw_message={"role": "assistant", "content": complete_response}
)
```

#### **After (OpenAI format):**
```python
# Track tool calls in OpenAI format
tool_calls_made = []  # List of tool call objects
tool_results = []     # List of tool results

# On tool start
tool_call = {
    "id": f"call_{uuid.uuid4()[:8]}",
    "type": "function",
    "function": {
        "name": tool_name,
        "arguments": json.dumps(tool_input)
    }
}
tool_calls_made.append(tool_call)

# Stream simple notification (optional)
yield f"\n\n🔧 **Tool: {tool_name}**\n"

# Save assistant message with tool_calls
assistant_msg = {
    "role": "assistant",
    "content": full_response,
    "tool_calls": tool_calls_made  # ← Structured data
}
await self.backend_client.save_message(
    content=full_response,
    raw_message=assistant_msg
)

# Save tool result messages separately
for tool_call, result in zip(tool_calls_made, tool_results):
    tool_msg = {
        "role": "tool",
        "tool_call_id": tool_call["id"],
        "name": result["name"],
        "content": result["output"]
    }
    await self.backend_client.save_message(
        role="tool",
        content=result["output"],
        tool_call_id=tool_call["id"],
        raw_message=tool_msg
    )
```

---

## 🎨 **Frontend Changes**

### **1. Updated Type Definitions**

#### **File: `frontend/src/types/models.ts`**

```typescript
export interface ToolCall {
  id: string
  type: 'function'
  function: {
    name: string
    arguments: string // JSON string
  }
}

export interface Message {
  id: string
  conversation_id: string
  role: 'user' | 'assistant' | 'tool' | 'system'
  content: string | null
  tool_call_id: string | null
  raw_message: {
    role: string
    content?: string | null
    tool_calls?: ToolCall[]  // ← Assistant messages
    tool_call_id?: string    // ← Tool messages
    name?: string            // ← Tool name
  }
  created_at: string
}
```

### **2. Enhanced ChatMessage Component**

#### **File: `frontend/src/features/chat/components/ChatMessage.tsx`**

```typescript
export function ChatMessage({ message, toolMessages }: ChatMessageProps) {
  // Skip rendering tool messages (shown with their calls)
  if (message.role === 'tool') return null
  
  // Parse tool calls from JSONB raw_message
  const toolCalls = message.raw_message?.tool_calls
  
  return (
    <div>
      {/* Regular message content */}
      {message.content && <div>{message.content}</div>}
      
      {/* Tool calls from JSONB */}
      {toolCalls?.map(toolCall => {
        // Find matching tool result
        const toolResult = toolMessages.find(
          tm => tm.tool_call_id === toolCall.id
        )
        
        // Parse arguments
        const args = JSON.parse(toolCall.function.arguments)
        
        return (
          <Tool status={toolResult ? "success" : "loading"}>
            <ToolHeader>🔧 {toolCall.function.name}</ToolHeader>
            <ToolContent>
              <ToolInput>{JSON.stringify(args, null, 2)}</ToolInput>
              <ToolOutput>{toolResult?.content}</ToolOutput>
            </ToolContent>
          </Tool>
        )
      })}
    </div>
  )
}
```

### **3. Updated ChatPage to Group Messages**

#### **File: `frontend/src/features/chat/pages/ChatPage.tsx`**

```typescript
{messages.map((message, index) => {
  // Find tool messages that follow this assistant message
  const toolMessages = []
  if (message.role === 'assistant') {
    for (let i = index + 1; i < messages.length; i++) {
      if (messages[i].role === 'tool') {
        toolMessages.push(messages[i])
      } else {
        break
      }
    }
  }
  
  return (
    <ChatMessage 
      message={message} 
      toolMessages={toolMessages}  // ← Pass related tools
    />
  )
})}
```

---

## 🔄 **Message Sequence Example**

### **Database Contents:**

```
Messages:
┌────────────────────────────────────────────────────────────────────────┐
│ 1. role: "user"                                                        │
│    content: "What is Docker?"                                          │
│    raw_message: { "role": "user", "content": "What is Docker?" }      │
├────────────────────────────────────────────────────────────────────────┤
│ 2. role: "assistant"                                                   │
│    content: "Based on the search results..."                          │
│    raw_message: {                                                      │
│      "role": "assistant",                                              │
│      "content": "Based on the search results...",                      │
│      "tool_calls": [                                                   │
│        {                                                               │
│          "id": "call_abc123",                                          │
│          "type": "function",                                           │
│          "function": {                                                 │
│            "name": "search_knowledge_base",                            │
│            "arguments": "{\"query\":\"Docker\",\"limit\":5}"           │
│          }                                                             │
│        }                                                               │
│      ]                                                                 │
│    }                                                                   │
├────────────────────────────────────────────────────────────────────────┤
│ 3. role: "tool"                                                        │
│    tool_call_id: "call_abc123"                                         │
│    content: "Result 1: Question: What is Docker?..."                  │
│    raw_message: {                                                      │
│      "role": "tool",                                                   │
│      "tool_call_id": "call_abc123",                                    │
│      "name": "search_knowledge_base",                                  │
│      "content": "Result 1: Question: What is Docker?..."              │
│    }                                                                   │
└────────────────────────────────────────────────────────────────────────┘
```

### **Frontend Rendering:**

```
┌─────────────────────────────────────────────────────────────────┐
│ 👤 You                                        3:45 PM           │
│ What is Docker?                                                 │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│ 🤖 AI Assistant                               3:45 PM           │
│                                                                 │
│ Based on the search results...                                  │
│                                                                 │
│ ┌─────────────────────────────────────────────────────────────┐ │
│ │ ✅ 🔧 search_knowledge_base                             ▼  │ │
│ ├─────────────────────────────────────────────────────────────┤ │
│ │ Input                                                       │ │
│ │ {                                                           │ │
│ │   "query": "Docker",                                        │ │
│ │   "limit": 5                                                │ │
│ │ }                                                           │ │
│ │                                                             │ │
│ │ Output                                                      │ │
│ │ Result 1:                                                   │ │
│ │ Question: What is Docker?                                   │ │
│ │ Answer: Docker is a containerization platform...            │ │
│ └─────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

---

## ✅ **Benefits of This Approach**

### **1. Proper Data Structure**
- ✅ Tool calls stored as structured JSONB, not text
- ✅ Can query by tool name, parse arguments programmatically
- ✅ Future-proof for analytics and tooling

### **2. OpenAI Compatibility**
- ✅ Standard format used by OpenAI, Anthropic, etc.
- ✅ Easy to migrate between providers
- ✅ Matches LangChain's message format

### **3. Clean Separation**
- ✅ Data layer (JSONB) vs presentation layer (UI components)
- ✅ Can change UI without touching data
- ✅ Tool results linked via `tool_call_id`

### **4. Database Queries**
```sql
-- Find all messages using a specific tool
SELECT * FROM messages 
WHERE raw_message->>'tool_calls' LIKE '%search_knowledge_base%';

-- Count tool usage
SELECT 
  raw_message->'tool_calls'->0->'function'->>'name' as tool_name,
  COUNT(*) 
FROM messages 
WHERE raw_message ? 'tool_calls'
GROUP BY tool_name;

-- Find failed tool calls
SELECT * FROM messages m1
JOIN messages m2 ON m2.tool_call_id = (m1.raw_message->'tool_calls'->0->>'id')
WHERE m2.content LIKE '%error%' OR m2.content LIKE '%not found%';
```

### **5. Type Safety**
- ✅ TypeScript interfaces match Go structs
- ✅ Frontend knows exact structure of tool calls
- ✅ No parsing fragile markdown

---

## 🔧 **Key Files Modified**

1. **Python Agent:**
   - `python-agent/agent/agent.py` - OpenAI format message creation
   
2. **Frontend Types:**
   - `frontend/src/types/models.ts` - ToolCall and Message interfaces
   - `frontend/src/features/chat/types.ts` - ChatMessage interface
   - `frontend/src/features/chat/api/chatApi.ts` - API Message interface

3. **Frontend Components:**
   - `frontend/src/features/chat/components/ChatMessage.tsx` - JSONB parsing
   - `frontend/src/features/chat/pages/ChatPage.tsx` - Message grouping

4. **Backend (No changes needed!):**
   - Already supported OpenAI format with raw_message JSONB
   - Already had tool_call_id field for linking

---

## 🚀 **Testing**

### **1. Start Services:**
```bash
docker-compose up -d
```

### **2. Test in Chat:**
```
User: "What is Docker?"

Expected database state after response:
- 1 user message
- 1 assistant message with tool_calls array in raw_message
- 1+ tool messages with tool_call_id linking back

Expected frontend display:
- User message
- Assistant message with expandable tool cards
- Each tool shows args (JSON) and results
```

### **3. Verify Database:**
```sql
SELECT 
  role, 
  content,
  raw_message->'tool_calls' as tool_calls,
  tool_call_id
FROM messages 
WHERE conversation_id = '<conversation_id>'
ORDER BY created_at;
```

### **4. Verify Frontend:**
- Open browser console
- Check network tab for `/api/conversations/{id}/messages`
- Verify response includes raw_message with tool_calls
- Verify Tool components render with correct data

---

## 📝 **Migration Notes**

### **Old Approach (Markdown embedded):**
```typescript
// Frontend had to parse this:
const content = `
──────────────────────────────────────────────────
🔧 **Tool Call: search_knowledge_base**

📋 **Arguments:**
  • **query:** Docker
...

Based on the search results, Docker is...
`

// Fragile parsing with regex
const toolPattern = /─{50}\n🔧 \*\*Tool Call: (.+?)\*\*/
```

### **New Approach (JSONB structured):**
```typescript
// Frontend directly accesses structured data:
const toolCalls = message.raw_message.tool_calls
// [{ id: "call_...", function: { name: "...", arguments: "..." } }]

// No parsing needed, just iterate and render
toolCalls.map(tc => <Tool key={tc.id} name={tc.function.name} />)
```

---

## 🎯 **Result**

✅ **Clean data architecture**  
✅ **OpenAI-compatible message format**  
✅ **Proper database normalization**  
✅ **Type-safe frontend**  
✅ **Easy to query and analyze**  
✅ **Tool calls persist correctly**  

**The system now uses industry-standard OpenAI message format throughout the entire stack!** 🚀

