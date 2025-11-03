// Feature-specific types for chat
export interface ChatMessage {
  id: string
  role: 'user' | 'assistant' | 'tool'
  content: string | null
  timestamp: Date
  tool_call_id?: string | null
  raw_message?: Record<string, unknown>
}


