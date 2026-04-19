import { useState, useEffect } from 'react'
import WeekView from '../components/WeekView'
import MonthView from '../components/MonthView'
import TaskForm from '../components/TaskForm'
import { Task, listTasks, updateTask } from '../api/tasks'
import { Category, listCategories } from '../api/categories'

type ViewMode = 'week' | 'month'

function formatPeriod(date: Date, view: ViewMode): string {
  if (view === 'month') {
    return date.toLocaleDateString('en-US', { month: 'long', year: 'numeric' })
  }
  const day = date.getDay()
  const diff = date.getDate() - day + (day === 0 ? -6 : 1)
  const mon = new Date(date)
  mon.setDate(diff)
  const sun = new Date(date)
  sun.setDate(diff + 6)
  const fmt = (d: Date) => d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
  return `${fmt(mon)} – ${fmt(sun)}, ${sun.getFullYear()}`
}

export default function CalendarPage() {
  const [view, setView] = useState<ViewMode>('week')
  const [currentDate, setCurrentDate] = useState(new Date())
  const [tasks, setTasks] = useState<Task[]>([])
  const [categories, setCategories] = useState<Category[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const [editingTask, setEditingTask] = useState<Task | null>(null)
  const [newTaskDate, setNewTaskDate] = useState<Date | null>(null)

  useEffect(() => {
    Promise.all([listTasks(), listCategories()])
      .then(([t, c]) => { setTasks(t); setCategories(c) })
      .catch(e => setError(e instanceof Error ? e.message : 'Failed to load'))
      .finally(() => setLoading(false))
  }, [])

  async function handleSaved() {
    const [t, c] = await Promise.all([listTasks(), listCategories()])
    setTasks(t)
    setCategories(c)
    setEditingTask(null)
    setNewTaskDate(null)
  }

  async function handleTaskMove(taskId: string, newStart: Date) {
    const task = tasks.find(t => t.id === taskId)
    if (!task) return

    const newStartISO = newStart.toISOString()

    // optimistic update
    setTasks(prev => prev.map(t => t.id === taskId ? { ...t, start_time: newStartISO } : t))

    try {
      await updateTask(taskId, {
        type: task.type,
        title: task.title,
        description: task.description,
        start_time: newStartISO,
        duration_minutes: task.duration_minutes,
        category_id: task.category_id,
      })
    } catch {
      // rollback
      setTasks(prev => prev.map(t => t.id === taskId ? task : t))
    }
  }

  function handleEmptySlotClick(date: Date) {
    setNewTaskDate(date)
  }

  function navigate(dir: -1 | 1) {
    setCurrentDate(prev => {
      const d = new Date(prev)
      if (view === 'week') d.setDate(d.getDate() + dir * 7)
      else d.setMonth(d.getMonth() + dir)
      return d
    })
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
        <div className="flex items-center gap-3">
          {/* View toggle */}
          <div className="flex items-center gap-0.5">
            {(['week', 'month'] as ViewMode[]).map(v => (
              <button
                key={v}
                onClick={() => setView(v)}
                className={`px-3 py-1.5 text-sm font-medium rounded transition-colors capitalize ${
                  view === v ? 'text-gold bg-gold-glow' : 'text-cream-dim hover:text-cream hover:bg-ink-raised'
                }`}
              >
                {v}
              </button>
            ))}
          </div>

          {/* Navigation */}
          <div className="flex items-center gap-1">
            <button
              onClick={() => navigate(-1)}
              className="px-2 py-1.5 text-cream-dim hover:text-cream hover:bg-ink-raised rounded transition-colors text-sm"
            >
              ←
            </button>
            <button
              onClick={() => setCurrentDate(new Date())}
              className="px-3 py-1.5 text-xs text-cream-faint hover:text-cream hover:bg-ink-raised rounded transition-colors"
            >
              Today
            </button>
            <button
              onClick={() => navigate(1)}
              className="px-2 py-1.5 text-cream-dim hover:text-cream hover:bg-ink-raised rounded transition-colors text-sm"
            >
              →
            </button>
          </div>

          <span className="text-sm text-cream-dim select-none">{formatPeriod(currentDate, view)}</span>
        </div>

        <button onClick={() => setNewTaskDate(new Date())} className="btn-primary py-1.5">
          + New Task
        </button>
      </div>

      {/* Calendar */}
      {view === 'week' ? (
        <WeekView
          currentDate={currentDate}
          tasks={tasks}
          categories={categories}
          onTaskClick={t => setEditingTask(t)}
          onEmptySlotClick={handleEmptySlotClick}
          onTaskMove={handleTaskMove}
        />
      ) : (
        <MonthView
          currentDate={currentDate}
          tasks={tasks}
          categories={categories}
          onTaskClick={t => setEditingTask(t)}
          onEmptySlotClick={handleEmptySlotClick}
        />
      )}

      {newTaskDate && (
        <TaskForm
          initialDate={newTaskDate}
          categories={categories}
          onClose={() => setNewTaskDate(null)}
          onSaved={handleSaved}
        />
      )}

      {editingTask && (
        <TaskForm
          task={editingTask}
          categories={categories}
          onClose={() => setEditingTask(null)}
          onSaved={handleSaved}
        />
      )}
    </div>
  )
}
