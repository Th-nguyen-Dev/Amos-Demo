import { cn } from '@/lib/utils'
import type { ChatMessage as ChatMessageType } from '../types'

interface ChatMessageProps {
  message: ChatMessageType
}

export function ChatMessage({ message }: ChatMessageProps) {
  const isUser = message.role === 'user'

  return (
    <div
      className={cn(
        'flex w-full gap-3 mb-4',
        isUser ? 'justify-end' : 'justify-start'
      )}
    >
      <div
        className={cn(
          'max-w-[80%] rounded-lg px-4 py-3',
          isUser
            ? 'bg-primary text-primary-foreground'
            : 'bg-muted'
        )}
      >
        <div className="text-sm font-medium mb-1">
          {isUser ? 'You' : 'AI Assistant'}
        </div>
        <div className="text-sm whitespace-pre-wrap">
          {message.content}
        </div>
        <div className="text-xs mt-2 opacity-70">
          {new Date(message.timestamp).toLocaleTimeString()}
        </div>
      </div>
    </div>
  )
}


