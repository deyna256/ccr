# Database Schema

## Tables

| Table          | Purpose                        |
|----------------|--------------------------------|
| users          | User accounts                  |
| categories     | Task categories with colors    |
| tasks          | Tasks               |
| attachments    | Task file / link attachments   |
| refresh_tokens | JWT refresh tokens             |

## Migrations

Managed via [`golang-migrate`](https://github.com/golang-migrate/migrate).
Migration files live in `backend/migrations/`.
