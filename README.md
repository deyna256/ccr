# CCR — Control Complete Result

> A personal task and calendar tool built around one idea: full ownership of your time and outcomes.

---

## Idea

Most productivity tools are slow, bloated, and designed for teams.
CCR is the opposite — a focused, fast, single-user tool that puts you in complete control.

No subscriptions. No collaboration overhead. No unnecessary features.
Just a clear view of what needs to be done and when.

## Ideology

**Control** — you decide how tasks are structured, scheduled, and categorized. Nothing is imposed.

**Complete** — every task has a defined scope: a time slot, a duration, a category. Nothing is vague or half-specified.

**Result** — the interface is oriented toward outcomes, not process. The calendar shows what was planned and what was done.

---

## Features

| # | Feature |
|---|---------|
| FR-1 | Week view as primary — current week always visible |
| FR-2 | Forward / backward week navigation |
| FR-3 | Month view as secondary with navigation |
| FR-4 | Fixed-slot tasks (start → end time) |
| FR-5 | Duration-only tasks — auto-scheduled into the next free slot |
| FR-6 | Task descriptions |
| FR-7 | Attachments: files and links |
| FR-8 | Categories with custom colors |
| FR-9 | Recurring tasks |
| FR-10 | Task archive with FIFO eviction policy |

---

## Stack

| Layer | Technology |
|-------|------------|
| Backend | Go, Chi, JWT + Refresh Tokens |
| Frontend | TypeScript, Vite, Service Worker, IndexedDB |
| Database | PostgreSQL |
| Deployment | Docker Compose |

Offline-first: the frontend works fully without a network connection via IndexedDB and syncs to the server when online.

---

## Documentation

- [`docs/api.md`](docs/api.md) — API reference
- [`docs/architecture.md`](docs/architecture.md) — system design, ADRs, project structure
- [`docs/requirements.md`](docs/requirements.md) — functional and non-functional requirements
- [`docs/database.md`](docs/database.md) — schema overview
- [`docs/guidelines/features.md`](docs/guidelines/features.md) — feature implementation guideline
- [`docs/guidelines/logging.md`](docs/guidelines/logging.md) — logging standard
- [`docs/guidelines/testing.md`](docs/guidelines/testing.md) — testing standard

---

## Development

```bash
just up       # build images and start all services
just down     # stop all services
just clean    # remove volumes, images, and build artifacts
```

Backend and frontend each have their own `Justfile` with `test`, `style`, `fmt`, and `publish` targets.

See [`backend/Justfile`](backend/Justfile) and [`frontend/Justfile`](frontend/Justfile).
