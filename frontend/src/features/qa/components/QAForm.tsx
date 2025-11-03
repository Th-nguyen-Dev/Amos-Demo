import { useState, useEffect } from 'react'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Button } from '@/components/ui/button'
import type { QAPair, CreateQARequest } from '../types'

interface QAFormProps {
  initialData?: QAPair
  onSubmit: (data: CreateQARequest) => void
  onCancel: () => void
  isLoading?: boolean
}

export function QAForm({ initialData, onSubmit, onCancel, isLoading }: QAFormProps) {
  const [question, setQuestion] = useState('')
  const [answer, setAnswer] = useState('')
  const [errors, setErrors] = useState<{ question?: string; answer?: string }>({})

  useEffect(() => {
    if (initialData) {
      setQuestion(initialData.question)
      setAnswer(initialData.answer)
    }
  }, [initialData])

  const validate = (): boolean => {
    const newErrors: { question?: string; answer?: string } = {}

    if (!question.trim()) {
      newErrors.question = 'Question is required'
    } else if (question.length < 3) {
      newErrors.question = 'Question must be at least 3 characters'
    }

    if (!answer.trim()) {
      newErrors.answer = 'Answer is required'
    } else if (answer.length < 3) {
      newErrors.answer = 'Answer must be at least 3 characters'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (validate()) {
      onSubmit({ question: question.trim(), answer: answer.trim() })
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-2">
        <Label htmlFor="question">Question</Label>
        <Input
          id="question"
          value={question}
          onChange={(e) => setQuestion(e.target.value)}
          placeholder="Enter question..."
          disabled={isLoading}
        />
        {errors.question && (
          <p className="text-sm text-destructive">{errors.question}</p>
        )}
      </div>

      <div className="space-y-2">
        <Label htmlFor="answer">Answer</Label>
        <Textarea
          id="answer"
          value={answer}
          onChange={(e) => setAnswer(e.target.value)}
          placeholder="Enter answer..."
          rows={4}
          disabled={isLoading}
        />
        {errors.answer && (
          <p className="text-sm text-destructive">{errors.answer}</p>
        )}
      </div>

      <div className="flex justify-end gap-2">
        <Button
          type="button"
          variant="outline"
          onClick={onCancel}
          disabled={isLoading}
        >
          Cancel
        </Button>
        <Button type="submit" disabled={isLoading}>
          {isLoading ? 'Saving...' : initialData ? 'Update' : 'Create'}
        </Button>
      </div>
    </form>
  )
}

