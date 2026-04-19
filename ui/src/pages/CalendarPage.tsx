import { useState, useEffect } from 'react'
import WeekView from '../components/WeekView'
import MonthView from '../components/MonthView'
import TaskForm from '../components/TaskForm'
import { Task, listTasks } from '../api/tasks'
import { Category, listCategories } from '../api/categories'

type ViewMode = 'week' | 'month'

export default function CalendarPage() {
  const [view, setView] = useState<ViewMode>('week')
  const [tasks, setTasks] = useState<Task[]>([])
  const [categories, setCategories] = useState<Category[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const [editingTask, setEditingTask] = useState<Task | null>(null)
  const [showNewTask, setShowNewTask] = useState(false)

  useEffect(() => {
    Promise.all([listTasks(), listCategories()])
      .then(([t, c]) => {
        setTasks(t)
        setCategories(c)
      })
      .catch(e => setError(e instanceof Error ? e.message : 'Failed to load'))
      .finally(() => setLoading(false))
  }, [])

  async function handleSaved() {
    const [t, c] = await Promise.all([listTasks(), listCategories()])
    setTasks(t)
    setCategories(c)
    setEditingTask(null)
    setShowNewTask(false)
  }

  function handleEmptySlotClick(_date: Date) {
    setShowNewTask(true)
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <p className="text-cream-faint text-sm">Loading…</p>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-full">
        <p className="text-ember text-sm">{error}</p>
      </div>
    )
  }

  return (
    <div className="flex flex-col h-[calc(100vh-3.25rem)] page-enter">
      {/* Toolbar */}
      <div className="flex items-center justify-between px-5 py-3 border-b border-ink-border">
        <div className="flex items-center gap-0.5">
          {(['week', 'month'] as ViewMode[]).map(v => (
            <button
              key={v}
              onClick={() => setView(v)}
              className={`px-3 py-1.5 text-sm font-medium rounded transition-colors capitalize ${
                view === v
                  ? 'text-gold bg-gold-glow'
                  : 'text-cream-dim hover:text-cream hover:bg-ink-raised'
              }`}
            >
              {v}
            </button>
          ))}
        </div>
        <button onClick={() => setShowNewTask(true)} className="btn-primary py-1.5">
          + New Task
        </button>
      </div>

      {/* Calendar */}
      {view === 'week' ? (
        <WeekView
          tasks={tasks}
          categories={categories}
          onTaskClick={t => setEditingTask(t)}
          onEmptySlotClick={handleEmptySlotClick}
        />
      ) : (
        <MonthView
          tasks={tasks}
          categories={categories}
          onTaskClick={t => setEditingTask(t)}
          onEmptySlotClick={handleEmptySlotClick}
        />
      )}

      {(showNewTask || editingTask) && (
        <TaskForm
          task={editingTask ?? undefined}
          categories={categories}
          onClose={() => {
            setEditingTask(null)
            setShowNewTask(false)
          }}
          onSaved={handleSaved}
        />
      )}
    </div>
  )
}
