import { useState, useEffect } from 'react'
import {
  Category,
  listCategories,
  createCategory,
  updateCategory,
  deleteCategory,
} from '../api/categories'

const COLORS = [
  '#ef4444', // red
  '#f97316', // orange
  '#eab308', // yellow
  '#22c55e', // green
  '#06b6d4', // cyan
  '#3b82f6', // blue
  '#8b5cf6', // violet
  '#ec4899', // pink
]

export default function SettingsPage() {
  const [categories, setCategories] = useState<Category[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const [showNew, setShowNew] = useState(false)
  const [newName, setNewName] = useState('')
  const [newColor, setNewColor] = useState(COLORS[0])

  const [editingId, setEditingId] = useState<string | null>(null)
  const [editName, setEditName] = useState('')
  const [editColor, setEditColor] = useState('')

  useEffect(() => {
    loadCategories()
  }, [])

  async function loadCategories() {
    try {
      const list = await listCategories()
      setCategories(list)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load')
    } finally {
      setLoading(false)
    }
  }

  async function handleCreate() {
    if (!newName.trim()) return
    try {
      await createCategory({ name: newName, color: newColor })
      setNewName('')
      setNewColor(COLORS[0])
      setShowNew(false)
      await loadCategories()
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to create')
    }
  }

  async function handleUpdate(id: string) {
    try {
      await updateCategory(id, { name: editName, color: editColor })
      setEditingId(null)
      await loadCategories()
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to update')
    }
  }

  async function handleDelete(id: string) {
    try {
      await deleteCategory(id)
      await loadCategories()
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to delete')
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <p className="text-zinc-500">Loading...</p>
      </div>
    )
  }

  return (
    <div className="p-6 max-w-2xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-xl font-semibold text-white">Categories</h1>
        <button onClick={() => setShowNew(true)} className="btn-primary">
          + New
        </button>
      </div>

      {error && (
        <p className="text-sm text-red-400 bg-red-900/20 p-2 rounded mb-4">
          {error}
        </p>
      )}

      {/* New category form */}
      {showNew && (
        <div className="card p-4 mb-4">
          <div className="flex items-center gap-4">
            <div className="flex gap-2">
              {COLORS.map(c => (
                <button
                  key={c}
                  onClick={() => setNewColor(c)}
                  className={`w-8 h-8 rounded-full ${
                    newColor === c
                      ? 'ring-2 ring-white ring-offset-2 ring-offset-zinc-900'
                      : ''
                  }`}
                  style={{ backgroundColor: c }}
                />
              ))}
            </div>
            <input
              type="text"
              value={newName}
              onChange={e => setNewName(e.target.value)}
              placeholder="Category name"
              className="input-field flex-1"
            />
            <button onClick={handleCreate} className="btn-primary">
              Add
            </button>
            <button onClick={() => setShowNew(false)} className="btn-ghost">
              Cancel
            </button>
          </div>
        </div>
      )}

      {/* Category list */}
      <div className="space-y-2">
        {categories.map(cat => (
          <div key={cat.id} className="card p-3 flex items-center gap-4">
            {editingId === cat.id ? (
              <>
                <div className="flex gap-2">
                  {COLORS.map(c => (
                    <button
                      key={c}
                      onClick={() => setEditColor(c)}
                      className={`w-6 h-6 rounded-full ${
                        editColor === c
                          ? 'ring-2 ring-white ring-offset-2 ring-offset-zinc-900'
                          : ''
                      }`}
                      style={{ backgroundColor: c }}
                    />
                  ))}
                </div>
                <input
                  type="text"
                  value={editName}
                  onChange={e => setEditName(e.target.value)}
                  className="input-field flex-1"
                />
                <button
                  onClick={() => handleUpdate(cat.id)}
                  className="btn-primary"
                >
                  Save
                </button>
                <button
                  onClick={() => setEditingId(null)}
                  className="btn-ghost"
                >
                  Cancel
                </button>
              </>
            ) : (
              <>
                <div
                  className="w-4 h-4 rounded-full"
                  style={{ backgroundColor: cat.color }}
                />
                <span className="text-white flex-1">{cat.name}</span>
                <button
                  onClick={() => {
                    setEditingId(cat.id)
                    setEditName(cat.name)
                    setEditColor(cat.color)
                  }}
                  className="text-sm text-zinc-500 hover:text-white"
                >
                  Edit
                </button>
                <button
                  onClick={() => handleDelete(cat.id)}
                  className="btn-danger"
                >
                  Delete
                </button>
              </>
            )}
          </div>
        ))}
      </div>

      {categories.length === 0 && !showNew && (
        <p className="text-center text-zinc-500 py-8">
          No categories yet. Create one to get started.
        </p>
      )}
    </div>
  )
}