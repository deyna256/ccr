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
    <div className="min-h-screen dot-grid flex items-center justify-center p-4">
      <form onSubmit={handleSubmit} className="card p-6 w-full max-w-sm space-y-4 animate-slide-up">
        <h1 className="text-xl font-semibold text-cream text-center">Create Account</h1>

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
            minLength={8}
          />
        </div>

        <button
          type="submit"
          disabled={loading}
          className="btn-primary w-full"
        >
          {loading ? 'Creating…' : 'Create Account'}
        </button>

        <p className="text-center text-xs text-cream-faint">
          Have an account?{' '}
          <a href="/login" className="text-gold hover:text-gold-light transition-colors">
            Sign In
          </a>
        </p>
      </form>
    </div>
  )
}
