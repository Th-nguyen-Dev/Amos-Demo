import { useState, useRef, useEffect } from 'react'
import { toast } from 'sonner'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { ChatMessage } from '../components/ChatMessage'
import { ChatInput } from '../components/ChatInput'
import { StreamingMessage } from '../components/StreamingMessage'
import {
  useListConversationsQuery,
  useGetMessagesQuery,
  useCreateConversationMutation,
  useDeleteConversationMutation,
  sendMessageStreaming,
} from '../api/chatApi'
import { PlusIcon, TrashIcon } from 'lucide-react'
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle } from '@/components/ui/alert-dialog'

export function ChatPage() {
  const [selectedConversationId, setSelectedConversationId] = useState<string | null>(null)
  const [streamingMessage, setStreamingMessage] = useState<string>('')
  const [streamingToolCalls, setStreamingToolCalls] = useState<Array<{id: string, name: string, arguments: string, status: 'loading' | 'complete'}>>([])
  const [pendingUserMessage, setPendingUserMessage] = useState<string>('')
  const [isStreaming, setIsStreaming] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [conversationToDelete, setConversationToDelete] = useState<string | null>(null)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const selectedConversationRef = useRef<HTMLDivElement>(null)

  // Fetch conversations
  const { data: conversationsData, isLoading: conversationsLoading } = useListConversationsQuery()
  const conversations = conversationsData?.data || []

  // Fetch messages for selected conversation
  const { data: messagesData, refetch: refetchMessages } = useGetMessagesQuery(
    selectedConversationId || '',
    { skip: !selectedConversationId }
  )
  const messages = messagesData?.data || []

  const [createConversation, { isLoading: isCreating }] = useCreateConversationMutation()
  const [deleteConversation, { isLoading: isDeleting }] = useDeleteConversationMutation()

  // Scroll to bottom when messages change
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages.length, streamingMessage, pendingUserMessage])

  // Scroll selected conversation into view
  useEffect(() => {
    if (selectedConversationId) {
      selectedConversationRef.current?.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
    }
  }, [selectedConversationId])

  const handleNewConversation = async () => {
    try {
      const result = await createConversation({
        title: `Chat ${new Date().toLocaleString()}`,
      }).unwrap()
      setSelectedConversationId(result.id)
      toast.success('New conversation created')
    } catch (err) {
      console.error('Failed to create conversation:', err)
      toast.error('Failed to create conversation')
    }
  }

  const handleDeleteConversation = async (id: string) => {
    setConversationToDelete(id)
    setDeleteDialogOpen(true)
  }

  const confirmDelete = async () => {
    if (!conversationToDelete) return

    try {
      await deleteConversation(conversationToDelete).unwrap()
      if (selectedConversationId === conversationToDelete) {
        setSelectedConversationId(null)
      }
      toast.success('Conversation deleted')
    } catch (err) {
      console.error('Failed to delete conversation:', err)
      toast.error('Failed to delete conversation')
    } finally {
      setDeleteDialogOpen(false)
      setConversationToDelete(null)
    }
  }

  const handleSubmit = async (message: string) => {
    if (!selectedConversationId) {
      toast.error('Please select or create a conversation first')
      return
    }

    setError(null)
    setIsStreaming(true)
    setPendingUserMessage(message) // Show user message immediately
    setStreamingMessage('')
    setStreamingToolCalls([])

    try {
      // Stream the response - handle LangChain events
      for await (const event of sendMessageStreaming(selectedConversationId, message)) {
        if (event.type === 'langchain_event') {
          // Text streaming
          if (event.data.event === 'on_chat_model_stream') {
            const chunk = event.data.data?.chunk
            
            // Accumulate text content
            if (chunk?.content) {
              setStreamingMessage(prev => prev + chunk.content)
            }
            
            // Tool calls announced - use LLM's native IDs
            if (chunk?.tool_calls && chunk.tool_calls.length > 0) {
              setStreamingToolCalls(chunk.tool_calls.map(tc => ({
                id: tc.id,  // LLM's UUID
                name: tc.name,
                arguments: JSON.stringify(tc.args),
                status: 'loading' as const
              })))
            }
          }
          
          // Tool result arrived - match by tool_call_id
          else if (event.data.event === 'on_tool_end') {
            const toolMessage = event.data.data?.output
            if (toolMessage?.tool_call_id) {
              setStreamingToolCalls(prev =>
                prev.map(tc =>
                  tc.id === toolMessage.tool_call_id
                    ? { ...tc, status: 'complete' as const }
                    : tc
                )
              )
            }
          }
        }
        else if (event.type === 'error') {
          console.error('Stream error:', event.data?.message)
          toast.error(event.data?.message || 'An error occurred')
        }
      }

      // After streaming is complete, clear streaming state and refetch messages
      setPendingUserMessage('') // Clear pending message
      setStreamingMessage('')
      setStreamingToolCalls([])
      setIsStreaming(false)
      await refetchMessages()
    } catch (err) {
      console.error('Failed to send message:', err)
      setError('Failed to send message. Please try again.')
      toast.error('Failed to send message')
      setPendingUserMessage('') // Clear on error
      setIsStreaming(false)
      setStreamingMessage('')
      setStreamingToolCalls([])
    }
  }

  return (
    <div className="flex h-[calc(100vh-8rem)] gap-4">
      {/* Sidebar - Conversations List */}
      <div className="w-64 shrink-0">
        <Card className="h-full flex flex-col">
          <CardHeader>
            <CardTitle className="text-lg">Conversations</CardTitle>
            <Button
              onClick={handleNewConversation}
              disabled={isCreating}
              size="sm"
              className="w-full mt-2"
            >
              <PlusIcon className="w-4 h-4 mr-2" />
              New Chat
            </Button>
          </CardHeader>
          <CardContent className="flex-1 overflow-y-auto p-2">
            {conversationsLoading ? (
              <div className="text-sm text-muted-foreground text-center py-4">
                Loading...
              </div>
            ) : conversations.length === 0 ? (
              <div className="text-sm text-muted-foreground text-center py-4">
                No conversations yet
              </div>
            ) : (
              <div className="space-y-1">
                {conversations.map((conv) => (
                  <div
                    key={conv.id}
                    ref={selectedConversationId === conv.id ? selectedConversationRef : null}
                    className={`group relative flex items-center gap-2 p-2 rounded-lg cursor-pointer transition-colors ${
                      selectedConversationId === conv.id
                        ? 'bg-primary text-primary-foreground'
                        : 'hover:bg-muted'
                    }`}
                    onClick={() => setSelectedConversationId(conv.id)}
                  >
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium truncate">{conv.title}</p>
                      <p className="text-xs opacity-70 truncate">
                        {new Date(conv.created_at).toLocaleDateString()}
                      </p>
                    </div>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="opacity-0 group-hover:opacity-100 h-6 w-6"
                      onClick={(e) => {
                        e.stopPropagation()
                        handleDeleteConversation(conv.id)
                      }}
                      disabled={isDeleting}
                    >
                      <TrashIcon className="w-3 h-3" />
                    </Button>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Main Chat Area */}
      <div className="flex-1">
        <Card className="h-full flex flex-col">
          <CardHeader>
            <CardTitle>AI Assistant</CardTitle>
            <CardDescription>
              Ask questions about your Q&A knowledge base
            </CardDescription>
          </CardHeader>
          <CardContent className="flex-1 flex flex-col space-y-4 overflow-hidden">
            {/* Error message */}
            {error && (
              <Alert variant="destructive">
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}

            {/* Messages */}
            <div className="flex-1 border rounded-lg p-4 overflow-y-auto bg-muted/20">
              {!selectedConversationId ? (
                <div className="flex items-center justify-center h-full text-muted-foreground">
                  <div className="text-center">
                    <p className="text-lg font-medium mb-2">No conversation selected</p>
                    <p className="text-sm">
                      Create or select a conversation to start chatting
                    </p>
                  </div>
                </div>
              ) : messages.length === 0 && !streamingMessage ? (
                <div className="flex items-center justify-center h-full text-muted-foreground">
                  <div className="text-center">
                    <p className="text-lg font-medium mb-2">No messages yet</p>
                    <p className="text-sm">Start by asking a question below</p>
                  </div>
                </div>
              ) : (
                <>
                  {messages.map((message, index) => {
                    // Find tool messages that follow this assistant message
                    const toolMessages: typeof messages = []
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
                        key={message.id}
                        message={{
                          id: message.id,
                          role: message.role,
                          content: message.content,
                          timestamp: new Date(message.created_at),
                          tool_call_id: message.tool_call_id,
                          raw_message: message.raw_message,
                        }}
                        toolMessages={toolMessages.map(tm => ({
                          id: tm.id,
                          role: tm.role,
                          content: tm.content,
                          timestamp: new Date(tm.created_at),
                          tool_call_id: tm.tool_call_id,
                          raw_message: tm.raw_message,
                        }))}
                      />
                    )
                  })}
                  {/* Show pending user message */}
                  {pendingUserMessage && (
                    <ChatMessage
                      message={{
                        id: 'pending-user',
                        role: 'user',
                        content: pendingUserMessage,
                        timestamp: new Date(),
                        tool_call_id: null,
                        raw_message: {},
                      }}
                    />
                  )}
                  {/* Show streaming AI response */}
                  {(streamingMessage || streamingToolCalls.length > 0) && (
                    <StreamingMessage 
                      content={streamingMessage}
                      toolCalls={streamingToolCalls}
                    />
                  )}
                  <div ref={messagesEndRef} />
                </>
              )}
            </div>

            {/* Input */}
            <ChatInput
              onSubmit={handleSubmit}
              isLoading={isStreaming}
              disabled={!selectedConversationId}
            />
          </CardContent>
        </Card>
      </div>

      {/* Delete Confirmation Dialog */}
      <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Conversation</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete this conversation? This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={confirmDelete}>Delete</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}

