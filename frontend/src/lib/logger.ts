type LogLevel = 'DEBUG' | 'INFO' | 'WARN' | 'ERROR'

interface LogEntry {
  level: LogLevel
  service: string
  msg: string
  [key: string]: unknown
}

function format(level: LogLevel, msg: string, data?: Record<string, unknown>): string {
  const entry: LogEntry = {
    level,
    service: 'ccr',
    msg,
    ...data,
  }
  return JSON.stringify(entry)
}

export function log(level: LogLevel, msg: string, data?: Record<string, unknown>) {
  console.log(format(level, msg, data))
}

export function debug(msg: string, data?: Record<string, unknown>) {
  log('DEBUG', msg, data)
}

export function info(msg: string, data?: Record<string, unknown>) {
  log('INFO', msg, data)
}

export function warn(msg: string, data?: Record<string, unknown>) {
  log('WARN', msg, data)
}

export function error(msg: string, data?: Record<string, unknown>) {
  log('ERROR', msg, data)
}