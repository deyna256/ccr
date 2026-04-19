import { authenticatedFetch } from './client'

export interface Category {
  id: string
  name: string
  color: string
}

export async function listCategories(): Promise<Category[]> {
  const response = await authenticatedFetch('/categories')
  return response.json()
}

export async function createCategory(data: { name: string; color: string }): Promise<Category> {
  const response = await authenticatedFetch('/categories', {
    method: 'POST',
    body: JSON.stringify(data),
  })
  return response.json()
}

export async function updateCategory(
  id: string,
  data: { name?: string; color?: string },
): Promise<Category> {
  const response = await authenticatedFetch(`/categories/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  })
  return response.json()
}

export async function deleteCategory(id: string): Promise<void> {
  await authenticatedFetch(`/categories/${id}`, {
    method: 'DELETE',
  })
}