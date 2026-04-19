import { useState, useEffect } from 'react'
import {
  Category,
  listCategories,
  createCategory,
  updateCategory,
  deleteCategory,
} from '../api/categories'

const COLORS = [
  '#ef4444',
  '#f97316',
  '#eab308',
  '#22c55e',
  '#06b6d4',
  '#3b82f6',
  '#8b5cf6',
  '#ec4899',
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
        <p className="text-cream-faint text-sm">Loading…</p>
      </div>
    )
  }

  return (
    <div className="p-6 max-w-2xl mx-auto page-enter">
      <div className="flex items-center justify-between mb-5">
        <div>
          <h1 className="text-xl font-semibold text-cream">Categories</h1>
          <p className="text-xs text-cream-faint mt-0.5">{categories.length} total</p>
        </div>
        <button onClick={() => setShowNew(true)} className="btn-primary py-1.5">
          + New
        </button>
      </div>

      {error && (
        <div className="mb-4 text-xs text-ember border border-ember/20 bg-ember/5 px-4 py-3 rounded">
          {error}
        </div>
      )}

      {/* New category form */}
      {showNew && (
        <div className="card p-4 mb-4 animate-slide-up">
          <div className="flex items-center gap-3 flex-wrap">
            <div className="flex gap-1.5">
              {COLORS.map(c => (
                <button
                  key={c}
                  onClick={() => setNewColor(c)}
                  className={`w-6 h-6 rounded-full transition-transform ${
                    newColor === c ? 'ring-2 ring-gold ring-offset-2 ring-offset-ink-surface scale-110' : ''
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
              className="input-field flex-1 min-w-[140px]"
            />
            <button onClick={handleCreate} className="btn-primary py-1.5">Add</button>
            <button onClick={() => setShowNew(false)} className="btn-ghost py-1.5">Cancel</button>
          </div>
        </div>
      )}

      {/* Category list */}
      <div className="space-y-2">
        {categories.map(cat => (
          <div key={cat.id} className="card p-3 flex items-center gap-3">
            {editingId === cat.id ? (
              <>
                <div className="flex gap-1.5">
                  {COLORS.map(c => (
                    <button
                      key={c}
                      onClick={() => setEditColor(c)}
                      className={`w-5 h-5 rounded-full transition-transform ${
                        editColor === c ? 'ring-2 ring-gold ring-offset-2 ring-offset-ink-surface scale-110' : ''
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
                <button onClick={() => handleUpdate(cat.id)} className="btn-primary py-1.5">Save</button>
                <button onClick={() => setEditingId(null)} className="btn-ghost py-1.5">Cancel</button>
              </>
            ) : (
              <>
                <div className="w-3.5 h-3.5 rounded-full shrink-0" style={{ backgroundColor: cat.color }} />
                <span className="text-cream text-sm flex-1">{cat.name}</span>
                <button
                  onClick={() => {
                    setEditingId(cat.id)
                    setEditName(cat.name)
                    setEditColor(cat.color)
                  }}
                  className="text-xs text-cream-faint hover:text-cream transition-colors"
                >
                  Edit
                </button>
                <button onClick={() => handleDelete(cat.id)} className="btn-danger">
                  Delete
                </button>
              </>
            )}
          </div>
        ))}
      </div>

      {categories.length === 0 && !showNew && (
        <div className="card px-5 py-12 text-center">
          <p className="text-cream-faint text-sm">No categories yet.</p>
          <button
            onClick={() => setShowNew(true)}
            className="mt-3 text-sm text-gold hover:text-gold-light transition-colors"
          >
            Create your first category →
          </button>
        </div>
      )}
    </div>
  )
}
