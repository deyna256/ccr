import { useState } from 'react'
import { Task, TaskType, CreateTaskRequest, createTask, updateTask, deleteTask, updateTaskStatus } from '../api/tasks'
import { Category } from '../api/categories'

interface TaskFormProps {
  task?: Task
  categories: Category[]
  onClose: () => void
  onSaved: () => void
}

export default function TaskForm({ task, categories, onClose, onSaved }: TaskFormProps) {
  const [type, setType] = useState<TaskType>(task?.type ?? 'task')
  const [title, setTitle] = useState(task?.title ?? '')
  const [description, setDescription] = useState(task?.description ?? '')
  const [date, setDate] = useState(
    task?.start_time ? task.start_time.split('T')[0] : new Date().toISOString().split('T')[0],
  )
  const [startTime, setStartTime] = useState(
    task?.start_time ? task.start_time.split('T')[1]?.slice(0, 5) : '09:00',
  )
  const [duration, setDuration] = useState(task?.duration_minutes?.toString() ?? '60')
  const [categoryId, setCategoryId] = useState(task?.category_id ?? '')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    setSaving(true)

    const data: CreateTaskRequest = {
      type,
      title,
      description,
      start_time: `${date}T${startTime}:00`,
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
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4">
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg w-full max-w-md">
        <div className="flex items-center justify-between p-4 border-b border-zinc-800">
          <h2 className="text-lg font-semibold text-white">
            {task ? 'Edit Task' : 'New Task'}
          </h2>
          <button onClick={onClose} className="text-zinc-500 hover:text-white">
            &times;
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-4 space-y-4">
          {error && (
            <p className="text-sm text-red-400 bg-red-900/20 p-2 rounded">{error}</p>
          )}

          <div className="flex gap-2">
            <button
              type="button"
              onClick={() => setType('task')}
              className={`flex-1 py-2 rounded-md text-sm transition-colors ${
                type === 'task'
                  ? 'bg-blue-600 text-white'
                  : 'bg-zinc-800 text-zinc-400 hover:text-white'
              }`}
            >
              Task
            </button>
            <button
              type="button"
              onClick={() => setType('event')}
              className={`flex-1 py-2 rounded-md text-sm transition-colors ${
                type === 'event'
                  ? 'bg-blue-600 text-white'
                  : 'bg-zinc-800 text-zinc-400 hover:text-white'
              }`}
            >
              Event
            </button>
          </div>

          <div>
            <label className="block text-sm text-zinc-400 mb-1">Title</label>
            <input
              type="text"
              value={title}
              onChange={e => setTitle(e.target.value)}
              className="input-field"
              required
            />
          </div>

          <div>
            <label className="block text-sm text-zinc-400 mb-1">Description</label>
            <textarea
              value={description}
              onChange={e => setDescription(e.target.value)}
              className="input-field resize-none"
              rows={2}
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm text-zinc-400 mb-1">Date</label>
              <input
                type="date"
                value={date}
                onChange={e => setDate(e.target.value)}
                className="input-field"
                required
              />
            </div>
            <div>
              <label className="block text-sm text-zinc-400 mb-1">Start Time</label>
              <input
                type="time"
                value={startTime}
                onChange={e => setStartTime(e.target.value)}
                className="input-field"
                required
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm text-zinc-400 mb-1">
                Duration (minutes)
              </label>
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
              <label className="block text-sm text-zinc-400 mb-1">Category</label>
              <select
                value={categoryId}
                onChange={e => setCategoryId(e.target.value)}
                className="input-field"
              >
                <option value="">None</option>
                {categories.map(c => (
                  <option key={c.id} value={c.id}>
                    {c.name}
                  </option>
                ))}
              </select>
            </div>
          </div>

          <div className="flex gap-2 pt-2">
            <button
              type="submit"
              disabled={saving}
              className="btn-primary flex-1"
            >
              {saving ? 'Saving...' : 'Save'}
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
            <div className="flex gap-2 pt-2 border-t border-zinc-800">
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