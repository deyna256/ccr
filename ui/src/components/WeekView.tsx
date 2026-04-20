import { useMemo, useRef, useState, useEffect } from 'react'
import { Task } from '../api/tasks'

interface WeekViewProps {
  currentDate: Date
  tasks: Task[]
  onTaskClick: (task: Task) => void
  onEmptySlotClick: (date: Date) => void
  onTaskMove?: (taskId: string, newStart: Date) => void
  onTaskResize?: (taskId: string, newStart: Date, durationMinutes: number) => void
}

const HOURS = Array.from({ length: 18 }, (_, i) => i + 6) // 06:00–23:00
const ROW_H = 3 // rem per hour
const GUTTER_PX = 48 // w-12
const MOVE_THRESHOLD = 6 // px before move activates

type Interaction =
  | {
      kind: 'move'
      task: Task
      offsetMin: number
      startX: number
      startY: number
      moved: boolean
      previewDayIdx: number
      previewMin: number
    }
  | {
      kind: 'resize'
      task: Task
      dir: 'top' | 'bottom'
      origStart: Date
      origDuration: number
      moved: boolean
      previewStart: Date
      previewDuration: number
    }

export default function WeekView({
  currentDate, tasks,
  onTaskClick, onEmptySlotClick, onTaskMove, onTaskResize,
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

  const bodyRef = useRef<HTMLDivElement>(null)
  const [interaction, setInteraction] = useState<Interaction | null>(null)
  const interRef = useRef<Interaction | null>(null)

  function pxPerMin(): number {
    if (!bodyRef.current) return (ROW_H * 16) / 60
    return bodyRef.current.offsetHeight / (HOURS.length * 60)
  }

  function snap15(min: number) {
    return Math.round(min / 15) * 15
  }

  function getDayIdxFromX(mouseX: number): number {
    if (!bodyRef.current) return 0
    const rect = bodyRef.current.getBoundingClientRect()
    const available = rect.width - GUTTER_PX
    const colWidth = available / 7
    const x = mouseX - rect.left - GUTTER_PX
    return Math.max(0, Math.min(6, Math.floor(x / colWidth)))
  }

  function getMinFromY(mouseY: number): number {
    if (!bodyRef.current) return HOURS[0] * 60
    const rect = bodyRef.current.getBoundingClientRect()
    const y = mouseY - rect.top
    const raw = HOURS[0] * 60 + y / pxPerMin()
    return Math.max(HOURS[0] * 60, Math.min(23 * 60, raw))
  }

  // Mouse-down on card body → start move interaction
  function handleMoveStart(e: React.MouseEvent, task: Task) {
    if (e.button !== 0) return
    e.preventDefault()
    e.stopPropagation()
    const start = new Date(task.start_time)
    const taskTopMin = start.getHours() * 60 + start.getMinutes()
    const offsetMin = (e.clientY - (bodyRef.current?.getBoundingClientRect().top ?? 0)) / pxPerMin() - (taskTopMin - HOURS[0] * 60)
    const dayIdx = getDayIdxFromX(e.clientX)
    const state: Interaction = {
      kind: 'move',
      task,
      offsetMin: Math.max(0, offsetMin),
      startX: e.clientX,
      startY: e.clientY,
      moved: false,
      previewDayIdx: dayIdx,
      previewMin: taskTopMin,
    }
    interRef.current = state
    setInteraction(state)
  }

  // Mouse-down on resize handle
  function handleResizeStart(e: React.MouseEvent, task: Task, dir: 'top' | 'bottom') {
    if (e.button !== 0) return
    e.preventDefault()
    e.stopPropagation()
    const start = new Date(task.start_time)
    const duration = task.duration_minutes ?? 60
    const state: Interaction = {
      kind: 'resize',
      task,
      dir,
      origStart: start,
      origDuration: duration,
      moved: false,
      previewStart: start,
      previewDuration: duration,
    }
    interRef.current = state
    setInteraction(state)
  }

  useEffect(() => {
    function onMouseMove(e: MouseEvent) {
      const cur = interRef.current
      if (!cur) return

      if (cur.kind === 'move') {
        const dx = Math.abs(e.clientX - cur.startX)
        const dy = Math.abs(e.clientY - cur.startY)
        const moved = cur.moved || dx > MOVE_THRESHOLD || dy > MOVE_THRESHOLD

        const dayIdx = getDayIdxFromX(e.clientX)
        const rawMin = getMinFromY(e.clientY) - cur.offsetMin
        const previewMin = snap15(Math.max(HOURS[0] * 60, Math.min(23 * 60 - 15, rawMin)))

        const next: Interaction = { ...cur, moved, previewDayIdx: dayIdx, previewMin }
        interRef.current = next
        setInteraction(next)
      } else {
        // resize
        const absoluteMin = getMinFromY(e.clientY)

        let previewStart = cur.origStart
        let previewDuration = cur.origDuration

        if (cur.dir === 'bottom') {
          const origStartMin = cur.origStart.getHours() * 60 + cur.origStart.getMinutes()
          const endMin = snap15(Math.max(origStartMin + 15, absoluteMin))
          previewDuration = endMin - origStartMin
        } else {
          const origEndMin = cur.origStart.getHours() * 60 + cur.origStart.getMinutes() + cur.origDuration
          const newStartMin = snap15(Math.min(origEndMin - 15, absoluteMin))
          const clamped = Math.max(HOURS[0] * 60, newStartMin)
          previewDuration = origEndMin - clamped
          previewStart = new Date(cur.origStart)
          previewStart.setHours(Math.floor(clamped / 60), clamped % 60, 0, 0)
        }

        const moved = cur.moved || previewDuration !== cur.origDuration || previewStart.getTime() !== cur.origStart.getTime()
        const next: Interaction = { ...cur, moved, previewStart, previewDuration }
        interRef.current = next
        setInteraction(next)
      }
    }

    function onMouseUp() {
      const cur = interRef.current
      if (!cur) return

      if (cur.kind === 'move') {
        if (cur.moved && onTaskMove) {
          const newStart = new Date(days[cur.previewDayIdx])
          newStart.setHours(Math.floor(cur.previewMin / 60), cur.previewMin % 60, 0, 0)
          onTaskMove(cur.task.id, newStart)
        } else if (!cur.moved) {
          onTaskClick(cur.task)
        }
      } else {
        if (cur.moved && onTaskResize) {
          onTaskResize(cur.task.id, cur.previewStart, cur.previewDuration)
        }
      }

      interRef.current = null
      setInteraction(null)
    }

    document.addEventListener('mousemove', onMouseMove)
    document.addEventListener('mouseup', onMouseUp)
    return () => {
      document.removeEventListener('mousemove', onMouseMove)
      document.removeEventListener('mouseup', onMouseUp)
    }
  }, [days, onTaskMove, onTaskResize, onTaskClick])

  const totalH = HOURS.length * ROW_H

  // For move preview: which day column shows indicator
  const movePreview = interaction?.kind === 'move' && interaction.moved ? interaction : null
  const moveDateStr = movePreview ? days[movePreview.previewDayIdx]?.toISOString().split('T')[0] : null

  const isAnyInteraction = interaction !== null

  return (
    <div className="flex-1 overflow-auto" style={{ userSelect: isAnyInteraction ? 'none' : undefined }}>
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
          {days.map((day) => {
            const dateStr = day.toISOString().split('T')[0]
            const dayTasks = tasksByDay.get(dateStr) ?? []
            const showMoveIndicator = moveDateStr === dateStr

            return (
              <div
                key={dateStr}
                className="flex-1 relative border-l border-ink-border"
              >
                {/* Hour grid lines */}
                {HOURS.map(hour => (
                  <div
                    key={hour}
                    className="absolute w-full border-t border-ink-subtle"
                    style={{ top: `${(hour - HOURS[0]) * ROW_H}rem` }}
                  />
                ))}

                {/* Move preview indicator */}
                {showMoveIndicator && movePreview && (
                  <div
                    className="absolute left-0 right-0 z-20 pointer-events-none"
                    style={{ top: `${(movePreview.previewMin - HOURS[0] * 60) / 60 * ROW_H}rem` }}
                  >
                    <div className="h-0.5 bg-gold w-full" />
                    <div className="absolute -top-1 -left-1 w-2 h-2 rounded-full bg-gold" />
                  </div>
                )}

                {/* Tasks */}
                {dayTasks.map(task => {
                  const isMoved = interaction?.kind === 'move' && interaction.task.id === task.id && interaction.moved
                  const isResizing = interaction?.kind === 'resize' && interaction.task.id === task.id

                  let displayStart = new Date(task.start_time)
                  let displayDuration = task.duration_minutes ?? 60

                  if (isResizing && interaction.kind === 'resize') {
                    displayStart = interaction.previewStart
                    displayDuration = interaction.previewDuration
                  }

                  const top = ((displayStart.getHours() - HOURS[0]) * 60 + displayStart.getMinutes()) / 60 * ROW_H
                  const height = Math.max(displayDuration / 60 * ROW_H, 0.75)
                  const color = task.color ?? '#57535e'

                  return (
                    <div
                      key={task.id}
                      className={`absolute left-0.5 right-0.5 rounded overflow-hidden z-10 select-none ${
                        isMoved ? 'opacity-30' : 'opacity-100'
                      }`}
                      style={{
                        top: `${top}rem`,
                        height: `${height}rem`,
                        backgroundColor: color + '4D',
                        borderLeft: `2px solid ${color}`,
                        cursor: isAnyInteraction ? (interaction?.kind === 'resize' ? 'ns-resize' : 'grabbing') : 'grab',
                      }}
                    >
                      {/* Top resize handle */}
                      <div
                        className="absolute top-0 left-0 right-0 h-3 z-20"
                        style={{ cursor: 'ns-resize' }}
                        onMouseDown={e => handleResizeStart(e, task, 'top')}
                      />

                      {/* Card body — drag to move */}
                      <div
                        className="absolute inset-0 top-3 bottom-3 px-1.5 py-1"
                        onMouseDown={e => handleMoveStart(e, task)}
                      >
                        <div className="text-xs text-cream truncate leading-tight">{task.title}</div>
                        {height >= 1.5 && (
                          <div className="text-xs text-cream-faint leading-tight">
                            {displayStart.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', hour12: false })}
                            {displayDuration ? ` · ${displayDuration}m` : ''}
                          </div>
                        )}
                      </div>

                      {/* Bottom resize handle */}
                      <div
                        className="absolute bottom-0 left-0 right-0 h-3 z-20"
                        style={{ cursor: 'ns-resize' }}
                        onMouseDown={e => handleResizeStart(e, task, 'bottom')}
                      />
                    </div>
                  )
                })}

                {/* Empty slot click */}
                <button
                  onClick={() => !isAnyInteraction && onEmptySlotClick(day)}
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
