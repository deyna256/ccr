import { logout } from '../api/auth'

type Tab = 'calendar' | 'settings'

interface NavBarProps {
  activeTab: Tab
  onTabChange: (tab: Tab) => void
}

export default function NavBar({ activeTab, onTabChange }: NavBarProps) {
  return (
    <nav className="h-12 bg-zinc-900 border-b border-zinc-800 flex items-center justify-between px-4">
      <div className="flex items-center gap-6">
        <span className="text-base font-semibold text-white">CCR</span>
        <div className="flex gap-4">
          <button
            onClick={() => onTabChange('calendar')}
            className={activeTab === 'calendar' ? 'nav-link-active' : 'nav-link'}
          >
            Calendar
          </button>
          <button
            onClick={() => onTabChange('settings')}
            className={activeTab === 'settings' ? 'nav-link-active' : 'nav-link'}
          >
            Settings
          </button>
        </div>
      </div>
      <button
        onClick={logout}
        className="text-sm text-zinc-500 hover:text-white transition-colors"
      >
        Logout
      </button>
    </nav>
  )
}