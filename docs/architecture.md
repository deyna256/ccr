# Architecture

## Overview

```
┌─────────────────────────────────────────┐
│          SPA Frontend (TS/Vite)         │
│    Service Worker + IndexedDB           │
│         Background Sync                 │
└────────────────┬────────────────────────┘
                 │ HTTPS JSON
                 ▼
┌─────────────────────────────────────────┐
│          Go + Chi Router                │
│   JWT + Refresh Token Auth Middleware   │
└────────────────┬────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────┐
│           Monolith Core                 │
│  ┌─────────┐ ┌──────────┐ ┌──────────┐ │
│  │  Task   │ │ Calendar │ │ Category │ │
│  │ Handler │ │ Handler  │ │ Handler  │ │
│  └─────────┘ └──────────┘ └──────────┘ │
│  ┌─────────┐ ┌──────────┐ ┌──────────┐ │
│  │  Task   │ │ Calendar │ │ Category │ │
│  │ Service │ │ Service  │ │ Service  │ │
│  └─────────┘ └──────────┘ └──────────┘ │
│  ┌─────────────────────────────────────┐│
│  │        Storage (Interface)          ││
│  └───────────────┬─────────────────────┘│
└──────────────────┼──────────────────────┘
                   │
         ┌─────────┴─────────┐
         ▼                   ▼
┌────────────────┐   ┌────────────────┐
│   PostgreSQL   │   │  File Storage  │
│                │   │ (docker vol)   │
└────────────────┘   └────────────────┘
```

## Technology Stack

| Component    | Technology                                  |
|--------------|---------------------------------------------|
| Router       | Chi                                         |
| Auth         | JWT + Refresh Tokens                        |
| Offline      | Service Worker + IndexedDB + Background Sync|
| Database     | PostgreSQL                                  |
| File Storage | Local (docker volume), S3-compatible        |
| Deployment   | Docker Compose                              |

## Project Structure

```
ccr/
├── backend/
│   ├── cmd/server/        # Entry point: wiring only
│   │   ├── main.go
│   │   └── config.go
│   ├── internal/          # Domain packages
│   │   ├── auth/
│   │   ├── task/
│   │   └── category/
│   ├── migrations/
│   ├── Justfile
│   └── Dockerfile
├── frontend/
│   ├── src/
│   ├── static/
│   ├── Justfile
│   └── Dockerfile
├── docs/
├── docker-compose.yml
├── Justfile
└── .env.example
```

## Architectural Decisions

| ADR   | Decision                                                                 |
|-------|--------------------------------------------------------------------------|
| ADR-1 | Monolith with clear package separation — no premature service split      |
| ADR-2 | SPA frontend — full client-side rendering for performance                |
| ADR-3 | Offline-first — IndexedDB as primary store, sync when online             |
| ADR-4 | Slot validation — overlapping tasks are rejected at the service layer    |
| ADR-5 | Domain-grouped packages (`auth/`, `task/`, `category/`)                  |
