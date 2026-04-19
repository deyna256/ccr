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
        <p className="text-zinc-500">Loading...</p>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-full">
        <p className="text-red-400">{error}</p>
      </div>
    )
  }

  return (
    <div className="flex flex-col h-[calc(100vh-3rem)]">
      {/* Toolbar */}
      <div className="flex items-center justify-between p-4 border-b border-zinc-800">
        <div className="flex items-center gap-2">
          <button
            onClick={() => setView('week')}
            className={`px-3 py-1.5 rounded text-sm ${
              view === 'week'
                ? 'bg-zinc-700 text-white'
                : 'text-zinc-400 hover:text-white'
            }`}
          >
            Week
          </button>
          <button
            onClick={() => setView('month')}
            className={`px-3 py-1.5 rounded text-sm ${
              view === 'month'
                ? 'bg-zinc-700 text-white'
                : 'text-zinc-400 hover:text-white'
            }`}
          >
            Month
          </button>
        </div>
        <button onClick={() => setShowNewTask(true)} className="btn-primary">
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

      {/* Modals */}
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