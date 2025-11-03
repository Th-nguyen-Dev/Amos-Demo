import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import type { QAPair } from '../types'

interface DeleteQADialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  qaPair: QAPair | null
  onConfirm: () => void
  isLoading?: boolean
}

export function DeleteQADialog({
  open,
  onOpenChange,
  qaPair,
  onConfirm,
  isLoading,
}: DeleteQADialogProps) {
  if (!qaPair) return null

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Are you sure?</AlertDialogTitle>
          <AlertDialogDescription>
            This will permanently delete the Q&A pair:
            <br />
            <br />
            <strong>&quot;{qaPair.question}&quot;</strong>
            <br />
            <br />
            This action cannot be undone.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={isLoading}>Cancel</AlertDialogCancel>
          <AlertDialogAction onClick={onConfirm} disabled={isLoading}>
            {isLoading ? 'Deleting...' : 'Delete'}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}


