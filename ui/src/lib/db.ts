import { openDB, DBSchema, IDBPDatabase } from 'idb'

export interface Task {
  id: string
  type: string
  title: string
  description?: string
  start_time: string
  end_time?: string
  duration_minutes?: number
  color?: string
  status: 'pending' | 'done' | 'archived'
  created_at: string
  updated_at: string
  completed_at?: string
  archived_at?: string
  [key: string]: unknown
}

export interface ChangeLogEntry {
  id: string
  entity_type: 'task'
  entity_id: string
  action: 'create' | 'update' | 'delete'
  old_values: Task | null
  new_values: Task | null
  client_time: string
  synced: boolean
  conflict_status?: 'pending' | 'resolved' | 'rejected'
}

export interface SyncMeta {
  last_sync_at: string | null
  device_id: string
  pending_count: number
}

interface CCRDB extends DBSchema {
  tasks: {
    key: string
    value: Task
    indexes: {
      'by-status': string
      'by-updated': string
    }
  }
  change_log: {
    key: string
    value: ChangeLogEntry
    indexes: {
      'by-entity': string
      'by-synced': number
      'by-client-time': string
    }
  }
  sync_meta: {
    key: 'meta'
    value: SyncMeta
  }
}

let dbPromise: Promise<IDBPDatabase<CCRDB>> | null = null

export async function getDB(): Promise<IDBPDatabase<CCRDB>> {
  if (!dbPromise) {
    dbPromise = openDB<CCRDB>('ccr', 1, {
      upgrade(db) {
        const taskStore = db.createObjectStore('tasks', { keyPath: 'id' })
        taskStore.createIndex('by-status', 'status')
        taskStore.createIndex('by-updated', 'updated_at')

        const logStore = db.createObjectStore('change_log', { keyPath: 'id' })
        logStore.createIndex('by-entity', 'entity_id')
        logStore.createIndex('by-synced', 'synced')
        logStore.createIndex('by-client-time', 'client_time')

        db.createObjectStore('sync_meta', { keyPath: 'device_id' })
      },
    })
  }
  return dbPromise
}

export async function getSyncMeta(): Promise<SyncMeta> {
  const db = await getDB()
  let meta = await db.get('sync_meta', 'meta')
  if (!meta) {
    meta = {
      device_id: crypto.randomUUID(),
      last_sync_at: null,
      pending_count: 0,
    }
    await db.put('sync_meta', meta)
  }
  return meta
}

export async function updateSyncMeta(updates: Partial<SyncMeta>): Promise<void> {
  const db = await getDB()
  const meta = await getSyncMeta()
  await db.put('sync_meta', { ...meta, ...updates })
}

export async function getUnsyncedChanges(): Promise<ChangeLogEntry[]> {
  const db = await getDB()
  return db.getAllFromIndex('change_log', 'by-synced', 0)
}

export async function markChangesSynced(ids: string[]): Promise<void> {
  const db = await getDB()
  const tx = db.transaction('change_log', 'readwrite')
  await Promise.all(
    ids.map(async (id) => {
      const entry = await tx.store.get(id)
      if (entry) {
        entry.synced = true
        await tx.store.put(entry)
      }
    })
  )
  await tx.done
}

export async function addChange(entry: ChangeLogEntry): Promise<void> {
  const db = await getDB()
  await db.put('change_log', entry)
  const meta = await getSyncMeta()
  await updateSyncMeta({ pending_count: meta.pending_count + 1 })
}

export async function getTasks(): Promise<Task[]> {
  const db = await getDB()
  return db.getAll('tasks')
}

export async function getTask(id: string): Promise<Task | undefined> {
  const db = await getDB()
  return db.get('tasks', id)
}

export async function putTask(task: Task): Promise<void> {
  const db = await getDB()
  await db.put('tasks', task)
}

export async function deleteTaskFromDB(id: string): Promise<void> {
  const db = await getDB()
  await db.delete('tasks', id)
}

export async function clearAllData(): Promise<void> {
  const db = await getDB()
  const tx = db.transaction(['tasks', 'change_log'], 'readwrite')
  await tx.objectStore('tasks').clear()
  await tx.objectStore('change_log').clear()
  await tx.done
}
