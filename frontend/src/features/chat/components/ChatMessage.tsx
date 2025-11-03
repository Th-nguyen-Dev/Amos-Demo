import { cn } from '@/lib/utils'
import type { ChatMessage as ChatMessageType } from '../types'
import { Bot, User } from 'lucide-react'
import { Tool, ToolHeader, ToolContent, ToolInput, ToolOutput } from '@/components/ai/tool'
import type { ToolCall } from '@/types/models'

interface ChatMessageProps {
  message: ChatMessageType
  toolMessages?: ChatMessageType[] // Tool result messages
}

export function ChatMessage({ message, toolMessages = [] }: ChatMessageProps) {
  const isUser = message.role === 'user'
  const isTool = message.role === 'tool'
  
  // Skip rendering tool messages (they're shown with their tool calls)
  if (isTool) {
    return null
  }
  
  // Parse tool calls from raw_message
  const toolCalls = (message.raw_message as { tool_calls?: ToolCall[] })?.tool_calls

  return (
    <div
      className={cn(
        'flex w-full gap-3 mb-6',
        isUser ? 'justify-end' : 'justify-start'
      )}
    >
      {/* Avatar for assistant */}
      {!isUser && (
        <div className="shrink-0 w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center">
          <Bot className="w-4 h-4 text-primary" />
        </div>
      )}

      <div className="max-w-[85%] flex flex-col gap-3">
        {/* Message content */}
        {message.content && (
          <div
            className={cn(
              'rounded-lg',
              isUser
                ? 'bg-primary text-primary-foreground px-4 py-3'
                : 'bg-muted/50 px-4 py-3'
            )}
          >
            <div className="flex items-center gap-2 mb-2">
              <span className="text-xs font-semibold">
                {isUser ? 'You' : 'AI Assistant'}
              </span>
              <span className="text-xs opacity-60">
                {new Date(message.timestamp).toLocaleTimeString()}
              </span>
            </div>

            <div className="text-sm whitespace-pre-wrap">
              {message.content}
            </div>
          </div>
        )}
        
        {/* Tool calls - rendered from JSONB raw_message */}
        {toolCalls && toolCalls.length > 0 && (
          <div className="flex flex-col gap-2">
            {toolCalls.map((toolCall, index) => {
              // Find the corresponding tool result message
              const toolResult = toolMessages.find(
                (msg) => msg.role === 'tool' && msg.tool_call_id === toolCall.id
              )
              
              let args: Record<string, unknown> = {}
              try {
                args = JSON.parse(toolCall.function.arguments)
              } catch {
                args = { arguments: toolCall.function.arguments }
              }
              
              const isSuccess = toolResult?.content && 
                              !toolResult.content.toLowerCase().includes('no relevant') &&
                              !toolResult.content.toLowerCase().includes('not found') &&
                              !toolResult.content.toLowerCase().includes('error')
              
              return (
                <Tool 
                  key={toolCall.id} 
                  status={toolResult ? (isSuccess ? "success" : "error") : "loading"}
                  defaultOpen={index === toolCalls.length - 1}
                >
                  <ToolHeader status={toolResult ? (isSuccess ? "success" : "error") : "loading"}>
                    🔧 {toolCall.function.name}
                  </ToolHeader>
                  <ToolContent>
                    <ToolInput>
                      <pre className="text-xs overflow-x-auto">
                        {JSON.stringify(args, null, 2)}
                      </pre>
                    </ToolInput>
                    {toolResult && (
                      <ToolOutput>
                        <div className="text-xs whitespace-pre-wrap max-h-40 overflow-y-auto">
                          {toolResult.content?.slice(0, 300)}
                          {toolResult.content && toolResult.content.length > 300 && '... (truncated)'}
                        </div>
                      </ToolOutput>
                    )}
                  </ToolContent>
                </Tool>
              )
            })}
          </div>
        )}
      </div>

      {/* Avatar for user */}
      {isUser && (
        <div className="shrink-0 w-8 h-8 rounded-full bg-primary flex items-center justify-center">
          <User className="w-4 h-4 text-primary-foreground" />
        </div>
      )}
    </div>
  )
}


