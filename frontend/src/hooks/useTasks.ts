import { openDB, DBSchema, IDBPDatabase } from 'idb'
import { Task, TaskListParams, listTasks } from '../api/tasks'

interface CCRDB extends DBSchema {
  tasks: {
    key: string
    value: Task
    indexes: { 'by-date': string }
  }
  pending: {
    key: string
    value: { id: string; action: 'create' | 'update' | 'delete'; data?: unknown; timestamp: number }
  }
}

let dbPromise: Promise<IDBPDatabase<CCRDB>> | null = null

function getDB() {
  if (!dbPromise) {
    dbPromise = openDB<CCRDB>('ccr', 1, {
      upgrade(db) {
        const taskStore = db.createObjectStore('tasks', { keyPath: 'id' })
        taskStore.createIndex('by-date', 'start_time')
        db.createObjectStore('pending', { keyPath: 'id' })
      },
    })
  }
  return dbPromise
}

export async function loadTasksFromCache(params?: TaskListParams): Promise<Task[]> {
  const db = await getDB()
  let tasks = await db.getAll('tasks')

  if (params?.from) {
    tasks = tasks.filter(t => t.start_time >= params.from!)
  }
  if (params?.to) {
    tasks = tasks.filter(t => t.start_time <= params.to!)
  }
  if (params?.status) {
    tasks = tasks.filter(t => t.status === params.status)
  }
  if (params?.category_id) {
    tasks = tasks.filter(t => t.category_id === params.category_id)
  }

  return tasks
}

export async function saveTasksToCache(tasks: Task[]) {
  const db = await getDB()
  const tx = db.transaction('tasks', 'readwrite')
  await Promise.all(tasks.map(t => tx.store.put(t)))
}

export async function clearTasksCache() {
  const db = await getDB()
  await db.clear('tasks')
}

export async function getPendingActions() {
  const db = await getDB()
  return db.getAll('pending')
}

export async function addPendingAction(
  action: 'create' | 'update' | 'delete',
  data?: unknown,
) {
  const db = await getDB()
  await db.put('pending', {
    id: crypto.randomUUID(),
    action,
    data,
    timestamp: Date.now(),
  })
}

export async function clearPendingActions() {
  const db = await getDB()
  await db.clear('pending')
}

export async function syncTasks() {
  try {
    const tasks = await listTasks()
    await saveTasksToCache(tasks)
  } catch {
    // Offline - load from cache
    return loadTasksFromCache()
  }
}