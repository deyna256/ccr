import { useState } from 'react'
import { register } from '../api/auth'

export default function RegisterPage() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)

    try {
      await register({ email, password })
      window.location.reload()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Registration failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-zinc-950 p-4">
      <form onSubmit={handleSubmit} className="card p-6 w-full max-w-sm space-y-4">
        <h1 className="text-xl font-semibold text-white text-center">Create Account</h1>

        {error && (
          <p className="text-sm text-red-400 bg-red-900/20 p-2 rounded">{error}</p>
        )}

        <div>
          <label className="block text-sm text-zinc-400 mb-1">Email</label>
          <input
            type="email"
            value={email}
            onChange={e => setEmail(e.target.value)}
            className="input-field"
            required
          />
        </div>

        <div>
          <label className="block text-sm text-zinc-400 mb-1">Password</label>
          <input
            type="password"
            value={password}
            onChange={e => setPassword(e.target.value)}
            className="input-field"
            required
            minLength={8}
          />
        </div>

        <button
          type="submit"
          disabled={loading}
          className="btn-primary w-full"
        >
          {loading ? 'Creating...' : 'Create Account'}
        </button>

        <p className="text-center text-sm text-zinc-500">
          Have an account?{' '}
          <a href="/login" className="text-blue-400 hover:text-blue-300">
            Sign In
          </a>
        </p>
      </form>
    </div>
  )
}