import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react'
import type {
  QAPair,
  CreateQARequest,
  UpdateQARequest,
  CreateQAResponse,
  UpdateQAResponse,
  ListQAResponse,
  CursorParams,
} from '@/types/api'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

export const qaApi = createApi({
  reducerPath: 'qaApi',
  baseQuery: fetchBaseQuery({ baseUrl: `${API_BASE_URL}/api` }),
  tagTypes: ['QAPair'],
  endpoints: (builder) => ({
    // List Q&A pairs with pagination and search
    listQAPairs: builder.query<ListQAResponse, CursorParams | void>({
      query: (params = {}) => {
        const searchParams = new URLSearchParams()
        if (params.limit) searchParams.append('limit', params.limit.toString())
        if (params.cursor) searchParams.append('cursor', params.cursor)
        if (params.direction) searchParams.append('direction', params.direction)
        if (params.search) searchParams.append('search', params.search)
        
        return {
          url: `/qa-pairs?${searchParams.toString()}`,
        }
      },
      providesTags: (result) =>
        result
          ? [
              ...result.data.map(({ id }) => ({ type: 'QAPair' as const, id })),
              { type: 'QAPair', id: 'LIST' },
            ]
          : [{ type: 'QAPair', id: 'LIST' }],
    }),

    // Get single Q&A pair
    getQAPair: builder.query<QAPair, string>({
      query: (id) => `/qa-pairs/${id}`,
      providesTags: (result, error, id) => [{ type: 'QAPair', id }],
    }),

    // Create Q&A pair
    createQAPair: builder.mutation<CreateQAResponse, CreateQARequest>({
      query: (body) => ({
        url: '/qa-pairs',
        method: 'POST',
        body,
      }),
      invalidatesTags: [{ type: 'QAPair', id: 'LIST' }],
    }),

    // Update Q&A pair
    updateQAPair: builder.mutation<UpdateQAResponse, { id: string; data: UpdateQARequest }>({
      query: ({ id, data }) => ({
        url: `/qa-pairs/${id}`,
        method: 'PUT',
        body: data,
      }),
      invalidatesTags: (result, error, { id }) => [
        { type: 'QAPair', id },
        { type: 'QAPair', id: 'LIST' },
      ],
    }),

    // Delete Q&A pair
    deleteQAPair: builder.mutation<{ success: boolean }, string>({
      query: (id) => ({
        url: `/qa-pairs/${id}`,
        method: 'DELETE',
      }),
      invalidatesTags: (result, error, id) => [
        { type: 'QAPair', id },
        { type: 'QAPair', id: 'LIST' },
      ],
    }),
  }),
})

export const {
  useListQAPairsQuery,
  useGetQAPairQuery,
  useCreateQAPairMutation,
  useUpdateQAPairMutation,
  useDeleteQAPairMutation,
} = qaApi

