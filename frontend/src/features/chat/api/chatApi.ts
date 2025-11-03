import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react'
import type { AskQuestionRequest, AskQuestionResponse } from '@/types/api'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

export const chatApi = createApi({
  reducerPath: 'chatApi',
  baseQuery: fetchBaseQuery({ baseUrl: `${API_BASE_URL}/api` }),
  endpoints: (builder) => ({
    // Ask a question to the AI assistant
    askQuestion: builder.mutation<AskQuestionResponse, AskQuestionRequest>({
      query: (body) => ({
        url: '/ask',
        method: 'POST',
        body,
      }),
    }),
  }),
})

export const {
  useAskQuestionMutation,
} = chatApi


