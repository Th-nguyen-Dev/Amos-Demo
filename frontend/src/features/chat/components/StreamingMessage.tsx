import { Bot } from 'lucide-react'
import { Tool, ToolHeader, ToolContent } from '@/components/ai/tool'
import { cn } from '@/lib/utils'

interface StreamingMessageProps {
  content: string
  toolCalls: Array<{ name: string, status: 'loading' | 'complete' }>
}

export function StreamingMessage({ content, toolCalls }: StreamingMessageProps) {
  return (
    <div className="flex w-full gap-3 mb-6 justify-start">
      {/* Avatar for assistant */}
      <div className="shrink-0 w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center">
        <Bot className="w-4 h-4 text-primary" />
      </div>

      <div className="max-w-[85%] flex flex-col gap-3">
        {/* Tool calls with pulsing animation */}
        {toolCalls.length > 0 && (
          <div className="flex flex-col gap-2">
            {toolCalls.map((toolCall, index) => (
              <div
                key={index}
                className={cn(
                  "transition-all duration-500",
                  toolCall.status === 'loading' && "animate-pulse"
                )}
                style={{
                  animation: toolCall.status === 'complete' 
                    ? 'fadeOut 0.5s ease-out forwards' 
                    : undefined
                }}
              >
                <Tool 
                  status="loading" 
                  defaultOpen={true}
                >
                  <ToolHeader status="loading">
                    🔧 {toolCall.name}
                  </ToolHeader>
                  <ToolContent>
                    <div className="px-3 py-2 text-xs text-muted-foreground flex items-center gap-2">
                      <div className="w-2 h-2 bg-blue-500 rounded-full animate-pulse" />
                      Executing...
                    </div>
                  </ToolContent>
                </Tool>
              </div>
            ))}
          </div>
        )}

        {/* Streaming text content */}
        {content && (
          <div className="rounded-lg bg-muted/50 px-4 py-3">
            <div className="flex items-center gap-2 mb-2">
              <span className="text-xs font-semibold">AI Assistant</span>
              <span className="text-xs opacity-60">
                {new Date().toLocaleTimeString()}
              </span>
            </div>

            <div className="text-sm whitespace-pre-wrap">
              {content}
              <span className="inline-block w-1 h-4 ml-1 bg-primary animate-pulse" />
            </div>
          </div>
        )}
      </div>
    </div>
  )
}



