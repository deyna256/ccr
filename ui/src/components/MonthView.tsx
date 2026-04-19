import { useMemo } from 'react'
import { Task } from '../api/tasks'
import { Category } from '../api/categories'

interface MonthViewProps {
  currentDate: Date
  tasks: Task[]
  categories: Category[]
  onTaskClick: (task: Task) => void
  onEmptySlotClick: (date: Date) => void
}

export default function MonthView({ currentDate, tasks, categories, onTaskClick, onEmptySlotClick }: MonthViewProps) {
  const today = new Date()
  const year = currentDate.getFullYear()
  const month = currentDate.getMonth()

  const days = useMemo(() => {
    const firstDay = new Date(year, month, 1)
    const lastDay = new Date(year, month + 1, 0)
    const startPadding = (firstDay.getDay() + 6) % 7

    return Array.from({ length: startPadding + lastDay.getDate() }, (_, i) => {
      const d = new Date(year, month, i - startPadding + 1)
      return d
    })
  }, [year, month])

  const tasksByDay = useMemo(() => {
    const map = new Map<string, Task[]>()
    for (const task of tasks) {
      const date = task.start_time.split('T')[0]
      const list = map.get(date) ?? []
      list.push(task)
      map.set(date, list)
    }
    return map
  }, [tasks])

  const getCategoryColor = (id?: string) => {
    if (!id) return '#57535e'
    return categories.find(c => c.id === id)?.color ?? '#57535e'
  }

  return (
    <div className="flex-1 overflow-auto">
      <div className="grid grid-cols-7 gap-px bg-ink-border">
        {['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'].map(day => (
          <div key={day} className="bg-ink p-2 text-center text-xs text-cream-faint uppercase tracking-wider">
            {day}
          </div>
        ))}
        {days.map(day => {
          const dateStr = day.toISOString().split('T')[0]
          const dayTasks = tasksByDay.get(dateStr) ?? []
          const isCurrentMonth = day.getMonth() === month
          const isToday = isCurrentMonth && day.toDateString() === today.toDateString()

          return (
            <div
              key={dateStr}
              className={`bg-ink-surface min-h-[100px] p-1.5 ${isCurrentMonth ? '' : 'opacity-25'}`}
            >
              <button onClick={() => onEmptySlotClick(day)} className="w-full text-left">
                <div className="mb-1">
                  {isToday ? (
                    <span className="inline-flex items-center justify-center w-6 h-6 rounded-full bg-gold text-ink text-xs font-semibold">
                      {day.getDate()}
                    </span>
                  ) : (
                    <span className="text-sm text-cream-dim">{day.getDate()}</span>
                  )}
                </div>
                <div className="space-y-0.5">
                  {dayTasks.slice(0, 3).map(task => (
                    <button
                      key={task.id}
                      onClick={e => { e.stopPropagation(); onTaskClick(task) }}
                      className="w-full text-left truncate text-xs px-1 py-0.5 rounded bg-ink-raised hover:bg-ink-subtle transition-colors"
                      style={{ borderLeft: `2px solid ${getCategoryColor(task.category_id)}` }}
                    >
                      <span className="text-cream-dim">{task.title}</span>
                    </button>
                  ))}
                  {dayTasks.length > 3 && (
                    <div className="text-xs text-cream-faint px-1">+{dayTasks.length - 3} more</div>
                  )}
                </div>
              </button>
            </div>
          )
        })}
      </div>
    </div>
  )
}
