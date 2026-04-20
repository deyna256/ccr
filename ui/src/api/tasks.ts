import { authenticatedFetch } from './client'

export type TaskStatus = 'pending' | 'done' | 'archived'

export interface Task {
  id: string
  type: string
  title: string
  description?: string
  start_time: string
  end_time?: string
  duration_minutes?: number
  color?: string
  status: TaskStatus
  created_at: string
  updated_at: string
}

export interface TaskListParams {
  from?: string
  to?: string
  status?: TaskStatus
}

export async function listTasks(params?: TaskListParams): Promise<Task[]> {
  const searchParams = new URLSearchParams()
  if (params?.from) searchParams.set('from', params.from)
  if (params?.to) searchParams.set('to', params.to)
  if (params?.status) searchParams.set('status', params.status)

  const query = searchParams.toString()
  const response = await authenticatedFetch(`/tasks${query ? `?${query}` : ''}`)
  return response.json()
}

export async function getTask(id: string): Promise<Task> {
  const response = await authenticatedFetch(`/tasks/${id}`)
  return response.json()
}

export interface CreateTaskRequest {
  title: string
  description?: string
  start_time: string
  duration_minutes?: number
  color?: string
}

export async function createTask(data: CreateTaskRequest): Promise<Task> {
  const response = await authenticatedFetch('/tasks', {
    method: 'POST',
    body: JSON.stringify({ ...data, type: 'task' }),
  })
  return response.json()
}

export interface UpdateTaskRequest {
  title?: string
  description?: string
  start_time?: string
  duration_minutes?: number
  color?: string
}

export async function updateTask(id: string, data: UpdateTaskRequest): Promise<Task> {
  const response = await authenticatedFetch(`/tasks/${id}`, {
    method: 'PUT',
    body: JSON.stringify({ ...data, type: 'task' }),
  })
  return response.json()
}

export async function deleteTask(id: string): Promise<void> {
  await authenticatedFetch(`/tasks/${id}`, { method: 'DELETE' })
}

export async function updateTaskStatus(id: string, status: TaskStatus): Promise<Task> {
  const response = await authenticatedFetch(`/tasks/${id}/status`, {
    method: 'PATCH',
    body: JSON.stringify({ status }),
  })
  return response.json()
}
