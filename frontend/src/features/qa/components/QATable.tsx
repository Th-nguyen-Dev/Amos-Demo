import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import type { QAPair } from '../types'

interface QATableProps {
  qaPairs: QAPair[]
  onEdit: (qaPair: QAPair) => void
  onDelete: (qaPair: QAPair) => void
  isLoading?: boolean
}

export function QATable({ qaPairs, onEdit, onDelete, isLoading }: QATableProps) {
  if (isLoading) {
    return (
      <div className="text-center py-8 text-muted-foreground">
        Loading Q&A pairs...
      </div>
    )
  }

  if (qaPairs.length === 0) {
    return (
      <div className="text-center py-8 text-muted-foreground">
        No Q&A pairs found. Create your first one!
      </div>
    )
  }

  return (
    <div className="border rounded-lg">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-[200px]">Question</TableHead>
            <TableHead>Answer</TableHead>
            <TableHead className="w-[100px] text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {qaPairs.map((qaPair) => (
            <TableRow key={qaPair.id}>
              <TableCell className="font-medium max-w-[200px] truncate">
                {qaPair.question}
              </TableCell>
              <TableCell className="max-w-[400px] truncate">
                {qaPair.answer}
              </TableCell>
              <TableCell className="text-right">
                <div className="flex justify-end gap-2">
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => onEdit(qaPair)}
                  >
                    Edit
                  </Button>
                  <Button
                    size="sm"
                    variant="destructive"
                    onClick={() => onDelete(qaPair)}
                  >
                    Delete
                  </Button>
                </div>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}


