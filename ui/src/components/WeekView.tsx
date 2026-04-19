import { useMemo, useRef, useState, useCallback } from 'react'
import { Task } from '../api/tasks'
import { Category } from '../api/categories'

interface WeekViewProps {
  currentDate: Date
  tasks: Task[]
  categories: Category[]
  onTaskClick: (task: Task) => void
  onEmptySlotClick: (date: Date) => void
  onTaskMove?: (taskId: string, newStart: Date) => void
}

const HOURS = Array.from({ length: 18 }, (_, i) => i + 6) // 06:00–23:00
const ROW_H = 3 // rem per hour

export default function WeekView({
  currentDate, tasks, categories,
  onTaskClick, onEmptySlotClick, onTaskMove,
}: WeekViewProps) {
  const days = useMemo(() => {
    const day = currentDate.getDay()
    const diff = currentDate.getDate() - day + (day === 0 ? -6 : 1)
    return Array.from({ length: 7 }, (_, i) => {
      const d = new Date(currentDate)
      d.setDate(diff + i)
      return d
    })
  }, [currentDate])

  const tasksByDay = useMemo(() => {
    const map = new Map<string, Task[]>()
    for (const task of tasks) {
      if (!task.start_time) continue
      const date = task.start_time.split('T')[0]
      const list = map.get(date) ?? []
      list.push(task)
      map.set(date, list)
    }
    return map
  }, [tasks])

  const getCategoryColor = useCallback((id?: string) => {
    if (!id) return '#57535e'
    return categories.find(c => c.id === id)?.color ?? '#57535e'
  }, [categories])

  // --- Drag state ---
  const bodyRef = useRef<HTMLDivElement>(null)
  const [dragTaskId, setDragTaskId] = useState<string | null>(null)
  const [dragOffsetMin, setDragOffsetMin] = useState(0)
  // snapMin = minutes from midnight where task top would land
  const [dropInfo, setDropInfo] = useState<{ dateStr: string; snapMin: number } | null>(null)

  function pxPerMin(): number {
    if (!bodyRef.current) return (ROW_H * 16) / 60
    return bodyRef.current.offsetHeight / (HOURS.length * 60)
  }

  function calcSnapMin(yInColumn: number): number {
    const raw = HOURS[0] * 60 + yInColumn / pxPerMin() - dragOffsetMin
    const clamped = Math.max(HOURS[0] * 60, Math.min(23 * 60, raw))
    return Math.round(clamped / 15) * 15
  }

  function handleDragStart(e: React.DragEvent, task: Task) {
    const rect = (e.currentTarget as HTMLElement).getBoundingClientRect()
    const offsetPx = e.clientY - rect.top
    setDragTaskId(task.id)
    setDragOffsetMin(Math.round(offsetPx / pxPerMin()))
    e.dataTransfer.effectAllowed = 'move'
  }

  function handleDragOver(e: React.DragEvent, dateStr: string) {
    e.preventDefault()
    e.dataTransfer.dropEffect = 'move'
    const rect = (e.currentTarget as HTMLElement).getBoundingClientRect()
    setDropInfo({ dateStr, snapMin: calcSnapMin(e.clientY - rect.top) })
  }

  function handleDragLeave(e: React.DragEvent) {
    if (!(e.currentTarget as HTMLElement).contains(e.relatedTarget as Node)) {
      setDropInfo(null)
    }
  }

  function handleDrop(e: React.DragEvent, day: Date) {
    e.preventDefault()
    if (!dragTaskId || !onTaskMove) { resetDrag(); return }

    const rect = (e.currentTarget as HTMLElement).getBoundingClientRect()
    const snapMin = calcSnapMin(e.clientY - rect.top)

    const newStart = new Date(day)
    newStart.setHours(Math.floor(snapMin / 60), snapMin % 60, 0, 0)
    onTaskMove(dragTaskId, newStart)
    resetDrag()
  }

  function resetDrag() {
    setDragTaskId(null)
    setDropInfo(null)
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
        <div ref={bodyRef} className="flex" style={{ height: `${totalH}rem` }}>
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
            const indicator = dropInfo?.dateStr === dateStr ? dropInfo.snapMin : null

            return (
              <div
                key={dateStr}
                className="flex-1 relative border-l border-ink-border"
                onDragOver={e => handleDragOver(e, dateStr)}
                onDragLeave={handleDragLeave}
                onDrop={e => handleDrop(e, day)}
              >
                {/* Hour grid lines */}
                {HOURS.map(hour => (
                  <div
                    key={hour}
                    className="absolute w-full border-t border-ink-subtle"
                    style={{ top: `${(hour - HOURS[0]) * ROW_H}rem` }}
                  />
                ))}

                {/* Drop indicator */}
                {indicator !== null && (
                  <div
                    className="absolute left-0 right-0 z-20 pointer-events-none"
                    style={{ top: `${(indicator - HOURS[0] * 60) / 60 * ROW_H}rem` }}
                  >
                    <div className="h-0.5 bg-gold w-full" />
                    <div className="absolute -top-1 -left-1 w-2 h-2 rounded-full bg-gold" />
                  </div>
                )}

                {/* Tasks */}
                {dayTasks.map(task => {
                  const start = new Date(task.start_time)
                  const top = ((start.getHours() - HOURS[0]) * 60 + start.getMinutes()) / 60 * ROW_H
                  const height = Math.max((task.duration_minutes ?? 60) / 60 * ROW_H, 0.75)
                  const color = getCategoryColor(task.category_id)
                  const isDragging = dragTaskId === task.id

                  return (
                    <div
                      key={task.id}
                      draggable
                      onDragStart={e => handleDragStart(e, task)}
                      onDragEnd={resetDrag}
                      onClick={() => onTaskClick(task)}
                      className={`absolute left-0.5 right-0.5 rounded px-1.5 py-1 cursor-grab active:cursor-grabbing overflow-hidden z-10 transition-opacity ${
                        isDragging ? 'opacity-30' : 'opacity-100 hover:opacity-80'
                      }`}
                      style={{
                        top: `${top}rem`,
                        height: `${height}rem`,
                        backgroundColor: color + '4D',
                        borderLeft: `2px solid ${color}`,
                      }}
                    >
                      <div className="text-xs text-cream truncate leading-tight select-none">{task.title}</div>
                      {height >= 1.5 && (
                        <div className="text-xs text-cream-faint leading-tight select-none">
                          {start.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', hour12: false })}
                          {task.duration_minutes ? ` · ${task.duration_minutes}m` : ''}
                        </div>
                      )}
                    </div>
                  )
                })}

                {/* Empty slot click (behind tasks) */}
                <button
                  onClick={() => onEmptySlotClick(day)}
                  className="absolute inset-0 w-full h-full"
                  tabIndex={-1}
                />
              </div>
            )
          })}
        </div>
      </div>
    </div>
  )
}
