import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react'

// Python agent URL (port 8000)
const PYTHON_AGENT_URL = import.meta.env.VITE_PYTHON_AGENT_URL || 'http://localhost:8000'

// Go backend URL (port 8080) for conversations
const BACKEND_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

export interface CreateConversationRequest {
  title: string
}

export interface Conversation {
  id: string
  title: string
  created_at: string
  updated_at: string
}

export interface SendMessageRequest {
  message: string
}

export interface Message {
  id: string
  conversation_id: string
  role: 'user' | 'assistant' | 'tool'
  content: string | null
  tool_call_id?: string | null
  raw_message: Record<string, unknown>
  created_at: string
}

export const chatApi = createApi({
  reducerPath: 'chatApi',
  baseQuery: fetchBaseQuery({ baseUrl: BACKEND_URL }),
  tagTypes: ['Conversation', 'Message'],
  endpoints: (builder) => ({
    // List all conversations
    listConversations: builder.query<{ data: Conversation[] }, void>({
      query: () => '/api/conversations',
      providesTags: ['Conversation'],
    }),

    // Get a single conversation
    getConversation: builder.query<Conversation, string>({
      query: (id) => `/api/conversations/${id}`,
      providesTags: (_result, _error, id) => [{ type: 'Conversation', id }],
    }),

    // Create a new conversation
    createConversation: builder.mutation<Conversation, CreateConversationRequest>({
      query: (body) => ({
        url: '/api/conversations',
        method: 'POST',
        body,
      }),
      invalidatesTags: ['Conversation'],
    }),

    // Delete a conversation
    deleteConversation: builder.mutation<void, string>({
      query: (id) => ({
        url: `/api/conversations/${id}`,
        method: 'DELETE',
      }),
      invalidatesTags: ['Conversation'],
    }),

    // Get messages for a conversation
    getMessages: builder.query<{ data: Message[] }, string>({
      query: (conversationId) => `/api/conversations/${conversationId}/messages`,
      providesTags: (_result, _error, conversationId) => [{ type: 'Message', id: conversationId }],
    }),
  }),
})

// Streaming event types from Python agent
export interface StreamEvent {
  type: 'langchain_event' | 'error'
  data: {
    // For langchain_event
    event?: string
    name?: string
    data?: {
      chunk?: {
        content?: string
        tool_calls?: Array<{
          id: string
          name: string
          args: Record<string, unknown>
        }>
      }
      input?: Record<string, unknown>
      output?: {
        tool_call_id?: string
        name?: string
        content?: string
      }
    }
    // For error
    message?: string
  }
}

// Hook for sending messages with streaming (using fetch directly for streaming)
export async function* sendMessageStreaming(
  conversationId: string,
  message: string
): AsyncGenerator<StreamEvent, void, unknown> {
  const response = await fetch(`${PYTHON_AGENT_URL}/chat/conversations/${conversationId}/messages`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ message }),
  })

  if (!response.ok) {
    throw new Error(`Failed to send message: ${response.statusText}`)
  }

  const reader = response.body?.getReader()
  const decoder = new TextDecoder()

  if (!reader) {
    throw new Error('No response body')
  }

  let buffer = ''

  while (true) {
    const { done, value } = await reader.read()
    if (done) break

    buffer += decoder.decode(value, { stream: true })
    
    // Process complete lines (newline-delimited JSON)
    const lines = buffer.split('\n')
    buffer = lines.pop() || '' // Keep incomplete line in buffer
    
    for (const line of lines) {
      if (line.trim()) {
        try {
          const event = JSON.parse(line) as StreamEvent
          yield event
        } catch (e) {
          console.error('Failed to parse event:', line, e)
        }
      }
    }
  }
}

export const {
  useListConversationsQuery,
  useGetConversationQuery,
  useCreateConversationMutation,
  useDeleteConversationMutation,
  useGetMessagesQuery,
} = chatApi


