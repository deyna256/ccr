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
const ROW_H = 3 // rem per hour

export default function WeekView({ tasks, categories, onTaskClick, onEmptySlotClick }: WeekViewProps) {
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
    if (!id) return '#57535e'
    return categories.find(c => c.id === id)?.color ?? '#57535e'
  }

  const totalH = HOURS.length * ROW_H

  return (
    <div className="flex-1 overflow-auto">
      <div className="min-w-[600px]">
        {/* Sticky header */}
        <div className="sticky top-0 z-10 flex bg-ink border-b border-ink-border">
          <div className="w-12 shrink-0" />
          {days.map(day => {
            const isToday = day.toDateString() === new Date().toDateString()
            return (
              <div key={day.toISOString()} className="flex-1 p-2 text-center">
                <div className="text-xs text-cream-faint uppercase tracking-wider">
                  {day.toLocaleDateString('en-US', { weekday: 'short' })}
                </div>
                <div className={`text-lg font-medium mt-0.5 ${isToday ? 'text-gold' : 'text-cream'}`}>
                  {day.getDate()}
                </div>
              </div>
            )
          })}
        </div>

        {/* Body */}
        <div className="flex" style={{ height: `${totalH}rem` }}>
          {/* Time gutter */}
          <div className="w-12 shrink-0 relative">
            {HOURS.map(hour => (
              <div
                key={hour}
                className="absolute w-full text-xs text-cream-faint font-mono pl-1 select-none"
                style={{ top: `${(hour - HOURS[0]) * ROW_H}rem` }}
              >
                {String(hour).padStart(2, '0')}:00
              </div>
            ))}
          </div>

          {/* Day columns */}
          {days.map(day => {
            const dateStr = day.toISOString().split('T')[0]
            const dayTasks = tasksByDay.get(dateStr) ?? []

            return (
              <div key={dateStr} className="flex-1 relative border-l border-ink-border">
                {/* Hour grid lines */}
                {HOURS.map(hour => (
                  <div
                    key={hour}
                    className="absolute w-full border-t border-ink-subtle"
                    style={{ top: `${(hour - HOURS[0]) * ROW_H}rem` }}
                  />
                ))}

                {/* Tasks */}
                {dayTasks.map(task => {
                  const start = new Date(task.start_time)
                  const top = ((start.getHours() - HOURS[0]) * 60 + start.getMinutes()) / 60 * ROW_H
                  const height = Math.max((task.duration_minutes ?? 60) / 60 * ROW_H, 0.75)
                  const color = getCategoryColor(task.category_id)
                  return (
                    <button
                      key={task.id}
                      onClick={e => { e.stopPropagation(); onTaskClick(task) }}
                      className="absolute left-0.5 right-0.5 rounded p-1 hover:opacity-80 transition-opacity overflow-hidden z-10"
                      style={{
                        top: `${top}rem`,
                        height: `${height}rem`,
                        backgroundColor: color + '28',
                        borderLeft: `2px solid ${color}`,
                      }}
                    >
                      <div className="text-xs text-cream truncate leading-tight">{task.title}</div>
                    </button>
                  )
                })}

                {/* Empty slot click */}
                <button
                  onClick={() => onEmptySlotClick(day)}
                  className="absolute inset-0 w-full h-full"
                />
              </div>
            )
          })}
        </div>
      </div>
    </div>
  )
}
