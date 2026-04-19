import { useState, useEffect } from 'react'
import { Task, listTasks, updateTaskStatus, deleteTask } from '../api/tasks'

export default function TasksPage() {
  const [tab, setTab] = useState<'done' | 'archived'>('done')
  const [tasks, setTasks] = useState<Task[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  async function load(status: 'done' | 'archived') {
    setLoading(true)
    setError('')
    try {
      const data = await listTasks({ status })
      setTasks(data)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load(tab) }, [tab])

  async function restore(task: Task) {
    try {
      await updateTaskStatus(task.id, 'pending')
      setTasks(prev => prev.filter(t => t.id !== task.id))
    } catch {
      // ignore
    }
  }

  async function remove(task: Task) {
    try {
      await deleteTask(task.id)
      setTasks(prev => prev.filter(t => t.id !== task.id))
    } catch {
      // ignore
    }
  }

  return (
    <div className="flex flex-col h-[calc(100vh-3.25rem)] page-enter">
      <div className="flex items-center gap-0.5 px-5 py-3 border-b border-ink-border">
        {(['done', 'archived'] as const).map(s => (
          <button
            key={s}
            onClick={() => setTab(s)}
            className={`px-3 py-1.5 text-sm font-medium rounded transition-colors capitalize ${
              tab === s ? 'text-gold bg-gold-glow' : 'text-cream-dim hover:text-cream hover:bg-ink-raised'
            }`}
          >
            {s === 'done' ? 'Completed' : 'Archived'}
          </button>
        ))}
      </div>

      <div className="flex-1 overflow-auto px-5 py-4">
        {loading && <p className="text-cream-faint text-sm">Loading…</p>}
        {error && <p className="text-ember text-sm">{error}</p>}
        {!loading && !error && tasks.length === 0 && (
          <p className="text-cream-faint text-sm">No {tab === 'done' ? 'completed' : 'archived'} tasks.</p>
        )}
        <div className="space-y-2">
          {tasks.map(task => (
            <div
              key={task.id}
              className="flex items-center gap-3 px-4 py-3 rounded bg-ink-surface border border-ink-border"
            >
              <div className="flex-1 min-w-0">
                <div className="text-sm text-cream truncate">{task.title}</div>
                {task.start_time && (
                  <div className="text-xs text-cream-faint mt-0.5">
                    {new Date(task.start_time).toLocaleDateString('en-US', {
                      weekday: 'short', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit', hour12: false,
                    })}
                    {task.duration_minutes ? ` · ${task.duration_minutes}m` : ''}
                  </div>
                )}
              </div>
              <div className="flex items-center gap-2 shrink-0">
                <button
                  onClick={() => restore(task)}
                  className="text-xs text-cream-dim hover:text-cream px-2 py-1 rounded hover:bg-ink-raised transition-colors"
                >
                  Restore
                </button>
                <button
                  onClick={() => remove(task)}
                  className="text-xs text-ember hover:text-ember px-2 py-1 rounded hover:bg-ember/10 transition-colors"
                >
                  Delete
                </button>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
