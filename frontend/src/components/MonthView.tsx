import { useMemo } from 'react'
import { Task } from '../api/tasks'
import { Category } from '../api/categories'

interface MonthViewProps {
  tasks: Task[]
  categories: Category[]
  onTaskClick: (task: Task) => void
  onEmptySlotClick: (date: Date) => void
}

export default function MonthView({
  tasks,
  categories,
  onTaskClick,
  onEmptySlotClick,
}: MonthViewProps) {
  const today = new Date()
  const year = today.getFullYear()
  const month = today.getMonth()

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
    if (!id) return '#666'
    return categories.find(c => c.id === id)?.color ?? '#666'
  }

  return (
    <div className="flex-1 overflow-auto">
      <div className="grid grid-cols-7 gap-px bg-zinc-800">
        {['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'].map(day => (
          <div key={day} className="bg-zinc-900 p-2 text-center text-xs text-zinc-500">
            {day}
          </div>
        ))}
        {days.map(day => {
          const dateStr = day.toISOString().split('T')[0]
          const dayTasks = tasksByDay.get(dateStr) ?? []
          const isCurrentMonth = day.getMonth() === month
          const isToday =
            isCurrentMonth &&
            day.toDateString() === today.toDateString()

          return (
            <div
              key={dateStr}
              className={`bg-zinc-900 min-h-[100px] p-1 ${
                isCurrentMonth ? '' : 'opacity-30'
              }`}
            >
              <button
                onClick={() => onEmptySlotClick(day)}
                className="w-full text-left"
              >
                <div
                  className={`text-sm mb-1 ${
                    isToday
                      ? 'w-6 h-6 rounded-full bg-blue-600 text-white flex items-center justify-center'
                      : 'text-zinc-400'
                  }`}
                >
                  {day.getDate()}
                </div>
                <div className="space-y-0.5">
                  {dayTasks.slice(0, 3).map(task => (
                    <button
                      key={task.id}
                      onClick={e => {
                        e.stopPropagation()
                        onTaskClick(task)
                      }}
                      className="w-full text-left truncate text-xs px-1 py-0.5 rounded bg-zinc-800 hover:bg-zinc-700"
                      style={{
                        borderLeft: `2px solid ${getCategoryColor(task.category_id)}`,
                      }}
                    >
                      {task.title}
                    </button>
                  ))}
                  {dayTasks.length > 3 && (
                    <div className="text-xs text-zinc-500 px-1">
                      +{dayTasks.length - 3} more
                    </div>
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