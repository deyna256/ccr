import { logout } from '../api/auth'

type Tab = 'calendar' | 'settings'

interface NavBarProps {
  activeTab: Tab
  onTabChange: (tab: Tab) => void
}

const TABS: { id: Tab; label: string }[] = [
  { id: 'calendar', label: 'Calendar' },
  { id: 'settings', label: 'Settings' },
]

export default function NavBar({ activeTab, onTabChange }: NavBarProps) {
  return (
    <header className="bg-ink border-b border-ink-border px-5 h-13 flex items-center gap-6 relative z-40">
      <span className="font-serif text-[17px] font-semibold text-cream tracking-tight select-none">
        CCR
      </span>
      <nav className="flex items-center gap-0.5">
        {TABS.map(tab => (
          <button
            key={tab.id}
            onClick={() => onTabChange(tab.id)}
            className={`px-3 py-1.5 text-sm font-medium rounded transition-colors ${
              activeTab === tab.id
                ? 'text-gold bg-gold-glow'
                : 'text-cream-dim hover:text-cream hover:bg-ink-raised'
            }`}
          >
            {tab.label}
          </button>
        ))}
      </nav>
      <div className="ml-auto">
        <button
          onClick={logout}
          className="text-sm text-cream-faint hover:text-cream transition-colors"
        >
          Logout
        </button>
      </div>
    </header>
  )
}
