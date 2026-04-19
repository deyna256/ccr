import { useState } from 'react'
import { login } from '../api/auth'

export default function LoginPage() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)

    try {
      await login({ email, password })
      window.location.href = '/'
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen dot-grid flex items-center justify-center p-4">
      <form onSubmit={handleSubmit} className="card p-6 w-full max-w-sm space-y-4 animate-slide-up">
        <h1 className="text-xl font-semibold text-cream text-center">Sign In</h1>

        {error && (
          <p className="text-xs text-ember border border-ember/20 bg-ember/5 px-4 py-3 rounded">
            {error}
          </p>
        )}

        <div>
          <label className="block text-xs text-cream-faint mb-1.5">Email</label>
          <input
            type="email"
            value={email}
            onChange={e => setEmail(e.target.value)}
            className="input-field"
            required
          />
        </div>

        <div>
          <label className="block text-xs text-cream-faint mb-1.5">Password</label>
          <input
            type="password"
            value={password}
            onChange={e => setPassword(e.target.value)}
            className="input-field"
            required
          />
        </div>

        <button
          type="submit"
          disabled={loading}
          className="btn-primary w-full"
        >
          {loading ? 'Signing in…' : 'Sign In'}
        </button>

        <p className="text-center text-xs text-cream-faint">
          No account?{' '}
          <a href="/register" className="text-gold hover:text-gold-light transition-colors">
            Register
          </a>
        </p>
      </form>
    </div>
  )
}
