import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { QAForm } from './QAForm'
import type { QAPair, UpdateQARequest } from '../types'

interface EditQADialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  qaPair: QAPair | null
  onSubmit: (data: UpdateQARequest) => void
  isLoading?: boolean
}

export function EditQADialog({
  open,
  onOpenChange,
  qaPair,
  onSubmit,
  isLoading,
}: EditQADialogProps) {
  if (!qaPair) return null

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Edit Q&A Pair</DialogTitle>
          <DialogDescription>
            Update the question and answer for this entry.
          </DialogDescription>
        </DialogHeader>
        <QAForm
          initialData={qaPair}
          onSubmit={onSubmit}
          onCancel={() => onOpenChange(false)}
          isLoading={isLoading}
        />
      </DialogContent>
    </Dialog>
  )
}


