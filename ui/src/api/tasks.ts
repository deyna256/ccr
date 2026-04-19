import { authenticatedFetch } from './client'

export type TaskType = 'task' | 'event'
export type TaskStatus = 'pending' | 'done' | 'archived'

export interface Task {
  id: string
  type: TaskType
  title: string
  description?: string
  start_time: string
  end_time?: string
  duration_minutes?: number
  category_id?: string
  color?: string
  status: TaskStatus
  recurrence_id?: string
  recurrence?: RecurrenceRule
  created_at: string
  updated_at: string
}

export interface RecurrenceRule {
  frequency: 'daily' | 'weekly' | 'monthly'
  interval: number
  end_date?: string
  count?: number
}

export interface TaskListParams {
  from?: string
  to?: string
  type?: TaskType
  status?: TaskStatus
  category_id?: string
}

export async function listTasks(params?: TaskListParams): Promise<Task[]> {
  const searchParams = new URLSearchParams()
  if (params?.from) searchParams.set('from', params.from)
  if (params?.to) searchParams.set('to', params.to)
  if (params?.type) searchParams.set('type', params.type)
  if (params?.status) searchParams.set('status', params.status)
  if (params?.category_id) searchParams.set('category_id', params.category_id)

  const query = searchParams.toString()
  const response = await authenticatedFetch(`/tasks${query ? `?${query}` : ''}`)
  return response.json()
}

export async function getTask(id: string): Promise<Task> {
  const response = await authenticatedFetch(`/tasks/${id}`)
  return response.json()
}

export interface CreateTaskRequest {
  type: TaskType
  title: string
  description?: string
  start_time: string
  end_time?: string
  duration_minutes?: number
  category_id?: string
  color?: string
  recurrence?: RecurrenceRule
}

export async function createTask(data: CreateTaskRequest): Promise<Task> {
  const response = await authenticatedFetch('/tasks', {
    method: 'POST',
    body: JSON.stringify(data),
  })
  return response.json()
}

export interface UpdateTaskRequest {
  type?: TaskType
  title?: string
  description?: string
  start_time?: string
  end_time?: string
  duration_minutes?: number
  category_id?: string
  color?: string
  status?: TaskStatus
}

export async function updateTask(id: string, data: UpdateTaskRequest): Promise<Task> {
  const response = await authenticatedFetch(`/tasks/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  })
  return response.json()
}

export async function deleteTask(id: string): Promise<void> {
  await authenticatedFetch(`/tasks/${id}`, {
    method: 'DELETE',
  })
}

export async function updateTaskStatus(
  id: string,
  status: TaskStatus,
): Promise<Task> {
  const response = await authenticatedFetch(`/tasks/${id}/status`, {
    method: 'PATCH',
    body: JSON.stringify({ status }),
  })
  return response.json()
}