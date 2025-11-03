import { useState, useRef, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'

interface ChatInputProps {
  onSubmit: (message: string) => void
  isLoading?: boolean
  disabled?: boolean
}

export function ChatInput({ onSubmit, isLoading, disabled }: ChatInputProps) {
  const [message, setMessage] = useState('')
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  useEffect(() => {
    // Auto-focus the textarea
    textareaRef.current?.focus()
  }, [])

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (message.trim() && !isLoading && !disabled) {
      onSubmit(message.trim())
      setMessage('')
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSubmit(e)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="flex gap-2">
      <Textarea
        ref={textareaRef}
        value={message}
        onChange={(e) => setMessage(e.target.value)}
        onKeyDown={handleKeyDown}
        placeholder="Ask a question..."
        disabled={isLoading || disabled}
        rows={3}
        className="resize-none"
      />
      <Button
        type="submit"
        disabled={!message.trim() || isLoading || disabled}
        className="self-end"
      >
        {isLoading ? 'Sending...' : 'Send'}
      </Button>
    </form>
  )
}


