import { configureStore } from '@reduxjs/toolkit'
import { qaApi } from '@/features/qa/api/qaApi'
import { chatApi } from '@/features/chat/api/chatApi'

export const store = configureStore({
  reducer: {
    [qaApi.reducerPath]: qaApi.reducer,
    [chatApi.reducerPath]: chatApi.reducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware().concat(qaApi.middleware, chatApi.middleware),
})

export type RootState = ReturnType<typeof store.getState>
export type AppDispatch = typeof store.dispatch

