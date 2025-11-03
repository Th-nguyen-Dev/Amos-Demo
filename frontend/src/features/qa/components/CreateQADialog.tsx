import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { QAForm } from './QAForm'
import type { CreateQARequest } from '../types'

interface CreateQADialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSubmit: (data: CreateQARequest) => void
  isLoading?: boolean
}

export function CreateQADialog({
  open,
  onOpenChange,
  onSubmit,
  isLoading,
}: CreateQADialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create New Q&A Pair</DialogTitle>
          <DialogDescription>
            Add a new question and answer to your knowledge base.
          </DialogDescription>
        </DialogHeader>
        <QAForm
          onSubmit={onSubmit}
          onCancel={() => onOpenChange(false)}
          isLoading={isLoading}
        />
      </DialogContent>
    </Dialog>
  )
}


