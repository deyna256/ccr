# Requirements

## Functional

| ID    | Requirement                                                              |
|-------|--------------------------------------------------------------------------|
| FR-1  | Week view — current week as the primary view                             |
| FR-2  | Week navigation — scroll forward / backward                              |
| FR-3  | Month view as secondary with navigation                                  |
| FR-4  | Tasks with a fixed time slot (start → end)                    |
| FR-5  | Task description (text); links are embedded in the text                  |
| FR-6  | Task attachments: files                                                  |
| FR-7  | Categories with custom colors                                            |
| FR-8  | Recurring tasks — template-based (move or delete affects the whole series)|
| FR-9  | Complete tasks; archive completed tasks                                   |
| FR-10 | Archive with FIFO eviction policy and size limit                         |

## Non-Functional

| ID    | Requirement                                                              |
|-------|--------------------------------------------------------------------------|
| NFR-1 | Maximum performance — aggressive optimization at every layer             |
| NFR-2 | Offline mode — fully functional without network                          |
| NFR-3 | Sync to server when online                                               |
| NFR-4 | Fast interactions — week switch target < 100 ms                          |

## Constraints

| Area       | Constraint                          |
|------------|-------------------------------------|
| Platform   | Web only                            |
| Stack      | Go, PostgreSQL, Docker Compose      |
| Deployment | Local or cloud (user's choice)      |
| Users      | Single user — no multi-tenancy      |

## Out of Scope

- Auto-scheduling (user places tasks manually)
- Edit single recurrence instance ("edit this one only")
- Mobile applications
- Team collaboration / multi-user
- Real-time sync between users
