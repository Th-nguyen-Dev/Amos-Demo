import { useState, useCallback, useMemo, useEffect } from 'react'
import { toast } from 'sonner'
import { useListQAPairsQuery, useCreateQAPairMutation, useUpdateQAPairMutation, useDeleteQAPairMutation } from '../api/qaApi'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { QATable } from '../components/QATable'
import { CreateQADialog } from '../components/CreateQADialog'
import { EditQADialog } from '../components/EditQADialog'
import { DeleteQADialog } from '../components/DeleteQADialog'
import type { QAPair, CreateQARequest, UpdateQARequest } from '../types'

// Debounce hook
function useDebounce<T>(value: T, delay: number): T {
  const [debouncedValue, setDebouncedValue] = useState<T>(value)

  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedValue(value)
    }, delay)

    return () => {
      clearTimeout(handler)
    }
  }, [value, delay])

  return debouncedValue
}

export function QAManagementPage() {
  // State
  const [search, setSearch] = useState('')
  const [cursor, setCursor] = useState<string | undefined>(undefined)
  const [cursorHistory, setCursorHistory] = useState<(string | undefined)[]>([])
  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  const [editDialogOpen, setEditDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [selectedQAPair, setSelectedQAPair] = useState<QAPair | null>(null)

  // Debounce search
  const debouncedSearch = useDebounce(search, 300)

  // Reset pagination when search changes
  useEffect(() => {
    setCursor(undefined)
    setCursorHistory([])
  }, [debouncedSearch])

  // Query params - always forward pagination, direction not needed
  const queryParams = useMemo(() => ({
    limit: 10,
    cursor,
    direction: 'next' as const,
    search: debouncedSearch || undefined,
  }), [cursor, debouncedSearch])

  // API hooks
  const { data, isLoading, error } = useListQAPairsQuery(queryParams, {
    refetchOnMountOrArgChange: 1, // Refetch if data is older than 1 second
    refetchOnFocus: true,
    refetchOnReconnect: true,
  })
  const [createQAPair, { isLoading: isCreating }] = useCreateQAPairMutation()
  const [updateQAPair, { isLoading: isUpdating }] = useUpdateQAPairMutation()
  const [deleteQAPair, { isLoading: isDeleting }] = useDeleteQAPairMutation()

  // Handlers
  const handleCreate = useCallback(async (formData: CreateQARequest) => {
    try {
      await createQAPair(formData).unwrap()
      setCreateDialogOpen(false)
      // Reset to first page after creating
      setCursor(undefined)
      setCursorHistory([])
      toast.success('Q&A pair created successfully')
    } catch (err) {
      console.error('Failed to create Q&A pair:', err)
      toast.error('Failed to create Q&A pair')
    }
  }, [createQAPair])

  const handleEdit = useCallback((qaPair: QAPair) => {
    setSelectedQAPair(qaPair)
    setEditDialogOpen(true)
  }, [])

  const handleUpdate = useCallback(async (formData: UpdateQARequest) => {
    if (!selectedQAPair) return

    try {
      await updateQAPair({ id: selectedQAPair.id, data: formData }).unwrap()
      setEditDialogOpen(false)
      setSelectedQAPair(null)
      toast.success('Q&A pair updated successfully')
    } catch (err) {
      console.error('Failed to update Q&A pair:', err)
      toast.error('Failed to update Q&A pair')
    }
  }, [selectedQAPair, updateQAPair])

  const handleDelete = useCallback((qaPair: QAPair) => {
    setSelectedQAPair(qaPair)
    setDeleteDialogOpen(true)
  }, [])

  const handleDeleteConfirm = useCallback(async () => {
    if (!selectedQAPair) return

    try {
      await deleteQAPair(selectedQAPair.id).unwrap()
      setDeleteDialogOpen(false)
      setSelectedQAPair(null)
      // Reset to first page after deleting
      setCursor(undefined)
      setCursorHistory([])
      toast.success('Q&A pair deleted successfully')
    } catch (err) {
      console.error('Failed to delete Q&A pair:', err)
      toast.error('Failed to delete Q&A pair')
    }
  }, [selectedQAPair, deleteQAPair])

  const handleNextPage = useCallback(() => {
    if (data?.pagination.next_cursor) {
      // Push current cursor to history before moving forward
      setCursorHistory(prev => [...prev, cursor])
      setCursor(data.pagination.next_cursor)
    }
  }, [data, cursor])

  const handlePrevPage = useCallback(() => {
    if (cursorHistory.length > 0) {
      // Pop from history to go back
      const newHistory = [...cursorHistory]
      const previousCursor = newHistory.pop()
      setCursorHistory(newHistory)
      setCursor(previousCursor)
    }
  }, [cursorHistory])

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>Q&A Knowledge Base</CardTitle>
          <CardDescription>
            Manage your question and answer pairs for the AI assistant
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Search and Create */}
          <div className="flex gap-4">
            <Input
              placeholder="Search questions and answers..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="flex-1"
            />
            <Button onClick={() => setCreateDialogOpen(true)}>
              Create New
            </Button>
          </div>

          {/* Error message */}
          {error && (
            <Alert variant="destructive">
              <AlertDescription>
                Failed to load Q&A pairs. Please try again.
              </AlertDescription>
            </Alert>
          )}

          {/* Table */}
          <QATable
            qaPairs={data?.data || []}
            onEdit={handleEdit}
            onDelete={handleDelete}
            isLoading={isLoading}
          />

          {/* Pagination */}
          {data && (data.data.length > 0 || cursorHistory.length > 0) && (
            <div className="flex justify-between items-center">
              <div className="text-sm text-muted-foreground">
                {data.data.length} Q&A pair{data.data.length !== 1 ? 's' : ''} displayed
              </div>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  onClick={handlePrevPage}
                  disabled={cursorHistory.length === 0}
                >
                  Previous
                </Button>
                <Button
                  variant="outline"
                  onClick={handleNextPage}
                  disabled={!data.pagination.has_next}
                >
                  Next
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Dialogs */}
      <CreateQADialog
        open={createDialogOpen}
        onOpenChange={setCreateDialogOpen}
        onSubmit={handleCreate}
        isLoading={isCreating}
      />

      <EditQADialog
        open={editDialogOpen}
        onOpenChange={setEditDialogOpen}
        qaPair={selectedQAPair}
        onSubmit={handleUpdate}
        isLoading={isUpdating}
      />

      <DeleteQADialog
        open={deleteDialogOpen}
        onOpenChange={setDeleteDialogOpen}
        qaPair={selectedQAPair}
        onConfirm={handleDeleteConfirm}
        isLoading={isDeleting}
      />
    </div>
  )
}

