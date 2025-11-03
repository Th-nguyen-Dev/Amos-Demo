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
      query: (params) => {
        const searchParams = new URLSearchParams()
        const p = params || {}
        if (p.limit) searchParams.append('limit', p.limit.toString())
        if (p.cursor) searchParams.append('cursor', p.cursor)
        if (p.direction) searchParams.append('direction', p.direction)
        if (p.search) searchParams.append('search', p.search)
        
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
      // Keep cache for 5 seconds and refetch when args change
      keepUnusedDataFor: 5,
      refetchOnMountOrArgChange: true,
    }),

    // Get single Q&A pair
    getQAPair: builder.query<QAPair, string>({
      query: (id) => `/qa-pairs/${id}`,
      providesTags: (_result, _error, id) => [{ type: 'QAPair', id }],
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
      invalidatesTags: (_result, _error, { id }) => [
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
      invalidatesTags: (_result, _error, id) => [
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

