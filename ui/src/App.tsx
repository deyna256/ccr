import { useState, useEffect } from 'react'
import { getAccessToken } from './api/client'
import LoginPage from './pages/LoginPage'
import RegisterPage from './pages/RegisterPage'
import CalendarPage from './pages/CalendarPage'
import SettingsPage from './pages/SettingsPage'
import NavBar from './components/NavBar'

type Tab = 'calendar' | 'settings'

export default function App() {
  const [loading, setLoading] = useState(true)
  const [token, setToken] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState<Tab>('calendar')

  useEffect(() => {
    setToken(getAccessToken())
    setLoading(false)
  }, [])

  useEffect(() => {
    const stored = localStorage.getItem('refresh_token')
    if (stored && !token) {
      setToken('has-refresh')
    }
  }, [token])

  if (loading) {
    return (
      <div className="min-h-screen bg-ink flex items-center justify-center">
        <p className="text-zinc-500">Loading...</p>
      </div>
    )
  }

  // No token at all - show login
  if (!token) {
    return <LoginPage />
  }

  // Show register only if explicitly navigating there
  if (window.location.pathname === '/register') {
    return <RegisterPage />
  }

  // Main app
  return (
    <div className="min-h-screen bg-ink">
      <NavBar activeTab={activeTab} onTabChange={setActiveTab} />
      {activeTab === 'calendar' && <CalendarPage />}
      {activeTab === 'settings' && <SettingsPage />}
    </div>
  )
}