import { authenticatedFetch } from '../api/client'
import {
  getSyncMeta,
  updateSyncMeta,
  getUnsyncedChanges,
  markChangesSynced,
  addChange,
  putTask,
  deleteTaskFromDB,
  getTask,
  type ChangeLogEntry,
  type Task,
} from './db'

export type SyncStatus = 'idle' | 'syncing' | 'error'

export interface SyncResult {
  pushed: number
  pulled: number
  errors: string[]
}

class SyncManager {
  private status: SyncStatus = 'idle'
  private lastError: string | null = null
  private listeners: Set<(status: SyncStatus, error: string | null) => void> = new Set()

  getStatus(): SyncStatus {
    return this.status
  }

  getLastError(): string | null {
    return this.lastError
  }

  subscribe(listener: (status: SyncStatus, error: string | null) => void): () => void {
    this.listeners.add(listener)
    return () => this.listeners.delete(listener)
  }

  private notify() {
    this.listeners.forEach((l) => l(this.status, this.lastError))
  }

  async sync(): Promise<SyncResult> {
    if (this.status === 'syncing') {
      return { pushed: 0, pulled: 0, errors: ['sync already in progress'] }
    }

    this.status = 'syncing'
    this.lastError = null
    this.notify()

    const result: SyncResult = { pushed: 0, pulled: 0, errors: [] }

    try {
      const meta = await getSyncMeta()

      const unsynced = await getUnsyncedChanges()
      if (unsynced.length > 0) {
        const pushResult = await this.pushChanges(unsynced, meta.device_id)
        result.pushed = pushResult.accepted.length
        if (pushResult.rejected.length > 0) {
          result.errors.push(`${pushResult.rejected.length} changes rejected`)
        }
        await markChangesSynced(pushResult.accepted)
        await updateSyncMeta({ pending_count: 0 })
      }

      const pullResult = await this.pullChanges(meta.last_sync_at, meta.device_id)
      result.pulled = pullResult.changes.length

      await updateSyncMeta({ last_sync_at: new Date().toISOString() })

      this.status = 'idle'
      this.notify()
      return result
    } catch (err) {
      this.status = 'error'
      this.lastError = err instanceof Error ? err.message : 'sync failed'
      this.notify()
      result.errors.push(this.lastError)
      return result
    }
  }

  private async pushChanges(
    changes: ChangeLogEntry[],
    deviceId: string
  ): Promise<{ accepted: string[]; rejected: { id: string; reason: string }[] }> {
    const response = await authenticatedFetch('/sync', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        device_id: deviceId,
        last_sync: (await getSyncMeta()).last_sync_at ?? '',
        changes: changes.map((c) => ({
          id: c.id,
          entity_type: c.entity_type,
          entity_id: c.entity_id,
          action: c.action,
          old_values: c.old_values,
          new_values: c.new_values,
          client_time: c.client_time,
        })),
      }),
    })

    const data = await response.json()
    return {
      accepted: data.accepted ?? [],
      rejected: data.rejected ?? [],
    }
  }

  private async pullChanges(
    lastSync: string | null,
    deviceId: string
  ): Promise<{ changes: ChangeLogEntry[] }> {
    const response = await authenticatedFetch('/sync', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        device_id: deviceId,
        last_sync: lastSync ?? '',
        changes: [],
      }),
    })

    const data = await response.json()

    for (const change of data.server_changes ?? []) {
      await this.applyServerChange(change)
    }

    return { changes: data.server_changes ?? [] }
  }

  private async applyServerChange(change: ChangeLogEntry): Promise<void> {
    if (change.entity_type !== 'task') return

    switch (change.action) {
      case 'create':
      case 'update': {
        if (change.new_values) {
          const task: Task = {
            id: change.entity_id,
            type: (change.new_values.type as string) ?? 'task',
            title: change.new_values.title as string,
            description: change.new_values.description as string | undefined,
            start_time: (change.new_values.start_time as string) ?? '',
            end_time: change.new_values.end_time as string | undefined,
            color: change.new_values.color as string | undefined,
            status: (change.new_values.status as Task['status']) ?? 'pending',
            created_at: (change.new_values.created_at as string) ?? new Date().toISOString(),
            updated_at: (change.new_values.updated_at as string) ?? new Date().toISOString(),
          }
          await putTask(task)
        }
        break
      }
      case 'delete': {
        await deleteTaskFromDB(change.entity_id)
        break
      }
    }
  }

  async localCreate(task: Task): Promise<void> {
    await putTask(task)
    await addChange({
      id: crypto.randomUUID(),
      entity_type: 'task',
      entity_id: task.id,
      action: 'create',
      old_values: null,
      new_values: task,
      client_time: new Date().toISOString(),
      synced: false,
    })
  }

  async localUpdate(task: Task, oldTask: Task): Promise<void> {
    await putTask(task)
    await addChange({
      id: crypto.randomUUID(),
      entity_type: 'task',
      entity_id: task.id,
      action: 'update',
      old_values: oldTask,
      new_values: task,
      client_time: new Date().toISOString(),
      synced: false,
    })
  }

  async localDelete(taskId: string): Promise<void> {
    const oldTask = await getTask(taskId)
    await deleteTaskFromDB(taskId)
    if (oldTask) {
      await addChange({
        id: crypto.randomUUID(),
        entity_type: 'task',
        entity_id: taskId,
        action: 'delete',
        old_values: oldTask,
        new_values: null,
        client_time: new Date().toISOString(),
        synced: false,
      })
    }
  }
}

export const syncManager = new SyncManager()
