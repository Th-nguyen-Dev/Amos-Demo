import { useState, useRef, useEffect } from 'react'
import { toast } from 'sonner'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { ChatMessage } from '../components/ChatMessage'
import { ChatInput } from '../components/ChatInput'
import { useAskQuestionMutation } from '../api/chatApi'
import type { ChatMessage as ChatMessageType } from '../types'

export function ChatPage() {
  const [messages, setMessages] = useState<ChatMessageType[]>([])
  const [error, setError] = useState<string | null>(null)
  const messagesEndRef = useRef<HTMLDivElement>(null)

  const [askQuestion, { isLoading }] = useAskQuestionMutation()

  // Scroll to bottom when messages change
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  const handleSubmit = async (question: string) => {
    // Clear any previous errors
    setError(null)

    // Add user message
    const userMessage: ChatMessageType = {
      id: Date.now().toString(),
      role: 'user',
      content: question,
      timestamp: new Date(),
    }
    setMessages((prev) => [...prev, userMessage])

    try {
      // Call the API
      const response = await askQuestion({ question }).unwrap()

      // Add assistant message
      const assistantMessage: ChatMessageType = {
        id: (Date.now() + 1).toString(),
        role: 'assistant',
        content: response.answer,
        timestamp: new Date(),
      }
      setMessages((prev) => [...prev, assistantMessage])
    } catch (err) {
      console.error('Failed to get answer:', err)
      setError('Failed to get an answer. Please try again.')
      toast.error('Failed to get an answer from the AI assistant')
      
      // Remove the user message since the request failed
      setMessages((prev) => prev.slice(0, -1))
    }
  }

  const handleClear = () => {
    setMessages([])
    setError(null)
    toast.success('Chat cleared')
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <div className="flex justify-between items-start">
            <div>
              <CardTitle>Ask the AI Assistant</CardTitle>
              <CardDescription>
                Ask questions about your Q&A knowledge base
              </CardDescription>
            </div>
            {messages.length > 0 && (
              <Button variant="outline" onClick={handleClear}>
                Clear Chat
              </Button>
            )}
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Error message */}
          {error && (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          {/* Messages */}
          <div className="border rounded-lg p-4 min-h-[400px] max-h-[600px] overflow-y-auto bg-muted/20">
            {messages.length === 0 ? (
              <div className="flex items-center justify-center h-[400px] text-muted-foreground">
                <div className="text-center">
                  <p className="text-lg font-medium mb-2">
                    No messages yet
                  </p>
                  <p className="text-sm">
                    Start by asking a question below
                  </p>
                </div>
              </div>
            ) : (
              <>
                {messages.map((message) => (
                  <ChatMessage key={message.id} message={message} />
                ))}
                <div ref={messagesEndRef} />
              </>
            )}
          </div>

          {/* Input */}
          <ChatInput
            onSubmit={handleSubmit}
            isLoading={isLoading}
          />
        </CardContent>
      </Card>
    </div>
  )
}

