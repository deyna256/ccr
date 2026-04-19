import { useState } from 'react'
import { Task, TaskType, CreateTaskRequest, createTask, updateTask, deleteTask, updateTaskStatus } from '../api/tasks'
import { Category } from '../api/categories'

interface TaskFormProps {
  task?: Task
  categories: Category[]
  initialDate?: Date
  onClose: () => void
  onSaved: () => void
}

function toDateInput(d: Date): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

function toTimeInput(d: Date): string {
  return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

export default function TaskForm({ task, categories, initialDate, onClose, onSaved }: TaskFormProps) {
  const [type, setType] = useState<TaskType>(task?.type ?? 'task')
  const [title, setTitle] = useState(task?.title ?? '')
  const [description, setDescription] = useState(task?.description ?? '')

  const startDate = task?.start_time ? new Date(task.start_time) : null

  const [date, setDate] = useState(
    startDate
      ? toDateInput(startDate)
      : initialDate
        ? toDateInput(initialDate)
        : toDateInput(new Date()),
  )
  const [startTime, setStartTime] = useState(
    startDate ? toTimeInput(startDate) : '09:00',
  )
  const [duration, setDuration] = useState(task?.duration_minutes?.toString() ?? '60')
  const [categoryId, setCategoryId] = useState(task?.category_id ?? '')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  const dateFixed = !task && !!initialDate

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    setSaving(true)

    const data: CreateTaskRequest = {
      type,
      title,
      description,
      start_time: new Date(`${date}T${startTime}:00`).toISOString(),
      duration_minutes: parseInt(duration, 10),
      category_id: categoryId || undefined,
    }

    try {
      if (task) {
        await updateTask(task.id, data)
      } else {
        await createTask(data)
      }
      onSaved()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save')
    } finally {
      setSaving(false)
    }
  }

  async function handleStatusChange(next: 'done' | 'archived') {
    if (!task) return
    setSaving(true)
    try {
      await updateTaskStatus(task.id, next)
      onSaved()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed')
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete() {
    if (!task) return
    setSaving(true)
    try {
      await deleteTask(task.id)
      onSaved()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="modal-backdrop">
      <div className="modal-panel max-w-md">
        <div className="flex items-center justify-between px-5 py-4 border-b border-ink-border">
          <h2 className="text-base font-semibold text-cream">
            {task ? 'Edit Task' : 'New Task'}
          </h2>
          <button onClick={onClose} className="text-cream-faint hover:text-cream transition-colors text-lg leading-none">
            ×
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-5 space-y-4">
          {error && (
            <div className="text-xs text-ember border border-ember/20 bg-ember/5 px-4 py-3 rounded">
              {error}
            </div>
          )}

          <div className="flex gap-1.5">
            {(['task', 'event'] as TaskType[]).map(t => (
              <button
                key={t}
                type="button"
                onClick={() => setType(t)}
                className={`flex-1 py-2 rounded text-sm font-medium transition-colors capitalize ${
                  type === t ? 'bg-gold text-ink' : 'bg-ink-raised text-cream-dim hover:text-cream'
                }`}
              >
                {t}
              </button>
            ))}
          </div>

          <div>
            <label className="block text-xs text-cream-faint mb-1.5">Title</label>
            <input
              type="text"
              value={title}
              onChange={e => setTitle(e.target.value)}
              className="input-field"
              required
              autoFocus
            />
          </div>

          <div>
            <label className="block text-xs text-cream-faint mb-1.5">Description</label>
            <textarea
              value={description}
              onChange={e => setDescription(e.target.value)}
              className="input-field resize-none"
              rows={2}
            />
          </div>

          <div className={`grid gap-3 ${dateFixed ? 'grid-cols-1' : 'grid-cols-2'}`}>
            {dateFixed ? (
              <div>
                <label className="block text-xs text-cream-faint mb-1.5">Time</label>
                <div className="flex items-center gap-2">
                  <span className="text-xs text-cream-dim border border-ink-border rounded px-3 py-2 bg-ink-raised select-none">
                    {new Date(date).toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' })}
                  </span>
                  <input
                    type="time"
                    value={startTime}
                    onChange={e => setStartTime(e.target.value)}
                    className="input-field flex-1"
                    required
                  />
                </div>
              </div>
            ) : (
              <>
                <div>
                  <label className="block text-xs text-cream-faint mb-1.5">Date</label>
                  <input
                    type="date"
                    value={date}
                    onChange={e => setDate(e.target.value)}
                    className="input-field"
                    required
                  />
                </div>
                <div>
                  <label className="block text-xs text-cream-faint mb-1.5">Start Time</label>
                  <input
                    type="time"
                    value={startTime}
                    onChange={e => setStartTime(e.target.value)}
                    className="input-field"
                    required
                  />
                </div>
              </>
            )}
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-xs text-cream-faint mb-1.5">Duration (min)</label>
              <input
                type="number"
                value={duration}
                onChange={e => setDuration(e.target.value)}
                className="input-field"
                min="15"
                step="15"
              />
            </div>
            <div>
              <label className="block text-xs text-cream-faint mb-1.5">Category</label>
              <select
                value={categoryId}
                onChange={e => setCategoryId(e.target.value)}
                className="input-field"
              >
                <option value="">None</option>
                {categories.map(c => (
                  <option key={c.id} value={c.id}>{c.name}</option>
                ))}
              </select>
            </div>
          </div>

          <div className="flex gap-2 pt-1">
            <button type="submit" disabled={saving} className="btn-primary flex-1">
              {saving ? 'Saving…' : 'Save'}
            </button>
            {task && (
              <button
                type="button"
                onClick={() => handleStatusChange('done')}
                disabled={saving || task.status === 'done'}
                className="btn-ghost"
              >
                Done
              </button>
            )}
          </div>

          {task && (
            <div className="flex gap-2 pt-2 border-t border-ink-border">
              <button
                type="button"
                onClick={() => handleStatusChange('archived')}
                disabled={saving || task.status === 'archived'}
                className="btn-ghost flex-1"
              >
                Archive
              </button>
              <button
                type="button"
                onClick={handleDelete}
                disabled={saving}
                className="btn-danger"
              >
                Delete
              </button>
            </div>
          )}
        </form>
      </div>
    </div>
  )
}
