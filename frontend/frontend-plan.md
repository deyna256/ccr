# Frontend Implementation Plan

Based on llm-provider/ui style + CCR API + offline-first requirements.

## Structure

```
frontend/src/
├── main.tsx                # entry point
├── App.tsx                 # router + layout
├── index.css               # Tailwind base
├── api/
│   ├── client.ts           # base fetch wrapper + auth handling
│   ├── auth.ts             # login, register, logout, refresh
│   ├── categories.ts       # category CRUD
│   ├── tasks.ts            # task CRUD + list
│   └── attachments.ts       # upload/download
├── components/
│   ├── NavBar.tsx
│   ├── WeekView.tsx
│   ├── MonthView.tsx
│   ├── TaskCard.tsx
│   ├── TaskForm.tsx
│   └── CategoryList.tsx
├── context/
│   └── AuthContext.tsx      # JWT + refresh token logic
├── hooks/
│   ├── useTasks.ts         # task fetching + caching
│   └── useOffline.ts       # IndexedDB sync
└── pages/
    ├── LoginPage.tsx
    ├── RegisterPage.tsx
    ├── CalendarPage.tsx    # week/month view
    └── SettingsPage.tsx   # categories, archive
```

## Offline Strategy

1. **IndexedDB** — primary store when offline
2. **Service Worker** — intercepts requests, serves from IndexedDB
3. **Background Sync** — when online, push pending changes

## Auth Flow

- Store access token in memory (JS variable)
- Refresh token in localStorage (or better: httpOnly cookie approach via backend)
- On 401 — attempt refresh, on failure redirect to login

## Key Components

| Component | Responsibility |
|-----------|---------------|
| `WeekView` | Primary view, 7-day grid, tasks per day |
| `MonthView` | Secondary calendar with dots |
| `TaskForm` | Create/edit: type, slot/duration, category, description |
| `CategoryList` | Color picker + CRUD |