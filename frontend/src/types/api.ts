// API request and response types

import type { QAPair, CursorPagination, Conversation, Message, SimilarityMatch } from './models'

// Re-export for convenience
export type { QAPair, CursorPagination }

// Error Response
export interface ErrorResponse {
  error: string
  code: string
  message: string
  details?: Record<string, unknown>
}

// Q&A API Types
export interface CreateQARequest {
  question: string
  answer: string
}

export interface UpdateQARequest {
  question: string
  answer: string
}

export interface CreateQAResponse {
  qa_pair: QAPair
}

export interface UpdateQAResponse {
  qa_pair: QAPair
}

export interface ListQAResponse {
  data: QAPair[]
  pagination: CursorPagination
}

export interface CursorParams {
  limit?: number
  cursor?: string
  direction?: 'next' | 'prev'
  search?: string
}

// Chat API Types
export interface AskQuestionRequest {
  question: string
  conversation_id?: string
}

export interface AskQuestionResponse {
  answer: string
  conversation_id?: string
  sources?: SimilarityMatch[]
}

// Conversation API Types
export interface CreateConversationRequest {
  title?: string
}

export interface CreateConversationResponse {
  conversation: Conversation
}

export interface ListConversationsResponse {
  data: Conversation[]
  pagination: CursorPagination
}

export interface ListMessagesResponse {
  data: Message[]
  pagination: CursorPagination
}

