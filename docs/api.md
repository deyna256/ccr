# API

Base path: `/api`  
Auth: JWT Bearer token, required on all routes except `/auth/*`  
Format: `application/json` unless noted

---

## Auth

| Method | Path | Description |
|--------|------|-------------|
| POST | `/auth/register` | Create account |
| POST | `/auth/login` | Get access + refresh tokens |
| POST | `/auth/logout` | Revoke refresh token |
| POST | `/auth/refresh` | Exchange refresh token for new access token |

---

## Categories

| Method | Path | Description |
|--------|------|-------------|
| GET | `/categories` | List all categories |
| POST | `/categories` | Create category |
| PUT | `/categories/{id}` | Update category |
| DELETE | `/categories/{id}` | Delete category |

---

## Tasks

| Method | Path | Description |
|--------|------|-------------|
| GET | `/tasks` | List tasks (see filters below) |
| POST | `/tasks` | Create task or event |
| GET | `/tasks/{id}` | Get single task |
| PUT | `/tasks/{id}` | Update task |
| DELETE | `/tasks/{id}` | Delete task (deletes full series if recurring) |
| PATCH | `/tasks/{id}/status` | Transition status: `pending → done → archived` |

### GET /tasks — query params

| Param | Type | Description |
|-------|------|-------------|
| `from` | date | Range start (inclusive) |
| `to` | date | Range end (inclusive) |
| `type` | `task\|event` | Filter by type |
| `status` | `pending\|done\|archived` | Filter by status |
| `category_id` | UUID | Filter by category |

Recurring tasks: the server expands the template into individual occurrences within the requested range. Each occurrence carries `recurrence_id` (parent task ID) so the client knows they belong to a series.

---

## Attachments

| Method | Path | Description |
|--------|------|-------------|
| GET | `/tasks/{id}/attachments` | List attachments for a task |
| POST | `/tasks/{id}/attachments` | Upload file (`multipart/form-data`) |
| DELETE | `/tasks/{id}/attachments/{attachment_id}` | Delete attachment |
| GET | `/attachments/{id}/file` | Download file |
