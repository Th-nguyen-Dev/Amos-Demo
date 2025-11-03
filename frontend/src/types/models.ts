// Domain models matching backend Go types

export interface QAPair {
  id: string // UUID
  question: string
  answer: string
  created_at: string // ISO 8601 timestamp
  updated_at: string // ISO 8601 timestamp
}

export interface Conversation {
  id: string // UUID
  title: string | null
  created_at: string
  updated_at: string
}

export interface ToolCall {
  id: string
  type: 'function'
  function: {
    name: string
    arguments: string // JSON string
  }
}

export interface Message {
  id: string // UUID
  conversation_id: string // UUID
  role: 'user' | 'assistant' | 'tool' | 'system'
  content: string | null
  tool_call_id: string | null
  raw_message: {
    role: string
    content?: string | null
    tool_calls?: ToolCall[]
    tool_call_id?: string
    name?: string // for tool messages
  }
  created_at: string
}

export interface CursorPagination {
  next_cursor?: string
  prev_cursor?: string
  has_next: boolean
  has_prev: boolean
}

export interface SimilarityMatch {
  qa_pair: QAPair
  score: number
}


