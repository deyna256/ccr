import { useMemo } from 'react'
import { Task } from '../api/tasks'
import { Category } from '../api/categories'

interface WeekViewProps {
  tasks: Task[]
  categories: Category[]
  onTaskClick: (task: Task) => void
  onEmptySlotClick: (date: Date) => void
}

const HOURS = Array.from({ length: 14 }, (_, i) => i + 7)

export default function WeekView({
  tasks,
  categories,
  onTaskClick,
  onEmptySlotClick,
}: WeekViewProps) {
  const days = useMemo(() => {
    const today = new Date()
    const day = today.getDay()
    const diff = today.getDate() - day + (day === 0 ? -6 : 1)
    return Array.from({ length: 7 }, (_, i) => {
      const d = new Date(today)
      d.setDate(diff + i)
      return d
    })
  }, [])

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
    if (!id) return '#666'
    return categories.find(c => c.id === id)?.color ?? '#666'
  }

  return (
    <div className="flex-1 overflow-auto">
      <div className="grid grid-cols-8 min-w-[600px]">
        {/* Header row */}
        <div className="sticky top-0 bg-zinc-950 z-10" />
        {days.map(day => (
          <div
            key={day.toISOString()}
            className="sticky top-0 bg-zinc-950 border-b border-zinc-800 p-2 text-center z-10"
          >
            <div className="text-xs text-zinc-500 uppercase">
              {day.toLocaleDateString('en-US', { weekday: 'short' })}
            </div>
            <div className="text-lg font-medium text-white">{day.getDate()}</div>
          </div>
        ))}

        {/* Time column */}
        {HOURS.map(hour => (
          <div key={hour} className="relative">
            <div className="absolute top-0 left-0 w-full text-xs text-zinc-600 font-mono pl-1">
              {hour.toString().padStart(2, '0')}:00
            </div>
          </div>
        ))}

        {/* Day columns */}
        {days.map(day => {
          const dateStr = day.toISOString().split('T')[0]
          const dayTasks = tasksByDay.get(dateStr) ?? []

          return (
            <div key={dateStr} className="relative min-h-[500px] border-l border-zinc-800">
              {dayTasks.map(task => {
                const start = new Date(task.start_time)
                const hour = start.getHours()
                const minute = start.getMinutes()
                const top = ((hour - 7) * 60 + minute) / 60 * 3 + 'rem'
                const height = ((task.duration_minutes ?? 60) / 60) * 3 + 'rem'

                return (
                  <button
                    key={task.id}
                    onClick={() => onTaskClick(task)}
                    className="absolute left-1 right-1 rounded p-1 hover:opacity-80 transition-opacity overflow-hidden"
                    style={{
                      top,
                      height,
                      backgroundColor: getCategoryColor(task.category_id) + '40',
                      borderLeft: `3px solid ${getCategoryColor(task.category_id)}`,
                    }}
                  >
                    <div className="text-xs text-white truncate">{task.title}</div>
                  </button>
                )
              })}

              {/* Empty slot click handler */}
              <button
                onClick={() => onEmptySlotClick(day)}
                className="absolute inset-0 w-full"
                style={{ height: '500px' }}
              />
            </div>
          )
        })}
      </div>
    </div>
  )
}